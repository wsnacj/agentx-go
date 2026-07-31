# AgentX Go

`agentx-go` 是 AgentX 面向 Go consumer 的独立源码仓库。当前根 module 提供经过
HS/M2 验证的最小执行合同；独立 `components` module拥有 provider-neutral LLM
合同；独立 `runtime` module 已逐步迁入协议、遥测、预算和 Workflow portable
implementation owner。

> 当前成熟度：**private validation / Experimental Runtime**。这不是 Public、
> Beta 或 Stable 发布，也还不是开箱即用的完整 embedded Agent SDK。

## 当前提供：根合同

- `Client`、`Config`、`New`
- 同步 `Run(ctx, RunRequest)`
- 稳定 `ErrorCode`、typed `Error` 和 `errors.Is/As`
- 六维 `ExecutionProfile`
- 有界、幂等的 `Shutdown(ctx)` 合同
- 供 Runtime/host 实现的窄 `ExecutionAdapter`

## 当前提供：Experimental 组件

- [`components/llm`](components/llm/API.md)：provider-neutral 的 LLM
  request/response、tool、stream、multimodal 与 usage 合同
- 独立 module：`github.com/wsnacj/agentx-go/components`
- 生产代码只依赖 Go 标准库；不提供 provider、credential、网络客户端或
  AgentX Runtime

## 当前提供：Experimental Runtime

- [`runtime`](runtime/README.md) 独立 module：
  `github.com/wsnacj/agentx-go/runtime`
- protocol、telemetry、budget、prompt context、tool error与 media artifact
  owner
- Workflow Spec、schema、validation、lowering、binding/state、transition、
  journal、node execution、orchestration与 composition owner
- 每个已迁 production package均提供中文 `API.md`、contract/external tests和
  import-direction gate
- Runtime production代码不依赖 HS、Scene、具体 provider或 backend

## 当前不提供

- 官方 Runtime 构造入口或模型/provider 接入
- 根 `agentx` Facade 的 Workflow、Objective、Resume 或长任务入口
- concrete Workflow validation/mapping policy、executor和 RunStore backend
- progress stream、HTTP API、Scene registry
- credential、真实网络 backend 或生产副作用

普通使用者要获得开箱即用的 Runtime 构造能力，仍需等待后续
`agentxruntime.New(ctx, Config)` 工作包。W1 的 `ExecutionAdapter` 面向扩展作者
和集成验证，不等于要求所有业务调用方自行实现 Runtime。

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
  v0.0.0-20260729101644-c7c26d427ac2

github.com/wsnacj/agentx-go/components
  v0.0.0-20260729125257-bb6949793309

github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260731070530-d41d75fadece
```

它们是不可变 private validation pseudo-version，不是正式发布版本。

## 文档

- [文档入口](docs/README.md)
- [快速开始](docs/quickstart.md)
- [执行模型](docs/concepts/execution-model.md)
- [Go API Reference](docs/reference/agentx.md)
- [自定义 Adapter](docs/guides/custom-adapter.md)
- [生命周期与错误处理](docs/guides/lifecycle-and-errors.md)
- [成熟度与兼容边界](docs/maturity.md)
- [`components/llm` 中文 API Reference](components/llm/API.md)
- [`runtime` 中文 package 导航](runtime/README.md)
- [`runtime/workflow/composition` 中文 API Reference](runtime/workflow/composition/API.md)
- [最小合同示例](examples/contract-basic)
- [自定义 Adapter 示例](examples/custom-adapter)
- [External-style consumer](conformance/consumer)

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
```

根 contract 与 `components/llm` 的 production代码只依赖 Go 标准库；Runtime
只依赖标准库及已批准的 canonical contract/component。当前私有验证阶段不创建
tag，不承诺正式 module版本，也不自动授权 W2-B、更多 components或 Scene迁移。
