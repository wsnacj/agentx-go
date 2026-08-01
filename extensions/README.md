# AgentX Go Extensions

本目录是 AgentX 可选扩展的共享 private-preview module：

```text
github.com/wsnacj/agentx-go/extensions
```

首批只包含 [`astock/contracts`](./astock/contracts/API.md)：A 股领域 portable
DTO、JSON normalization、status和 assessment mechanism。它不包含行情 provider、
livekit、pack/workflow、工具执行、credential、缓存或真实网络。

依赖方向固定为：

```text
extensions -> contract/runtime/components
contract/runtime/components -X-> extensions
```

当前 package均为 Experimental extension。共享 module不表示每个 Scene都获得独立
发行资格，也不构成 Public/Beta/Stable、正式 tag或 semver承诺。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```
