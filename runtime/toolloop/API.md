# runtime/toolloop API

导入路径：

```go
import toolloop "github.com/wsnacj/agentx-go/runtime/toolloop"
```

成熟度：**Experimental / private validation**。

该 package 是 portable Run/Open Tool Loop mechanism owner，负责：

- 从 round 1开始、最多执行 `MaxRounds` 次 host-owned step；
- 对 continue、completed、terminated和 max-rounds outcome做确定性收口；
- 原样传播 step error identity和发生错误的 round；
- 对 tool call/result observation做稳定 signature；
- 按 no-progress→repeat→ping-pong优先级检测循环；
- 检测 immediate successful replay；
- 对同一 normalized tool的连续失败计数，并让 `invalid_args` 最多两次即熔断。
- 对 host round结果做 outcome验证、continuation state更新，并按 failure
  fuse→host continuation policy→loop detector的顺序决定是否继续。

它不拥有 model request、tool resolution/execution、authorization、budget、
retry、answer-contract、queue/status、checkpoint、event、session persistence、
用户可见回复、RunStore、provider、credential或 Scene。

## Runtime

```go
type Config struct {
    MaxRounds int
}

type StepInput struct {
    Round     int
    MaxRounds int
}

type StepResult struct {
    Kind OutcomeKind
}

type Result struct {
    Kind  OutcomeKind
    Round int
}

type Stepper interface {
    Step(context.Context, StepInput) (StepResult, error)
}

func New(Config, Stepper) (*Runtime, error)
func (*Runtime) Run(context.Context) (Result, error)
```

`MaxRounds`必须大于0，`Stepper`必须非 nil。`Run`把调用方 context原样传给
每个 step，不建立第二套 timeout、retry或 cancellation policy。step error不
包装，调用方可以继续使用 `errors.Is/As`；`Result.Round`指出失败或终止发生的
round。

`OutcomeContinue`进入下一轮；`OutcomeCompleted`和`OutcomeTerminated`立即返回；
循环耗尽时返回 `OutcomeMaxRounds`。用户可见 max-rounds回复、checkpoint和
event仍由 host负责。

## Host-Backed Assembly

```go
type AssemblyConfig struct {
    MaxRounds   int
    Coordinator CoordinatorConfig
    Initial     RoundState
}

type AssemblyResult struct {
    Driver      Result
    State       RoundState
    Termination *TerminationSignal
}

func NewAssembly(AssemblyConfig) (*Assembly, error)
func (*Assembly) Run(context.Context) (AssemblyResult, error)
```

`Assembly`把本 package已有的多轮 `Runtime`与`Coordinator`组合成一次逻辑 Run，
并统一返回 driver outcome、最终 portable state与 termination fact。它是真实的
组合 implementation，不创建或复制 model/tool/backend，也不定义产品默认值。

`CoordinatorConfig.Executor`与可选 policy仍由 host注入。构造时先验证
coordinator，再验证 `MaxRounds`；调用方 context和 round error原样传播。
`State`与`Termination`均为防御性副本。`Assembly`持有每次 Run的可变状态，面向
一次逻辑 Run，不应并发调用或跨 Run复用；它不创建 goroutine，也不提供
`Shutdown`。

## Round Coordinator

```go
type RoundState struct {
    Chunks           []string
    ForceNoToolCalls bool
    FinalReply       string
}

type RoundExecutor interface {
    ExecuteRound(context.Context, RoundExecutionInput) (RoundExecutionResult, error)
}

type ContinuationPolicy interface {
    ObserveContinuation(
        context.Context,
        ContinuationObservation,
    ) (code string, stop bool, err error)
}

func NewCoordinator(CoordinatorConfig, RoundState) (*Coordinator, error)
func (*Coordinator) Step(context.Context, StepInput) (StepResult, error)
func (*Coordinator) State() RoundState
```

`Coordinator`本身实现 `Stepper`，可以直接交给 `New(Config, Stepper)`。具体一轮
model request、response observation、tool execution、recovery、budget和
persistence由 host `RoundExecutor`完成；coordinator只接收 portable
`RoundExecutionResult`。

`OutcomeContinue`必须携带 `RoundContinuation`。coordinator会防御性复制
`NextChunks`，只用非空白的 raw reply更新 `FinalReply`，再按以下固定顺序判定：

```text
FailureFuse -> ContinuationPolicy -> LoopDetector
```

`ContinuationPolicy`是可选的 host policy seam，例如产品可在该位置保留 scheduler
queue stall规则。portable `TerminationSignal`只记录 failure、host-policy或 loop
事实；用户可见回复、diagnostics、checkpoint和 event投影仍由 host负责。executor
或 policy error不会包装，`errors.Is/As` identity保持。

## Round Phase Coordinator

```go
type RoundPhaseInput struct {
    Round     int
    MaxRounds int
}

type RoundPhaseExecutor interface {
    Request(context.Context, RoundPhaseInput) (RoundRequestResult, error)
    Observe(context.Context, RoundPhaseInput) (RoundObserveResult, error)
    BeforeAction(context.Context, RoundPhaseInput) (bool, error)
    Act(context.Context, RoundPhaseInput) error
}

func NewRoundPhaseCoordinator(RoundPhaseExecutor) (*RoundPhaseCoordinator, error)
func (*RoundPhaseCoordinator) Execute(
    context.Context,
    RoundPhaseInput,
) (RoundPhaseResult, error)
```

`RoundPhaseCoordinator`拥有以下固定顺序：

```text
Request -> Observe -> no action
                   -> BeforeAction -> host stopped
                                   -> Act -> action completed
```

同一个调用方 context会原样传给每个 phase；phase error不包装，结果中的
`LastPhase`指出失败或最后完成的 phase。`Observe`提供的 raw reply不 trim。
无 action时不会调用 gate或 `Act`，`BeforeAction`返回 false时也不会调用
`Act`。

具体 model request/response、tool call数据、budget检查、authorization、
persistence与产品 policy都由同一个 host executor持有。executor可在四个方法间
保存单轮临时状态，因此每次并发 round必须使用独立实例，除非 host自行同步。

## LoopDetector

```go
type LoopDetectorConfig struct {
    Enabled             bool
    RepeatThreshold     int
    PingPongThreshold   int
    NoProgressThreshold int
}

func NewLoopDetector(LoopDetectorConfig) *LoopDetector
func (*LoopDetector) Observe(
    round int,
    calls []Call,
    runs []RunObservation,
) (LoopSignal, bool)
func (*LoopDetector) ShouldSuppressReplay([]Call) (LoopSignal, bool)
```

call argument是合法 JSON时按 compact JSON签名，否则按小写、空白折叠后的文本
签名；result output使用同样的文本 normalization和 FNV-1a 64-bit digest。
`LoopSignal`只提供机制事实，不决定用户可见回复或产品停止策略。

## FailureFuse

```go
type FailureFuseConfig struct {
    Enabled   bool
    Threshold int
}

func NewFailureFuse(FailureFuseConfig) *FailureFuse
func (*FailureFuse) Observe(
    round int,
    observations []FailureObservation,
) (FailureSignal, bool)
```

host必须先完成 tool name normalization与 error classification，再投影为
`FailureObservation`。任一成功 observation会 reset连续失败；普通错误使用
host提供的 threshold，`invalid_args` 的有效 threshold最多为2。

## 并发与生命周期

`Runtime`自身不保存单次 Run进度，但会调用同一个 `Stepper`；并发调用是否安全
取决于 host Stepper。`LoopDetector`和 `FailureFuse`保存每次 Run的可变历史，
不得跨并发 Run共享；每个 Run应创建新实例。package不创建 goroutine，也不提供
Shutdown。

## 非目标

- 不提供无需 Host 的根 `agentxruntime.New`；
- 不直接接受 `engine.Runner`、HS Config或具体 LLM/tool类型；
- 不定义 max-rounds、loop threshold或 failure threshold的产品默认值；
- 不实现真实网络、filesystem、provider、backend或 Scene side effect；
- 不构成 Public、Beta、Stable或 production-ready声明。
