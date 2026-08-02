# 安全报告政策

AgentX Go当前是私有Developer Preview，不是Public、Beta、Stable或生产支持版本。

## 如何报告

请优先使用GitHub仓库的私有Security Advisory入口报告漏洞。若该入口不可用，请通过
仓库Owner批准的私密渠道联系`@wsnacj`。不要在公开Issue、讨论、日志或提交中披露漏洞
细节、凭据、token、cookie、真实客户数据或可复用攻击载荷。

报告至少应包含：

- 受影响module、package和固定版本；
- 无真实secret的最小复现；
- 影响范围、攻击前提和是否涉及网络/文件/命令副作用；
- 已知缓解方式；
- 报告者希望使用的后续联络方式。

## 响应目标

私有Developer Preview阶段的目标是3个工作日内确认收到、7个工作日内给出初步分级和
下一步。这是best-effort响应目标，不是生产SLA或修复期限。若怀疑凭据已经暴露，应由
对应Host/Provider owner立即吊销或轮换；AgentX Core不保存或恢复调用方secret。

## 范围

本仓安全入口覆盖root、components、runtime和extensions四个module。HS、Scene、具体
provider、credential、授权策略、真实backend和部署环境仍由各自owner负责，但若问题
跨越边界，Core maintainer会协助路由。当前只支持文档记录的统一固定版本，不承诺旧
pseudo-version安全回补。

正式Public/Beta发布前仍需具名security approver、漏洞响应授权、license和release
流程；本文件不代表这些门禁已经关闭。
