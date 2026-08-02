# `scenes/companyresearch` 中文 API Reference

成熟度：**Developer Preview candidate**。当前没有 semver、Public、Beta 或 Stable
兼容承诺；调用方必须固定 pseudo-version，并在升级时执行 API/行为 differential。

本包是公司研究领域的 portable source authority：提供单公司/多公司意图与结果合同、
主体解析 seam、任务分解、证据 readiness、回答边界、确定性 evaluator、LLM tool schema
以及 Pack/Workflow Definition。它不查询财报、行情或新闻，不选择 provider，不读取
credential，也不生成投资建议。

## Pack 与 Workflow

```go
coordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(coordinator)
_ = companyresearch.RegisterInto(registry)
single, _ := companyresearch.MaterializedDefaultWorkflow(coordinator)
compare, _ := companyresearch.MaterializedCompareWorkflow(coordinator)
```

- `Definition`：返回 caller-owned Pack Definition。
- `RegisterInto`：注册 Definition；nil registry 保持兼容的 no-op 行为。
- `MaterializedDefaultWorkflow`、`MaterializedCompareWorkflow`：要求显式 canonical
  `pack.Coordinator`，不内置 HS validation/lowering policy。
- `PackID`、`DefaultWorkflow`、`CompareWorkflow`、case/tool/skill 常量：标识当前
  portable 合同，但不代表正式版本承诺。

## Tool 与数据合同

- `CompanyResearchIntent`、`CompanyResearchEvidence`、`CompanyResearchPayload`：主数据
  合同；JSON tag 保持迁移前兼容。
- `CompanyResearchLookupTool`、`CompanyCompareLookupTool`、
  `CompanyResearchGuardTool`：返回 `components/llm.Tool` schema。
- `RegisterStandardTools`：只注册 Host 显式提供的 lookup/compare/guard handler。
- `IntentFromParams`、`ParamsFromIntent`：确定性参数转换，不解析自然语言。
- `SubjectResolutionRequestFromIntent`：建立主体解析请求；真实解析来源由 Host 注入。

## Task force 与 evaluator

- `BuildCompanyResearchTaskPlan`：按 subject→finance/market/news/risk→guard→synthesis
  建立确定性任务图，不启动 goroutine、scheduler 或子 Session。
- `TaskResultFromEvidence`、`TaskResultFromReadiness`、
  `TaskResultFromAnswerContract`：把 Host 已提供的结果投影为 task status。
- `EvaluateCompanyResearchEvidence`：检查主体、维度、来源、freshness 和回答边界；
  它不拉取证据，也不作投资结论。

## 非目标

- finance/stock/news 的 concrete executor、provider、网络和 RunStore；
- credential、授权、安全策略、客户规则和投资/合规判断；
- Runner、HS registry、Scene lifecycle 或完整长任务调度；
- Public/Beta/Stable、semver 或正式发行承诺。
