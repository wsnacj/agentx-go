# 公共执行模型

## M3D 代码合同

自定义 Adapter路径：

```text
RunRequest -> Client -> ExecutionAdapter -> AdapterRunResult -> RunResult
```

Host Kit路径进一步落地为：

```text
RunRequest
  -> Client
  -> ExecutionAdapter（由 runtime/hostkit 组合 runtime/execution 实现）
  -> typed execution.Host
  -> portable toolloop Assembly
  -> ModelToolRoundAdapter
  -> caller model/tool functions / optional policy ports
  -> AdapterRunResult
  -> RunResult
```

`RunRequest` 只包含输入与可选 Session identity。`RunResult` 只公开调用方稳定需要
的 identity、状态、回复、证据、blocker、next action 和执行画像。底层 Runner
状态、provider 响应、credential、Scene registry 和内部 diagnostics 不进入根合同。

## 六维执行画像

AgentX 使用六个正交维度描述执行，而不是把所有能力折叠为一个不断增长的 mode
枚举：

| 维度 | 当前根合同值 | 含义 |
| --- | --- | --- |
| Activation | `off` | Objective 控制面未开启 |
| ControlMode | `tool` | 当前控制方式是工具执行 |
| ExecutionIntensity | `l2_bounded_tool_loop` | 有界工具循环 |
| Driver | `open_tool_loop` | Runner 动态执行路径 |
| ResultPolicy | `runner_final_reply` | 由 Runner 最终回复收口 |
| Lifecycle | `synchronous_run` | 单次同步调用生命周期 |

当前根合同只接受这一个完整画像。字段使用字符串是为了保留 M2已验证合同，并不代表任意
组合都获得支持。

## 与既有七种“模式”的关系

| 既有概念 | M3D 状态 |
| --- | --- |
| A0 控制面关闭 | 进入画像：`Activation=off` |
| Open Tool Loop | 进入画像：`Driver=open_tool_loop`；Host Kit已有真实 portable implementation |
| Tool Direct Answer | 暂不进入首版 ResultPolicy |
| Workflow | canonical Runtime extensions已有真实实现；尚未进入根 Client Facade |
| Objective Runtime Loop | non-goal |
| 长任务编排 | non-goal |
| Deterministic Scene | 生态验收路径，不属于根 Client 模式 |

因此，M3D不是把七项删成一种模式，而是为根 Client只选择已经能够稳定解释和验证的
最小垂直切片。Workflow的 Spec、validation、lowering、orchestration和composition
可作为 Experimental extension直接使用，但不伪装成根 Facade能力；Objective、
Resume和完整 durable lifecycle仍不提供空实现，也不通过文档暗示已支持。

## Open Tool Loop 与 Workflow 边界

`runtime/hostkit`和 `runtime/toolloop`拥有 Open Tool Loop的 portable机制，但不拥有
provider、授权、sandbox、RunStore或产品默认值。`runtime/workflow/*`已经拥有一组
真实 portable implementation owner，但 concrete executor、backend、产品 validation
policy和根 Facade construction继续由 Host拥有。

因此“代码已经迁入”与“成为 Developer Preview标准入口”是两项不同结论：当前标准
入口是自定义 ExecutionAdapter和 Host Kit + Model/Tool Adapter；Workflow包继续按
Experimental extension使用。

## 状态与下一步

根合同只输出四种状态：

| Status | Blocker | NextAction |
| --- | --- | --- |
| `completed` | 无 | 空 |
| `blocked` | `execution_incomplete` | `resolve_execution_blocker` |
| `canceled` | `canceled` | `caller_decides_retry` |
| `failed` | 稳定错误码 | 与错误码对应的稳定动作 |

`Evidence` 当前只包含非敏感 identity，例如 `run:<id>` 和 `session:<id>`。
底层原始错误不会进入 `Reply`、`Blockers` 或可展示错误文本。
