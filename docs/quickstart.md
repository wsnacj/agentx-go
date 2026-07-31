# 快速开始

当前 W1 是 contract-only 验证版本。它不负责创建模型、工具或底层 Runtime；
调用方需要传入一个实现 `ExecutionAdapter` 的确定性 adapter。
如果希望复用 canonical Open Tool Loop 而不依赖 HS Runner，可改用
[`runtime/hostkit`](../runtime/hostkit/API.md)；可运行的 fixed-version 集成见
[`runtime/conformance/hostkit-consumer`](../runtime/conformance/hostkit-consumer)。

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

重要边界：

- `New` 成功后，不再绕过 `Client` 直接调用同一个 adapter。
- 同一 `Client` 上的重叠 `Run` 会被串行化。
- `Shutdown` 可以与正在执行的 `Run` 并发；adapter 应取消并收敛该执行。
- `Shutdown` 一旦开始，新的 `Run` 返回 `CodeClientClosed`。
- 示例中的 echo adapter 只证明合同可用，不证明完整 Agent Runtime 已交付。
