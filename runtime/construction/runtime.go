// Package construction owns the Experimental substrate-neutral Runtime
// construction lifecycle.
package construction

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

const cleanupTimeout = 5 * time.Second

var errNilConstructedClient = errors.New("agentx runtime construction: client factory returned nil")

// Config is the substrate-neutral Runtime construction input.
type Config struct {
	WorkspaceRoot string
	ModelProfile  string
	Profile       agentx.ExecutionProfile
}

// ModelRuntime is a host-created model resource owned by the construction
// lifecycle until ownership moves to a later stage.
type ModelRuntime interface {
	Shutdown(context.Context) error
}

// RunnerRuntime is a host-created runner resource owned by the construction
// lifecycle until ownership moves to an adapter.
type RunnerRuntime interface {
	Shutdown(context.Context) error
}

// Host creates concrete resources without exposing their implementation types
// to the construction lifecycle.
type Host interface {
	ResolveModel(context.Context, Config) (ModelRuntime, error)
	NewRunner(context.Context, Config, ModelRuntime) (RunnerRuntime, error)
	NewAdapter(context.Context, Config, RunnerRuntime, ModelRuntime) (agentx.ExecutionAdapter, error)
	ClassifyError(error) agentx.ErrorCode
}

type clientFactory func(agentx.Config) (*agentx.Client, error)

// New validates config and constructs an AgentX Client through the supplied
// host. Successful construction transfers adapter ownership to the Client.
func New(ctx context.Context, config Config, host Host) (*agentx.Client, error) {
	return newWithClientFactory(ctx, config, host, agentx.New)
}

func newWithClientFactory(
	ctx context.Context,
	config Config,
	host Host,
	newClient clientFactory,
) (*agentx.Client, error) {
	validated, err := validateConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.New("construction host is required"),
		)
	}
	if newClient == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.New("client factory is required"),
		)
	}

	modelRuntime, resolveErr := host.ResolveModel(ctx, validated)
	if resolveErr != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, resolveErr, agentx.CodeExecutionFailed),
			errors.Join(resolveErr, shutdownModelRuntime(modelRuntime)),
		)
	}
	if modelRuntime == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.New("model resolver returned an empty runtime"),
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, err, agentx.CodeExecutionFailed),
			errors.Join(err, shutdownModelRuntime(modelRuntime)),
		)
	}

	runnerRuntime, runnerErr := host.NewRunner(ctx, validated, modelRuntime)
	if runnerErr != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, runnerErr, agentx.CodeExecutionFailed),
			errors.Join(
				runnerErr,
				shutdownRunnerRuntime(runnerRuntime),
				shutdownModelRuntime(modelRuntime),
			),
		)
	}
	if runnerRuntime == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.Join(
				errors.New("runner factory returned nil"),
				shutdownModelRuntime(modelRuntime),
			),
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, err, agentx.CodeExecutionFailed),
			errors.Join(
				err,
				shutdownRunnerRuntime(runnerRuntime),
				shutdownModelRuntime(modelRuntime),
			),
		)
	}

	adapter, adapterErr := host.NewAdapter(ctx, validated, runnerRuntime, modelRuntime)
	if adapterErr != nil {
		var cleanupErr error
		if adapter != nil {
			cleanupErr = shutdownAdapter(adapter)
		} else {
			cleanupErr = shutdownRunnerAndModel(runnerRuntime, modelRuntime)
		}
		return nil, newConstructionError(
			constructionErrorCode(host, adapterErr, agentx.CodeExecutionFailed),
			errors.Join(adapterErr, cleanupErr),
		)
	}
	if adapter == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.Join(
				errors.New("adapter factory returned nil"),
				shutdownRunnerAndModel(runnerRuntime, modelRuntime),
			),
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, err, agentx.CodeExecutionFailed),
			errors.Join(err, shutdownAdapter(adapter)),
		)
	}

	client, clientErr := newClient(agentx.Config{
		Adapter: adapter,
		Profile: validated.Profile,
	})
	if clientErr != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, clientErr, agentx.CodeExecutionFailed),
			errors.Join(clientErr, shutdownAdapter(adapter)),
		)
	}
	if client == nil {
		return nil, newConstructionError(
			agentx.CodeExecutionFailed,
			errors.Join(errNilConstructedClient, shutdownAdapter(adapter)),
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, newConstructionError(
			constructionErrorCode(host, err, agentx.CodeExecutionFailed),
			errors.Join(err, shutdownAdapter(adapter)),
		)
	}
	return client, nil
}

func validateConfig(ctx context.Context, config Config) (Config, error) {
	if ctx == nil {
		return Config{}, newConstructionError(
			agentx.CodeInvalidArgument,
			errors.New("nil construction context"),
		)
	}
	if err := ctx.Err(); err != nil {
		return Config{}, newConstructionError(
			constructionErrorCode(nil, err, agentx.CodeExecutionFailed),
			err,
		)
	}

	workspaceRoot := filepath.Clean(strings.TrimSpace(config.WorkspaceRoot))
	if workspaceRoot == "." || !filepath.IsAbs(workspaceRoot) {
		return Config{}, newConstructionError(
			agentx.CodeInvalidArgument,
			errors.New("workspace root must be absolute"),
		)
	}
	modelProfile := strings.TrimSpace(config.ModelProfile)
	if !validModelProfileName(modelProfile) {
		return Config{}, newConstructionError(
			agentx.CodeInvalidArgument,
			errors.New("model profile must be a logical catalog name"),
		)
	}
	if err := validateExecutionProfile(config.Profile); err != nil {
		return Config{}, err
	}
	return Config{
		WorkspaceRoot: workspaceRoot,
		ModelProfile:  modelProfile,
		Profile:       config.Profile,
	}, nil
}

func validModelProfileName(profile string) bool {
	return profile != "" &&
		!strings.Contains(profile, "/") &&
		!strings.Contains(profile, `\`) &&
		!strings.Contains(profile, "://") &&
		!strings.Contains(profile, "=")
}

func validateExecutionProfile(profile agentx.ExecutionProfile) error {
	_, err := agentx.New(agentx.Config{
		Adapter: profileValidationAdapter{},
		Profile: profile,
	})
	return err
}

func constructionErrorCode(host Host, err error, fallback agentx.ErrorCode) agentx.ErrorCode {
	switch {
	case errors.Is(err, context.Canceled):
		return agentx.CodeCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return agentx.CodeDeadlineExceeded
	}
	var typed *agentx.Error
	if errors.As(err, &typed) && typed != nil {
		return normalizeConstructionErrorCode(typed.Code, fallback)
	}
	if host != nil {
		return normalizeConstructionErrorCode(host.ClassifyError(err), fallback)
	}
	return fallback
}

func normalizeConstructionErrorCode(code, fallback agentx.ErrorCode) agentx.ErrorCode {
	switch code {
	case agentx.CodeInvalidArgument,
		agentx.CodeCanceled,
		agentx.CodeDeadlineExceeded,
		agentx.CodeUnsupportedProfile,
		agentx.CodeExecutionFailed:
		return code
	default:
		return fallback
	}
}

func newConstructionError(code agentx.ErrorCode, cause error) error {
	typed := &agentx.Error{
		Code:      code,
		Retryable: false,
		Message:   constructionMessage(code),
	}
	if cause == nil {
		return typed
	}
	return &constructionError{typed: typed, cause: cause}
}

func constructionMessage(code agentx.ErrorCode) string {
	switch code {
	case agentx.CodeInvalidArgument:
		return "invalid runtime configuration"
	case agentx.CodeCanceled:
		return "runtime construction canceled by caller"
	case agentx.CodeDeadlineExceeded:
		return "runtime construction exceeded caller deadline"
	default:
		return "runtime construction failed"
	}
}

type constructionError struct {
	typed *agentx.Error
	cause error
}

func (e *constructionError) Error() string {
	if e == nil || e.typed == nil {
		return ""
	}
	return e.typed.Error()
}

func (e *constructionError) Unwrap() []error {
	if e == nil {
		return nil
	}
	return []error{e.typed, e.cause}
}

func shutdownRunnerRuntime(runner RunnerRuntime) error {
	if runner == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		return fmt.Errorf("cleanup runner: %w", err)
	}
	return nil
}

func shutdownModelRuntime(modelRuntime ModelRuntime) error {
	if modelRuntime == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	if err := modelRuntime.Shutdown(ctx); err != nil {
		return fmt.Errorf("cleanup model runtime: %w", err)
	}
	return nil
}

func shutdownRunnerAndModel(runner RunnerRuntime, modelRuntime ModelRuntime) error {
	return errors.Join(
		shutdownRunnerRuntime(runner),
		shutdownModelRuntime(modelRuntime),
	)
}

func shutdownAdapter(adapter agentx.ExecutionAdapter) error {
	if adapter == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	if err := adapter.Shutdown(ctx); err != nil {
		return fmt.Errorf("cleanup adapter: %w", err)
	}
	return nil
}

type profileValidationAdapter struct{}

func (profileValidationAdapter) Run(context.Context, agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
	return nil, errors.New("profile validation adapter cannot run")
}

func (profileValidationAdapter) Shutdown(context.Context) error {
	return nil
}

func (profileValidationAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
