# A 股 Portable Domain Extension 接入

`github.com/wsnacj/agentx-go/scenes/astock`是 A股 portable能力的推荐 Go入口。
它可以离线提供 Manifest、不可变 skill/tool资产、7个模型工具定义、3组 Pack、
确定性 evaluator和一个显式注入 handler的 Host Kit。

它不是行情服务，也不会自行创建 provider、读取凭据、访问网络或注册 Runner。

## 安装

```bash
go get github.com/wsnacj/agentx-go/scenes@v0.2.1
```

可重复构建的项目应把解析结果固定在`go.mod`和`go.sum`中。

## 读取目录与资产

```go
manifest := astock.Manifest()
provider := astock.Assets()
skill, err := fs.ReadFile(astock.ExtensionFS(), "skills/a-stock-data/SKILL.md")
```

`Manifest()`及名称 slice均返回 caller-owned副本；`Assets()`提供 snapshot后的只读
provider与 fingerprint。`ExtensionFS()`包含一份 skill和7份 declarative tool
manifest，不包含 plugin command、provider或 credential。

`ToolDefinitions()`返回模型运行时使用的 `components/llm.Tool`；`tools/*.tool.json`
是带安装/目录元数据的 declarative资产。两者共享稳定 tool identity，但属于不同
消费合同，调用方不应假定完整 JSON逐字节相同。

## 注册并选择 Pack

```go
coordinator, err := pack.NewCoordinator(hostValidator, hostLowerer)
registry, err := pack.NewMemoryRegistry(coordinator)
err = astock.RegisterPacks(registry)

selection, matched := pack.SelectBinding(
    registry,
    "请查询平安银行 A股估值和行情快照",
    pack.SelectOptions{},
)
binding, ok, err := coordinator.ResolveBinding(
    registry,
    selection.Selected.PackID,
    selection.Selected.CaseType,
    selection.Selected.WorkflowID,
)
```

`RegisterPacks`按 valuation→research→signal稳定顺序注册。Validator与
ToolArgumentLowerer必须由 Host显式提供；扩展不会把产品 validation policy、具体
工具映射或 backend藏进默认值。

## 接入 Host handler

```go
payload, err := astockhostkit.BuildAStockInvestigationPayload(
    ctx,
    astockhostkit.InvestigationConfig{
        Source: "my-host",
        Handlers: astockhostkit.InvestigationHandlers{
            Quote: quoteHandler,
            Research: researchHandler,
            Signal: signalHandler,
        },
    },
    modelSuppliedParams,
)
```

handler负责证券身份、来源、时间、freshness、授权和错误证据；Host Kit只按既有
task frame协调调用并聚合 readiness。未注入的能力会产生结构化 missing/unsupported
结果，不会偷偷访问公共站点。

`EvaluateValuationEvidence`、`EvaluateResearchEvidence`和
`EvaluateSignalEvidence`只检查调用方提供的证据。`FormatAStockAnswer`只格式化带
原 tool identity与 readiness合同的 payload，不生成新事实。

## 可运行证据

[`scenes/conformance/astock-consumer`](../../scenes/conformance/astock-consumer)
是独立 nested module，固定远端不可变版本，不使用专有Runner、长期 `replace`
或网络。它完整验证：

```text
Manifest → embedded asset → 三 Pack注册 → route/binding/materialize
→ fixture Host handler → readiness → evaluator
```

```bash
GOWORK=off go -C scenes/conformance/astock-consumer test ./...
GOWORK=off go -C scenes/conformance/astock-consumer run .
```

`scenes`会固定其所需的`extensions`与`runtime`依赖；只有调用方源码直接import这些module时
才需要把它们列为直接依赖。

需要真实行情、研报或信号时，调用方应提供具体 Host adapter。provider、credential、
cache、source priority、fallback、网络与生产 readiness不属于本 portable API。
