# runtime/runstore API

导入路径：

```go
import runstore "github.com/wsnacj/agentx-go/runtime/runstore"
```

成熟度：**Experimental / private validation**。

`runtime/runstore` 是 AgentX Run 数据平面的 portable implementation owner。它
定义 Run、NodeExecution、Event 存储合同，提供并发安全的进程内参考实现，
并把已存储的 NodeExecution 转换为 canonical workflow node projection。

本 package 不提供数据库、文件、远程服务、迁移、事务、租约或生产 durable
backend。`MemoryStore` 的数据随进程结束丢失，只适合测试、示例和明确接受该
生命周期语义的宿主。

## 数据合同

### Run

```go
type Run struct {
    RunID       string `json:"run_id"`
    CaseID      string `json:"case_id,omitempty"`
    WorkflowID  string `json:"workflow_id,omitempty"`
    WorkflowVer string `json:"workflow_ver,omitempty"`
    Status      string `json:"status,omitempty"`
    Attempt     int    `json:"attempt,omitempty"`
    ParentRunID string `json:"parent_run_id,omitempty"`
    RootRunID   string `json:"root_run_id,omitempty"`
    ContractID  string `json:"contract_id,omitempty"`
    StartedAt   int64  `json:"started_at,omitempty"`
    FinishedAt  int64  `json:"finished_at,omitempty"`
    Summary     string `json:"summary,omitempty"`
}
```

`RunID` 是必填 identity。写入时只 trim `RunID`；其他 string 字段原样保留。
负数 `Attempt` 归零。

### NodeExecution

```go
type NodeExecution struct {
    NodeExecID                string `json:"node_exec_id"`
    RunID                     string `json:"run_id"`
    BranchID                  string `json:"branch_id,omitempty"`
    ParentNodeExecID          string `json:"parent_node_exec_id,omitempty"`
    NodeID                    string `json:"node_id,omitempty"`
    Kind                      string `json:"kind,omitempty"`
    Status                    string `json:"status,omitempty"`
    Attempt                   int    `json:"attempt,omitempty"`
    InputStateRef             string `json:"input_state_ref,omitempty"`
    OutputStateRef            string `json:"output_state_ref,omitempty"`
    ExecutionContractID       string `json:"execution_contract_id,omitempty"`
    ExecutionContractDiffJSON string `json:"execution_contract_diff_json,omitempty"`
    TerminationJSON           string `json:"termination_json,omitempty"`
    DelegatedExecutionJSON    string `json:"delegated_execution_json,omitempty"`
    StartedAt                 int64  `json:"started_at,omitempty"`
    FinishedAt                int64  `json:"finished_at,omitempty"`
}
```

`NodeExecID` 和 `RunID` 必填并在写入时 trim，其他字段原样保留，负数
`Attempt` 归零。以下方法读取 opaque JSON 字段；无效或空 JSON 不返回错误，
而是返回 nil/空 projection：

```go
func (NodeExecution) Projection() *NodeExecutionProjection
func (NodeExecution) ExecutionContractDiff() []string
func (NodeExecution) TerminationProjection() *NodeTerminationProjection
func (NodeExecution) DelegatedExecutionProjection() *NodeDelegatedExecutionProjection
```

### Event

```go
type Event struct {
    EventID     string `json:"event_id"`
    RunID       string `json:"run_id"`
    BranchID    string `json:"branch_id,omitempty"`
    NodeExecID  string `json:"node_exec_id,omitempty"`
    Name        string `json:"name,omitempty"`
    PayloadJSON string `json:"payload_json,omitempty"`
    CreatedAt   int64  `json:"created_at,omitempty"`
}
```

`EventID` 和 `RunID` 必填并在写入时 trim，其他字段原样保留。

## Store port

```go
type Store interface {
    CreateRun(context.Context, Run) error
    UpdateRun(context.Context, Run) error
    GetRun(context.Context, string) (Run, error)

    AppendEvent(context.Context, Event) error
    ListEvents(context.Context, string, int) ([]Event, error)

    UpsertNodeExecution(context.Context, NodeExecution) error
    ListNodeExecutions(context.Context, string) ([]NodeExecution, error)
}
```

该 port 只定义最小 Run 数据面，不规定具体 backend、事务或 durable guarantee。
宿主可以实现数据库或远程 adapter，但 adapter 必须保持本合同的错误 identity、
字段和排序语义。

## MemoryStore

```go
func NewMemoryStore() *MemoryStore
```

`MemoryStore` 实现 `Store`，并用 `sync.RWMutex` 保护所有内部 map，可由多个
goroutine 并发调用。它的行为如下：

- `CreateRun` 要求唯一 `RunID`；重复时返回 `ErrAlreadyExists`；
- `UpdateRun` 和 `GetRun` 要求 Run 已存在；不存在时返回 `ErrNotFound`；
- `AppendEvent` 要求 Run 已存在，且 `EventID` 在整个 store 中全局唯一；
- `UpsertNodeExecution` 要求 Run 已存在，并按 Run 内的 `NodeExecID` 覆盖写入；
- `ListEvents` 按 `CreatedAt` 升序、再按 `EventID` 升序；`limit > 0` 返回
  排序后前 N 项，`limit <= 0` 不限；
- `ListNodeExecutions` 按 `StartedAt` 升序、再按 `NodeExecID` 升序；
- 查询尚无 event/node 的合法或未知 RunID 都返回空 slice 和 nil error；
- 当前实现保留既有 nil receiver 合同：nil store 的写方法返回 nil，
  `GetRun` 返回 `ErrNotFound`，两个 List 方法返回 nil slice 和 nil error；
- 方法当前不读取 context 的 cancellation/deadline；context 参数用于与可替换
  backend 的统一 port。调用方不能把 MemoryStore 当作 cancellation boundary。

## 错误合同

```go
var (
    ErrNotFound      = errors.New("agentx/runstore: not found")
    ErrAlreadyExists = errors.New("agentx/runstore: already exists")
)
```

带 identity 的错误会用 `%w` 包装，调用方应使用 `errors.Is`。缺少必填 ID 的
validation error 不匹配上述 sentinel，并保留以下 display text：

```text
agentx/runstore: run id is required
agentx/runstore: event id and run id are required
agentx/runstore: node exec id and run id are required
```

## Node projection aliases

下列名称是 `runtime/workflow/nodeexec` canonical 合同的 type alias，不复制
Workflow execution evidence：

```go
type NodeExecutionProjection = nodeexec.NodeExecutionProjection
type NodeTerminationProjection = nodeexec.Termination
type NodeDelegatedExecutionProjection = nodeexec.DelegatedExecution
type NodeDelegatedRoundProjection = nodeexec.DelegatedRound
```

因此它们与 nodeexec 对应类型具有相同的类型 identity 和 JSON shape。

### JSON 与 clone helpers

```go
func NodeExecutionContractDiffFromJSON(string) []string
func NodeTerminationProjectionFromJSON(string) *NodeTerminationProjection
func NodeDelegatedExecutionProjectionFromJSON(string) *NodeDelegatedExecutionProjection
func CloneNodeExecutionProjection(*NodeExecutionProjection) *NodeExecutionProjection
func CloneNodeExecutionProjections([]NodeExecutionProjection) []NodeExecutionProjection
func CloneNodeTerminationProjection(*NodeTerminationProjection) *NodeTerminationProjection
func CloneNodeDelegatedExecutionProjection(*NodeDelegatedExecutionProjection) *NodeDelegatedExecutionProjection
```

Parser 对空白、无效 JSON 和语义为空的 projection 返回 nil，不把 parse failure
提升为执行错误。非空 string 字段保留原始空白。Clone helper 会复制 slice、
递归 child projection 和嵌套 evidence，返回值可独立修改；语义为空的 delegated
execution clone 归一为 nil。

### selection 与 tree helpers

```go
func SelectTerminalNodeExecutionProjection([]NodeExecutionProjection) *NodeExecutionProjection
func SelectChildNodeExecutionProjections([]NodeExecutionProjection, string) []NodeExecutionProjection
func AttachChildNodeExecutionProjections([]NodeExecutionProjection) []NodeExecutionProjection
func FindNodeExecutionProjection([]NodeExecutionProjection, string) *NodeExecutionProjection
func NodeDelegatedExecutionProjectionFromChildNodeExecutions([]NodeExecutionProjection) *NodeDelegatedExecutionProjection
func NodeDelegatedExecutionLastStopReason(*NodeDelegatedExecutionProjection) string
```

`SelectTerminalNodeExecutionProjection` 在存在 top-level node 时只比较 top-level
候选，依次选择最大的 `FinishedAt`、`StartedAt`、`Attempt`、`NodeExecID`，
并返回深拷贝。

child selection/attachment 使用未经 trim 的 `ParentNodeExecID == NodeExecID`
精确匹配。child 顺序依次使用 round、`StartedAt`、`FinishedAt` 和
`NodeExecID`。Round 优先读取 delegated round；否则只解析 `NodeID` 中最后一个
`round:` 后紧随的十进制数字。`FindNodeExecutionProjection` 同样使用未经 trim
的 exact ID，并返回深拷贝。

delegated aggregation 会 flatten child tree，按 round/timestamp/ID 排序，保留
raw driver、outcome 和 stop reason，并汇总 round count 与 tool calls。

## 外部调用示例

```go
store := runstore.NewMemoryStore()

if err := store.CreateRun(ctx, runstore.Run{
    RunID:     "run-1",
    Status:    "running",
    StartedAt: time.Now().UnixMilli(),
}); err != nil {
    return err
}

if err := store.AppendEvent(ctx, runstore.Event{
    EventID:   "event-1",
    RunID:     "run-1",
    Name:      "run.start",
    CreatedAt: time.Now().UnixMilli(),
}); err != nil {
    return err
}

events, err := store.ListEvents(ctx, "run-1", 100)
```

生产宿主应把自己的 durable backend 适配成 `Store`，而不是依赖 MemoryStore
获得进程重启后的恢复能力。

## 非目标

- 不提供 SQL/SQLite/PostgreSQL、文件或远程 RunStore backend；
- 不提供 schema migration、事务、租约、锁、resume 或 scheduler；
- 不拥有 Workflow execution、journal 写入顺序或 retry policy；
- 不解释 status、event name、opaque JSON 或业务结果；
- 不依赖 HS、Runner、Scene、provider、credential 或具体工具；
- 不构成 Public、Beta、Stable、生产持久化或正式发行声明。
