package llm

import (
	"slices"
	"strings"
)

// StreamEventType identifies the normalized event kind in an event stream.
type StreamEventType string

const (
	StreamEventStart         StreamEventType = "start"
	StreamEventTextStart     StreamEventType = "text_start"
	StreamEventTextDelta     StreamEventType = "text_delta"
	StreamEventTextEnd       StreamEventType = "text_end"
	StreamEventThinkingStart StreamEventType = "thinking_start"
	StreamEventThinkingDelta StreamEventType = "thinking_delta"
	StreamEventThinkingEnd   StreamEventType = "thinking_end"
	StreamEventToolCallStart StreamEventType = "toolcall_start"
	StreamEventToolCallDelta StreamEventType = "toolcall_delta"
	StreamEventToolCallEnd   StreamEventType = "toolcall_end"
	StreamEventUsage         StreamEventType = "usage"
	StreamEventDone          StreamEventType = "done"
	StreamEventError         StreamEventType = "error"
)

// StreamStopReason identifies the normalized terminal stop reason when the provider exposes one.
type StreamStopReason string

const (
	StreamStopReasonStop          StreamStopReason = "stop"
	StreamStopReasonLength        StreamStopReason = "length"
	StreamStopReasonToolUse       StreamStopReason = "tool_use"
	StreamStopReasonContentFilter StreamStopReason = "content_filter"
)

// StreamChunk 表示流式响应的单个增量数据。
type StreamChunk struct {
	Content   string
	Raw       []byte
	Usage     *Usage
	Done      bool
	Err       error
	ToolCalls []FunctionCallDelta
}

// StreamResult 描述 Provider 返回的流式通道及取消函数。
type StreamResult struct {
	Ch     <-chan StreamChunk
	Cancel func()
}

// StreamMessageSnapshot captures the accumulated message state seen so far.
type StreamMessageSnapshot struct {
	Text      []string
	Thinking  []string
	ToolCalls []FunctionCall
}

// StreamEvent describes a richer normalized streaming event.
type StreamEvent struct {
	Type             StreamEventType
	ContentIndex     *int
	TextDelta        string
	TextSnapshot     string
	ThinkingDelta    string
	ThinkingSnapshot string
	ToolCall         *FunctionCallDelta
	ToolCallSnapshot *FunctionCall
	Usage            *Usage
	ResponseID       string
	FinishReason     string
	StopReason       StreamStopReason
	MessageSnapshot  *StreamMessageSnapshot
	Err              error
	Raw              []byte
}

// EventStreamResult describes a stream of normalized events and a cancel hook.
type EventStreamResult struct {
	Ch     <-chan StreamEvent
	Cancel func()
}

// SimpleStreamChunk 表示更轻量的文本流增量。
type SimpleStreamChunk struct {
	Text string
	Done bool
	Err  error
}

// SimpleStreamResult 描述简化文本流及取消函数。
type SimpleStreamResult struct {
	Ch     <-chan SimpleStreamChunk
	Cancel func()
}

// BridgeEventStreamResult downgrades a normalized event stream into the legacy chunk stream.
func BridgeEventStreamResult(res *EventStreamResult) *StreamResult {
	if res == nil {
		return nil
	}
	out := make(chan StreamChunk)
	go func() {
		defer close(out)
		for event := range res.Ch {
			switch event.Type {
			case StreamEventTextStart, StreamEventTextEnd, StreamEventThinkingStart, StreamEventThinkingEnd:
				continue
			case StreamEventTextDelta:
				if event.TextDelta == "" {
					continue
				}
				out <- StreamChunk{Content: event.TextDelta, Raw: event.Raw}
			case StreamEventToolCallStart, StreamEventToolCallDelta, StreamEventToolCallEnd:
				if event.ToolCall == nil {
					continue
				}
				out <- StreamChunk{ToolCalls: []FunctionCallDelta{*event.ToolCall}, Raw: event.Raw}
			case StreamEventUsage:
				if event.Usage == nil {
					continue
				}
				out <- StreamChunk{Usage: event.Usage, Raw: event.Raw}
			case StreamEventError:
				out <- StreamChunk{Err: event.Err, Raw: event.Raw}
			case StreamEventDone:
				out <- StreamChunk{Done: true, Raw: event.Raw}
			}
		}
	}()
	return &StreamResult{
		Ch: out,
		Cancel: func() {
			if res.Cancel != nil {
				res.Cancel()
			}
		},
	}
}

// BridgeEventStreamToSimple downgrades a normalized event stream into a lightweight text-only stream.
func BridgeEventStreamToSimple(res *EventStreamResult) *SimpleStreamResult {
	if res == nil {
		return nil
	}
	out := make(chan SimpleStreamChunk)
	go func() {
		defer close(out)
		for event := range res.Ch {
			switch event.Type {
			case StreamEventTextStart, StreamEventTextEnd, StreamEventThinkingStart, StreamEventThinkingEnd:
				continue
			case StreamEventTextDelta:
				if event.TextDelta == "" {
					continue
				}
				out <- SimpleStreamChunk{Text: event.TextDelta}
			case StreamEventError:
				out <- SimpleStreamChunk{Err: event.Err}
			case StreamEventDone:
				out <- SimpleStreamChunk{Done: true}
			}
		}
	}()
	return &SimpleStreamResult{
		Ch: out,
		Cancel: func() {
			if res.Cancel != nil {
				res.Cancel()
			}
		},
	}
}

// BridgeLegacyStreamResult upgrades the legacy chunk stream into normalized stream events.
func BridgeLegacyStreamResult(res *StreamResult) *EventStreamResult {
	if res == nil {
		return nil
	}
	out := make(chan StreamEvent)
	go func() {
		defer close(out)
		out <- StreamEvent{Type: StreamEventStart}
		textSnapshotByIndex := map[int]string{}
		thinkingSnapshotByIndex := map[int]string{}
		toolCallSnapshotByIndex := map[int]FunctionCall{}
		for chunk := range res.Ch {
			if chunk.Content != "" {
				contentIndex := 0
				if textSnapshotByIndex[contentIndex] == "" {
					textSnapshotByIndex[contentIndex] = chunk.Content
					out <- StreamEvent{
						Type:            StreamEventTextStart,
						ContentIndex:    &contentIndex,
						TextSnapshot:    textSnapshotByIndex[contentIndex],
						MessageSnapshot: BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
						Raw:             chunk.Raw,
					}
				} else {
					textSnapshotByIndex[contentIndex] += chunk.Content
				}
				out <- StreamEvent{
					Type:            StreamEventTextDelta,
					ContentIndex:    &contentIndex,
					TextDelta:       chunk.Content,
					TextSnapshot:    textSnapshotByIndex[contentIndex],
					MessageSnapshot: BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
					Raw:             chunk.Raw,
				}
			}
			for _, call := range chunk.ToolCalls {
				callCopy := call
				contentIndex := call.Index
				_, seen := toolCallSnapshotByIndex[contentIndex]
				snapshot := MergeToolCallSnapshot(toolCallSnapshotByIndex[contentIndex], call)
				toolCallSnapshotByIndex[contentIndex] = snapshot
				if !seen {
					out <- StreamEvent{
						Type:             StreamEventToolCallStart,
						ContentIndex:     &contentIndex,
						ToolCallSnapshot: cloneFunctionCallPointer(snapshot),
						MessageSnapshot:  BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
						Raw:              chunk.Raw,
					}
				}
				out <- StreamEvent{
					Type:             StreamEventToolCallDelta,
					ContentIndex:     &contentIndex,
					ToolCall:         &callCopy,
					ToolCallSnapshot: &snapshot,
					MessageSnapshot:  BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
					Raw:              chunk.Raw,
				}
			}
			if chunk.Usage != nil {
				out <- StreamEvent{
					Type:  StreamEventUsage,
					Usage: chunk.Usage,
					Raw:   chunk.Raw,
				}
			}
			if chunk.Err != nil {
				out <- StreamEvent{
					Type: StreamEventError,
					Err:  chunk.Err,
					Raw:  chunk.Raw,
				}
			}
			if chunk.Done {
				for _, contentIndex := range SortedStringSnapshotIndexes(textSnapshotByIndex) {
					out <- StreamEvent{
						Type:            StreamEventTextEnd,
						ContentIndex:    &contentIndex,
						TextSnapshot:    textSnapshotByIndex[contentIndex],
						MessageSnapshot: BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
						Raw:             chunk.Raw,
					}
				}
				for _, contentIndex := range SortedFunctionCallIndexes(toolCallSnapshotByIndex) {
					out <- StreamEvent{
						Type:             StreamEventToolCallEnd,
						ContentIndex:     &contentIndex,
						ToolCallSnapshot: cloneFunctionCallPointer(toolCallSnapshotByIndex[contentIndex]),
						MessageSnapshot:  BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
						Raw:              chunk.Raw,
					}
				}
				out <- StreamEvent{
					Type:            StreamEventDone,
					MessageSnapshot: BuildStreamMessageSnapshot(textSnapshotByIndex, thinkingSnapshotByIndex, toolCallSnapshotByIndex),
					Raw:             chunk.Raw,
				}
			}
		}
	}()
	return &EventStreamResult{
		Ch: out,
		Cancel: func() {
			if res.Cancel != nil {
				res.Cancel()
			}
		},
	}
}

// NormalizeStreamStopReason maps provider-specific finish reasons into a stable normalized stop reason.
func NormalizeStreamStopReason(reason string) StreamStopReason {
	switch strings.TrimSpace(strings.ToLower(reason)) {
	case "stop":
		return StreamStopReasonStop
	case "length", "max_tokens", "max_output_tokens":
		return StreamStopReasonLength
	case "tool_calls", "function_call":
		return StreamStopReasonToolUse
	case "content_filter", "safety", "recitation", "blocklist", "prohibited_content", "spii":
		return StreamStopReasonContentFilter
	default:
		return ""
	}
}

// MergeToolCallSnapshot folds a streamed tool-call delta into a cumulative tool-call snapshot.
func MergeToolCallSnapshot(current FunctionCall, delta FunctionCallDelta) FunctionCall {
	if delta.ID != "" {
		current.ID = delta.ID
	}
	if delta.Type != "" {
		current.Type = delta.Type
	}
	if delta.Name != "" {
		current.Name = delta.Name
	}
	if delta.Arguments != "" {
		current.Arguments += delta.Arguments
	}
	if delta.ContinuationToken != "" {
		current.ContinuationToken = delta.ContinuationToken
	}
	return current
}

func cloneFunctionCallPointer(call FunctionCall) *FunctionCall {
	callCopy := call
	return &callCopy
}

// BuildStreamMessageSnapshot converts per-content-index partial state into a stable accumulated snapshot.
func BuildStreamMessageSnapshot(textByIndex map[int]string, thinkingByIndex map[int]string, toolCallsByIndex map[int]FunctionCall) *StreamMessageSnapshot {
	maxIndex := maxIndexedSnapshot(textByIndex, thinkingByIndex, toolCallsByIndex)
	if maxIndex < 0 {
		return nil
	}
	snapshot := &StreamMessageSnapshot{
		Text:      make([]string, maxIndex+1),
		Thinking:  make([]string, maxIndex+1),
		ToolCalls: make([]FunctionCall, maxIndex+1),
	}
	for index, text := range textByIndex {
		snapshot.Text[index] = text
	}
	for index, thinking := range thinkingByIndex {
		snapshot.Thinking[index] = thinking
	}
	for index, toolCall := range toolCallsByIndex {
		snapshot.ToolCalls[index] = toolCall
	}
	return snapshot
}

func maxIndexedSnapshot(textByIndex map[int]string, thinkingByIndex map[int]string, toolCallsByIndex map[int]FunctionCall) int {
	maxIndex := -1
	for index, text := range textByIndex {
		if text != "" && index > maxIndex {
			maxIndex = index
		}
	}
	for index, thinking := range thinkingByIndex {
		if thinking != "" && index > maxIndex {
			maxIndex = index
		}
	}
	for index, toolCall := range toolCallsByIndex {
		if (toolCall != FunctionCall{}) && index > maxIndex {
			maxIndex = index
		}
	}
	return maxIndex
}

func SortedFunctionCallIndexes(toolCallsByIndex map[int]FunctionCall) []int {
	if len(toolCallsByIndex) == 0 {
		return nil
	}
	indexes := make([]int, 0, len(toolCallsByIndex))
	for index, toolCall := range toolCallsByIndex {
		if toolCall == (FunctionCall{}) {
			continue
		}
		indexes = append(indexes, index)
	}
	slices.Sort(indexes)
	return indexes
}

// SortedStringSnapshotIndexes returns sorted indexes for non-empty string snapshots.
func SortedStringSnapshotIndexes(valuesByIndex map[int]string) []int {
	if len(valuesByIndex) == 0 {
		return nil
	}
	indexes := make([]int, 0, len(valuesByIndex))
	for index, value := range valuesByIndex {
		if value == "" {
			continue
		}
		indexes = append(indexes, index)
	}
	slices.Sort(indexes)
	return indexes
}
