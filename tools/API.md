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

## Invocation kernel

```go
type ChainExecutor struct {
    Executors []tool.Executor
}
func (ChainExecutor) Execute(context.Context, tool.Call) (tool.Result, error)
func (ChainExecutor) Definitions() []tool.Definition
func NormalizeToolName(string) string
func SortByName([]tool.Definition)
func SanitizeToolDefinitionForBackendCompatibility(tool.Definition) tool.Definition
func SanitizeToolDefinitionsForBackendCompatibility([]tool.Definition) []tool.Definition
```

`ChainExecutor`只在当前executor不知道该tool时继续查找；一旦definition已经确认归属，handler
错误会原样返回，不能被后续executor吞掉。合并definition会去重并稳定排序。schema sanitizer
返回深层复制后的backend-compatible JSON Schema，不修改调用方原对象。

## Result middleware

```go
type ToolContentClassifier func(toolName string) (externalContent, untrustedContent bool)
func BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput) ToolResultMiddlewareEvent
func AppendToolResultMiddlewareTelemetryAttrs(map[string]any, ToolResultMiddlewareEvent) map[string]any
func BuildControlledToolResultTransform(ToolResultTransformInput) ToolResultTransformResult
func AppendToolResultTransformTelemetryAttrs(map[string]any, ToolResultTransformResult) map[string]any
```

observation阶段只生成大小、schema drift、敏感key、binary payload、terminal observation和建议
策略，不静默修改原始结果。controlled transform只有在明确需要且存在raw/artifact reference时
才产生summary envelope；error result始终保留。Host可以用`ClassifyContent`注入Browser、
Document、process或网络的产品信任分类，canonical module不拥有这些策略。

## 当前通用tool

- [`tools/diffs`](./diffs/API.md)：纯文本diff，不读取文件、Git或网络，不产生副作用。

authorization、approval、sandbox、credential、内容信任分类、具体filesystem/process/network/store backend和
产品allowlist/default均由Host或后续显式adapter提供。
