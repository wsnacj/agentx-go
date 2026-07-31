package construction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/construction"
)

func TestExternalHostConstructsRunsAndShutsDownClient(t *testing.T) {
	host := externalHost{}
	client, err := construction.New(context.Background(), construction.Config{
		WorkspaceRoot: t.TempDir(),
		ModelProfile:  "external-model",
	}, host)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "hello",
		SessionID: "external-session",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "completed" ||
		result.Reply != "external-ok" ||
		result.SessionID != "external-session" {
		t.Fatalf("Run() result = %#v", result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestExternalHostErrorPreservesTypedAndOriginalIdentity(t *testing.T) {
	cause := errors.New("external host unavailable")
	host := externalHost{resolveErr: cause}
	client, err := construction.New(context.Background(), construction.Config{
		WorkspaceRoot: t.TempDir(),
		ModelProfile:  "external-model",
	}, host)
	if client != nil {
		t.Fatal("New() returned a client")
	}
	if !errors.Is(err, cause) ||
		!errors.Is(err, &agentx.Error{Code: agentx.CodeExecutionFailed}) {
		t.Fatalf("New() error = %v", err)
	}
}

type externalHost struct {
	resolveErr error
}

func (h externalHost) ResolveModel(
	context.Context,
	construction.Config,
) (construction.ModelRuntime, error) {
	if h.resolveErr != nil {
		return nil, h.resolveErr
	}
	return externalResource{}, nil
}

func (externalHost) NewRunner(
	context.Context,
	construction.Config,
	construction.ModelRuntime,
) (construction.RunnerRuntime, error) {
	return externalResource{}, nil
}

func (externalHost) NewAdapter(
	context.Context,
	construction.Config,
	construction.RunnerRuntime,
	construction.ModelRuntime,
) (agentx.ExecutionAdapter, error) {
	return externalAdapter{}, nil
}

func (externalHost) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

type externalResource struct{}

func (externalResource) Shutdown(context.Context) error {
	return nil
}

type externalAdapter struct{}

func (externalAdapter) Run(
	_ context.Context,
	request agentx.AdapterRunRequest,
) (*agentx.AdapterRunResult, error) {
	return &agentx.AdapterRunResult{
		RunID:     "external-run",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "external-ok",
	}, nil
}

func (externalAdapter) Shutdown(context.Context) error {
	return nil
}

func (externalAdapter) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
