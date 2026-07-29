// contract-basic 展示 W1 contract-only Client 的最小成功路径。
//
// 这个示例不访问网络、不读取 secret，也不代表完整 Runtime 已交付。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

type echoAdapter struct{}

func (echoAdapter) Run(
	_ context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	return &agentx.AdapterRunResult{
		RunID:     "run-contract-basic",
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
		exit(err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "hello",
		SessionID: "contract-basic-session",
	})
	if err != nil {
		exit(err)
	}
	if result.Status != "completed" ||
		result.Reply != "echo: hello" ||
		result.SessionID != "contract-basic-session" {
		exit(fmt.Errorf("unexpected result: %#v", result))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		exit(err)
	}
	fmt.Printf("%s %s\n", result.Status, result.SessionID)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
