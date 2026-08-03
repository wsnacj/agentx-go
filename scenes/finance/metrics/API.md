# `scenes/finance/metrics` 中文 API Reference

成熟度：**Experimental extension**。

本包拥有最新财报/多期趋势 Pack、结构化 case frame、字段级 evidence 与确定性 evaluator。

- `Definition` / `RegisterInto` / `MaterializedDefaultWorkflow`：Pack 与 Workflow 入口；
- `EvaluateLatestMetrics`：检查主体、期间、来源、字段 evidence、增长一致性、guard 与 trend；
- `LatestMetricsFieldSourcesAccepted`：拒绝搜索结果页、跨来源字段和缺少局部证据的值；
- `BuildLatestMetricsCaseInput` / `BuildReportMetricsCaseInput`：仅作为显式 legacy fallback，
  不应替代模型结构化 intent 与 Host source verification。

Document parser 在本包中只是接口；具体 docparse/OCR/PDF 实现由 Host 注入。

