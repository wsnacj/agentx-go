# `scenes/docparse/profile` 中文 API Reference

成熟度：**internalization candidate**。

- `ExtractionProfile`、`Registry`、`Matcher`：保存 Host 显式注册的 profile 并做确定性匹配。
- `MatchInput`、`MatchResult`、`Candidate`、`Proposal`：描述 verified、candidate 与 unknown
  三种结果；unknown 不会被当作已验证 profile。
- `PromoteReviewedProposal`：只在显式 review decision 与 regression seed 齐备时生成
  promotion artifact，不写磁盘或注册生产配置。

本包不内置发票、合同或客户文档规则，不调用模型或搜索 provider。
