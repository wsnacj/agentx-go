# AgentX Go Extensions

本目录是 AgentX 可选扩展的共享 private-preview module：

```text
github.com/wsnacj/agentx-go/extensions
```

当前包含：

- [`astock`](./astock/API.md)：M5D technical checkpoint已完成的 A股 portable
  Domain Extension推荐入口，组合 Manifest、immutable assets、tool schema、三组 Pack Definition与
  确定性 evaluator；
- [`astock/contracts`](./astock/contracts/API.md)：A 股领域 portable DTO、JSON
  normalization、status和 assessment mechanism；
- [`astock/hostkit`](./astock/hostkit/API.md)：Host显式注入 handler的 A股 intent、
  investigation、readiness与回答格式化机制；
- [`domainmodule`](./domainmodule/API.md)：编译期 Domain Module的portable
  manifest、config、diagnostics与顺序注册编排。
- [`pack`](./pack/API.md)：Domain Pack定义、显式校验、注册、Workflow选择/物化、
  路由选择与 Binding机制。
- [`skills`](./skills/API.md)：M5E迁入的 Skill合同、loader/cache、activation、
  requested semantics与资源引用机制。

它不包含行情 provider、livekit、具体 Workflow policy、工具执行、`pack/runtime`
memory/eval backend、Skill prompt ranking/安装执行/安全策略、credential或真实网络。

依赖方向固定为：

```text
extensions -> contract/runtime/components
contract/runtime/components -X-> extensions
```

M5C Portable Pack Core与 M5D `extensions/astock` checkpoint已经 Owner接受；M5E
`extensions/skills`技术 checkpoint已完成并等待 Owner接受。`extensions/astock`是
Developer Preview candidate，`skills`、`astock/hostkit`与其它 extension owner仍为
Experimental，三组 Pack implementation
位于 `astock/internal`且不允许外部直接依赖。共享 module不表示其它 Scene获得独立
发行资格，也不构成 Public/Beta/Stable、正式 tag或 semver承诺。

固定版本、无 HS、无长期 `replace` 的组合验证位于：

- [`conformance/astock-contract-consumer`](conformance/astock-contract-consumer)；
- [`conformance/domain-module-consumer`](conformance/domain-module-consumer)；
- [`conformance/pack-consumer`](conformance/pack-consumer)；
- [`conformance/astock-consumer`](conformance/astock-consumer)：组合完整 portable A股
  Manifest、资产、三组 Pack、fixture Host Kit与 evaluator；
- [`conformance/skills-consumer`](conformance/skills-consumer)：使用固定版本验证
  immutable AssetFS加载、缓存、路径激活、requested semantics、资源完整性和 deep clone。

这些 consumer分别证明 contracts、Domain Module、Pack、A股组合路径与 Portable
Skills可以在无 HS、Runner、长期 `replace`和网络时运行；Skills consumer也不执行
命令或安装副作用。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```
