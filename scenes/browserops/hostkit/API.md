# `scenes/browserops/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

本包把五个 Browser Ops 语义工具协调到 Host 显式注入的 canonical `tools.Executor`。它只负责
参数投影、runtime tool 选择、统一 unsupported payload 和任务观察，不创建真实浏览器或网络 client。

## 构造与执行

- `Config`：注入 `Executor`、source、runtime tool 名称、`TaskObserver` 和有界默认参数；
- `DefaultConfig`、`New`、`BuildStandardToolHandlers`：构造 Kit 和标准 handler；
- `Kit.OpenTarget`、`FillFields`、`CapturePageSnapshot`、`CaptureSubmissionEvidence`、
  `DownloadFile`：每次最多调用一次 Host executor，并遵守传入 `context.Context`；
- Executor 为空时返回结构化 `unsupported` payload，不执行隐式 fallback 或真实副作用。

## Tool 注册与解码 seam

- `Browser*Tool` 与 `ToolNames`：返回五个语义工具 schema 和稳定名称；
- `RegisterTools`：使用严格 JSON object decoder 注册到 canonical `tools.Registry`；
- `RegisterToolsWithDecoder`：允许 Host 注入 `ArgumentDecoder`，用于保留既有兼容解码策略；
- `DecodeToolArguments`：空字符串解码为空 object，非 object 或非法 JSON 返回带
  `decode browser-ops tool arguments` 前缀的错误；
- decoder repair policy 不属于本包。HS adapter可注入既有 tolerant decoder，但 canonical 不
  复制其 retrieval、model-output repair 或产品策略。

## 观察与边界

`TaskObservation`、`TaskObserver`、`TaskObserverFunc` 和 `MultiTaskObserver` 提供同步观察 seam。
调用方负责 observer 的并发安全、延迟和副作用；本包不会持久化、调度或重试 observation。

真实 browser backend、profile/login state、credential、approval、文件路径授权、artifact retention、
站点策略、审计和产品编排必须由 Host 提供。本包不依赖 HS、Runner 或旧 Scene。
