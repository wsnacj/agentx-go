# 成熟度与兼容边界

## 当前状态

| 项目 | 状态 |
| --- | --- |
| 仓库 | private validation |
| 当前产品里程碑 | M5N Portable Objective Runtime Step, Executor Contract and Productization Projection：`technical_checkpoint_complete_awaiting_owner_acceptance` |
| 根 contract | Developer Preview candidate；未承诺兼容性 |
| LLM contract component | W3-01 已落地，Experimental |
| agentx-go production packages | M5N canonical landing：43个package、181个production source、59,412行；仍为8个Developer Preview candidate |
| Immutable AssetFS | M5A迁入完整 snapshot/fingerprint/resolver implementation，Experimental |
| Extensions module | 单一private-preview共享module；ProductShell temporary planning继续使用既有Experimental package，不新增package或成熟度等级 |
| Shared Host HTTP | M4A 已迁入3个 Experimental owner；只提供 transport/request/policy mechanism，不拥有 Scene handler或 backend |
| Workflow portable Runtime | composition及其下游 canonical owner已落地；Workflow Host Kit已形成标准入口 |
| Runtime construction lifecycle | W5-A 已迁移并完成 HS production cutover |
| Run/Open Tool Loop mechanism | W5-B～W5-E 已迁移多轮驱动、检测器/fuse、round coordination、phase ordering与 Host-backed assembly；host executor 仍必需 |
| Run/Artifact数据平面 | M5F已完成 `runtime/runstore`与`runtime/artifact` Experimental owner、fixed consumer和 HS production cutover；durable/file/object backend仍由 Host拥有 |
| Core Run dispatch/result assembly | W5-F 已迁入 `runtime/execution`并完成 HS production cutover；engine投影仍由 host持有 |
| Portable Core Host Kit | W5-G 已组合 no-HS-Runner执行；W5-H 已迁入 model/tool round implementation并完成 HS production cutover |
| 普通新项目接入 | Open Tool Loop可用 `NewModelToolClient`；Workflow可用 `workflow/hostkit.New`，两者仍显式要求 Host capabilities |
| 无需 host-provided adapter/policy 的完整 Runtime | `not_ready_for_hostless_w2b` |
| Examples/conformance | 根合同、LLM、Runtime及Extension fixed-version consumer已提供；M5N controlcontract consumer固定版本并覆盖Objective runtime/executor/productization；均无HS/Runner/Scene/长期replace/network |
| HS canonical import | W1-C、M5A～M5N production consumer均使用固定 private pseudo-version |
| HS LLM contract authority | W3-01 已切换到 `components/llm`，旧路径为 Deprecated shim |
| 当前 package surface | 43个全部纳入中文 Reference矩阵；8个进入 Developer Preview candidate signature/doc gate |
| Public/Beta/Stable | 未授权；Developer Preview candidate不等于任一正式等级 |
| 正式 tag/semver | 未授权 |
| License/NOTICE/release security | 仍是发行门禁 |

## 三类结论必须分开

- **API文档正文覆盖**：当前43个 production package均纳入中文 Reference；只说明
  实际签名、语义和 non-goal已经被描述。
- **兼容性承诺**：当前没有 semver、Public/Beta/Stable承诺；Developer Preview
  candidate签名门禁用于发现意外漂移，不等于禁止经审阅的变更。
- **正式发布成熟度**：tag、license/NOTICE、security/legal、维护 owner和 release
  authorization仍未完成，继续 fail closed。

根合同的 authoring/source authority 已在 W1-C 后转移到本仓库。HS 旧
`experimental/facade` 只保留 Deprecated alias/forwarder，不再接受独立行为修改。

LLM 合同类型的 source authority 已在 W3-01 后转移到
`github.com/wsnacj/agentx-go/components/llm`。HS
`core/llmx/types` 只保留 Deprecated alias/forwarder；其余 consumer 将在后续
小批次切换，不能继续向旧路径增加行为。

Runtime module 已拥有 protocol、telemetry、budget、prompt context以及
Workflow Spec/schema/validation/lowering/state/transition/journal/nodeexec/
orchestration/composition等 portable source authority，也拥有基于窄 Host port
的 construction lifecycle，以及基于窄 `Stepper` port 的确定性 Run/Open
Tool Loop 驱动、循环检测和 failure fuse，并拥有组合 driver、coordinator、
termination capture与 final portable state的单次 Run `Assembly`。HS production
consumer固定使用 canonical pseudo-version；W5-F又让 `runtime/execution`
成为根 Client到 Host的 adapter dispatch/result implementation owner。对应通用
implementation不得在 HS恢复双写。

W5-G 又让 `runtime/hostkit` 统一拥有 per-run portable assembly、
outcome/result projection、execution adapter与根 Client组合。W5-H继续迁入真实
`ModelToolRoundAdapter`，固定 model request、response observe、before-tools gate
和 tool execution的 portable协调机制；HS Open Tool Loop production path已消费
该 owner并删除同义通用 phase executor。`NewModelToolClient`进一步让普通新项目
无需自定义 Factory、BuildRun assembly或 RoundExecutor，但 provider、授权、
backend和产品策略仍由调用方显式注入。

M3E又让 `runtime/workflow/hostkit`组合已经迁入的 Workflow真实 owner。普通调用方
只导入 `workflow`和 `workflow/hostkit`，显式提供 validator、mapper、executor、
identity、clock和可选 durable port；HS `ExecuteInline`已切到同一入口。Host Kit
不复制 lowering/orchestration语义，也不提供 concrete policy或 backend。

M4A将原 HS shared Host HTTP实现迁入 `runtime/hosthttp/hostserver`、
`requestjson`和 `resourcepolicy`。它们拥有 bounded transport、request identity、
严格 JSON解码和 Host资源收窄 mechanism；HS旧路径只保留 Deprecated薄转发。
数据库检查的 handler、SQLite collector、readiness policy、OpenAPI和产品错误投影
继续由 HS Scene拥有，因此这三个 package仍是 Experimental extension，而不是通用
Scene Facade或发布承诺。

M5A已迁入 `runtime/assetfs`真实 immutable asset mechanism，并创建单一
`extensions` private-preview module承载首个 `astock/contracts` portable
implementation。它们不包含 A股 provider、livekit、pack/workflow、tool executor、
credential或网络，因此不表示完整 A股 Scene已可外部分发。

M5B又迁入 `extensions/domainmodule`真实portable registration coordinator，并让
九个 HS Scene module直接使用canonical合同。该 package仍为 Experimental extension；
HS Target和具体宿主/Scene策略没有被伪装成已迁移或可发布。

M5C已落地 `runtime/executionpolicy` portable DTO与 `extensions/pack`真实机制，
并已获 Owner接受。
Pack coordinator通过显式 Validator和 ToolArgumentLowerer ports保留 Host policy与
mapping边界；HS旧 Pack Core降为薄 compatibility adapter，真实 production与 fixed
external consumer均已切换。M5C checkpoint时 canonical production为32个 package、
90个 source、17,685行；`pack/runtime` memory/eval backend及具体 Scene内容仍由
HS拥有。该接受不构成 Public/Beta/Stable或发行授权。

M5D已建立 `extensions/astock`推荐入口与 `extensions/astock/hostkit`，并把三组
A股 Pack Definition/evaluator放入不可外部导入的 internal owner。当前37个 package
均已有中文 Reference，`extensions/astock`作为第八个 Developer Preview candidate
进入 focused签名门禁。fixed-version组合 consumer和 HS production cutover已经完成，
旧 Pack/tool-schema/Host Kit通用实现已删除或降薄；canonical production由 M5C的
32个 package、90个 source、17,685行增至37个 package、104个 source、20,674行。
fixed extensions版本为 `v0.0.0-20260801071806-57b903334bf5`，无 HS/Runner/长期
`replace`/网络的组合 consumer、module zip/cache、四 module gate和签名门禁均已
通过。HS本 cohort production由3,079行收缩为452行，整个 cutover提交净减少
2,631行；保留的是具体 registry/executor、Host adapter、provider/livekit以及
兼容资产入口。唯一一次 HS完整回归中新出现的陈旧 owner guard已迁到 canonical
source并 focused复测通过，剩余仍只有既有三个 evidence stale/mismatch；该
checkpoint已经 Owner接受。A股 provider、livekit、credential、cache、source
priority和真实网络继续由 HS拥有。

M5E又把7份、3,497行 Portable Skills真实实现迁入 `extensions/skills`，使 canonical
production增至38个 package、112个 source、24,176行。canonical owner覆盖 types/
normalization、loader/cache/generation/watcher、activation、requested semantics、
clone和 resource refs；HS删除对应实现，仅保留155行 alias/薄 forwarder及其合法的
catalog/filter/install/safety产品 owner。HS Skills从11个 source、5,781行降为5个
source、2,439行，净减3,342行，并已有真实 production consumer直接使用固定
`v0.0.0-20260801100244-e9b2f8a65ee4`。fixed consumer在无 HS、Runner、长期
`replace`、网络或命令执行时验证 immutable loader/cache、activation、semantics、
资源完整性和 deep clone。`extensions/skills`因目录 watcher尚无公共 `Shutdown`
合同而保持 Experimental，不自动成为第九个 Developer Preview candidate。该
checkpoint已经 Owner接受；上述source/LOC是M5E历史实测，不代表M5F最终规模。

M5F把Run数据合同、NodeExecution/Event存储port与内存实现收口到
`runtime/runstore`，并把Artifact metadata、lineage、registry/blob port与内存实现
收口到 `runtime/artifact`。两个package均保持 Experimental，Developer Preview
candidate仍为8个。canonical code commit为 `075f8158b05a`，fixed runtime版本为
`v0.0.0-20260801104654-075f8158b05a`；无HS/Runner/长期`replace`/网络的 consumer
输出 `agentx-run-data-plane-ok:event-1:artifact-consumer-1:1:1`。HS的durable
RunStore backend、Store adapter、文件BlobStore、private file安全写入和具体Artifact
runtime投影继续由Host拥有。canonical production实测为40个package、119个source、
25,369行；M5F checkpoint已经Owner接受，但不构成成熟度晋级或发行授权。

M5G在不迁入产品策略的前提下新增两个真实 owner：`runtime/cases`拥有 Case数据合同、
规范化/复制 helper和最小Store port；`extensions/productshell`拥有无副作用输入投影、
Shell/Case/Workflow绑定helper，以及通过显式 `PreparationRuntime` Host port运行的固定
准备顺序。两者均为 **Experimental extension**，没有进入8个 Developer Preview
candidate，更不构成 Public、Beta或Stable。canonical提交分别为 `651dab4f0a53`和
`cd5d97b84728`；当前fixed extensions版本为
`v0.0.0-20260801114445-cd5d97b84728`。M5G新增13个production source、1,959行，
使canonical实测增至42个package、132个source、27,328行。

ProductShell接入明确分成两阶段：先用canonical helper完成纯输入投影，再由
`PreparationPipeline`按固定顺序调用Host提供的session binding、command、Pack/Case、
Workflow和validation port。fixed-version consumer位于
`extensions/conformance/productshell-consumer`，不依赖HS、Runner、长期`replace`、
provider或网络。自然语言规划、LLM、产品推断、Pack/Workflow/Case backend、持久化、
Objective和Scene继续留在Host；`PrepareResult`只表示准备完成，不表示Workflow已经执行。
当前W2-A closure按既有口径为 `406/57/2369/65`，owner分布为
`HS 119 / canonical 32 / Scene 0`。该增量只证明source authority边界可迁移，不改变
`not_ready_for_hostless_w2b`，也不解除任何发行门禁。

M5H继续在同一个 `extensions/productshell` Experimental owner内迁入typed Session、
HostProcess与OperatorLine observation归一化，以及display-safe Host UI handoff、consumer
conformance和runtime-use机制。canonical提交为 `af05058a8a7f`，fixed extensions版本为
`v0.0.0-20260801133815-af05058a8a7f`。本轮没有新增package或Developer Preview
candidate；新增5个production source、1,239行，使canonical实测增至42个package、
137个production source、28,567行。

HS cutover提交 `ada1b785b` 已锁定上述fixed版本：portable typed contract、builder、
handoff、conformance和runtime-use旧实现退出或降为alias/薄兼容层；engine与真实Host
surface改用canonical owner。raw聚合、parser、inventory和delivery仍留在HS。

fixed-version consumer位于
`extensions/conformance/productshell-observation-consumer`，在无HS、Runner、Scene、
长期`replace`、provider、网络或凭据时验证typed Session/HostProcess/OperatorLine到
display-safe envelope、conformance与runtime-use的纵向路径。raw diagnostics/tool
output parser、完整 `ObservationSnapshot`、process inventory、历史存储、readback、
authorization以及真实log/UI/HTTP delivery继续由Host拥有；本包也不把Runtime声明为
delivery source。该landing不改变 `not_ready_for_hostless_w2b`，不表示ProductShell
Runtime已经整体迁入，也不构成任何正式发行或兼容承诺。

M5H最终验证中，root/components/runtime/extensions的test、race、vet、tidy-diff和
list全部通过；`runtime/hosthttp/hostserver`在受限沙箱内的两项loopback启动测试在
非沙箱环境复测通过。fixed consumer、42/8 API/doc gate、import direction及module
cache/zip回读均通过。HS唯一一次完整回归为149个package PASS、2个既有治理package
FAIL、20个无测试package，3个失败测试仍由同一条既有evidence source-scope
stale/mismatch产生，没有新增功能回归。W2-A closure为 `406/57/2374/65`、
`HS119/canonical32/Scene0`；新增的5个source位于既有canonical package内，因此不改变
package owner分布或 `not_ready_for_hostless_w2b`。本状态是
`technical_checkpoint_complete_awaiting_owner_acceptance`，不是Developer Preview、
Public、Beta、Stable或发行授权。

M5I继续复用`extensions/productshell`，canonical提交为`d6bac55`和`16d9426`，fixed
extensions版本为`v0.0.0-20260801144943-16d9426fd82a`；新增2个production source、
1,021行，使canonical实测增至42个package、139个production source、29,588行，仍为
8个Developer Preview candidate。fixed consumer无HS、Runner、Scene、长期`replace`、
真实provider、credential或网络。

HS cutover提交`9adb1d460`已让production consumer使用canonical planner，旧portable主体
删除或降为薄adapter。root、components、runtime、extensions的test/race/vet/tidy/list、
fixed consumer及API/doc/import gate均通过；完整HS回归为149个package PASS、2个既有
治理package FAIL、20个无测试package SKIP，没有新增失败。W2-A closure为
`406/57/2377/65`。M5I已于2026-08-01获Owner接受，
仍保持Experimental和`not_ready_for_hostless_w2b`，不构成Developer Preview晋级、
Public/Beta/Stable、正式tag、semver或任何发行授权。

M5J在独立准入后新增Experimental `runtime/controlcontract`，canonical production
达到43个package、141个source、32,447行；新增2,859行真实 status/evidence/
blocker/next-action、display-safe、projection与approval/budget/idempotency/lifecycle
implementation。runtime fixed版本为
`v0.0.0-20260801153451-11cc3fc9419e`，fixed consumer、module zip/cache、中文
Reference与43/8 API/doc gate已通过。HS对应2,856行通用source收缩为1,164行
alias/薄forwarder和未迁移文件仍需的private compatibility helper，production净减
1,692行；exported type identity、方法可调用性、JSON与reducer行为由focused
differential保持。完整AgentX回归中的no-space连锁失败经释放本轮临时cache后focused
复测全部通过，最终仍只有2个既有治理package/3个evidence stale测试失败，没有新增
功能回归。M5J已于2026-08-02获Owner接受，仍不构成Developer Preview晋级、
Public/Beta/Stable或发行授权。

M5K随后在同一Experimental package内新增8个Objective definition/strategy/graph
production source与一个private helper source，共6,365行真实implementation。HS对应
6,322行source已收缩为648行alias/forwarder，并只另留187行未迁移controlplane代码仍需的
private compatibility seam；fixed consumer现覆盖Objective Graph validation。四module、
43/8 API/doc gate、module zip/cache与完整HS回归均已闭合，技术checkpoint已获Owner接受。
M5K不构成Developer Preview晋级、Public/Beta/Stable或发行授权。

M5L继续在同一Experimental package内迁入8个、4,330行Objective required-evidence、
semantic verification、verification gate、recovery和replanning完整source，并把portable
`ObservationNormalizationResult`合同与private helper收口到canonical。canonical新增
10个production source、4,481行，达到43个package、160个production source、43,293行；
Developer Preview candidate仍为8个。具体runtime/production adapter result翻译、
authorization、Objective executor/scheduler、RunStore/backend、ProductShellRuntime与
Scene继续由HS拥有。fixed runtime版本为
`v0.0.0-20260801164525-a99d16de1fcd`，consumer覆盖verification与recovery proposal；
HS `45d90a208`已完成production cutover，八个完整source由4,330行降为379行
alias/forwarder，具体normalizer继续留在Host。四module、fixed consumer、43/8 API/doc
gate、module zip/cache与单次完整HS回归均闭合；完整回归无新增失败，closure为
407/57/2399/65、HS119/canonical33/Scene0。当前状态为
`accepted_checkpoint`；Owner已于2026-08-02接受。该landing仍不构成Developer Preview
晋级、Public/Beta/Stable或发行授权。

M5M继续在同一Experimental package内迁入8个production source、7,592行真实Host effect
contract/admission/invocation implementation：统一independent-effect gate、workflow runtime
executor gate、capability/scheduler request-result-readback及adapter gate、repeated-success
memory proposal和memory apply gate。canonical提交`9a3f0e5e1d5c`使总量达到43个package、
168个production source、50,885行，Developer Preview candidate仍为8个。runtime fixed版本为
`v0.0.0-20260801172253-9a3f0e5e1d5c`；HS `22fdb7f78`把8个source由7,609行降为533行
alias/forwarder/private compatibility seam，净减7,076行，42个alias method退出HS manifest，
stable symbols由1,964降为1,922，stable debt保持590/590。managed ingress、Objective runtime
loop/executor/productization、delegation observation normalization、具体adapter/backend、
authorization、ProductShellRuntime与Scene继续由HS拥有。四module、fixed consumer、
43/8 API/doc gate、module zip/cache与单次完整HS回归均已闭合；完整回归为149 PASS、
2个既有治理FAIL、20 SKIP且无新增失败，closure为407/57/2407/65、
HS119/canonical33/Scene0。当前状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`；不构成成熟度晋级或发行授权，
Owner已于2026-08-02接受；该接受不构成成熟度晋级或发行授权。

M5N继续在同一Experimental package内迁入13个production source、8,527行真实Objective
runtime step、auto-delegation、Host-owned executor request/result/readback、adapter gate和
runtime productization implementation，使canonical达到43个package、181个production
source、59,412行。具体adapter/executor、authorization、RunStore/backend、durable write、
ProductShellRuntime与Scene仍由HS拥有。runtime fixed版本为
`v0.0.0-20260801180223-57ea36658ea2`；fixed consumer、中文Reference、四module、43/8
API/doc gate、module zip/cache与单次完整HS回归均已闭合，回归没有新增功能失败。HS按
完整owner拆分口径由8,396行收缩为795行，净减7,601行；closure为407/57/2421/65、
HS119/canonical33/Scene0。当前状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`，W2-B继续
`not_ready_for_hostless_w2b`；等待Owner接受，不自动进入下一wave。

## 明确 non-goal

以下能力没有进入当前根 Facade，不得根据 package名、文档愿景或未来目录推断
已支持：

- 无需 host-provided model/tool adapter和 policy的完整 embedded Runtime根构造；
- Tool Direct Answer 的独立结果策略；
- 根 `agentx` Facade直接暴露的 Workflow 图执行；
- Objective Runtime Loop；
- 长任务调度、子 Session 和 durable lifecycle；
- Resume、progress/event stream；
- concrete Scene service/CLI/provider/credential与带真实副作用的 handler；
- 真实网络或生产副作用。

## 后续晋级

M3D/M3E/M4A checkpoint只证明：

- 公共合同可以作为独立零第三方依赖 module 构建和测试；
- external-style consumer 能通过 canonical import 使用它；
- HS adapter 可以依赖同一份合同而不反向污染 owner package。

W2-A 已验证普通使用者需要的窄 Runtime construction形状；W5-A 又把通用构造
生命周期迁入 canonical `runtime/construction`，W5-B～W5-E 又把真实多轮驱动、
循环/重放检测、failure fuse、portable round-result coordination和 phase
ordering，以及 Host-backed single-run assembly迁入 `runtime/toolloop`，W5-F
再迁入顶层 adapter dispatch/result assembly，并分别完成 fixed-version consumer
与 HS production cutover。W5-G/W5-H 已把这些 owner组合为可运行 Host Kit和
portable model/tool round adapter；低样板路径仍要求调用方显式提供 model request、
concrete tool execution及可选 policy。answer-contract、持久化和官方 concrete
provider/tool kit尚未成为 canonical owner，因此 W2-B结论保持
`not_ready_for_hostless_w2b`。当前 Experimental package拥有真实实现和
consumer，不代表其全部导出符号已经通过 Public API审批。Pre-Beta还必须完成
hostless construction决策、surface consolidation、版本兼容、维护 owner、
Changelog、license、security和发布授权门禁。任何一个步骤都不能自动把当前合同
升级为 Public、Beta或 Stable。
