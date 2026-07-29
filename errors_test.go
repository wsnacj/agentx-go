package agentx_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

func TestRunCancellationAndDeadlineReturnTypedErrors(t *testing.T) {
	tests := []struct {
		name    string
		code    agentx.ErrorCode
		wantIs  error
		makeCtx func() (context.Context, context.CancelFunc)
		start   func(context.CancelFunc)
	}{
		{
			name:   "cancellation",
			code:   agentx.CodeCanceled,
			wantIs: context.Canceled,
			makeCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			start: func(cancel context.CancelFunc) { cancel() },
		},
		{
			name:   "deadline",
			code:   agentx.CodeDeadlineExceeded,
			wantIs: context.DeadlineExceeded,
			makeCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 30*time.Millisecond)
			},
			start: func(context.CancelFunc) {},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			started := make(chan struct{})
			client := newClient(t, &fakeAdapter{
				runFn: func(ctx context.Context, _ agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
					close(started)
					<-ctx.Done()
					return nil, ctx.Err()
				},
			})

			ctx, cancel := test.makeCtx()
			defer cancel()
			done := make(chan struct{})
			var (
				result agentx.RunResult
				runErr error
			)
			go func() {
				defer close(done)
				result, runErr = client.Run(ctx, agentx.RunRequest{Input: "wait"})
			}()
			select {
			case <-started:
				test.start(cancel)
			case <-ctx.Done():
			}
			<-done

			if !errors.Is(runErr, test.wantIs) ||
				!errors.Is(runErr, &agentx.Error{Code: test.code}) {
				t.Fatalf("Run() error = %v", runErr)
			}
			var typed *agentx.Error
			if !errors.As(runErr, &typed) || typed.Code != test.code || typed.Retryable {
				t.Fatalf("Run() typed error = %#v", typed)
			}
			if len(result.Blockers) != 1 || result.Blockers[0] != string(test.code) {
				t.Fatalf("Run() result = %#v", result)
			}
			shutdownClient(t, client)
		})
	}
}

func TestRunErrorIsDisplaySafeAndPreservesCause(t *testing.T) {
	ownerErr := errors.New("provider secret detail must not be displayed")
	client := newClient(t, &fakeAdapter{
		runFn: func(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
			return nil, ownerErr
		},
	})
	defer shutdownClient(t, client)

	_, err := client.Run(context.Background(), agentx.RunRequest{Input: "fail safely"})
	var typed *agentx.Error
	if !errors.As(err, &typed) || typed.Code != agentx.CodeExecutionFailed {
		t.Fatalf("Run() typed error = %#v, err = %v", typed, err)
	}
	if !errors.Is(err, ownerErr) {
		t.Fatalf("Run() did not preserve cause: %v", err)
	}
	if strings.Contains(err.Error(), "provider secret") {
		t.Fatalf("Run() exposed owner error: %q", err.Error())
	}
}

func TestErrorIsComparesStableCode(t *testing.T) {
	err := &agentx.Error{
		Code:      agentx.CodeClientClosed,
		Retryable: false,
		Message:   "safe message",
	}
	if !errors.Is(err, &agentx.Error{Code: agentx.CodeClientClosed}) {
		t.Fatal("errors.Is did not compare ErrorCode")
	}
	if errors.Is(err, &agentx.Error{Code: agentx.CodeExecutionFailed}) {
		t.Fatal("errors.Is matched a different ErrorCode")
	}
}
