package transport_test

import (
	"context"
	"net/http"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/transport"
)

func TestResolveAndApplyPreserveExplicitValues(t *testing.T) {
	settings := transport.Resolve(transport.Config{Mode: "default", Headers: map[string]string{"X-Key": "default"}}, llm.RequestOptions{Transport: "request", SessionID: "s1", CacheControl: "no-store", Headers: map[string]string{"X-Key": "request"}})
	payload := map[string]any{"session_id": "explicit"}
	transport.ApplyPayload(payload, settings)
	headers := http.Header{"X-Key": []string{"explicit"}}
	transport.ApplyHeaders(headers, settings)
	if settings.Mode != "request" || payload["session_id"] != "explicit" || headers.Get("X-Key") != "explicit" {
		t.Fatalf("settings=%#v payload=%#v headers=%#v", settings, payload, headers)
	}
	ctx := transport.WithRequestOptions(context.Background(), llm.RequestOptions{SessionID: "ctx"})
	if got := transport.ResolveFromContext(ctx, transport.Config{}).SessionID; got != "ctx" {
		t.Fatalf("session = %q", got)
	}
}
