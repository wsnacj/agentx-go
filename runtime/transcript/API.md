# runtime/transcript API

导入路径：

```go
import "github.com/wsnacj/agentx-go/runtime/transcript"
```

成熟度：**Experimental / private validation**。

`transcript` 是 AgentX 的纯内存对话处理 owner：它估算上下文字符预算、给出无副作用的
阈值决策、压缩较早的工具/消息正文、修复严格工具调用协议并按完整协议片段裁剪历史。

它不拥有 provider 选择、模型目录、tokenizer、Session/RunStore、日志、telemetry、Hook、
credential 或网络。Host 必须把模型与产品配置映射为显式 policy。

## 预算估算与 Guard

```go
type EstimateInput struct {
    SystemPrompt string
    Chunks       []string
    Messages     llm.Conversation
    RoleAware    bool
}

type GuardPolicy struct {
    WarnChars int
    MaxChars  int
}

type GuardDecision struct {
    EstimatedChars int
    Warn           bool
    Overflow       bool
}

func EstimateChars(input EstimateInput) int
func Evaluate(input EstimateInput, policy GuardPolicy) GuardDecision
func OverflowMessage(estimatedChars, maxChars int) string
```

`RoleAware=true` 时只计算 `Messages`；否则计算 `Chunks`。工具调用名称和参数计入估算，
但本 API 不把字符数冒充精确 token 数。阈值为零表示关闭；等于阈值即触发。
`OverflowMessage` 提供兼容的 display text，错误 identity 与 recovery 仍由 Host 决定。

## 压缩

```go
type AnchorSelector func(content string) int

type CompactionPolicy struct {
    MaxChars         int
    ToolOutputAnchor AnchorSelector
}

func Compact(messages llm.Conversation, policy CompactionPolicy) (llm.Conversation, Diagnostic)
func CompactToolOutputs(messages llm.Conversation, policy CompactionPolicy) (llm.Conversation, int)
func CompactHistoryBodies(messages llm.Conversation, maxChars int) (llm.Conversation, int)
func TruncateToolOutput(content string, maxChars int, selector AnchorSelector) string
```

压缩顺序固定为：较早的 tool output，再到较早的 user/assistant 正文；最后一条消息和工具
协议字段不会作为普通历史正文压缩。`ToolOutputAnchor` 是可选 Host policy，用于保留调用方
认为重要的尾部片段；canonical package 不内置 PDF、citation 或业务标记。

启用压缩时返回 defensive copy，不修改调用方切片。selector 必须确定性且可安全并发调用。
`TruncateToolOutput` 暴露与 `Compact` 相同的单条 head/tail 截断机制，便于 Host 保留已有
package-private 兼容入口；它不执行整段预算判断或历史正文压缩。
两个 `Compact*` 分阶段函数用于已有 Host 维持原诊断计数和调用顺序；新调用方通常应直接
使用组合后的 `Compact`。

## 协议修复与历史裁剪

```go
type SanitizePolicy struct {
    StrictToolProtocol     bool
    StripInternalReasoning bool
}

type HistoryPolicy struct {
    MaxEvents          int
    StrictToolProtocol bool
}

func Sanitize(messages llm.Conversation, policy SanitizePolicy) (llm.Conversation, Diagnostic)
func Prune(messages llm.Conversation, policy HistoryPolicy) (llm.Conversation, int)
func PruneTailPreservingSystemPrefix(messages llm.Conversation, policy HistoryPolicy) (llm.Conversation, int)
```

严格模式会为缺失 ID 的 tool call 生成稳定的 `agentx_call_N`，在只有一个 pending call 时
恢复 legacy tool result 的 ID；无法消歧的 tool result 降为 assistant 消息。是否对某个
provider 启用严格模式属于 Host policy，不由本 package 猜测。

`Prune` 在严格模式下把 assistant tool-call 与随后的 tool result 视为不可拆分片段；
`PruneTailPreservingSystemPrefix` 额外保留开头连续的 system 消息。返回的整数是因协议片段
边界而放弃的消息数。

## Diagnostic

`Diagnostic` 只含转换计数，不含 prompt、消息内容或凭据。Host 可把它投影到自己的
observability contract。本 package 不记录日志，不发送事件，也不持久化内容。
