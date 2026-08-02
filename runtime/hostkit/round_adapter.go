package hostkit

import (
	"context"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

// ModelResult is the substrate-neutral result of one host model request.
// Response uses the canonical LLM contract; provider payload and recovery
// policy remain captured by the host function.
type ModelResult struct {
	Response  llm.ChatResponse
	Model     string
	Recovered bool
}

// ModelToolRoundExchange is the immutable view passed from response
// observation to the optional tool gate and concrete tool executor. Reply is
// empty while ObserveResponse is running and populated for later phases.
type ModelToolRoundExchange struct {
	Input toolloop.RoundExecutionInput
	Model ModelResult
	Reply string
}

// ToolResult is the portable result of executing one model-requested tool
// batch. The host still owns authorization, scheduling, persistence, and
// concrete tool implementations.
type ToolResult struct {
	Runs             []toolloop.RunObservation
	Failures         []toolloop.FailureObservation
	NextChunks       []string
	ForceNoToolCalls bool
	DirectAnswer     *ToolDirectAnswer
}

// ToolDirectAnswer is an explicit host decision to finish the current Run
// with a display-safe tool result instead of asking the model for another
// synthesis round. The host owns policy, sanitization, and source selection;
// the portable Host Kit owns outcome projection.
type ToolDirectAnswer struct {
	Reply  string
	Source string
	Reason string
}

// ModelToolRoundConfig supplies concrete model/tool operations while leaving
// round phase ordering and portable projection to ModelToolRoundAdapter.
//
// RequestModel is required. ObserveResponse defaults to Response.Content,
// BeforeTools defaults to true, and ExecuteTools is required only when the
// model response contains tool calls.
type ModelToolRoundConfig struct {
	RequestModel    func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error)
	ObserveResponse func(context.Context, ModelToolRoundExchange) (string, error)
	BeforeTools     func(context.Context, ModelToolRoundExchange) (bool, error)
	ExecuteTools    func(context.Context, ModelToolRoundExchange) (ToolResult, error)
}

// ModelToolRoundResult reports the canonical phase result together with
// defensive copies of the model and tool results.
type ModelToolRoundResult struct {
	Phase toolloop.RoundPhaseResult
	Model ModelResult
	Tools ToolResult
}

// ModelToolRoundAdapter coordinates a model request, response observation,
// optional host tool gate, and concrete tool execution. It is stateless;
// concurrent safety depends only on the supplied host functions.
type ModelToolRoundAdapter struct {
	config ModelToolRoundConfig
}

var _ toolloop.RoundExecutor = (*ModelToolRoundAdapter)(nil)

// NewModelToolRoundAdapter validates and constructs a portable round adapter.
func NewModelToolRoundAdapter(config ModelToolRoundConfig) (*ModelToolRoundAdapter, error) {
	if config.RequestModel == nil {
		return nil, fmt.Errorf("agentx host kit: model requester is required")
	}
	return &ModelToolRoundAdapter{config: config}, nil
}

// Execute coordinates one model/tool round and preserves phase error identity.
// It is useful to hosts that retain richer post-round product policy while
// delegating the generic model/tool phase mechanism.
func (adapter *ModelToolRoundAdapter) Execute(ctx context.Context, input toolloop.RoundExecutionInput) (ModelToolRoundResult, error) {
	if adapter == nil || adapter.config.RequestModel == nil {
		return ModelToolRoundResult{}, fmt.Errorf("agentx host kit: model/tool round adapter is required")
	}
	if ctx == nil {
		return ModelToolRoundResult{}, fmt.Errorf("agentx host kit: nil round context")
	}
	if err := ctx.Err(); err != nil {
		return ModelToolRoundResult{}, err
	}
	execution := &modelToolRoundExecution{
		config: adapter.config,
		input:  cloneRoundExecutionInput(input),
	}
	coordinator, err := toolloop.NewRoundPhaseCoordinator(execution)
	if err != nil {
		return ModelToolRoundResult{}, err
	}
	phase, executeErr := coordinator.Execute(ctx, toolloop.RoundPhaseInput{
		Round:     input.Round,
		MaxRounds: input.MaxRounds,
	})
	return ModelToolRoundResult{
		Phase: phase,
		Model: cloneModelResult(execution.model),
		Tools: cloneToolResult(execution.tools),
	}, executeErr
}

// ExecuteRound implements toolloop.RoundExecutor with the minimal portable
// projection: a model-only response completes, a host gate stop terminates,
// and an executed tool batch continues with ToolResult state.
func (adapter *ModelToolRoundAdapter) ExecuteRound(ctx context.Context, input toolloop.RoundExecutionInput) (toolloop.RoundExecutionResult, error) {
	result, err := adapter.Execute(ctx, input)
	if err != nil {
		return toolloop.RoundExecutionResult{Reply: result.Phase.Reply}, err
	}
	return result.ExecutionResult()
}

// ExecutionResult projects a rich phase result into the portable tool-loop
// outcome. Hosts that need product-specific persistence or telemetry can call
// Execute first and then use this method without reimplementing result policy.
func (result ModelToolRoundResult) ExecutionResult() (toolloop.RoundExecutionResult, error) {
	projected := toolloop.RoundExecutionResult{Reply: result.Phase.Reply}
	switch result.Phase.Kind {
	case toolloop.RoundPhaseNoAction:
		projected.Kind = toolloop.OutcomeCompleted
	case toolloop.RoundPhaseHostStopped:
		projected.Kind = toolloop.OutcomeTerminated
	case toolloop.RoundPhaseActionCompleted:
		if result.Tools.DirectAnswer != nil {
			if strings.TrimSpace(result.Tools.DirectAnswer.Reply) == "" {
				return projected, fmt.Errorf("agentx host kit: tool direct answer reply is required")
			}
			projected.Kind = toolloop.OutcomeCompleted
			projected.Reply = result.Tools.DirectAnswer.Reply
			return projected, nil
		}
		projected.Kind = toolloop.OutcomeContinue
		projected.Continuation = &toolloop.RoundContinuation{
			Calls:            projectToolCalls(result.Model.Response.Calls),
			Runs:             append([]toolloop.RunObservation(nil), result.Tools.Runs...),
			Failures:         append([]toolloop.FailureObservation(nil), result.Tools.Failures...),
			NextChunks:       append([]string(nil), result.Tools.NextChunks...),
			ForceNoToolCalls: result.Tools.ForceNoToolCalls,
		}
	default:
		return projected, fmt.Errorf(
			"agentx host kit: unsupported model/tool round outcome %q",
			result.Phase.Kind,
		)
	}
	return projected, nil
}

type modelToolRoundExecution struct {
	config ModelToolRoundConfig
	input  toolloop.RoundExecutionInput
	model  ModelResult
	reply  string
	tools  ToolResult
}

func (execution *modelToolRoundExecution) Request(ctx context.Context, _ toolloop.RoundPhaseInput) (toolloop.RoundRequestResult, error) {
	model, err := execution.config.RequestModel(ctx, cloneRoundExecutionInput(execution.input))
	execution.model = cloneModelResult(model)
	return toolloop.RoundRequestResult{ActionRequired: len(model.Response.Calls) > 0}, err
}

func (execution *modelToolRoundExecution) Observe(ctx context.Context, _ toolloop.RoundPhaseInput) (toolloop.RoundObserveResult, error) {
	exchange := execution.exchange("")
	if execution.config.ObserveResponse == nil {
		execution.reply = exchange.Model.Response.Content
		return toolloop.RoundObserveResult{Reply: execution.reply}, nil
	}
	reply, err := execution.config.ObserveResponse(ctx, exchange)
	execution.reply = reply
	return toolloop.RoundObserveResult{Reply: reply}, err
}

func (execution *modelToolRoundExecution) BeforeAction(ctx context.Context, _ toolloop.RoundPhaseInput) (bool, error) {
	if execution.config.BeforeTools == nil {
		return true, nil
	}
	return execution.config.BeforeTools(ctx, execution.exchange(execution.reply))
}

func (execution *modelToolRoundExecution) Act(ctx context.Context, _ toolloop.RoundPhaseInput) error {
	if execution.config.ExecuteTools == nil {
		return fmt.Errorf("agentx host kit: tool executor is required")
	}
	tools, err := execution.config.ExecuteTools(ctx, execution.exchange(execution.reply))
	execution.tools = cloneToolResult(tools)
	return err
}

func (execution *modelToolRoundExecution) exchange(reply string) ModelToolRoundExchange {
	return ModelToolRoundExchange{
		Input: cloneRoundExecutionInput(execution.input),
		Model: cloneModelResult(execution.model),
		Reply: reply,
	}
}

func projectToolCalls(calls []llm.FunctionCall) []toolloop.Call {
	if len(calls) == 0 {
		return nil
	}
	projected := make([]toolloop.Call, 0, len(calls))
	for _, call := range calls {
		projected = append(projected, toolloop.Call{
			Name:      call.Name,
			Arguments: call.Arguments,
		})
	}
	return projected
}

func cloneRoundExecutionInput(input toolloop.RoundExecutionInput) toolloop.RoundExecutionInput {
	input.State.Chunks = append([]string(nil), input.State.Chunks...)
	return input
}

func cloneModelResult(result ModelResult) ModelResult {
	result.Response.Raw = append([]byte(nil), result.Response.Raw...)
	result.Response.Calls = append([]llm.FunctionCall(nil), result.Response.Calls...)
	if result.Response.Usage != nil {
		usage := *result.Response.Usage
		result.Response.Usage = &usage
	}
	return result
}

func cloneToolResult(result ToolResult) ToolResult {
	result.Runs = append([]toolloop.RunObservation(nil), result.Runs...)
	result.Failures = append([]toolloop.FailureObservation(nil), result.Failures...)
	result.NextChunks = append([]string(nil), result.NextChunks...)
	if result.DirectAnswer != nil {
		direct := *result.DirectAnswer
		result.DirectAnswer = &direct
	}
	return result
}
