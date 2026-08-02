# PDF API（Experimental）

`pdf` 提供 provider-neutral 的 PDF 结果合同、表格/文本格式化、结构分析和显式 backend
协调器。package 名暂保留为 `pdfparser`，降低 HS consumer cutover 风险。

## 推荐构造

```go
runner, err := python.New(python.Config{
    PythonPath: "/absolute/path/to/python3",
    ScriptPath: scriptPath,
    Environment: controlledEnv,
})
parser, err := pdfparser.NewParser(runner)
result, err := parser.ParsePDFContext(ctx, file, false)
```

`NewParser` 不发现解释器、不读取环境变量、不启动进程。真实执行只发生在调用 parse 方法后，
且必须经调用方显式注入的 `Runner`。测试、远程 service 或其它 native backend 可以实现同一端口。

## context 与并发

- `ParsePDFContext`、`ParsePDFWithOptionsContext` 传播 cancellation/deadline；
- 未传 deadline 时，parser 使用五分钟兼容上限，可用 `WithTimeout` 收窄；
- parser 构造后只读取 Runner 与 timeout；是否可并发最终取决于注入 Runner；
- `python.Runner` 每次调用创建独立 command，不保存请求状态，可并发使用。

## Python adapter

`pdf/python` 是显式 optional adapter，不是默认 backend。它包含既有解析脚本以支持 module zip，
但 `PythonPath`、`ScriptPath`、环境变量和工作目录均由 Host 明确提供。canonical module 不扫描
PATH、不读取 credential、不选择虚拟环境，也不授予进程权限。

## Non-goal

当前不承诺 Python 依赖自动安装、生产 sandbox、文件授权、artifact 发布或客户字段策略；这些
属于 Host/Product 层。`Experimental` 也不构成 Public/Beta/Stable 兼容承诺。
