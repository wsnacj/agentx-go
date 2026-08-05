# AgentX Go Extensions

本目录是 AgentX 可选扩展的共享 private-preview module：

```text
github.com/wsnacj/agentx-go/extensions
```

当前包含：

- [`catalog`](./catalog/API.md)：Tool、Skill、Plugin、Connector、Expert、Team 的统一只读
  discovery envelope、规范化、确定性检索、fingerprint与差异；不执行、安装或路由资产。
- [`connector`](./connector/API.md)：credential-free Connector identity、protocol、transport与
  discovery投影；不包含endpoint、进程、凭据或连接生命周期。
- [`plugin`](./plugin/API.md)：可安装能力包的portable manifest、contained path、依赖请求、
  权限请求与typed error；不安装、不授权且不执行包内容。
- [`mcp`](./mcp/API.md)：稳定MCP `2025-11-25`生命周期、Tool发现/调用和现有Tool合同适配；
  concrete transport、credential与授权仍由Host拥有。
- [`expert`](./expert/API.md)：portable Expert角色、untrusted instruction与显式资产requirements；
  不选择model、不创建Session/Subagent且不执行资产。
- [`team`](./team/API.md)：引用Expert的Team DAG与确定性topological stage plan；不拥有scheduler、
  queue、budget或第二套Agent Loop。
- [`domainmodule`](./domainmodule/API.md)：编译期 Domain Module的portable
  manifest、config、diagnostics与顺序注册编排。
- [`pack`](./pack/API.md)：Domain Pack定义、显式校验、注册、Workflow选择/物化、
  路由选择与 Binding机制。
- [`skills`](./skills/API.md)：M5E迁入的 Skill合同、loader/cache、activation、
  requested semantics与资源引用机制。
- [`productshell`](./productshell/API.md)：M5G迁入的portable input/preparation机制、
  M5H新增的typed observation与display-safe Host UI handoff，以及M5I新增的临时
  Workflow typed plan、有限重试、binding lowering与固定planning stage。

它不包含领域Scene、行情 provider、livekit、具体 Workflow policy、工具执行、`pack/runtime`
memory/eval backend、Skill prompt ranking/安装执行/安全策略、credential或真实网络。A股领域
owner已在P4-A迁入独立[`scenes/astock`](../scenes/astock/API.md)。

依赖方向固定为：

```text
extensions -> contract/runtime/components
contract/runtime/components -X-> extensions
```

M5C Portable Pack Core与M5E Skills、ProductShell等checkpoint已经完成；A股pilot已在P4-A
迁入`scenes` module并删除本module的重复source authority。`skills`与其它 extension owner仍为
Experimental。共享 module不表示任意 Scene获得独立
发行资格，也不构成 Public/Beta/Stable、正式 tag或 semver承诺。

固定版本、无 HS、无长期 `replace` 的组合验证位于：

- [`conformance/domain-module-consumer`](conformance/domain-module-consumer)；
- [`conformance/pack-consumer`](conformance/pack-consumer)；
- [`conformance/skills-consumer`](conformance/skills-consumer)：使用固定版本验证
  immutable AssetFS加载、缓存、路径激活、requested semantics、资源完整性和 deep clone。
- [`conformance/plugin-consumer`](conformance/plugin-consumer)：使用固定pseudo-version验证
  portable Plugin manifest、权限请求边界与typed error，不安装或执行包内容；
- [`conformance/productshell-consumer`](conformance/productshell-consumer)：验证portable
  ProductShell两阶段准备、临时Workflow规划与Host port组合；
- [`conformance/productshell-observation-consumer`](conformance/productshell-observation-consumer)：
  验证typed Session/HostProcess/OperatorLine到display-safe handoff及runtime-use。

这些consumer分别证明Domain Module、Pack、Portable Skills
及ProductShell准备/临时规划/观测路径可以在无HS、Runner、长期`replace`和网络时运行；
它们也不执行命令、安装或真实UI/log delivery副作用。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```
