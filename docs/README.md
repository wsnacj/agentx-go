# AgentX Go 中文文档

当前文档对应根合同、LLM 组件和已迁 Experimental Runtime owner，目标是让调用方
能够准确判断：

1. 根 `agentx` 包现在真实提供什么；
2. `Run`、context、并发、错误和关闭行为如何工作；
3. `ExecutionAdapter` 由谁实现、拥有什么责任；
4. 哪些 Workflow portable mechanism已经进入 Runtime module；
5. 哪些能力仍是后续工作，不能从 package 名或示例中误判为完整 SDK。

建议阅读顺序：

1. [快速开始](quickstart.md)
2. [公共执行模型](concepts/execution-model.md)
3. [Go API Reference](reference/agentx.md)
4. [生命周期与错误处理](guides/lifecycle-and-errors.md)
5. [自定义 Adapter](guides/custom-adapter.md)
6. [成熟度与兼容边界](maturity.md)
7. [`components/llm` 中文 API Reference](../components/llm/API.md)
8. [`runtime` 中文 package 导航](../runtime/README.md)
9. [`runtime/construction` 中文 API Reference](../runtime/construction/API.md)
10. [`runtime/execution` 中文 API Reference](../runtime/execution/API.md)
11. [`runtime/toolloop` 中文 API Reference](../runtime/toolloop/API.md)
12. [`runtime/workflow/composition` 中文 API Reference](../runtime/workflow/composition/API.md)

可运行验证位于：

- [`examples/contract-basic`](../examples/contract-basic)
- [`examples/custom-adapter`](../examples/custom-adapter)
- [`conformance/consumer`](../conformance/consumer)
- [`runtime/conformance/protocol-consumer`](../runtime/conformance/protocol-consumer)
- [`runtime/conformance/construction-consumer`](../runtime/conformance/construction-consumer)
- [`runtime/conformance/execution-consumer`](../runtime/conformance/execution-consumer)
- [`runtime/conformance/toolloop-consumer`](../runtime/conformance/toolloop-consumer)

`docs/**` 的主体页面描述根 contract module；`components/llm/API.md` 描述 LLM
合同；`runtime/**/API.md` 描述已经真实落地的 Experimental Runtime owner。
尚未落地的无需 Host 的根 Runtime construction、更多 components、extensions和
Scene不会预先获得虚假 API 页面。
