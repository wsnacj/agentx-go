# `scenes/publicsource` 中文 API Reference

成熟度：**Experimental extension**。当前没有 semver、Developer Preview、Public、Beta 或
Stable 兼容承诺；调用方应固定 pseudo-version 并在升级时执行 API/行为 differential。

本包是通用公开来源只读采集的 portable source authority：拥有 typed request/report/evidence/
display-summary合同、Host-supplied source policy 机制、provider-neutral Collector、exact-once
Coordinator、search/document 投影、确定性 evaluator 和 Pack/Workflow。它不访问网络，
不选择 provider，不读取 credential，也不保存 raw page body。

## 最小接入

```go
type collector struct{}

func (collector) CollectPublicSourceEvidence(
    ctx context.Context,
    req publicsource.Request,
) (publicsource.Report, error) {
    // Host 在这里接入 HTTP、Browser、fixture、cache 或其他 service。
    return publicsource.Report{/* display-safe evidence */}, nil
}

coordinator := publicsource.NewCoordinator(collector{})
execution := coordinator.Execute(ctx, runtimeRequest)
```

- `Collector`：唯一运行时 port；每次 `Execute` 最多调用一次；
- `Request`：只携带 runtime request 与 display-safe query/source/policy refs；
- `Report`、`Evidence`、`DisplaySummary`：无 provider 的读回合同；
- `Report.Normalize`：对 missing evidence、raw output 和弱摘要 fail closed；
- `Coordinator`、`Execution`：返回 generic control-contract result 与领域 report。

## Search / Document Host Kit

- `SearchExecutor`、`SearchCollector`：Host 只需返回 `SearchPayload` 与可选
  `DisplaySummary`；canonical 负责 allowlist 过滤、证据 identity 与状态投影；
- `DocumentFetcher`、`DocumentCollector`：Host 返回无 header/cookie/provider diagnostics 的
  `DocumentExecution`；raw URL/text 不进入 control-contract result；
- `SourcePolicy`、`SourcePolicy.CheckURL`、`SourcePolicy.Filter`：实现 HTTPS、host、
  host suffix 与 URL prefix 匹配；规则值始终由 Host 显式注入；
- `BuildReportFromSearch`、`QueryResultEvidenceRef`：纯函数投影，不发起网络。

## 评估与 Pack

- `Evaluate(report, requireAttestation)` 验证 evidence、display summary、attestation 与
  raw-output 边界；
- `Definition` / `PackDefinition`、`RegisterInto`、`MaterializedDefaultWorkflow` 提供
  一个只读、一次 tool-call 上限的 Pack/Workflow；
- `PackID`、`CaseTypeAcquire`、`DefaultWorkflow`、`AcquireTool` 是当前 portable identity。

## 取消、并发与错误

Coordinator 原样传递 `context.Context`，不创建 goroutine、不隐藏 deadline、不自动重试。
只要 Host Collector 支持并发，同一 Coordinator 可并发使用。Collector error 投影为
`external_dependency_unavailable`；缺 Collector、request 未 ready、identity 不匹配均返回
typed blocked result。

## 非目标

- HTTP client、DNS/private-host policy、Browser、provider selection/fallback、credential与真实网络；
- raw page cache、本地文件、内容授权、登录态、付费订阅或业务来源偏好；
- HS Objective/HTTP/CLI、产品回复本地化、Public/Beta/Stable 或正式发行。

