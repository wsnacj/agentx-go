# `extensions/astock/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

本包提供无 provider的 typed Host协调：把模型参数转为 A股 intent，调用 Host显式
注入的 fixture/live handlers，聚合 readiness，并只格式化已经验证的 payload。

## 调查协调

- `InvestigationConfig`：Host拥有 source标签、默认 source policy与 handlers。
- `InvestigationHandlers`：Quote/Research/Signal/Announcement/Profile五个显式函数。
- `BuildAStockInvestigationPayload`：遵守 `context.Context`，按既有稳定顺序调用
  handler并投影 partial/unsupported/error readiness。
- `BuildAStockInvestigationHandler`：返回 provider-neutral `ToolPayloadHandler`，方便
  HS或新 Host接入自己的工具 registry。

## 参数与回答

- `IntentFromParams`、`ParamsFromIntent`：确定性转换 task frame；不解析自然语言。
- `StringArg`、`StringSliceArg`、`FreshnessArg`、`BoolArg`：兼容现有宽输入边界。
- `FormatAStockAnswer`：只接受携带原 tool identity与 readiness的标准 payload；
  模型自行拼装或缺少证据的 payload会返回稳定 skipped文本，不会伪造事实。
- `BuildAStockAnswerFormatHandler`：把 formatter包装为 Host handler。

本包不会构造 HTTP client、选择 provider、读取 credential/cache或执行网络；这些能力
必须由 Host handler显式提供。
