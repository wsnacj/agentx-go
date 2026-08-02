package scheduler_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxtools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/scheduler"
)

func TestExternalConsumerRoutesAllActions(t *testing.T) {
	called := make([]string, 0, 5)
	handle := func(_ context.Context, request scheduler.Request) (string, error) {
		called = append(called, request.Action)
		return `{"action":"` + request.Action + `"}`, nil
	}
	registry := agentxtools.NewRegistry()
	scheduler.Register(registry, scheduler.BackendFuncs{
		AddFunc: handle, ListFunc: handle, StatusFunc: handle, RunFunc: handle, RemoveFunc: handle,
	})
	for _, action := range []string{"add", "list", "status", "run", "remove"} {
		result, err := registry.Execute(context.Background(), toolcontract.Call{Name: scheduler.Name, Arguments: `{"action":"` + action + `"}`})
		if err != nil || result == "" {
			t.Fatalf("action=%s result=%q err=%v", action, result, err)
		}
	}
	if !reflect.DeepEqual(called, []string{"add", "list", "status", "run", "remove"}) {
		t.Fatalf("called=%v", called)
	}
}

func TestTypedErrorsAndCancellation(t *testing.T) {
	backend := scheduler.BackendFuncs{ListFunc: func(ctx context.Context, _ scheduler.Request) (string, error) { return "", ctx.Err() }}
	registry := agentxtools.NewRegistry()
	scheduler.Register(registry, backend)
	_, err := registry.Execute(context.Background(), toolcontract.Call{Name: scheduler.Name, Arguments: `{}`})
	argumentError, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok || argumentError.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("argument error=%T %v", err, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = registry.Execute(ctx, toolcontract.Call{Name: scheduler.Name, Arguments: `{"action":"list"}`})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestDefinitionIsClosedOverActionContract(t *testing.T) {
	definition := scheduler.Definition()
	if definition.Function.Name != scheduler.Name {
		t.Fatalf("definition=%#v", definition)
	}
	required, ok := definition.Function.Parameters["required"].([]string)
	if !ok || !reflect.DeepEqual(required, []string{"action"}) {
		t.Fatalf("required=%#v", definition.Function.Parameters["required"])
	}
}
