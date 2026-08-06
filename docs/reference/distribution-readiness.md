# 分发 Readiness

本页描述首个正式公开版本需要满足的技术和人工条件，不记录内部执行批次。

## 当前结论

`v0.2.1`是首个公开Developer Preview发行组合：九个library module使用同一版本，推荐的9包核心
兼容候选面进入签名、中文Reference和external consumer门禁，其余可导入package保持Experimental。

正式Release必须证明九module artifact来自同一revision，且test、race、vet、tidy、list、module zip、
clean-room/offline consumer、双平台API签名、文档、License和安全扫描全部通过。创建tag之前的源码或
本地artifact不构成可分发版本。

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
| 发行版本 | `v0.2.1`九module同版组合 |
| Developer Preview兼容范围 | 9包核心候选面；其它可导入包保持Experimental |
| 安全响应责任 | `@wsnacj`；目标3个工作日确认收到，无修复SLA |
| 发布审批与回滚责任 | `@wsnacj`；首版暂无backup maintainer |
| 依赖安全 | Go与Node依赖均须在发布revision重新扫描；可达漏洞必须为0 |
| 远端tag / Release readback | 九个tag和Release必须与同一revision及版本矩阵一致 |
| 仓库可见性 | 以GitHub仓库页面显示状态为准 |

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
