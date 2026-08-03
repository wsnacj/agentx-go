# AgentX Go 中文文档

P1已完成Core七类能力，P2已完成Providers与通用Tools，P3已完成Browser与Document，
P4-A至P4-E已迁入首批互补Portable Scenes并完成Portfolio Checkpoint。P5已完成
Developer Preview接入、中文API导航、固定版本和HS portable source-authority closure，
没有发现需要补迁的公共能力，也没有继续按Scene逐个扩张。九个module当前共有123份中文`API.md`，覆盖全部120个可外部import
的production package；focused gate管理root/components/runtime/extensions/scenes的
77个production package和14个Developer
Preview candidate。其它可选module仍按Experimental边界独立验证。

这些结论仍处于private validation，不构成Public、Beta、Stable、semver、正式发行或
production-ready声明。API正文覆盖、兼容性承诺和正式发布成熟度是三件不同的事。

P1-A又补齐Model Conversation与Tool Direct Answer推荐路径；当前文档对应 M3E Core
Developer Preview candidate、M4A Experimental Host
HTTP owner、M5A AssetFS/首个 portable extension合同、M5B Domain Module及
已获 Owner接受的 M5C Portable Pack Core与 M5D A股 portable Domain Extension，
已获 Owner接受的 M5E Portable Skills Core与 M5F Run/Artifact Data Plane，以及
M5G已落地的 Experimental Case合同和 ProductShell Preparation owner、M5H
新增的 typed Observation / display-safe Host Handoff机制，以及M5I新增的portable
temporary Workflow planning机制、LLM组件和
已迁 Runtime owner，
目标是让调用方
能够准确判断：

1. 根 `agentx` 包现在真实提供什么；
2. `Run`、context、并发、错误和关闭行为如何工作；
3. Model Conversation、Open Tool Loop、Tool Direct Answer与 Workflow路径应该如何选择；
4. `ExecutionAdapter`、Model/Tool Adapter和 Host分别拥有什么责任；
5. Workflow Host Kit如何组合已经进入 Runtime module的 portable mechanism；
6. shared Host HTTP transport与具体 Scene HTTP API之间如何分工；
7. A股推荐 extension入口如何组合 Manifest、Pack、tool schema和 Host handler；
8. 新项目如何从目录或 immutable `fs.FS`加载 Skill，同时把 catalog/filter、安全和
   安装副作用保留在 Host；
9. 哪些能力仍是后续工作，不能从 package 名或示例中误判为完整 SDK。

建议阅读顺序：

先用[七类能力与标准接入路径](guides/capability-map.md)选择入口，再按以下顺序阅读合同和
专题文档：

1. [快速开始](quickstart.md)
2. [安装与多 Module 引用](guides/installation-and-modules.md)
3. [版本、升级与回滚](guides/versioning-and-upgrades.md)
4. [Developer Preview兼容与分发政策](guides/developer-preview-policy.md)
5. [分发Readiness与Beta阻断](reference/distribution-readiness.md)，以及
   [Pre-Beta准入合同](reference/pre-beta-admission.md)
6. [公共执行模型](concepts/execution-model.md)
7. [Go API Reference](reference/agentx.md)
8. [生命周期与错误处理](guides/lifecycle-and-errors.md)
9. [自定义 Adapter](guides/custom-adapter.md)
10. [Host Kit + Model/Tool Adapter](guides/model-tool-hostkit.md)
11. [Task / Session / Subagent Host Kit](guides/session-subagent-hostkit.md)
12. [模型对话](guides/chat.md)
13. [Tool Direct Answer](guides/tool-direct-answer.md)
14. [Workflow Host Kit](guides/workflow-hostkit.md)
15. [A 股 Portable Domain Extension](guides/astock-extension.md)
16. [Portable Skills 接入](guides/portable-skills.md)
17. [Run 与 Artifact 数据平面](guides/run-data-plane.md)
18. [ProductShell 两阶段准备](guides/product-shell-preparation.md)
19. [ProductShell Observation与Host Handoff](guides/product-shell-observation-handoff.md)
20. [Package API 索引与成熟度矩阵](reference/package-maturity.md)
21. [成熟度与兼容边界](maturity.md)
22. [HS 迁移说明](guides/hs-migration.md)
23. [`components/llm` 中文 API Reference](../components/llm/API.md)
24. [`runtime` 中文 package 导航](../runtime/README.md)
25. [`extensions` 中文 package 导航](../extensions/README.md)
26. [`runtime/assetfs` 中文 API Reference](../runtime/assetfs/API.md)
27. [`runtime/runstore` 中文 API Reference](../runtime/runstore/API.md)
28. [`runtime/artifact` 中文 API Reference](../runtime/artifact/API.md)
29. [`runtime/cases` 中文 API Reference](../runtime/cases/API.md)
30. [`scenes/astock` 推荐入口中文 API Reference](../scenes/astock/API.md)
31. [`scenes/astock/contracts` 中文 API Reference](../scenes/astock/contracts/API.md)
32. [`scenes/astock/hostkit` 中文 API Reference](../scenes/astock/hostkit/API.md)
33. [`scenes/publicnews` 中文 API Reference](../scenes/publicnews/API.md)
34. [`scenes/publicnews/hostkit` 中文 API Reference](../scenes/publicnews/hostkit/API.md)
35. [`scenes/companyresearch` 中文 API Reference](../scenes/companyresearch/API.md)
36. [`scenes/companyresearch/hostkit` 中文 API Reference](../scenes/companyresearch/hostkit/API.md)
37. [`scenes/docparse` 中文 API Reference](../scenes/docparse/API.md)
38. [`scenes/docparse/hostkit` 中文 API Reference](../scenes/docparse/hostkit/API.md)
39. [`scenes/docparse/adapters` 中文 API Reference](../scenes/docparse/adapters/API.md)
40. [`scenes/docparse/fusion` 中文 API Reference](../scenes/docparse/fusion/API.md)
41. [`scenes/docparse/planner` 中文 API Reference](../scenes/docparse/planner/API.md)
42. [`scenes/docparse/profile` 中文 API Reference](../scenes/docparse/profile/API.md)
43. [`scenes/docparse/qualityevidence` 中文 API Reference](../scenes/docparse/qualityevidence/API.md)
44. [`scenes/docparse/representation` 中文 API Reference](../scenes/docparse/representation/API.md)
45. [`scenes/docparse/understanding` 中文 API Reference](../scenes/docparse/understanding/API.md)
46. [`extensions/domainkit` 中文 API Reference](../extensions/domainkit/API.md)
47. [`extensions/domainmodule` 中文 API Reference](../extensions/domainmodule/API.md)
48. [`extensions/pack` 中文 API Reference](../extensions/pack/API.md)
49. [`extensions/productshell` 中文 API Reference](../extensions/productshell/API.md)
50. [`extensions/skills` 中文 API Reference](../extensions/skills/API.md)
51. [`runtime/construction` 中文 API Reference](../runtime/construction/API.md)
52. [`runtime/controlcontract` 中文 API Reference](../runtime/controlcontract/API.md)
53. [`runtime/execution` 中文 API Reference](../runtime/execution/API.md)
54. [`runtime/executionpolicy` 中文 API Reference](../runtime/executionpolicy/API.md)
55. [`runtime/channel` 中文 API Reference](../runtime/channel/API.md)
56. [`runtime/hostkit` 中文 API Reference](../runtime/hostkit/API.md)
57. [`runtime/hosthttp/hostserver` 中文 API Reference](../runtime/hosthttp/hostserver/API.md)
58. [`runtime/hosthttp/requestjson` 中文 API Reference](../runtime/hosthttp/requestjson/API.md)
59. [`runtime/hosthttp/resourcepolicy` 中文 API Reference](../runtime/hosthttp/resourcepolicy/API.md)
60. [`runtime/toolloop` 中文 API Reference](../runtime/toolloop/API.md)
61. [`runtime/workflow/composition` 中文 API Reference](../runtime/workflow/composition/API.md)
62. [`runtime/workflow/hostkit` 中文 API Reference](../runtime/workflow/hostkit/API.md)
63. [Objective Host Kit 接入指南](guides/objective-hostkit.md)
64. [`runtime/objective` 中文 API Reference](../runtime/objective/API.md)
65. [`runtime/objective/hostkit` 中文 API Reference](../runtime/objective/hostkit/API.md)
66. [`runtime/session` 中文 API Reference](../runtime/session/API.md)
67. [`runtime/session/hostkit` 中文 API Reference](../runtime/session/hostkit/API.md)
68. [`runtime/scheduler` 中文 API Reference](../runtime/scheduler/API.md)
69. [`runtime/session/resume` 中文 API Reference](../runtime/session/resume/API.md)
70. [`providers` 中文 API Reference](../providers/API.md)
71. [`providers/openaicompat` 中文 API Reference](../providers/openaicompat/API.md)
72. [`providers/anthropic` 中文 API Reference](../providers/anthropic/API.md)
73. [`providers/codex` 中文 API Reference](../providers/codex/API.md)
74. [`components/tool` 中文 API Reference](../components/tool/API.md)
75. [`tools` 中文 API Reference](../tools/API.md)
76. [`tools/diffs` 中文 API Reference](../tools/diffs/API.md)
77. [`browser` 中文 API 总览](../browser/API.md)
78. [`browser/runtime` 中文 API Reference](../browser/runtime/API.md)
79. [`scenes/browserops` 中文 API Reference](../scenes/browserops/API.md)
80. [`scenes/browserops/hostkit` 中文 API Reference](../scenes/browserops/hostkit/API.md)
81. [`scenes/publictransport` 中文 API Reference](../scenes/publictransport/API.md)
82. [`scenes/publicsource` 中文 API Reference](../scenes/publicsource/API.md)
83. [`scenes/wechatarticle` 中文 API Reference](../scenes/wechatarticle/API.md)
84. [`document` 中文 API 总览](../document/API.md)
85. [`document/ocr` 中文 API Reference](../document/ocr/API.md)
86. [`document/pdf` 中文 API Reference](../document/pdf/API.md)
87. [`document/pipeline` 中文 API Reference](../document/pipeline/API.md)
88. [`document/tools` 中文 API Reference](../document/tools/API.md)
89. [`scenes/globalstock` 中文 API Reference](../scenes/globalstock/API.md)
90. [`scenes/globalstock/hostkit` 中文 API Reference](../scenes/globalstock/hostkit/API.md)
91. [`scenes/finance` 中文 API Reference](../scenes/finance/API.md)
92. [`scenes/finance/metrics` 中文 API Reference](../scenes/finance/metrics/API.md)
93. [`scenes/finance/brief` 中文 API Reference](../scenes/finance/brief/API.md)
94. [`scenes/finance/hostkit` 中文 API Reference](../scenes/finance/hostkit/API.md)

可运行验证位于：

- [`examples/contract-basic`](../examples/contract-basic)
- [`examples/custom-adapter`](../examples/custom-adapter)
- [`conformance/consumer`](../conformance/consumer)
- [`runtime/conformance/protocol-consumer`](../runtime/conformance/protocol-consumer)
- [`runtime/conformance/construction-consumer`](../runtime/conformance/construction-consumer)
- [`runtime/conformance/execution-consumer`](../runtime/conformance/execution-consumer)
- [`runtime/conformance/channel-consumer`](../runtime/conformance/channel-consumer)
- [`runtime/conformance/hostkit-consumer`](../runtime/conformance/hostkit-consumer)
- [`runtime/conformance/controlcontract-consumer`](../runtime/conformance/controlcontract-consumer)
- [`scenes/conformance/browserops-consumer`](../scenes/conformance/browserops-consumer)
- [`scenes/conformance/publictransport-consumer`](../scenes/conformance/publictransport-consumer)
- [`runtime/conformance/toolloop-consumer`](../runtime/conformance/toolloop-consumer)
- [`runtime/conformance/workflow-hostkit-consumer`](../runtime/conformance/workflow-hostkit-consumer)
- [`runtime/conformance/objective-hostkit-consumer`](../runtime/conformance/objective-hostkit-consumer)
- [`runtime/conformance/session-hostkit-consumer`](../runtime/conformance/session-hostkit-consumer)
- [`runtime/conformance/run-data-plane-consumer`](../runtime/conformance/run-data-plane-consumer)
- [`scenes/conformance/astock-consumer`](../scenes/conformance/astock-consumer)
- [`scenes/conformance/research-consumer`](../scenes/conformance/research-consumer)
- [`scenes/conformance/docparse-consumer`](../scenes/conformance/docparse-consumer)
- [`extensions/conformance/domain-module-consumer`](../extensions/conformance/domain-module-consumer)
- [`extensions/conformance/pack-consumer`](../extensions/conformance/pack-consumer)
- [`extensions/conformance/skills-consumer`](../extensions/conformance/skills-consumer)
- [`extensions/conformance/productshell-consumer`](../extensions/conformance/productshell-consumer)
- [`extensions/conformance/productshell-observation-consumer`](../extensions/conformance/productshell-observation-consumer)
- [`providers/conformance/openaicompat-consumer`](../providers/conformance/openaicompat-consumer)
- [`providers/conformance/provider-cohort-consumer`](../providers/conformance/provider-cohort-consumer)
- [`tools/conformance/catalog-diffs-consumer`](../tools/conformance/catalog-diffs-consumer)
- [`tools/conformance/general-tools-consumer`](../tools/conformance/general-tools-consumer)
- [`browser/conformance/browser-platform-consumer`](../browser/conformance/browser-platform-consumer)
- [`document/conformance/ocr-consumer`](../document/conformance/ocr-consumer)
- [`document/conformance/pdf-consumer`](../document/conformance/pdf-consumer)
- [`document/conformance/pipeline-consumer`](../document/conformance/pipeline-consumer)
- [`document/conformance/tools-consumer`](../document/conformance/tools-consumer)
- [`scenes/conformance/sourceacquisition-consumer`](../scenes/conformance/sourceacquisition-consumer)
- [`scenes/conformance/finance-consumer`](../scenes/conformance/finance-consumer)

`docs/**` 的主体页面描述根 contract module；`components/llm/API.md` 描述 LLM
合同；`runtime/**/API.md` 描述已经真实落地且各自标注成熟度的 Runtime owner；
`extensions/**/API.md`描述获准迁入的 portable extension合同。
`providers/**/API.md`描述可选provider module；它不属于当前五module focused API gate，
但自身必须完成fixed consumer、test/race/vet/tidy/list和无HS反向依赖验证。
`tools/**/API.md`描述可选通用tool module；它同样不属于Core API gate，必须独立证明
固定版本消费、行为合同和无HS/Runner/Scene反向依赖。
`browser/**/API.md`描述可选重型Browser module；P3已完成runtime、browserd host、tools、
fixed consumer与HS cutover。`document/**/API.md`描述Document推荐高层入口；低层OCR/
Pipeline实现包仍按internalization candidate治理，不能因可import自动视为稳定API。
`scenes/**/API.md`描述P4已选定的portable Domain Kits；P4 Portfolio Checkpoint已经停止
继续按Scene逐个迁移。
M5D只批准且已完成 A股 portable Manifest/assets/tool schema/Pack/hostkit切片。M5E
只迁入 Skill contracts、loader/cache、activation、requested semantics与 resource
refs；prompt catalog/filter、安全规则、安装执行、bundled内容和 Runner仍由 Host拥有。
M5F只收口 Run/NodeExecution/Event与 Artifact registry/lineage的 portable合同和
内存实现；fixed consumer与 HS production cutover已经完成，durable backend、文件/
对象存储、安全写入、保留期与产品投影仍由 Host拥有。
M5G新增 `runtime/cases`和 `extensions/productshell`两个 Experimental owner：前者只
固定 Case合同/存储port，后者只拥有无副作用输入投影与 Host port驱动的固定准备顺序。
自然语言规划、LLM/provider、Pack/Workflow/Case backend、Objective、Scene和产品策略
继续由 Host拥有；`PrepareResult`也不代表已经执行 Workflow。
M5H继续在同一个 `extensions/productshell` Experimental package内增加typed session/
host-process/operator-line observation与display-safe Host UI handoff；raw parser、完整
ObservationSnapshot、inventory、readback和真实delivery继续由Host拥有。
M5I继续在同一package内增加临时Workflow typed plan、prompt/schema构造、有限重试、
binding lowering、Workflow Spec构造和固定`Should → Resolve → Apply`阶段；具体model/
provider、tool policy、validator policy、execution snapshot和Workflow执行继续由Host拥有。
尚未落地的hostless完整Runtime construction、默认provider/credential、产品backend与
其余HS业务Scene不会预先获得虚假API页面。示例与fixed consumer的职责和映射见
[`examples/README.md`](../examples/README.md)。
