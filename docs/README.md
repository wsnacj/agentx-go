# AgentX Go 中文文档

M5S已形成并获Owner接受的Core Developer Preview Candidate技术checkpoint；P1-C当前扩展为五条标准construction路径、
48个package中文Reference、10个候选API snapshot/平台/类型闭包门禁及统一fixed-version
consumer已经闭合。M5T四module统一版本、升级/回滚说明与clean-room消费证据也已形成
技术checkpoint并获Owner接受。M6A已补齐Developer Preview兼容、维护、支持、安全报告、
版本epoch与分发预检合同并获Owner接受。M6B已把现有正文交付为可本地构建、导航、
搜索的85页Core中文Developer Portal Candidate，并形成等待Owner接受的技术checkpoint。
这些结论仍处于private validation，不构成Public、Beta、Stable、semver或正式发行。

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

1. [快速开始](quickstart.md)
2. [安装与多 Module 引用](guides/installation-and-modules.md)
3. [版本、升级与回滚](guides/versioning-and-upgrades.md)
4. [Developer Preview兼容与分发政策](guides/developer-preview-policy.md)
5. [分发Readiness与Beta阻断](reference/distribution-readiness.md)
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
30. [`extensions/astock` 推荐入口中文 API Reference](../extensions/astock/API.md)
31. [`extensions/astock/contracts` 中文 API Reference](../extensions/astock/contracts/API.md)
32. [`extensions/astock/hostkit` 中文 API Reference](../extensions/astock/hostkit/API.md)
33. [`extensions/domainmodule` 中文 API Reference](../extensions/domainmodule/API.md)
34. [`extensions/pack` 中文 API Reference](../extensions/pack/API.md)
35. [`extensions/productshell` 中文 API Reference](../extensions/productshell/API.md)
36. [`extensions/skills` 中文 API Reference](../extensions/skills/API.md)
37. [`runtime/construction` 中文 API Reference](../runtime/construction/API.md)
38. [`runtime/controlcontract` 中文 API Reference](../runtime/controlcontract/API.md)
39. [`runtime/execution` 中文 API Reference](../runtime/execution/API.md)
40. [`runtime/executionpolicy` 中文 API Reference](../runtime/executionpolicy/API.md)
41. [`runtime/channel` 中文 API Reference](../runtime/channel/API.md)
42. [`runtime/hostkit` 中文 API Reference](../runtime/hostkit/API.md)
43. [`runtime/hosthttp/hostserver` 中文 API Reference](../runtime/hosthttp/hostserver/API.md)
44. [`runtime/hosthttp/requestjson` 中文 API Reference](../runtime/hosthttp/requestjson/API.md)
45. [`runtime/hosthttp/resourcepolicy` 中文 API Reference](../runtime/hosthttp/resourcepolicy/API.md)
46. [`runtime/toolloop` 中文 API Reference](../runtime/toolloop/API.md)
47. [`runtime/workflow/composition` 中文 API Reference](../runtime/workflow/composition/API.md)
48. [`runtime/workflow/hostkit` 中文 API Reference](../runtime/workflow/hostkit/API.md)
49. [Objective Host Kit 接入指南](guides/objective-hostkit.md)
50. [`runtime/objective` 中文 API Reference](../runtime/objective/API.md)
51. [`runtime/objective/hostkit` 中文 API Reference](../runtime/objective/hostkit/API.md)
52. [`runtime/session` 中文 API Reference](../runtime/session/API.md)
53. [`runtime/session/hostkit` 中文 API Reference](../runtime/session/hostkit/API.md)

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
- [`runtime/conformance/toolloop-consumer`](../runtime/conformance/toolloop-consumer)
- [`runtime/conformance/workflow-hostkit-consumer`](../runtime/conformance/workflow-hostkit-consumer)
- [`runtime/conformance/objective-hostkit-consumer`](../runtime/conformance/objective-hostkit-consumer)
- [`runtime/conformance/session-hostkit-consumer`](../runtime/conformance/session-hostkit-consumer)
- [`runtime/conformance/run-data-plane-consumer`](../runtime/conformance/run-data-plane-consumer)
- [`extensions/conformance/astock-contract-consumer`](../extensions/conformance/astock-contract-consumer)
- [`extensions/conformance/domain-module-consumer`](../extensions/conformance/domain-module-consumer)
- [`extensions/conformance/pack-consumer`](../extensions/conformance/pack-consumer)
- [`extensions/conformance/astock-consumer`](../extensions/conformance/astock-consumer)
- [`extensions/conformance/skills-consumer`](../extensions/conformance/skills-consumer)
- [`extensions/conformance/productshell-consumer`](../extensions/conformance/productshell-consumer)
- [`extensions/conformance/productshell-observation-consumer`](../extensions/conformance/productshell-observation-consumer)

`docs/**` 的主体页面描述根 contract module；`components/llm/API.md` 描述 LLM
合同；`runtime/**/API.md` 描述已经真实落地且各自标注成熟度的 Runtime owner；
`extensions/**/API.md`描述获准迁入的 portable extension合同。
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
尚未落地的 hostless完整 Runtime construction、A股 livekit/provider、更多
components、其它 extensions和 Scene不会预先获得虚假 API 页面。
