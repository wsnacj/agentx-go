# `runtime/protocol` 中文 API Reference

## 当前定位

导入路径：

```go
import agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
```

本 package 定义 AgentX Runtime 的版本化 wire/schema、normalization 和
validation。它从 HS `core/agentx/runtime/protocol` 原样迁移，是该基础协议的
唯一 source authority，目前处于 **private validation / Experimental**。

本 package：

- 不执行 Agent、Workflow、Tool 或 Model；
- 不决定 auth、approval、sandbox、budget、timeout、retry 或 replay；
- 不持久化 Run/Event/Artifact，也不拥有 durable source-of-truth；
- 不从日志、transcript 或业务 payload 反推运行事实；
- 不包含 projection、provider、credential、网络或生产副作用；
- production 代码只依赖 Go 标准库。

`execution` 仍是 runtime control source；`runstore`、artifact、telemetry 等仍
拥有原始或持久化事实。HS `runtime/protocol/projection` 可以把这些事实投影成
本 package 的结构，但 projection 不是本 module 的一部分。

## Schema identity

| Constant | Value | 用途 |
| --- | --- | --- |
| `RuntimeSchemaV1` | `agentx.runtime.v1` | 统一 Runtime envelope identity |
| `RunEventSchemaV1` | `agentx.run_event.v1` | Run/turn/tool/handoff 事件 |
| `TraceSpanSchemaV1` | `agentx.trace_span.v1` | reserved trace/span 结构 |
| `ToolExecutionPlanSchemaV1` | `agentx.tool_execution_plan.v1` | Tool 执行前计划 |
| `HandoffSchemaV1` | `agentx.handoff.v1` | Agent/task/handoff 记录 |
| `SandboxManifestSchemaV1` | `agentx.sandbox_manifest.v1` | Sandbox/workspace 描述 |
| `ArtifactVersionSchemaV1` | `agentx.artifact_version.v1` | Artifact 版本与 payload |
| `ArtifactLinkSchemaV1` | `agentx.artifact_link.v1` | Artifact lineage link |

schema string 与 JSON field/tag 是兼容面。新增 optional field 可以在独立 review
后进入同版本；改变字段含义、required 规则或删除字段必须另行版本化。

## 统一 Envelope

`Envelope` 提供跨协议关联 identity：

```go
type Envelope struct {
    SchemaVersion      string
    Kind               string
    TimestampUnixMilli int64
    RunID              string
    RootRunID          string
    ParentRunID        string
    SessionID          string
    TurnID             string
    BranchID           string
    NodeExecID         string
    TraceID            string
    SpanID             string
    ParentSpanID       string
}
```

字段都通过 JSON tag 序列化；除 `schema_version` 外，多数字段按
`omitempty` 输出。具体协议的 validator 决定哪些字段在该对象中必填。

## RunEvent

`RunEvent` 描述可审计、可投影的 Runtime 事件，包含：

- envelope 与 `SourceEvent`/`SourceEventID`；
- status、reason、stage、tool/model/provider identity；
- duration、attempt、cache、retry；
- `ErrorInfo`、`Usage`、`RuntimeDecisionSnapshot`；
- execution contract identity/diff；
- 可选 `Attrs`。

```go
func NormalizeRunEvent(RunEvent) RunEvent
func ValidateRunEvent(RunEvent) error
```

normalize 会 trim/lower 已定义字段、复制 `Attrs`，并归一化嵌套 error、usage
和 runtime decision。validate 要求正确的 `RunEventSchemaV1`、非空 `kind`
和 `run_id`。

`ErrorInfo.Message` 只是人类可读辅助字段；机器逻辑应优先使用 class、code、
reason、retryable 和 degraded。

## TraceSpan

`TraceSpan` 当前是 **reserved schema**：

```go
func NormalizeTraceSpan(TraceSpan) TraceSpan
func ValidateTraceSpan(TraceSpan) error
```

package 只保证 type、normalization、validation 与 JSON shape。它不承诺 span
producer、时钟/lifecycle、树投影、redaction、operator report 或 CI gate。
validator 要求 schema、kind、run ID、trace ID、span ID 和 type。

## ToolExecutionPlan

`ToolExecutionPlan` 表达执行前的可观察计划：

- `PlanID`、`MaxConcurrency`；
- execution contract identity/diff；
- `ToolPlanCall`；
- `ToolPlanInterruption`；
- `ToolPlanBlockedCall`；
- `RuntimeDecisionSummary`。

```go
func NormalizeToolExecutionPlan(ToolExecutionPlan) ToolExecutionPlan
func ValidateToolExecutionPlan(ToolExecutionPlan) error
```

validator 要求 plan ID，并要求每个 call 具有 tool-call ID 和 tool name。
本结构不执行计划，也不能覆盖 execution owner 的授权结果。

## HandoffRecord

`HandoffRecord` 记录 source/target、input filter、isolation、status、reason 和
error：

```go
func NormalizeHandoffRecord(HandoffRecord) HandoffRecord
func ValidateHandoffRecord(HandoffRecord) error
```

`HandoffEndpoint` 可以用 agent、pack、workflow、node、run、session、task 或
tool-call identity 描述端点。validator 要求 handoff ID，并要求 source/target
至少包含一种 identity。

handoff kind 常量：

```text
HandoffKindContextOnly
HandoffKindArtifact
HandoffKindTaskChild
HandoffKindAgentAsTool
HandoffKindWorkflowNode
HandoffKindExternalHost
```

## SandboxManifest

`SandboxManifest` 描述已解析的 workspace/sandbox 事实：

- platform、root、backend；
- entries、environment、path grants；
- network 与 command policy；
- degraded/reason。

```go
func NormalizeSandboxManifest(SandboxManifest) SandboxManifest
func ValidateSandboxManifest(SandboxManifest) error
```

validator 要求 manifest ID、root，并拒绝 entry path 中的 `..` segment。
manifest 是描述性协议，不负责创建 sandbox、扩大 host authority 或解释策略。

## ArtifactVersion 与 ArtifactLink

`ArtifactVersion` 描述 artifact identity、version、URI、scope、MIME type、
creator、metadata 与 payload：

```go
func NormalizeArtifactVersion(ArtifactVersion) ArtifactVersion
func ValidateArtifactVersion(ArtifactVersion) error
```

`ArtifactLink` 描述两个 artifact version 的 lineage：

```go
func NormalizeArtifactLink(ArtifactLink) ArtifactLink
func ValidateArtifactLink(ArtifactLink) error
```

protocol 只定义结构，不读取 path/URL/blob、不验证 digest 内容、不保存 artifact
也不拥有 lineage store。

## Status、Kind 与 host-process 常量

通用状态：

```text
StatusPlanned
StatusBlocked
StatusApprovalPending
StatusReady
StatusRunning
StatusCompleted
StatusFailed
StatusSkipped
```

已定义的 tool/sandbox/handoff/host-process kind 只是协议 identity。它们不构成
自动状态机，也不授权调用方跳过 execution、scheduler、runstore 或 host owner。

## 值、并发与可变字段

大部分协议是普通 Go value，但包含 map、slice 和 pointer：

- normalize 会复制本 package 明确处理的 `Attrs`、slice 和嵌套值；
- 调用方把值交给其它 goroutine 或 store 后，不应并发修改可变字段；
- 本 package 不提供全对象图 deep clone；
- JSON marshal/unmarshal 遵循 Go 标准库行为。

package 没有 goroutine、global mutable state、I/O 或 shutdown 生命周期。

## 错误合同

validation error 当前是稳定的 display-safe string contract，前缀为：

```text
agentx/runtime/protocol:
```

本轮迁移逐字节保留现有文本。它们尚未升级为根 AgentX typed error/code；
调用方不应据此推断业务 retry 或 authorization。未来改成 typed error 必须另开
兼容任务，不能借 source migration 顺手修改。

## Non-goal

- Runtime construction、Runner 或 Facade；
- projection、durable store、telemetry producer；
- auth/approval/sandbox/budget/retry policy；
- provider、credential、HTTP、CLI 或 Scene；
- TraceSpan materialization；
- Public/Beta/Stable 或生产就绪承诺。

完整嵌入式 Runtime、module version、发布策略和 Scene distribution 仍以
AgentX 产品化规划与后续 work packet 为准。
