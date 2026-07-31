# runtime/workflow/orchestration API

导入路径：

```go
import orchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
```

成熟度：**Experimental / private validation**。

该 package 是已经完成 lowering 的 Workflow run orchestration
implementation owner。它组合 canonical `bindingstate`、`transition`、
`journal` 和 `nodeexec`，负责：

- 创建或恢复 run，并在结果初始化后写入 `workflow.start`；
- 逐节点 materialize input binding；
- 按 durable journal 顺序记录 node start/finish；
- 每个节点通过一个 `NodeExecution` capability 恰好执行一次；
- 投影 outcome、error、state、status 和 stop reason；
- success/failure/incomplete edge routing；
- required state 校验与 run finalization。

它不提供 Workflow validation 或 lowering，不选择具体 tool/model/provider，
不拥有 RunStore backend、retry/resume/replanning、credential 或 Scene policy。

## Construction

host 先把自己的 lowered representation 投影为：

```go
type Plan struct {
    WorkflowID  string
    Version     string
    EntryNode   string
    NodeIDs     []string
    Nodes       map[string]PlannedNode
    Edges       []workflow.EdgeSpec
    StateSchema []workflow.StateSlotSpec
}
```

`PlannedNode` 只包含 canonical `NodeSpec`、portable `nodeexec.Call`、kind、
execution mode 和 opaque original config。不能把 host Runner、HS config、
具体 executor 或 backend 放入 Plan。

运行依赖为：

```go
type Dependencies struct {
    Journal            *journal.Journal
    NodeExecution      NodeExecution
    NewNodeExecutionID func() string
    NowUnixMilli       func() int64
    ProjectError       func(error) string
}
```

`Journal` 和通常由 `nodeexec.Coordinator` 实现的 `NodeExecution` 是既有
canonical mechanism。ID、clock 和 display-safe error projection 由 host
注入，因此 package 不拥有 UUID、时钟或业务错误展示策略。

## Run

```go
func Run(
    context.Context,
    Plan,
    Inputs,
    Dependencies,
) (Result, error)
```

`Run` 同步执行一个单路径、单 pass 的 lowered plan。context 原样传递给
journal port 和 node execution；package 不覆盖 dependency 返回的
cancellation/deadline error。

缺少 `NodeExecution` 或 clock 时分别 fail closed：

```text
workflow orchestration: node execution is required
workflow orchestration: clock is required
```

缺少或返回空 node execution ID 时同样返回稳定错误，不会调用 executor。

executor error 会写入 host 投影后的 display-safe 文本，但返回错误继续使用
`%w` 保留 `errors.Is/As` identity。binding、transition、journal error 维持
既有 first-error/fail-fast 顺序。

durable 顺序固定为：

```text
run ensure
workflow.start
input state snapshot
running node
workflow.node.start
output state snapshot
final node
workflow.node.finish
run update
workflow.finish
```

## Result

`Result` 返回 RunID、最终 node/status/stop reason、逐节点结果、非空 node
output 和隔离的 state snapshot。它不包含 host-only lowered debug
representation；host 可在兼容 wrapper 中附加该信息。

## 并发与生命周期

每次 `Run` 创建独立 binding state 和 transition machine。能否并发调用取决于
调用方提供的 Journal Port、NodeExecution、ID、clock 和 error projector。
package 不为非并发安全 dependency 增加锁，也不提供 background goroutine 或
Shutdown。

## 非目标

- 不构成 Public、Beta、Stable 或 production-ready 声明；
- 不提供完整 Agent Runtime construction；
- 不提供 Workflow validation、lowering、retry、resume 或 durable backend；
- 不注册 tool/model/provider，不处理 credential 或真实网络副作用；
- 不拥有 pack、product default、Scene 或业务错误 policy。
