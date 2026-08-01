# AgentX Go 中文文档

当前文档对应 M3E Core Developer Preview candidate、M4A Experimental Host
HTTP owner、LLM组件和已迁 Runtime owner，目标是让调用方
能够准确判断：

1. 根 `agentx` 包现在真实提供什么；
2. `Run`、context、并发、错误和关闭行为如何工作；
3. Open Tool Loop与 Workflow两类标准执行路径应该如何选择；
4. `ExecutionAdapter`、Model/Tool Adapter和 Host分别拥有什么责任；
5. Workflow Host Kit如何组合已经进入 Runtime module的 portable mechanism；
6. shared Host HTTP transport与具体 Scene HTTP API之间如何分工；
7. 哪些能力仍是后续工作，不能从 package 名或示例中误判为完整 SDK。

建议阅读顺序：

1. [快速开始](quickstart.md)
2. [安装与多 Module 引用](guides/installation-and-modules.md)
3. [公共执行模型](concepts/execution-model.md)
4. [Go API Reference](reference/agentx.md)
5. [生命周期与错误处理](guides/lifecycle-and-errors.md)
6. [自定义 Adapter](guides/custom-adapter.md)
7. [Host Kit + Model/Tool Adapter](guides/model-tool-hostkit.md)
8. [Workflow Host Kit](guides/workflow-hostkit.md)
9. [Package API 索引与成熟度矩阵](reference/package-maturity.md)
10. [成熟度与兼容边界](maturity.md)
11. [HS 迁移说明](guides/hs-migration.md)
12. [`components/llm` 中文 API Reference](../components/llm/API.md)
13. [`runtime` 中文 package 导航](../runtime/README.md)
14. [`runtime/construction` 中文 API Reference](../runtime/construction/API.md)
15. [`runtime/execution` 中文 API Reference](../runtime/execution/API.md)
16. [`runtime/hostkit` 中文 API Reference](../runtime/hostkit/API.md)
17. [`runtime/hosthttp/hostserver` 中文 API Reference](../runtime/hosthttp/hostserver/API.md)
18. [`runtime/hosthttp/requestjson` 中文 API Reference](../runtime/hosthttp/requestjson/API.md)
19. [`runtime/hosthttp/resourcepolicy` 中文 API Reference](../runtime/hosthttp/resourcepolicy/API.md)
20. [`runtime/toolloop` 中文 API Reference](../runtime/toolloop/API.md)
21. [`runtime/workflow/composition` 中文 API Reference](../runtime/workflow/composition/API.md)
22. [`runtime/workflow/hostkit` 中文 API Reference](../runtime/workflow/hostkit/API.md)

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

`docs/**` 的主体页面描述根 contract module；`components/llm/API.md` 描述 LLM
合同；`runtime/**/API.md` 描述已经真实落地且各自标注成熟度的 Runtime owner。
尚未落地的 hostless完整 Runtime construction、更多 components、其它 extensions和
Scene不会预先获得虚假 API 页面。
