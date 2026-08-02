package tools_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/tools"
)

func TestResultMiddlewareUsesExplicitHostClassifier(t *testing.T) {
	called := false
	event := tools.BuildToolResultMiddlewareEvent(tools.ToolResultMiddlewareInput{
		ToolName: "host_private_tool",
		Output:   `{"ok":true}`,
		ClassifyContent: func(name string) (bool, bool) {
			called = true
			if name != "host_private_tool" {
				t.Fatalf("classifier name = %q", name)
			}
			return true, true
		},
	})
	if !called || !event.ExternalContent || !event.UntrustedContent {
		t.Fatalf("event = %#v, called = %v", event, called)
	}
}

func TestControlledTransformPreservesRawReference(t *testing.T) {
	event := tools.BuildToolResultMiddlewareEvent(tools.ToolResultMiddlewareInput{
		ToolName:         "open_page",
		Output:           "line\nline\nline",
		RawResultRef:     "artifact://raw/1",
		LargeResultLines: 2,
	})
	result := tools.BuildControlledToolResultTransform(tools.ToolResultTransformInput{
		Event:  event,
		Output: "line\nline\nline",
	})
	if !result.Applied || result.RawResultRef != "artifact://raw/1" || result.Summary == nil {
		t.Fatalf("result = %#v", result)
	}
}
