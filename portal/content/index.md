---
layout: home
title: AgentX Go
titleTemplate: 中文开发者文档

hero:
  name: AgentX Go
  text: 可组合、可治理的 Agent Runtime
  tagline: 独立 Go SDK · 明确 Host 边界 · 中文 API Reference
  actions:
    - theme: brand
      text: 5 分钟开始
      link: /docs/quickstart
    - theme: alt
      text: 选择接入路径
      link: /docs/concepts/execution-model
    - theme: alt
      text: 浏览 Package API
      link: /packages

features:
  - title: 自定义 ExecutionAdapter
    details: 从最小 Client、Run、typed error 与 Shutdown 合同开始，完全控制执行底座。
    link: /docs/guides/custom-adapter
  - title: Model / Tool Host Kit
    details: 显式注入模型和工具能力，以更少样板代码运行 Open Tool Loop。
    link: /docs/guides/model-tool-hostkit
  - title: Workflow Host Kit
    details: 组合 lowering、validation、journal、node execution 与 orchestration portable mechanism。
    link: /docs/guides/workflow-hostkit
  - title: 成熟度可见
    details: 44 个 package 与 8 个 Developer Preview candidate 分级展示，不把 exported 误报为 Public API。
    link: /docs/reference/package-maturity
  - title: 生命周期合同
    details: 明确并发、context cancellation、deadline、typed error 与有界幂等 Shutdown。
    link: /docs/guides/lifecycle-and-errors
  - title: 分发边界
    details: Private validation 已可重复验证；Public Beta 仍对许可证、安全与发行授权 fail closed。
    link: /docs/reference/distribution-readiness
---

::: warning Developer Preview
当前站点是private Developer Preview Candidate，不构成Public、Beta、Stable、生产SLA或
正式发行授权。调用方应固定文档记录的四module伪版本。
:::

## 两类标准执行路径

- **Open Tool Loop**：使用Model / Tool Host Kit组合模型请求、工具调用与结果协调；
- **Workflow**：使用Workflow Host Kit执行显式图、结构校验、状态转换和durable journal。

根`agentx.Client`与自定义`ExecutionAdapter`是最窄合同接入方式，不是第三种执行语义。

