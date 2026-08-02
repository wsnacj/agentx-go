# Browser Platform fixed-version consumer

该独立 Go module 只使用固定 pseudo-version，不使用 `replace`，也不依赖 HS、Runner、Scene、
provider、credential 或真实网络。它以显式 fake backend 验证统一 `browser` tool 的注册、执行、
取消与受约束参数修复，并以显式 status probe 验证 browserd Host 的无副作用构造和状态合同。

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```

这是 Developer Preview / Experimental 的外部接入证据，不是 Public/Beta/Stable 或正式发行证明。
