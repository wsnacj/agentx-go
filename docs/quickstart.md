# 快速开始

M3D 提供两条标准接入路径：

1. 已有完整 Runtime时，实现自定义 `ExecutionAdapter`；
2. 新项目需要 portable Open Tool Loop时，使用 Host Kit并显式提供 Model/Tool
   Adapter。

两条路径最终都返回同一个根 `Client`，共享 Run、typed error、context、并发和
Shutdown合同。

## 路径一：自定义 ExecutionAdapter

根合同不创建模型、工具或底层 Runtime；调用方把已有执行系统包装为确定性
`ExecutionAdapter`。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

type echoAdapter struct{}

func (echoAdapter) Run(
	ctx context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	return &agentx.AdapterRunResult{
		RunID:     "run-example",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "echo: " + request.Input,
	}, nil
}

func (echoAdapter) Shutdown(context.Context) error {
	return nil
}

func (echoAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

func main() {
	client, err := agentx.New(agentx.Config{Adapter: echoAdapter{}})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "hello",
		SessionID: "session-example",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Status, result.Reply)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
```

输出：

```text
completed echo: hello
```

完整责任边界见[自定义 Adapter](guides/custom-adapter.md)。

## 路径二：Host Kit + Model/Tool Adapter

普通新项目不需要手写 Factory、BuildRun assembly或 RoundExecutor：

```go
client, err := hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
    MaxRounds: 3,
    ResolveIdentity: func(request execution.Request) (string, string) {
        return "run-example", request.SessionID
    },
    BuildRound: func(
        context.Context,
        execution.Request,
    ) (hostkit.ModelToolRoundConfig, error) {
        return hostkit.ModelToolRoundConfig{
            RequestModel: requestModel,
            ExecuteTools: executeTools,
        }, nil
    },
})
```

`requestModel`和 `executeTools`由调用方实现；Host Kit不安装默认 provider、凭据、
授权或网络副作用。完整代码、可选 ports和 owner边界见
[Host Kit接入指南](guides/model-tool-hostkit.md)，fixed-version运行证据见
[`runtime/conformance/hostkit-consumer`](../runtime/conformance/hostkit-consumer)。

安装前先阅读[安装与多 Module 引用](guides/installation-and-modules.md)。

重要边界：

- `New` 成功后，不再绕过 `Client` 直接调用同一个 adapter。
- 同一 `Client` 上的重叠 `Run` 会被串行化。
- `Shutdown` 可以与正在执行的 `Run` 并发；adapter 应取消并收敛该执行。
- `Shutdown` 一旦开始，新的 `Run` 返回 `CodeClientClosed`。
- 两个示例只证明候选合同和 portable Runtime切片可用，不证明 hostless完整 Agent
  Runtime或正式发布已交付。
