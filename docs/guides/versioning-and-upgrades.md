# 版本、升级与回滚

AgentX Go 当前处于 Developer Preview。正式版本、tag 前缀和长期支持周期以 Release
说明为准。

## 版本原则

- 每个 module 使用 Go module 版本语义；
- 调用方应提交 `go.mod` 和 `go.sum`，固定精确不可变版本；
- 不把分支名、HTTP API 的 `v1` 或数据 schema 的 `*_v1` 解释为稳定 Go module 版本；
- Experimental package 的变更风险高于 Developer Preview candidate；
- 版本升级必须可以回滚到升级前的 `go.mod`/`go.sum`。

## 九 Module 发行前缀

计划采用标准 nested-module tag：

| Module | Tag 前缀 |
| --- | --- |
| root | `vX.Y.Z` |
| components | `components/vX.Y.Z` |
| runtime | `runtime/vX.Y.Z` |
| extensions | `extensions/vX.Y.Z` |
| providers | `providers/vX.Y.Z` |
| tools | `tools/vX.Y.Z` |
| browser | `browser/vX.Y.Z` |
| document | `document/vX.Y.Z` |
| scenes | `scenes/vX.Y.Z` |

首次正式版本发布前，该规则仍是发行候选设计，不代表仓库已经存在相应 tag。

## 升级步骤

1. 阅读 [CHANGELOG](../../CHANGELOG.md)；
2. 检查 [Package 成熟度](../reference/package-maturity.md)；
3. 在独立分支更新一个依赖 cohort；
4. 运行 `go mod tidy -diff`、`go test`、`go test -race` 和 `go vet`；
5. 运行真实 Host 的代表集成案例；
6. 检查 error、JSON、状态、取消、持久化顺序和副作用授权；
7. 合入后保留升级前版本和回滚步骤。

## Breaking change

Developer Preview 的 breaking change 必须：

- 在 CHANGELOG 明确列出；
- 提供替代入口和迁移示例；
- 更新中文 Reference 与签名快照；
- 使用新的不可变版本；
- 通过 external-style consumer；
- 不静默改变安全和副作用边界。

## 回滚

回滚应恢复完整 module 组合，而不是只降级一个高层 module。若问题涉及真实 provider、
Browser、Document 或 Scene，应同时恢复 Host 配置、backend schema 和部署 artifact。
