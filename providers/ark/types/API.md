# `providers/ark/types` 中文 API Reference

成熟度：`Experimental extension`

该 package 定义 Ark Responses、Files、Image Generation、tool 与 stream event 的 provider
wire contract，并提供无网络副作用的构造和解码 helper。它不读取 credential、不选择 endpoint、
不发起 HTTP 请求，也不拥有模型路由或产品 fallback。

## 主要合同

- `ResponseRequest`、`InputUnion`、`InputItem`、`InputContent`、`ContentItem`：Responses 输入；
- `Response`、`OutputItem`、`OutputContentItem`、`OutputAnnotation`、`Usage`、`APIError`：响应；
- `Tool` 及 Function/Web Search/Knowledge Search/Image Process/MCP/Doubao App 相关类型：tool wire；
- `ResponseEvent` 与 typed stream event：Responses SSE；
- `ImageGenerationResponse`、`ImageGenerationEvent` 及 typed image events：图像生成；
- `UploadFileRequest`、`FileObject`、`FileList`、`InputItemList`：Files API；
- `ThinkingConfig`、`ReasoningConfig`、`CachingConfig`、`ContextManagement`：请求控制。

`ResponseRequest.ParallelToolCalls`直接映射provider wire字段。Host只有在目标model/endpoint已经完成能力
核对后才应设置；`false`表示要求Provider不要在同一响应中并行生成多个Function Call，不等同于总Tool预算。
本合同不提供`max_tool_calls`数量字段，调用数量、allowlist和副作用准入仍必须由Host独立fail closed。

## 构造与解码

`NewInputText/Items`、`NewMessage`、`NewInput*Item`、`NewFunctionTool` 及各 provider tool builder
只组装 DTO；`ObjectSchema/ArraySchema/StringSchema/...` 只构造 JSON Schema map。
`DecodeResponseEvent`、`DecodeTypedResponseEvent`、`DecodeImageGenerationEvent` 与
`DecodeTypedImageGenerationEvent` 只解析调用方提供的 bytes，不启动 stream 或 goroutine。

## 兼容与边界

这些类型直接映射 provider JSON，字段名、omitempty、union marshal/unmarshal 与 event type 是
协议兼容面，但当前仍为 Experimental，不形成 Public/Beta/Stable 承诺。认证头、HTTP client、
timeout、SSE 生命周期、模型级字段能力、工具授权和执行由 `providers/ark` Client 与 Host 负责。
