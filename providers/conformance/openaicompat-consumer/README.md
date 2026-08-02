# OpenAI-compatible fixed-version consumer

该独立 module 只依赖固定 pseudo-version 的 `agentx-go/components` 与
`agentx-go/providers`，没有 `replace`、HS、Runner 或 Scene import。它通过显式 `HTTPDoer`
运行离线 fixture，不访问真实网络，也不读取真实 credential。
