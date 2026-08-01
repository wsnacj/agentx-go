# AgentX Go 中文文档

当前文档对应 M3E Core Developer Preview candidate、M4A Experimental Host
HTTP owner、M5A AssetFS/首个 portable extension合同、M5B Domain Module及
已获 Owner接受的 M5C Portable Pack Core，以及当前 active的 M5D A股 portable
Domain Extension、LLM组件和已迁 Runtime owner，目标是让调用方
能够准确判断：

1. 根 `agentx` 包现在真实提供什么；
2. `Run`、context、并发、错误和关闭行为如何工作；
3. Open Tool Loop与 Workflow两类标准执行路径应该如何选择；
4. `ExecutionAdapter`、Model/Tool Adapter和 Host分别拥有什么责任；
5. Workflow Host Kit如何组合已经进入 Runtime module的 portable mechanism；
6. shared Host HTTP transport与具体 Scene HTTP API之间如何分工；
7. A股推荐 extension入口如何组合 Manifest、Pack、tool schema和 Host handler；
8. 哪些能力仍是后续工作，不能从 package 名或示例中误判为完整 SDK。

建议阅读顺序：

1. [快速开始](quickstart.md)
2. [安装与多 Module 引用](guides/installation-and-modules.md)
3. [公共执行模型](concepts/execution-model.md)
4. [Go API Reference](reference/agentx.md)
5. [生命周期与错误处理](guides/lifecycle-and-errors.md)
6. [自定义 Adapter](guides/custom-adapter.md)
7. [Host Kit + Model/Tool Adapter](guides/model-tool-hostkit.md)
8. [Workflow Host Kit](guides/workflow-hostkit.md)
9. [A 股 Portable Domain Extension](guides/astock-extension.md)
10. [Package API 索引与成熟度矩阵](reference/package-maturity.md)
11. [成熟度与兼容边界](maturity.md)
12. [HS 迁移说明](guides/hs-migration.md)
13. [`components/llm` 中文 API Reference](../components/llm/API.md)
14. [`runtime` 中文 package 导航](../runtime/README.md)
15. [`extensions` 中文 package 导航](../extensions/README.md)
16. [`runtime/assetfs` 中文 API Reference](../runtime/assetfs/API.md)
17. [`extensions/astock` 推荐入口中文 API Reference](../extensions/astock/API.md)
18. [`extensions/astock/contracts` 中文 API Reference](../extensions/astock/contracts/API.md)
19. [`extensions/astock/hostkit` 中文 API Reference](../extensions/astock/hostkit/API.md)
20. [`extensions/domainmodule` 中文 API Reference](../extensions/domainmodule/API.md)
21. [`extensions/pack` 中文 API Reference](../extensions/pack/API.md)
22. [`runtime/construction` 中文 API Reference](../runtime/construction/API.md)
23. [`runtime/execution` 中文 API Reference](../runtime/execution/API.md)
24. [`runtime/executionpolicy` 中文 API Reference](../runtime/executionpolicy/API.md)
25. [`runtime/hostkit` 中文 API Reference](../runtime/hostkit/API.md)
26. [`runtime/hosthttp/hostserver` 中文 API Reference](../runtime/hosthttp/hostserver/API.md)
27. [`runtime/hosthttp/requestjson` 中文 API Reference](../runtime/hosthttp/requestjson/API.md)
28. [`runtime/hosthttp/resourcepolicy` 中文 API Reference](../runtime/hosthttp/resourcepolicy/API.md)
29. [`runtime/toolloop` 中文 API Reference](../runtime/toolloop/API.md)
30. [`runtime/workflow/composition` 中文 API Reference](../runtime/workflow/composition/API.md)
31. [`runtime/workflow/hostkit` 中文 API Reference](../runtime/workflow/hostkit/API.md)

可运行验证位于：

- [`examples/contract-basic`](../examples/contract-basic)
- [`examples/custom-adapter`](../examples/custom-adapter)
- [`conformance/consumer`](../conformance/consumer)
- [`runtime/conformance/protocol-consumer`](../runtime/conformance/protocol-consumer)
- [`runtime/conformance/construction-consumer`](../runtime/conformance/construction-consumer)
- [`runtime/conformance/execution-consumer`](../runtime/conformance/execution-consumer)
- [`runtime/conformance/hostkit-consumer`](../runtime/conformance/hostkit-consumer)
- [`runtime/conformance/toolloop-consumer`](../runtime/conformance/toolloop-consumer)
- [`runtime/conformance/workflow-hostkit-consumer`](../runtime/conformance/workflow-hostkit-consumer)
- [`extensions/conformance/astock-contract-consumer`](../extensions/conformance/astock-contract-consumer)
- [`extensions/conformance/domain-module-consumer`](../extensions/conformance/domain-module-consumer)
- [`extensions/conformance/pack-consumer`](../extensions/conformance/pack-consumer)
- [`extensions/conformance/astock-consumer`](../extensions/conformance/astock-consumer)

`docs/**` 的主体页面描述根 contract module；`components/llm/API.md` 描述 LLM
合同；`runtime/**/API.md` 描述已经真实落地且各自标注成熟度的 Runtime owner；
`extensions/**/API.md`描述获准迁入的 portable extension合同。
M5D当前只批准 A股 portable Manifest/assets/tool schema/Pack/hostkit切片；尚未落地
的 hostless完整 Runtime construction、A股 livekit/provider、更多 components、其它
extensions和 Scene不会预先获得虚假 API 页面。
