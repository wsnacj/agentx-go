# `scenes/globalstock` 中文 API Reference

成熟度：**Experimental extension**。当前没有 semver、Developer Preview、Public、Beta 或
Stable 兼容承诺；调用方应固定 pseudo-version，并在升级时执行 API/JSON/Workflow differential。

本包是港股/美股只读分析 Kit 的推荐入口，提供工具 identity、quote/valuation Pack、两条
Workflow 与确定性 evaluator。它不访问行情网络、不选择 provider、不读取 token、不交易，
也不生成投资建议或业务免责声明。

## Pack 接入

```go
coordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(coordinator)
_ = globalstock.RegisterInto(registry)

quote, _ := globalstock.MaterializedDefaultWorkflow(coordinator)
comparison, _ := globalstock.MaterializedComparisonWorkflow(coordinator)
```

- `Definition` / `RegisterInto`：返回或注册 caller-owned Pack definition；
- `MaterializedDefaultWorkflow` / `MaterializedComparisonWorkflow`：必须显式提供 canonical
  `pack.Coordinator`，Workflow validation 与 tool argument lowering 仍由 Host 控制；
- `PackID`、`CaseTypeQuote`、`CaseTypeValuation`、`CaseTypeComparison` 与 tool 常量保持
  当前 Experimental identity；
- `ToolNames` / `SkillNames` 返回新的 slice，不注册 Runner 或 provider。

typed payload 与 readiness 在 [`contracts`](contracts/API.md)；handler 协调在
[`hostkit`](hostkit/API.md)。具体行情、公告、研报和信号 provider 由调用方实现。

## 非目标

- 付费行情/研报、provider fallback、credential、cookie、proxy 或真实网络；
- 交易、组合管理、投资结论、适当性、客户策略和业务免责声明；
- HS Runner/Scene registry、Public/Beta/Stable 或正式发行。

