# `scenes/docparse/fusion` 中文 API Reference

成熟度：**internalization candidate**。

`Merge` 把多个 adapter output 合并为 `Result`，按字段 key 和 confidence 做确定性选择，
归一化 page refs、bbox、table cells 与 warnings，并计算 `answer_ready`、
`review_required` 或 `failed`。缺值或缺 evidence 不会被补造。

本包不调用模型、不验证外部事实，也不代表业务审核已经通过。
