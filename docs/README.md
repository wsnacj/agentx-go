# AgentX Go 中文文档

从一个可运行示例开始，再根据项目需要选择 Runtime、能力模块和领域 Kit。

## 第一次接入

1. [快速开始](quickstart.md)：运行最小 Chat；
2. [能力与接入路径](guides/capability-map.md)：确认七类能力和推荐入口；
3. [安装与多 Module 引用](guides/installation-and-modules.md)：选择依赖；
4. [生命周期与错误](guides/lifecycle-and-errors.md)：处理取消、并发和关闭；
5. [Package API 与成熟度](reference/package-maturity.md)：确认兼容等级。

## 标准接入路径

- [Chat](guides/chat.md)
- [Model / Tool Host Kit](guides/model-tool-hostkit.md)
- [Tool Direct Answer](guides/tool-direct-answer.md)
- [Workflow Host Kit](guides/workflow-hostkit.md)
- [Objective Host Kit](guides/objective-hostkit.md)
- [Session / Subagent Host Kit](guides/session-subagent-hostkit.md)
- [自定义 ExecutionAdapter](guides/custom-adapter.md)

## 扩展与能力模块

- [Portable Skills](guides/portable-skills.md)
- [A股 Portable Extension](guides/astock-extension.md)
- [Reference Host](guides/reference-host.md)
- [Product Shell Preparation](guides/product-shell-preparation.md)
- [Product Shell Observation Handoff](guides/product-shell-observation-handoff.md)
- [Run Data Plane](guides/run-data-plane.md)

各 package 的完整中文 Reference 位于对应源码目录的 `API.md`。统一索引见
[Package API 与成熟度](reference/package-maturity.md)或本地文档站的 Package API 页面。

## 概念与政策

- [执行模型](concepts/execution-model.md)
- [成熟度与兼容边界](maturity.md)
- [Developer Preview 政策](guides/developer-preview-policy.md)
- [版本、升级与回滚](guides/versioning-and-upgrades.md)
- [分发 Readiness](reference/distribution-readiness.md)
- [安全政策](../SECURITY.md)
- [支持政策](../SUPPORT.md)

## 示例与验证

- [七类能力示例](../examples/README.md)
- [Reference Host](../examples/reference-host)
- [External-style consumer](../conformance/consumer)
- [Opt-in Provider Smoke](../conformance/live/README.md)

教学 examples 默认无 credential、真实网络、外部进程和文件写入。需要真实环境的验证必须
显式 opt-in，并与教学示例分开。

## 本地文档站

```bash
npm ci
npm run docs:check
npm run docs:dev
```

Markdown 和 package `API.md` 是正文事实源；`portal/.generated` 与构建产物不提交。
