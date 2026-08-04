# Gemini Native Provider API（Experimental）

`github.com/wsnacj/agentx-go/providers/gemini` 承载 Gemini native HTTP 协议、JSON/SSE
解析、embedding 和 Files API。该包不读取环境变量，不发现 API key，也不选择模型。

通过 `New(Config)` 构造并发安全的 `Client`。`Config.Authorize` 由 Host 注入认证头，
`Config.HTTPClient` 可用于代理、测试或受控网络出口，`Config.Transport` 提供显式 header
与 payload/response hook。构造过程不发起网络请求。

主要入口包括：

- `GenerateContent`、`StreamGenerateContent`；
- `EmbedContent`、`BatchEmbedContents`；
- `UploadFile`、`GetFile`、`DeleteFile`；
- native request/response、content、tool、usage 和 file 类型。

`NewProvider(Config)` 进一步提供面向 `components/llm` 的 `Chat`、`Vision`、
`Embedding` 与 normalized stream event。模型默认值通过 `ModelConfig` 或
`EmbeddingConfig` 显式传入；本地媒体只能通过 `Config.ResolveMedia` 由 Host 批准和解析，
provider 本身不读取文件系统。

`ModelConfig.ModelCapabilities()` 返回 provider-neutral 能力描述。当前高层 `Provider`
会拒绝带 tools 的请求，因此即使 native API 类型包含 tools，该映射也保持
`ToolCalling=false`，不会虚报尚未开放的高层路径。

默认 endpoint 仅用于协议兼容。生产 Host 仍应显式确认 endpoint、凭据、代理、配额、
重试和网络授权。当前包为 Experimental，不构成 Public/Beta/Stable 承诺。
