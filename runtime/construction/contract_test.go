package construction

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

func TestNewValidatesBeforeHostResolution(t *testing.T) {
	tests := map[string]struct {
		ctx    context.Context
		config Config
		code   agentx.ErrorCode
		cause  error
	}{
		"nil context": {
			config: validConfig(t),
			code:   agentx.CodeInvalidArgument,
		},
		"canceled context": {
			ctx:    canceledContext(),
			config: validConfig(t),
			code:   agentx.CodeCanceled,
			cause:  context.Canceled,
		},
		"relative workspace": {
			ctx: context.Background(),
			config: Config{
				WorkspaceRoot: "relative",
				ModelProfile:  "model-a",
			},
			code: agentx.CodeInvalidArgument,
		},
		"empty model profile": {
			ctx: context.Background(),
			config: Config{
				WorkspaceRoot: t.TempDir(),
			},
			code: agentx.CodeInvalidArgument,
		},
		"path model profile": {
			ctx: context.Background(),
			config: Config{
				WorkspaceRoot: t.TempDir(),
				ModelProfile:  "../model.yaml",
			},
			code: agentx.CodeInvalidArgument,
		},
		"unsupported execution profile": {
			ctx: context.Background(),
			config: Config{
				WorkspaceRoot: t.TempDir(),
				ModelProfile:  "model-a",
				Profile:       agentx.ExecutionProfile{Driver: "workflow"},
			},
			code: agentx.CodeUnsupportedProfile,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var resolveCalls atomic.Int32
			host := &fakeHost{
				resolve: func(context.Context, Config) (ModelRuntime, error) {
					resolveCalls.Add(1)
					return nil, errors.New("must not resolve")
				},
			}
			client, err := New(test.ctx, test.config, host)
			if client != nil {
				t.Fatal("New() returned a client")
			}
			if !errors.Is(err, &agentx.Error{Code: test.code}) {
				t.Fatalf("New() error = %v, want code %q", err, test.code)
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("New() error = %v, want cause %v", err, test.cause)
			}
			if resolveCalls.Load() != 0 {
				t.Fatalf("ResolveModel calls = %d, want 0", resolveCalls.Load())
			}
		})
	}
}

func TestNewPassesNormalizedConfigToEveryHostStage(t *testing.T) {
	var configs []Config
	var mu sync.Mutex
	record := func(config Config) {
		mu.Lock()
		defer mu.Unlock()
		configs = append(configs, config)
	}
	model := &fakeResource{}
	runner := &fakeResource{}
	adapter := &fakeAdapter{}
	host := &fakeHost{
		resolve: func(_ context.Context, config Config) (ModelRuntime, error) {
			record(config)
			return model, nil
		},
		newRunner: func(_ context.Context, config Config, _ ModelRuntime) (RunnerRuntime, error) {
			record(config)
			return runner, nil
		},
		newAdapter: func(
			_ context.Context,
			config Config,
			_ RunnerRuntime,
			_ ModelRuntime,
		) (agentx.ExecutionAdapter, error) {
			record(config)
			return adapter, nil
		},
	}
	root := filepath.Join(t.TempDir(), "workspace")
	client, err := New(context.Background(), Config{
		WorkspaceRoot: "  " + root + "  ",
		ModelProfile:  "  model-a  ",
	}, host)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	shutdownClient(t, client)

	want := Config{WorkspaceRoot: root, ModelProfile: "model-a"}
	if len(configs) != 3 {
		t.Fatalf("host config count = %d, want 3", len(configs))
	}
	for i, got := range configs {
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("host config[%d] = %#v, want %#v", i, got, want)
		}
	}
}

func TestNewClassifiesHostErrorWithoutDisplayingPrivateCause(t *testing.T) {
	privateCause := errors.New("private /catalog/path secret-token")
	host := &fakeHost{
		resolve: func(context.Context, Config) (ModelRuntime, error) {
			return nil, privateCause
		},
		classify: func(error) agentx.ErrorCode {
			return agentx.CodeInvalidArgument
		},
	}
	client, err := New(context.Background(), validConfig(t), host)
	if client != nil {
		t.Fatal("New() returned a client")
	}
	if !errors.Is(err, privateCause) ||
		!errors.Is(err, &agentx.Error{Code: agentx.CodeInvalidArgument}) {
		t.Fatalf("New() error = %v", err)
	}
	if strings.Contains(err.Error(), "private") ||
		strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("New() exposed private cause: %q", err.Error())
	}
}

func TestNewFailureCleanupOrderAndOwnership(t *testing.T) {
	stageErr := errors.New("stage failed")
	tests := map[string]struct {
		configure func(*fakeHost, *fakeResource, *fakeResource, *fakeAdapter)
		wantOrder []string
		newClient clientFactory
	}{
		"partial model": {
			configure: func(host *fakeHost, model, _ *fakeResource, _ *fakeAdapter) {
				host.resolve = func(context.Context, Config) (ModelRuntime, error) {
					return model, stageErr
				}
			},
			wantOrder: []string{"model"},
		},
		"partial runner": {
			configure: func(host *fakeHost, _ *fakeResource, runner *fakeResource, _ *fakeAdapter) {
				host.newRunner = func(context.Context, Config, ModelRuntime) (RunnerRuntime, error) {
					return runner, stageErr
				}
			},
			wantOrder: []string{"runner", "model"},
		},
		"nil adapter": {
			configure: func(host *fakeHost, _ *fakeResource, _ *fakeResource, _ *fakeAdapter) {
				host.newAdapter = func(
					context.Context,
					Config,
					RunnerRuntime,
					ModelRuntime,
				) (agentx.ExecutionAdapter, error) {
					return nil, stageErr
				}
			},
			wantOrder: []string{"runner", "model"},
		},
		"partial adapter owns upstream": {
			configure: func(host *fakeHost, _ *fakeResource, _ *fakeResource, adapter *fakeAdapter) {
				host.newAdapter = func(
					context.Context,
					Config,
					RunnerRuntime,
					ModelRuntime,
				) (agentx.ExecutionAdapter, error) {
					return adapter, stageErr
				}
			},
			wantOrder: []string{"adapter"},
		},
		"client failure": {
			configure: func(*fakeHost, *fakeResource, *fakeResource, *fakeAdapter) {},
			wantOrder: []string{"adapter"},
			newClient: func(agentx.Config) (*agentx.Client, error) {
				return nil, stageErr
			},
		},
		"nil client": {
			configure: func(*fakeHost, *fakeResource, *fakeResource, *fakeAdapter) {},
			wantOrder: []string{"adapter"},
			newClient: func(agentx.Config) (*agentx.Client, error) {
				return nil, nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var order []string
			model := &fakeResource{name: "model", order: &order}
			runner := &fakeResource{name: "runner", order: &order}
			adapter := &fakeAdapter{
				fakeResource: fakeResource{name: "adapter", order: &order},
			}
			host := successfulHost(model, runner, adapter)
			test.configure(host, model, runner, adapter)
			newClient := test.newClient
			if newClient == nil {
				newClient = agentx.New
			}
			client, err := newWithClientFactory(
				context.Background(),
				validConfig(t),
				host,
				newClient,
			)
			if client != nil {
				t.Fatal("construction returned a client")
			}
			if err == nil {
				t.Fatal("construction returned nil error")
			}
			if !reflect.DeepEqual(order, test.wantOrder) {
				t.Fatalf("cleanup order = %v, want %v", order, test.wantOrder)
			}
			for _, resource := range []*fakeResource{model, runner, &adapter.fakeResource} {
				if resource.calls.Load() > 1 {
					t.Fatalf("%s cleanup calls = %d, want at most 1", resource.name, resource.calls.Load())
				}
			}
		})
	}
}

func TestNewContextCancellationAfterResolveCleansModel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var order []string
	model := &fakeResource{name: "model", order: &order}
	host := &fakeHost{
		resolve: func(context.Context, Config) (ModelRuntime, error) {
			cancel()
			return model, nil
		},
		newRunner: func(context.Context, Config, ModelRuntime) (RunnerRuntime, error) {
			t.Fatal("NewRunner must not run after cancellation")
			return nil, nil
		},
	}
	client, err := New(ctx, validConfig(t), host)
	if client != nil {
		t.Fatal("New() returned a client")
	}
	if !errors.Is(err, context.Canceled) ||
		!errors.Is(err, &agentx.Error{Code: agentx.CodeCanceled}) {
		t.Fatalf("New() error = %v", err)
	}
	if !reflect.DeepEqual(order, []string{"model"}) {
		t.Fatalf("cleanup order = %v", order)
	}
}

func TestNewSuccessTransfersAdapterOwnershipToClient(t *testing.T) {
	var order []string
	model := &fakeResource{name: "model", order: &order}
	runner := &fakeResource{name: "runner", order: &order}
	adapter := &fakeAdapter{
		fakeResource: fakeResource{name: "adapter", order: &order},
		runResult: &agentx.AdapterRunResult{
			RunID:     "run-1",
			SessionID: "session-1",
			Status:    "completed",
			Reply:     "ok",
		},
	}
	host := successfulHost(model, runner, adapter)
	client, err := New(context.Background(), validConfig(t), host)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(order) != 0 {
		t.Fatalf("resources closed before ownership transfer: %v", order)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "run",
		SessionID: "session-1",
	})
	if err != nil || result.Reply != "ok" || result.RunID != "run-1" {
		t.Fatalf("Run() result=%#v error=%v", result, err)
	}
	shutdownClient(t, client)
	if !reflect.DeepEqual(order, []string{"adapter"}) {
		t.Fatalf("client cleanup order = %v, want adapter only", order)
	}
}

func TestConcurrentNewDoesNotShareLifecycleState(t *testing.T) {
	const count = 8
	var sequence atomic.Int32
	host := &fakeHost{
		resolve: func(context.Context, Config) (ModelRuntime, error) {
			return &fakeResource{}, nil
		},
		newRunner: func(context.Context, Config, ModelRuntime) (RunnerRuntime, error) {
			return &fakeResource{}, nil
		},
		newAdapter: func(
			context.Context,
			Config,
			RunnerRuntime,
			ModelRuntime,
		) (agentx.ExecutionAdapter, error) {
			id := sequence.Add(1)
			return &fakeAdapter{runResult: &agentx.AdapterRunResult{
				RunID:  "run",
				Status: "completed",
				Reply:  string(rune('a' + id)),
			}}, nil
		},
	}
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := New(context.Background(), validConfig(t), host)
			if err == nil {
				shutdownClient(t, client)
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent New() error = %v", err)
		}
	}
	if sequence.Load() != count {
		t.Fatalf("adapter constructions = %d, want %d", sequence.Load(), count)
	}
}

type fakeHost struct {
	resolve    func(context.Context, Config) (ModelRuntime, error)
	newRunner  func(context.Context, Config, ModelRuntime) (RunnerRuntime, error)
	newAdapter func(context.Context, Config, RunnerRuntime, ModelRuntime) (agentx.ExecutionAdapter, error)
	classify   func(error) agentx.ErrorCode
}

func (h *fakeHost) ResolveModel(ctx context.Context, config Config) (ModelRuntime, error) {
	if h.resolve == nil {
		return nil, errors.New("resolve model not configured")
	}
	return h.resolve(ctx, config)
}

func (h *fakeHost) NewRunner(
	ctx context.Context,
	config Config,
	model ModelRuntime,
) (RunnerRuntime, error) {
	if h.newRunner == nil {
		return nil, errors.New("new runner not configured")
	}
	return h.newRunner(ctx, config, model)
}

func (h *fakeHost) NewAdapter(
	ctx context.Context,
	config Config,
	runner RunnerRuntime,
	model ModelRuntime,
) (agentx.ExecutionAdapter, error) {
	if h.newAdapter == nil {
		return nil, errors.New("new adapter not configured")
	}
	return h.newAdapter(ctx, config, runner, model)
}

func (h *fakeHost) ClassifyError(err error) agentx.ErrorCode {
	if h.classify != nil {
		return h.classify(err)
	}
	return agentx.CodeExecutionFailed
}

type fakeResource struct {
	name  string
	order *[]string
	err   error
	calls atomic.Int32
}

func (r *fakeResource) Shutdown(context.Context) error {
	r.calls.Add(1)
	if r.order != nil {
		*r.order = append(*r.order, r.name)
	}
	return r.err
}

type fakeAdapter struct {
	fakeResource
	runResult *agentx.AdapterRunResult
	runErr    error
}

func (a *fakeAdapter) Run(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
	return a.runResult, a.runErr
}

func (a *fakeAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

func successfulHost(
	model ModelRuntime,
	runner RunnerRuntime,
	adapter agentx.ExecutionAdapter,
) *fakeHost {
	return &fakeHost{
		resolve: func(context.Context, Config) (ModelRuntime, error) {
			return model, nil
		},
		newRunner: func(context.Context, Config, ModelRuntime) (RunnerRuntime, error) {
			return runner, nil
		},
		newAdapter: func(
			context.Context,
			Config,
			RunnerRuntime,
			ModelRuntime,
		) (agentx.ExecutionAdapter, error) {
			return adapter, nil
		},
	}
}

func validConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		WorkspaceRoot: t.TempDir(),
		ModelProfile:  "model-a",
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func shutdownClient(t *testing.T, client *agentx.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
