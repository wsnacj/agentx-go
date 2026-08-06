# `scenes/publicnews` 中文 API Reference

成熟度：**Experimental extension**。本包随 `scenes/v0.2.0`提供，但不进入9包核心
兼容候选面；调用方必须固定精确版本，并在升级时执行 API/行为 differential。

本包是公开新闻领域的 portable source authority：提供结构化意图、来源与证据合同、
确定性 freshness/source-quality/independence guard、回答边界、LLM tool schema 以及
Pack/Workflow Definition。它不执行搜索或网络访问，不选择 provider，也不读取
credential、浏览器缓存或 HS 配置。

## Pack 与 Workflow

```go
coordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(coordinator)
_ = publicnews.RegisterInto(registry)
spec, _ := publicnews.MaterializedDefaultWorkflow(coordinator)
```

- `Definition`、`PackDefinition`：返回 caller-owned Pack Definition。
- `RegisterInto`、`RegisterPacksIntoRegistry`：注册 Definition；nil registry 保持兼容
  的 no-op 行为。
- `MaterializedDefaultWorkflow`：要求显式注入 canonical `pack.Coordinator`，因此
  Workflow admission 和 tool-argument lowering 仍由 Host 拥有。
- `PackID`、`CaseTypeLatestBrief`、`DefaultWorkflow`：稳定标识当前 portable 合同；
  它们不代表正式版本承诺。

## Tool 合同

- `LatestNewsLookupTool`、`LatestNewsExtractTool`、`LatestNewsGuardTool`：返回
  `components/llm.Tool` schema。
- `RegisterTools`、`RegisterLatestNewsLookupTool`：把 Host 显式提供的 handler 注册到
  canonical `tools.Registry`；不会创建 provider 或网络 client。
- `DecodeToolArguments`：只执行 JSON object 解码，错误保留
  `decode public-news tool arguments` 前缀。
- `ToolNames`、`SkillNames`：返回领域声明的 tool/skill identity。

## Evidence 与回答边界

- `LatestNewsLookupIntent`、`LatestNewsLookupSource`、`LatestNewsLookupPayload`：公开新闻
  查询的主合同；JSON tag 保持迁移前兼容。
- `ExtractEvidence*`、`BuildGuardPayload`、`BuildEvaluation`：对 Host 已提供的页面文本和
  来源执行确定性抽取与 guard，不抓取网页。
- `DefaultEvidenceQualityPolicy`、`DefaultSourceRelevancePolicy`：portable mechanism；
  站点 allowlist、publisher/provider 偏好仍由 Host 拥有。
- `LatestNewsLookupAnswerReadiness`、`LatestNewsLookupAnswerContract`、
  `BuildLatestNewsEvaluatorReport`：生成 bounded answer/evidence 状态，不生成或背书新闻
  事实。
- `EvidenceReviewer`：可选的 Host seam；调用方负责并发、安全和模型/provider 实现。

## 非目标

- 搜索、打开页面、浏览器、真实网络和 provider fallback；
- credential、授权、安全策略、站点白名单和地区/客户策略；
- Runner、HS registry、Scene lifecycle 或具体 product orchestration；
- Public/Beta/Stable、semver 或正式发行承诺。
