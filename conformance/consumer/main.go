// conformance/consumer 是 W1 的 external-style canonical import 验收程序。
//
// 它是独立 Go module，不使用 local replace，不访问网络或 secret。
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

var (
	errProbeFailed = errors.New("conformance probe failed")
	errProbeClosed = errors.New("conformance adapter closed")
)

type conformanceAdapter struct {
	closed atomic.Bool
}

func (a *conformanceAdapter) Run(
	_ context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	if a.closed.Load() {
		return nil, errProbeClosed
	}
	if request.Input == "fail" {
		return nil, errProbeFailed
	}
	return &agentx.AdapterRunResult{
		RunID:     "run-conformance",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "conformance-ok",
	}, nil
}

func (a *conformanceAdapter) Shutdown(context.Context) error {
	a.closed.Store(true)
	return nil
}

func (*conformanceAdapter) ClassifyError(err error) agentx.ErrorCode {
	if errors.Is(err, errProbeClosed) {
		return agentx.CodeClientClosed
	}
	return agentx.CodeExecutionFailed
}

func runProbe() (string, error) {
	client, err := agentx.New(agentx.Config{Adapter: &conformanceAdapter{}})
	if err != nil {
		return "", err
	}

	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "run",
		SessionID: "conformance-session",
	})
	if err != nil {
		return "", err
	}
	if result.RunID != "run-conformance" ||
		result.SessionID != "conformance-session" ||
		result.Status != "completed" ||
		result.Reply != "conformance-ok" ||
		len(result.Evidence) != 2 {
		return "", fmt.Errorf("unexpected success result: %#v", result)
	}

	_, runErr := client.Run(context.Background(), agentx.RunRequest{Input: "fail"})
	var typed *agentx.Error
	if !errors.As(runErr, &typed) ||
		typed.Code != agentx.CodeExecutionFailed ||
		!errors.Is(runErr, errProbeFailed) {
		return "", fmt.Errorf("unexpected typed error: %v", runErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		return "", err
	}
	if err := client.Shutdown(ctx); err != nil {
		return "", err
	}
	_, closedErr := client.Run(context.Background(), agentx.RunRequest{Input: "after close"})
	if !errors.Is(closedErr, &agentx.Error{Code: agentx.CodeClientClosed}) {
		return "", fmt.Errorf("unexpected closed error: %v", closedErr)
	}
	return result.Status + " " + result.SessionID, nil
}

func main() {
	output, err := runAllProbes(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(output)
}
