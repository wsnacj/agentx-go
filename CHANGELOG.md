# Developer Preview 变更记录

本文件记录可供private consumer复现的Developer Preview checkpoint，不代表正式release、
semver兼容承诺或Public/Beta/Stable声明。

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
