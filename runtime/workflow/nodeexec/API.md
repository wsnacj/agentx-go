# runtime/workflow/nodeexec API

导入路径：

```go
import nodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
```

成熟度：**Experimental / private validation**。

该 package 是 Workflow 单节点执行协调的 portable implementation owner。它
负责：

- 可选 node context binding；
- `OutcomeExecutor > NodeExecutor > BasicExecutor` 的固定能力优先级；
- 每个 request 只调用一个 executor 一次；
- basic/node string output 与 rich `Outcome` 的统一结果。

它不提供 lowering、工具 registry、具体 LLM/tool executor、RunStore、retry、
provider、credential 或 Scene policy。

## Request

```go
type Call struct {
    Name      string `json:"name,omitempty"`
    Arguments string `json:"arguments,omitempty"`
}

type Request struct {
    NodeExecutionID string                 `json:"node_execution_id,omitempty"`
    NodeID          string                 `json:"node_id,omitempty"`
    Kind            workflow.NodeKind      `json:"kind,omitempty"`
    ExecutionMode   workflow.ExecutionMode `json:"execution_mode,omitempty"`
    OriginalConfig  map[string]any         `json:"original_config,omitempty"`
    Spec            workflow.NodeSpec      `json:"spec,omitempty"`
    Call            Call                   `json:"call"`
}
```

`Kind`、`ExecutionMode` 和 `Spec` 使用 canonical Workflow 数据合同。
`OriginalConfig` 是 opaque node-local 数据；Coordinator 不读取、复制、trim
或修改 map。tool/model/task mapping 与配置验证属于 host adapter。

## Outcome

```go
type Outcome struct {
    Output                string
    FinalStatus           string
    StopReason            string
    ExecutionContractID   string
    ExecutionContractDiff []string
    Termination           *Termination
    DelegatedExecution    *DelegatedExecution
    ChildNodeExecutions   []ChildNodeExecutionProjection
}
```

`Termination`、`DelegatedExecution` 和递归
`ChildNodeExecutionProjection` 是 portable execution evidence。字段 JSON
名称与既有 AgentX projection shape 对齐，包括 `node_exec_id`、
`parent_node_exec_id`、`termination`、`delegated_execution` 和
`child_node_executions`。

Coordinator 不解释 `FinalStatus`、`StopReason` 或 extension 字段，也不做
status normalization、error projection 或 durable serialization。调用方
继续拥有这些 policy。

## Executor ports

```go
type BasicExecutor interface {
    Execute(context.Context, Call) (string, error)
}

type NodeExecutor interface {
    ExecuteNode(context.Context, Request) (string, error)
}

type OutcomeExecutor interface {
    ExecuteNodeWithOutcome(
        context.Context,
        Request,
    ) (Outcome, error)
}
```

三个 port 都是 substrate-neutral。具体实现可适配 LLM、tool registry、
remote worker 或 test double，但相关类型不能进入本 package。

## Construction

```go
type Dependencies struct {
    Basic       BasicExecutor
    Node        NodeExecutor
    Outcome     OutcomeExecutor
    BindContext func(
        context.Context,
        string,
        string,
    ) context.Context
}

func New(Dependencies) *Coordinator
```

`BindContext` 的两个 string 参数依次是 `NodeExecutionID` 和 `NodeID`。binder
为空时不调用；binder 返回 nil 时继续使用原 context。

Coordinator 构造后不修改 dependencies，能否并发使用取决于调用方提供的
executor 和 binder。package 不承诺为非并发安全 dependency 增加同步。

## Execute

```go
func (*Coordinator) Execute(
    context.Context,
    Request,
) (Outcome, error)
```

执行顺序：

1. 调用可选 `BindContext`；
2. 若 `OutcomeExecutor != nil`，只调用
   `ExecuteNodeWithOutcome`；
3. 否则若 `NodeExecutor != nil`，只调用 `ExecuteNode`，将 output 放入
   `Outcome.Output`；
4. 否则调用 `BasicExecutor.Execute`，同样统一为 `Outcome.Output`。

executor 返回 output/outcome 与 error 时，两者同时保留。Port error 原样返回，
不 wrap，因此 `errors.Is/As` identity 保持。Coordinator 不根据 context
状态覆盖 executor 结果；cancellation/deadline 通过同一个 context 传给
binder/executor，由 dependency 遵守。

`BasicExecutor` 在没有 outcome/node capability 时必须存在。缺失时 fail
closed：

```text
workflow nodeexec: basic executor is required
```

nil Coordinator 同样 fail closed。正常 host construction 应在调用前提供
basic adapter。

## 示例

```go
coordinator := nodeexec.New(nodeexec.Dependencies{
    Basic: basicExecutor,
    Node: nodeAwareExecutor,
    Outcome: outcomeExecutor,
})

outcome, err := coordinator.Execute(ctx, nodeexec.Request{
    NodeExecutionID: "nodeexec-1",
    NodeID:           "collect",
    Kind:             workflow.NodeTool,
    Call: nodeexec.Call{
        Name:      "public_source",
        Arguments: `{"query":"risk"}`,
    },
})
```

当三个 capability 都存在时，只调用 `OutcomeExecutor`。

## 非目标

- 不提供 Workflow validation、lowering、transition 或完整执行循环；
- 不注册或选择具体 tool/model/provider；
- 不提供 retry、resume、replanning 或 scheduler；
- 不持久化 run/node/event，不拥有 RunStore backend；
- 不定义 Scene、安全、credential 或业务 policy；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
