# Source Acquisition fixed-version consumer

本目录是独立 Go module，固定 `agentx-go/scenes` pseudo-version，不使用 `replace`、
HS、Runner、credential、真实网络或文件副作用。它同时验证 `publicsource` 与
`wechatarticle` 的推荐 Host-port 组合路径。

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```
