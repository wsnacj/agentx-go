# 模型对话接入

`runtime/hostkit.NewChatClient`是 C1 Model Conversation 的推荐入口。它为每次
`Client.Run`执行一次宿主提供的模型请求，并复用根合同的 identity、typed error、取消、
并发和`Shutdown(ctx)`语义。

## 最小代码

```go
client, err := hostkit.NewChatClient(hostkit.ChatClientConfig{
    Model:  "your-model",
    System: "请用中文简洁回答",
    RequestModel: func(
        ctx context.Context,
        request llm.ChatRequest,
    ) (llm.ChatResponse, error) {
        return model.Chat(ctx, request)
    },
})
```

`RequestModel`必填；SDK不创建provider、网络client或credential。默认请求把
`RunRequest.Input`转换成一条`Role: "user"`消息，并透传`Model`和`System`。

## 多轮历史

Client不会私自持有conversation backend。需要多轮对话时，调用方在`BuildRequest`中使用
`execution.Request.SessionID`读取历史，并返回完整`llm.ChatRequest`：

```go
BuildRequest: func(
    ctx context.Context,
    request execution.Request,
) (llm.ChatRequest, error) {
    history, err := conversations.Load(ctx, request.SessionID)
    if err != nil {
        return llm.ChatRequest{}, err
    }
    history = append(history, llm.Message{Role: "user", Content: request.Input})
    return llm.ChatRequest{Model: "your-model", Messages: history}, nil
},
```

历史持久化、裁剪、审计和租户隔离属于 Host，不进入 canonical Core。

## 错误、取消与关闭

- caller取消或deadline继续映射为根`CodeCanceled`/`CodeDeadlineExceeded`；
- model错误的 identity通过`errors.Is/As`保留，display-safe文本由根错误合同控制；
- 模型若返回tool calls，Chat路径以`CodeExecutionFailed`失败，不静默执行工具；
- 同一Client的重叠Run按根合同串行化；
- `Shutdown(ctx)`有界且幂等，开始关闭后新Run返回`CodeClientClosed`。

## Non-goal

本入口不提供默认provider、对话数据库、streaming、tool call、Workflow、Objective或
durable Session。需要工具时使用`NewModelToolClient`。
