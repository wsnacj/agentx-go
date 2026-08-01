# HS 到 agentx-go 的迁移边界

本页说明当前 source authority 在哪里，帮助 HS维护者避免恢复双写。它不是要求
业务 consumer 一次性迁仓，也不授权 Scene迁移。

## 当前依赖方向

```text
agentx-go portable implementation
        ↓ contracts / ports
HS host adapter、policy、backend
        ↓
provider、credential、Scene和真实副作用
```

agentx-go production code不得 import `hs/...`、Runner或 `scene/...`。HS可以固定
pseudo-version消费 canonical owner，但 owner package不得反向依赖 Facade或 HS。

## W5-H round owner 收口

`runtime/hostkit.ModelToolRoundAdapter`现在拥有：

- 模型请求、响应观察、工具前 gate和工具执行的固定阶段顺序；
- canonical LLM response与 portable toolloop之间的防御性投影；
- model-only完成、host停止和tool continuation的 round outcome；
- port error identity与失败阶段保留。

HS `core/agentx/engine` production path已切换到该 implementation。HS仍拥有：

- concrete model/provider请求、recovery和配置映射；
- response persistence、telemetry和answer-contract；
- authorization、approval、sandbox和真实工具执行；
- RunStore、scheduler、budget owner和产品默认值。

因此 HS不再直接构造通用 `RoundPhaseCoordinator`，也不再保留一份同义的通用
phase executor；它只保留上述具体 host闭包和结果投影。后续修改 portable阶段顺序
应先进入 agentx-go并通过 compatibility differential，不得在 HS恢复分叉。

## Open Tool Loop 两类 consumer

- HS production consumer：验证现有 Open Tool Loop业务行为不变；
- `runtime/conformance/hostkit-consumer`：用固定版本验证无 HS、无 Runner、无长期
  `replace` 的新项目接入。

后者使用 `NewModelToolClient`，不再手写 Factory、BuildRun assembly或
RoundExecutor。它仍显式提供 Model/Tool Adapter，避免把 provider、credential和
产品策略下沉到通用 Runtime。

## M3E Workflow Host Kit 收口

HS `core/agentx/workflow.ExecuteInline`现在固定使用
`runtime/workflow/hostkit.New`。HS只保留：

- validation和 tool/model/task mapping policy；
- concrete executor capability adapter；
- RunStore到 portable `JournalPort`的 backend adapter；
- identity、clock、error display与结果兼容投影。

lowering、journal construction、node executor能力优先级、orchestration和
composition继续由 agentx-go owner持有。fixed-version
`runtime/conformance/workflow-hostkit-consumer`证明新项目无需 HS、Runner或长期
`replace`即可运行显式 Workflow。

## 尚未迁移

W2-B结论统一为 `not_ready_for_hostless_w2b`。当前没有无需调用方提供 model/tool
adapter、policy或 backend的完整 embedded Runtime；Workflow已有独立 Developer
Preview Host Kit，但没有成为根 `Client` mode。Scene、Objective、Resume、concrete
durable backend、正式发行和 Public/Beta/Stable均不在 M3E范围。
