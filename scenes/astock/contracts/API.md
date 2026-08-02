# A 股 Portable Contracts API（Experimental）

`astock/contracts`拥有 A 股数据源、调查、估值和信号结果之间共享的 portable
数据合同，以及确定性的 JSON normalization、状态和 assessment实现。它只依赖
Go标准库。

## 能力边界

- DTO保留 JSON field、`omitempty`和嵌套结构；
- normalization只清理既有字符串、集合和 JSON形状，不访问 provider或网络；
- status/assessment按已有字段确定性投影，不读取 host配置；
- 输入不会被用于发起交易、写缓存或调用模型。

## Evidence evaluator 数据合同

`ResearchEvaluationInput`、`SignalEvaluationInput`和
`ValuationEvaluationInput`描述 Host 已经取得的 evidence；对应的
`ResearchEvaluation`、`SignalEvaluation`和`ValuationEvaluation`保留稳定 JSON
字段，用于承载确定性 guard 结果。实际 evaluator 由推荐入口
`scenes/astock`暴露，本包只拥有 portable DTO，避免公开类型依赖 Go
`internal`实现包。

这些输入不会自行拉取行情、研报或信号，也不会推断交易日、选择 provider 或生成
投资建议。

本包不包含：证券身份解析、交易日/freshness策略、来源优先级、provider client、
credential、缓存、pack/workflow、tool executor或最终回答策略。这些责任继续由
具体 A股 Host/Scene拥有。

## M5D 组合关系

M5D已把本合同与 `scenes/astock`推荐入口、三组 portable Pack Definition及
`scenes/astock/hostkit`组合为一个可独立验证的 A股 Domain Extension，并形成
等待 Owner接受的技术 checkpoint。该组合不会改变本包的 JSON或状态 source
authority，也不会把 provider、livekit、网络或 Host策略下沉到 contracts。

新项目应从 `scenes/astock`开始接入完整 portable extension；只有只需要 DTO、
JSON normalization或 readiness/assessment投影时，才直接导入本包。

## 接入示例

```go
code, market, ok := contracts.NormalizeAStockCode("sz000001")
readiness := contracts.BuildReadiness(
    contracts.AdapterStatusOK,
    contracts.FailureCodeNone,
    true,
    nil,
    nil,
)
```

示例中的 `code`、`market` 和 `ok` 分别表示规范证券代码、市场和解析结果；
`readiness` 是不访问 provider 的确定性证据就绪投影。

调用方应固定 pseudo-version，并针对使用的 JSON字段执行兼容测试。当前
Experimental分级不承诺 semver稳定性，也不表示完整 A股 Scene可独立运行。
