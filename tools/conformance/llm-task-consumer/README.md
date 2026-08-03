# LLM Task fixed-version consumer

该隔离 module 固定 `github.com/wsnacj/agentx-go/tools` pseudo-version，不使用 `replace`，也不
导入 HS、Runner 或 Scene。它通过显式 fake model adapter 注册并执行 `llm_task`，验证 JSON
提取、schema、model identity 和 `tool_choice=none` 的 external-style 接入合同。

该 consumer 不访问真实 provider、credential 或网络，不构成正式发布证据。
