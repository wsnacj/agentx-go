# 七类能力与标准接入路径

本页回答“AgentX现在支持什么、应该从哪个入口接入”。七项概念不是七个互斥的
`mode`枚举：A0是控制面状态，Tool Direct Answer是结果策略，长任务由多个runtime
owner组合，Deterministic Scene属于扩展生态。根`agentx.Client`也不会为了统一名称而
伪装并不存在的Workflow或Objective mode。

## 能力矩阵

| 能力 | 当前实现 | 推荐入口 | 可运行证据 | Host仍拥有 |
| --- | --- | --- | --- | --- |
| A0 控制面关闭 | 根执行画像的默认状态 | `agentx.New`或`hostkit.NewChatClient` | [`examples/chat`](../../examples/chat) | 是否启用Objective及产品policy |
| Open Tool Loop | portable round、driver、phase、终止与assembly真实实现 | [`runtime/hostkit`](../../runtime/hostkit/API.md)；已有Runtime时可实现根`ExecutionAdapter` | [`examples/tool-loop`](../../examples/tool-loop) | model/tool实现、授权、sandbox、provider和backend |
| Tool Direct Answer | Host Kit中的显式结果策略 | [`ToolDirectAnswer`](tool-direct-answer.md) | [`examples/tool-loop`](../../examples/tool-loop) | 哪类工具结果允许直接展示的安全决策 |
| Workflow | Spec、validation、lowering、journal、node execution、orchestration与composition真实实现 | [`runtime/workflow/hostkit`](../../runtime/workflow/hostkit/API.md) | [`examples/workflow`](../../examples/workflow) | validation policy、concrete executor和durable backend |
| Objective Runtime Loop | Objective kernel与Host-owned dispatch/verification组合 | [`runtime/objective/hostkit`](../../runtime/objective/hostkit/API.md) | [`examples/objective`](../../examples/objective) | catalog/policy、真实执行、授权和业务完成标准 |
| 长任务编排 | Task/Session/Subagent、portable scheduler、resume/readback和bounded lifecycle | [`runtime/session/hostkit`](../../runtime/session/hostkit/API.md) | [`examples/session-subagent`](../../examples/session-subagent) | process/system scheduler、durable store、worker部署与副作用 |
| Deterministic Scene | Domain Kit、Pack及多个provider-neutral Scene owner | [`extensions/domainkit`](../../extensions/domainkit/API.md)与[Scene成熟度索引](../reference/package-maturity.md) | [`examples/deterministic-scene`](../../examples/deterministic-scene) | provider、credential、真实网络、审批和产品规则 |

模型对话是最小的模型执行路径，使用
[`hostkit.NewChatClient`](model-tool-hostkit.md#纯模型对话)；它不需要工具循环，也不等于
Objective或Workflow。通用tool、provider、Browser和Document能力分别位于独立module，
调用方只引入实际使用的依赖。

## 代码成熟度边界

- 根合同与选定Host Kit是Developer Preview candidate；Scene及其余实现按各自
  `API.md`标记为Experimental extension或internalization candidate。
- “真实实现”表示canonical仓库拥有portable algorithm/state/orchestration，并已有
  fixed-version consumer；不表示提供默认provider、credential或生产backend。
- 当前是Developer Preview，不构成Beta、Stable、SLA或production-ready承诺。
- `not_ready_for_hostless_w2b`仍成立：新项目必须显式注入model/tool、policy和需要的
  backend。AgentX不会静默读取环境凭据或替Host决定生产副作用。

## 选择顺序

1. 只做单轮模型请求：使用[模型对话](chat.md)。
2. 需要模型动态调用工具：使用[Host Kit + Model/Tool Adapter](model-tool-hostkit.md)。
3. 工具结果可直接收口：在上一条路径中显式返回[Tool Direct Answer](tool-direct-answer.md)。
4. 执行图在运行前已声明：使用[Workflow Host Kit](workflow-hostkit.md)。
5. 需要目标驱动的Host执行与验证：使用[Objective Host Kit](objective-hostkit.md)。
6. 需要子任务、调度和恢复：使用[Task / Session / Subagent Host Kit](session-subagent-hostkit.md)。
7. 需要无模型或领域确定性验收：组合Domain Kit与目标Scene，并自行注入Host handler。

如果调用方已经有完整Runtime，只需要对齐根Run/错误/取消/Shutdown合同，可使用
[自定义ExecutionAdapter](custom-adapter.md)。它是高级construction seam，不是普通新项目
使用Open Tool Loop的默认样板。

希望先看完整组合方式的新项目可从[Reference Host](reference-host.md)开始。该样板默认
离线且显式选择provider、tools和RunStore；它不是默认provider或生产Host。
