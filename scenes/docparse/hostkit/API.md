# `scenes/docparse/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

本包把 Docparse semantic tools、结构化 parse result 与 evidence evaluator 组合成无
provider Host Kit。

## 构造与 Ports

- `Config.Executor`：可选 canonical `tools.Executor`，用于显式代理 Host 的
  `document_parse` / `document_spec_recommend`。
- `Config.ResultLoader`：可选、Host-owned 的本地结果读取 port；未配置时 `result_path`
  fail closed，`parse_result` 仍可纯内存使用。
- `New`、`DefaultConfig`、`BuildStandardToolHandlers`：构造 Kit 与 tool handlers。

## 操作

- `SpecSelect`、`ProfileProbe`、`ExtractFields`、`ExtractTable`：调用 Host executor 或投影
  caller-provided parse result。
- `TraceEvidence`、`Validate`、`Guard`：确定性生成 evidence、failure class、
  review-required 与 answer boundary。
- `Docparse*Tool`、`RegisterTools`、`DecodeToolArguments`：提供 canonical LLM tool schema
  与 registry wiring。

Host 必须自行保证 loader/executor 的路径授权、取消、并发、credential、安全、审计和副作用
边界。本包不打开文件或网络，也不选择 OCR/LLM provider。
