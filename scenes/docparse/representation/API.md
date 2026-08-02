# `scenes/docparse/representation` 中文 API Reference

成熟度：**internalization candidate**。

本包定义 provider-neutral 的 `Document`、`Page`、`Diagnostics`、`PageRange` 与
`ExtractionMode`，并提供 `FromTextPages`、`SelectPageRange`、`PromptText`、
`ParsePageRange`、`SplitTextByChars` 等纯内存机制。

本包不读取文件、不检测 MIME、不创建 OCR/PDF client。Host 应从 canonical
`agentx-go/document` 或自己的授权后端获得页文本，再调用 `FromTextPages`。JSON 字段沿用
HS 迁移前合同；进入 Beta 前仍可能被收进更高层 facade。
