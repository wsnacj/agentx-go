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
| Go module基线 | `go 1.24.1`；private consumer应使用Go 1.24.1或更高版本 |
| 实际完整测试主机 | macOS arm64，Go 1.25.5 |
| API/build surface | darwin/arm64与linux/amd64，`CGO_ENABLED=0` |
| Ubuntu真实运行 | 尚未形成M6A正式证据，继续阻断Public Beta |
| CGO/native | Core推荐路径不要求；未声明的native能力不属于支持面 |

provider、credential、授权/审批、真实网络、durable backend、Scene业务规则、部署容量和
生产可用性不属于本仓Developer Preview支持承诺。
