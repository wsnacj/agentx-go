# Package API 索引与成熟度矩阵

本页只评估当前root、components、runtime、extensions与scenes五个module中实际存在的
77个 production package。它不是历史
surface inventory，也不会把任何符号自动升级为 Public、Beta 或 Stable。

## 分级含义

| 分级 | 本阶段含义 | 兼容性含义 |
| --- | --- | --- |
| Developer Preview candidate | M3E 选定的 Core标准路径或 M5D选定的 A股推荐入口直接需要，签名和中文 Reference 进入 focused gate | 仅表示候选；当前没有 semver 或长期兼容承诺 |
| Experimental extension | 已有真实 implementation 和 consumer，但仍可能在 Beta 前调整 owner 或入口 | 调用方应固定伪版本并评估升级差异 |
| internalization candidate | 当前属于低层迁移 owner，或已位于 `internal`并由上层 Facade隐藏 | 新项目不得或不应直接依赖 |

## 当前矩阵

| Package | 分级 | 中文 Reference | 主要用途 |
| --- | --- | --- | --- |
| `agentx` | Developer Preview candidate | [根 API](agentx.md) | Client、Run、错误、画像和生命周期 |
| `components/llm` | Developer Preview candidate | [LLM API](../../components/llm/API.md) | provider-neutral 模型、消息、工具和 usage 合同 |
| `components/tool` | Experimental extension | [Tool Contract API](../../components/tool/API.md) | provider-neutral工具声明、调用、结果与执行合同 |
| `runtime/execution` | Developer Preview candidate | [Execution API](../../runtime/execution/API.md) | 根 Client 与 Host 之间的 typed dispatch |
| `runtime/executionpolicy` | Experimental extension | [API](../../runtime/executionpolicy/API.md) | 执行policy DTO、Host编译port与metadata/retry/decision/report reducer |
| `runtime/channel` | Experimental extension | [API](../../runtime/channel/API.md) | portable message/sender ports、bounded ingress与session delivery合同 |
| `runtime/hostkit` | Developer Preview candidate | [Host Kit API](../../runtime/hostkit/API.md) | Model Conversation、Model/Tool Adapter 与 Tool Direct Answer 构造 |
| `runtime/assetfs` | Experimental extension | [API](../../runtime/assetfs/API.md) | immutable asset snapshot、fingerprint与 resolver |
| `runtime/artifact` | Experimental extension | [API](../../runtime/artifact/API.md) | Artifact identity、lineage、registry合同与并发安全内存实现 |
| `scenes/astock` | Developer Preview candidate | [A股 Extension API](../../scenes/astock/API.md) | A股 Manifest、assets、tool schema、Pack catalog与 evaluator推荐入口 |
| `scenes/astock/contracts` | Experimental extension | [API](../../scenes/astock/contracts/API.md) | A股 portable DTO、JSON normalization与 assessment |
| `scenes/astock/hostkit` | Experimental extension | [API](../../scenes/astock/hostkit/API.md) | 无 provider的 A股 intent、handler协调、readiness与回答格式化 |
| `scenes/astock/internal/packresearch` | internalization candidate | [API](../../scenes/astock/internal/packresearch/API.md) | Research Pack Definition与 evaluator内部 owner |
| `scenes/astock/internal/packsignal` | internalization candidate | [API](../../scenes/astock/internal/packsignal/API.md) | Signal Pack Definition与 evaluator内部 owner |
| `scenes/astock/internal/packvaluation` | internalization candidate | [API](../../scenes/astock/internal/packvaluation/API.md) | Valuation Pack Definition与 evaluator内部 owner |
| `scenes/browserops` | Developer Preview candidate | [API](../../scenes/browserops/API.md) | Browser Ops Pack、证据投影与确定性 evaluator |
| `scenes/browserops/hostkit` | Experimental extension | [API](../../scenes/browserops/hostkit/API.md) | 显式Host tool executor驱动的Browser Ops协调层 |
| `scenes/publictransport` | Experimental extension | [API](../../scenes/publictransport/API.md) | 公共交通只读合同、provider-neutral协调、证据evaluator与Pack；依赖Experimental controlcontract |
| `scenes/publicnews` | Developer Preview candidate | [API](../../scenes/publicnews/API.md) | 公开新闻合同、Pack、证据质量与确定性回答投影 |
| `scenes/publicnews/hostkit` | Experimental extension | [API](../../scenes/publicnews/hostkit/API.md) | 显式Search/Fetch ports驱动的无provider协调 |
| `scenes/publicsource` | Experimental extension | [API](../../scenes/publicsource/API.md) | 通用公开来源合同、source policy、search/document投影、协调、evaluator与Pack |
| `scenes/wechatarticle` | Experimental extension | [API](../../scenes/wechatarticle/API.md) | 公众号文章typed合同、Host Client协调、evidence、evaluator与Pack |
| `scenes/companyresearch` | Developer Preview candidate | [API](../../scenes/companyresearch/API.md) | 公司研究任务、证据guard、Pack与结果投影 |
| `scenes/companyresearch/hostkit` | Experimental extension | [API](../../scenes/companyresearch/hostkit/API.md) | 显式研究数据ports驱动的无provider协调 |
| `scenes/docparse` | Developer Preview candidate | [API](../../scenes/docparse/API.md) | 文档解析Pack、资产、tool schema与推荐入口 |
| `scenes/docparse/hostkit` | Experimental extension | [API](../../scenes/docparse/hostkit/API.md) | 显式Parse/ResultLoader ports驱动的文档协调与结果投影 |
| `scenes/docparse/adapters` | internalization candidate | [API](../../scenes/docparse/adapters/API.md) | portable字段与表格适配机制 |
| `scenes/docparse/fusion` | internalization candidate | [API](../../scenes/docparse/fusion/API.md) | 多来源文档结果融合机制 |
| `scenes/docparse/planner` | internalization candidate | [API](../../scenes/docparse/planner/API.md) | profile驱动的文档路由规划 |
| `scenes/docparse/profile` | internalization candidate | [API](../../scenes/docparse/profile/API.md) | 文档profile探测与规范化 |
| `scenes/docparse/qualityevidence` | internalization candidate | [API](../../scenes/docparse/qualityevidence/API.md) | 质量证据与评估投影 |
| `scenes/docparse/representation` | internalization candidate | [API](../../scenes/docparse/representation/API.md) | provider-neutral Document/Page表征 |
| `scenes/docparse/understanding` | internalization candidate | [API](../../scenes/docparse/understanding/API.md) | 文档理解与review-required判定机制 |
| `scenes/globalstock` | Experimental extension | [API](../../scenes/globalstock/API.md) | 港股/美股只读Pack、Workflow、tool identity与evaluator入口 |
| `scenes/globalstock/contracts` | Experimental extension | [API](../../scenes/globalstock/contracts/API.md) | HK/US证券、行情、证据、identity与readiness合同 |
| `scenes/globalstock/hostkit` | Experimental extension | [API](../../scenes/globalstock/hostkit/API.md) | 显式handler ports驱动的provider-neutral investigation协调 |
| `scenes/finance` | Experimental extension | [API](../../scenes/finance/API.md) | 财报合同、period/metric/evidence/readiness机制与Pack组合入口 |
| `scenes/finance/metrics` | Experimental extension | [API](../../scenes/finance/metrics/API.md) | 财报指标Pack、字段来源guard和确定性evaluator |
| `scenes/finance/brief` | Experimental extension | [API](../../scenes/finance/brief/API.md) | 财报简报Pack、evidence与确定性evaluator |
| `scenes/finance/hostkit` | Experimental extension | [API](../../scenes/finance/hostkit/API.md) | candidates→metrics→guard→optional brief协调 |
| `extensions/domainkit` | Experimental extension | [API](../../extensions/domainkit/API.md) | 无模型module/tool dispatch、typed error与deterministic output digest |
| `extensions/domainmodule` | Experimental extension | [API](../../extensions/domainmodule/API.md) | portable manifest、config、diagnostics与顺序注册编排 |
| `extensions/pack` | Experimental extension | [API](../../extensions/pack/API.md) | Pack定义、显式校验、注册、选择、物化与 Binding |
| `extensions/productshell` | Experimental extension | [API](../../extensions/productshell/API.md) | 输入投影与准备顺序、临时Workflow planning、typed observation及display-safe Host handoff |
| `extensions/skills` | Experimental extension | [API](../../extensions/skills/API.md) | Skill合同、加载、缓存、activation、semantics与资源引用机制 |
| `runtime/hosthttp/hostserver` | Experimental extension | [API](../../runtime/hosthttp/hostserver/API.md) | Host-deployed HTTP transport、request identity 与有界关闭 |
| `runtime/hosthttp/requestjson` | Experimental extension | [API](../../runtime/hosthttp/requestjson/API.md) | 有界严格 JSON 请求解码 |
| `runtime/hosthttp/resourcepolicy` | Experimental extension | [API](../../runtime/hosthttp/resourcepolicy/API.md) | Host 资源 allowlist 与预算收窄 |
| `runtime/toolloop` | Developer Preview candidate | [Tool Loop API](../../runtime/toolloop/API.md) | portable 多轮驱动、phase、检测和 assembly |
| `runtime/workflow` | Developer Preview candidate | [API](../../runtime/workflow/API.md) | Workflow Spec 数据合同与 Host admission seam |
| `runtime/workflow/hostkit` | Developer Preview candidate | [Workflow Host Kit API](../../runtime/workflow/hostkit/API.md) | Workflow标准构造入口；组合 portable lowering、journal、node execution和orchestration |
| `runtime/budget` | Experimental extension | [API](../../runtime/budget/API.md) | 预算判定 mechanism |
| `runtime/cases` | Experimental extension | [API](../../runtime/cases/API.md) | Case数据合同、规范化/复制 helper与最小Store port |
| `runtime/construction` | Experimental extension | [API](../../runtime/construction/API.md) | 高级 Host 构造生命周期 |
| `runtime/controlcontract` | Experimental extension | [API](../../runtime/controlcontract/API.md) | 执行控制合同、Objective kernel、Host effect、adapter/ingress/closeout、durable projection、trace-backed final answer及delegation coordination |
| `runtime/mediaartifact` | Experimental extension | [API](../../runtime/mediaartifact/API.md) | 媒体产物描述合同 |
| `runtime/objective` | Experimental extension | [API](../../runtime/objective/API.md) | Objective推荐类型入口；保持kernel类型与JSON identity |
| `runtime/objective/hostkit` | Developer Preview candidate | [API](../../runtime/objective/hostkit/API.md) | Managed ingress、显式Host dispatch、observation normalization与verification标准入口 |
| `runtime/session` | Experimental extension | [API](../../runtime/session/API.md) | Task/Session/Subagent identity、delegation与parent verification推荐命名边界 |
| `runtime/session/hostkit` | Developer Preview candidate | [API](../../runtime/session/hostkit/API.md) | child worker闭环与bounded Scheduler/Resume标准入口 |
| `runtime/session/resume` | Experimental extension | [API](../../runtime/session/resume/API.md) | continuation readback、Host wake dispatch与bounded service机制 |
| `runtime/promptcontext` | Experimental extension | [API](../../runtime/promptcontext/API.md) | prompt context 投影 |
| `runtime/protocol` | Experimental extension | [API](../../runtime/protocol/API.md) | Runtime wire/schema |
| `runtime/runstore` | Experimental extension | [API](../../runtime/runstore/API.md) | Run、NodeExecution、Event存储合同、投影与并发安全内存实现 |
| `runtime/scheduler` | Experimental extension | [API](../../runtime/scheduler/API.md) | portable queue、dispatcher、lease heartbeat、终态协调与metrics |
| `runtime/telemetry` | Experimental extension | [API](../../runtime/telemetry/API.md) | portable event 与 replay |
| `runtime/telemetry/safeerror` | Experimental extension | [API](../../runtime/telemetry/safeerror/API.md) | observation-safe error 投影 |
| `runtime/toolerrors` | Experimental extension | [API](../../runtime/toolerrors/API.md) | 结构化工具参数错误 |
| `runtime/workflow/composition` | Experimental extension | [API](../../runtime/workflow/composition/API.md) | lower→run 组合入口 |
| `runtime/workflow/journal` | Experimental extension | [API](../../runtime/workflow/journal/API.md) | durable 顺序与 host port |
| `runtime/workflow/lowering` | Experimental extension | [API](../../runtime/workflow/lowering/API.md) | portable lowering |
| `runtime/workflow/validation` | Experimental extension | [API](../../runtime/workflow/validation/API.md) | structural validation 与 policy port |
| `runtime/workflow/bindingstate` | internalization candidate | [API](../../runtime/workflow/bindingstate/API.md) | binding/state mechanism |
| `runtime/workflow/nodeexec` | internalization candidate | [API](../../runtime/workflow/nodeexec/API.md) | node execution coordination |
| `runtime/workflow/orchestration` | internalization candidate | [API](../../runtime/workflow/orchestration/API.md) | lowered-plan orchestration 内核 |
| `runtime/workflow/schema` | internalization candidate | [API](../../runtime/workflow/schema/API.md) | schema normalization/validation 内核 |
| `runtime/workflow/transition` | internalization candidate | [API](../../runtime/workflow/transition/API.md) | state transition 内核 |

## 可选 Providers module

下列8个package不进入当前root/components/runtime/extensions/scenes五module的52包focused
gate，全部保持Experimental：

| Package | 分级 | 中文 Reference | 主要用途 |
| --- | --- | --- | --- |
| `providers` | Experimental extension | [API](../../providers/API.md) | provider错误与capability sentinel |
| `providers/openaicompat` | Experimental extension | [API](../../providers/openaicompat/API.md) | OpenAI-compatible HTTP与SSE真实client |
| `providers/anthropic` | Experimental extension | [API](../../providers/anthropic/API.md) | Anthropic Messages真实client |
| `providers/codex` | Experimental extension | [API](../../providers/codex/API.md) | Codex Responses/SSE真实client |
| `providers/transport` | Experimental extension | [API](../../providers/transport/API.md) | request settings、headers与hooks |
| `providers/fault` | Experimental extension | [API](../../providers/fault/API.md) | 稳定错误分类与retryability |
| `providers/retry` | Experimental extension | [API](../../providers/retry/API.md) | bounded context-aware retry |
| `providers/usage` | Experimental extension | [API](../../providers/usage/API.md) | usage collector port |

providers module使用独立fixed pseudo-version与consumer验证；不因出现在本表而升级为
Developer Preview candidate、Public、Beta或Stable。

## 可选 Tools module

下列tools module package不进入当前五module的52包focused gate，全部保持Experimental：

| Package | 分级 | 中文 Reference | 主要用途 |
| --- | --- | --- | --- |
| `tools` | Experimental extension | [API](../../tools/API.md) | 并发安全catalog、稳定definition投影与保守名称修复 |
| `tools/diffs` | Experimental extension | [API](../../tools/diffs/API.md) | 无文件、Git或网络副作用的纯文本diff |

tools module使用独立fixed pseudo-version与consumer验证；不拥有授权、sandbox或具体
backend，也不因出现在本表而升级为Developer Preview candidate、Public、Beta或Stable。

机器可检查的同源清单位于
[`developer-preview-packages.tsv`](developer-preview-packages.tsv)。它只服务当前
77个 package 的覆盖与漂移门禁，不是新的全仓 API registry。

## 漂移门禁

```bash
GOWORK=off go run scripts/check_developer_preview_api.go
GOWORK=off go run scripts/check_developer_preview_api.go -check-platforms
GOWORK=off go run scripts/check_docs_links.go
```

门禁会确认：

1. 五个 module 当前 production package 与矩阵一一对应；
2. 每个 package 都有非空中文 Reference；
3. 十四个 Developer Preview candidate 的 `go doc -all` 哈希与可读快照未漂移；
4. 候选公开类型不泄漏 `hs/`、Go `internal`包或不推荐入口
   `runtime/controlcontract`；
5. darwin/arm64与linux/amd64的CGO-disabled候选签名一致；
6. 中文正文中扫描到的仓库本地链接均可解析且不越出仓库。

更新候选签名必须同时完成 focused owner/consumer review、中文 Reference 修订和
baseline及`docs/reference/api-snapshots/`更新。`-check-platforms`只把目标平台参数
传给子级`go list`/`go doc`，gate程序本身仍在当前Host执行，因此不会执行交叉编译
产物。上述门禁不生成完整文档站，也不替代未来semver/API compatibility工具。
