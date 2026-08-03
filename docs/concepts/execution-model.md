# 公共执行模型

## M3E 代码合同

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

Workflow保持独立显式图路径：

```text
workflow.Spec
  -> workflow/hostkit
  -> validation + lowering
  -> journal + nodeexec + orchestration
  -> composition.Result
```

它与根 Client共享 context/error identity等原则，但不虚构根 Client的 Workflow
mode或 Shutdown；Host继续拥有 concrete policy、executor和 backend生命周期。

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

| 既有概念 | 当前状态 |
| --- | --- |
| A0 控制面关闭 | 进入画像：`Activation=off` |
| Open Tool Loop | 进入画像：`Driver=open_tool_loop`；Host Kit已有真实 portable implementation |
| Tool Direct Answer | 已进入 Host Kit 显式结果策略；不新增根 `ExecutionProfile` 组合 |
| Workflow | canonical Runtime已有真实实现和独立 Host Kit标准入口；不进入根 Client Facade |
| Objective Runtime Loop | `runtime/objective/hostkit`已有真实portable控制与Host dispatch/verification组合；不进入根Client Facade |
| 长任务编排 | `runtime/session/hostkit`已组合Task/Session/Subagent、Scheduler与Resume内核；生产scheduler/backend仍由Host拥有 |
| Deterministic Scene | `extensions/domainkit`与`scenes`已有真实实现和fixed consumer；不属于根Client模式 |

因此，这不是把七项删成一种 mode，也不把 Workflow强塞进根 Client。Model Conversation
使用`NewChatClient`，Open Tool Loop与 Tool Direct Answer通过根 Client/Host Kit接入；
Workflow通过独立 Workflow Host Kit接入。Objective和长任务通过各自Host Kit接入；它们
已提供portable机制，但不伪装为根Client mode，也不声称包含生产backend。七类能力的
入口、证据和Host责任见[能力矩阵](../guides/capability-map.md)。

## Open Tool Loop 与 Workflow 边界

`runtime/hostkit`和 `runtime/toolloop`拥有 Open Tool Loop的 portable机制，但不拥有
provider、授权、sandbox、RunStore或产品默认值。`runtime/workflow/*`已经拥有一组
真实 portable implementation owner，但 concrete executor、backend、产品 validation
policy和根 Facade construction继续由 Host拥有。

M3E补齐了“代码已经迁入”与“可被新项目标准接入”之间的缺口：Workflow调用方只
依赖 `workflow`和 `workflow/hostkit`，低层 composition/journal/nodeexec等 package
仍按 Experimental或 internalization candidate治理。

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
