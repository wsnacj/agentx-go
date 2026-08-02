# `scenes/docparse/planner` 中文 API Reference

成熟度：**internalization candidate**。

`Planner.PlanRoutes` 根据显式 spec、profile route hints、文档文本可用性和 Host adapter
能力生成按优先级排序的 `Plan`。`Route` 明确 kind、owner、profile/spec 与 reason；
`StatusNeedsReview`、`StatusNoUsableRoute` 不会被静默升级为 ready。

本包不执行 route、不读取文档，也不选择具体 OCR、LLM、card 或 table backend。
