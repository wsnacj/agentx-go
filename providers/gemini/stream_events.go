package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/transport"
)

func (p *Provider) StreamChatEvents(ctx context.Context, cfg ModelConfig, req types.ChatRequest) (*types.EventStreamResult, error) {
	if !cfg.Capability.Streaming {
		return nil, providers.ErrUnsupported
	}
	req = resolveChatRequest(cfg, req)
	return p.streamEvents(ctx, cfg, req, nil)
}

func (p *Provider) StreamVisionEvents(ctx context.Context, cfg ModelConfig, req types.VisualRequest) (*types.EventStreamResult, error) {
	if !cfg.Capability.Vision {
		return nil, fmt.Errorf("model %s does not support vision", cfg.Name)
	}
	if !cfg.Capability.Streaming {
		return nil, providers.ErrUnsupported
	}
	req = resolveVisualRequest(cfg, req)
	if len(req.Tools) > 0 || req.ToolChoice != nil {
		return nil, providers.ErrUnsupported
	}
	return p.streamEvents(ctx, cfg, chatRequestFromVisualRequest(req), req.Visual)
}

func (p *Provider) streamEvents(ctx context.Context, cfg ModelConfig, req types.ChatRequest, visuals []types.VisualContent) (*types.EventStreamResult, error) {
	payload, err := buildGeneratePayload(ctx, p.resolveMedia, cfg, req, visuals)
	if err != nil {
		return nil, err
	}
	settings := transport.Resolve(p.cfg.Transport, types.RequestOptionsFromMap(req.Options))
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	streamCtx := ctx
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		streamCtx, cancel = context.WithCancel(ctx)
	} else {
		streamCtx, cancel = context.WithTimeout(ctx, defaultStreamTimeout)
	}

	endpoint := addAltSSE(p.buildEndpoint(req.Model, "streamGenerateContent"))
	reqHTTP, err := http.NewRequestWithContext(streamCtx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("build request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Accept", "text/event-stream")
	transport.ApplyHeaders(reqHTTP.Header, settings)
	if p.cfg.Authorize != nil {
		if err := p.cfg.Authorize(streamCtx, reqHTTP.Header); err != nil {
			cancel()
			return nil, err
		}
	}

	resp, err := p.httpClient.Do(reqHTTP)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("send request: %w", err)
	}
	if err := transport.ApplyResponseHook(streamCtx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, endpoint, resp)); err != nil {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("apply response hook: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		return nil, &providers.APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	ch := make(chan types.StreamEvent)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		defer cancel()

		ch <- types.StreamEvent{Type: types.StreamEventStart}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, streamScannerBuffer), streamScannerMaxBuffer)
		parser := geminiStreamEventParser{}

		for scanner.Scan() {
			select {
			case <-streamCtx.Done():
				ch <- types.StreamEvent{Type: types.StreamEventError, Err: streamCtx.Err()}
				for _, event := range parser.DoneEvents(nil) {
					ch <- event
				}
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}
			payloadLine := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payloadLine == "" || payloadLine == "[DONE]" {
				continue
			}

			var streamResp GenerateContentResponse
			if err := json.Unmarshal([]byte(payloadLine), &streamResp); err != nil {
				ch <- types.StreamEvent{Type: types.StreamEventError, Err: fmt.Errorf("decode stream payload: %w", err), Raw: []byte(payloadLine)}
				return
			}
			for _, event := range parser.ParseResponse(&streamResp, []byte(payloadLine)) {
				ch <- event
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- types.StreamEvent{Type: types.StreamEventError, Err: fmt.Errorf("stream read error: %w", err)}
			return
		}
		if streamCtx.Err() != nil {
			ch <- types.StreamEvent{Type: types.StreamEventError, Err: streamCtx.Err()}
		}
		for _, event := range parser.DoneEvents(nil) {
			ch <- event
		}
	}()

	return &types.EventStreamResult{
		Ch: ch,
		Cancel: func() {
			cancel()
			resp.Body.Close()
		},
	}, nil
}

type geminiStreamEventParser struct {
	pendingFinishReason string
	textSnapshots       map[int]string
	thinkingSnapshots   map[int]string
	toolSnapshots       map[int]types.FunctionCall
}

func (p *geminiStreamEventParser) ParseResponse(resp *GenerateContentResponse, raw []byte) []types.StreamEvent {
	if resp == nil {
		return nil
	}
	if p.textSnapshots == nil {
		p.textSnapshots = map[int]string{}
	}
	if p.thinkingSnapshots == nil {
		p.thinkingSnapshots = map[int]string{}
	}
	if p.toolSnapshots == nil {
		p.toolSnapshots = map[int]types.FunctionCall{}
	}

	events := make([]types.StreamEvent, 0, 4)
	if usage := extractStreamUsage(resp); usage != nil {
		events = append(events, types.StreamEvent{
			Type:  types.StreamEventUsage,
			Usage: usage,
			Raw:   raw,
		})
	}
	if len(resp.Candidates) == 0 {
		return events
	}
	candidate := resp.Candidates[0]
	for idx, part := range candidate.Content.Parts {
		if part.FunctionCall != nil {
			argsMap := part.FunctionCall.Args
			if argsMap == nil {
				argsMap = map[string]any{}
			}
			arguments, err := json.Marshal(argsMap)
			if err != nil {
				events = append(events, types.StreamEvent{Type: types.StreamEventError, Err: fmt.Errorf("encode streamed function arguments: %w", err), Raw: raw})
				continue
			}
			incoming := types.FunctionCall{
				ID: part.FunctionCall.ID, Type: "function", Name: part.FunctionCall.Name,
				Arguments: string(arguments), ContinuationToken: part.ThoughtSignature,
			}
			if previous, exists := p.toolSnapshots[idx]; exists {
				if previous != incoming {
					events = append(events, types.StreamEvent{Type: types.StreamEventError, Err: fmt.Errorf("gemini: streamed function call changed after emission"), Raw: raw})
				}
				continue
			}
			p.toolSnapshots[idx] = incoming
			contentIndex := idx
			delta := types.FunctionCallDelta{
				ID: incoming.ID, Type: "function", Name: incoming.Name, Arguments: incoming.Arguments,
				ContinuationToken: incoming.ContinuationToken, Index: contentIndex,
			}
			events = append(events,
				types.StreamEvent{
					Type: types.StreamEventToolCallStart, ContentIndex: &contentIndex,
					ToolCallSnapshot: cloneGeminiFunctionCall(incoming),
					MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots), Raw: raw,
				},
				types.StreamEvent{
					Type: types.StreamEventToolCallDelta, ContentIndex: &contentIndex, ToolCall: &delta,
					ToolCallSnapshot: cloneGeminiFunctionCall(incoming),
					MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots), Raw: raw,
				},
			)
			continue
		}
		if part.Text == "" {
			continue
		}
		contentIndex := idx
		if part.Thought {
			previous := p.thinkingSnapshots[contentIndex]
			if previous == "" {
				p.thinkingSnapshots[contentIndex] = part.Text
				events = append(events, types.StreamEvent{
					Type:             types.StreamEventThinkingStart,
					ContentIndex:     &contentIndex,
					ThinkingSnapshot: p.thinkingSnapshots[contentIndex],
					MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
					Raw:              raw,
				})
			}
			delta, snapshot := advanceSnapshot(previous, part.Text)
			p.thinkingSnapshots[contentIndex] = snapshot
			if delta == "" {
				continue
			}
			events = append(events, types.StreamEvent{
				Type:             types.StreamEventThinkingDelta,
				ContentIndex:     &contentIndex,
				ThinkingDelta:    delta,
				ThinkingSnapshot: snapshot,
				MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
				Raw:              raw,
			})
			continue
		}
		previous := p.textSnapshots[contentIndex]
		if previous == "" {
			p.textSnapshots[contentIndex] = part.Text
			events = append(events, types.StreamEvent{
				Type:            types.StreamEventTextStart,
				ContentIndex:    &contentIndex,
				TextSnapshot:    p.textSnapshots[contentIndex],
				MessageSnapshot: types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
				Raw:             raw,
			})
		}
		delta, snapshot := advanceSnapshot(previous, part.Text)
		p.textSnapshots[contentIndex] = snapshot
		if delta == "" {
			continue
		}
		events = append(events, types.StreamEvent{
			Type:            types.StreamEventTextDelta,
			ContentIndex:    &contentIndex,
			TextDelta:       delta,
			TextSnapshot:    snapshot,
			MessageSnapshot: types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
			Raw:             raw,
		})
	}
	if candidate.FinishReason != "" {
		p.pendingFinishReason = candidate.FinishReason
	}
	return events
}

func (p *geminiStreamEventParser) DoneEvents(raw []byte) []types.StreamEvent {
	events := make([]types.StreamEvent, 0, len(p.thinkingSnapshots)+len(p.textSnapshots)+len(p.toolSnapshots)+1)
	for _, contentIndex := range types.SortedStringSnapshotIndexes(p.thinkingSnapshots) {
		events = append(events, types.StreamEvent{
			Type:             types.StreamEventThinkingEnd,
			ContentIndex:     &contentIndex,
			ThinkingSnapshot: p.thinkingSnapshots[contentIndex],
			MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
			Raw:              raw,
		})
	}
	for _, contentIndex := range types.SortedStringSnapshotIndexes(p.textSnapshots) {
		events = append(events, types.StreamEvent{
			Type:            types.StreamEventTextEnd,
			ContentIndex:    &contentIndex,
			TextSnapshot:    p.textSnapshots[contentIndex],
			MessageSnapshot: types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
			Raw:             raw,
		})
	}
	for _, contentIndex := range types.SortedFunctionCallIndexes(p.toolSnapshots) {
		events = append(events, types.StreamEvent{
			Type:             types.StreamEventToolCallEnd,
			ContentIndex:     &contentIndex,
			ToolCallSnapshot: cloneGeminiFunctionCall(p.toolSnapshots[contentIndex]),
			MessageSnapshot:  types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
			Raw:              raw,
		})
	}
	stopReason := types.NormalizeStreamStopReason(p.pendingFinishReason)
	if len(p.toolSnapshots) > 0 && (stopReason == "" || stopReason == types.StreamStopReasonStop) {
		stopReason = types.StreamStopReasonToolUse
	}
	events = append(events, types.StreamEvent{
		Type:            types.StreamEventDone,
		FinishReason:    p.pendingFinishReason,
		StopReason:      stopReason,
		MessageSnapshot: types.BuildStreamMessageSnapshot(p.textSnapshots, p.thinkingSnapshots, p.toolSnapshots),
		Raw:             raw,
	})
	return events
}

func cloneGeminiFunctionCall(call types.FunctionCall) *types.FunctionCall {
	copy := call
	return &copy
}

func advanceSnapshot(previous string, incoming string) (string, string) {
	if incoming == "" {
		return "", previous
	}
	if previous == "" {
		return incoming, incoming
	}
	if incoming == previous {
		return "", previous
	}
	if strings.HasPrefix(incoming, previous) {
		return strings.TrimPrefix(incoming, previous), incoming
	}
	return incoming, previous + incoming
}

func extractStreamUsage(resp *GenerateContentResponse) *types.Usage {
	if resp == nil || resp.UsageMetadata == nil {
		return nil
	}
	usage := resp.UsageMetadata
	return &types.Usage{
		PromptTokens:     usage.PromptTokenCount,
		CompletionTokens: usage.CandidatesTokenCount,
		TotalTokens:      usage.TotalTokenCount,
		ReasoningTokens:  usage.ThoughtsTokenCount,
	}
}
