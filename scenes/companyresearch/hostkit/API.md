# `scenes/companyresearch/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

本包把公司研究合同组合成 provider-neutral Host Kit。调用方显式注入财报、A 股、
全球行情、公开新闻、公告、研报、风险和主体解析 handler；本包负责调用顺序、预算内
聚合、readiness、task summary 与 bounded answer contract。

## 推荐入口

- `CompanyResearchConfig`：声明 source、handler、主体解析与任务执行 seam。
- `CompanyResearchHandlers`：按 evidence dimension 注入 concrete adapters。
- `BuildCompanyResearchLookupPayload`：单主体研究协调。
- `BuildCompanyComparePayload`：多主体比较协调。
- `BuildCompanyResearchGuardPayload`：只对已提供 evidence 做 guard。
- `BuildCompanyResearchLookupHandler`、`BuildCompanyCompareLookupHandler`、
  `BuildCompanyResearchGuardHandler`：包装为 portable tool handler。
- `SubjectResolver`、`TaskExecutor`：可选 Host ports；实现必须遵守传入 context，并自行
  保证并发、安全、授权和审计边界。

Host Kit 会过滤 task diagnostics 中的 model/provider/token/secret 等敏感键，但这不是
完整安全系统。credential 管理、provider 选择、真实网络、RunStore、scheduler、审批和
产品展示仍由 Host 拥有。本包不依赖 HS、Runner 或 Scene。
