# Developer Preview 变更记录

本文件记录可供private consumer复现的Developer Preview checkpoint，不代表正式release、
semver兼容承诺或Public/Beta/Stable声明。

## Unreleased（M6D技术候选）

- 新增可删除的四module同版file-proxy候选、module zip SHA-256、依赖图、无replace
  consumer和只读cache验证；模拟版本`v0.0.0-m6d.0`不是首次正式版本；
- 固定Go 1.25.12与`govulncheck@v1.6.0`。Go 1.25.5扫描命中的10个标准库可达漏洞已由
  安全patch工具链解除；当前可达漏洞为0；
- 保留extensions中Windows-only `GO-2026-5024`不可达module finding，等待具名
  security approver决定升级或接受残余平台边界；
- 新增无PR依赖的M6D Ubuntu workflow和value-safe候选manifest；不创建tag/release，
  `public_beta_ready=false`。

## v0.0.0-20260802021959-5a41fb0ccb87（M5S / M5T基线）

- 固定根`agentx.Client`自定义`ExecutionAdapter`、Model/Tool Host Kit和Workflow Host Kit
  三条host-provided construction路径；
- 44个production package有中文Reference，8个Developer Preview candidate有可读API
  snapshot、hash、公开类型闭包与darwin/linux签名门禁；
- A股推荐入口的六个evaluator DTO由公开`extensions/astock/contracts`持有，避免公共
  签名依赖Go `internal`包；
- 统一consumer固定root/components/runtime/extensions四module，使用fixture验证三条
  Core路径和A股推荐入口，无HS、Runner、Scene、provider、credential或网络；
- 完全hostless Runtime仍为`not_ready_for_hostless_w2b`；Objective/control大面、Scene、
  正式tag与发行均未进入该基线。

升级和回滚步骤见[版本、升级与回滚](docs/guides/versioning-and-upgrades.md)。
