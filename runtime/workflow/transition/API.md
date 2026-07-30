# runtime/workflow/transition API

导入路径：

```go
import transition "github.com/wsnacj/agentx-go/runtime/workflow/transition"
```

成熟度：**Experimental / private validation**。

该 package 是 Workflow Runtime 的 portable traversal、final-status
normalization 和 edge-routing implementation owner。它维护一次执行的当前
node 与 visited state，检查 lowered node 是否存在和是否形成运行时循环，并
按 success/failure/always 规则选择唯一下一条 edge。

它不提供 Workflow validation、lowering、node/tool/model executor、RunStore、
durable write、retry、resume、cancellation、replanning、provider、credential
或 Scene policy。调用方必须在 host 层组合这些能力。

## Plan 与构造

```go
type Plan struct {
    EntryNode string
    NodeIDs   []string
    Edges     []workflow.EdgeSpec
}

type Machine struct {
    // contains filtered or unexported fields
}

func New(Plan) *Machine
```

`New` 会复制 `NodeIDs` 的 identity set 和 `Edges` slice；调用方在构造后修改
原 slice 不会改变 Machine。node/edge ID 使用精确字符串比较，不 trim、不改写
大小写。

`Machine` 对应一次 workflow execution，不是并发安全容器。多个 goroutine
不能在没有外部同步的情况下并发调用同一实例。

## 遍历

```go
func (*Machine) Enter() (nodeID string, err error)
```

host 在执行当前 node 前调用 `Enter`：

- 返回空 node ID 表示没有待执行 node；
- 当前 node 不在 `NodeIDs` 时返回 missing-lowered-node error；
- 当前 node 已经进入过时返回 runtime-cycle error；
- 成功后把当前 node 记录为 visited。

Machine 不读取 `workflow.NodeSpec`，也不执行 node。

## Edge routing

```go
type Trigger string

const (
    TriggerSuccess Trigger = "success"
    TriggerFailure Trigger = "failure"
)

func (*Machine) Advance(Trigger) (nextNodeID string, err error)
```

`Advance` 在 host 完成本次 node 的所有既有副作用后调用。匹配规则：

- edge `From` 必须精确等于当前 node ID；
- `On=""` 等价 `success`；
- `On="always"` 同时匹配 success 与 failure；
- 没有匹配 edge 时返回空字符串并进入 terminal state；
- 多条 edge 同时匹配时返回 error，Machine 不推进；
- 唯一匹配时更新 current node 并返回其 `To`。

该方法不检查目标 node；目标存在性与 cycle 在下一次 `Enter` 检查，因此 host
仍可保持 node finish、snapshot 和 durable write 的原有顺序。

## Final status normalization

```go
func NormalizeFinalStatus(
    finalStatus string,
    failed bool,
) string
```

精确的 `completed`、`failed`、`incomplete` 在 `failed=false` 时原样保留。
其它值回落为 `completed`，不 trim、不 canonicalize。`failed=true` 始终返回
`failed`，表示 host dependency error 优先。

## Error 合同

当前 error 是普通 Go error，没有新增 typed error code。package 保留既有
Workflow inline host 的 error 文本和返回时点：

```text
workflow: detected cycle at node %q during inline execution
workflow: lowered node %q missing
workflow: node %q has multiple outgoing %s edges; inline executor only supports a single path
```

调用方不应从这些文本推导新的业务 policy。

## 示例

```go
machine := transition.New(transition.Plan{
    EntryNode: "collect",
    NodeIDs:   []string{"collect", "report"},
    Edges: []workflow.EdgeSpec{{
        From: "collect",
        To:   "report",
        On:   "success",
    }},
})

for {
    nodeID, err := machine.Enter()
    if err != nil {
        return err
    }
    if nodeID == "" {
        break
    }

    // host executes nodeID and completes its durable writes here.

    next, err := machine.Advance(transition.TriggerSuccess)
    if err != nil {
        return err
    }
    if next == "" {
        break
    }
}
```

## 非目标

- 不声明完整 Workflow Runtime 或 `ExecuteInline`；
- 不拥有 validation、lowering、executor、RunStore 或 durable lifecycle；
- 不实现 retry、resume、cancellation、condition evaluation 或 parallel graph；
- 不包含 tool/model/task mapping、provider、credential 或 Scene policy；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
