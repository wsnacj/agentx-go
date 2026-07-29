# AgentX Go

`agentx-go` 是 AgentX 面向 Go consumer 的独立源码仓库。当前 W1 只落地经过
HS/M2 验证的最小执行合同，为后续 Runtime 内核和通用组件迁移建立稳定的依赖方向。

> 当前成熟度：**private validation / contract-only**。这不是 Public、Beta 或
> Stable 发布，也不是完整 embedded Agent SDK。

## 当前提供

- `Client`、`Config`、`New`
- 同步 `Run(ctx, RunRequest)`
- 稳定 `ErrorCode`、typed `Error` 和 `errors.Is/As`
- 六维 `ExecutionProfile`
- 有界、幂等的 `Shutdown(ctx)` 合同
- 供 Runtime/host 实现的窄 `ExecutionAdapter`

## 当前不提供

- 官方 Runtime 构造入口或模型/provider 接入
- Workflow、Objective Runtime Loop、Resume、durable lifecycle
- progress stream、HTTP API、Scene registry
- credential、真实网络 backend 或生产副作用

普通使用者要获得开箱即用的 Runtime 构造能力，仍需等待后续
`agentxruntime.New(ctx, Config)` 工作包。W1 的 `ExecutionAdapter` 面向扩展作者
和集成验证，不等于要求所有业务调用方自行实现 Runtime。

## 文档

- [文档入口](docs/README.md)
- [快速开始](docs/quickstart.md)
- [执行模型](docs/concepts/execution-model.md)
- [Go API Reference](docs/reference/agentx.md)
- [自定义 Adapter](docs/guides/custom-adapter.md)
- [生命周期与错误处理](docs/guides/lifecycle-and-errors.md)
- [成熟度与兼容边界](docs/maturity.md)
- [最小合同示例](examples/contract-basic)
- [自定义 Adapter 示例](examples/custom-adapter)
- [External-style consumer](conformance/consumer)

## 本地验证

```bash
GOWORK=off go test ./... -count=1
GOWORK=off go test -race ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go mod tidy -diff
```

根 contract 的 production 代码只依赖 Go 标准库。当前私有验证阶段不创建 tag，
不承诺正式 module 版本，也不授权 Runtime、components 或 Scene 迁移。
