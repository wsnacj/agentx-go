# Skills fixed-version consumer

该 consumer 只依赖已推送的 `agentx-go/extensions` 与 `agentx-go/runtime` 固定
pseudo-version，不使用 HS、Runner、长期 `replace`、网络或命令执行。它验证 immutable
AssetFS、Skill加载与缓存、路径激活、requested semantics及资源引用检查。

```bash
GOWORK=off GOPROXY=off go test ./...
GOWORK=off GOPROXY=off go run .
```
