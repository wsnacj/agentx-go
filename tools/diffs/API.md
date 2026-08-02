# `tools/diffs` 中文 API Reference

成熟度：`Experimental extension`

`diffs`是纯文本通用tool。调用方显式提供before/after；实现不会读取文件、Git状态、网络或
工作区，也不会写入任何内容。

```go
func Definition() tool.Definition
func Register(tool.Registrar)
func Execute(context.Context, tool.Call) (tool.Result, error)
func Run(Request) (tool.Result, error)
```

`Run`适合Host已经完成兼容参数归一化的场景；`Execute`接受标准JSON arguments。支持
`unified`和`semantic`两种格式，返回兼容JSON字符串及additions/deletions/changes/unchanged。
缺少before/after或格式无效时返回`runtime/toolerrors`既有typed error。
