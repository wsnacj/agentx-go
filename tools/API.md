# `tools` 中文 API Reference

成熟度：`Experimental extension`

该module提供线程安全、provider-neutral的工具catalog与可选通用tool implementation。它不
属于Core Runtime依赖，Runtime不会反向导入本module。

## Registry

```go
func NewRegistry() *Registry
func (*Registry) Register(tool.Definition, tool.Handler)
func (*Registry) Execute(context.Context, tool.Call) (tool.Result, error)
func (*Registry) Definitions() []tool.Definition
func (*Registry) Version() uint64
func (*Registry) Reset()
func Ensure(tool.Executor) tool.Executor
```

`Registry`可并发使用；definition按名称稳定排序。重复名称覆盖旧handler并推进version。
缺失名称返回`*ToolNameError`，可用`errors.As`或`AsToolNameError`识别。

## 名称修复

```go
func RepairToolName(string, []string) ToolNameRepairResolution
func NewToolNameError(string, ToolNameRepairResolution) *ToolNameError
func AsToolNameError(error) (*ToolNameError, bool)
```

名称修复只处理大小写、空白、常见分隔符和`tool`后缀，不做模糊猜测。候选不唯一时fail
closed，并返回稳定`ambiguous_tool_name` code。

## 当前通用tool

- [`tools/diffs`](./diffs/API.md)：纯文本diff，不读取文件、Git或网络，不产生副作用。

authorization、approval、sandbox、credential、具体filesystem/process/network/store backend和
产品allowlist/default均由Host或后续显式adapter提供。
