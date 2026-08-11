# 版本、升级与回滚

AgentX Go 当前发布 `v0.2.2 Developer Preview`。该版本不是Beta或Stable，长期支持周期
仍以Release说明为准。

`v0.2.0` 的公开标签曾在首次发布后移动，公共模块代理可能持有不同源码和校验和，因此不得作为
新依赖使用。`v0.2.1`九个module artifact保持不可移动，但tag源码内7个nested module/consumer仍保存
旧root module checksum，不能作为仓库examples与conformance的完整可复现基线。`v0.2.2`只修复发行元数据、
版本矩阵和fixed consumer，不改变API或运行时行为，是当前推荐组合。

## 版本原则

- 每个 module 使用 Go module 版本语义；
- 调用方应提交 `go.mod` 和 `go.sum`，固定精确不可变版本；
- 不把分支名、HTTP API 的 `v1` 或数据 schema 的 `*_v1` 解释为稳定 Go module 版本；
- Experimental package 的变更风险高于 Developer Preview candidate；
- 版本升级必须可以回滚到升级前的 `go.mod`/`go.sum`。

## v0.2.2 九 Module 发行前缀

首版采用标准nested-module同版tag：

| Module | Tag 前缀 |
| --- | --- |
| root | `v0.2.2` |
| components | `components/v0.2.2` |
| runtime | `runtime/v0.2.2` |
| extensions | `extensions/v0.2.2` |
| providers | `providers/v0.2.2` |
| tools | `tools/v0.2.2` |
| browser | `browser/v0.2.2` |
| document | `document/v0.2.2` |
| scenes | `scenes/v0.2.2` |

九个tag必须指向同一经过验证的release commit；只升级其中一部分不属于受验证组合。

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
