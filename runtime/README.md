# AgentX Go Runtime

本目录是 `agentx-go` 的 Runtime owner module：

```text
github.com/wsnacj/agentx-go/runtime
```

当前只落地首个无外部依赖的叶子 package：

- [`protocol`](./protocol/API.md)：版本化 Runtime wire/schema、normalization
  与 validation。

当前成熟度为 **private validation / Experimental**。本 module 尚未提供根
`agentxruntime.New`、Runner、真实 backend、provider、credential、Scene 或完整
embedded Runtime，不能据此宣称 Runtime 已达到 Public、Beta、Stable 或
production-ready。

依赖方向固定为：

```text
Runtime owner package -> Go 标准库或已批准的 agentx-go contract/component
```

production package 不得 import `hs/...`、`scene/...` 或机器本地路径。新增
package 必须有独立 owner/consumer review、中文 API 文档、focused/race/module
验证和 source-authority cutover，不创建杂项 `common` package。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```

`conformance/` 下的 external-style consumer 是独立 nested module，必须单独
验证；根 module 的 `go test ./...` 不会自动跨越 nested module。
