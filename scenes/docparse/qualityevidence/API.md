# `scenes/docparse/qualityevidence` 中文 API Reference

成熟度：**internalization candidate**。

`Evaluate` 对 Host 已构造的 `representation.Document` 计算页覆盖、required phrase recall、
截断与 OCR 状态；`Skipped` 显式记录未执行 provider lane，永不冒充成功。

该报告只适用于隐私安全 fixture 的可重复质量 readback，不替代真实客户 corpus、OCR
provider 或人工验收。
