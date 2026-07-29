// custom-adapter 展示如何实现 W1 ExecutionAdapter，并验证 typed error。
//
// 示例完全确定性，不包含 provider、credential、真实网络或生产副作用。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

var errRejected = errors.New("deterministic adapter rejected input")

type controlledAdapter struct {
	closed atomic.Bool
}

func (a *controlledAdapter) Run(
	_ context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	if a.closed.Load() {
		return nil, errRejected
	}
	if request.Input == "reject" {
		return nil, errRejected
	}
	return &agentx.AdapterRunResult{
		RunID:     "run-custom-adapter",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "accepted",
	}, nil
}

func (a *controlledAdapter) Shutdown(context.Context) error {
	a.closed.Store(true)
	return nil
}

func (*controlledAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

func main() {
	client, err := agentx.New(agentx.Config{Adapter: &controlledAdapter{}})
	if err != nil {
		exit(err)
	}
	_, err = client.Run(context.Background(), agentx.RunRequest{Input: "reject"})
	var typed *agentx.Error
	if !errors.As(err, &typed) ||
		typed.Code != agentx.CodeExecutionFailed ||
		!errors.Is(err, errRejected) {
		exit(fmt.Errorf("unexpected typed error: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		exit(err)
	}
	fmt.Printf("%s display-safe=%q\n", typed.Code, typed.Message)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
