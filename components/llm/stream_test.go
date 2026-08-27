package llm

import "testing"

func TestNormalizeStreamStopReason(t *testing.T) {
	cases := map[string]StreamStopReason{
		"stop":           StreamStopReasonStop,
		"length":         StreamStopReasonLength,
		"tool_calls":     StreamStopReasonToolUse,
		"function_call":  StreamStopReasonToolUse,
		"content_filter": StreamStopReasonContentFilter,
		"unknown":        "",
	}

	for in, want := range cases {
		if got := NormalizeStreamStopReason(in); got != want {
			t.Fatalf("normalize %q: got %q want %q", in, got, want)
		}
	}
}

func TestMergeToolCallSnapshot(t *testing.T) {
	snapshot := MergeToolCallSnapshot(FunctionCall{}, FunctionCallDelta{
		ID:                "call_1",
		Type:              "function",
		Name:              "lookup",
		Arguments:         "{\"q\":\"hel",
		ContinuationToken: "opaque-provider-token",
		Index:             0,
	})
	snapshot = MergeToolCallSnapshot(snapshot, FunctionCallDelta{
		Arguments: "lo\"}",
		Index:     0,
	})
	if snapshot.ID != "call_1" || snapshot.Type != "function" || snapshot.Name != "lookup" || snapshot.ContinuationToken != "opaque-provider-token" {
		t.Fatalf("expected snapshot metadata preserved, got %#v", snapshot)
	}
	if snapshot.Arguments != "{\"q\":\"hello\"}" {
		t.Fatalf("expected cumulative arguments, got %#v", snapshot)
	}
}

func TestBridgeEventStreamResultDowngradesToLegacyChunks(t *testing.T) {
	events := make(chan StreamEvent, 4)
	events <- StreamEvent{Type: StreamEventStart}
	events <- StreamEvent{Type: StreamEventTextDelta, TextDelta: "hi", Raw: []byte("text")}
	events <- StreamEvent{Type: StreamEventUsage, Usage: &Usage{TotalTokens: 3}, Raw: []byte("usage")}
	events <- StreamEvent{Type: StreamEventDone, Raw: []byte("done")}
	close(events)

	stream := BridgeEventStreamResult(&EventStreamResult{Ch: events, Cancel: func() {}})
	var chunks []StreamChunk
	for chunk := range stream.Ch {
		chunks = append(chunks, chunk)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 downgraded chunks, got %d", len(chunks))
	}
	if got := chunks[0].Content; got != "hi" {
		t.Fatalf("expected text delta chunk, got %q", got)
	}
	if chunks[1].Usage == nil || chunks[1].Usage.TotalTokens != 3 {
		t.Fatalf("expected usage chunk, got %#v", chunks[1].Usage)
	}
	if !chunks[2].Done {
		t.Fatalf("expected done chunk, got %#v", chunks[2])
	}
}

func TestBridgeLegacyStreamResultUpgradesToEvents(t *testing.T) {
	ch := make(chan StreamChunk, 4)
	ch <- StreamChunk{Content: "hi", Raw: []byte("text")}
	ch <- StreamChunk{Content: "!", Raw: []byte("text2")}
	ch <- StreamChunk{ToolCalls: []FunctionCallDelta{{Name: "lookup", Arguments: "{}", Index: 2}}, Raw: []byte("tool")}
	ch <- StreamChunk{Done: true, Raw: []byte("done")}
	close(ch)

	events := BridgeLegacyStreamResult(&StreamResult{Ch: ch, Cancel: func() {}})
	var got []StreamEvent
	for event := range events.Ch {
		got = append(got, event)
	}
	if len(got) != 9 {
		t.Fatalf("expected 9 events including text/tool lifecycle, got %d", len(got))
	}
	if got[0].Type != StreamEventStart {
		t.Fatalf("expected first event start, got %s", got[0].Type)
	}
	if got[1].Type != StreamEventTextStart || got[1].TextSnapshot != "hi" {
		t.Fatalf("expected text start event, got %#v", got[1])
	}
	if got[2].Type != StreamEventTextDelta || got[2].TextDelta != "hi" {
		t.Fatalf("expected text delta event, got %#v", got[2])
	}
	if got[2].ContentIndex == nil || *got[2].ContentIndex != 0 {
		t.Fatalf("expected text delta content index 0, got %#v", got[2].ContentIndex)
	}
	if got[2].TextSnapshot != "hi" {
		t.Fatalf("expected first text snapshot hi, got %#v", got[2].TextSnapshot)
	}
	if got[2].MessageSnapshot == nil || got[2].MessageSnapshot.Text[0] != "hi" {
		t.Fatalf("expected text delta message snapshot hi, got %#v", got[2].MessageSnapshot)
	}
	if got[3].Type != StreamEventTextDelta || got[3].TextDelta != "!" {
		t.Fatalf("expected second text delta event, got %#v", got[3])
	}
	if got[3].TextSnapshot != "hi!" {
		t.Fatalf("expected cumulative text snapshot hi!, got %#v", got[3].TextSnapshot)
	}
	if got[4].Type != StreamEventToolCallStart || got[4].ToolCallSnapshot == nil || got[4].ToolCallSnapshot.Name != "lookup" {
		t.Fatalf("expected tool call start event, got %#v", got[4])
	}
	if got[5].Type != StreamEventToolCallDelta || got[5].ToolCall == nil || got[5].ToolCall.Name != "lookup" {
		t.Fatalf("expected tool call delta event, got %#v", got[5])
	}
	if got[5].ContentIndex == nil || *got[5].ContentIndex != 2 {
		t.Fatalf("expected tool call content index 2, got %#v", got[5].ContentIndex)
	}
	if got[5].MessageSnapshot == nil || got[5].MessageSnapshot.ToolCalls[2].Name != "lookup" {
		t.Fatalf("expected tool delta message snapshot lookup, got %#v", got[5].MessageSnapshot)
	}
	if got[6].Type != StreamEventTextEnd || got[6].TextSnapshot != "hi!" {
		t.Fatalf("expected text end event, got %#v", got[6])
	}
	if got[7].Type != StreamEventToolCallEnd || got[7].ToolCallSnapshot == nil || got[7].ToolCallSnapshot.Name != "lookup" {
		t.Fatalf("expected tool call end event, got %#v", got[7])
	}
	if got[8].Type != StreamEventDone {
		t.Fatalf("expected done event, got %#v", got[8])
	}
	if got[8].MessageSnapshot == nil {
		t.Fatalf("expected done event message snapshot, got %#v", got[8])
	}
	if got[8].MessageSnapshot.Text[0] != "hi!" {
		t.Fatalf("expected final text snapshot hi!, got %#v", got[8].MessageSnapshot.Text)
	}
	if got[8].MessageSnapshot.ToolCalls[2].Name != "lookup" {
		t.Fatalf("expected final tool snapshot lookup, got %#v", got[8].MessageSnapshot.ToolCalls)
	}
}

func TestBridgeLegacyStreamResultEmitsToolCallLifecycle(t *testing.T) {
	ch := make(chan StreamChunk, 2)
	ch <- StreamChunk{ToolCalls: []FunctionCallDelta{{Index: 1, ID: "call_1", Type: "function", Name: "lookup", Arguments: "{}"}}, Raw: []byte("tool")}
	ch <- StreamChunk{Done: true, Raw: []byte("done")}
	close(ch)

	events := BridgeLegacyStreamResult(&StreamResult{Ch: ch, Cancel: func() {}})
	var toolEvents []StreamEvent
	for event := range events.Ch {
		if event.Type == StreamEventToolCallStart || event.Type == StreamEventToolCallDelta || event.Type == StreamEventToolCallEnd {
			toolEvents = append(toolEvents, event)
		}
	}
	if len(toolEvents) != 3 {
		t.Fatalf("expected tool lifecycle events, got %#v", toolEvents)
	}
	if toolEvents[0].Type != StreamEventToolCallStart {
		t.Fatalf("expected toolcall_start, got %#v", toolEvents[0])
	}
	if toolEvents[1].Type != StreamEventToolCallDelta {
		t.Fatalf("expected toolcall_delta, got %#v", toolEvents[1])
	}
	if toolEvents[2].Type != StreamEventToolCallEnd {
		t.Fatalf("expected toolcall_end, got %#v", toolEvents[2])
	}
}

func TestBridgeLegacyStreamResultEmitsTextLifecycle(t *testing.T) {
	ch := make(chan StreamChunk, 2)
	ch <- StreamChunk{Content: "hello", Raw: []byte("text")}
	ch <- StreamChunk{Done: true, Raw: []byte("done")}
	close(ch)

	events := BridgeLegacyStreamResult(&StreamResult{Ch: ch, Cancel: func() {}})
	var textEvents []StreamEvent
	for event := range events.Ch {
		if event.Type == StreamEventTextStart || event.Type == StreamEventTextDelta || event.Type == StreamEventTextEnd {
			textEvents = append(textEvents, event)
		}
	}
	if len(textEvents) != 3 {
		t.Fatalf("expected text lifecycle events, got %#v", textEvents)
	}
	if textEvents[0].Type != StreamEventTextStart {
		t.Fatalf("expected text_start, got %#v", textEvents[0])
	}
	if textEvents[1].Type != StreamEventTextDelta {
		t.Fatalf("expected text_delta, got %#v", textEvents[1])
	}
	if textEvents[2].Type != StreamEventTextEnd {
		t.Fatalf("expected text_end, got %#v", textEvents[2])
	}
}

func TestBridgeLegacyStreamResultBuildsTextSnapshotPerContentIndex(t *testing.T) {
	ch := make(chan StreamChunk, 3)
	ch <- StreamChunk{Content: "he", Raw: []byte("text")}
	ch <- StreamChunk{Content: "llo", Raw: []byte("text2")}
	ch <- StreamChunk{Done: true, Raw: []byte("done")}
	close(ch)

	events := BridgeLegacyStreamResult(&StreamResult{Ch: ch, Cancel: func() {}})
	var textEvents []StreamEvent
	for event := range events.Ch {
		if event.Type == StreamEventTextDelta {
			textEvents = append(textEvents, event)
		}
	}
	if len(textEvents) != 2 {
		t.Fatalf("expected 2 text events, got %d", len(textEvents))
	}
	if textEvents[0].TextSnapshot != "he" {
		t.Fatalf("expected first text snapshot he, got %#v", textEvents[0].TextSnapshot)
	}
	if textEvents[1].TextSnapshot != "hello" {
		t.Fatalf("expected cumulative text snapshot hello, got %#v", textEvents[1].TextSnapshot)
	}
}

func TestBridgeLegacyStreamResultBuildsToolCallSnapshotPerContentIndex(t *testing.T) {
	ch := make(chan StreamChunk, 3)
	ch <- StreamChunk{ToolCalls: []FunctionCallDelta{{Index: 1, ID: "call_1", Type: "function", Name: "lookup", Arguments: "{\"q\":\"hel", ContinuationToken: "opaque-provider-token"}}, Raw: []byte("tool1")}
	ch <- StreamChunk{ToolCalls: []FunctionCallDelta{{Index: 1, Arguments: "lo\"}"}}, Raw: []byte("tool2")}
	ch <- StreamChunk{Done: true, Raw: []byte("done")}
	close(ch)

	events := BridgeLegacyStreamResult(&StreamResult{Ch: ch, Cancel: func() {}})
	var toolEvents []StreamEvent
	for event := range events.Ch {
		if event.Type == StreamEventToolCallDelta {
			toolEvents = append(toolEvents, event)
		}
	}
	if len(toolEvents) != 2 {
		t.Fatalf("expected 2 tool events, got %d", len(toolEvents))
	}
	if toolEvents[0].ToolCallSnapshot == nil || toolEvents[0].ToolCallSnapshot.Arguments != "{\"q\":\"hel" || toolEvents[0].ToolCallSnapshot.ContinuationToken != "opaque-provider-token" {
		t.Fatalf("expected first tool-call snapshot, got %#v", toolEvents[0].ToolCallSnapshot)
	}
	if toolEvents[1].ToolCallSnapshot == nil || toolEvents[1].ToolCallSnapshot.Arguments != "{\"q\":\"hello\"}" {
		t.Fatalf("expected cumulative tool-call snapshot, got %#v", toolEvents[1].ToolCallSnapshot)
	}
}

func TestBridgeEventStreamToSimpleDowngradesToTextOnlyChunks(t *testing.T) {
	events := make(chan StreamEvent, 6)
	events <- StreamEvent{Type: StreamEventStart}
	events <- StreamEvent{Type: StreamEventThinkingDelta, ThinkingDelta: "plan"}
	events <- StreamEvent{Type: StreamEventTextDelta, TextDelta: "hello"}
	events <- StreamEvent{Type: StreamEventUsage, Usage: &Usage{TotalTokens: 5}}
	events <- StreamEvent{Type: StreamEventDone}
	close(events)

	stream := BridgeEventStreamToSimple(&EventStreamResult{Ch: events, Cancel: func() {}})
	var got []SimpleStreamChunk
	for chunk := range stream.Ch {
		got = append(got, chunk)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 simple chunks, got %d", len(got))
	}
	if got[0].Text != "hello" {
		t.Fatalf("expected text delta preserved, got %#v", got[0])
	}
	if !got[1].Done {
		t.Fatalf("expected done chunk, got %#v", got[1])
	}
}

func TestBuildStreamMessageSnapshotCombinesPerIndexState(t *testing.T) {
	snapshot := BuildStreamMessageSnapshot(
		map[int]string{1: "hello"},
		map[int]string{0: "plan"},
		map[int]FunctionCall{2: {Name: "lookup", Arguments: "{}"}},
	)
	if snapshot == nil {
		t.Fatalf("expected snapshot")
	}
	if len(snapshot.Text) != 3 || len(snapshot.Thinking) != 3 || len(snapshot.ToolCalls) != 3 {
		t.Fatalf("expected snapshot slices sized to max index, got %#v", snapshot)
	}
	if snapshot.Thinking[0] != "plan" {
		t.Fatalf("expected thinking snapshot at index 0, got %#v", snapshot.Thinking)
	}
	if snapshot.Text[1] != "hello" {
		t.Fatalf("expected text snapshot at index 1, got %#v", snapshot.Text)
	}
	if snapshot.ToolCalls[2].Name != "lookup" {
		t.Fatalf("expected tool snapshot at index 2, got %#v", snapshot.ToolCalls)
	}
}
