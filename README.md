# AgentX Go

`agentx-go` 是 AgentX 面向 Go consumer 的独立源码仓库。当前根 module 提供经过
HS/M2 验证的最小执行合同；独立 `components` module拥有 provider-neutral LLM
合同；独立 `runtime` module 已逐步迁入协议、遥测、预算、Workflow portable
implementation owner 和 Run/Open Tool Loop 通用机制。

> M6A Core Pre-Beta Contract and Distribution Preflight Closure已获Owner接受。
> M6B Core中文Developer Portal与API Reference交付闭环已获Owner接受。
> M6C Core Ubuntu Runtime与跨平台分发证据闭环已获Owner接受。
> 当前里程碑：**M6D Core Pre-Beta候选安全与发行工程闭环已完成技术
> checkpoint，等待Owner接受**。
> M5S三条标准construction、44包中文Reference、8个候选API gate与统一fixed consumer
> 已获Owner接受；P1-B现新增第四条Objective Host Kit、46包分类和第9个候选API。
> M5T也已收口四module固定版本、升级/回滚说明与独立消费证据并获接受。
> M6B已把现有中文正文、44个package Reference与8个candidate交付成可构建、导航、
> 搜索的85页Developer Portal Candidate。M6C已经证明Ubuntu真实运行与跨平台分发；
> M6D已用Go 1.25.12、固定`govulncheck`、临时同版四module proxy、只读cache
> consumer和Ubuntu远端run形成可复验技术候选。它不新增Runtime能力，也不是Public、
> Beta、Stable或production-ready发布。

## 当前提供：根合同

- `Client`、`Config`、`New`
- 同步 `Run(ctx, RunRequest)`
- 稳定 `ErrorCode`、typed `Error` 和 `errors.Is/As`
- 六维 `ExecutionProfile`
- 有界、幂等的 `Shutdown(ctx)` 合同
- 供 Runtime/host 实现的窄 `ExecutionAdapter`

安装、四module固定版本和升级边界见[安装与多 Module 引用](docs/guides/installation-and-modules.md)
与[版本、升级与回滚](docs/guides/versioning-and-upgrades.md)。当前Developer Preview
变更摘要见[CHANGELOG](CHANGELOG.md)。

维护、兼容和分发边界见[Developer Preview政策](docs/guides/developer-preview-policy.md)、
[分发Readiness](docs/reference/distribution-readiness.md)、[安全报告](SECURITY.md)和
[支持边界](SUPPORT.md)。参与修改请阅读[CONTRIBUTING](CONTRIBUTING.md)；PR不是必需
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

## 当前提供：Experimental Extensions

- [`extensions/astock`](extensions/astock/API.md)：A股 Manifest、不可变资产、7个
  tool schema、3组 Pack与 evaluator推荐入口；进入 Developer Preview candidate
  focused签名门禁
- [`extensions/astock/hostkit`](extensions/astock/hostkit/API.md)：显式注入 Host
  handler的 intent、readiness与回答格式化
- [`extensions/domainmodule`](extensions/domainmodule/API.md)与
  [`extensions/pack`](extensions/pack/API.md)：portable注册、选择、binding与物化机制
- [`extensions/skills`](extensions/skills/API.md)：Skill数据合同、目录与 immutable
  `fs.FS` loader/cache、activation、requested semantics和资源引用检查；保持
  Experimental，不包含 prompt catalog/filter、安全策略、安装执行或 bundled内容
- [`extensions/productshell`](extensions/productshell/API.md)：输入/preparation、临时
  Workflow planning、typed observation和display-safe Host handoff；具体model/tool
  policy、provider、backend与执行继续由Host显式注入
- extension不安装 provider、credential、Runner、网络或生产 backend

## 当前不提供

- 无需任何 host-provided model/tool adapter和 policy的开箱即用 Runtime
- 官方模型/provider、credential或生产网络接入
- 根 `agentx` Facade 的 Workflow、Objective、Resume 或长任务入口
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
  v0.0.0-20260802080954-21919fd8e06a

github.com/wsnacj/agentx-go/components
  v0.0.0-20260802080954-21919fd8e06a

github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260802080954-21919fd8e06a

github.com/wsnacj/agentx-go/extensions
  v0.0.0-20260802080954-21919fd8e06a
```

它们是不可变 private validation pseudo-version，不是正式发布版本。

## 文档

- [文档入口](docs/README.md)
- [快速开始](docs/quickstart.md)
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
- [`extensions/astock` 中文 API Reference](extensions/astock/API.md)
- [`extensions/astock/contracts` 中文 API Reference](extensions/astock/contracts/API.md)
- [`extensions/astock/hostkit` 中文 API Reference](extensions/astock/hostkit/API.md)
- [`extensions/domainmodule` 中文 API Reference](extensions/domainmodule/API.md)
- [`extensions/skills` 中文 API Reference](extensions/skills/API.md)
- [最小合同示例](examples/contract-basic)
- [自定义 Adapter 示例](examples/custom-adapter)
- [三条标准路径统一 External-style consumer](conformance/consumer)
- [Objective Host Kit external-style consumer](runtime/conformance/objective-hostkit-consumer)
- [Control Contract external-style consumer](runtime/conformance/controlcontract-consumer)
- [Extension external-style consumer](extensions/conformance/astock-contract-consumer)
- [Domain Module external-style consumer](extensions/conformance/domain-module-consumer)
- [A 股组合 external-style consumer](extensions/conformance/astock-consumer)
- [Skills external-style consumer](extensions/conformance/skills-consumer)

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
GOWORK=off GOPROXY=off go -C extensions/conformance/astock-contract-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C extensions/conformance/domain-module-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C extensions/conformance/astock-consumer test ./... -count=1
GOWORK=off GOPROXY=off go -C extensions/conformance/skills-consumer test ./... -count=1
GOWORK=off go run scripts/check_developer_preview_api.go
GOWORK=off go run scripts/check_developer_preview_api.go -check-platforms
GOWORK=off go run scripts/check_docs_links.go
GOWORK=off go run scripts/check_developer_preview_distribution.go
```

根 contract 与 `components/llm` 的 production代码只依赖 Go 标准库；Runtime
只依赖标准库及已批准的 canonical contract/component。当前私有验证阶段不创建
tag，不承诺正式 module版本，也不自动授权 W2-B、更多 components或 Scene迁移。
