package main

import (
	"context"
	"fmt"
	"os"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
)

func main() {
	client, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "execution conformance",
		SessionID: "execution-conformance",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("agentx-execution-ok:%s:%s\n", result.Status, result.SessionID)
}

func newClient() (*agentx.Client, error) {
	runtime, err := execution.New(conformanceHost{})
	if err != nil {
		return nil, err
	}
	return agentx.New(agentx.Config{Adapter: runtime})
}

type conformanceHost struct{}

func (conformanceHost) Run(_ context.Context, request execution.Request) (*execution.Result, error) {
	return &execution.Result{
		RunID:     "execution-conformance-run",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "agentx-execution-ok",
	}, nil
}

func (conformanceHost) Shutdown(context.Context) error {
	return nil
}

func (conformanceHost) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
