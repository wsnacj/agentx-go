package main

import (
	"context"
	"fmt"
	"os"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/construction"
)

func main() {
	client, err := newClient(context.Background(), os.TempDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "construction conformance",
		SessionID: "construction-conformance",
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
	fmt.Println(result.Reply)
}

func newClient(ctx context.Context, workspaceRoot string) (*agentx.Client, error) {
	return construction.New(ctx, construction.Config{
		WorkspaceRoot: workspaceRoot,
		ModelProfile:  "conformance-model",
	}, conformanceHost{})
}

type conformanceHost struct{}

func (conformanceHost) ResolveModel(
	context.Context,
	construction.Config,
) (construction.ModelRuntime, error) {
	return conformanceResource{}, nil
}

func (conformanceHost) NewRunner(
	context.Context,
	construction.Config,
	construction.ModelRuntime,
) (construction.RunnerRuntime, error) {
	return conformanceResource{}, nil
}

func (conformanceHost) NewAdapter(
	context.Context,
	construction.Config,
	construction.RunnerRuntime,
	construction.ModelRuntime,
) (agentx.ExecutionAdapter, error) {
	return conformanceAdapter{}, nil
}

func (conformanceHost) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

type conformanceResource struct{}

func (conformanceResource) Shutdown(context.Context) error {
	return nil
}

type conformanceAdapter struct{}

func (conformanceAdapter) Run(
	_ context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	return &agentx.AdapterRunResult{
		RunID:     "construction-conformance-run",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "agentx-runtime-construction-ok",
	}, nil
}

func (conformanceAdapter) Shutdown(context.Context) error {
	return nil
}

func (conformanceAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
