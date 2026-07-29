package agentx_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

var errAdapterClosed = errors.New("adapter closed")

func TestShutdownIsBoundedIdempotentAndRejectsNewRun(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	runDone := make(chan struct{})
	var cancelRun context.CancelFunc
	var lifecycleMu sync.Mutex
	var shutdownCalls atomic.Int32

	adapter := &fakeAdapter{}
	adapter.runFn = func(ctx context.Context, _ agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
		runCtx, cancel := context.WithCancel(ctx)
		lifecycleMu.Lock()
		cancelRun = cancel
		lifecycleMu.Unlock()
		close(started)
		<-runCtx.Done()
		<-release
		close(runDone)
		return nil, errAdapterClosed
	}
	adapter.shutdownFn = func(ctx context.Context) error {
		shutdownCalls.Add(1)
		lifecycleMu.Lock()
		cancel := cancelRun
		lifecycleMu.Unlock()
		if cancel != nil {
			cancel()
		}
		select {
		case <-runDone:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	adapter.classifyFn = func(err error) agentx.ErrorCode {
		if errors.Is(err, errAdapterClosed) {
			return agentx.CodeClientClosed
		}
		return agentx.CodeExecutionFailed
	}
	client := newClient(t, adapter)

	activeDone := make(chan error, 1)
	go func() {
		_, err := client.Run(context.Background(), agentx.RunRequest{Input: "active"})
		activeDone <- err
	}()
	<-started

	shortCtx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	shutdownErr := client.Shutdown(shortCtx)
	if !errors.Is(shutdownErr, context.DeadlineExceeded) ||
		!errors.Is(shutdownErr, &agentx.Error{Code: agentx.CodeShutdownFailed}) {
		t.Fatalf("bounded Shutdown() error = %v", shutdownErr)
	}

	_, closedErr := client.Run(context.Background(), agentx.RunRequest{Input: "after shutdown starts"})
	if !errors.Is(closedErr, &agentx.Error{Code: agentx.CodeClientClosed}) {
		t.Fatalf("Run() after shutdown start error = %v", closedErr)
	}

	close(release)
	if activeErr := <-activeDone; !errors.Is(activeErr, &agentx.Error{Code: agentx.CodeClientClosed}) {
		t.Fatalf("active Run() shutdown error = %v", activeErr)
	}

	shutdownClient(t, client)
	shutdownClient(t, client)
	if got := shutdownCalls.Load(); got != 3 {
		t.Fatalf("adapter Shutdown() calls = %d, want 3 bounded/idempotent attempts", got)
	}
}

func TestShutdownValidatesContext(t *testing.T) {
	client := newClient(t, &fakeAdapter{})
	if err := client.Shutdown(nil); !errors.Is(err, &agentx.Error{Code: agentx.CodeInvalidArgument}) {
		t.Fatalf("Shutdown(nil) error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.Shutdown(ctx); !errors.Is(err, context.Canceled) ||
		!errors.Is(err, &agentx.Error{Code: agentx.CodeShutdownFailed}) {
		t.Fatalf("Shutdown(canceled) error = %v", err)
	}
	shutdownClient(t, client)
}
