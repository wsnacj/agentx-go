# Ark Provider API（Experimental）

`github.com/wsnacj/agentx-go/providers/ark` 承载 Ark Responses/Files 协议的可移植 HTTP
client；`providers/ark/types` 是对应的 request、response、tool、stream event、image 和 file
数据合同。该包不读取环境变量，不发现 API key，也不拥有模型别名或产品 fallback 策略。

通过 `New(Config)` 构造 `Client`。Host 使用 `Config.Authorize` 注入认证头，并可通过
`Config.HTTPClient`、`Config.Transport`、`Timeout` 和 `StreamTimeout` 控制网络出口与超时。
构造过程不发起网络请求。

当前 client 提供：

- `DoJSON`：Responses、Files 等 JSON endpoint；
- `DoStream`：SSE response body；
- `UploadFile`：multipart Files API；
- `ListFiles`、`ListInputItems`：分页查询。

能力探测、模型路由、降级策略、凭据管理和业务重试继续由 Host 负责。当前包为
Experimental，不构成 Public/Beta/Stable 承诺。

