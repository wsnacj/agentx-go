# `scenes/astock` 中文 API Reference

成熟度：**Experimental extension**。本包随 `scenes/v0.2.2`提供，但不进入9包核心
兼容候选面；调用方必须固定精确版本并执行升级 differential。

本包是 A股 portable领域扩展的推荐 Go入口，组合 Manifest、不可变 skill/tool资产、
LLM tool schema、三组 Pack Definition与确定性 evaluator。它不注册 Runner，不访问
网络，也不拥有 provider、credential、cache或 source priority。

## 推荐入口

```go
manifest := astock.Manifest()
assets := astock.Assets()

coordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(coordinator)
_ = astock.RegisterPacks(registry)
```

- `Manifest()`：返回新的 `domainmodule.Manifest`，描述 module、skill、tool、Pack和
  Workflow identity；不执行 Host注册。
- `Assets()`：返回 snapshot后的 `assetfs.Provider`及稳定 fingerprint。
- `ExtensionFS()`：返回只读 `fs.FS`，包含 `skills/`与 declarative `tools/`。
- `Definitions()`：按 valuation→research→signal稳定顺序返回新的 Definition。
- `RegisterPacks(PackRegistrar)`：只要求 `Register(pack.Definition) error`；nil保持
  旧兼容行为，不执行 provider或工具。
- `ToolDefinitions()`：返回7个 caller-owned `components/llm.Tool` schema。

`ToolDefinitions()`是运行时模型工具合同；`ExtensionFS()`中的 `.tool.json`还包含
安装/目录元数据。两者共享 tool identity，但保留迁移前各自的字段和说明，不承诺
完整 JSON逐字节相同。升级时应分别对两类合同做 differential。

## Pack 与 evaluator

- `ValuationDefinition`、`ResearchDefinition`、`SignalDefinition`返回 caller-owned
  Pack Definition。
- `Materialize*Workflow`要求显式 canonical `pack.Coordinator`，因此 Workflow
  admission与 tool-argument lowering继续由 Host拥有。
- `EvaluateValuationEvidence`、`EvaluateResearchEvidence`、
  `EvaluateSignalEvidence`只对已提供 evidence做确定性 guard判断，不拉取数据、
  不生成投资建议。

六个 evaluator 输入/结果 DTO 的 source authority 位于
`scenes/astock/contracts`；本包通过 type alias 提供推荐入口。公开签名不会依赖
Go `internal`实现包，调用方也不应直接导入三组内部 Pack evaluator。

所有 module/tool/skill/Pack/case/workflow identity均有具名常量。返回的 slice、map、
Definition和 Manifest不得被解释为共享可变状态。

## 非目标

- HS `module.Module()`、Runner、extension runtime和 registry adapter；
- livekit、行情/研报/信号 provider、HTTP、credential、cookie、proxy与 cache；
- source优先级、fallback、交易日核验和生产 readiness；
- concrete tool executor、CLI或真实网络；
- Public/Beta/Stable、semver或正式发行承诺。
