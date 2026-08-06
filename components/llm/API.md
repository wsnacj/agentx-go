# `components/llm` 中文 API Reference

## 当前定位

成熟度：**v0.2.1 Developer Preview candidate**。本包进入9包核心兼容候选面及
签名/文档漂移门禁，但不构成Beta、Stable或生产SLA。

导入路径：

```go
import llm "github.com/wsnacj/agentx-go/components/llm"
```

本package定义provider-neutral的LLM请求、响应、工具调用、多模态、usage
和流式事件合同，是AgentX Go内对应合同的单一source authority。

本 package：

- 不创建 AgentX `Client` 或 Runtime；
- 不选择模型/provider，不读取配置或 credential；
- 不执行 HTTP、文件、数据库或其它生产副作用；
- 不包含 provider adapter、重试、metrics、tracing 或业务 Scene；
- 生产代码只依赖 Go 标准库。

## 使用边界

### 值与所有权

输入、响应和事件大多是普通 Go 值，但 slice、map、pointer、`Raw []byte`、
callback 和 channel 仍遵循 Go 的引用语义。调用方在把值交给其它 goroutine 或
provider adapter 后，不应并发修改其可变成员。

`ChatInput.Clone`、`VisionInput.Clone`、`BotInput.Clone` 和
`EmbedInput.Clone` 会复制顶层 slice/pointer，并调用 options 的 `Clone`；
它们不是任意对象图的通用深拷贝。尤其是 `Tool.Function.Parameters` 等嵌套 map
仍应视为调用方冻结的数据。

`RequestOptions.Clone`、`EmbeddingOptions.Clone` 会复制 package 明确拥有的
map/slice/pointer 字段；callback 函数本身按值保留。`ToMap` 返回的新 map 可以由
调用方修改，但 map 中无法识别的任意对象仍可能保留其原生引用语义。

### Callback 与 context

`PayloadHook` 和 `ResponseHook` 由真正的 provider/runtime owner 调用。本 package
只定义函数签名，不决定调用次数、并发度、超时或重试。实现必须尊重收到的
`context.Context`，不得把 credential 或原始敏感响应写入可展示错误。

### Stream 与 Cancel

`StreamResult`、`EventStreamResult`、`SimpleStreamResult` 的 `Ch` 为只读
channel，producer 负责关闭。调用方应持续消费直到 channel 关闭或调用
`Cancel`。`Cancel` 可为 nil；本合同不额外承诺任意 provider 的 cancel 幂等性，
该责任由创建 stream 的 owner 说明。

三个 bridge 会启动 goroutine 转换事件；下游停止消费时应先调用 `Cancel` 并继续
按 owner 合同收敛，避免 producer 因发送阻塞。

## 请求、响应与消息

| API | 说明 |
| --- | --- |
| `ChatInput` | 面向 service 的 typed chat 输入，包含逻辑配置名、系统提示、消息、请求选项和工具 |
| `VisionInput` | 在 chat 输入上增加多模态内容 |
| `BotInput` | 高层 bot 请求输入 |
| `EmbedInput` | 文本/图片 embedding 输入 |
| `ChatRequest` | provider-neutral chat 请求快照 |
| `VisualRequest` | provider-neutral multimodal 请求快照 |
| `EmbeddingRequest` | provider-neutral embedding 请求快照 |
| `ChatResponse` | 文本、工具调用、usage 和原始载荷响应 |
| `VisualResponse` | 多模态响应 |
| `EmbeddingResponse` | dense/sparse vector 响应 |
| `BotResponse` | 高层内容、引用、usage、request identity 响应 |
| `BotReference` | 高层响应引用 |
| `BotUsage` | bot usage 聚合 |
| `BotUsageAction` | action usage |
| `BotUsageModel` | model usage |
| `Conversation` | 有序 `Message` 集合 |
| `Message` | role、content、tool identity 和 tool calls |
| `FunctionResult` | 工具调用及其 output/error |
| `SparseEntry` | sparse vector 的 index/value |

## Options 与 Hook

| API | 说明 |
| --- | --- |
| `RequestOptions` | chat/vision 的 typed options；未知 provider 字段保留在 `ProviderFields` |
| `EmbeddingOptions` | embedding 的 typed options |
| `ThinkingOptions` | thinking enable/mode |
| `ReasoningOptions` | reasoning effort |
| `PayloadHook` | provider payload marshal/send 前的 hook 签名 |
| `ResponseHook` | provider response decode 前的 hook 签名 |
| `ResponseMetadata` | method、URL、status 和 headers |

转换函数：

```go
func RequestOptionsFromMap(map[string]any) RequestOptions
func (RequestOptions) ToMap() map[string]any
func (RequestOptions) Clone() RequestOptions

func EmbeddingOptionsFromMap(map[string]any) EmbeddingOptions
func (EmbeddingOptions) ToMap() map[string]any
func (EmbeddingOptions) Clone() EmbeddingOptions
```

未被 typed field 接受的 legacy key 会进入 `ProviderFields`。这用于迁移兼容，
不是鼓励把新的通用合同继续堆入任意 map。

## Tool 合同

| API | 说明 |
| --- | --- |
| `Tool` | 当前承载 function tool |
| `Function` | name、description、parameters、output schema、strict |
| `FunctionCall` | 完整工具调用 identity/name/arguments |
| `FunctionCallDelta` | 流式工具调用增量 |
| `ToolChoice` | 工具选择策略 |
| `ToolChoiceFunction` | 指定 function name |

相关函数：

```go
func SanitizeToolSchemas([]Tool) []Tool
func SanitizeFunctionParametersSchema(map[string]any) map[string]any
func MergeToolCallSnapshot(FunctionCall, FunctionCallDelta) FunctionCall
func SortedFunctionCallIndexes(map[int]FunctionCall) []int
func (FunctionCallDelta) HasName() bool
func (FunctionCallDelta) HasArguments() bool
```

`SanitizeToolSchemas` 返回复制后的 tool slice，不修改 `OutputSchema`；
`OutputSchema` 仍是上层内部输出合同，不应自动发送给 OpenAI-compatible provider。

## Stream 合同

| API | 说明 |
| --- | --- |
| `StreamChunk` | legacy text/tool/usage/error/done 增量 |
| `StreamResult` | legacy chunk channel 与 cancel |
| `StreamEventType` | normalized event kind |
| `StreamStopReason` | normalized stop reason |
| `StreamEvent` | text/thinking/tool/usage/error/done 事件 |
| `EventStreamResult` | normalized event channel 与 cancel |
| `SimpleStreamChunk` | 简化的 text/error/done 增量 |
| `SimpleStreamResult` | 简化 channel 与 cancel |
| `StreamMessageSnapshot` | text/thinking/tool calls 的累计快照 |

事件常量：

```text
StreamEventStart
StreamEventTextStart
StreamEventTextDelta
StreamEventTextEnd
StreamEventThinkingStart
StreamEventThinkingDelta
StreamEventThinkingEnd
StreamEventToolCallStart
StreamEventToolCallDelta
StreamEventToolCallEnd
StreamEventUsage
StreamEventDone
StreamEventError
```

停止原因：

```text
StreamStopReasonStop
StreamStopReasonLength
StreamStopReasonToolUse
StreamStopReasonContentFilter
```

相关函数：

```go
func BridgeLegacyStreamResult(*StreamResult) *EventStreamResult
func BridgeEventStreamResult(*EventStreamResult) *StreamResult
func BridgeEventStreamToSimple(*EventStreamResult) *SimpleStreamResult
func BuildStreamMessageSnapshot(map[int]string, map[int]string, map[int]FunctionCall) *StreamMessageSnapshot
func SortedStringSnapshotIndexes(map[int]string) []int
func NormalizeStreamStopReason(string) StreamStopReason
```

bridge 是 legacy compatibility surface。本次迁移保持其行为，不代表这些 bridge
已被选为未来 Stable API。

## 多模态

| API | 说明 |
| --- | --- |
| `VisualContent` | text/image/video/data URI 内容块 |
| `VisualOption` | 构造内容块的 option |
| `DetailAuto` | 自动 detail |
| `DetailHigh` | high detail |

构造函数：

```go
func NewTextBlock(string, ...VisualOption) VisualContent
func NewImageURL(string, ...VisualOption) VisualContent
func NewVideoURL(string, ...VisualOption) VisualContent
func NewLocalImage(string, ...VisualOption) VisualContent
func NewImageList([]string, ...VisualOption) []VisualContent
func WithDetail(string) VisualOption
func WithFPS(float32) VisualOption
func WithLabels(...string) VisualOption
```

这些函数只构造值，不读取本地图片、下载 URL 或验证 provider 能力。

## Usage

| API | 说明 |
| --- | --- |
| `Usage` | provider-reported prompt/completion/total、cache 和 reasoning token 计数 |
| `UsageRecord` | 带时间、feature 和 metadata 的 usage 记录 |

`PromptTokens`、`CompletionTokens` 与 `TotalTokens` 保持既有 JSON 字段；
`CachedInputTokens`、`CacheWriteTokens` 和 `ReasoningTokens` 是可选补充。它们只承载 provider
返回的实际 usage，不应由请求前估算值填充。Provider 没有返回某一维度时保持零值。

## Model limits 与请求前 Token Count

```go
type ModelLimits struct {
    ContextWindowTokens int64
    MaxInputTokens      int64
    MaxOutputTokens     int64
}

func (ModelLimits) Normalize() ModelLimits

type TokenCount struct {
    Tokens int64
    Exact  bool
    Source string
}

type TokenCountRequest struct {
    Model    string
    System   string
    Messages Conversation
    Tools    []Tool
}

type TokenCounter interface {
    CountInput(context.Context, TokenCountRequest) (TokenCount, error)
}

type TokenCounterFunc func(context.Context, TokenCountRequest) (TokenCount, error)
func (TokenCounterFunc) CountInput(context.Context, TokenCountRequest) (TokenCount, error)
```

`ModelLimits` 是 Host 对一个已配置模型的 provider-neutral 限制快照；零值表示未知，不会被
`Normalize` 猜测。`TokenCounter` 只用于请求前输入计数或保守估算；`Exact=false` 必须继续
作为估算传播，不能冒充 provider-reported `Usage`。Tokenizer 选择、模型 catalog、价格、凭据
和路由属于 Host/Platform。

## ModelCapabilities

```go
type ModelCapabilities struct {
    TextGeneration   bool
    ToolCalling      bool
    VisionInput      bool
    Streaming        bool
    LocalMediaInput  bool
    ReasoningControl bool
    ParallelTools    bool
    BotCompletion    bool
}
```

这是一个最小、provider-neutral 的 adapter 能力描述，不是模型 catalog。字段为 `true`
表示当前 provider adapter/config 明确声明调用方可以使用；`false` 表示调用方不得假设可用，
不代表对应 provider 的所有上游模型永久不支持。Files、Images 等 provider service API 不混入
本结构，credential、配额、地区和租户策略也不属于能力描述。

## 完整导出面

W3-01 冻结以下 exact candidate surface，防止迁移时遗漏；冻结不等于成熟度晋级。

Types（51）：

```text
BotInput BotReference BotResponse BotUsage BotUsageAction BotUsageModel
ChatInput ChatRequest ChatResponse Conversation
EmbedInput EmbeddingOptions EmbeddingRequest EmbeddingResponse
EventStreamResult Function FunctionCall FunctionCallDelta FunctionResult Message ModelCapabilities ModelLimits
PayloadHook ReasoningOptions RequestOptions ResponseHook ResponseMetadata
SimpleStreamChunk SimpleStreamResult SparseEntry StreamChunk StreamEvent
StreamEventType StreamMessageSnapshot StreamResult StreamStopReason ThinkingOptions
TokenCount TokenCountRequest TokenCounter TokenCounterFunc Tool ToolChoice ToolChoiceFunction Usage UsageRecord VisionInput VisualContent
VisualOption VisualRequest VisualResponse
```

Package functions（21）：

```text
BridgeEventStreamResult BridgeEventStreamToSimple BridgeLegacyStreamResult
BuildStreamMessageSnapshot EmbeddingOptionsFromMap MergeToolCallSnapshot
NewImageList NewImageURL NewLocalImage NewTextBlock NewVideoURL
NormalizeStreamStopReason RequestOptionsFromMap SanitizeFunctionParametersSchema
SanitizeProviderOptionMap SanitizeToolSchemas SortedFunctionCallIndexes
SortedStringSnapshotIndexes WithDetail WithFPS WithLabels
```

Methods（12）：

```text
BotInput.Clone ChatInput.Clone EmbedInput.Clone
EmbeddingOptions.Clone EmbeddingOptions.ToMap
FunctionCallDelta.HasArguments FunctionCallDelta.HasName
ModelLimits.Normalize RequestOptions.Clone RequestOptions.ToMap
TokenCounterFunc.CountInput VisionInput.Clone
```

Constants（19）：

```text
DetailAuto DetailHigh
StreamEventStart StreamEventTextStart StreamEventTextDelta StreamEventTextEnd
StreamEventThinkingStart StreamEventThinkingDelta StreamEventThinkingEnd
StreamEventToolCallStart StreamEventToolCallDelta StreamEventToolCallEnd
StreamEventUsage StreamEventDone StreamEventError
StreamStopReasonStop StreamStopReasonLength StreamStopReasonToolUse
StreamStopReasonContentFilter
```

## 明确 non-goal

- AgentX `Client`、`Run`、`Shutdown` 和执行状态机；
- provider 注册、模型 catalog、配置加载和 credential；
- HTTP transport、retry、metrics、tracing、usage persistence；
- Workflow、Objective、Resume、durable lifecycle；
- Runtime protocol、Scene、CLI、HTTP API；
- 对本页 51/21/12/19 个导出项作 Public/Beta/Stable 承诺。
