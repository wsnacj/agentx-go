# `runtime/contextwindow` API 参考

导入路径：

```go
import "github.com/wsnacj/agentx-go/runtime/contextwindow"
```

成熟度：**Experimental**。本包不属于 `v0.2.1` 的 Developer Preview
兼容候选面，也不构成 Public、Beta 或 Stable 承诺。

`contextwindow` 提供 provider-neutral 的上下文窗口准备编排。它复用
`runtime/transcript` 的协议修复、确定性 tool output 压缩和历史裁剪，并在仍超限时调用 Host
显式注入的 `Summarizer`。本包不选择模型/provider，不读取凭据，不发送网络请求，也不保存
Session、summary 或用户记忆。

## 构造与执行

```go
type Orchestrator struct {
    Policy       Policy
    Summarizer   Summarizer
    TokenCounter llm.TokenCounter
}

func (o Orchestrator) Prepare(
    ctx context.Context,
    request Request,
) (Result, error)
```

`Prepare` 的顺序固定为：

1. 检查 caller cancellation/deadline；
2. `transcript.Sanitize`；
3. 可选 protocol-aware history prune；
4. 确定性 tool output 压缩；
5. 仍超限时，保护 system prefix、显式 head segment、最新用户消息及其后的 tail segment；
6. 只把中间窗口交给 Host summarizer；
7. 以带稳定 marker 的 system message 替换中间窗口并重新验算。

错误路径返回原始消息的 defensive copy，不返回可被误持久化的半成品历史。

## Policy

```go
type Policy struct {
    WarnChars              int
    MaxChars               int
    MaxEvents              int
    StrictToolProtocol     bool
    StripInternalReasoning bool
    ToolOutputAnchor       transcript.AnchorSelector
    ProtectedHeadSegments  int
    ProtectedTailSegments  int
    SummaryTargetChars     int
    WarnInputTokens        int64
    MaxInputTokens         int64
}
```

- `MaxChars` 与 `SummaryTargetChars` 必须大于零；
- `ProtectedTailSegments` 至少为 1，canonical 还会把最新 user segment 到末尾全部保护；
- assistant tool call 与紧随的 tool results 是一个不可拆分 protocol segment；
- `MaxEvents=0` 表示不做 history prune；
- 字符预算继续兼容；只有 Host 设置 token limit 并注入 `TokenCounter` 时才启用 token-aware
  溢出判定和压缩后复核；
- `TokenCount.Exact=false` 会原样进入 report，不冒充精确 tokenizer 或 provider usage。

`Request` 可同时携带 `Model` 和 `Tools`，使 Host tokenizer 能覆盖 system、conversation 和
tool schema。token counter 失败返回 `token_count_failed`；调用方取消仍优先映射为
`canceled`/`deadline_exceeded`。

## Summarizer port

```go
type Summarizer interface {
    Summarize(context.Context, SummaryRequest) (Summary, error)
}

type SummaryRequest struct {
    Messages        llm.Conversation
    PreviousSummary string
    TargetChars     int
}

type Summary struct {
    Content string
}
```

Host 负责 summary prompt、模型/provider、credential、retry/backoff、计费和 telemetry。实现必须
传播 `ctx`，不得修改 `Messages`。空 summary 或超过 `TargetChars` 的 summary 会 fail closed。

`PreviousSummary` 是显式输入；本包没有跨调用状态。若 Host 未显式提供，但消息中含本包上次
生成的 summary marker，`Prepare` 会恢复最近一份 summary 并避免重复注入。

## Result 与 Report

```go
type Result struct {
    Messages llm.Conversation
    Summary  string
    Report   Report
}
```

`Report` 只包含前后字符/token 数、token count 的 exact/source、转换计数和保护 segment 数，
不包含 prompt、消息、summary、凭据或 provider 错误。Host 可以把这些 bounded 计数投影到
自己的 observability；source 会移除换行并限制长度。

## Typed error

```go
type Error struct {
    Code  ErrorCode
    Cause error
}

func AsError(error) (*Error, bool)
```

稳定 code 包括 `invalid_policy`、`canceled`、`deadline_exceeded`、
`summarizer_unavailable`、`summarization_failed`、`invalid_summary`、`token_count_failed` 和
`limit_unresolved`。`Error()` 只返回 display-safe 文本；底层原因通过 `errors.Unwrap` 提供，
因此 `errors.Is(err, context.Canceled)` 与 `errors.As` 均可用。

## 并发和状态边界

`Orchestrator` 不维护可变状态，可被并发调用；`ToolOutputAnchor` 和 `Summarizer` 的并发能力由
Host 保证。Session summary、SQLite/vector store、memory ranking、visibility、用户画像和
Shutdown lifecycle 不属于本包，应由 stateful Host 持有。
