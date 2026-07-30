# runtime/telemetry API

导入路径：

```go
import telemetry "github.com/wsnacj/agentx-go/runtime/telemetry"
```

成熟度：**Experimental / private validation**。

该 package 定义 AgentX Runtime 的 observation event、tool/semantic run
projection、stored-event replay、summary 与私有 JSONL sink。它不拥有
Runner、RunStore、workflow、provider、网络 exporter 或发布策略。

## 基础事件

```go
type Level string

const (
    LevelDebug Level = "debug"
    LevelInfo  Level = "info"
    LevelWarn  Level = "warn"
    LevelError Level = "error"

    EventSchemaV1 = "v1"
)

type Event struct {
    SchemaVersion string         `json:"schema_version"`
    Timestamp     time.Time      `json:"timestamp"`
    Component     string         `json:"component"`
    Name          string         `json:"name"`
    Level         Level          `json:"level"`
    SessionID     string         `json:"session_id,omitempty"`
    Round         int            `json:"round,omitempty"`
    Tool          string         `json:"tool,omitempty"`
    Model         string         `json:"model,omitempty"`
    Status        string         `json:"status,omitempty"`
    Attrs         map[string]any `json:"attrs,omitempty"`
}

type Sink interface {
    Emit(context.Context, Event) error
}
```

`Event` 写入 sink 前会执行既有 normalization：

- schema 统一为 `v1`；
- timestamp 为空时取当前 UTC，非空时转为 UTC；
-文本 identity/status 字段去除首尾空白；
-空 level 默认为 `info`，空 attrs 省略；
-`tool.start` 的 raw arguments 字段会被删除，避免把原始工具参数写入
  observation。

本 normalization 是兼容行为，不是输入验证 API。调用方仍应只放入可展示、
可持久化的受控字段。

## Sink

```go
type MultiSink struct {
    Sinks []Sink
}

func (MultiSink) Emit(context.Context, Event) error

func NewJSONLSink(path string) (*JSONLSink, error)
func (*JSONLSink) Emit(context.Context, Event) error

func NewToolEventJSONLSink(path string) (*ToolEventJSONLSink, error)
func (*ToolEventJSONLSink) Emit(context.Context, Event) error

func NewSemanticRunEventJSONLSink(path string) (*SemanticRunEventJSONLSink, error)
func (*SemanticRunEventJSONLSink) Emit(context.Context, Event) error
```

`MultiSink` 按声明顺序调用非 nil sink，并将错误文本用 `"; "` 聚合；它不提供
typed aggregate error。三个 nil sink pointer 的 `Emit` 均为 no-op。

JSONL sink：

- 构造时把路径转为绝对路径；
- 缺失父目录以 `0700` 创建；
- event 文件以 `0600` 创建，并修复已有普通文件权限；
- 拒绝 symlink、非普通文件、非真实父目录与打开期间发生的文件 identity
  替换；
- 每条记录为一行 JSON；
- 通过进程内 mutex 串行同一 sink 的写入。

本 package 不提供 rotation、fsync、跨进程锁、remote exporter 或 delivery
guarantee。需要这些能力时应由 host adapter 显式实现。

## Tool event projection

```go
func ProjectToolEvents(Event) []ToolEvent
func NormalizeToolEventKind(string) string
func ReplayToolEventsFromStoredRecords([]StoredRawEventRecord) (
    []ToolEvent,
    ToolEventProjectionTrace,
)
func SummarizeToolEvents([]ToolEvent) ToolEventSummary
func ToolEventSurfaceForTool(string) string
```

tool projection 使用版本化 schema：

```go
const ToolEventSchemaV1 = "tool_event_v1"
```

它把受支持的 `tool.start`、approval/runtime-decision、repair 与
`tool.finish` observation 投影为 typed event。一次 `tool.finish` 可按固定顺序
产生 retry、provider fallback、result middleware 与最终 completed/failed
多个 event。

`ToolEvent` 及其 projection 类型保留受控 execution-contract、repair、
soft-rejection、result-middleware、output-schema 与 provider-fallback
字段。字段与 JSON tag 是当前 wire compatibility contract；未知 kind
normalize 为空字符串。

summary 提供：

- kind/tool/surface 计数；
- failed/retried/provider fallback；
- runtime decision 与 soft rejection；
- result middleware 与 output-schema drift。

surface 仅为当前稳定分类字符串：`retrieval`、`browser`、`pdf`、`exec`、
`other`。

## Semantic run event projection

```go
func ProjectSemanticRunEvents(Event) []SemanticRunEvent
func NormalizeSemanticRunEventKind(string) string
func SemanticRunEventProjectableSourceEvents() []string
func IsSemanticRunEventProjectableSourceEvent(string) bool
func ReplaySemanticRunEventsFromStoredRecords([]StoredRawEventRecord) (
    []SemanticRunEvent,
    SemanticRunEventProjectionTrace,
)
func SummarizeSemanticRunEvents([]SemanticRunEvent) SemanticRunEventSummary
```

semantic projection 使用版本化 schema：

```go
const SemanticRunEventSchemaV1 = "semantic_run_event_v1"
```

当前 typed kind 为：

- `run.interrupted`
- `run.resumed`
- `approval.requested`
- `approval.resolved`

source allowlist 由 `SemanticRunEventProjectableSourceEvents` 返回 defensive
copy；调用方修改返回 slice 不会改变 package 内部状态。

projection 保留 run/session/branch/node identity、checkpoint、termination 与
approval 的受控字段。它不执行审批、不恢复 Run，也不替代 execution 或
workflow owner。

## Stored event replay

```go
type StoredRawEventRecord struct {
    EventID     string `json:"event_id,omitempty"`
    RunID       string `json:"run_id,omitempty"`
    BranchID    string `json:"branch_id,omitempty"`
    NodeExecID  string `json:"node_exec_id,omitempty"`
    Name        string `json:"name,omitempty"`
    PayloadJSON string `json:"payload_json,omitempty"`
    CreatedAt   int64  `json:"created_at,omitempty"`
}
```

replay 只解析调用方已经读出的 record：

- 不连接 RunStore 或文件；
- payload timestamp 为空时使用 `CreatedAt` 毫秒时间；
- 将 record identity 投影到 typed event；
- invalid JSON 记录为 `invalid_payload_json`，并进入 trace 计数；
- 不修改原 record，不执行 retry 或恢复。

trace 的 source identity 只有 `runstore_events` 与 `telemetry_jsonl`。这两个值
描述输入来源，不意味着本 package 拥有对应存储实现。

## 并发、错误与安全边界

- `MultiSink` 和 JSONL sink 可被并发调用；Event、projection、summary 为值/
  map/slice 数据，调用方负责其自身共享内存同步；
- projection/summary 不发起网络、provider、credential 或生产副作用；
- JSONL sink 的 error 可能包含本地路径，host 在返回给最终用户前必须经过
  display-safe projection；
- package 不承诺跨进程原子写、磁盘持久性或日志保留策略；
- schema 常量的 `V1` 表示该 observation wire schema，不等于 AgentX 产品
  Stable v1。

## 非目标

当前 package 不提供：

- `agentxruntime.New`、Client、Run、Shutdown 或 Runner；
- metrics/tracing backend、OTLP、HTTP 或云服务 exporter；
- RunStore、artifact storage、workflow/objective lifecycle；
- authorization、credential redaction policy 或业务审计审批；
- Public、Beta、Stable 或 production-ready 声明。
