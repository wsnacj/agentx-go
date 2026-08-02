# 支持边界

AgentX Go当前只提供private Developer Preview的best-effort开发支持，没有生产SLA、长期
支持版本或兼容保证。

## 当前支持对象

- 文档中记录的四module统一fixed pseudo-version；
- 自定义`ExecutionAdapter`、Model/Tool Host Kit和Workflow Host Kit三条标准路径；
- 成熟度矩阵中标为Developer Preview candidate的8个package入口；
- 无HS、无Runner、无长期`replace`的conformance consumer。

非敏感缺陷和文档问题可通过仓库Issue或Owner批准的协作渠道提交。安全问题必须遵循
[SECURITY.md](SECURITY.md)，不得进入公开Issue。

问题报告应包含`go version`、`go env GOOS GOARCH CGO_ENABLED`、四module最终
`go list -m`结果、最小复现、期望/实际行为和已运行的gate。不得附带真实credential。

## 当前验证矩阵

| 维度 | 当前事实 |
| --- | --- |
| Go module语言基线 | `go 1.24.1`；该directive不是Go 1.24 patch安全支持承诺 |
| 当前安全候选工具链 | Go 1.25.12；M6D发现Go 1.25.5标准库可达漏洞后已fail closed升级 |
| 已接受完整测试主机 | macOS arm64，以及M6C的Ubuntu 24.04.4 amd64/Go 1.25.5历史证据 |
| API/build surface | darwin/arm64与linux/amd64，`CGO_ENABLED=0` |
| Ubuntu真实运行 | M6C run `30732109611`在`3c3c7fa46a28`完成四module normal/race/vet/tidy/list、fixed consumer与artifact gate |
| CGO/native | Ubuntu批准矩阵以`CGO_ENABLED=1`通过；Core推荐路径仍不要求CGO，未声明的native能力不属于支持面 |

M6C只证明上述单一Ubuntu/amd64/Go版本矩阵；它不自动承诺其它发行版、架构、Go版本、
系统级native依赖或生产SLA。可复跑入口为`.github/workflows/m6c-ubuntu-runtime.yml`与
`go run ./scripts/check_ubuntu_runtime.go`，后者会拒绝非Linux amd64或非Go 1.25.12主机。
M6D把当前安全候选固定到Go 1.25.12；在正式批准最低支持工具链前，不得从`go.mod`的
语言directive推导Go 1.24运行时仍受安全支持。

provider、credential、授权/审批、真实网络、durable backend、Scene业务规则、部署容量和
生产可用性不属于本仓Developer Preview支持承诺。
