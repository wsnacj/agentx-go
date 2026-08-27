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
`Embedding` 与 normalized stream event。`Chat` 与 `StreamChatEvents` 可以把 canonical function
Tool 映射为 Gemini `functionDeclarations`、`functionCallingConfig`、`functionCall` 和
`functionResponse`；返回值统一为 `llm.FunctionCall` 或 ToolCall stream event。模型默认值通过 `ModelConfig` 或
`EmbeddingConfig` 显式传入；本地媒体只能通过 `Config.ResolveMedia` 由 Host 批准和解析，
provider 本身不读取文件系统。

`ModelConfig.ModelCapabilities()` 返回 provider-neutral 能力描述。Tool 能力必须由 Host 对具体模型显式设置
`Capability.ToolCalling=true`；默认值为 false，带 Tool 的请求会以 `providers.ErrUnsupported` 关闭。开启后：

- `auto`、`none`、`required` 和指定函数选择均映射到 native function calling mode；
- assistant function call 与 Host tool result 可作为后续轮次历史输入；
- non-stream Chat 返回函数名和 JSON 参数，stream 返回 start/delta/end 与最终 snapshot；
- `Vision`/`StreamVisionEvents` 仍拒绝 Tool，因为当前 `llm.VisualResponse` 没有函数调用结果面。

这些 capability 相互独立：`Streaming=true` 与 `ToolCalling=true` 只声明文本 Chat 的组合实现，不自动声明
Visual Tool、并行 Tool 或任意上游模型均支持。协议字段依据
[Gemini Function Calling](https://ai.google.dev/gemini-api/docs/function-calling)；Host仍负责核对所选模型。

默认 endpoint 仅用于协议兼容。生产 Host 仍应显式确认 endpoint、凭据、代理、配额、
重试和网络授权。当前包为 Experimental，不构成 Public/Beta/Stable 承诺。
