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

type fakeAdapter struct {
	runFn      func(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error)
	shutdownFn func(context.Context) error
	classifyFn func(error) agentx.ErrorCode
}

func (a *fakeAdapter) Run(ctx context.Context, request agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
	if a.runFn != nil {
		return a.runFn(ctx, request)
	}
	return &agentx.AdapterRunResult{
		RunID:     "run-default",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "ok",
	}, nil
}

func (a *fakeAdapter) Shutdown(ctx context.Context) error {
	if a.shutdownFn != nil {
		return a.shutdownFn(ctx)
	}
	return nil
}

func (a *fakeAdapter) ClassifyError(err error) agentx.ErrorCode {
	if a.classifyFn != nil {
		return a.classifyFn(err)
	}
	return agentx.CodeExecutionFailed
}

func newClient(t *testing.T, adapter agentx.ExecutionAdapter) *agentx.Client {
	t.Helper()
	client, err := agentx.New(agentx.Config{Adapter: adapter})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return client
}

func shutdownClient(t *testing.T, client *agentx.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestRunSuccessAndIdentityProjection(t *testing.T) {
	adapter := &fakeAdapter{
		runFn: func(_ context.Context, request agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
			if request.Input != "hello" || request.SessionID != "session-1" {
				t.Fatalf("adapter request = %#v", request)
			}
			return &agentx.AdapterRunResult{
				RunID:     "run-1",
				SessionID: request.SessionID,
				Status:    "success",
				Reply:     "world",
			}, nil
		},
	}
	client := newClient(t, adapter)
	defer shutdownClient(t, client)

	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "hello",
		SessionID: " session-1 ",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RunID != "run-1" || result.SessionID != "session-1" ||
		result.Status != "completed" || result.Reply != "world" {
		t.Fatalf("Run() result = %#v", result)
	}
	if len(result.Evidence) != 2 ||
		result.Evidence[0] != "run:run-1" ||
		result.Evidence[1] != "session:session-1" {
		t.Fatalf("Run() evidence = %#v", result.Evidence)
	}
	if result.Profile != supportedProfile() {
		t.Fatalf("Run() profile = %#v", result.Profile)
	}
}

func TestRunValidatesContextInputAndProfile(t *testing.T) {
	client := newClient(t, &fakeAdapter{})
	defer shutdownClient(t, client)

	tests := []struct {
		name    string
		ctx     context.Context
		request agentx.RunRequest
	}{
		{name: "nil context", request: agentx.RunRequest{Input: "valid"}},
		{name: "empty input", ctx: context.Background(), request: agentx.RunRequest{Input: " \t "}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := client.Run(test.ctx, test.request)
			if !errors.Is(err, &agentx.Error{Code: agentx.CodeInvalidArgument}) {
				t.Fatalf("Run() error = %v", err)
			}
			if result.Status != "failed" ||
				len(result.Blockers) != 1 ||
				result.Blockers[0] != string(agentx.CodeInvalidArgument) ||
				result.NextAction != "fix_request" {
				t.Fatalf("Run() result = %#v", result)
			}
		})
	}

	if _, err := agentx.New(agentx.Config{}); !errors.Is(err, &agentx.Error{Code: agentx.CodeInvalidArgument}) {
		t.Fatalf("New(empty Config) error = %v", err)
	}
	if _, err := agentx.New(agentx.Config{
		Adapter: &fakeAdapter{},
		Profile: agentx.ExecutionProfile{
			Activation:         "managed",
			ControlMode:        "objective",
			ExecutionIntensity: "l3_managed_objective",
			Driver:             "objective_runtime_loop",
			ResultPolicy:       "objective_closure",
			Lifecycle:          "durable",
		},
	}); !errors.Is(err, &agentx.Error{Code: agentx.CodeUnsupportedProfile}) {
		t.Fatalf("New(unsupported profile) error = %v", err)
	}
}

func TestRunMapsOwnerStatus(t *testing.T) {
	tests := []struct {
		ownerStatus string
		status      string
		blocker     string
		nextAction  string
	}{
		{ownerStatus: "", status: "completed"},
		{ownerStatus: "review_required", status: "blocked", blocker: "execution_incomplete", nextAction: "resolve_execution_blocker"},
		{ownerStatus: "cancelled", status: "canceled", blocker: string(agentx.CodeCanceled), nextAction: "caller_decides_retry"},
		{ownerStatus: "failure", status: "failed", blocker: string(agentx.CodeExecutionFailed), nextAction: "inspect_owner_diagnostics"},
		{ownerStatus: "owner_specific", status: "failed", blocker: string(agentx.CodeExecutionFailed), nextAction: "inspect_owner_diagnostics"},
	}
	for _, test := range tests {
		t.Run(test.ownerStatus, func(t *testing.T) {
			client := newClient(t, &fakeAdapter{
				runFn: func(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
					return &agentx.AdapterRunResult{Status: test.ownerStatus}, nil
				},
			})
			defer shutdownClient(t, client)
			result, err := client.Run(context.Background(), agentx.RunRequest{Input: "status"})
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if result.Status != test.status || result.NextAction != test.nextAction {
				t.Fatalf("Run() result = %#v", result)
			}
			if test.blocker == "" {
				if len(result.Blockers) != 0 {
					t.Fatalf("Run() blockers = %#v", result.Blockers)
				}
			} else if len(result.Blockers) != 1 || result.Blockers[0] != test.blocker {
				t.Fatalf("Run() blockers = %#v", result.Blockers)
			}
		})
	}
}

func TestClientSerializesOverlappingRun(t *testing.T) {
	var active atomic.Int32
	var maxActive atomic.Int32
	client := newClient(t, &fakeAdapter{
		runFn: func(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
			current := active.Add(1)
			defer active.Add(-1)
			for {
				observed := maxActive.Load()
				if current <= observed || maxActive.CompareAndSwap(observed, current) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			return &agentx.AdapterRunResult{Status: "completed"}, nil
		},
	})
	defer shutdownClient(t, client)

	const runs = 4
	var group sync.WaitGroup
	errs := make(chan error, runs)
	for index := 0; index < runs; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := client.Run(context.Background(), agentx.RunRequest{Input: "concurrent"})
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Run() error = %v", err)
		}
	}
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("adapter calls overlapped: max active = %d", got)
	}
}

func TestQueuedRunRespectsDeadline(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	client := newClient(t, &fakeAdapter{
		runFn: func(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
			if calls.Add(1) == 1 {
				close(started)
				<-release
			}
			return &agentx.AdapterRunResult{Status: "completed"}, nil
		},
	})

	firstDone := make(chan error, 1)
	go func() {
		_, err := client.Run(context.Background(), agentx.RunRequest{Input: "hold gate"})
		firstDone <- err
	}()
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := client.Run(ctx, agentx.RunRequest{Input: "expire while queued"})
	if !errors.Is(err, context.DeadlineExceeded) ||
		!errors.Is(err, &agentx.Error{Code: agentx.CodeDeadlineExceeded}) {
		t.Fatalf("queued Run() error = %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("queued Run() reached adapter: calls = %d", got)
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	shutdownClient(t, client)
}

func supportedProfile() agentx.ExecutionProfile {
	return agentx.DefaultExecutionProfile()
}
