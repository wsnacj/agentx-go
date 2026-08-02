# `providers/codex` 中文 API Reference

成熟度：`Experimental extension`

该包拥有OpenAI Codex Responses协议的真实payload、SSE收集、message/function-call响应与
usage解码、HTTP状态错误实现。它不拥有登录、OAuth刷新、token store或用户目录读取。

## 构造

```go
func New(Config) (*Client, error)
```

`Config`要求Host显式提供`BaseURL`，并允许注入transport defaults、`Authorizer`和`HTTPDoer`。
默认User-Agent/originator保持AgentX既有Codex协议身份，但可由Host覆盖。构造不发起网络请求。

## Chat

```go
func (*Client) Chat(context.Context, ModelConfig, llm.ChatRequest) (*llm.ChatResponse, *llm.Usage, error)
```

该调用把Responses SSE收集为现有`llm.ChatResponse`，支持文本、function call、tool result、
reasoning effort、prompt cache key和usage。非200响应保持`*providers.APIError` identity；
failed/incomplete/缺少terminal event保持稳定typed-text错误边界。

## Host责任

Host必须显式完成token发现/刷新、account ID投影、credential rotation、endpoint/proxy、审计、
模型路由和生产出网授权。`Authorizer`可在每次请求时解析最新token；canonical代码不读取环境、
keychain、auth文件或用户目录。

## 并发与取消

`Client`构造后不修改内部状态，可并发调用。`Chat`遵守调用方context；无deadline时使用有界
默认超时。Host注入的`HTTPDoer`和`Authorizer`也必须支持并发。
