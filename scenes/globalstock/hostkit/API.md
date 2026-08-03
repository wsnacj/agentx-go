# `scenes/globalstock/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

`BuildGlobalStockInvestigationPayload` 通过 `InvestigationHandlers` 在 quote/profile/
announcement/research/signal 中选择一条路径；每个选中的 handler 每次请求最多调用一次。
多主体 comparison 按首次出现顺序调用 quote handler，并确定性聚合 readiness。

```go
cfg := hostkit.InvestigationConfig{
    Handlers: hostkit.InvestigationHandlers{Quote: quoteHandler},
    AnswerContract: hostOwnedComparisonDraft,
}
result, err := hostkit.BuildGlobalStockInvestigationPayload(ctx, cfg, args)
```

`AnswerContract` 是可选 Host seam：canonical 不拥有回答文案、投资免责声明或产品策略。
context 原样传递，不创建 goroutine、不自动重试。handler 必须自行遵守取消、deadline、并发、
provider授权和真实副作用边界。

