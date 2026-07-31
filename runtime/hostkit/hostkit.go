// Package hostkit 组合 portable tool-loop、execution adapter 和根 AgentX
// Client，为不依赖特定 Runner 的宿主提供最小执行内核。
//
// 当前 package 处于 Experimental/private validation。它不提供模型、工具、
// provider、credential、Workflow 显式图或 durable backend。
package hostkit

import (
	"context"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

// Config 绑定宿主 Factory 与根 Client 执行画像。
//
// New 成功后，Factory 的 Run 构造和 Shutdown 访问权转移给返回的
// Client；调用方不应再直接使用该 Factory。
type Config struct {
	Factory Factory
	Profile agentx.ExecutionProfile
}

// Factory 按 Run 构造已解析的 portable assembly，并持有宿主资源生命周期。
//
// BuildRun 只应组装 identity、RoundExecutor 和已解析 policy ports，
// 不应重新实现 portable round ordering。Shutdown 必须有界、幂等，并允许
// 前一次调用因 context 到期后继续收敛。
type Factory interface {
	BuildRun(context.Context, execution.Request) (RunConfig, error)
	Shutdown(context.Context) error
	ClassifyError(error) agentx.ErrorCode
}

// RunConfig 是 Factory 为一次 Run 提供的完整 portable assembly 输入。
//
// RunID 和 SessionID 由宿主产生或保留；hostkit 不伪造 identity。
// Assembly 直接复用 canonical toolloop 合同。
type RunConfig struct {
	RunID     string
	SessionID string
	Assembly  toolloop.AssemblyConfig
}

// RunResult 是一次 portable assembly 的完整底层无关结果。
//
// Status 只使用 execution adapter 已支持的 completed、incomplete 和
// failed。Driver、State 和 Termination 供宿主做诊断或兼容投影。
type RunResult struct {
	RunID       string
	SessionID   string
	Status      string
	Reply       string
	Driver      toolloop.Result
	State       toolloop.RoundState
	Termination *toolloop.TerminationSignal
}

// Execute 构造并执行一次 portable tool-loop assembly。
//
// context 和 host error identity 不被包装。执行错误发生时，返回值仍
// 保留已产生的 identity、portable state 和 reply。
func Execute(ctx context.Context, config RunConfig) (RunResult, error) {
	result := RunResult{
		RunID:     config.RunID,
		SessionID: config.SessionID,
		Status:    "failed",
	}
	if ctx == nil {
		return result, fmt.Errorf("agentx host kit: nil run context")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	assembly, err := toolloop.NewAssembly(config.Assembly)
	if err != nil {
		return result, err
	}
	assembled, err := assembly.Run(ctx)
	result.Driver = assembled.Driver
	result.State = assembled.State
	result.Termination = assembled.Termination
	result.Reply = assembled.State.FinalReply
	if err != nil {
		return result, err
	}
	switch assembled.Driver.Kind {
	case toolloop.OutcomeCompleted:
		result.Status = "completed"
	case toolloop.OutcomeTerminated, toolloop.OutcomeMaxRounds:
		result.Status = "incomplete"
	default:
		return result, fmt.Errorf(
			"agentx host kit: unsupported driver outcome %q",
			assembled.Driver.Kind,
		)
	}
	return result, nil
}

// New 构造一个不依赖特定 Runner 的根 AgentX Client。
//
// New 不执行模型或工具请求。Factory 只有在 New 成功后才转移给
// Client；失败时仍由调用方持有。
func New(config Config) (*agentx.Client, error) {
	if config.Factory == nil {
		return nil, fmt.Errorf("agentx host kit: factory is required")
	}
	host := &runtimeHost{factory: config.Factory}
	adapter, err := execution.New(host)
	if err != nil {
		return nil, err
	}
	return agentx.New(agentx.Config{
		Adapter: adapter,
		Profile: config.Profile,
	})
}

type runtimeHost struct {
	factory Factory
}

var _ execution.Host = (*runtimeHost)(nil)

func (host *runtimeHost) Run(ctx context.Context, request execution.Request) (*execution.Result, error) {
	if host == nil || host.factory == nil {
		return nil, fmt.Errorf("agentx host kit: runtime is required")
	}
	config, buildErr := host.factory.BuildRun(ctx, request)
	if buildErr != nil {
		return &execution.Result{
			RunID:     config.RunID,
			SessionID: config.SessionID,
			Status:    "failed",
		}, buildErr
	}
	result, runErr := Execute(ctx, config)
	return &execution.Result{
		RunID:     result.RunID,
		SessionID: result.SessionID,
		Status:    result.Status,
		Reply:     result.Reply,
	}, runErr
}

func (host *runtimeHost) Shutdown(ctx context.Context) error {
	if host == nil || host.factory == nil {
		return fmt.Errorf("agentx host kit: runtime is required")
	}
	return host.factory.Shutdown(ctx)
}

func (host *runtimeHost) ClassifyError(err error) agentx.ErrorCode {
	if host == nil || host.factory == nil {
		return agentx.CodeExecutionFailed
	}
	return host.factory.ClassifyError(err)
}
