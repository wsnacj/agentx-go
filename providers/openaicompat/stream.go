package openaicompat

import (
	"encoding/json"
	"fmt"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

const maxDecodeErrors = 5

type streamToolCall struct {
	Index    int              `json:"index"`
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function toolCallFunction `json:"function"`
}
type streamChoiceDelta struct {
	Content   string           `json:"content"`
	Role      string           `json:"role"`
	ToolCalls []streamToolCall `json:"tool_calls"`
}
type streamChoiceData struct {
	Delta        streamChoiceDelta `json:"delta"`
	FinishReason string            `json:"finish_reason"`
}
type streamResponse struct {
	ID      string             `json:"id"`
	Choices []streamChoiceData `json:"choices"`
	Usage   *llm.Usage         `json:"usage"`
}
type streamParser struct {
	decodeErrors int
	responseID   string
	finishReason string
	text         map[int]string
	calls        map[int]llm.FunctionCall
}

func (p *streamParser) parse(payload string) ([]llm.StreamEvent, bool) {
	if payload == "[DONE]" {
		events := make([]llm.StreamEvent, 0, len(p.text)+len(p.calls)+1)
		for _, index := range llm.SortedStringSnapshotIndexes(p.text) {
			events = append(events, llm.StreamEvent{Type: llm.StreamEventTextEnd, ContentIndex: &index, TextSnapshot: p.text[index], ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
		}
		for _, index := range llm.SortedFunctionCallIndexes(p.calls) {
			call := p.calls[index]
			events = append(events, llm.StreamEvent{Type: llm.StreamEventToolCallEnd, ContentIndex: &index, ToolCallSnapshot: &call, ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
		}
		events = append(events, llm.StreamEvent{Type: llm.StreamEventDone, ResponseID: p.responseID, FinishReason: p.finishReason, StopReason: llm.NormalizeStreamStopReason(p.finishReason), MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
		return events, true
	}
	var response streamResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		p.decodeErrors++
		event := llm.StreamEvent{Type: llm.StreamEventError, Err: fmt.Errorf("decode stream chunk: %w", err), Raw: []byte(payload)}
		if p.decodeErrors >= maxDecodeErrors {
			event.Err = fmt.Errorf("llmx: too many decode errors: %w", err)
			return []llm.StreamEvent{event, {Type: llm.StreamEventDone, Raw: []byte(payload)}}, true
		}
		return []llm.StreamEvent{event}, false
	}
	p.decodeErrors = 0
	if response.ID != "" {
		p.responseID = response.ID
	}
	if p.text == nil {
		p.text = map[int]string{}
	}
	if p.calls == nil {
		p.calls = map[int]llm.FunctionCall{}
	}
	events := make([]llm.StreamEvent, 0, len(response.Choices)+1)
	for _, choice := range response.Choices {
		if choice.Delta.Content != "" {
			index := 0
			if p.text[index] == "" {
				p.text[index] = choice.Delta.Content
				events = append(events, llm.StreamEvent{Type: llm.StreamEventTextStart, ContentIndex: &index, TextSnapshot: p.text[index], ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
			} else {
				p.text[index] += choice.Delta.Content
			}
			events = append(events, llm.StreamEvent{Type: llm.StreamEventTextDelta, ContentIndex: &index, TextDelta: choice.Delta.Content, TextSnapshot: p.text[index], ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
		}
		for _, call := range choice.Delta.ToolCalls {
			index := call.Index
			delta := llm.FunctionCallDelta{ID: call.ID, Type: call.Type, Name: call.Function.Name, Arguments: call.Function.Arguments, Index: index}
			snapshot := llm.MergeToolCallSnapshot(p.calls[index], delta)
			_, seen := p.calls[index]
			p.calls[index] = snapshot
			if !seen {
				copy := snapshot
				events = append(events, llm.StreamEvent{Type: llm.StreamEventToolCallStart, ContentIndex: &index, ToolCallSnapshot: &copy, ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
			}
			deltaCopy, snapshotCopy := delta, snapshot
			events = append(events, llm.StreamEvent{Type: llm.StreamEventToolCallDelta, ContentIndex: &index, ToolCall: &deltaCopy, ToolCallSnapshot: &snapshotCopy, ResponseID: p.responseID, MessageSnapshot: llm.BuildStreamMessageSnapshot(p.text, nil, p.calls), Raw: []byte(payload)})
		}
		if choice.FinishReason != "" {
			p.finishReason = choice.FinishReason
		}
	}
	if response.Usage != nil {
		events = append(events, llm.StreamEvent{Type: llm.StreamEventUsage, Usage: response.Usage, ResponseID: p.responseID, Raw: []byte(payload)})
	}
	return events, false
}
