# `scenes/finance` 中文 API Reference

成熟度：**Experimental extension**。当前没有 semver、Developer Preview、Public、Beta 或
Stable 兼容承诺。

本包是公开财报证据分析的 portable owner：定义 lookup intent/result、候选、指标、简报、
assessment、period、identity、readiness 与 answer-contract，并组合 metrics/brief 两个只读 Pack。
它不下载报告、不调用模型、不解析 PDF、不选择数据商，也不输出投资建议。

## 主要 API

- `PackDefinitions` / `RegisterPacksIntoRegistry`：返回或注册 metrics 与 brief Pack；
- `MetricSpecFromFields`、`NormalizeMetricFields`：规范化指标请求并检查缺失/review字段；
- `NormalizePeriodScope`、trend helpers：确定性处理 latest/recent-years period scope；
- `FinanceReportLookupAnswerReadiness`、`FinanceReportLookupAnswerContract`：仅基于已提供 evidence
  计算能否回答和允许的摘要范围；
- `FinanceReportAssessmentFromPayload`：投影业务表现/风险 assessment，不生成估值或交易结论；
- `NormalizeBriefEvidence`、`ReviewBriefEvidence`、`BuildBriefText`：对已提取简报证据做纯函数收口。

运行时协调见 [`hostkit`](hostkit/API.md)，Pack/evaluator 分别见
[`metrics`](metrics/API.md) 与 [`brief`](brief/API.md)。

## 非目标

- Browser、HTTP、报告下载、docparse/OCR/PDF、LLM tool registry/executor；
- provider、token、付费数据、来源优先级、客户策略、免责声明与真实副作用；
- Public/Beta/Stable、正式 module tag 或发行授权。

