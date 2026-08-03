# `scenes/wechatarticle` 中文 API Reference

成熟度：**Experimental extension**。当前没有 semver、Developer Preview、Public、Beta 或
Stable 兼容承诺；调用方应固定 pseudo-version 并执行 API/行为 differential。

本包是公众号文章只读采集的 portable Domain Kit：拥有 login/account/article/dedup/download
typed contracts、显式 Host `Client` port、provider-neutral Coordinator、display-safe evidence
projection、确定性 evaluator 与 Pack/Workflow。它不实现 exporter HTTP client、登录二维码、
credential/cookie、redirect、filesystem 或 artifact writer。

## 最小接入

```go
coordinator := wechatarticle.NewCoordinator(hostClient)
coordinator.AccountKeyword = "公众号名称"
coordinator.DownloadFirst = true

execution := coordinator.Execute(ctx, runtimeRequest)
evaluation := wechatarticle.Evaluate(execution, true)
```

- `Client`：Host 实现 `CheckLogin`、`SearchAccounts`、`ListArticles` 和
  `DownloadArticle`；canonical 不接触 endpoint、auth key、basic auth 或 cookie；
- `Coordinator`、`NewCoordinator`、`Coordinator.Execute`：按显式 strategy 执行 login probe、
  account search、article list 与可选 first download；每个步骤最多调用一次；
- `Execution`：保留 typed account/article/download readback；AgentX generic result 只包含
  count、digest 与 display-safe refs，不包含 credential/cookie；
- `ErrorClassifier`、`Failure`：由 HS/Host 把具体 exporter error 映射到 typed
  failure class、missing input 与 next action；canonical 不依赖具体 error package。

## 合同与 identity

- `LoginStatus`、`Account`、`Article`、`ArticleDedupKey`、`ArticleListResult`、
  `DownloadResult`、`SyncOptions`、`SyncResult`：当前 typed 数据合同；
- `Descriptor`、`Registry`与 `DefaultAdapterRef` / `Strategy*` / `Evidence*` 常量：
  用于 AgentX control-contract 组合；
- `NormalizeDownloadFormat`、`ClampPageSize`：无 I/O 的有界 normalization。

`DownloadResult.Body` 只在 Host Client 与调用方之间传递，JSON 不序列化；Coordinator 仅把
SHA-256 digest 和字符数投影为 evidence。artifact 路径、写入和保留策略属于 Host。

## 评估与 Pack

- `Evaluate(execution, requireDownload)` 确定性验证 login、article list、download 与 evidence；
- `Definition` / `PackDefinition`、`RegisterInto`、`MaterializedDefaultWorkflow` 提供
  一个只读、一次 tool-call 上限的 Pack/Workflow；
- `PackID`、`CaseTypeAcquire`、`DefaultWorkflow`、`AcquireTool` 是当前 portable identity。

## 取消、并发与错误

Coordinator 原样传递 `context.Context`，不创建 goroutine、不自动重试。同一 Coordinator
只在调用方不并发修改其字段且 Host Client 支持并发时可并发使用。未注入 error
classifier 时保守投影为 `external_dependency_unavailable`；登录失效、缺 client、缺输入和
strategy 不匹配都 fail closed。

## 非目标

- exporter HTTP/API 实现、basic auth/auth key、cookie、secret file 与环境变量；
- 登录二维码创建/轮询、redirect policy、真实网络、rate limit 和 provider readiness；
- artifact/filesystem、产品回复、HTTP/CLI、Objective projection、Public/Beta/Stable 或正式发行。

