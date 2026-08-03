# `scenes/finance/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

`BuildFinanceReportLookupPayload` 依次协调 candidates、metrics extract、metrics guard，以及
调用方明确要求时的 brief extract/guard。所有步骤由 `FinanceReportLookupHandlers` 显式注入；
canonical 不拥有具体 source、docparse、model 或 tool executor。

```go
cfg := hostkit.FinanceReportLookupConfig{
    Handlers: hostkit.FinanceReportLookupHandlers{
        Candidates: candidateHandler,
        MetricsExtract: extractHandler,
        MetricsGuard: guardHandler,
    },
}
payload, err := hostkit.BuildFinanceReportLookupPayload(ctx, cfg, args)
```

已配置 handler 每阶段最多调用一次，error 原样返回；未配置阶段返回 typed unsupported/
needs-review payload，不伪造成功。context 原样传递，不创建 goroutine、不隐藏 deadline、不自动重试。

