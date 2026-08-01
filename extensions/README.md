# AgentX Go Extensions

本目录是 AgentX 可选扩展的共享 private-preview module：

```text
github.com/wsnacj/agentx-go/extensions
```

当前包含：

- [`astock`](./astock/API.md)：M5D active的 A股 portable Domain Extension推荐
  入口，组合 Manifest、immutable assets、tool schema、三组 Pack Definition与
  确定性 evaluator；
- [`astock/contracts`](./astock/contracts/API.md)：A 股领域 portable DTO、JSON
  normalization、status和 assessment mechanism；
- [`astock/hostkit`](./astock/hostkit/API.md)：Host显式注入 handler的 A股 intent、
  investigation、readiness与回答格式化机制；
- [`domainmodule`](./domainmodule/API.md)：编译期 Domain Module的portable
  manifest、config、diagnostics与顺序注册编排。
- [`pack`](./pack/API.md)：Domain Pack定义、显式校验、注册、Workflow选择/物化、
  路由选择与 Binding机制。

它不包含行情 provider、livekit、具体 Workflow policy、工具执行、`pack/runtime`
memory/eval backend、credential、缓存或真实网络。

依赖方向固定为：

```text
extensions -> contract/runtime/components
contract/runtime/components -X-> extensions
```

M5C Portable Pack Core checkpoint已经 Owner接受；M5D现以 `extensions/astock`作为
唯一 active的领域产品切片。`extensions/astock`是 Developer Preview candidate，
`astock/hostkit`与其它 extension owner仍为 Experimental，三组 Pack implementation
位于 `astock/internal`且不允许外部直接依赖。共享 module不表示其它 Scene获得独立
发行资格，也不构成 Public/Beta/Stable、正式 tag或 semver承诺。

固定版本、无 HS、无长期 `replace` 的组合验证位于：

- [`conformance/astock-contract-consumer`](conformance/astock-contract-consumer)；
- [`conformance/domain-module-consumer`](conformance/domain-module-consumer)；
- [`conformance/pack-consumer`](conformance/pack-consumer)。

前三个 consumer分别证明 contracts、Domain Module和 Pack机制可独立消费；M5D仍需
新增一个组合真实 A股 Manifest、Pack和 fixture Host handler的 fixed-version
consumer，完成前不得把 A股 Extension描述为完整 checkpoint。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```
