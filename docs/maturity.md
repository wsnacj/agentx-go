# 成熟度与兼容边界

## 当前状态

| 项目 | 状态 |
| --- | --- |
| 仓库 | private validation |
| 当前产品里程碑 | M5B Domain Module Developer Preview foundation active |
| 根 contract | Developer Preview candidate；未承诺兼容性 |
| LLM contract component | W3-01 已落地，Experimental |
| agentx-go production packages | 根合同、LLM、26个 Runtime owner及2个 Extension owner，共30个 |
| Immutable AssetFS | M5A迁入完整 snapshot/fingerprint/resolver implementation，Experimental |
| Extensions module | 单一 private-preview共享 module；当前含 A股 portable contracts与 Domain Module coordinator |
| Shared Host HTTP | M4A 已迁入3个 Experimental owner；只提供 transport/request/policy mechanism，不拥有 Scene handler或 backend |
| Workflow portable Runtime | composition及其下游 canonical owner已落地；Workflow Host Kit已形成标准入口 |
| Runtime construction lifecycle | W5-A 已迁移并完成 HS production cutover |
| Run/Open Tool Loop mechanism | W5-B～W5-E 已迁移多轮驱动、检测器/fuse、round coordination、phase ordering与 Host-backed assembly；host executor 仍必需 |
| Core Run dispatch/result assembly | W5-F 已迁入 `runtime/execution`并完成 HS production cutover；engine投影仍由 host持有 |
| Portable Core Host Kit | W5-G 已组合 no-HS-Runner执行；W5-H 已迁入 model/tool round implementation并完成 HS production cutover |
| 普通新项目接入 | Open Tool Loop可用 `NewModelToolClient`；Workflow可用 `workflow/hostkit.New`，两者仍显式要求 Host capabilities |
| 无需 host-provided adapter/policy 的完整 Runtime | `not_ready_for_hostless_w2b` |
| Examples/conformance | 根合同、LLM、Runtime和 Extension fixed-version consumer已提供 |
| HS canonical import | W1-C、M5A及 M5B consumer已切换固定 private pseudo-version |
| HS LLM contract authority | W3-01 已切换到 `components/llm`，旧路径为 Deprecated shim |
| 当前 package surface | 30个全部有中文 Reference；7个进入 Developer Preview candidate signature/doc gate |
| Public/Beta/Stable | 未授权；Developer Preview candidate不等于任一正式等级 |
| 正式 tag/semver | 未授权 |
| License/NOTICE/release security | 仍是发行门禁 |

## 三类结论必须分开

- **API文档正文覆盖**：当前30个 production package均有中文 Reference；只说明
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

## 明确 non-goal

以下能力没有进入当前根 Facade，不得根据 package名、文档愿景或未来目录推断
已支持：

- 无需 host-provided model/tool adapter和 policy的完整 embedded Runtime根构造；
- Tool Direct Answer 的独立结果策略；
- 根 `agentx` Facade直接暴露的 Workflow 图执行；
- Objective Runtime Loop；
- 长任务调度、子 Session 和 durable lifecycle；
- Resume、progress/event stream；
- Scene领域 handler/service、CLI、provider、credential；
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
