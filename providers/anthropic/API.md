# `providers/anthropic` 中文 API Reference

成熟度：`Experimental extension`

该包拥有Anthropic Messages协议的真实请求序列化、tool use/result映射、响应与usage解码、
HTTP状态错误和transport hook实现。它不读取环境变量、credential store或本地文件。

## 构造

```go
func New(Config) (*Client, error)
```

`Config`显式提供`BaseURL`、版本、超时、transport defaults、`Authorizer`和可选`HTTPDoer`。
未提供`BaseURL`时保持Anthropic官方endpoint默认值；构造过程不发起网络请求。

## Chat

```go
func (*Client) Chat(context.Context, ModelConfig, llm.ChatRequest) (*llm.ChatResponse, *llm.Usage, error)
```

`ModelConfig`只包含Host选择的模型默认值。请求支持system、assistant tool call、tool result、
tool schema和tool choice；非200响应返回`*providers.APIError`。credential、模型路由、配额、
proxy、审计与生产出网授权必须由Host提供。

`ModelConfig.ModelCapabilities()` 声明当前 canonical Messages adapter 的 text generation
与 tool calling；未实现的 stream/vision 等字段保持 fail-closed。

## 并发与取消

`Client`构造后不修改内部状态，可并发调用。`Chat`遵守调用方context；若调用方没有更短
deadline，则应用有界请求超时。Host注入的`HTTPDoer`和`Authorizer`也必须支持并发。
