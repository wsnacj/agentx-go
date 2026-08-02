# `scenes/docparse/understanding` 中文 API Reference

成熟度：**internalization candidate**。

`New(Options)` 构造纯内存 `Engine`；`PlanOnly` 只匹配 profile 与规划 route，`Run` 按计划
调用显式注入的 `adapters.Registry` 并通过 `fusion.Merge` 聚合结果。未处理 route 会保留在
`UnhandledRoutes`，adapter error 与 context cancellation 会使用 `%w` 保留 identity。

本包不拥有文件、provider、credential、RunStore 或 Scene 生命周期。
