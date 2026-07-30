# runtime/workflow/bindingstate API

导入路径：

```go
import bindingstate "github.com/wsnacj/agentx-go/runtime/workflow/bindingstate"
```

成熟度：**Experimental / private validation**。

该 package 是 Workflow Runtime 的 portable binding/state implementation
owner。它负责把 `workflow.BindingSpec` 应用于 JSON arguments、记录 node
result、将输出投影到内存 state，以及验证 required state slots。

它不提供 lowering、node/tool/model executor、route、retry、cancellation、
RunStore、durable write、replanning、provider、credential 或 Scene policy。
调用方必须在 host 层组合这些能力。

## 构造输入

```go
type Inputs struct {
    InitialState map[string]any
    SessionInput map[string]any
    CaseInput    map[string]any
}

func New(Inputs) *Runtime
```

`New` 会递归复制三个输入 map 及其中的 `map[string]any`/`[]any`。调用方在
构造后修改原 map，不会改变 Runtime 内部状态。nil、empty map 均形成空状态。

## Runtime

```go
type Runtime struct {
    // contains filtered or unexported fields
}

func (*Runtime) State() map[string]any

func (*Runtime) MaterializeArguments(
    nodeID string,
    argumentsJSON string,
    bindings []workflow.BindingSpec,
) (string, error)

func (*Runtime) ApplyNodeOutputs(
    nodeID string,
    bindings []workflow.BindingSpec,
    result NodeResult,
) error

func (*Runtime) ValidateRequiredSlots(
    slots []workflow.StateSlotSpec,
) error
```

`State` 每次返回递归副本；修改 snapshot 不会回写 Runtime。

`Runtime` 不是并发安全容器。一个 workflow execution 应拥有一个实例；多个
goroutine 不得在没有外部同步的情况下并发调用会读写该实例的方法。

## NodeResult

```go
type NodeResult struct {
    // contains filtered or unexported fields
}

func NewNodeResult(
    status string,
    output string,
    errorText string,
) NodeResult
```

`NodeResult` 是 opaque value，必须通过 `NewNodeResult` 构造。`output` 只有在
它是没有首尾空白的合法 JSON 时才会形成 structured result；原始 output
字符串始终保留。该规则避免调用方无意中改变既有 raw-output 行为。

## Input binding

`MaterializeArguments` 只接受空字符串或没有首尾空白的 JSON object。空字符串
等价于空 object。支持的 source：

- `state.<path>`
- `session.input.<path>`
- `case.input.<path>`
- `node.<node_id>.output[.<path>]`
- `node.<node_id>.result[.<path>]`
- `node.<node_id>.status`
- `node.<node_id>.error`

target 必须是 `args.<path>`。path 使用点分隔；纯十进制 segment 表示 array
index。binding 按 slice 顺序应用，随后使用 `encoding/json` 重新编码
arguments，因此 object key 顺序遵循 Go JSON 的确定性编码。

`Optional=true` 只忽略 `value "<path>" not found`；空 source、非法 path、
unsupported source/target 和 scalar dereference 仍返回 error。

## Output binding 与 state transition

`ApplyNodeOutputs` 先记录 node result，再按 slice 顺序写 state。支持的 source：

- `output` / `output.<path>`
- `result` / `result.<path>`
- `status`
- `error`

target 必须是 `state.<path>`。如果后续 binding 失败，已经记录的 node result
和之前成功的 state write 不回滚；这是既有 state-transition/write-order
合同，不应把该方法当成事务。

## Required slot

```go
func ValidateSlotName(string) error
```

state slot 名称已经以 workflow state 为根，不能带 `state.` 前缀，不能为空、
含空 segment 或 segment/整体首尾空白。

`ValidateRequiredSlots` 按 slice 顺序先验证名称，再检查 `Required=true` 的
path 是否存在。value 为 `nil` 但 path 存在时视为已填充。

## Error 合同

当前 error 是普通 Go error，没有新增 typed error code。W4-18 保留既有
`workflow: ...` 文本、binding 顺序和返回时点，供 HS compatibility
differential 使用；调用方不应从未文档化的子串重建业务 policy。

## 示例

```go
runtime := bindingstate.New(bindingstate.Inputs{
    SessionInput: map[string]any{"query": "risk"},
})

args, err := runtime.MaterializeArguments(
    "collect",
    `{}`,
    []workflow.BindingSpec{{
        From: "session.input.query",
        To:   "args.query",
    }},
)
if err != nil {
    return err
}

result := bindingstate.NewNodeResult(
    "completed",
    `{"report":"ready"}`,
    "",
)
if err := runtime.ApplyNodeOutputs(
    "collect",
    []workflow.BindingSpec{{
        From: "result.report",
        To:   "state.report",
    }},
    result,
); err != nil {
    return err
}
```

## 非目标

- 不声明完整 Workflow Runtime、executor 或 durable lifecycle；
- 不拥有具体 tool/model/task config mapping；
- 不执行 status default、route、retry、cancel 或 error projection；
- 不提供 backend、provider、credential 或真实副作用；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
