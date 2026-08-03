# Finance Domain Kit fixed-version consumer

本目录是独立 Go module，固定 `agentx-go/scenes` pseudo-version，不使用 `replace`、HS、
Runner、credential、真实网络、交易或文件副作用。它同时验证 `globalstock` 与 `finance`
的推荐 Host-port coordination。

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```

