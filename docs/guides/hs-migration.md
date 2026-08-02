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

## M5C Portable Pack Core收口

Pack Definition、显式 validation协调、并发 registry、route/binding与 Workflow
materialization的 source authority位于 `extensions/pack`。HS `core/agentx/pack`
保留旧 Definition方法集及 Host validator/tool-lowering adapter；`pack/runtime`的
memory/eval backend仍由 HS拥有。

## M5D A股 Portable Extension收口

`extensions/astock`现在拥有 A股 Manifest、compiled skill/tool资产、运行时 tool
schema、三组 Pack Definition/evaluator和聚合注册；`extensions/astock/hostkit`拥有
intent转换、显式 handler协调、readiness聚合和确定性回答格式化。

HS `scene/agentx_a_stock/module`已经直接消费 canonical Manifest、资产与 Pack注册；
原 `packs`、portable `tools` schema和 `hostkit`实现已删除或降为薄转发。HS继续拥有：

- livekit、provider、HTTP/source adapter、credential、cache与 fallback；
- concrete tool registry/executor、Host Target mutation和 Runner wiring；
- filesystem plugin bundle、诊断 CLI和真实网络。

root filesystem skill/tool文件暂保留为显式路径 API的 compatibility mirror，并由
differential测试约束；它们不再是 compiled asset source authority。plugin bundle的
command/risk metadata是独立分发合同，不得用 canonical lookup manifest覆盖。

`extensions/conformance/astock-consumer`以固定版本证明无 HS、Runner、`replace`和
网络的 Manifest→asset→Pack→binding→fixture Host Kit→evaluator路径。

## M5E Portable Skills Core收口

`extensions/skills`现在拥有 Skill types/normalization、`SKILL.md` loader、source
precedence、cache/generation/watcher、activation path、requested semantics、deep
clone和 resource refs的 source authority。HS `core/agentx/skills`对这些能力只保留
canonical type alias与薄 forwarder；新的通用 loader/cache/activation修改必须先落在
`extensions/skills`，不得在 HS恢复双写。

HS继续拥有 prompt catalog/ranking及 memory/browser启发式、eligibility/filter、
safety规则、install plan/apply、approval/rollback/命令执行、bundled Skill内容和
Runner/ProductShell集成。这些是产品策略或真实副作用，不能为了减少旧 import而下沉
到 canonical extension。

`extensions/conformance/skills-consumer`固定 extensions/runtime pseudo-version，
证明新项目可在无 HS、Runner、`replace`、网络和命令执行时完成 immutable source
加载、缓存、activation、requested semantics、资源完整性与 deep clone。

## P1-B Objective Host Kit收口

`runtime/objective/hostkit`现在拥有managed ingress之后的显式Host dispatch、handler
选择、exact-once调用、observation normalization和Objective verification组合。
`runtime/objective`提供普通Host所需的最小类型/构造名称；大型`controlcontract`继续作为
Experimental kernel，不作为推荐业务入口。

HS `productshellruntime.ManagedObjectiveRuntimeAdapterProductEntrypoint`继续拥有产品refs、
配置来源、operator approval和具体handlers，但其production dispatch已通过固定版本使用
canonical implementation。HS原315行dispatch和通用observation normalization已删除为薄
alias/forwarder；owner gate禁止这些机制回流。

`runtime/conformance/objective-hostkit-consumer`证明新项目无需HS、Runner、Scene、provider
或长期`replace`即可完成一次Objective Host adapter闭环。Task/Session/Subagent、
Scheduler/Resume、durable backend和真实副作用仍由后续能力wave或Host拥有。

## 尚未迁移

M5S没有开启新的HS source-authority迁移。它把已经完成cutover的根Client、
Model/Tool Host Kit和Workflow Host Kit固定为三条标准接入路径，并用一个新的
fixed-version external consumer统一验收。HS继续固定各owner已验收的不可变
pseudo-version；本轮不为了版本外观一致而触碰Scene或重写业务consumer。

A股evaluator DTO的canonical定义已收口到`extensions/astock/contracts`，
`extensions/astock`和三组内部Pack只保留同一类型身份的alias。该调整用于消除候选
公开类型对Go`internal`路径的泄漏，不改变HS evaluator调用、JSON或产品策略。

W2-B结论统一为 `not_ready_for_hostless_w2b`。当前没有无需调用方提供 model/tool
adapter、policy或 backend的完整 embedded Runtime；Workflow和Objective已有独立
Developer Preview Host Kit，但没有成为根 `Client` mode。完整 Scene、Task/Session、
Resume、concrete durable backend、正式发行和 Public/Beta/Stable仍不在当前范围。
