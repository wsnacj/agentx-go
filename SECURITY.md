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

本仓安全入口覆盖root、components、runtime、extensions、providers、tools、browser、
document和scenes九个library module。这里的scenes是新仓中的portable Domain Kit；HS中的
具体Scene Host、credential、授权策略、真实backend和部署环境仍由各自owner负责。若问题
跨越边界，仓库maintainer会协助路由。当前只支持文档记录的固定版本矩阵，不承诺旧
pseudo-version安全回补。

## 自动扫描边界

Pre-Beta候选使用固定
`golang.org/x/vuln/cmd/govulncheck@v1.6.0`扫描九module全部package。执行工具链与
source-mode标准库模型均显式固定为Go 1.25.12；仅用新版Go启动扫描器不能替代后者。
M6D曾在Go 1.25.5上命中10个标准库可达漏洞并按fail-closed停止，升级到Go 1.25.12后
当前可达漏洞为0。漏洞数据库或工具下载不可达时只允许有界重试，不能把
执行失败写成零漏洞。

当前extensions与document graph各有1个不可达module-level finding：`GO-2026-5024`位于
`golang.org/x/sys`；scenes另有10个`golang.org/x/net` module-only finding。九module
production import/call graph均未触达这些记录，已知可达漏洞为0。直接把extensions升级到
当前修复版本会同时提升最低Go版本，因此该结论不是永久豁免；具名security approver仍需
在Beta前决定兼容升级、工具链提升或接受残余平台边界。自动扫描也不能覆盖
反射/unsafe不可见调用、Host credential、授权策略、部署配置或业务readback。

正式Public/Beta发布前仍需具名security approver、漏洞响应授权、license和release
流程；本文件不代表这些门禁已经关闭。
