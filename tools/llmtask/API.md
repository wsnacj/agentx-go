# `tools/llmtask` 中文 API Reference

成熟度：`Experimental extension / Developer Preview candidate`

`llmtask` 提供一个有界的 LLM-only JSON 子任务工具。canonical implementation 负责参数兼容、
模型输入构造、`tool_choice=none`、JSON Schema 子集校验、响应提取、取消与 deadline；Host 显式
注入一次模型调用，不会在包内发现 provider、credential、endpoint 或全局模型配置。

## 接入

```go
registry := tools.NewRegistry()
llmtask.Register(registry, llmtask.Options{
    ModelConfig: "analysis-model",
    ChatWithInput: func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error) {
        return model.Chat(ctx, input)
    },
})
```

公共合同：

```go
const Name = "llm_task"

type ChatWithInputFunc func(context.Context, llm.ChatInput) (*llm.ChatResponse, error)

type Options struct {
    ModelConfig        string
    AllowModelOverride bool
    DefaultTimeoutMs   int
    MaxContentChars    int
    ChatWithInput      ChatWithInputFunc
}

func Definition() tool.Definition
func Register(tool.Registrar, Options)
func NewHandler(Options) tool.Handler
```

`ChatWithInput` 是唯一模型副作用边界。未提供该函数时 `Register` 不注册工具；直接构造的
`NewHandler` 也会 fail closed，不回退到环境变量或全局 provider。

## 行为合同

- `instruction` 必填；兼容读取 `task/prompt/goal/request/query`，但这些别名不进入推荐 schema；
- `input` 可选，兼容 `context`；默认模型由 Host 固定，只有 `AllowModelOverride=true` 时才能切换；
- 默认请求 JSON object；传入 `schema/output_schema` 时请求 `json_schema` 且默认 `strict=true`；
- 模型调用固定 `tool_choice=none`，避免子任务递归执行工具；
- 输出必须是 JSON object/array，并兼容 fenced JSON、字符串外注释和前后说明文本；
- schema 校验支持 object、array、string、number、integer、boolean、null，以及常用
  `required/properties/items/enum/const/range/pattern/additionalProperties` 约束；
- `DefaultTimeoutMs` 默认 45 秒，`MaxContentChars` 默认 120000 字符，调用方可继续收窄。

## 错误、取消与并发

- 缺少 instruction/model 或 JSON 参数无效时返回 `runtime/toolerrors.ToolArgumentError`；
- 模型错误保留 `errors.Is/As` identity，仅增加 `llm_task: chat failed` 上下文；
- cancellation/deadline 通过派生 context 交给 Host adapter，不启动后台 goroutine；
- handler 本身无共享可变状态，可并发调用；注入的模型 adapter 是否并发安全由 Host 声明；
- schema 不匹配、空响应和非 JSON 响应在返回前失败，不产生其他副作用。

## Host 责任与 non-goal

Host 继续拥有 provider/model 注册、credential、租户路由、授权、限流、计费、重试、审计与内容
策略。该工具不是 Session、Task、Subagent 或 durable lifecycle：需要 fanout、子 Session、恢复、
队列和长任务时，应使用 Runtime/Host Kit 的相应能力，而不是把它们伪装成 `llm_task`。
