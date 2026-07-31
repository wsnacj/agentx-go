# runtime/workflow/journal API

导入路径：

```go
import journal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
```

成熟度：**Experimental / private validation**。

该 package 是 Workflow Runtime 的 portable durable journal implementation
owner。它通过调用方提供的最小 `Port`，确定 run/node/event 的 fail-fast
写入顺序、snapshot reference 和 workflow lifecycle event。

它不提供具体 RunStore backend、查询/List API、validation、lowering、node
executor、retry、resume、provider、credential 或 Scene policy。

## Port

```go
type Port interface {
    LoadRun(context.Context, string) (Run, bool, error)
    CreateRun(context.Context, Run) error
    UpdateRun(context.Context, Run) error
    UpsertNodeExecution(context.Context, NodeExecution) error
    AppendEvent(context.Context, Event) error
}
```

`LoadRun` 用独立的 `found` 表达不存在；不存在不是 error。具体 backend 的
not-found error identity 由 host adapter 转换。其它方法的 error 原样返回，
Journal 不重试、不吞错、不改写 error identity。

每次调用都使用调用方传入的原始 `context.Context`。Port 应遵守 cancellation
和 deadline；Journal 不创建 background context。

## Portable records

```go
type Run struct {
    RunID           string
    CaseID          string
    WorkflowID      string
    WorkflowVersion string
    Status          string
    Attempt         int
    ParentRunID     string
    RootRunID       string
    ContractID      string
    StartedAt       int64
    FinishedAt      int64
    Summary         string
}

type NodeExecution struct {
    NodeExecutionID           string
    RunID                     string
    BranchID                  string
    ParentNodeExecutionID     string
    NodeID                    string
    Kind                      string
    Status                    string
    Attempt                   int
    InputStateRef             string
    OutputStateRef            string
    ExecutionContractID       string
    ExecutionContractDiffJSON string
    TerminationJSON           string
    DelegatedExecutionJSON    string
    StartedAt                 int64
    FinishedAt                int64
}

type Event struct {
    EventID         string
    RunID           string
    BranchID        string
    NodeExecutionID string
    Name            string
    PayloadJSON     string
    CreatedAt       int64
}
```

这些类型是 host-port record，不是已承诺的 wire/storage schema。Journal 不
trim、normalize 或重写字符串字段。JSON extension 字段由 host 构造，具体
业务类型不会进入本 package。

## Construction

```go
type Dependencies struct {
    Port         Port
    NewRunID     func() string
    NewEventID   func() string
    NowUnixMilli func() int64
}

func New(Dependencies) *Journal
```

identity 与 clock 由 host 注入，Journal 不选择 UUID 实现或 wall clock。
当 `Port=nil` 时，durable methods 是有界 no-op；`EnsureRun` 仍会保留或生成
run ID。需要生成 ID/时间但相应 dependency 缺失或返回空字符串时，方法
fail closed。

`Journal` 面向一次 host execution 组合，不承诺并发安全。Port、ID source
和 clock 的并发能力由调用方负责。

## Run lifecycle

```go
type EnsureRunRequest struct {
    RunID           string
    CaseID          string
    WorkflowID      string
    WorkflowVersion string
}

func (*Journal) EnsureRun(
    context.Context,
    EnsureRunRequest,
) (string, error)
```

规则：

- `RunID=""` 时调用 `NewRunID`；
- `LoadRun` found 时，Case/Workflow/Version 只使用非空 request 覆盖；
- status 设置为 `running`；
- Attempt 小于等于零时设置为 1；
- RootRunID 为空时设置为当前 RunID；
- StartedAt 小于等于零时调用 `NowUnixMilli`；
- found run 调用 `UpdateRun`，missing run 调用 `CreateRun`；
- persistence error 后不执行后续操作。

run persistence 与 start event 有意分成两个方法：

```go
type StartRunEventRequest struct {
    RunID      string
    BranchID   string
    WorkflowID string
    EntryNode  string
}

func (*Journal) AppendRunStart(
    context.Context,
    StartRunEventRequest,
) error
```

这允许 host 在两者之间初始化可观察执行结果：run persistence 失败可以返回空
result，`workflow.start` event 失败可以返回已经包含 run identity/state 的
result。`AppendRunStart` 使用注入的 event ID 和 clock。

run finish：

```go
type FinishRunRequest struct {
    RunID           string
    WorkflowID      string
    WorkflowVersion string
    Status          string
    FinishedAt      int64
    ErrorText       string
}

func (*Journal) FinishRun(
    context.Context,
    FinishRunRequest,
) error
```

顺序固定为：

1. `LoadRun`；
2. 补齐 RunID/Workflow/Version/Attempt/RootRunID，写入 status、finishedAt
   和 summary；
3. `UpdateRun`；
4. `AppendEvent(workflow.finish)`。

update 失败时不会写 finish event。

## Node lifecycle

```go
type StartNodeRequest struct {
    Node         NodeExecution
    State        map[string]any
    EventPayload map[string]any
}

func (*Journal) StartNode(
    context.Context,
    StartNodeRequest,
) (inputStateRef string, err error)
```

顺序固定为：

1. 将 `{node_id, state}` 编码为 `workflow.node.state.input`；
2. snapshot event ID 成为 `InputStateRef`；
3. `UpsertNodeExecution`；
4. 编码并追加 `workflow.node.start`。

node finish：

```go
type FinishNodeRequest struct {
    Node         NodeExecution
    State        map[string]any
    EventPayload map[string]any
}

func (*Journal) FinishNode(
    context.Context,
    FinishNodeRequest,
) (outputStateRef string, err error)
```

顺序固定为：

1. 写 `workflow.node.state.output`；
2. snapshot event ID 成为 `OutputStateRef`；
3. `UpsertNodeExecution`；
4. 追加 `workflow.node.finish`。

任一步失败都会立即停止；Journal 不回滚已经成功的 durable write。这是
append/upsert journal，不是事务。

## JSON 与 error

state 和 event payload 使用 Go `encoding/json` 编码。map key 使用 Go JSON
确定性顺序；nil/empty payload 保留既有 empty `PayloadJSON` 行为。marshal
error 在调用 Port 前返回。

dependency validation error 使用 `workflow journal: ...` 前缀。Port error
不 wrap，因此 `errors.Is/As` identity 保持。该 package不新增业务 error
code，也不从 error 文本推导 retry policy。

## 非目标

- 不提供 RunStore backend、database schema 或查询/List API；
- 不执行 Workflow validation、lowering 或 node execution；
- 不拥有 retry、resume、replanning、condition 或 parallel execution；
- 不定义 provider、credential、安全或 Scene policy；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
