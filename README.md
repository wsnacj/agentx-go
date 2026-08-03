# AgentX Go

`agentx-go` 是 AgentX 面向 Go consumer 的独立源码仓库。当前根 module 提供经过
HS/M2 验证的最小执行合同；独立 `components` module拥有 provider-neutral LLM
合同；独立 `runtime` module 已迁入协议、遥测、预算、Workflow、Objective、长任务与
Run/Open Tool Loop 的 portable implementation owner；`providers`、`tools`、`browser`、
`document`和`scenes`提供显式选择的可选能力。

> M6A Core Pre-Beta Contract and Distribution Preflight Closure已获Owner接受。
> M6B Core中文Developer Portal与API Reference交付闭环已获Owner接受。
> M6C Core Ubuntu Runtime与跨平台分发证据闭环已获Owner接受。
> 当前里程碑：**M7A P1 Core七类能力、P2 Providers/通用Tools、P3 Browser/Document、
> P4首批Portable Scenes与P5 Core Developer Preview产品化均已完成私有候选闭环**。
> 当前九个module共有
> 123份中文`API.md`，覆盖全部120个可外部import的production package；其中
> root/components/runtime/extensions/scenes的77个候选范围
> package和14个Developer Preview candidate进入focused API gate。
> P4 Portfolio Checkpoint已经停止按Scene逐个扩张；P5-C closure没有发现需要补迁的
> 公共能力，因此未增加新cohort。M6D Core Foundation发行门禁继续独立fail closed。
> M6C已经为历史Core四module快照形成Ubuntu真实运行与跨平台分发证据；
> M6D历史四module证据继续保留；当前Pre-Beta本地、distribution与Ubuntu入口已升级到
> 九module release train。九module本地技术候选已经通过，当前仍等待九module Ubuntu复验
> 及具名License、安全、兼容范围和发行决定。它不新增Runtime能力，也不是Public、Beta、
> Stable或production-ready发布。

P5-C最终验证覆盖九module normal/race/vet/tidy/list、33个无`replace` fixed-version
consumer、仓库外clean-room consumer、九module zip/Sum/Origin、反向依赖和中文Portal。
它证明private Developer Preview Candidate可复验，不构成正式发行或兼容性承诺。

## 当前提供：根合同

- `Client`、`Config`、`New`
- 同步 `Run(ctx, RunRequest)`
- 稳定 `ErrorCode`、typed `Error` 和 `errors.Is/As`
- 六维 `ExecutionProfile`
- 有界、幂等的 `Shutdown(ctx)` 合同
- 供 Runtime/host 实现的窄 `ExecutionAdapter`

安装、九module固定版本和升级边界见[安装与多 Module 引用](docs/guides/installation-and-modules.md)
与[版本、升级与回滚](docs/guides/versioning-and-upgrades.md)。当前Developer Preview
变更摘要见[CHANGELOG](CHANGELOG.md)。

维护、兼容和分发边界见[Developer Preview政策](docs/guides/developer-preview-policy.md)、
[分发Readiness](docs/reference/distribution-readiness.md)、[安全报告](SECURITY.md)和
[Pre-Beta准入合同](docs/reference/pre-beta-admission.md)、[支持边界](SUPPORT.md)。参与修改请阅读[CONTRIBUTING](CONTRIBUTING.md)；PR不是必需
流程，本地gate与commit-range人工审阅是一等路径。

中文Developer Portal本地构建：

```bash
npm ci
npm run docs:check
```

站点只投影现有Markdown/API事实源；生成目录和静态产物不提交。工程说明见
[Portal README](portal/README.md)。

## 当前提供：Experimental 组件

- [`components/llm`](components/llm/API.md)：provider-neutral 的 LLM
  request/response、tool、stream、multimodal 与 usage 合同
- [`components/tool`](components/tool/API.md)：复用同一份 LLM wire type identity 的
  provider-neutral tool catalog/handler/executor合同
- 独立 module：`github.com/wsnacj/agentx-go/components`
- 生产代码只依赖 Go 标准库；不提供 provider、credential、网络客户端或
  AgentX Runtime

## 当前提供：Experimental Runtime

- [`runtime`](runtime/README.md) 独立 module：
  `github.com/wsnacj/agentx-go/runtime`
- protocol、telemetry、budget、prompt context、tool error、media artifact与
  Runtime construction lifecycle owner
- [`runtime/execution`](runtime/execution/API.md) 的根 Client→Host Run分派、
  adapter result组装、Shutdown转发与 error classification委托；具体 engine
  request/result投影仍由 host拥有
- [`runtime/hostkit`](runtime/hostkit/API.md) 的 portable model/tool round
  adapter、per-run assembly、outcome/result projection、execution adapter与根
  Client组合；普通新项目通过 `NewModelToolClient`显式注入 model/tool函数，无需
  手写 Factory、RoundExecutor或依赖 HS Runner
- [`runtime/session/hostkit`](runtime/session/hostkit/API.md) 的Host-owned child worker
  invoke、record/readback、parent verification与可选Objective handoff，以及bounded
  Scheduler/Resume组合；具体process、系统scheduler、authorization和backend由Host注入
- [`runtime/scheduler`](runtime/scheduler/API.md) 的portable queue、dispatcher、lease
  heartbeat、terminal coordination与metrics；生产durable backend和产品policy由Host注入
- [`runtime/session/resume`](runtime/session/resume/API.md) 的continuation readback、Host
  wake dispatch与bounded service实现；普通调用方优先使用`session/hostkit.NewResumeRuntime`
- [`runtime/toolloop`](runtime/toolloop/API.md) 的确定性多轮驱动、round 结果
  收口/continuation state 更新、request→observe→action phase编排、循环/重放
  检测与连续工具失败熔断，以及把 driver/coordinator/termination/final state
  组合成单次 Run 的 Host-backed `Assembly`；具体 model/tool执行、持久化和产品
  策略仍由 host注入
- Workflow Spec、schema、validation、lowering、binding/state、transition、
  journal、node execution、orchestration与 composition owner；
  [`runtime/workflow/hostkit`](runtime/workflow/hostkit/API.md)把这些真实 owner
  组合为只需显式注入 validator、mapper、executor、identity、clock和可选
  durable port的标准入口
- 每个已迁 production package均提供中文 `API.md`、contract/external tests和
  import-direction gate
- Runtime production代码不依赖 HS、Scene、具体 provider或 backend

## 当前提供：Experimental Providers

- [`providers`](providers/API.md) 独立可选 module：
  `github.com/wsnacj/agentx-go/providers`
- [`providers/openaicompat`](providers/openaicompat/API.md) 提供真实 OpenAI-compatible
  chat、vision、embedding、bot和SSE stream实现
- [`providers/anthropic`](providers/anthropic/API.md) 提供真实Anthropic Messages实现；
  [`providers/codex`](providers/codex/API.md) 提供真实Codex Responses/SSE实现
- `transport`、`fault`、`retry`和`usage`提供provider-neutral机制；固定版本
  external consumer不依赖HS、Runner、Scene、`replace`或真实网络
- endpoint、credential、proxy、配额、审计、模型选择和生产出网授权仍由Host显式注入；
  默认构造不读环境变量、文件或credential store

## 当前提供：Experimental Tools

- [`tools`](tools/API.md) 独立可选 module：`github.com/wsnacj/agentx-go/tools`
- 线程安全Registry、稳定definition排序/version/reset、保守名称修复和typed name error
  已形成真实implementation owner
- [`tools/diffs`](tools/diffs/API.md)提供首个真实通用tool：只处理调用方显式传入的
  before/after文本，不读取文件、Git、网络或工作区
- [`tools/message`](tools/message/API.md)、[`tools/httprequest`](tools/httprequest/API.md)、
  [`tools/filesystem`](tools/filesystem/API.md)、[`tools/memory`](tools/memory/API.md)与
  [`tools/scheduler`](tools/scheduler/API.md)提供真实portable coordination；固定consumer
  以fake/in-memory ports执行10个入口
- authorization、approval、sandbox、credential、filesystem/process/network/store/
  scheduler backend和产品allowlist仍由Host拥有

## 当前提供：Experimental Browser

- [`browser`](browser/API.md)独立可选module：`github.com/wsnacj/agentx-go/browser`
- [`browser/runtime`](browser/runtime/API.md)拥有30,906行provider-neutral Browser
  session/action/route/snapshot/capability/result与state/recovery/watch implementation；
  HS production consumer已使用固定版本
- [`browser/host/browserd`](browser/host/browserd/API.md)提供显式Plan/StatusProbe驱动的
  browserd process manager、内置Node资产和Playwright bootstrap/cache；构造阶段无副作用
- [`browser/tools`](browser/tools/API.md)与统一fixed consumer已完成P3闭环；默认仍不启动
  browserd、不访问网络、不读取credential，真实代理、登录态和企业网络由Host注入

## 当前提供：Experimental Document

- [`document`](document/API.md)独立可选module：`github.com/wsnacj/agentx-go/document`
- [`document/ocr`](document/ocr/API.md)、[`document/pdf`](document/pdf/API.md)与
  [`document/pipeline`](document/pipeline/API.md)提供OCR、PDF和结构化文档处理的真实实现
- [`document/tools`](document/tools/API.md)提供显式Host能力边界内的推荐Document tools；
  fixed consumer与HS production cutover已经完成
- provider credential、文件权限、外部Python环境、生产存储和资源预算仍由Host治理；
  默认接入不隐式读取credential或启动外部服务

## 当前提供：Experimental Extensions 与 Portable Scenes

- [`scenes/astock`](scenes/astock/API.md)：A股 Manifest、不可变资产、7个
  tool schema、3组 Pack与 evaluator推荐入口；进入 Developer Preview candidate
  focused签名门禁
- [`scenes/astock/hostkit`](scenes/astock/hostkit/API.md)：显式注入 Host
  handler的 intent、readiness与回答格式化
- [`scenes/publicnews`](scenes/publicnews/API.md)与
  [`scenes/companyresearch`](scenes/companyresearch/API.md)：公开新闻与公司研究的只读
  contract、Pack、evidence/quality机制及显式Host ports
- [`scenes/docparse`](scenes/docparse/API.md)：文档profile/planner/fusion/understanding、
  Pack、质量证据与无文件/无provider Host Kit；真实OCR/PDF、文件访问和私有schema由Host注入
- [`scenes/browserops`](scenes/browserops/API.md)：Browser Ops Pack、证据投影、六类确定性
  evaluator与显式[Host Kit](scenes/browserops/hostkit/API.md)；真实浏览器、profile/login、
  credential、审批、文件/artifact和站点副作用策略由Host注入
- [`scenes/publictransport`](scenes/publictransport/API.md)：公共交通只读合同、协调、证据
  evaluator与Pack；真实票务provider、endpoint、限流和合规策略由Host注入
- [`scenes/publicsource`](scenes/publicsource/API.md)与
  [`scenes/wechatarticle`](scenes/wechatarticle/API.md)：公开来源与公众号文章的typed合同、
  evidence、evaluator和Host Client协调
- [`scenes/globalstock`](scenes/globalstock/API.md)与
  [`scenes/finance`](scenes/finance/API.md)：港/美股与财报的只读合同、Pack、Workflow、
  evidence/readiness和显式Host Kit；行情、财报provider及真实网络继续留在Host
- [`extensions/domainkit`](extensions/domainkit/API.md)：无模型module/tool dispatch、typed
  error与deterministic output digest；provider与真实副作用必须由Host handler注入
- [`extensions/domainmodule`](extensions/domainmodule/API.md)与
  [`extensions/pack`](extensions/pack/API.md)：portable注册、选择、binding与物化机制
- [`extensions/skills`](extensions/skills/API.md)：Skill数据合同、目录与 immutable
  `fs.FS` loader/cache、activation、requested semantics和资源引用检查；保持
  Experimental，不包含 prompt catalog/filter、安全策略、安装执行或 bundled内容
- [`extensions/productshell`](extensions/productshell/API.md)：输入/preparation、临时
  Workflow planning、typed observation和display-safe Host handoff；具体model/tool
  policy、provider、backend与执行继续由Host显式注入
- extension与portable Scene不安装 provider、credential、Runner、网络或生产 backend

## 当前不提供

- 无需任何 host-provided model/tool adapter和 policy的开箱即用 Runtime
- 默认模型目录、credential发现/轮换、生产网络授权或开箱即用provider配置
- 根 `agentx` Facade 的 Workflow、Objective、Resume 或长任务入口；当前Resume位于
  Experimental Runtime Host Kit
- concrete Workflow validation/mapping policy、executor和 RunStore backend
- progress stream、HTTP API、Scene registry
- credential、真实网络 backend 或生产副作用

[`runtime/construction`](runtime/construction/API.md) 已提供基于窄 `Host`
port 的 Experimental 构造生命周期；[`runtime/hostkit`](runtime/hostkit/API.md)
又已提供无 HS Runner 的真实执行组合和低样板 `NewModelToolClient`。普通使用者
仍需显式提供 concrete model/tool adapter；无需这些 host能力的完整 embedded
Runtime结论保持 `not_ready_for_hostless_w2b`。根合同的
`ExecutionAdapter` 面向扩展作者和集成验证，不等于要求所有业务调用方自行实现
Runtime。

## Private validation 访问

当前仓库是 private。consumer 环境需要：

```bash
export GOPRIVATE=github.com/wsnacj/agentx-go
export GONOSUMDB=github.com/wsnacj/agentx-go
export GOPROXY=direct
export GOWORK=off
```

Git 还必须能够通过 HTTPS token 或 SSH URL rewrite 访问该私有仓库。凭据和 URL
rewrite 属于开发/CI 环境配置，不写入源码、`go.mod`、示例或日志。当前固定验证
版本为：

```text
github.com/wsnacj/agentx-go
  v0.0.0-20260802113655-f41de95ec5be

github.com/wsnacj/agentx-go/components
  v0.0.0-20260802130858-34ec103e09d9

github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260802113655-f41de95ec5be

github.com/wsnacj/agentx-go/extensions
  v0.0.0-20260802113655-f41de95ec5be

github.com/wsnacj/agentx-go/providers
  v0.0.0-20260802124746-c7f90139a1cc

github.com/wsnacj/agentx-go/tools
  v0.0.0-20260802165151-c51d7391dbb4

github.com/wsnacj/agentx-go/browser
  v0.0.0-20260802183055-f15e2f99ed1a

github.com/wsnacj/agentx-go/document
  v0.0.0-20260802203835-ec74047cbe60

github.com/wsnacj/agentx-go/scenes
  v0.0.0-20260803022834-4043fbe78ff3
```

它们是不可变 private validation pseudo-version，不是正式发布版本。

## 文档

- [文档入口](docs/README.md)
- [快速开始](docs/quickstart.md)
- [七类能力与标准接入路径](docs/guides/capability-map.md)
- [安装与多 Module 引用](docs/guides/installation-and-modules.md)
- [执行模型](docs/concepts/execution-model.md)
- [Go API Reference](docs/reference/agentx.md)
- [自定义 Adapter](docs/guides/custom-adapter.md)
- [Host Kit + Model/Tool Adapter](docs/guides/model-tool-hostkit.md)
- [Workflow Host Kit](docs/guides/workflow-hostkit.md)
- [Objective Host Kit](docs/guides/objective-hostkit.md)
- [A 股 Portable Domain Extension](docs/guides/astock-extension.md)
- [Portable Skills 接入](docs/guides/portable-skills.md)
- [生命周期与错误处理](docs/guides/lifecycle-and-errors.md)
- [Package API 索引与成熟度矩阵](docs/reference/package-maturity.md)
- [HS 迁移说明](docs/guides/hs-migration.md)
- [成熟度与兼容边界](docs/maturity.md)
- [`components/llm` 中文 API Reference](components/llm/API.md)
- [`components/tool` 中文 API Reference](components/tool/API.md)
- [`providers` 中文 API Reference](providers/API.md)
- [`tools` 中文 API Reference](tools/API.md)
- [`tools/diffs` 中文 API Reference](tools/diffs/API.md)
- [`browser` 中文 API 总览](browser/API.md)
- [`browser/runtime` 中文 API Reference](browser/runtime/API.md)
- [`browser/host/browserd` 中文 API Reference](browser/host/browserd/API.md)
- [`runtime` 中文 package 导航](runtime/README.md)
- [`runtime/construction` 中文 API Reference](runtime/construction/API.md)
- [`runtime/controlcontract` 中文 API Reference](runtime/controlcontract/API.md)
- [`runtime/execution` 中文 API Reference](runtime/execution/API.md)
- [`runtime/hostkit` 中文 API Reference](runtime/hostkit/API.md)
- [`runtime/toolloop` 中文 API Reference](runtime/toolloop/API.md)
- [`runtime/workflow/composition` 中文 API Reference](runtime/workflow/composition/API.md)
- [`runtime/workflow/hostkit` 中文 API Reference](runtime/workflow/hostkit/API.md)
- [`runtime/objective` 中文 API Reference](runtime/objective/API.md)
- [`runtime/objective/hostkit` 中文 API Reference](runtime/objective/hostkit/API.md)
- [`runtime/assetfs` 中文 API Reference](runtime/assetfs/API.md)
- [`scenes/astock` 中文 API Reference](scenes/astock/API.md)
- [`scenes/astock/contracts` 中文 API Reference](scenes/astock/contracts/API.md)
- [`scenes/astock/hostkit` 中文 API Reference](scenes/astock/hostkit/API.md)
- [`scenes/publicnews` 中文 API Reference](scenes/publicnews/API.md)
- [`scenes/publicnews/hostkit` 中文 API Reference](scenes/publicnews/hostkit/API.md)
- [`scenes/companyresearch` 中文 API Reference](scenes/companyresearch/API.md)
- [`scenes/companyresearch/hostkit` 中文 API Reference](scenes/companyresearch/hostkit/API.md)
- [`scenes/docparse` 中文 API Reference](scenes/docparse/API.md)
- [`scenes/docparse/hostkit` 中文 API Reference](scenes/docparse/hostkit/API.md)
- [`extensions/domainkit` 中文 API Reference](extensions/domainkit/API.md)
- [`extensions/domainmodule` 中文 API Reference](extensions/domainmodule/API.md)
- [`extensions/skills` 中文 API Reference](extensions/skills/API.md)
- [最小合同示例](examples/contract-basic)
- [自定义 Adapter 示例](examples/custom-adapter)
- [示例与可运行消费证据](examples/README.md)
- [三条标准路径统一 External-style consumer](conformance/consumer)
- [Objective Host Kit external-style consumer](runtime/conformance/objective-hostkit-consumer)
- [Session/Subagent Host Kit external-style consumer](runtime/conformance/session-hostkit-consumer)
- [Control Contract external-style consumer](runtime/conformance/controlcontract-consumer)
- [Domain Module external-style consumer](extensions/conformance/domain-module-consumer)
- [A 股组合 external-style consumer](scenes/conformance/astock-consumer)
- [公开新闻/公司研究 external-style consumer](scenes/conformance/research-consumer)
- [文档解析 external-style consumer](scenes/conformance/docparse-consumer)
- [Skills external-style consumer](extensions/conformance/skills-consumer)
- [Tool catalog/diffs external-style consumer](tools/conformance/catalog-diffs-consumer)

## 本地验证

```bash
GOWORK=off go test ./... -count=1
GOWORK=off go test -race ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go mod tidy -diff
GOWORK=off go -C components test ./... -count=1
GOWORK=off go -C components test -race ./... -count=1
GOWORK=off go -C components vet ./...
GOWORK=off go -C components mod tidy -diff
GOWORK=off GOPROXY=off go -C conformance/consumer test ./... -count=1
GOWORK=off go -C runtime test ./... -count=1
GOWORK=off go -C runtime test -race ./... -count=1
GOWORK=off go -C runtime vet ./...
GOWORK=off go -C runtime mod tidy -diff
GOWORK=off GOPROXY=off go -C runtime/conformance/protocol-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C runtime/conformance/construction-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C runtime/conformance/toolloop-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C runtime/conformance/hostkit-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C runtime/conformance/controlcontract-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C runtime/conformance/workflow-hostkit-consumer test ./... -count=1
GOWORK=off go -C extensions test ./... -count=1
GOWORK=off go -C extensions test -race ./... -count=1
GOWORK=off go -C extensions vet ./...
GOWORK=off go -C extensions mod tidy -diff
GOWORK=off GOPROXY=off go -C extensions/conformance/domain-module-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C scenes/conformance/astock-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C extensions/conformance/skills-consumer test ./... -count=1
GOWORK=off go -C scenes test ./... -count=1
GOWORK=off go -C scenes test -race ./... -count=1
GOWORK=off go -C scenes vet ./...
GOWORK=off go -C scenes mod tidy -diff
GOWORK=off go run scripts/check_developer_preview_api.go
GOWORK=off go run scripts/check_developer_preview_api.go -check-platforms
GOWORK=off go run scripts/check_package_api_docs.go
GOWORK=off go run scripts/check_docs_links.go
GOWORK=off go run scripts/check_developer_preview_distribution.go
```

根 contract 与 `components/llm` 的 production代码只依赖 Go 标准库；Runtime
只依赖标准库及已批准的 canonical contract/component。当前私有验证阶段不创建
tag，不承诺正式 module版本，也不自动授权 W2-B、更多 components或 Scene迁移。
