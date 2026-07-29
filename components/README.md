# AgentX Go Components

本目录是 `agentx-go` 仓库中经过 focused owner/consumer review 的通用组件
module：

```text
github.com/wsnacj/agentx-go/components
```

当前只包含 [`llm`](./llm/API.md) 合同类型候选。它来自 HS
`core/llmx/types` 的机械迁移，用于让 AgentX、HS 业务代码和外部项目共享同一份
LLM 请求、响应、工具与流式事件合同。

当前成熟度为 **private validation / Experimental**：

- 不是根 `agentx` 执行合同；
- 不提供 AgentX Runtime、模型 provider、credential 或网络客户端；
- 不代表目录内所有导出符号已获得 Public/Beta/Stable 承诺；
- 正式发布仍受 license、版本、平台、安全和 release authorization 门禁约束。

当前 HS/conformance 固定验证版本：

```text
v0.0.0-20260729125257-bb6949793309
```

这是不可变 private pseudo-version，不是正式 tag 或 semver 发布。

组件之间不得通过 `common` 或杂项 helper 形成隐式依赖。每个新增 package 都必须
有独立 owner、consumer、API 文档和 module-local 验证。
