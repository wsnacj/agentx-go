# 快速开始

M3E 提供两类标准执行路径、三个接入选择：

1. Open Tool Loop：已有完整 Runtime时实现自定义 `ExecutionAdapter`，普通新项目
   使用 Host Kit并显式提供 Model/Tool Adapter；
2. Workflow：使用 Workflow Host Kit并显式提供 validator、mapper、executor、
   identity、clock和可选 durable port。

Open Tool Loop两种接入最终都返回根 `Client`，共享 Run、typed error、context、
并发和 Shutdown合同。Workflow保持独立显式图 Runtime，不伪装成根 Client mode。

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

## 路径三：Workflow Host Kit

Workflow调用方只需导入 `workflow`和 `workflow/hostkit`，不再手工组合六个低层
package：

```go
runtime, err := workflowhostkit.New(workflowhostkit.Config{
    Validator:          validator,
    Mapper:             mapper,
    BasicExecutor:      executor,
    NewRunID:           newRunID,
    NewNodeExecutionID: newNodeExecutionID,
    NowUnixMilli:       nowUnixMilli,
})
result, err := runtime.Run(ctx, spec, workflowhostkit.Inputs{})
```

Host Kit复用 canonical validation/lowering/journal/nodeexec/orchestration/composition，
不安装产品 policy、provider或 backend。完整示例与 durable、取消、并发边界见
[Workflow Host Kit指南](guides/workflow-hostkit.md)，fixed-version证据见
[`runtime/conformance/workflow-hostkit-consumer`](../runtime/conformance/workflow-hostkit-consumer)。

安装前先阅读[安装与多 Module 引用](guides/installation-and-modules.md)。

重要边界：

- 以下 Client/Shutdown条目只适用于路径一和路径二；Workflow Host Kit不拥有 Host
  资源，因此没有虚假的 Shutdown。
- `New` 成功后，不再绕过 `Client` 直接调用同一个 adapter。
- 同一 `Client` 上的重叠 `Run` 会被串行化。
- `Shutdown` 可以与正在执行的 `Run` 并发；adapter 应取消并收敛该执行。
- `Shutdown` 一旦开始，新的 `Run` 返回 `CodeClientClosed`。
- 这些示例只证明候选合同和 portable Runtime切片可用，不证明 hostless完整 Agent
  Runtime或正式发布已交付。
