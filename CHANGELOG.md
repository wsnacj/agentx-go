# Developer Preview 变更记录

本文件记录可供private consumer复现的Developer Preview checkpoint，不代表正式release、
semver兼容承诺或Public/Beta/Stable声明。

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
