# AgentX Go

AgentX Go 是面向 Go 项目的可组合 Agent Runtime。它提供稳定、窄小的执行合同，
并把模型、工具、工作流、长任务、浏览器、文档和领域能力拆成显式选择的模块。

> 当前成熟度：`v0.2.2 Developer Preview`。已有真实实现、外部消费示例和中文 API Reference，
> 只对文档明确列出的9包核心候选面实施受治理变更。Experimental package仍可能调整；
> 本版本不是Beta、Stable、production-ready或生产SLA。

最低工具链为 Go 1.25.0；发布验证固定使用 Go 1.25.12。

## 为什么使用 AgentX

- 一个统一的 `Run`、typed error、context cancellation 和有界 `Shutdown` 合同；
- 模型、工具、存储、调度器和副作用由 Host 显式注入；
- Open Tool Loop、Workflow、Objective 和长任务使用独立、可测试的组合入口；
- Browser、Document 和 Scenes 是可选模块，不强迫轻量项目引入全部能力；
- 所有外部可导入 package 都有中文 `API.md`，推荐入口还受到签名漂移检查；
- 默认示例不读取凭据、不访问网络、不启动进程，也不写入文件。

## 五分钟开始

安装最小 Chat 路径：

```bash
go get github.com/wsnacj/agentx-go@v0.2.2
go get github.com/wsnacj/agentx-go/components@v0.2.2
go get github.com/wsnacj/agentx-go/runtime@v0.2.2
```

调用方显式提供模型函数：

```go
package main

import (
    "context"
    "fmt"

    agentx "github.com/wsnacj/agentx-go"
    llm "github.com/wsnacj/agentx-go/components/llm"
    "github.com/wsnacj/agentx-go/runtime/hostkit"
)

func main() {
    client, err := hostkit.NewChatClient(hostkit.ChatClientConfig{
        Model: "my-model",
        RequestModel: func(_ context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
            // 在这里调用你选择的模型 provider。
            return llm.ChatResponse{Content: "hello from AgentX"}, nil
        },
    })
    if err != nil {
        panic(err)
    }
    defer client.Shutdown(context.Background())

    result, err := client.Run(context.Background(), agentx.RunRequest{
        Input: "hello",
        SessionID: "quickstart",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Reply)
}
```

完整说明见[快速开始](docs/quickstart.md)和[Chat 示例](examples/chat)。

## 七类能力

| 能力 | 推荐入口 | Host 需要提供 |
| --- | --- | --- |
| Model Conversation / A0 | `runtime/hostkit.NewChatClient` | 模型请求函数 |
| Open Tool Loop | `runtime/hostkit.NewModelToolClient` | 模型与工具执行 |
| Tool Direct Answer | `runtime/hostkit.ToolDirectAnswer` | 结果确认策略 |
| Workflow | `runtime/workflow/hostkit` | validator、mapper、executor、identity、clock |
| Objective Runtime Loop | `runtime/objective/hostkit` | strategy、policy、approval、handler |
| 长任务、Session、Subagent、Resume | `runtime/session/hostkit`、`runtime/scheduler` | worker、state、queue、持久化 |
| Deterministic Scene | `extensions/domainkit`、`scenes` | 领域 handler 和真实 backend |

这些是执行路径和组合能力，不是一个可任意切换的 mode 枚举。根 `Client` 当前只接受
`DefaultExecutionProfile`；其它能力由对应 Host Kit 显式构造。

## 选择模块

| Module | 用途 | 适合 |
| --- | --- | --- |
| `github.com/wsnacj/agentx-go` | 根 Client、Run、错误和生命周期 | 所有接入 |
| `.../components` | LLM 与 Tool 的 provider-neutral 合同 | Provider、Runtime、工具扩展 |
| `.../runtime` | Tool Loop、Workflow、Objective、Session、Scheduler、RunStore | Agent Runtime |
| `.../extensions` | Pack、Skills、Domain Kit、Product Shell | 可组合扩展 |
| `.../providers` | OpenAI-compatible、Anthropic、Codex、Ark、Gemini | 显式选择的模型接入 |
| `.../tools` | Catalog、Web、LLM Task、Process、Filesystem 等 | 通用工具能力 |
| `.../browser` | Browser Runtime、browserd Host、Browser Tools | 浏览器自动化 |
| `.../document` | OCR、PDF、Pipeline、Document Tools | 文档处理 |
| `.../scenes` | 可移植领域 Kit | 领域能力组合 |

`browser`、`document` 和 `scenes` 是可选 batteries。`scenes` 的 module graph 包含
Browser 与 Document，但轻量 package 的实际 import graph 仍保持按需加载；在对二进制大小
或依赖下载敏感的项目中，应先使用 `go list -deps` 测量所选 package。

模块安装、私有仓库认证和版本组合见
[安装与多 Module 引用](docs/guides/installation-and-modules.md)。

## 两条标准接入路径

### Host Kit

普通项目优先使用 Host Kit。它组合 AgentX 已有实现，调用方只提供模型、工具、策略或
backend port：

- [Model / Tool Host Kit](docs/guides/model-tool-hostkit.md)
- [Workflow Host Kit](docs/guides/workflow-hostkit.md)
- [Objective Host Kit](docs/guides/objective-hostkit.md)
- [Session / Subagent Host Kit](docs/guides/session-subagent-hostkit.md)

### 自定义 ExecutionAdapter

已有完整 Runtime 或需要接入自有执行底座的项目，可以实现最窄
`ExecutionAdapter`。详见[自定义 Adapter](docs/guides/custom-adapter.md)。

## 显式 Host 边界

AgentX 提供 portable mechanism，不替调用方决定生产策略：

- credential、provider 选择和网络出口由 Host 注入；
- authorization、approval、sandbox 和工具 allowlist 由 Host 管理；
- durable store、scheduler backend 和资源预算由 Host 管理；
- Scene 的真实数据源、客户 schema 和业务规则由领域 Host 管理；
- 本地进程、浏览器、Python、OCR 等副作用 adapter 必须显式构造和启用。

`runtime/hosthttp/hostserver.DefaultConfig` 不读取环境变量。需要采用约定部署环境变量时，
Host 必须显式调用 `ConfigFromEnv`。

## 示例与验证

[examples](examples/README.md) 提供七类核心能力、Browser/Document 重型扩展和 Reference Host
的可运行示例。它们使用
fixture、纯函数工具或内存 backend，适合作为新项目模板。

[conformance](conformance/consumer) 和各 module 下的 `conformance/*-consumer` 用于验证
固定版本、跨 module 类型身份、错误、取消、并发和依赖方向。需要真实凭据的 provider
验证单独放在 [opt-in live smoke](conformance/live/README.md)，不会被默认测试执行。

## 文档

- [文档首页](docs/README.md)
- [能力与接入路径](docs/guides/capability-map.md)
- [执行模型](docs/concepts/execution-model.md)
- [生命周期与错误](docs/guides/lifecycle-and-errors.md)
- [Package API 与成熟度](docs/reference/package-maturity.md)
- [Developer Preview 政策](docs/guides/developer-preview-policy.md)
- [版本、升级与回滚](docs/guides/versioning-and-upgrades.md)
- [安全报告](SECURITY.md)
- [支持边界](SUPPORT.md)

## 本地开发

每个 library module 都应独立执行：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy -diff
```

中文文档站验证：

```bash
npm ci
npm run docs:check
```

参与开发前请阅读[贡献指南](CONTRIBUTING.md)。当前没有稳定兼容承诺；成熟度和公开
发行状态以[成熟度说明](docs/maturity.md)与正式 Release 为准。

## License

AgentX Go 以 [Apache License 2.0](LICENSE) 提供。直接依赖及上游归属摘要见
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)；各第三方组件仍遵循其自身许可证。
