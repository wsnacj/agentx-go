# Anthropic / Codex fixed-version consumer

该独立module只依赖固定pseudo-version的`agentx-go/components`与`agentx-go/providers`，
没有`replace`、HS、Runner或Scene import。Anthropic与Codex均通过显式`HTTPDoer`和
`Authorizer`运行离线fixture，不读取真实credential或访问网络。
