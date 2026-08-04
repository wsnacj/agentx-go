# HS 到 agentx-go 的迁移边界

本页说明当前 source authority 在哪里，帮助 HS维护者避免恢复双写。它不是要求
业务 consumer 一次性迁仓，也不授权 Scene迁移。

## 当前完成度不是“删除HS目录”

截至2026-08-04，HS根`go.mod`已经固定消费九个agentx-go module；`core/agentx`有427个
production Go文件、`scene/agentx_*`有120个production Go文件直接import canonical路径。
agentx-go本身有868个production Go文件、258,751行，且production代码没有HS反向依赖。

HS仍有大量AgentX代码并不等于通用实现仍在双写。它主要分为：

- concrete model/tool/backend、授权、安全、配置、存储与产品策略等Host实现；
- 为仍使用`hs/core/agentx/...`的调用方保留的Deprecated alias/forwarder；
- 大型examples、selftest、live harness和历史治理代码；
- 具体Scene的provider、credential、HTTP/CLI、业务聚合和真实副作用。

当前production范围仍有2,166处`hs/core/agentx/...` import，分布于1,029个Go文件：大多数是
`core/agentx`内部package关系，其次是Scene，另有少量其它HS consumer。因此不能把
`core/agentx`目录直接删除，或把所有import机械替换成新仓路径：canonical没有、也不应拥有
客户/部署Host语义。

如果最终目标进一步要求“HS不再提供任何`core/agentx`旧路径，只作为agentx-go的产品
Host”，还需要一个独立的兼容退役阶段：先把所有portable authority核对为canonical唯一
owner，再把HS专属实现移到明确的`internal/agentxhost`或Scene/应用owner，逐批切完旧import，
最后删除closure为零的shim。P1至P5完成的是公共portable Platform与接入文档，不包含这次
大范围目录/旧路径退役，不能把两种完成定义混为一谈。

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

## P6-B Host owner与compatibility checkpoint

P6-B1至P6-B3已把Runner registration、Tool Registry、Pack、Extension Runtime与AfterRun接线
归位到HS `internal/agentxhost/runnerwiring`，并切换core及应用production consumer。旧
`core/agentx/hostwiring`不再拥有Runner mutation实现。

P6-B4又删除4个零production closure直连helper，并将其行为测试归位真实internal owner。
P6-B0.5完整live/NL矩阵复测Core 5/5、只读Scene 3/3；完整AgentX回归没有新增功能失败。

P6-C随后完成该残留边界的Owner决定与收口：不新增canonical公共Surface。17个Scene导出
facade改为返回HS-only `domainmodule.HostRegistration`构造值，应用由
`domainmodule.Target.RegisterHost`统一应用；真正Runner mutation继续只有
`internal/agentxhost/runnerwiring`一个owner。4个Core残留中的Host diagnostics归位
`internal/agentxhost/diagnostics`，guardrail归位`runnerwiring`；
`core/agentx/hostwiring`与`internal/agentxhost/scenecompat`均已删除。

该变化没有扩大agentx-go API，也没有要求新项目实现HS compatibility。当前严格旧路径仍有
2,166处/1,029文件，主要是HS内部Host/package关系；它们只能在后续owner-level package
internalization与consumer closure中逐批处理，不能机械替换或复制到canonical。

这不会增加新项目的接入成本：仓库外项目不需要、也不应实现HS compatibility层，应直接使用
`agentx-go`的根Client、`runtime/hostkit`、Workflow/Objective/Session Host Kit和各Domain Kit。
剩余compatibility只服务HS现有Scene与历史调用路径。

## P2-A OpenAI-compatible provider收口

`providers/openaicompat`现在拥有请求序列化、chat/vision/embedding/bot响应解码、
SSE事件流和HTTP状态错误的真实implementation；`providers/transport`、`fault`、
`retry`与`usage`拥有相应provider-neutral机制。

HS `core/llmx/provider/http`保留原`HTTPProvider`签名和metadata，只把HS Config映射为
canonical Config，并注入Host拥有的auth resolver与local-media resolver。旧HTTP
protocol实现已删除；fault/retry/transport/usage旧包降为alias或薄forwarder。HS继续拥有：

- API key来源、环境/credential store、rotation与认证header策略；
- 默认模型、provider注册、endpoint/proxy/配额与生产出网授权；
- tracing、审计、FileCollector和其它具体provider。

固定consumer位于`providers/conformance/openaicompat-consumer`，不依赖HS、Runner、Scene、
长期`replace`或真实网络。新的通用OpenAI-compatible协议行为必须先进入canonical module，
不得在HS adapter恢复双写。

## P2-B Anthropic/Codex provider收口

`providers/anthropic`现在拥有Messages payload、tool use/result、响应与usage解码；
`providers/codex`拥有Responses payload、SSE terminal/fallback、function call和usage解码。
两者都复用P2-A transport/error seam并通过显式`Authorizer`获取Host解析后的认证信息。

HS `core/llmx/provider/anthropic`只保留原`Provider`接口和Config/auth映射；
`core/llmx/provider/codex`只保留同一兼容adapter以及具体token store/OAuth refresh。旧Codex
payload、response和stream实现已删除。token文件、环境变量、用户目录、refresh写入、模型
发现与account identity投影继续是Host责任，不得迁入canonical protocol package。

固定consumer位于`providers/conformance/provider-cohort-consumer`，无HS、Runner、Scene、
长期`replace`或真实网络。新Anthropic/Codex协议行为必须先进入canonical module；HS只能
增加credential、routing、audit或其它明确Host-owned policy。

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
机制的 source authority现位于 `scenes/astock/contracts`。HS
`scene/agentx_a_stock/contracts`只保留 Deprecated alias与四个函数转发；A股
production consumer直接使用 canonical contracts。

以下能力仍由 HS A股 Scene拥有，不属于portable `scenes/astock`：

- provider、HTTP client、credential、cookie、proxy和缓存；
- livekit、pack/workflow、工具执行和模型调用；
- 来源优先级、fallback、freshness核验和最终回答策略；
- 任何真实网络、交易或其它生产副作用。

`scenes/conformance/astock-consumer`固定 scenes/runtime/extensions
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
外部项目无需 HS、Runner或 `replace`即可使用portable coordinator与无模型Domain Kit
fixture执行；它不包含具体Scene、provider、credential或真实副作用。

## M5C Portable Pack Core收口

Pack Definition、显式 validation协调、并发 registry、route/binding与 Workflow
materialization的 source authority位于 `extensions/pack`。HS `core/agentx/pack`
保留旧 Definition方法集及 Host validator/tool-lowering adapter；`pack/runtime`的
memory/eval backend仍由 HS拥有。

## M5D A股 Portable Extension收口

`scenes/astock`现在拥有 A股 Manifest、compiled skill/tool资产、运行时 tool
schema、三组 Pack Definition/evaluator和聚合注册；`scenes/astock/hostkit`拥有
intent转换、显式 handler协调、readiness聚合和确定性回答格式化。

HS `scene/agentx_a_stock/module`已经直接消费 canonical Manifest、资产与 Pack注册；
原 `packs`、portable `tools` schema和 `hostkit`实现已删除或降为薄转发。HS继续拥有：

- livekit、provider、HTTP/source adapter、credential、cache与 fallback；
- concrete tool registry/executor、Host Target mutation和 Runner wiring；
- filesystem plugin bundle、诊断 CLI和真实网络。

root filesystem skill/tool文件暂保留为显式路径 API的 compatibility mirror，并由
differential测试约束；它们不再是 compiled asset source authority。plugin bundle的
command/risk metadata是独立分发合同，不得用 canonical lookup manifest覆盖。

`scenes/conformance/astock-consumer`以固定版本证明无 HS、Runner、`replace`和
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

## P1-C Task / Session / Subagent Host Kit收口

`runtime/session/hostkit`现在拥有child worker lifecycle的portable协调机制：
exact-once invoke、durable record、按ref回读、一致性校验、parent verification和
可选Objective handoff。`runtime/session`提供普通Host所需的窄类型与构造名称，
避免推荐路径直接泄漏大型`controlcontract`。

HS `delegationruntime`、ProductShell sample和interactive CLI trace已使用固定
pseudo-version。原backend、Objective closure/handoff和async completion通用实现已删除
或降为薄alias/forwarder；HS仅保留concrete command worker、产品投影、配置、
scheduler/queue、credential和backend。

`runtime/conformance/session-hostkit-consumer`固定runtime pseudo-version，证明新项目
无需HS、Runner、Scene、provider或长期`replace`即可完成child worker
invoke→record→readback→verification闭环。Scheduler/Resume、progress/event stream、
concrete durable backend和真实副作用仍由P1-D或Host拥有。

## P2-C Tool Catalog / Diffs收口

`components/tool`与独立`tools` module现在拥有provider-neutral tool合同、线程安全catalog、
保守名称修复和纯文本diffs的source authority。HS `core/llmx/tools`只保留同类型alias和
薄forwarder；`core/agentx/tools`的diffs builtin只负责既有兼容参数归一化，再调用canonical
`diffs.Run`。owner gate禁止这些机制回流。

HS继续拥有authorization、approval、sandbox、credential，以及filesystem、process、
network、memory、retrieval和task的具体backend/policy。后续迁移不能把这些Host责任伪装成
通用tool实现，也不能让canonical module反向导入HS、Runner或Scene。

`tools/conformance/catalog-diffs-consumer`固定components/runtime/tools pseudo-version，证明
新项目可以在无HS、Runner、Scene、长期`replace`和真实副作用时完成catalog注册、名称修复
及diffs执行。该证据不代表其余通用tool已经迁移完成。

## P2-D 通用Tool Invocation与五类工具收口

canonical `tools`现在进一步拥有chain、schema sanitizer、result middleware/controlled
transform，以及message、filesystem、HTTP、memory和scheduler五类portable coordination。
这些package提供真实参数解析、稳定definition/result、取消传播和确定性路由；所有副作用
均通过`TextSender`、`Workspace`、HTTP `Preparer`/`HTTPDoer`、memory `Backend`与scheduler
`Backend`显式注入。

HS production tools已经固定引用canonical pseudo-version。原有通用协调、schema、result
middleware和五类handler被删除或降为薄Host adapter；HS继续拥有channel credential、
workspace root/symlink/atomic transaction、SSRF/proxy/network policy、memory store/visibility/
ranking、scheduler/RunStore/queue、authorization、approval与产品默认。owner gate禁止已迁
机制回流，canonical也不得反向导入HS、Runner或Scene。

`tools/conformance/general-tools-consumer`在无HS、Runner、Scene、长期`replace`和真实副作用
的独立module中注册、执行全部10个工具入口。它证明外部Host可组合P2-D合同，不表示任何
默认网络、文件、memory或scheduler backend，也不形成Public/Beta/Stable承诺。

## P6-A2 LLM Task收口

canonical `tools/llmtask`拥有单次LLM-only JSON子任务的参数兼容、model input、
`tool_choice=none`、JSON/schema校验、response budget和timeout/cancellation。HS原入口继续
存在，但只负责把既有全局`llmx.ChatWithInput`选择作为Host adapter注入；新项目必须显式提供
`ChatWithInputFunc`，canonical不会读取credential或选择provider。

`tools/conformance/llm-task-consumer`固定tools pseudo-version且无`replace`/HS import，证明新
项目可以只依赖canonical package完成注册和执行。Session/Task Host Kit可复用advanced
composition API，但Store、Scheduler、visibility、durable lifecycle和生产授权仍属于Host。

## 尚未迁移

M5S没有开启新的HS source-authority迁移。它把已经完成cutover的根Client、
Model/Tool Host Kit和Workflow Host Kit固定为三条标准接入路径，并用一个新的
fixed-version external consumer统一验收。HS继续固定各owner已验收的不可变
pseudo-version；本轮不为了版本外观一致而触碰Scene或重写业务consumer。

A股evaluator DTO的canonical定义已收口到`scenes/astock/contracts`，
`scenes/astock`和三组内部Pack只保留同一类型身份的alias。该调整用于消除候选
公开类型对Go`internal`路径的泄漏，不改变HS evaluator调用、JSON或产品策略。

W2-B结论统一为 `not_ready_for_hostless_w2b`。当前没有无需调用方提供 model/tool
adapter、policy或 backend的完整 embedded Runtime；Workflow和Objective已有独立
Developer Preview Host Kit，但没有成为根 `Client` mode。完整 Scene、Task/Session、
Resume、concrete durable backend、正式发行和 Public/Beta/Stable仍不在当前范围。
