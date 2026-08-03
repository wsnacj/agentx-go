# Agent tools fixed-version consumer

该consumer只依赖固定pseudo-version，不使用`replace`，不import HS，也不启动真实worker、网络、
进程或持久化backend。它证明新项目可以用自己的Backend连接`tasks_*`、`subagents`和
`agent_step`模型工具合同。

```bash
go test ./...
go test -race ./...
go vet ./...
go run .
```
