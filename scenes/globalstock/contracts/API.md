# `scenes/globalstock/contracts` 中文 API Reference

成熟度：**Experimental extension**。

本包定义 HK/US security、quote、profile、announcement、research、signal、investigation、
freshness、evidence、identity-resolution 与 readiness JSON 合同。

- `NormalizeSecurityCode` 规范化 HK/US code/ticker，不发起身份查询；
- `BuildReadiness` 对缺字段、review-required、source failure 做 fail-closed 投影；
- `QuotePayload`、`ProfilePayload`、`AnnouncementPayload`、`ResearchPayload`、
  `SignalPayload` 保留来源、时点、币种与 provider-attempt 证据；
- `InvestigationIntent` / `InvestigationPayload` / `ToolHandoff` 描述高层只读分析和显式跨域移交。

这些类型不代表数据真实或适合交易；Host 必须验证 provider、授权、时效和合规边界。

