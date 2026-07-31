# runtime/workflow/lowering API

导入路径：

```go
import lowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
```

成熟度：**Experimental / private validation**。

该 package 是 Workflow lowering plan mechanism 的 canonical implementation
owner，负责：

- 按既有顺序调用 Spec 与 Node validator；
- 为 exact-empty `ExecutionMode` 补 `inline`；
- 按 Node 声明顺序调用一次 host `Mapper`；
- 将 mapper 返回的 argument object编码为 JSON；
- 组装 portable node/edge/state plan；
- 投影为 canonical `orchestration.Plan`。

它不选择具体 tool、model、task、queue、evaluator、provider、credential、
executor 或 backend。调用方必须显式注入 `Validator` 与 `Mapper`。

## Dependencies

```go
type Validator interface {
    ValidateSpec(workflow.Spec) error
    ValidateNode(workflow.NodeSpec) error
}

type Mapper interface {
    MapNode(
        workflow.NodeSpec,
        workflow.ExecutionMode,
    ) (MappedCall, error)
}

type MappedCall struct {
    Name      string
    Arguments map[string]any
}

type Dependencies struct {
    Validator Validator
    Mapper    Mapper
}
```

`Validator` 决定 host admission policy；package 不提供 permissive default。
`Mapper` 只负责把一个已验证 Node 映射成调用名称与 argument object，不取得
lowering loop、JSON marshal 或 orchestration state。

两个依赖都必须非 nil：

```text
workflow lowering: validator is required
workflow lowering: mapper is required
```

## LowerSpec

```go
func LowerSpec(
    spec workflow.Spec,
    dependencies Dependencies,
) (Plan, error)
```

调用顺序固定为：

```text
Validator.ValidateSpec
for node in spec.Nodes:
    Validator.ValidateNode
    Mapper.MapNode
    JSON marshal
```

node error 使用以下 context 和 `%w` wrapping：

```text
workflow: lower node "<node-id>": <cause>
```

因此 validator/mapper error 的 `errors.Is/As` identity 保持。

## LowerNode

```go
func LowerNode(
    node workflow.NodeSpec,
    dependencies Dependencies,
) (Node, error)
```

exact-empty execution mode默认成 `workflow.ExecInline`；其它非空值原样交给
validator/mapper，不 trim、不 canonicalize。

空 argument map编码为 `{}`。其它值使用标准 `encoding/json`；失败文本为：

```text
marshal tool arguments: <cause>
```

## Plan

```go
type Node struct {
    NodeID         string
    Spec           workflow.NodeSpec
    Kind           workflow.NodeKind
    ExecutionMode  workflow.ExecutionMode
    Call           nodeexec.Call
    OriginalConfig map[string]any
}

type Plan struct {
    SpecID      string
    Version     string
    EntryNode   string
    Nodes       map[string]Node
    Edges       []workflow.EdgeSpec
    StateSchema []workflow.StateSlotSpec
}

func (Plan) OrchestrationPlan(workflowID string) orchestration.Plan
```

`OrchestrationPlan` 只做 portable projection，不创建 executor、RunStore、
journal 或 goroutine，也不执行 Workflow。`workflowID` 是 host 选择的运行期
identity，可以与 Spec ID 不同。

`Nodes` 是 map；调用方不得把 `NodeIDs` 的 map iteration 顺序当作稳定合同。
实际 traversal 由 entry/edge transition 决定。

## 并发与生命周期

package 不保存全局状态、不创建 goroutine，也不提供 Shutdown。每次调用创建
独立 Plan。是否能并发调用取决于 host 提供的 Validator/Mapper；package
不会为非并发安全依赖加锁。

## 非目标

- 不提供 HS/Scene mapping/default policy；
- 不固定 `llm_task`、`agent_step`、`tasks_wait` 等工具名称；
- 不生成 evaluator prompt/schema；
- 不选择 concrete executor、RunStore、provider 或 credential；
- 不执行 retry、resume、durable lifecycle 或 Scene side effect；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
