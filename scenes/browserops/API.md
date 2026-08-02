# `scenes/browserops` 中文 API Reference

成熟度：**Developer Preview candidate**。当前没有 semver、Public、Beta 或 Stable 兼容承诺；
调用方应固定 pseudo-version，并在升级时执行 API 与行为 differential。

本包是 Browser Ops Domain Kit 的 portable source authority：提供 Pack/Workflow Definition、
浏览器证据结构、状态投影与确定性 evaluator。它组合 canonical Browser/Runtime 能力，但不启动
浏览器、不访问网络、不读取文件，也不拥有 profile、登录态、凭据、审批、站点或副作用策略。

## Pack 与 Workflow

```go
coordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(coordinator)
_ = browserops.RegisterInto(registry)
spec, _ := browserops.MaterializedDefaultWorkflow(coordinator)
```

- `Definition`：返回 caller-owned `extensions/pack.Definition`，包含表单提交、失败证据、页面状态、
  结构化提取、站内搜索和下载六条 Workflow；
- `RegisterInto`：注册到显式 canonical Pack registry；nil registry 保持 no-op；
- `MaterializedDefaultWorkflow`：要求显式 `pack.Coordinator`，Host继续拥有 Workflow validation
  与 tool-argument lowering；
- `PackID`、`DefaultWorkflow` 及其余 Workflow 常量：portable identity，不构成正式版本承诺。

## Evidence 合同与状态投影

- `BrowserEvidenceBundle`：聚合 final URL、page snapshot、screenshot、action trace、downloaded file、
  task plan、failure reason 和 artifact ref；
- `BuildBrowserEvidenceBundleFromState`：从已有 Workflow state 做纯内存投影，不读 artifact 文件；
- `EvaluateBrowserEvidenceReadiness`：按 `BrowserEvidenceRequirements` 生成 score、ready flags、
  evidence 和 failure reasons；
- `BrowserArtifactType*`、`BrowserEvidenceKind*`：稳定的 portable evidence identity。

## 确定性 Evaluator

- `EvaluateBrowserVisualEvidenceGate`：校验 snapshot、screenshot 与 final URL；
- `EvaluateBrowserActionFailurePayloadEvidence`：校验失败 actionability payload、trace 与 snapshot；
- `EvaluateBrowserPageStateEvidence`：校验页面 required/forbidden text、URL 与截图；
- `EvaluateBrowserStructuredDataEvidence`：校验字段提取、页面证据与 URL；
- `EvaluateBrowserSiteSearchEvidence`：校验查询、提交动作、结果与页面证据；
- `EvaluateBrowserDownloadFileEvidence`：校验下载状态、文件元数据、URL 与截图；
- 对应 `*Schema` 函数返回 evaluator output schema。所有 evaluator 只处理调用方提供的数据，
  不执行重试、网络、浏览器或文件操作。

## 并发、取消与错误

`Definition` 与 evaluator 不保存可变共享状态，可由调用方并发调用。它们不创建后台 goroutine。
Pack materialization 的取消/并发边界由调用方注入的 Coordinator 和 Host 实现承担。当前错误文本与
evidence/failure reason 顺序纳入迁移 differential，但在正式版本冻结前仍属于 Developer Preview。

## 非目标

- browser backend、browserd、profile、tab/session、登录态与 credential；
- authorization、sandbox、approval、download/file allowlist、retention 和站点策略；
- live capture、artifact 持久化、真实网络、副作用执行和产品结果权威；
- Runner、HS registry、旧 Scene lifecycle、Public/Beta/Stable、semver 或正式发行。
