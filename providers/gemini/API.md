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

默认 endpoint 仅用于协议兼容。生产 Host 仍应显式确认 endpoint、凭据、代理、配额、
重试和网络授权。当前包为 Experimental，不构成 Public/Beta/Stable 承诺。

