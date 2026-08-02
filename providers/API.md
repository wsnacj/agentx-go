# Providers 模块 API（Experimental）

`github.com/wsnacj/agentx-go/providers` 是 AgentX 的可选模型供应商模块。它承载可移植的
HTTP 协议、错误分类、有限重试、usage 收集接缝和具体 provider client；不负责选择模型、
读取环境变量、发现或轮换凭据、租户配额、endpoint allowlist、代理策略和生产网络授权。

当前首个实现是 [`openaicompat`](./openaicompat/API.md)。调用方必须显式提供 endpoint，
需要认证时通过 `Authorize` 注入。`New` 只构造 client，不发起网络请求。

## 包索引

- `providers`：`APIError` 与 `ErrUnsupported`；
- `providers/openaicompat`：OpenAI-compatible chat、vision、embedding、bot 与 SSE stream；
- `providers/transport`：request options、header、payload/response hook；
- `providers/fault`：稳定错误分类和 retryability；
- `providers/retry`：受 context、次数与 backoff 约束的调用重试；
- `providers/usage`：`Collector` 与 `NoopCollector`。

所有包目前均为 Experimental，不构成 Public/Beta/Stable 或正式发行承诺。
