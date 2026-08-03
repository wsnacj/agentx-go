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
