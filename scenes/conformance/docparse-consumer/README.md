# Docparse fixed-version consumer

本 consumer 只依赖固定 pseudo-version 的 `github.com/wsnacj/agentx-go/scenes`，不使用
`replace`、HS、Runner、文件系统或真实 OCR/LLM provider。它验证 Docparse Pack identity、
tool surface 与 inline parse-result evidence projection。

运行：

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```
