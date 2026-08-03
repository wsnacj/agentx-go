# Developer Preview 变更记录

本文件记录可供private consumer复现的Developer Preview checkpoint，不代表正式release、
semver兼容承诺或Public/Beta/Stable声明。

## Unreleased（P5 Core Developer Preview Closure）

- Pre-Beta本地技术门禁、distribution full gate和Ubuntu入口已从历史四module范围升级为
  当前九module release train，并新增无HS、无`replace`的九module聚合consumer；
- 九module本地候选已在Go 1.25.12上完成132 package、同版zip、固定漏洞扫描、聚合consumer
  与离线cache复验，已知可达漏洞为0；Public Beta仍因Ubuntu和具名责任决定保持关闭；
- 新增Pre-Beta准入合同与九module候选tag前缀。该合同不选择License、不创建tag，具名
  security/release责任和Public Beta授权继续fail closed；
- 九module的120个可外部import production package均有中文Reference，共123份`API.md`；
  focused signature gate覆盖77 package、14个Developer Preview candidate和2个平台；
- 新增九module机器可读fixed-version矩阵，并让版本与文档门禁核对代表性consumer，修复
  历史四module同版假设；
- 九modulenormal/race/vet/tidy/list、33个无`replace`external-style consumer、仓库外
  clean-room consumer及九module zip/Sum/Origin复验通过；production code无HS、Runner或
  Scene反向依赖；
- P5 closure没有发现阻断七类能力或推荐接入路径的portable owner缺口，因此没有继续迁移
  Operations、Capability Install、Database Query或其它Scene；
- 当前状态为private Developer Preview Candidate。Pre-Beta、Public/Beta/Stable、正式tag、
  License、security/legal和release authorization继续fail closed。

## v0.0.0-20260802131439-56dd598eef59（P2-C Tool Catalog / Diffs）

- `components/tool`复用既有LLM wire type identity，提供provider-neutral
  Definition/Call/Result/Handler/Executor合同，不复制第二套DTO；
- 新增独立Experimental `tools` module，落地并发安全Registry、稳定definition排序、
  version/reset、保守名称修复和typed name error真实implementation；
- 新增无文件、Git、网络或进程副作用的`tools/diffs`真实通用tool；fixed consumer不使用
  HS、Runner、Scene或长期`replace`；
- HS `core/llmx/tools`与AgentX diffs production consumer切至fixed canonical版本，旧通用
  source从551行收缩为49行；授权、sandbox和具体backend继续由Host拥有；
- 本版本仍是private Experimental验证，不授权Public/Beta/Stable、正式tag或发行。

## v0.0.0-20260802124746-c7f90139a1cc（P2-B Remaining Provider Cohort）

- 新增真实`providers/anthropic` Messages和`providers/codex` Responses/SSE client；
- Codex token store、OAuth刷新、用户目录、credential、模型路由和生产网络授权继续留在Host，
  canonical只通过显式`Authorizer`与`HTTPDoer`消费；
- fixed provider-cohort consumer在无HS、Runner、Scene、长期`replace`和真实网络下同时验证
  Anthropic文本/usage与Codex文本/tool call/usage/account identity；
- HS Anthropic/Codex production provider切到固定版本，payload/response/SSE旧通用源码删除，
  对应production source从1,815行收缩为834行；
- 本版本仍是private Experimental验证，不授权Public/Beta/Stable、正式tag或发行。

## v0.0.0-20260802121436-b8b4d7efb134（P2-A OpenAI-compatible Provider）

- 新增独立Experimental `providers` module与真实OpenAI-compatible chat、vision、
  embedding、bot和SSE stream client；
- provider-neutral transport hooks、fault分类、bounded retry与usage collector seam一并
  迁入，production code不依赖HS、Runner或Scene；
- credential、endpoint、local media和HTTP client通过显式Host seam注入，默认构造不读取
  环境变量、文件或credential store；
- fixed consumer无`replace`、HS或真实网络；HS `core/llmx/provider/http`切为canonical
  production consumer，旧portable authority从1,912行收缩为295行；
- 本版本仍是private Experimental验证，不授权Public/Beta/Stable、正式tag或发行。

## v0.0.0-20260802113655-f41de95ec5be（P1-E Deterministic Domain Kit）

- 新增Experimental `extensions/domainkit`真实无模型execution implementation：构造时
  校验完整manifest/handler闭包，运行时按显式module/tool exact-once调用Host handler；
- 输出保持string/bytes/JSON兼容，并附带稳定SHA-256 digest；typed error支持
  `errors.Is/As`且将display-safe message与cause文本分离；
- fixed `domain-module-consumer`不依赖HS、Runner、Scene或长期`replace`，覆盖注册与
  重复fixture执行；HS A股production tool protocol切换到同一canonical Runtime；
- provider、credential、authorization、真实backend/副作用与Scene业务判断仍由Host拥有；
  本版本仍是private Experimental/Developer Preview验证，不授权正式发行。

## v0.0.0-20260802103826-c7d80001682e（P1-D Scheduler / Resume / Long Task）

- 新增Experimental `runtime/scheduler`真实queue、dispatcher、lease heartbeat、terminal
  persistence、retry/dead-letter与metrics实现；
- 新增Experimental `runtime/session/resume`和Developer Preview candidate
  `runtime/session/hostkit.NewResumeRuntime`，组合continuation readback、Host wake dispatch、
  bounded daemon/service、幂等Shutdown和关闭后调用合同；
- fixed consumer无HS、Runner、Scene或长期`replace`，同时覆盖child worker和resume路径；
  HS scheduler/resume production consumer已cutover，原portable authority净减少约3,506行；
- 具体durable backend、process/system scheduler、credential、authorization和产品policy仍由
  Host拥有；本版本仍是private Developer Preview candidate，不授权Public/Beta/Stable、tag
  或发行。

## v0.0.0-20260802091415-920282587efc（P1-C Task / Session / Subagent Host Kit）

- 新增Experimental `runtime/session`和Developer Preview candidate
  `runtime/session/hostkit`，把child worker调用、durable record、回读校验、
  parent verification与可选Objective handoff组合为真实portable implementation。
- `WorkerRuntime`和`StateStore`由Host显式注入；concrete worker、scheduler、queue、
  credential、backend与产品policy没有下沉到canonical Runtime。
- HS三条production consumer已固定该pseudo-version；四个通用兼容文件从
  1,562行收缩为79行alias/forwarder与必要Host投影，并增加owner gate防止实现回流。
- 新增无HS、无Runner、无长期`replace`的fixed-version consumer，中文Reference现覆盖
  48个package和10个Developer Preview candidate。
- Scheduler/Resume/long-task orchestration保留给P1-D；本版本仍是private
  Developer Preview candidate，不授权Public/Beta/Stable、tag或发行。

## v0.0.0-20260802080954-21919fd8e06a（P1-B Objective Host Kit）

- 新增Experimental `runtime/objective`最小推荐类型入口和Developer Preview candidate
  `runtime/objective/hostkit`；Host Kit真实组合managed ingress、explicit host dispatch、
  observation normalization和Objective verification。
- 将通用observation normalization从HS迁入canonical `runtime/controlcontract`，并让HS
  production productshell路径通过固定版本使用canonical dispatch；HS旧实现降为薄兼容层。
- 新增无HS、无Runner、无长期`replace`的fixed consumer，覆盖成功、显式确认、取消、
  typed result和verification；中文Reference与Objective Host Kit接入指南同步落地。
- 本版本仍是private Developer Preview candidate，不授权Public/Beta/Stable、tag或发行。

## v0.0.0-20260802072349-5311859981e9（P1-A Model Conversation / Tool Direct Answer）

- 新增`runtime/hostkit.NewChatClient`单轮 Model Conversation 推荐构造；conversation
  backend、provider和历史加载继续由 Host注入；
- 新增显式`ToolDirectAnswer`与统一`ExecutionResult`投影；Host负责业务判断、授权和
  display-safe处理，canonical Host Kit负责不再发起模型合成轮次并完成 portable Run；
- HS Open Tool Loop 的`answer_contract`检测、recovery、persistence和telemetry保持原
  owner，但 production completion已切换 canonical结果策略；
- 固定版本consumer覆盖无HS的 Chat与 Direct Answer；本版本仍是private validation
  pseudo-version，不是tag、semver或Public/Beta/Stable声明。

## Unreleased（M6D技术候选）

- 新增可删除的四module同版file-proxy候选、module zip SHA-256、依赖图、无replace
  consumer和只读cache验证；模拟版本`v0.0.0-m6d.0`不是首次正式版本；
- 固定Go 1.25.12与`govulncheck@v1.6.0`。Go 1.25.5扫描命中的10个标准库可达漏洞已由
  安全patch工具链解除；当前可达漏洞为0；
- 保留extensions中Windows-only `GO-2026-5024`不可达module finding，等待具名
  security approver决定升级或接受残余平台边界；
- 新增无PR依赖的M6D Ubuntu workflow和value-safe候选manifest；不创建tag/release，
  run `30733996721`已在revision `2034622f991e`完成技术复验并上传artifact
  `8829013105`；`pre_beta_technical_candidate_ready=true`、`public_beta_ready=false`。

## v0.0.0-20260802021959-5a41fb0ccb87（M5S / M5T基线）

- 固定根`agentx.Client`自定义`ExecutionAdapter`、Model/Tool Host Kit和Workflow Host Kit
  三条host-provided construction路径；
- 44个production package有中文Reference，8个Developer Preview candidate有可读API
  snapshot、hash、公开类型闭包与darwin/linux签名门禁；
- A股推荐入口的六个evaluator DTO由公开`extensions/astock/contracts`持有，避免公共
  签名依赖Go `internal`包；
- 统一consumer固定root/components/runtime/extensions四module，使用fixture验证三条
  Core路径和A股推荐入口，无HS、Runner、Scene、provider、credential或网络；
- 完全hostless Runtime仍为`not_ready_for_hostless_w2b`；Objective/control大面、Scene、
  正式tag与发行均未进入该基线。

升级和回滚步骤见[版本、升级与回滚](docs/guides/versioning-and-upgrades.md)。
