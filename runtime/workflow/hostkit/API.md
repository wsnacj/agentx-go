# `runtime/workflow/hostkit` 中文 API Reference

状态：**v0.2.1 Developer Preview candidate**。本包进入9包核心兼容候选面及
签名、文档和consumer门禁，但不是Beta、Stable或生产SLA。

本 package是 Workflow面向普通 Host的标准 construction/access seam。它把已存在
的 canonical lowering、journal、node execution coordination、orchestration和
composition组装为一个不可变 `Runtime`，不复制这些 owner的执行语义。

## `Config`

```go
type Config struct {
    Validator Validator
    Mapper    Mapper

    BasicExecutor   BasicExecutor
    NodeExecutor    NodeExecutor
    OutcomeExecutor OutcomeExecutor

    JournalPort JournalPort

    NewRunID           func() string
    NewEventID         func() string
    NewNodeExecutionID func() string
    NowUnixMilli       func() int64

    BindNodeExecutionContext func(context.Context, string, string) context.Context
    ProjectError             func(error) string
}
```

调用方必须显式提供 `Validator`、`Mapper`、至少一个 executor、Run/Node identity
生成器和 clock。配置 `JournalPort` 时还必须提供 `NewEventID`。package不安装
UUID、系统时钟、默认 validator、provider、存储或产品 policy。

executor优先级复用 canonical `nodeexec` owner：

```text
OutcomeExecutor -> NodeExecutor -> BasicExecutor
```

每个节点只调用一个能力。`BindNodeExecutionContext`可选；返回 nil时保留原
context。`ProjectError`只影响可展示的 node error投影，不替换返回 error identity。

## Portable 类型

为使普通 consumer只导入 `workflow`和 `workflow/hostkit`，本 package为已有
canonical类型提供同一类型身份的名称：

- `MappedCall`；
- `Call`、`NodeRequest`、`NodeOutcome`；
- `JournalRun`、`JournalNodeExecution`、`JournalEvent`、`JournalPort`；
- `Inputs`、`RunInputs`、`Result`。

这些名称不是第二份数据模型，也不做 JSON或字段转换；真实 owner仍分别位于
lowering、nodeexec、journal、orchestration和composition。

## `New`

```go
func New(Config) (*Runtime, error)
```

`New`只验证和组合依赖，不执行 Workflow、不访问网络或存储。缺少依赖时按以下
顺序 fail closed：validator、mapper、executor、Run ID、配置 durable port时的
Event ID、Node Execution ID、clock。

成功后 `Runtime`保留传入的 interface和函数。调用方必须保证这些依赖在 Runtime
存活期间有效；package不接管 concrete资源关闭，因此不提供虚假的 `Shutdown`。

## `Run`

```go
func (*Runtime) Run(
    context.Context,
    workflow.Spec,
    Inputs,
) (Result, error)
```

调用顺序保持为：

```text
Validate -> Map/Lower -> Ensure Run -> Execute Nodes -> Durable Finish
```

实际顺序和 fail-fast语义由 canonical composition及其下游 owner持有。lowering失败
返回零结果；orchestration开始后失败会保留 lowering plan和 partial execution。
validator、mapper、executor、journal和 context cancellation/deadline error均保留
`errors.Is/As` identity。

`Inputs.WorkflowID`为空时回退到 `Spec.ID`；`RunInputs`保存 Run/Case/Branch identity
以及 initial/session/case binding roots。Host Kit不 trim、不生成调用方显式传入的
identity，也不改变 state-transition或 durable write顺序。

## 并发、取消与 durable 边界

`Runtime`本身不可变且不持有单次 Run state，也不创建 goroutine。并发调用是否安全
取决于 Host提供的 Validator、Mapper、executors、JournalPort、identity generator
和 clock是否并发安全。

调用方 context原样传给 journal和 executor；package不替换 cancellation/deadline，
也不后台继续执行。`JournalPort=nil`表示无具体 durable backend，但仍运行同一
journal/composition机制；配置 port后，写入顺序由 canonical journal保证。

## 非目标

- 不进入根 `agentx.Client` mode，也不提供 Objective、Resume或长任务；
- 不提供 tool/model/task mapping、产品 validation policy或 retry默认；
- 不提供 RunStore、filesystem、queue、scheduler或网络 backend；
- 不提供 provider、credential、Scene、HTTP或 CLI；
- 不拥有 Host资源生命周期，不伪造 `Shutdown`；
- 不构成 Public、Beta、Stable或正式发布声明。
