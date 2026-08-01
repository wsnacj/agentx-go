package hostkit

import (
	"context"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

// ModelToolClientConfig is the low-boilerplate Host Kit path for callers that
// can build model/tool functions per Run but do not need custom assembly
// policy. Advanced hosts can continue to use Config and Factory directly.
type ModelToolClientConfig struct {
	Profile   agentx.ExecutionProfile
	MaxRounds int

	BuildRound func(context.Context, execution.Request) (ModelToolRoundConfig, error)

	ResolveIdentity func(execution.Request) (runID string, sessionID string)
	InitialState    func(execution.Request) toolloop.RoundState
	Shutdown        func(context.Context) error
	ClassifyError   func(error) agentx.ErrorCode
}

// NewModelToolClient constructs a root Client without requiring callers to
// implement Factory, BuildRun, or toolloop.RoundExecutor. It does not install
// provider, authorization, persistence, detector, or product-policy defaults.
func NewModelToolClient(config ModelToolClientConfig) (*agentx.Client, error) {
	if config.MaxRounds <= 0 {
		return nil, fmt.Errorf("agentx host kit: max rounds must be positive")
	}
	if config.BuildRound == nil {
		return nil, fmt.Errorf("agentx host kit: round builder is required")
	}
	return New(Config{
		Factory: &modelToolFactory{config: config},
		Profile: config.Profile,
	})
}

type modelToolFactory struct {
	config ModelToolClientConfig
}

func (factory *modelToolFactory) BuildRun(ctx context.Context, request execution.Request) (RunConfig, error) {
	config := RunConfig{SessionID: request.SessionID}
	if factory == nil || factory.config.BuildRound == nil || factory.config.MaxRounds <= 0 {
		return config, fmt.Errorf("agentx host kit: model/tool factory is required")
	}
	if ctx == nil {
		return config, fmt.Errorf("agentx host kit: nil run context")
	}
	if err := ctx.Err(); err != nil {
		return config, err
	}
	if factory.config.ResolveIdentity != nil {
		config.RunID, config.SessionID = factory.config.ResolveIdentity(request)
	}
	roundConfig, err := factory.config.BuildRound(ctx, request)
	if err != nil {
		return config, err
	}
	round, err := NewModelToolRoundAdapter(roundConfig)
	if err != nil {
		return config, err
	}
	initial := toolloop.RoundState{Chunks: []string{request.Input}}
	if factory.config.InitialState != nil {
		initial = factory.config.InitialState(request)
	}
	config.Assembly = toolloop.AssemblyConfig{
		MaxRounds: factory.config.MaxRounds,
		Coordinator: toolloop.CoordinatorConfig{
			Executor: round,
		},
		Initial: cloneRoundState(initial),
	}
	return config, nil
}

func (factory *modelToolFactory) Shutdown(ctx context.Context) error {
	if factory == nil {
		return fmt.Errorf("agentx host kit: model/tool factory is required")
	}
	if ctx == nil {
		return fmt.Errorf("agentx host kit: nil shutdown context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if factory.config.Shutdown == nil {
		return nil
	}
	return factory.config.Shutdown(ctx)
}

func (factory *modelToolFactory) ClassifyError(err error) agentx.ErrorCode {
	if factory == nil || factory.config.ClassifyError == nil {
		return agentx.CodeExecutionFailed
	}
	return factory.config.ClassifyError(err)
}

func cloneRoundState(state toolloop.RoundState) toolloop.RoundState {
	state.Chunks = append([]string(nil), state.Chunks...)
	return state
}
