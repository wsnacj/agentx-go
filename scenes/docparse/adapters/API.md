# `scenes/docparse/adapters` 中文 API Reference

成熟度：**internalization candidate**。

- `Adapter`、`Registry`、`Func`：显式注册并执行首个支持某 route 的 adapter。
- `RouteAdapter` 与 `New*Adapter`：把 Host function 适配为 fixed-form、table 或 OCR+LLM
  route；没有注入函数时 fail closed。
- `SpecDocparseAdapter`：通过 canonical `tools.Executor` 调用 Host 的 `document_parse`。
- `OCRXHTMLLLMParser`：只定义 provider-neutral port；credential、prompt、cache、retry 与
  cost policy 始终由 Host 持有。

本包不会创建 `core/card`、`core/table`、OCR 或模型实例。
