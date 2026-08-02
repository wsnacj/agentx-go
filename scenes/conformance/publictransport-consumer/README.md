# Public Transport fixed-version consumer

该 consumer 固定依赖 `agentx-go/scenes` 与 `agentx-go/runtime` pseudo-version，不使用
`replace`、HS、Runner、旧 Scene import、真实网络、provider 或 credential。它以内存 Collector
验证 provider-neutral exact-once coordination、typed report、库存 evaluator 与 Pack identity。

当前 fixed scenes 版本为 `v0.0.0-20260802235605-d32ccb29a700`；它是私有仓 Developer Preview
验证证据，不是 Public/Beta/Stable、semver 或正式发行声明。
