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

## `ModelToolRoundAdapter`

`ModelToolRoundAdapter` 是 W5-H 的真实 portable model/tool round
implementation。它复用 `runtime/toolloop.RoundPhaseCoordinator`，固定执行：

```text
RequestModel -> ObserveResponse -> no tools
                                 -> BeforeTools -> host stopped
                                                -> ExecuteTools
```

它不选择 provider、不做授权或持久化，只负责调用顺序、model/tool结果的
defensive copy，以及到 portable round outcome/continuation的投影。

```go
type ModelToolRoundConfig struct {
    RequestModel func(
        context.Context,
        toolloop.RoundExecutionInput,
    ) (ModelResult, error)

    ObserveResponse func(
        context.Context,
        ModelToolRoundExchange,
    ) (string, error)

    BeforeTools func(
        context.Context,
        ModelToolRoundExchange,
    ) (bool, error)

    ExecuteTools func(
        context.Context,
        ModelToolRoundExchange,
    ) (ToolResult, error)
}

func NewModelToolRoundAdapter(
    ModelToolRoundConfig,
) (*ModelToolRoundAdapter, error)
```

- `RequestModel`必填，返回 canonical `components/llm.ChatResponse`；
- `ObserveResponse`可选，默认直接使用 `Response.Content`；
- `BeforeTools`可选，默认允许执行；authorization、approval、budget等规则应由
  Host在这里调用既有 owner，不得复制进 adapter；
- `ExecuteTools`只在模型返回 tool calls时需要；没有 tool calls时不会调用；
- 所有 port error identity原样传播，`ModelToolRoundResult.Phase.LastPhase`
  保留失败阶段。

```go
type ModelResult struct {
    Response  llm.ChatResponse
    Model     string
    Recovered bool
}

type ModelToolRoundExchange struct {
    Input toolloop.RoundExecutionInput
    Model ModelResult
    Reply string
}

type ToolResult struct {
    Runs             []toolloop.RunObservation
    Failures         []toolloop.FailureObservation
    NextChunks       []string
    ForceNoToolCalls bool
}
```

`ModelToolRoundExchange`和所有返回 slice均为防御性副本。provider私有 payload、
recovery plan、RunStore handle或具体 tool executor不得塞入这些合同；Host可在
各函数闭包中保留自己的单轮状态。

### `Execute` 与 `ExecuteRound`

```go
func (*ModelToolRoundAdapter) Execute(
    context.Context,
    toolloop.RoundExecutionInput,
) (ModelToolRoundResult, error)

func (*ModelToolRoundAdapter) ExecuteRound(
    context.Context,
    toolloop.RoundExecutionInput,
) (toolloop.RoundExecutionResult, error)
```

`Execute`面向仍需在执行后保留 answer-contract、recovery、budget、telemetry等
产品策略的 Host；它返回 phase/model/tool事实，产品投影继续由 Host完成。

`ExecuteRound`直接实现 `toolloop.RoundExecutor`，适合普通新项目：

- 无 tool call → `OutcomeCompleted`；
- Host gate停止 → `OutcomeTerminated`；
- tool batch完成 → `OutcomeContinue`，并自动投影 Calls、Runs、Failures、
  NextChunks和 ForceNoToolCalls。

因此 Host Kit调用方可以直接把 adapter放入
`toolloop.CoordinatorConfig.Executor`，不再自行编写 `RoundExecutor`或复制
round phase ordering。

adapter自身无单轮可变状态；并发安全性只取决于调用方提供的函数是否安全。

## `NewModelToolClient`

普通新项目不需要实现 `Factory`、`BuildRun`或 `toolloop.RoundExecutor`：

```go
type ModelToolClientConfig struct {
    Profile   agentx.ExecutionProfile
    MaxRounds int

    BuildRound func(
        context.Context,
        execution.Request,
    ) (ModelToolRoundConfig, error)

    ResolveIdentity func(
        execution.Request,
    ) (runID string, sessionID string)

    InitialState  func(execution.Request) toolloop.RoundState
    Shutdown      func(context.Context) error
    ClassifyError func(error) agentx.ErrorCode
}

func NewModelToolClient(ModelToolClientConfig) (*agentx.Client, error)
```

- `MaxRounds`必须大于0，`BuildRound`必填；
- `BuildRound`只为一次 Run提供具体 model/tool函数，不组装 Assembly；
- `ResolveIdentity`可选；未提供时保留 request SessionID，RunID保持空，不伪造
  identity；
- `InitialState`可选，默认把 request Input作为唯一初始 chunk；
- `Shutdown`可选，默认无资源可关闭；`ClassifyError`默认
  `CodeExecutionFailed`；
- 该便捷路径不安装 loop detector、failure fuse、authorization、persistence或
  产品默认值。需要这些已解析 policy ports的高级 Host继续使用 `Config + Factory`。

相比 W5-G consumer，该路径删除了调用方自定义 Factory三个方法、手写
`BuildRun` Assembly和自定义 `RoundExecutor`，同时保持所有能力显式注入。

## 明确 non-goal

本 package 不提供：

- concrete model/tool/provider/credential 或网络 client；
- RunStore、queue、scheduler、telemetry backend 或产品 policy；
- Workflow/Objective/Resume/durable lifecycle 入口；
- Scene registry、HTTP/CLI 或发行配置；
- Public、Beta、Stable 或正式 module 发行承诺。
