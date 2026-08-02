package domainkit_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/domainkit"
	"github.com/wsnacj/agentx-go/extensions/domainmodule"
)

func TestRuntimeRunsModelFreeCaseAndProducesStableDigest(t *testing.T) {
	input := map[string]any{"symbol": "000001", "nested": map[string]any{"limit": float64(1)}}
	runtime := newRuntime(t, domainkit.Handler(func(_ context.Context, args map[string]any) (any, error) {
		args["symbol"] = "mutated"
		args["nested"].(map[string]any)["limit"] = float64(2)
		return map[string]any{"status": "ready", "symbol": "000001"}, nil
	}))

	first, err := runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: " SAMPLE ", CaseID: " quote-fixture ", Tool: "lookup", Arguments: input})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	second, err := runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "sample", CaseID: "quote-fixture", Tool: "lookup", Arguments: input})
	if err != nil {
		t.Fatalf("Run() repeat: %v", err)
	}
	if first.Output != `{"status":"ready","symbol":"000001"}` || first.OutputDigest == "" || first.OutputDigest != second.OutputDigest {
		t.Fatalf("unstable result: first=%#v second=%#v", first, second)
	}
	if input["symbol"] != "000001" || input["nested"].(map[string]any)["limit"] != float64(1) {
		t.Fatalf("handler mutated caller arguments: %#v", input)
	}
}

func TestRuntimePreservesStringAndBytes(t *testing.T) {
	for name, handler := range map[string]domainkit.Handler{
		"string": func(context.Context, map[string]any) (any, error) { return "plain", nil },
		"bytes":  func(context.Context, map[string]any) (any, error) { return []byte("blob"), nil },
	} {
		t.Run(name, func(t *testing.T) {
			runtime := newRuntime(t, handler)
			result, err := runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "sample", Tool: "lookup"})
			if err != nil {
				t.Fatalf("Run(): %v", err)
			}
			want := map[string]string{"string": "plain", "bytes": "blob"}[name]
			if result.Output != want {
				t.Fatalf("output = %q, want %q", result.Output, want)
			}
		})
	}
}

func TestRuntimeTypedErrorsAndCauseIdentity(t *testing.T) {
	boom := errors.New("fixture failed")
	runtime := newRuntime(t, func(context.Context, map[string]any) (any, error) { return nil, boom })
	_, err := runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "sample", Tool: "lookup"})
	if !errors.Is(err, domainkit.ErrHandlerFailed) || !errors.Is(err, boom) {
		t.Fatalf("error identity = %v", err)
	}
	var typed *domainkit.Error
	if !errors.As(err, &typed) || typed.Code != domainkit.ErrorCodeHandlerFailed || typed.Retryable || typed.DisplaySafeMessage != "domain kit handler failed" || err.Error() != boom.Error() {
		t.Fatalf("typed error = %#v (%v)", typed, err)
	}

	_, err = runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "missing", Tool: "lookup"})
	if !errors.Is(err, domainkit.ErrModuleNotFound) {
		t.Fatalf("module error = %v", err)
	}
	_, err = runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "sample", Tool: "missing"})
	if !errors.Is(err, domainkit.ErrToolNotFound) {
		t.Fatalf("tool error = %v", err)
	}
}

func TestRuntimeCancellationAndInvalidConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runtime := newRuntime(t, func(context.Context, map[string]any) (any, error) { return "unexpected", nil })
	if _, err := runtime.Run(ctx, domainkit.RunRequest{ModuleID: "sample", Tool: "lookup"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}

	_, err := domainkit.New(domainkit.Config{Modules: []domainkit.Module{{
		Manifest: domainmodule.Manifest{ID: "sample", Tools: []string{"lookup", "missing"}},
		Handlers: map[string]domainkit.Handler{"lookup": func(context.Context, map[string]any) (any, error) { return nil, nil }},
	}}})
	if !errors.Is(err, domainkit.ErrInvalidConfig) || !strings.Contains(err.Error(), "missing tool handlers") {
		t.Fatalf("config error = %v", err)
	}
}

func TestRuntimeEncodingErrorIsTyped(t *testing.T) {
	runtime := newRuntime(t, func(context.Context, map[string]any) (any, error) { return make(chan int), nil })
	_, err := runtime.Run(context.Background(), domainkit.RunRequest{ModuleID: "sample", Tool: "lookup"})
	if !errors.Is(err, domainkit.ErrEncodingFailed) {
		t.Fatalf("encoding error = %v", err)
	}
}

func newRuntime(t *testing.T, handler domainkit.Handler) *domainkit.Runtime {
	t.Helper()
	runtime, err := domainkit.New(domainkit.Config{Modules: []domainkit.Module{{
		Manifest: domainmodule.Manifest{ID: "sample", Tools: []string{"lookup"}},
		Handlers: map[string]domainkit.Handler{"lookup": handler},
	}}})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	return runtime
}
