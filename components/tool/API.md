# `components/tool` 中文 API Reference

成熟度：`Experimental`

该包定义provider-neutral的工具声明、调用、结果、Handler、Executor、DefinitionProvider和
Registrar合同。`Definition`、`Function`、`Choice`与`Call`有意复用`components/llm`现有
wire type，避免AgentX出现两套JSON/模型调用DTO。

## 合同

```go
type Definition = llm.Tool
type Function = llm.Function
type Choice = llm.ToolChoice
type Call = llm.FunctionCall
type Result = string

type Handler func(context.Context, Call) (Result, error)
type Executor interface {
    Execute(context.Context, Call) (Result, error)
}
type DefinitionProvider interface {
    Definitions() []Definition
}
type Registrar interface {
    Register(Definition, Handler)
}
```

`Result`当前保持文本/JSON字符串合同，以兼容模型tool-result消息和现有HS consumer；它不代表
未来必须把所有artifact或结构化结果压成字符串。授权、sandbox、approval、credential与真实
副作用由Host或具体tool implementation负责，本包不提供默认执行策略。
