---
layout: home
title: AgentX Go
titleTemplate: 中文开发者文档

hero:
  name: AgentX Go
  text: 可组合、可治理的 Agent Runtime
  tagline: 小而稳定的执行合同 · 显式 Host 边界 · 可选 Batteries
  actions:
    - theme: brand
      text: 5 分钟开始
      link: /docs/quickstart
    - theme: alt
      text: 选择能力
      link: /docs/guides/capability-map
    - theme: alt
      text: Package API
      link: /packages

features:
  - title: Model / Tool Host Kit
    details: 显式注入模型和工具，以较少样板代码运行 Chat、Open Tool Loop 和 Tool Direct Answer。
    link: /docs/guides/model-tool-hostkit
  - title: Workflow 与 Objective
    details: 使用独立 Host Kit 组合校验、执行、状态、持久化和目标控制。
    link: /docs/guides/workflow-hostkit
  - title: 长任务与子任务
    details: 组合 Session、Subagent、Scheduler 和 Resume，backend 仍由 Host 选择。
    link: /docs/guides/session-subagent-hostkit
  - title: 可选 Batteries
    details: Providers、Tools、Browser、Document 和 Scenes 按需引用，不污染根合同。
    link: /docs/guides/installation-and-modules
  - title: API 成熟度可见
    details: 推荐入口、Experimental 扩展和 Internal 实现分级展示。
    link: /docs/reference/package-maturity
  - title: 生命周期合同
    details: 明确 context cancellation、deadline、typed error 和有界幂等 Shutdown。
    link: /docs/guides/lifecycle-and-errors
---

::: warning Developer Preview
当前 API 仍可能在首个稳定版本前调整。请固定精确版本、阅读 CHANGELOG，并在升级时运行
项目自己的回归测试。
:::

## 从哪条路径开始

- 只需要模型对话：从 [Chat](/docs/guides/chat) 开始；
- 需要模型调用工具：使用 [Model / Tool Host Kit](/docs/guides/model-tool-hostkit)；
- 需要显式执行图：使用 [Workflow Host Kit](/docs/guides/workflow-hostkit)；
- 需要目标控制：使用 [Objective Host Kit](/docs/guides/objective-hostkit)；
- 需要子任务、调度和恢复：使用
  [Session / Subagent Host Kit](/docs/guides/session-subagent-hostkit)；
- 已有自己的 Runtime：实现 [自定义 ExecutionAdapter](/docs/guides/custom-adapter)。

AgentX 默认不选择 provider、不发现 credential，也不启动真实副作用。模型、工具、授权、
backend 和业务策略始终由调用方 Host 显式提供。
