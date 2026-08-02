# 参与开发

AgentX Go优先保持架构owner和公共合同清晰。Core只接受substrate-neutral合同与机制；
provider、credential、授权/安全策略、具体backend、Scene规则和真实副作用应留在Host。

## 工作流程

本仓不要求GitHub付费PR能力。受信维护者可以使用以下等价流程：

1. 从已接受checkpoint创建短期分支；
2. 运行受影响package的focused test/race/vet；
3. 修改Developer Preview candidate时更新中文Reference、API snapshot、CHANGELOG和迁移说明；
4. 运行`GOWORK=off go run scripts/check_developer_preview_distribution.go`；
5. 进入Pre-Beta候选时再运行`GOWORK=off go run scripts/check_pre_beta_candidate.go`；
6. 以单一责任组织commit并推送分支；
7. 由CODEOWNERS对明确commit range做人工审阅后合并或继续推进。

PR可以作为审阅载体，但不是准入证据的唯一形式。审阅必须能定位source commit、变更
范围、测试结果和回滚点，不能以聊天确认替代。

## API变更

- 不得静默改变Run、error identity/code、JSON、取消/deadline、并发或Shutdown语义；
- Developer Preview的breaking change必须先获Owner批准，再同时更新snapshot、中文
  Reference、CHANGELOG、consumer和升级说明；
- 仅在确认变更后使用`check_developer_preview_api.go -update-snapshots`更新baseline；
- Experimental package可以调整，但若已有external consumer，仍需记录迁移影响；
- 不得把新的exported symbol自动标为Developer Preview candidate。

## 验证与安全

四个nested module必须分别在`GOWORK=off`下测试、vet和tidy check。示例使用fixture、
占位URL和无副作用adapter，不访问真实provider。不得提交`.env`、token、cookie、私钥、
客户数据、机器本地路径或长期`replace`。

安全问题请遵循[SECURITY.md](SECURITY.md)，不要通过公开commit提供可利用细节。
`check_pre_beta_candidate.go`固定Go 1.25.12与`govulncheck@v1.6.0`，只从tracked source
构建临时候选；不得用`-mod=mod`、长期`replace`或忽略扫描失败来伪造通过。
