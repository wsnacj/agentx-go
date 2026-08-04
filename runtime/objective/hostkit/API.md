# `runtime/objective/hostkit` 中文 API Reference

状态：**v0.1.0 Developer Preview candidate**。本包进入首版9包核心兼容候选面及
签名、文档和consumer门禁，但不是Beta、Stable或生产SLA。

本package是C5 Objective Runtime Loop面向普通Host的推荐construction路径。它真实组合：

```text
Managed Ingress -> Ready Runtime Adapter Request -> Explicit Host Confirmation
-> Exactly-one Host Handler -> Observation Normalization -> Objective Verification
```

## `Handler`

```go
type Handler func(
    context.Context,
    objective.RuntimeAdapterRequest,
) objective.RuntimeAdapterResult
```

Handler由Host拥有。Host负责具体模型、工具、Workflow、HTTP、数据库或其它backend，并用
canonical result builder返回structured observations和evidence。Host Kit不读取credential，
不推断授权，也不重试或后台执行Handler。

## `Config`和`New`

```go
type Config struct {
    Handler  Handler
    Handlers map[objective.DisplaySafeRef]Handler
}

func New(Config) (*Runtime, error)
```

`Handler`是默认handler；`Handlers`按normalized adapter ref选择。至少提供一个非nil
handler，否则`New`返回`agentx objective host kit: handler is required`。非法map key会
返回`agentx objective host kit: handler ref must be display-safe`。

`New`只复制handler map并验证refs，不运行Objective、不访问网络或存储。`Runtime`自身
不可变；并发安全取决于Host提供的handlers是否并发安全。

## `RunRequest`

```go
type RunRequest struct {
    Ingress               objective.ManagedObjectiveIngressInput
    DispatchEnabled       bool
    DispatchHostConfirmed bool
    Boundaries            []objective.Boundary
}
```

`Ingress`必须由Host提供goal digest、policy、budget、approval、strategy catalog、adapter
registry、idempotency和display-safe input refs。即使ingress ready，只有两个dispatch布尔值
同时为true时才会调用handler。

## `Run`和`RunResult`

```go
func (*Runtime) Run(context.Context, RunRequest) RunResult
```

结果同时保留side-effect-free ingress projection和dispatch/verification readback，并明确
`ReadyForDispatch`、`DispatchAttempted`、`Completed`、status、failure、missing inputs、
boundaries和next action。nil `Runtime` fail closed；非nil cancellation/deadline原样传给
handler，Host Kit不创建goroutine，也不会后台继续执行。

## `Dispatch`

```go
func Dispatch(DispatchInput) DispatchResult
```

这是低层兼容入口，用于已经拥有ready request的Host。它保持HS既有default-off、explicit
confirmation、handler selection、normalization、verification、状态和边界顺序。普通新项目
优先使用`New`+`Run`。

## 非目标

- 不解析原始goal或提供默认strategy/policy；
- 不拥有credential、authorization、approval、provider或backend；
- 不提供durable scheduler、Resume、subagent或long-task生命周期；
- 不把`runtime/controlcontract`全部升级为公共API。
