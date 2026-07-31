package execution

import (
	"context"
	"errors"
	"reflect"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
)

type hostStub struct {
	run      func(context.Context, Request) (*Result, error)
	shutdown func(context.Context) error
	code     agentx.ErrorCode
}

func (h hostStub) Run(ctx context.Context, request Request) (*Result, error) {
	return h.run(ctx, request)
}

func (h hostStub) Shutdown(ctx context.Context) error {
	if h.shutdown == nil {
		return nil
	}
	return h.shutdown(ctx)
}

func (h hostStub) ClassifyError(error) agentx.ErrorCode {
	return h.code
}

func TestRuntimeDispatchesAndAssemblesWithoutChangingIdentity(t *testing.T) {
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("identity"), "kept")
	wantRequest := Request{Input: " keep input ", SessionID: "session-1"}
	wantResult := &Result{
		RunID:     "run-1",
		SessionID: "session-1",
		Status:    "incomplete",
		Reply:     "partial",
	}
	sentinel := errors.New("host failure")
	runtime, err := New(hostStub{
		run: func(got context.Context, request Request) (*Result, error) {
			if got != ctx || got.Value(contextKey("identity")) != "kept" {
				t.Fatal("context identity changed")
			}
			if !reflect.DeepEqual(request, wantRequest) {
				t.Fatalf("request = %#v, want %#v", request, wantRequest)
			}
			return wantResult, sentinel
		},
	})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, runErr := runtime.Run(ctx, agentx.AdapterRunRequest{
		Input:     wantRequest.Input,
		SessionID: wantRequest.SessionID,
	})
	if !errors.Is(runErr, sentinel) {
		t.Fatalf("Run() error = %v", runErr)
	}
	wantAdapterResult := &agentx.AdapterRunResult{
		RunID:     wantResult.RunID,
		SessionID: wantResult.SessionID,
		Status:    wantResult.Status,
		Reply:     wantResult.Reply,
	}
	if !reflect.DeepEqual(result, wantAdapterResult) {
		t.Fatalf("Run() result = %#v, want %#v", result, wantAdapterResult)
	}
}

func TestRuntimePreservesNilResultShutdownAndClassification(t *testing.T) {
	sentinel := errors.New("host failure")
	shutdownCtx := context.Background()
	runtime, err := New(hostStub{
		run: func(context.Context, Request) (*Result, error) {
			return nil, sentinel
		},
		shutdown: func(ctx context.Context) error {
			if ctx != shutdownCtx {
				t.Fatal("shutdown context identity changed")
			}
			return sentinel
		},
		code: agentx.CodeClientClosed,
	})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, runErr := runtime.Run(context.Background(), agentx.AdapterRunRequest{Input: "run"})
	if result != nil || !errors.Is(runErr, sentinel) {
		t.Fatalf("Run() = %#v, %v", result, runErr)
	}
	if err := runtime.Shutdown(shutdownCtx); !errors.Is(err, sentinel) {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if code := runtime.ClassifyError(sentinel); code != agentx.CodeClientClosed {
		t.Fatalf("ClassifyError() = %q", code)
	}
}

func TestRuntimeRejectsMissingHost(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) error = nil")
	}
	var runtime *Runtime
	if _, err := runtime.Run(context.Background(), agentx.AdapterRunRequest{}); err == nil {
		t.Fatal("nil Runtime.Run() error = nil")
	}
	if err := runtime.Shutdown(context.Background()); err == nil {
		t.Fatal("nil Runtime.Shutdown() error = nil")
	}
	if code := runtime.ClassifyError(errors.New("failure")); code != agentx.CodeExecutionFailed {
		t.Fatalf("nil Runtime.ClassifyError() = %q", code)
	}
}
