package agent_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxtools "github.com/wsnacj/agentx-go/tools"
	agent "github.com/wsnacj/agentx-go/tools/agent"
)

func TestRegisterRoutesTaskAndNormalizesSubagentRun(t *testing.T) {
	reg := agentxtools.NewRegistry()
	var requests []agent.Request
	agent.Register(reg, agent.Options{Backend: agent.BackendFuncs{
		TaskFunc: func(_ context.Context, request agent.Request) (string, error) {
			requests = append(requests, request)
			return `{"task_id":"task-1"}`, nil
		},
		SubagentFunc: func(_ context.Context, request agent.Request) (string, error) {
			requests = append(requests, request)
			return `{"action":"run"}`, nil
		},
		AgentStepFunc: func(_ context.Context, request agent.Request) (string, error) {
			requests = append(requests, request)
			return `{"result":"ok"}`, nil
		},
	}})

	if _, err := reg.Execute(context.Background(), llm.FunctionCall{Name: agent.TasksRunName, Arguments: `{"seed_message":"inspect"}`}); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Execute(context.Background(), llm.FunctionCall{Name: agent.SubagentsName, Arguments: `{"action":"run","message":"inspect"}`}); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d", len(requests))
	}
	if requests[0].Name != agent.TasksRunName || requests[1].Action != "run" {
		t.Fatalf("unexpected requests: %#v", requests)
	}
	if got := requests[1].Arguments["seed_message"]; got != "inspect" {
		t.Fatalf("normalized seed_message = %#v", got)
	}
	if got := requests[1].Arguments["seed_role"]; got != "user" {
		t.Fatalf("normalized seed_role = %#v", got)
	}
}

func TestSubagentsTypedErrorsAndFanoutCount(t *testing.T) {
	reg := agentxtools.NewRegistry()
	agent.Register(reg, agent.Options{Backend: agent.BackendFuncs{
		TaskFunc:      func(context.Context, agent.Request) (string, error) { return `{}`, nil },
		SubagentFunc:  func(context.Context, agent.Request) (string, error) { return `{}`, nil },
		AgentStepFunc: func(context.Context, agent.Request) (string, error) { return `{}`, nil },
	}})

	for _, call := range []llm.FunctionCall{
		{Name: agent.SubagentsName, Arguments: `{}`},
		{Name: agent.SubagentsName, Arguments: `{"action":"unknown"}`},
		{Name: agent.SubagentsName, Arguments: `{"action":"fanout","expected_count":2,"items":[{"message":"one"}]}`},
	} {
		_, err := reg.Execute(context.Background(), call)
		if err == nil {
			t.Fatalf("expected typed error for %s", call.Arguments)
		}
		var argumentError *agentxtoolerrors.ToolArgumentError
		if !errors.As(err, &argumentError) {
			t.Fatalf("error type = %T: %v", err, err)
		}
	}
}

func TestCancellationAndFencedJSONReachBackend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reg := agentxtools.NewRegistry()
	agent.Register(reg, agent.Options{Enabled: []string{agent.TasksWaitName}, Backend: agent.BackendFuncs{
		TaskFunc: func(ctx context.Context, request agent.Request) (string, error) {
			if request.Arguments["task_id"] != "task-1" {
				t.Fatalf("arguments = %#v", request.Arguments)
			}
			return "", ctx.Err()
		},
	}})
	_, err := reg.Execute(ctx, llm.FunctionCall{Name: agent.TasksWaitName, Arguments: "```json\n{\"task_id\":\"task-1\"}\n```"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if got := reg.Definitions(); len(got) != 1 || got[0].Function.Name != agent.TasksWaitName {
		t.Fatalf("definitions = %#v", got)
	}
}

func TestDefinitionsAreStable(t *testing.T) {
	want := []string{
		agent.TasksSpawnName, agent.TasksWaitName, agent.TasksRunName,
		agent.TasksCancelName, agent.TasksReplayName, agent.TasksCollectName,
		agent.TasksDeadletterListName, agent.SubagentsName, agent.AgentStepName,
	}
	defs := agent.Definitions()
	got := make([]string, 0, len(defs))
	for _, definition := range defs {
		got = append(got, definition.Function.Name)
		if definition.Type != "function" || definition.Function.Parameters == nil {
			t.Fatalf("invalid definition: %#v", definition)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %#v, want %#v", got, want)
	}
}
