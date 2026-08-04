# 支持政策

## 当前状态

AgentX Go 当前是 Developer Preview：

- 推荐入口有中文 Reference、examples 和可重复测试；
- Experimental package 可能在正式稳定版本前调整；
- 当前不提供生产 SLA、长期支持期限或跨大版本兼容承诺；
- 真实 provider、网络、浏览器、OCR、进程和 durable backend 由调用方环境决定。

## 获取帮助

非敏感问题可以通过仓库 Issue 提交。请包含：

- 使用的 Go 版本、操作系统和架构；
- 相关 module 与精确版本；
- 最小复现代码；
- 期望行为和实际行为；
- 已去除 credential、客户数据和私有 endpoint 的日志。

安全问题必须按照 [SECURITY.md](SECURITY.md) 私下报告。

## 支持的接入路径

优先支持以下路径：

1. 根 `Client` + 自定义 `ExecutionAdapter`；
2. Model / Tool Host Kit；
3. Workflow Host Kit；
4. Objective Host Kit；
5. Session / Subagent Host Kit；
6. 文档中列出的可选 Providers、Tools、Browser、Document 与 Scenes 推荐入口。

直接依赖低层 Experimental package 的调用方需要自行评估升级差异。

## 不包含

- 调用方私有 provider、credential、网络、代理或账号问题；
- 未经授权的生产副作用和第三方服务条款；
- 调用方自定义 Host policy、backend 或 Scene 业务规则；
- 未记录版本、无法复现或包含敏感信息的问题；
- 非正式发布 commit 的长期维护承诺。
