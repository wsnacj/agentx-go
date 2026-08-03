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

## Catalog metadata

```go
type ToolMetadata struct {
    Plugin          string
    Groups          []string
    Type            string
    Source          string
    Capabilities    []string
    AuditTags       []string
    RiskProfile     string
    ReadOnly        *bool
    ConcurrencySafe *bool
    Destructive     *bool
}
```

`ToolMetadata`只描述catalog条目，不执行authorization、approval、sandbox或backend选择。
三个布尔字段使用指针，确保“明确为false”和“未声明”可以区分。`ToolSourceBuiltin`、
`ToolSourceExtension`、`ToolSourceProject`、`ToolSourceCustom`和`ToolSourceUnknown`提供稳定
source值。

## Runtime context

```go
func WithToolSessionID(context.Context, string) context.Context
func ToolSessionIDFromContext(context.Context) string
func WithToolRuntimeNetworkGuard(context.Context, RuntimeNetworkGuard) context.Context
func ToolRuntimeNetworkGuardFromContext(context.Context) (RuntimeNetworkGuard, bool)
```

该上下文合同只携带run/session identity和Host已选择的网络约束覆盖，不决定默认策略，也不
执行网络请求。空值保持原context；CIDR和port列表会去空、去重并稳定排序。命令授权、approval
hook和具体network backend不属于该合同。

## 当前通用tool

- [`tools/diffs`](./diffs/API.md)：纯文本diff，不读取文件、Git或网络，不产生副作用。
- [`tools/message`](./message/API.md)：通过显式`runtime/channel` sender/target执行发送、回复、
  广播、转发、编辑、删除和reaction协调；不拥有credential、平台选择或真实网络。
- [`tools/httprequest`](./httprequest/API.md)：拥有request/response协调与预算收窄，通过显式
  `Preparer`/`HTTPDoer`接入Host URL/redirect/proxy/network policy。
- [`tools/filesystem`](./filesystem/API.md)：拥有`read/write/edit/apply_patch`的参数、schema、
  文本选择、精确替换和patch语法，通过显式`Workspace`接入Host安全/原子文件后端。
- [`tools/memory`](./memory/API.md)：拥有`memory_search/memory_get`的参数、source归一化、
  typed request与预算协调，通过显式`Backend`接入Host store、visibility和ranking策略。
- [`tools/scheduler`](./scheduler/API.md)：拥有`cron`的action解析与命令路由，通过五个显式
  Backend方法接入Host scheduler、RunStore、授权和durable lifecycle。
- [`tools/llmtask`](./llmtask/API.md)：拥有单次LLM-only JSON子任务的参数兼容、模型输入、
  schema校验、响应提取和timeout/cancellation，通过显式`ChatWithInputFunc`接入Host模型。
- [`tools/web`](./web/API.md)：拥有Search、WebFetch、OpenPage、FindInPage的provider协议、
  正文提取、缓存与模型调用协调，通过显式`retrieval.Preparer`接入Host URL、redirect、proxy与
  network policy；凭据只允许显式注入。

authorization、approval、sandbox、credential、内容信任分类、具体filesystem/process/network/store backend和
产品allowlist/default均由Host或后续显式adapter提供。

固定版本组合验证位于
[`tools/conformance/general-tools-consumer`](./conformance/general-tools-consumer)：它在一个
独立module中注册并执行`diffs`、`message`、`http_request`、四个filesystem入口、两个
memory入口和`cron`，只使用内存/fake ports，不依赖HS、Runner、Scene、长期`replace`或
真实副作用。该consumer证明portable coordination可被外部Host组合，不证明Host授权、
sandbox、credential或具体backend已经迁入。

`llm_task`的独立fixed-version验证位于
[`tools/conformance/llm-task-consumer`](./conformance/llm-task-consumer)。它不使用`replace`，
通过fake model adapter验证JSON/schema/model identity和`tool_choice=none`，不访问真实provider。
