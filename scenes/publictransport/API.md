# `scenes/publictransport` 中文 API Reference

成熟度：**Experimental extension**。当前没有 semver、Developer Preview、Public、Beta 或 Stable
兼容承诺；调用方应固定 pseudo-version，并在升级时执行 API 与行为 differential。当前公开签名
依赖同为 Experimental 的 `runtime/controlcontract`，因此不会提前升级为 Developer Preview candidate。

本包是公共交通只读 Domain Kit 的 portable source authority：它提供 typed request/report、
Collector port、provider-neutral Coordinator、确定性库存证据 evaluator 和 AgentX Pack。
它不选择票务或地图服务商，不访问网络，不读取凭据，也不执行订票、购票或支付。

## 最小接入

```go
type collector struct{}

func (collector) CollectPublicTransportTicketEvidence(
    ctx context.Context,
    req publictransport.Request,
) (publictransport.Report, error) {
    // 调用方在这里接入获准的 provider、fixture、cache 或 host service。
    return publictransport.Report{/* display-safe evidence */}, nil
}

coordinator := publictransport.NewCoordinator(collector{})
execution := coordinator.Execute(ctx, runtimeRequest)
```

- `Collector`：唯一运行时 port；Host 负责 provider、endpoint、凭据、网络、限流和合规；
- `Request`：只携带 AgentX runtime request、query/route/date/source/policy refs 和 observed time；
- `Report`、`Evidence`、`InventoryRow`：provider-neutral、display-safe readback；
- `Report.Normalize`：复制并规范化结果；缺 evidence、原始输出或订票/购票尝试会 fail closed；
- `Coordinator`、`NewCoordinator`、`Coordinator.Execute`：每次调用最多执行一次 Collector，
  不重试、不选择 provider、不创建后台 goroutine；
- `Execution`：同时返回 generic `RuntimeAdapterExecutionResult` 与领域 `Report`。

## 身份与 AgentX Control Contract

- `Descriptor`：返回只读 production-adapter descriptor；
- `Registry`：返回单条显式 Host adapter registry snapshot；
- `Strategy`：返回 Objective 模式的 read-only strategy metadata；
- `InventoryEvidenceRef`：从 display-safe query ref 派生确定性 evidence ref；
- `DefaultAdapterRef`、`DefaultStrategyRef`、`DefaultCapabilityRef` 等仅是当前 Developer Preview identity。

## 确定性证据评估

`EvaluateInventory` 接受 `InventoryEvaluationInput`，验证：

- 是否观察到库存 evidence；
- 最小行数；
- 车次前缀；
- 席别字段已观察或确有余票；
- 可选票价 evidence；
- 未发生订票或购票。

`FilterInventoryRowsByEvidence`、`FilterInventoryRowsByObservedEvidence`、`RowsHave*` 和
`InventoryRowMatches*` 提供同一套纯函数判断。`SeatEvidenceModeAvailable` 要求目标席别有余票；
`SeatEvidenceModeObserved` 允许售罄但要求目标席别字段已实际出现。

## Pack 与 Workflow

```go
packCoordinator, _ := pack.NewCoordinator(hostValidator, hostLowerer)
registry, _ := pack.NewMemoryRegistry(packCoordinator)
_ = publictransport.RegisterInto(registry)
spec, _ := publictransport.MaterializedDefaultWorkflow(packCoordinator)
```

- `Definition` / `PackDefinition`：返回 caller-owned Pack Definition；
- `RegisterInto`：注册到显式 canonical Pack registry；nil registry 保持 no-op；
- `MaterializedDefaultWorkflow`：通过调用方注入的 validation/lowering seam 物化默认 Workflow；
- `PackID`、`CaseTypeTicketLookup`、`DefaultWorkflow`、`LookupTool`：portable identity。

Pack 只声明一个 read-only tool，预算为一次调用，side-effect 上限为 read-only。真实 provider、
站码与日期物化、重试、缓存、票价补采和产品回复均由 Host 决定。

## 并发、取消与错误

`Coordinator` 不修改自身字段；只要调用方注入的 Collector 支持并发，同一个 Coordinator 可并发使用。
`Execute` 原样传递 `context.Context`，不隐藏 deadline，不创建后台 goroutine。Collector error 被投影为
`external_dependency_unavailable`；缺 Collector、request 未 ready 或 identity 不匹配均返回 typed
control-contract blocked result。当前 failure reason 与顺序纳入迁移 differential，但正式冻结前仍是
Developer Preview。

## 非目标

- 票务/地图 provider、endpoint、headers、cookie、credential、代理与真实网络；
- 站码目录、日期默认、候选路线放大、重试、缓存与 rate limit；
- 商业条款、票务可售性权威、产品免责声明、authorization、approval 与安全策略；
- 订票、购票、支付、交易确认或任何写副作用；
- HS HTTP API、CLI、产品 Objective wiring、Public/Beta/Stable、semver 或正式发行。
