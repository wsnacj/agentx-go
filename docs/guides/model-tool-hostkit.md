# 使用 Host Kit 接入 Model/Tool Adapter

这是 M3E 推荐给普通新项目的 Open Tool Loop 接入路径。调用方显式提供模型请求和
工具执行函数，Host Kit 负责 portable round 顺序、多轮 assembly、根 Client、
typed error 和 Shutdown 组合。整个路径不依赖 HS、Runner、Scene 或长期
`replace`。

如果已有完整 Runtime，只需要接入稳定根合同，请使用
[自定义 ExecutionAdapter](custom-adapter.md)。

## 最小示例

```go
package main

import (
    "context"
    "fmt"

    agentx "github.com/wsnacj/agentx-go"
    llm "github.com/wsnacj/agentx-go/components/llm"
    "github.com/wsnacj/agentx-go/runtime/execution"
    "github.com/wsnacj/agentx-go/runtime/hostkit"
    "github.com/wsnacj/agentx-go/runtime/toolloop"
)

func buildRound(
    context.Context,
    execution.Request,
) (hostkit.ModelToolRoundConfig, error) {
    return hostkit.ModelToolRoundConfig{
        RequestModel: func(
            _ context.Context,
            input toolloop.RoundExecutionInput,
        ) (hostkit.ModelResult, error) {
            if input.Round == 1 {
                return hostkit.ModelResult{Response: llm.ChatResponse{
                    Calls: []llm.FunctionCall{{Name: "lookup"}},
                }}, nil
            }
            return hostkit.ModelResult{Response: llm.ChatResponse{
                Content: "done",
            }}, nil
        },
        ExecuteTools: func(
            context.Context,
            hostkit.ModelToolRoundExchange,
        ) (hostkit.ToolResult, error) {
            return hostkit.ToolResult{
                NextChunks: []string{"tool result"},
            }, nil
        },
    }, nil
}

func main() {
    client, err := hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
        MaxRounds: 3,
        ResolveIdentity: func(request execution.Request) (string, string) {
            return "run-example", request.SessionID
        },
        BuildRound: buildRound,
    })
    if err != nil {
        panic(err)
    }

    result, err := client.Run(context.Background(), agentx.RunRequest{
        Input: "inspect agentx",
        SessionID: "session-example",
    })
    if err != nil {
        panic(err)
    }
    defer client.Shutdown(context.Background())
    fmt.Println(result.Status, result.Reply)
}
```

可直接运行的固定版本 consumer 位于
[`runtime/conformance/hostkit-consumer`](../../runtime/conformance/hostkit-consumer)。

## 谁负责什么

Host Kit负责：

- `request → observe → optional gate → tools`固定阶段顺序；
- 无工具调用完成、工具结果 continuation、Host gate终止的结果收口；
- 有界多轮执行与根 `Client` 状态投影；
- context、typed error、并发 Run gate和关闭后调用合同。

调用方负责：

- provider/model请求和响应转换；
- 工具名称解析、授权、approval、sandbox与真实执行；
- RunStore、telemetry backend和产品策略；
- 资源关闭函数以及错误到允许的根错误码分类。

`BuildRound` 每次 Run 返回一组闭包，因此 provider私有对象和 host状态可以留在
闭包中，不进入 canonical API。

## 可选接缝

- `ObserveResponse`：在执行工具前保存或观察模型响应；
- `BeforeTools`：调用 Host拥有的授权、预算或审批 owner；返回 `false`终止本轮；
- `InitialState`：替换默认的单一 input chunk；
- `Shutdown`：有界、幂等地关闭 host资源；
- `ClassifyError`：只返回根合同允许运行期使用的错误码。

需要 loop detector、failure fuse、自定义 continuation policy 或完整
`AssemblyConfig` 时，改用高级 `hostkit.Config + Factory`，不要把这些 policy
硬编码进 `ModelToolRoundAdapter`。

## 并发、取消和关闭

同一个 Client 的重叠 Run仍由根合同串行化。每个 model/tool函数必须使用收到的
context；deadline/cancellation不得替换为字符串错误。`Shutdown(ctx)`开始后新 Run
返回 `CodeClientClosed`。如果 Shutdown 的 context过期，调用方可以用新 context
继续收敛。

## Non-goal

这条路径不提供默认 provider、credential、网络 client、授权策略、RunStore、
Workflow/Objective/Resume 或 Scene。它是 Developer Preview candidate 接入面，
不是 Public、Beta、Stable 或 production-ready 声明。
