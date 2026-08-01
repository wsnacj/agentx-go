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

## M5A AssetFS 与 A 股 contracts 收口

immutable asset机制的 source authority现位于 `runtime/assetfs`。它拥有输入
`fs.FS`快照、content fingerprint、provider identity、startup-only resolver、
seal和 `assetfs://`解析；HS `core/agentx/assetfs`只保留 Deprecated alias与
forwarder。Core和现有 Scene consumer均直接固定 canonical runtime版本，不得在
HS兼容路径恢复 snapshot或 resolver实现。

A股 portable DTO、JSON行为、证券代码 normalization、readiness和 assessment
机制的 source authority现位于 `extensions/astock/contracts`。HS
`scene/agentx_a_stock/contracts`只保留 Deprecated alias与四个函数转发；A股
production consumer直接使用 canonical contracts。

以下能力仍由 HS A股 Scene拥有，不属于 extensions：

- provider、HTTP client、credential、cookie、proxy和缓存；
- livekit、pack/workflow、工具执行和模型调用；
- 来源优先级、fallback、freshness核验和最终回答策略；
- 任何真实网络、交易或其它生产副作用。

`extensions/conformance/astock-contract-consumer`固定 runtime/extensions
pseudo-version，以无 HS、无 `replace`方式组合 `assetfs`与 A股 JSON合同。该
consumer只证明 portable合同可独立消费，不表示完整 A股 Scene已经迁仓或可发布。

## M5B Domain Module注册编排收口

portable manifest、module-scoped config、diagnostics、config resolver、重复ID检查、
顺序注册和 report聚合的 source authority现位于
`extensions/domainmodule`。HS `core/agentx/domainmodule`对这些类型只保留
Deprecated alias/forwarder，并把既有 `RegisterAll`委托给 canonical coordinator。

HS adapter继续拥有 `Target`、Runner sealed检查、extension filesystem/runtime、
pack registry、tool executor与 skill投影。九个现有 Scene module已直接使用canonical
Manifest/ConfigRequirement/Diagnostics，同时继续通过 HS Target helper接入具体宿主。
迁移没有改变注册顺序、错误文本、diagnostic JSON或“失败前的 mutation不回滚”语义。

`extensions/conformance/domain-module-consumer`固定 extensions pseudo-version，证明
外部项目无需 HS、Runner或 `replace`即可使用portable coordinator；它不包含具体
Scene、pack、tool executor、provider或credential。

## 尚未迁移

W2-B结论统一为 `not_ready_for_hostless_w2b`。当前没有无需调用方提供 model/tool
adapter、policy或 backend的完整 embedded Runtime；Workflow已有独立 Developer
Preview Host Kit，但没有成为根 `Client` mode。完整 Scene、Objective、Resume、
concrete durable backend、正式发行和 Public/Beta/Stable均不在当前范围。
