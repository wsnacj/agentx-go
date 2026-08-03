# 示例与可运行消费证据

`examples`只保留适合复制阅读的最小入口；更完整的跨module验证位于各
`conformance/*-consumer`独立module。两者职责不同：example解释怎么接入，consumer证明
固定pseudo-version在无HS、无Runner、无长期`replace`条件下可以构建和运行。

## 最小示例

| 示例 | 说明 |
| --- | --- |
| [`contract-basic`](contract-basic) | 根`Client`、`Run`、identity和有界`Shutdown` |
| [`custom-adapter`](custom-adapter) | 自定义`ExecutionAdapter`、typed error与`errors.Is/As` |

运行：

```bash
GOWORK=off go run ./contract-basic
GOWORK=off go run ./custom-adapter
```

## 七类能力的验证入口

| 能力 | 推荐文档 | fixed/external-style consumer |
| --- | --- | --- |
| A0 / 根合同 | [快速开始](../docs/quickstart.md) | [`conformance/consumer`](../conformance/consumer) |
| Open Tool Loop | [Model/Tool Host Kit](../docs/guides/model-tool-hostkit.md) | [`runtime/conformance/hostkit-consumer`](../runtime/conformance/hostkit-consumer) |
| Tool Direct Answer | [Tool Direct Answer](../docs/guides/tool-direct-answer.md) | [`conformance/consumer`](../conformance/consumer) |
| Workflow | [Workflow Host Kit](../docs/guides/workflow-hostkit.md) | [`runtime/conformance/workflow-hostkit-consumer`](../runtime/conformance/workflow-hostkit-consumer) |
| Objective | [Objective Host Kit](../docs/guides/objective-hostkit.md) | [`runtime/conformance/objective-hostkit-consumer`](../runtime/conformance/objective-hostkit-consumer) |
| 长任务 / 子任务 | [Session/Subagent Host Kit](../docs/guides/session-subagent-hostkit.md) | [`runtime/conformance/session-hostkit-consumer`](../runtime/conformance/session-hostkit-consumer) |
| Deterministic Scene | [能力矩阵](../docs/guides/capability-map.md) | [`scenes/conformance/astock-consumer`](../scenes/conformance/astock-consumer) |

Browser、Document、Provider、Tools和其它Scene也各有独立conformance consumer；完整入口
见[中文文档首页](../docs/README.md)。所有consumer默认使用fixture、fake或in-memory port，
不读取credential、不访问真实网络，也不代表正式发布证据。
