# General Tools Fixed-Version Consumer

这个独立 Go module 用固定 pseudo-version 组合 `tools`、`diffs`、`message`、
`httprequest`、`filesystem`、`memory` 与 `scheduler`。它不使用 `replace`，不依赖
HS、Runner 或 Scene，也不会访问真实网络、文件系统、凭据或调度后端。

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```

程序通过内存 workspace 和显式 fake ports 执行十个已注册工具，证明 P2-D 的
portable coordination 可以被仓库外 Host 按固定版本组合。这里验证的是
Developer Preview 接入，不是正式发布、兼容性或发行授权证据。
