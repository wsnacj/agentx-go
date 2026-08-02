package codex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

type sseEvent struct {
	event string
	data  []byte
}

func readServerSentEvents(reader io.Reader, handle func(sseEvent) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	var eventType string
	var dataLines []string
	flush := func() error {
		if len(dataLines) == 0 {
			eventType = ""
			return nil
		}
		data := []byte(strings.Join(dataLines, "\n"))
		event := eventType
		eventType = ""
		dataLines = dataLines[:0]
		return handle(sseEvent{event: event, data: data})
	}
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, ok := strings.Cut(line, ":")
		if !ok {
			field, value = line, ""
		} else if strings.HasPrefix(value, " ") {
			value = strings.TrimPrefix(value, " ")
		}
		switch field {
		case "event":
			eventType = value
		case "data":
			dataLines = append(dataLines, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read codex stream: %w", err)
	}
	return flush()
}

type streamCollector struct {
	textParts       []string
	textDone        string
	outputItems     []collectedOutputItem
	completed       json.RawMessage
	incomplete      json.RawMessage
	failed          json.RawMessage
	failedMessage   string
	terminalStatus  string
	argumentDeltas  map[string][]string
	completedEvents int
}

type collectedOutputItem struct {
	itemType string
	raw      json.RawMessage
}

func newStreamCollector() *streamCollector {
	return &streamCollector{argumentDeltas: map[string][]string{}}
}

func (c *streamCollector) handle(event sseEvent) error {
	data := bytes.TrimSpace(event.data)
	if len(data) == 0 || bytes.Equal(data, []byte("[DONE]")) {
		return nil
	}
	var payload struct {
		Type        string          `json:"type"`
		Response    json.RawMessage `json:"response"`
		Item        json.RawMessage `json:"item"`
		Delta       string          `json:"delta"`
		Text        string          `json:"text"`
		Arguments   string          `json:"arguments"`
		ItemID      string          `json:"item_id"`
		OutputIndex any             `json:"output_index"`
		Error       any             `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("decode codex stream event %q: %w", event.event, err)
	}
	eventType := strings.TrimSpace(event.event)
	if eventType == "" {
		eventType = strings.TrimSpace(payload.Type)
	}
	switch eventType {
	case "response.output_text.delta":
		if payload.Delta != "" {
			c.textParts = append(c.textParts, payload.Delta)
		}
	case "response.output_text.done":
		if payload.Text != "" {
			c.textDone = payload.Text
		}
	case "response.output_item.done":
		if len(payload.Item) > 0 {
			c.collectOutputItem(payload.Item)
		}
	case "response.function_call_arguments.delta":
		if payload.Delta != "" {
			key := streamArgumentKey(payload.ItemID, payload.OutputIndex)
			c.argumentDeltas[key] = append(c.argumentDeltas[key], payload.Delta)
		}
	case "response.function_call_arguments.done":
		if payload.Arguments != "" {
			key := streamArgumentKey(payload.ItemID, payload.OutputIndex)
			c.argumentDeltas[key] = []string{payload.Arguments}
		}
	case "response.completed":
		c.completedEvents++
		c.terminalStatus = "completed"
		c.completed = responsePayloadOrSelf(payload.Response, data)
	case "response.incomplete":
		c.terminalStatus = "incomplete"
		c.incomplete = responsePayloadOrSelf(payload.Response, data)
	case "response.failed":
		c.terminalStatus = "failed"
		c.failed = responsePayloadOrSelf(payload.Response, data)
		c.failedMessage = stringFromAny(payload.Error)
	case "error":
		c.terminalStatus = "failed"
		c.failed = data
		c.failedMessage = stringFromAny(payload.Error)
	}
	return nil
}

func (c *streamCollector) collectOutputItem(raw json.RawMessage) {
	var item struct {
		Type   string `json:"type"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return
	}
	itemType := strings.ToLower(strings.TrimSpace(item.Type))
	if itemType == "" {
		return
	}
	c.outputItems = append(c.outputItems, collectedOutputItem{itemType: itemType, raw: append(json.RawMessage(nil), raw...)})
}

func (c *streamCollector) finish() (*llm.ChatResponse, *llm.Usage, error) {
	if len(c.failed) > 0 {
		message := strings.TrimSpace(c.failedMessage)
		if message == "" {
			message = string(c.failed)
		}
		return nil, nil, fmt.Errorf("codex stream failed: %s", message)
	}
	if len(c.completed) > 0 {
		response, err := parseResponse(c.completed)
		if err == nil {
			return response, extractUsage(c.completed), nil
		}
		if !c.hasFallbackOutput() {
			return nil, nil, err
		}
	}
	if c.hasFallbackOutput() {
		raw := c.fallbackResponseRaw()
		response, err := parseResponse(raw)
		if err != nil {
			return nil, nil, err
		}
		usage := extractUsage(c.completed)
		if usage == nil {
			usage = extractUsage(c.incomplete)
		}
		return response, usage, nil
	}
	if len(c.incomplete) > 0 {
		return nil, nil, fmt.Errorf("codex stream incomplete: %s", string(c.incomplete))
	}
	if c.completedEvents == 0 {
		return nil, nil, fmt.Errorf("codex stream ended without response.completed")
	}
	return nil, nil, fmt.Errorf("codex stream has no output")
}

func (c *streamCollector) hasFallbackOutput() bool {
	return len(c.outputItems) > 0 || c.fallbackText() != ""
}

func (c *streamCollector) fallbackText() string {
	if strings.TrimSpace(c.textDone) != "" {
		return c.textDone
	}
	return strings.TrimSpace(strings.Join(c.textParts, ""))
}

func (c *streamCollector) fallbackResponseRaw() []byte {
	output := make([]json.RawMessage, 0, len(c.outputItems)+1)
	hasMessage := false
	for _, item := range c.outputItems {
		if item.itemType == "message" {
			hasMessage = true
		}
		output = append(output, item.raw)
	}
	if text := c.fallbackText(); text != "" && !hasMessage {
		message := map[string]any{
			"type": "message", "role": "assistant", "status": "completed",
			"content": []map[string]any{{"type": "output_text", "text": text}},
		}
		raw, _ := json.Marshal(message)
		output = append([]json.RawMessage{raw}, output...)
	}
	status := "completed"
	if c.terminalStatus == "incomplete" {
		status = "incomplete"
	}
	raw, _ := json.Marshal(map[string]any{"status": status, "output": output, "output_text": c.fallbackText()})
	return raw
}

func responsePayloadOrSelf(response json.RawMessage, data []byte) json.RawMessage {
	if len(bytes.TrimSpace(response)) > 0 && !bytes.Equal(bytes.TrimSpace(response), []byte("null")) {
		return append(json.RawMessage(nil), response...)
	}
	return append(json.RawMessage(nil), data...)
}

func streamArgumentKey(itemID string, outputIndex any) string {
	if itemID = strings.TrimSpace(itemID); itemID != "" {
		return itemID
	}
	return stringFromAny(outputIndex)
}
