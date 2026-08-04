# OpenAI-compatible Client API（Experimental）

## 构造

```go
client, err := openaicompat.New(openaicompat.Config{
    Name:    "my-provider",
    BaseURL: "https://provider.example/v1",
    Authorize: func(ctx context.Context, h http.Header) error {
        // 凭据由 Host 获取；canonical module 不读取环境变量。
        h.Set("Authorization", "Bearer "+token)
        return nil
    },
})
```

`Config` 在构造时复制静态 headers。`Client` 构造后可被多个 goroutine 并发调用；调用方注入
的 `HTTPDoer`、`Authorizer`、hook 和 `MediaResolver` 也必须满足相同并发边界。

## 调用

- `Chat(ctx, ModelConfig, llm.ChatRequest)`；
- `Vision(ctx, ModelConfig, llm.VisualRequest)`；
- `Embedding(ctx, EmbeddingConfig, llm.EmbeddingRequest)`；
- `Bot(ctx, ModelConfig, llm.ChatRequest)`；
- `StreamChatEvents` / `StreamVisionEvents`：推荐的规范化事件流；
- `StreamChat` / `StreamVision`：兼容旧 chunk stream。

`ModelConfig` 和 `EmbeddingConfig` 是 Host 已完成模型选择后的显式输入，不是产品模型目录。
`Capability` 控制可选协议字段；缺少 streaming 等能力时返回 `providers.ErrUnsupported`。
`ModelConfig.ModelCapabilities()` 将已显式配置的 text/tool/vision/stream/local-media/
reasoning/parallel-tools/bot 能力投影为 `llm.ModelCapabilities`；它不探测远端模型。

## 取消、错误与 usage

调用遵守调用方的 cancellation/deadline。调用方未设置 deadline 时，非流式 chat、embedding
和流式调用分别保留有限默认超时。非 200 响应返回 `*providers.APIError`，可用
`errors.As` 检查；`fault.Classify` 提供稳定 kind 与 retryability。成功响应返回
`*llm.Usage`；是否记录、计费或上报由 Host 决定。

## 安全边界

默认构造不读取文件、环境变量或 credential store。只有 Host 显式设置 `ResolveMedia` 且模型
允许 local files 时才可解析本地媒体。真实 endpoint、proxy、credential、审计、配额与出网
授权仍属于 Host。
