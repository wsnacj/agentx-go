# Document Pipeline fixed-version consumer

该隔离模块只依赖固定 pseudo-version 的 `github.com/wsnacj/agentx-go/document`，不依赖 HS，也不使用 `replace`。它验证调用方可显式注入文档加载与章节切分能力，并执行 canonical regex extraction pipeline。

```bash
go test ./...
go run .
```

预期输出：

```json
{"revenue":42,"status":"parsed"}
```

