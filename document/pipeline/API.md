# Document Pipeline API（Experimental）

`document/pipeline` 是 AgentX Document 的可移植文档解析运行时。它负责配置加载、页眉页脚处理、章节协调、字段抽取、候选选择、派生字段、校验、诊断、缓存和产物格式；文件解码、章节规则后端和模型调用由宿主显式注入。

> 当前成熟度：Experimental / Developer Preview candidate。它不是 Public、Beta 或 Stable 承诺。

## 构造

```go
runtime, err := pipeline.New(pipeline.Dependencies{
    Loader:    loader,
    Sectioner: sectioner,
    Model:     model,
})
```

- `DocumentLoader`：读取文件、选择 OCR/PDF 策略并返回页面文本；凭据与具体后端留在宿主。
- `Sectioner`：执行具体章节规则，返回可移植 `section.Node`。
- `Model`：完成一次文本模型请求；provider、credential、授权、网络和实际重试均由宿主管理。
- `Observer`：可选，只接收展示安全的阶段事件，不参与决策。

`New` 不读取环境变量，不选择默认 provider，也不产生网络副作用。构造后的 `Runtime` 在注入依赖支持并发的前提下可并发调用。

## 执行

```go
result, err := runtime.Run(ctx, pipeline.ParseRequest{
    DocPath:        "report.pdf",
    SpecPath:       "./spec",
    ModelName:      "document-model",
    ArtifactPolicy: pipeline.ArtifactPolicySummary,
})
```

`Run` 保留调用方的 `context.Context` 取消和 deadline。`ParseBudget.TotalTimeout` 可在调用 deadline 内再收紧总预算。`ArtifactPolicyNone` 不写解析产物；缓存只在调用方显式给出策略和目录时启用。

结果类型位于 `document/pipeline/types`，配置合同位于 `document/pipeline/configs`。JSON 字段保持与 HS 既有 docparse 合同一致。

低层扩展入口：[`configs`](configs/API.md)、[`types`](types/API.md)、
[`section`](section/API.md)、[`preprocessing`](preprocessing/API.md)、
[`extractors`](extractors/API.md)、[`expr`](expr/API.md)、[`derive`](derive/API.md)和
[`utils`](utils/API.md)。这些package有独立中文Reference，但仍为Experimental。

## 明确边界

本包不拥有：

- provider、credential 或环境变量发现；
- OCR/PDF/文件系统后端选择；
- HS `Runner`、Scene 或业务规则；
- 客户专属 spec、prompt 和默认模型；
- 发布稳定性承诺。

HS 迁移期间可以保留兼容入口，但通用解析实现应只存在于本包。
