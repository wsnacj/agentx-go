# 分发 Readiness

本页描述首个正式公开版本需要满足的技术和人工条件，不记录内部执行批次。

## 当前结论

代码、examples、中文 API 和多 module 验证已经达到 Private Developer Preview 候选水平。
Apache-2.0、历史`v0.1.0`九module release-train、9包核心兼容候选面及具名责任已经批准，但
`v0.1.0`曾被重建，不同不可变Go代理、当前origin和既有consumer保存了不一致的zip/checksum。
这些tag不得再次重写，也不得作为首次公开安装基线。

当前production-code基线为`378edb9eb58a1948e13a6db7d43440aa7b02acaa`，包含`v0.1.0`之后的Context、Memory、
Plugin、Connector/MCP、Expert/Team等Experimental增量；agentx-platform和HS开发分支通过固定
`v0.1.1-0...`pseudo-version消费。Public Developer Preview必须选择一个从未发布的新版本，对该
精确source重新执行九module、module zip、clean proxy/direct/cache、文档和安全readback。仓库
public开关继续需要独立授权。

当前active development baseline为`v0.1.1-0.20260805062037-378edb9eb58a`：九个production module的
仓内依赖、examples、HS和agentx-platform均已对齐，normal/vet/tidy、examples与Platform race、API与
中文Reference门禁通过。该pseudo-version只是私有开发基线，不是新tag、Release或兼容性承诺；历史
conformance/release matrix中的`v0.1.0`仅保留为旧train证据，必须由未来全新版本替换后再做公开readback。

## 技术门禁

九个 library module需要独立通过：

- `go test ./...`；
- `go test -race ./...`；
- `go vet ./...`；
- `go mod tidy -diff`；
- `go list ./...`；
- module zip、module cache 和 origin/readback；
- 无本地 `replace` 的 clean-room consumer；
- API signature、package Reference、链接和文档站构建；
- dependency、credential 和可达漏洞扫描。

Browser、Document、CGO 和系统命令路径还需要在受支持平台执行实际验证。

## 人工门禁

| 条件 | 当前状态 |
| --- | --- |
| License / NOTICE | 已采用 Apache-2.0；九个 library module 与 examples module 均携带一致副本 |
| 直接依赖归属摘要 | 已提交；发行二进制时仍须按实际依赖闭包重新复核 |
| 历史私有版本与九 module tag | `v0.1.0`已存在但有重建/checksum冲突，不得作为首次公开基线 |
| 首个公开版本 | 待批准：必须使用从未发布、不可变的新版本并重做完整readback |
| Developer Preview兼容范围 | 已批准：9包核心候选面；其它可导入包保持Experimental |
| 安全响应责任 | `@wsnacj`；目标3个工作日确认收到，无修复SLA |
| 发布审批与回滚责任 | `@wsnacj`；首版暂无backup owner |
| 历史`v0.1.0`技术与安全门禁 | 历史checkpoint曾通过；不能替代新版本重验 |
| 新版本远端tag / Release readback | 未开始；Public准入前保持fail closed |
| GitHub dependency alert | 2026-08-06推送时默认分支报告1个high告警；尚未归因或关闭，阻断Public准入 |
| 正式公开开关 | 待批准 |

任何技术测试通过都不能替代这些决定。

## Artifact 边界

发行 artifact 不得包含：

- credential、真实客户数据和未脱敏日志；
- 本地 `replace`、`go.work`、机器绝对路径和缓存；
- 测试临时文件、Portal 构建目录和 node_modules；
- 私有 Host 配置、业务 policy 和生产 backend；
- 未声明的平台二进制或本地 native 依赖。

## 发布结论

只有正式 Release 中列出的 module/version/platform 才构成可分发版本。开发分支、
任意 commit、文档站构建或测试通过不自动形成 Beta、Stable、SLA 或长期维护承诺。
