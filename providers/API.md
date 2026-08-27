# Providers 模块 API（Experimental）

`github.com/wsnacj/agentx-go/providers` 是 AgentX 的可选模型供应商模块。它承载可移植的
HTTP 协议、错误分类、有限重试、usage 收集接缝和具体 provider client；不负责选择模型、
读取环境变量、发现或轮换凭据、租户配额、endpoint allowlist、代理策略和生产网络授权。

当前实现包括 [`openaicompat`](./openaicompat/API.md)、[`anthropic`](./anthropic/API.md)、
[`codex`](./codex/API.md)、[`gemini`](./gemini/API.md) 和 [`ark`](./ark/API.md)。调用方必须显式提供或确认endpoint，需要认证时通过
`Authorize`注入。`New`只构造client，不发起网络请求。

## 包索引

- `providers`：`APIError` 与 `ErrUnsupported`；
- `providers/openaicompat`：OpenAI-compatible chat、vision、embedding、bot 与 SSE stream；
- `providers/anthropic`：Anthropic Messages payload、tool use/result、响应与usage；
- `providers/codex`：Codex Responses payload、SSE收集、function call与usage；
- `providers/gemini`：Gemini native JSON/SSE、embedding 与 Files API；
- `providers/ark`、`providers/ark/types`：Ark Responses/Files HTTP client 与数据合同；
- `providers/transport`：request options、header、payload/response hook；
- `providers/fault`：稳定错误分类和 retryability；
- `providers/retry`：受 context、次数与 backoff 约束的调用重试；
- `providers/usage`：`Collector` 与 `NoopCollector`。

所有包目前均为 Experimental，不构成 Public/Beta/Stable 或正式发行承诺。

## 高层能力一致性基线

下表描述当前 canonical 高层 adapter 已真正实现的合同，不代表对应供应商的全部模型能力。`显式`表示 Host
必须对所选模型开启 capability；`透传`表示只提供受控 provider 字段映射，不构成统一强类型保证。

| Provider | Stream | Structured output | Function Tool | Thinking / reasoning | Cache | Usage |
|---|---|---|---|---|---|---|
| OpenAI-compatible | 显式 | `response_format`透传 | 是；并行能力显式 | 显式 | transport透传 | 是 |
| Anthropic | 否 | 否 | 是；non-stream | 否 | 否 | 基础input/output |
| Gemini native | 显式 | `generationConfig`透传 | 显式；Chat与StreamChat | 显式thinking | native低层类型；高层未开放 | prompt/output/total/reasoning |
| Ark Responses | 是 | native Responses合同 | 是 | 是 | provider报告为准 | prompt/output/total/cache/reasoning |
| Codex Responses | 内部SSE收集；无公开normalized stream | provider合同 | 是 | 显式 | prompt cache key | prompt/output/total/cache/reasoning |

所有 adapter 的网络/API失败统一保留 `providers.APIError` cause；Platform或其它Host必须在外部边界再次映射为
display-safe error，不能把响应正文、prompt、模型输出或凭据直接显示给调用方。缺少的组合能力必须返回
`ErrUnsupported`或保持 capability=false，不允许通过静默换模、自动fallback或空实现伪装支持。
