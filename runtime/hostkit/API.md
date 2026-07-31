# `runtime/hostkit` 中文 API Reference

状态：**Experimental / private validation**。

本 package 把 canonical `runtime/toolloop` assembly、`runtime/execution`
adapter 和根 `agentx.Client` 组合成一个 portable Host Kit。它让新项目
只提供每次 Run 的 concrete round executor 和已解析 policy ports，无需
导入 HS Runner 或重新实现 portable round ordering。

## `Config`

```go
type Config struct {
    Factory Factory
    Profile agentx.ExecutionProfile
}
```

- `Factory` 必填，负责为每次 Run 提供宿主资源和 assembly。
- `Profile` 交由根 `agentx.New` 验证；当前只支持根合同已测量的
  Open Tool Loop 画像。
- `New` 成功后 Factory ownership 转移给 Client；失败时不转移。

## `Factory`

```go
type Factory interface {
    BuildRun(context.Context, execution.Request) (RunConfig, error)
    Shutdown(context.Context) error
    ClassifyError(error) agentx.ErrorCode
}
```

`BuildRun` 按请求返回已解析的 `RunConfig`。它可以选择 concrete
model/tool/provider/backend，但这些类型不会进入 canonical API。

`Shutdown` 必须有界、幂等，并允许前一次调用因 context 到期后
继续收敛。`ClassifyError` 只返回根 AgentX 合同支持的稳定错误码，
不得把 backend 文本当作 display-safe message。

## `RunConfig`

```go
type RunConfig struct {
    RunID     string
    SessionID string
    Assembly  toolloop.AssemblyConfig
}
```

- identity 由 Host 产生或保留；Host Kit 不伪造 RunID/SessionID。
- `Assembly` 直接复用 canonical toolloop 类型，包含 `MaxRounds`、
  `RoundExecutor`、detectors、continuation policy 和 initial state。
- concrete model/tool 操作仍在 `toolloop.RoundExecutor` 实现中。

## `RunResult`

```go
type RunResult struct {
    RunID       string
    SessionID   string
    Status      string
    Reply       string
    Driver      toolloop.Result
    State       toolloop.RoundState
    Termination *toolloop.TerminationSignal
}
```

`Status` 映射规则固定：

- `OutcomeCompleted` → `completed`；
- `OutcomeTerminated` / `OutcomeMaxRounds` → `incomplete`；
- assembly 构造或执行失败 → `failed`。

返回值保留 portable Driver、State 和 Termination；发生执行错误时仍
返回已产生的 identity/state/reply，error identity 不被包装。

## `Execute`

```go
func Execute(context.Context, RunConfig) (RunResult, error)
```

`Execute` 在当前调用内构造并执行一个 `toolloop.Assembly`。它不保存
全局状态，不包装 caller context 或 round error。同一 `RunConfig`
中的 stateful executor/detector 不应被并发重用。

## `New`

```go
func New(Config) (*agentx.Client, error)
```

`New` 只组合 Host Kit、execution adapter 和根 Client，不发送模型请求。
返回后，并发 Run 串行化、取消/deadline、typed error、关闭后调用和
bounded/idempotent `Shutdown(ctx)` 继续由根 Client、execution adapter 和
Factory 的现有合同共同保证。

## 明确 non-goal

本 package 不提供：

- concrete model/tool/provider/credential 或网络 client；
- RunStore、queue、scheduler、telemetry backend 或产品 policy；
- Workflow/Objective/Resume/durable lifecycle 入口；
- Scene registry、HTTP/CLI 或发行配置；
- Public、Beta、Stable 或正式 module 发行承诺。
