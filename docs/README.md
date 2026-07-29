# AgentX Go 中文文档

当前文档对应 W1 private validation 合同，目标是让调用方能够准确判断：

1. 根 `agentx` 包现在真实提供什么；
2. `Run`、context、并发、错误和关闭行为如何工作；
3. `ExecutionAdapter` 由谁实现、拥有什么责任；
4. 哪些能力仍是后续工作，不能从空类型或示例中误判为已支持。

建议阅读顺序：

1. [快速开始](quickstart.md)
2. [公共执行模型](concepts/execution-model.md)
3. [Go API Reference](reference/agentx.md)
4. [生命周期与错误处理](guides/lifecycle-and-errors.md)
5. [自定义 Adapter](guides/custom-adapter.md)
6. [成熟度与兼容边界](maturity.md)

这些页面只描述根 contract module。后续 Runtime、components、extensions 和 Scene
只有在真实代码、consumer 与门禁落地后，才会获得各自的中文 API 文档入口。
