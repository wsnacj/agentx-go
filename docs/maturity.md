# 成熟度与兼容边界

## 当前状态

| 项目 | 状态 |
| --- | --- |
| 仓库 | private validation |
| 当前产品里程碑 | P1 Core七能力、P2 Providers/Tools、P3 Browser/Document与P4-A/P4-B Portable Scenes闭环完成；仍处于private validation |
| Developer Portal | 68 package/14 candidate focused API gate已通过；Public Transport因依赖Experimental controlcontract保持Experimental；完整Scene站点编排仍属于后续文档产品化 |
| Private validation readiness | `true`；空缓存私有VCS fixed-version consumer已通过 |
| Pre-Beta technical candidate | `true`；四module同版候选、Go 1.25.12漏洞扫描、离线consumer和Ubuntu远端复验已通过 |
| Public Beta readiness | `false`；license、具名安全/发行复核、正式兼容等级和release authorization仍阻断 |
| 根 contract | Developer Preview candidate；未承诺兼容性 |
| LLM contract component | W3-01 已落地，Experimental |
| agentx-go production packages | root/components/runtime/extensions/scenes focused surface当前68个package、14个Developer Preview candidate；providers/tools/browser/document使用各自独立module gate |
| Experimental providers | P2-A/P2-B已落地OpenAI-compatible、Anthropic Messages、Codex Responses真实client、transport/fault/retry/usage机制和fixed consumer；credential、token store、模型选择与生产网络仍由Host拥有 |
| Experimental tools | 7个package已落地tool合同、invocation kernel、diffs及message/filesystem/HTTP/memory/scheduler真实coordination，并完成统一fixed consumer与HS production cutover；授权、sandbox、credential与具体backend仍由Host拥有 |
| Experimental browser | P3-A已完成Browser runtime、browserd host、Browser tools、fixed consumer、HS cutover与跨平台focused gate |
| Immutable AssetFS | M5A迁入完整 snapshot/fingerprint/resolver implementation，Experimental |
| Extensions module | 单一private-preview共享module；ProductShell temporary planning继续使用既有Experimental package，不新增package或成熟度等级 |
| Shared Host HTTP | M4A 已迁入3个 Experimental owner；只提供 transport/request/policy mechanism，不拥有 Scene handler或 backend |
| Workflow portable Runtime | composition及其下游 canonical owner已落地；Workflow Host Kit已形成标准入口 |
| Runtime construction lifecycle | W5-A 已迁移并完成 HS production cutover |
| Run/Open Tool Loop mechanism | W5-B～W5-E 已迁移多轮驱动、检测器/fuse、round coordination、phase ordering与 Host-backed assembly；host executor 仍必需 |
| Run/Artifact数据平面 | M5F已完成 `runtime/runstore`与`runtime/artifact` Experimental owner、fixed consumer和 HS production cutover；durable/file/object backend仍由 Host拥有 |
| Core Run dispatch/result assembly | W5-F 已迁入 `runtime/execution`并完成 HS production cutover；engine投影仍由 host持有 |
| Portable Core Host Kit | W5-G 已组合 no-HS-Runner执行；W5-H 已迁入 model/tool round implementation并完成 HS production cutover |
| Portable Scheduler/Resume | P1-D已落地queue、dispatcher、lease heartbeat与bounded Resume Host Kit；生产durable backend、process与产品policy仍由Host拥有 |
| 普通新项目接入 | Model Conversation/Tool Direct Answer可用`runtime/hostkit`，Open Tool Loop可用 `NewModelToolClient`，Workflow可用 `workflow/hostkit.New`，Objective可用`objective/hostkit.New`，child worker lifecycle可用`session/hostkit.New`，bounded Resume可用`session/hostkit.NewResumeRuntime`；均显式要求 Host capabilities |
| 无需 host-provided adapter/policy 的完整 Runtime | `not_ready_for_hostless_w2b` |
| Examples/conformance | fixed-version consumer覆盖五条标准construction；其它focused consumer继续覆盖LLM、Runtime、Extension、channel、provider及P2-D十个tool入口；均无HS/Runner/Scene/长期replace，P2-D consumer也无真实副作用 |
| HS canonical import | W1-C、M5A～M5R production consumer已使用固定 private pseudo-version；M5R cutover已完成 |
| HS LLM contract authority | W3-01 已切换到 `components/llm`，旧路径为 Deprecated shim |
| 当前 package surface | 51个全部纳入中文Reference矩阵；10个候选有hash、可读snapshot、公开类型闭包和darwin/linux平台一致性门禁 |
| Public/Beta/Stable | 未授权；Developer Preview candidate不等于任一正式等级 |
| 正式 tag/semver | 未授权 |
| License/NOTICE/release security | 仍是发行门禁 |

## 三类结论必须分开

- **API文档正文覆盖**：当前51个 production package均纳入中文 Reference；只说明
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

M5D已建立 `scenes/astock`推荐入口与 `scenes/astock/hostkit`，并把三组
A股 Pack Definition/evaluator放入不可外部导入的 internal owner。当前37个 package
均已有中文 Reference，`scenes/astock`作为第八个 Developer Preview candidate
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
`not_ready_for_hostless_w2b`；Owner已于2026-08-02接受，该接受不构成成熟度晋级或发行授权。

M5O在同一Experimental package内落地31个、23,158行portable Host adapter、Managed
Objective Ingress与closeout projection source，使canonical达到43个package、212个
production source、82,570行。原31组合同测试和external-package fail-closed测试通过；
该cohort只拥有display-safe request/result/readback、gate和composition，不执行真实
adapter、authorization、store、Runner或UI副作用。具体Host实现继续留在HS。fixed
runtime为`v0.0.0-20260801184120-1fdddbc8d76f`；HS 23,217行基线已收缩为1,276行
alias/forwarder/private Host seam，净减21,941行。fixed consumer、中文Reference、四module、
43/8 API/doc gate与有效完整回归均闭合，closure为407/57/2453/65、
HS119/canonical33/Scene0。当前状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`，W2-B继续
`not_ready_for_hostless_w2b`；Owner已于2026-08-02接受，该接受不构成成熟度晋级或
发行授权。

M5P随后以单一cohort迁入7个、4,210行portable Objective completion、durable-store
projection、async delegation/controller/handoff和strategy metadata真实implementation，
使canonical达到43个package、219个production source、86,780行。fixed runtime为
`v0.0.0-20260801192832-e4c46c0a6b76`；HS 4,210行基线收缩为296行alias/forwarder，
净减3,914行。fixed consumer、中文Reference、四module、43/8 API/doc gate和有效完整
回归均已闭合，closure为407/57/2460/65、HS119/canonical33/Scene0。concrete
store/backend、LLM/provider、worker scheduler/dispatch、`observation_normalizer`、Runner、
Scene和真实副作用继续留在HS。当前状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`；该checkpoint不构成新增成熟度
承诺或发行授权。

M5P已于2026-08-02获Owner接受；该接受不构成成熟度晋级或发行授权。M5Q随后准入
现有`runtime/executionpolicy`类型直接需要的6-source、1,487行portable Snapshot
metadata、bounded retry command、runtime decision/packet、execution-loop report与
soft-rejection kernel。compiler/derive、authorization、sandbox、具体browser/network/
gateway policy、Host backend和真实副作用不进入M5Q。

M5Q canonical已在既有Experimental `runtime/executionpolicy`落地上述6个source和原
合同测试，使agentx-go达到43个package、225个production source、88,267行；新增kernel
只生成typed command/report/reducer，不执行Host副作用。fixed runtime为
`v0.0.0-20260801195859-a83a74b77995`；无HS/Runner/Scene/长期`replace`的consumer、中文
Reference、四module test/race/vet/tidy/list、43/8 API/doc gate与module zip/cache均已闭合。
HS原6-source、1,487行authority已收缩为285行alias/forwarder，并保留57行只服务具体
browser/network/gateway/runtime policy的private Host seam；engine、hostruntime和hostwiring
production consumer已直接使用canonical owner。完整回归为149 PASS、2个既有治理FAIL、
20 SKIP且无新增功能失败，closure为407/57/2467/65、HS119/canonical33/Scene0。状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`；该checkpoint不构成成熟度晋级
或发行授权。

M5Q已于2026-08-02获Owner接受；该接受不构成成熟度晋级或发行授权。M5R随后准入
HS `runtime/channel`中除pairing外的8个production source、1,400行portable message、
sender/runner port、routing、chunking、in-memory dedupe、bounded ingress runtime与
session delivery contract。隔离module复用全部非pairing原合同测试后通过；production只
依赖标准库与canonical safeerror。pairing challenge/approval、durable state、access policy、
具体platform sender/SDK、credential、backend和真实网络副作用不进入M5R。

M5R canonical `9f67493e1e32`已新增Experimental `runtime/channel`，落地8个迁入
implementation source与1个package doc source，共1,405行；agentx-go达到44个package、
234个production source、89,672行。runtime fixed版本为
`v0.0.0-20260801203033-9f67493e1e32`，sum为
`h1:UjoEhQXVLUgAxv/APnvBpgmwIalAZa8RsId5UeNdxng=`；中文Reference和无HS/Runner/Scene/
长期`replace`/network的consumer由`02ebc3d`闭合。HS `de3b100d9`已把18个production
callsite直接切到canonical owner，原8-source、1,400行portable authority收缩为249行
alias/forwarder，净减1,151行；pairing、durable state、access policy与具体平台发送继续
留在HS。四module test/race/vet/tidy/list、44/8 API/doc gate、module zip/cache、fixed
consumer与唯一完整HS回归均已闭合；回归为149 PASS、2个既有治理FAIL、20 SKIP，
没有新增功能失败。closure保持407/57/2467/65，但owner分布从HS119/canonical33变为
HS118/canonical34、Scene0，证明HS channel package退出production construction closure。
当前状态为`technical_checkpoint_complete_awaiting_owner_acceptance`，W2-B继续
`not_ready_for_hostless_w2b`；本checkpoint不构成Developer Preview晋级、
Public/Beta/Stable或发行授权。

M5R已于2026-08-02获Owner接受；该接受不构成成熟度晋级或发行授权。M5S随后停止以
迁移LOC扩展Core，转向现有44个package和8个Developer Preview candidate的产品收口。
当前标准construction固定为自定义`ExecutionAdapter`、Model/Tool Host Kit与Workflow
Host Kit三条显式Host capability路径；完全hostless Runtime继续
`not_ready_for_hostless_w2b`，但作为本次Developer Preview的deferred non-goal。
`runtime/controlcontract`继续是Experimental迁移/组合底座，不进入推荐入口，也不得由
M5S自动升级其巨大exported surface。M5S只补候选API/平台差异、fixed consumer、依赖
足迹与中文文档门禁，不进入Scene、完整文档站、Beta/Stable或发行。

M5S技术checkpoint已完成。三条标准construction现在由根
`conformance/consumer`在同一组固定版本
`v0.0.0-20260802021959-5a41fb0ccb87`上验证：自定义`ExecutionAdapter`、
`runtime/hostkit.NewModelToolClient`和`runtime/workflow/hostkit.New`。该consumer
无HS、Runner、Scene、长期`replace`、provider、credential或网络副作用；test、race、
vet、tidy、offline run和darwin/arm64、linux/amd64 CGO-disabled build均通过。

8个Developer Preview candidate现在同时拥有`go doc -all`hash与可读snapshot；gate
检查44个package分类、中文Reference、公开类型不得泄漏`hs/`、Go`internal`或
`runtime/controlcontract`，并比较darwin/arm64与linux/amd64签名。为关闭唯一真实泄漏，
A股六个evaluator DTO的source authority从三组`internal` Pack移到
`scenes/astock/contracts`，顶层推荐入口和内部实现均使用type alias；evaluator逻辑、
JSON字段和调用顺序未改变。Markdown门禁验证68个文件、248个本地链接。

同一fixed commit的root/components/runtime/extensions module zip分别为102,907、
27,733、1,026,836和246,270字节，Origin均回读`5a41fb0ccb87`；统一consumer只有
canonical module和标准库依赖，总依赖包79个，darwin/arm64二进制3,327,682字节、
linux/amd64二进制3,369,181字节。四module test/race/vet/tidy/list、module cache/zip、
import direction和HS focused compatibility全部通过；唯一完整HS回归为149 PASS、2个
既有治理FAIL、20 SKIP，仍仅有3个既有evidence stale/mismatch。W2-A closure保持
407/57/2467/65、HS118/canonical34/Scene0；M5S不新增HS construction依赖，因此不借
产品收口修改该闭包。M5S已于2026-08-02获Owner接受，状态为`accepted_checkpoint`；
该接受不构成Public/Beta/Stable、semver、tag或发行授权。

M5T随后只关闭消费与升级层面的版本偏差：以
`v0.0.0-20260802021959-5a41fb0ccb87`作为四module候选统一基线，补中文升级/回滚说明、
轻量版本一致性门禁与clean-room/no-replace验证，并在兼容差分通过后统一HS要求。
M5T不新增Runtime owner、不改变Scene业务逻辑，也不进入完整文档站或正式发行。

M5T技术checkpoint现已完成。根`conformance/consumer`直接固定四module统一版本，并
使用fixture覆盖三条标准construction和A股推荐extension入口；版本gate与仓库外
clean-room gate均通过。HS只修改`go.mod/go.sum`并统一同一版本，未修改Core或Scene
生产源码；外部依赖集合不变。canonical四module test/race/vet/tidy/list、44/8双平台
API gate、69文件/254链接文档gate和cache Origin回读通过；完整HS回归保持149 PASS、
2个既有治理FAIL、20 SKIP，仍只有原3条evidence stale/mismatch。production规模与
closure保持44包、235 source、89,691行及407/57/2467/65、HS118/canonical34/Scene0。
M5T已于2026-08-02获Owner接受，状态为`accepted_checkpoint`；该接受不构成
Public/Beta/Stable、semver、tag、Scene迁移或发行授权。

M6A随后只补Developer Preview兼容、维护、支持、安全报告、版本epoch和分发预检合同，
并将private validation与Public Beta readiness分开。M6A不新增Runtime owner、不修改
公共行为、不进入Scene或完整文档站，也不擅自选择license或创建tag/release。

M6A技术checkpoint现已完成：仓库新增`SECURITY.md`、`SUPPORT.md`、
`CONTRIBUTING.md`、`CODEOWNERS`、Developer Preview政策和分发Readiness页；本地
preflight复用既有版本、API、文档、clean-room与module artifact gate。四module固定
版本已从空`GOMODCACHE`经私有VCS重新获取，Origin均回读`5a41fb0ccb87`，无HS
consumer完成构建和运行。四module `test/vet/tidy-diff/list`全部通过，HS Facade、
Workflow、Engine与ProductShell focused兼容性通过。M6A没有修改固定消费版本、Runtime
实现或HS生产代码，故不重复M5T完整回归，继续引用其149 PASS、2个既有治理FAIL、
20 SKIP且无新增失败的最近基线。当前结论为`private_validation_ready=true`、
`public_beta_ready=false`；状态为`technical_checkpoint_complete_awaiting_owner_acceptance`。

M6A已于2026-08-02获Owner接受；该接受不改变`public_beta_ready=false`。M6B随后只把
已有中文正文、44个package Reference、8个Developer Preview candidate和三条标准接入
路径交付为可构建、导航、搜索的Core Developer Portal Candidate。站点复用现有
Markdown/API snapshot事实源，不新增Runtime owner、公共API、Scene或正式发行声明。

M6B技术checkpoint现已完成：VitePress 1.6.4与修复后的Vite 6.4.3由lockfile固定，
clean `npm ci`和`npm audit`零漏洞通过；确定性prepare只把现有Markdown、package
`API.md`和maturity TSV投影到ignored staging。production build生成85个HTML页面，
44 package/8 candidate coverage、本地全文搜索、断链、生成产物clean和敏感占位gate
通过。浏览器已验证首页、Package API、`Shutdown`搜索、Developer Preview badge、
source commit backlink、侧边栏、390×844响应式无水平溢出和零console error。四module
smoke tests全部通过。M6B不改变fixed版本、Go production source、Runtime行为或HS代码，
故不重复完整AgentX回归；当前状态为
`technical_checkpoint_complete_awaiting_owner_acceptance`。

## P1-A Model Conversation / Tool Direct Answer checkpoint

P1-A在不增加新module、不引入provider或业务规则的前提下，为`runtime/hostkit`增加
`NewChatClient`、`ToolDirectAnswer`和统一`ExecutionResult`投影。固定版本
`v0.0.0-20260802072349-5311859981e9`已由无HS consumer验证；HS Open Tool Loop生产路径
使用同一portable结果策略，既有`answer_contract` JSON、recovery、persistence、telemetry、
错误文本与stop reason保持Host owner。8个Developer Preview candidate数量不变，
`runtime/hostkit`签名基线经审阅更新；这不构成Public/Beta/Stable或正式发行。

## P1-B Objective Host Kit checkpoint

P1-B新增Experimental `runtime/objective`最小类型入口和Developer Preview candidate
`runtime/objective/hostkit`。推荐路径现在可以在无HS、无Runner、无provider默认的进程中
组合managed ingress、显式Host runtime-adapter dispatch、observation normalization和
Objective verification。固定版本`v0.0.0-20260802080954-21919fd8e06a`由独立consumer
验证成功、确认门禁和取消路径。

通用observation normalization与315行dispatch facade的source authority已从HS转移到
canonical；HS production productshell consumer通过固定版本使用新实现，原代码降为薄
alias/forwarder与共享private格式helper。具体policy、approval、credential、handler、
backend、持久化和副作用仍由Host拥有。当前只完成C5同步Host adapter闭环，不包含C6
Task/Session/Subagent、Scheduler/Resume或durable long-run，也不构成正式发行。

## P1-C Task / Session / Subagent Host Kit checkpoint

P1-C新增Experimental `runtime/session`和Developer Preview candidate
`runtime/session/hostkit`。Host显式提供child `WorkerRuntime`、`StateStore`与
control-contract readiness；canonical implementation负责exact-once worker invoke、durable
record、按ref回读、一致性校验、parent verification以及可选Objective handoff。

固定版本`v0.0.0-20260802091415-920282587efc`由无HS、无Runner、无长期
`replace`的独立consumer验证；HS delegation、ProductShell和interactive CLI生产路径
已直接使用canonical合同。HS四个通用兼容文件从1,562行收缩为79行，净减
1,483行，并有owner gate防止实现回流。concrete worker、queue、scheduler、
credential、backend和产品policy仍属于Host。P1-C闭合C6的child worker lifecycle
垂直切片，不声称Scheduler/Resume/long-task orchestration或正式发行已完成。

## P1-D Scheduler / Resume / Long Task implementation landing

P1-D新增Experimental `runtime/scheduler`真实实现，并在Developer Preview candidate
`runtime/session/hostkit`内组合bounded `ResumeRuntime`。canonical当前拥有typed Job/Result、
三lane并发安全Queue、QueueProxy、Dispatcher、lease heartbeat、cancellation、terminal
coordination、metrics，以及continuation readback到Host wake dispatch的有界service loop。

Host仍显式拥有concrete durable queue/store、process/service lifecycle、authorization、
credential、backend与产品retry/priority policy。该landing不等于系统scheduler或无Host的
完整long-task平台。fixed版本`v0.0.0-20260802103826-c7d80001682e`的独立consumer同时验证
child worker与resume路径；HS production已经切到canonical scheduler和resume service，原
10个portable source共3,902行收缩为396行兼容/Host helper，净减少约3,506行。canonical
四module test/race/vet/tidy/list、50 package/10 candidate API gate、fixed consumer与分发
检查均通过；HS有效完整回归为149 PASS、2个既有治理FAIL、20 SKIP且无新增功能失败。
P1-D状态为`completed_checkpoint`，不构成Public/Beta/Stable或正式发行授权。

## P1-E Deterministic Domain Kit implementation landing

P1-E新增Experimental `extensions/domainkit`真实implementation：构造时统一规范化manifest、
拒绝重复module、manifest外handler与缺失handler；运行时按显式module/tool exact-once调用
Host handler，不调用模型、不进行自然语言路由或provider选择，并输出稳定SHA-256 digest。
typed error保留cause identity，同时提供独立display-safe message。该实现不包含provider、
credential或真实副作用。fixed版本`v0.0.0-20260802113655-f41de95ec5be`由无HS/Runner/Scene/
长期`replace`的domain-module consumer重复fixture验证；HS A股production tool protocol已
使用canonical Runtime并删除本地通用结果收口。canonical四module、51/10/2 API gate、
83份Markdown/303链接、95页Portal与artifact Origin通过；HS完整回归为149 PASS、2个既有
治理FAIL、20 SKIP且无新增失败。P1-E状态为`completed_checkpoint`，不构成Scene迁移或
Public/Beta/Stable授权。

## P2-C Tool Catalog / Diffs implementation landing

P2-C先在`components/tool`收口provider-neutral tool合同，直接复用
`components/llm`的Definition、Call和Result type identity，避免形成第二套wire DTO图。
独立`tools` module拥有并发安全Registry、稳定definition排序、version/reset、保守名称
修复与typed name error的真实source authority；`tools/diffs`拥有首个纯文本通用tool实现。

固定版本`v0.0.0-20260802131439-56dd598eef59`由无HS、Runner、Scene、长期
`replace`、网络或credential的独立consumer验证。HS `core/llmx/tools`和AgentX diffs
production consumer已直接使用canonical版本，原551行通用source收缩为49行alias/adapter，
净减502行。authorization、approval、sandbox及filesystem/process/network/store backend
仍由Host拥有；P2-C不是“全部通用tools已迁完”的声明，后续cohort只能通过显式窄port继续
迁入portable mechanism。本checkpoint不构成Public/Beta/Stable或正式发行授权。

## P4-A 首批 Portable Scenes source-authority 闭环

P4-A建立独立`scenes` module并完成三组互补领域Kit。`scenes/astock`收正A股唯一
canonical owner；`scenes/publicnews`与`scenes/companyresearch`拥有只读研究合同、Pack、
evidence/quality机制和显式Host ports；`scenes/docparse`拥有profile、planner、fusion、
understanding、quality evidence、Pack、资产与无文件/无provider Host Kit，并复用固定
`document` module合同而不复制OCR/PDF实现。

四组canonical production共79个Go source、19,534行；三个fixed consumer分别锁定
`v0.0.0-20260802210630-68ea44b615cc`、
`v0.0.0-20260802213138-f16c8c734592`和
`v0.0.0-20260802220352-a8c81c2c9118`，在无HS、Runner、旧Scene路径、长期`replace`、
credential和真实网络的条件下通过normal/race/vet/run及module zip/cache/Origin回读。
HS production consumer已切换到fixed canonical owner；research提交净减少15,765行，docparse
提交净减少5,744行，A股旧`extensions/astock`重复authority已收正。

scenes module normal/race/vet/tidy/list、Linux/amd64 CGO-disabled build、65 package/13
candidate双平台API gate和import-direction gate均通过。HS四Scene focused normal/race/vet
通过；有效完整AgentX回归为148 PASS、2个既有治理package FAIL、20 SKIP。失败仍属于
deferred governance：3个既有evidence source-scope stale/mismatch，以及1个P3-A删除旧
`browserruntime`后尚未刷新的runtime-boundary inventory条目；没有P4-A功能回归。

本checkpoint仍不包含live provider、credential、客户policy、真实网络/文件副作用、业务结果
权威，也不构成Public、Beta、Stable、semver、tag或发行授权。

## P4-B Browser Ops Domain Kit source-authority 闭环

P4-B把HS既有Browser Ops portable实现机械迁入`scenes/browserops`：14个production Go
source、4,974行，覆盖六条Pack/Workflow、evidence bundle与state projection、六类确定性
evaluator，以及通过显式canonical Tool Executor驱动的Host Kit。真实browser backend、
profile/login、credential、approval、文件/artifact、livecapture与L1/L2/L3副作用策略继续留HS。

fixed consumer锁定`v0.0.0-20260802230400-97a7e59508f5`，不含`replace`、HS、Runner、旧
Scene import、真实浏览器、网络或credential；normal/race/vet/run及module zip/cache/Sum/Origin
回读通过。HS production路径升级到同一版本，Pack注册与module manifest直接消费canonical
owner；旧12个portable production文件删除，HS提交净减少4,496行，Browser Ops production
由12,139行降至7,528行，保留部分均为Host/Product职责或兼容forwarder。

scenes module normal/race/vet/tidy/list与Linux/amd64 CGO-disabled build通过，67 package/14
candidate双平台API/import gate通过。完整AgentX回归仍为148 PASS、2个既有治理package FAIL、
20 SKIP；失败仍只有3个evidence source-scope stale/mismatch与1个P3-A旧runtime-boundary
inventory条目，没有P4-B功能回归。首次并行全module验证因本轮隔离cache耗尽磁盘而失败；
仅清理本轮可重建cache后串行原命令通过，不计为代码失败。

本checkpoint不授权Public/Beta/Stable、semver、license、tag或release，也不把HS live与安全
策略视为待迁portable实现。

## P4-C Public Transport Domain Kit source-authority 闭环

P4-C在`scenes/publictransport`落地5个production Go source、1,132行真实implementation：
typed request/report/evidence/inventory合同、fail-closed normalization、provider-neutral exact-once
Coordinator、control-contract投影、库存evidence纯函数、确定性Evaluator与read-only
Pack/Workflow。真实ticket/map provider、endpoint/template、headers/cookie、credential、
station catalog、warmup/retry、缓存/限流、真实网络与购票副作用仍由HS Host拥有。

无HS、Runner、长期`replace`、网络或credential的fixed consumer锁定
`scenes@v0.0.0-20260802235605-d32ccb29a700`，normal/race/vet/run通过，输出
`agentx-publictransport-ok:public-transport-readonly-pack:public_transport_ticket_lookup_v1:1:true`。
HS production路径已升级到同一版本：`service`直接消费canonical evaluator，原
ticketlookup portable types、coordination、projection与evaluator删除或降为121行Deprecated
alias/forwarder。cutover提交增加174行、删除785行，净减611行；整个Scene
production source由12,644行降至11,996行。

scenes module normal/race/vet/tidy/list、import boundary与68 package/14 candidate/2 target API gate
通过。`publictransport`因导出签名显式依赖Experimental `runtime/controlcontract`，依据
fail-closed gate保持`Experimental extension`，没有被错误升级为Developer Preview。HS全Scene
normal、focused race/vet、engine/pack/inlinespec/business-host兼容测试通过。本wave唯一
完整AgentX回归仅失败2个既有治理package中的4项历史断言：3项evidence
source-scope stale/mismatch与1项P3-A删除旧browserruntime后的runtime-boundary记录漂移；
没有P4-C功能回归。

本checkpoint不授权Public/Beta/Stable、semver、license、tag或release，也不把真实
票务后端、凭据、安全/合规、商业条款或产品结果权威下沉canonical。

## P4-D Source Acquisition Domain Kit source-authority 闭环

P4-D在`scenes/publicsource`与`scenes/wechatarticle`落地12个production Go source、
1,631行真实implementation。Public Source拥有typed request/report/evidence/display-summary、
Host规则驱动的HTTPS/allowlist机制、provider-neutral search/document投影、exact-once
Collector协调、确定性Evaluator与read-only Pack/Workflow；WeChat Article拥有
login/account/article/dedup/download合同、显式Host Client port、可注入typed错误分类、
search/list/download协调、evidence digest、Evaluator与read-only Pack/Workflow。

无HS、Runner、长期`replace`、credential、真实网络或文件副作用的fixed consumer锁定
`scenes@v0.0.0-20260803013146-6ba7e01b9d26`，normal/race/vet/run与`GOPROXY=off`
module graph验证通过，输出为
`agentx-sourceacquisition-ok:public-source-readonly-pack:wechat-article-readonly-pack:1:1`。
HS Public Source service直接消费canonical Report/Collector/SourcePolicy；WeChat service与
objective直接消费canonical合同、strategy和evidence identity，原portable协调/类型实现删除
或降为Host translation/error-classifier薄层。HS cutover提交增加261行、删除1,404行；
Public Source Host package从3,474行降至2,708行，两Scene合计production从10,716行降至
10,259行。

scenes module normal/race/vet/tidy/list、Linux/amd64 CGO-disabled build、import boundary、
70 package/14 candidate/2 target API gate与77文件/399本地链接文档gate通过。HS相关Scene
normal/race/vet、engine/pack/inlinespec/business-host兼容通过；本wave唯一完整AgentX回归
仍只失败2个既有治理package的4项历史断言，没有P4-D功能回归。

HTTP/retrieval、exporter concrete client、credential/cookie、login QR、artifact/filesystem、
真实网络和产品结果权威继续留在Host。本checkpoint不授权Public/Beta/Stable、semver、
license、tag或release。

## P4-E Finance Domain Kit source-authority 闭环

P4-E在`scenes/globalstock`与`scenes/finance`七个package落地37个production Go文件、
7,274行真实implementation。Global Stock拥有证券事实合同、代码规范化、evidence/readiness、
exact-once investigation coordination与quote Pack/Workflow；Finance拥有财报合同、period/
metric/trend/evidence/readiness、exact-once lookup coordination与brief/metrics Pack/Workflow。
产品回答格式通过显式Host seam注入，付费数据/provider、credential、交易、docparse/tool
concrete executor、真实网络与免责声明不属于canonical owner。

无HS、Runner、长期`replace`、credential、网络或交易副作用的fixed consumer锁定
`scenes@v0.0.0-20260803022834-4043fbe78ff3`，normal/race/vet/run及module
zip/cache/Sum/Origin通过。HS两个production Scene已切换同一版本；旧portable contracts、
coordination、evaluator与Pack删除或降薄，production source从35,600行降至28,940行，净减
6,660行。canonical scenes模块normal/race/vet/tidy/list、Linux/amd64 CGO-disabled build、
77 package/14 candidate/2 target API gate与77文件/406链接文档gate通过；HS focused
兼容验证通过，完整AgentX回归没有遗留P4-E功能失败。

全部新package继续标记为Experimental extension，不构成Public/Beta/Stable、semver、
license、tag或release授权。

## 明确 non-goal

以下能力没有进入当前根 Facade，不得根据 package名、文档愿景或未来目录推断
已支持：

- 无需 host-provided model/tool adapter和 policy的完整 embedded Runtime根构造；
- 自动解析任意工具输出并决定 Tool Direct Answer 的产品策略（Host Kit只支持 Host显式提供
  display-safe `ToolDirectAnswer`后的 portable completion投影）；
- 根 `agentx` Facade直接暴露的 Workflow 图执行；
- 无Host policy/catalog/adapter注入的全自动Objective Runtime；
- 长任务调度、queue ownership和无Host的durable backend；
- Resume、progress/event stream与完整long-task orchestration；
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
