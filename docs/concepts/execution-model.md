# 公共执行模型

## 首版代码合同

W1 将一次公共执行压缩为：

```text
RunRequest
  -> Client
  -> ExecutionAdapter（由 runtime/hostkit 组合 runtime/execution 实现）
  -> typed execution.Host
  -> portable toolloop Assembly
  -> host RoundExecutor / policy ports
  -> AdapterRunResult
  -> RunResult
```

`RunRequest` 只包含输入与可选 Session identity。`RunResult` 只公开调用方稳定需要
的 identity、状态、回复、证据、blocker、next action 和执行画像。底层 Runner
状态、provider 响应、credential、Scene registry 和内部 diagnostics 不进入根合同。

## 六维执行画像

AgentX 使用六个正交维度描述执行，而不是把所有能力折叠为一个不断增长的 mode
枚举：

| 维度 | W1 值 | 含义 |
| --- | --- | --- |
| Activation | `off` | Objective 控制面未开启 |
| ControlMode | `tool` | 当前控制方式是工具执行 |
| ExecutionIntensity | `l2_bounded_tool_loop` | 有界工具循环 |
| Driver | `open_tool_loop` | Runner 动态执行路径 |
| ResultPolicy | `runner_final_reply` | 由 Runner 最终回复收口 |
| Lifecycle | `synchronous_run` | 单次同步调用生命周期 |

W1 只接受这一个完整画像。字段使用字符串是为了保留 M2 已验证合同，并不代表任意
组合都获得支持。

## 与既有七种“模式”的关系

| 既有概念 | W1 状态 |
| --- | --- |
| A0 控制面关闭 | 进入画像：`Activation=off` |
| Open Tool Loop | 进入画像：`Driver=open_tool_loop` |
| Tool Direct Answer | 暂不进入首版 ResultPolicy |
| Workflow | non-goal |
| Objective Runtime Loop | non-goal |
| 长任务编排 | non-goal |
| Deterministic Scene | 生态验收路径，不属于根 Client 模式 |

因此，W1 不是把七项删成一种模式，而是只发布已经能够稳定解释和验证的最小垂直
切片。Workflow、Objective、Resume 和 durable lifecycle 不提供空实现，也不通过
文档暗示已支持。

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
