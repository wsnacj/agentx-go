package nodeexec

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type contextKey string

func TestCoordinatorPrefersRichOutcomeBindsContextAndInvokesExactlyOnce(t *testing.T) {
	basic := &captureBasic{}
	node := &captureNode{}
	outcome := &captureOutcome{outcome: Outcome{
		Output:                " raw output ",
		FinalStatus:           " incomplete ",
		StopReason:            " raw reason ",
		ExecutionContractID:   " contract-1 ",
		ExecutionContractDiff: []string{" raw field "},
		Termination:           &Termination{Kind: " raw termination "},
		DelegatedExecution: &DelegatedExecution{
			Driver: " raw driver ",
			Rounds: []DelegatedRound{{
				NodeExecID: " child-1 ",
				Round:      1,
			}},
		},
		ChildNodeExecutions: []NodeExecutionProjection{{
			NodeExecID:       " child-1 ",
			ParentNodeExecID: " parent-1 ",
			NodeID:           " child ",
			Status:           " completed ",
			ChildNodeExecutions: []NodeExecutionProjection{{
				NodeExecID: " grandchild-1 ",
			}},
		}},
	}}
	coordinator := New(Dependencies{
		Basic:   basic,
		Node:    node,
		Outcome: outcome,
		BindContext: func(ctx context.Context, executionID string, nodeID string) context.Context {
			if executionID != "nodeexec-1" || nodeID != "node-1" {
				t.Fatalf("binding identity = %q/%q", executionID, nodeID)
			}
			return context.WithValue(ctx, contextKey("bound"), "yes")
		},
	})
	request := Request{
		NodeExecutionID: "nodeexec-1",
		NodeID:          "node-1",
		Call:            Call{Name: "echo", Arguments: `{"value":"ok"}`},
	}
	got, err := coordinator.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if !reflect.DeepEqual(got, outcome.outcome) {
		t.Fatalf("outcome = %#v, want %#v", got, outcome.outcome)
	}
	if basic.calls != 0 || node.calls != 0 || outcome.calls != 1 {
		t.Fatalf("calls basic/node/outcome = %d/%d/%d", basic.calls, node.calls, outcome.calls)
	}
	if outcome.contextValue != "yes" || !reflect.DeepEqual(outcome.request, request) {
		t.Fatalf("captured context/request = %q/%#v", outcome.contextValue, outcome.request)
	}
}

func TestCoordinatorPreservesOutputAndErrorAcrossFallbacks(t *testing.T) {
	sentinel := errors.New("raw failure")
	rich := &captureOutcome{
		outcome: Outcome{Output: "rich output", FinalStatus: " raw status "},
		err:     sentinel,
	}
	got, err := New(Dependencies{
		Basic:   &captureBasic{},
		Outcome: rich,
	}).Execute(context.Background(), Request{})
	if !errors.Is(err, sentinel) ||
		got.Output != "rich output" ||
		got.FinalStatus != " raw status " ||
		rich.calls != 1 {
		t.Fatalf("outcome result = %#v, %v, calls=%d", got, err, rich.calls)
	}

	node := &captureNode{output: "node output", err: sentinel}
	got, err = New(Dependencies{
		Basic: &captureBasic{output: "basic output"},
		Node:  node,
	}).Execute(context.Background(), Request{})
	if !errors.Is(err, sentinel) || got.Output != "node output" || node.calls != 1 {
		t.Fatalf("node result = %#v, %v, calls=%d", got, err, node.calls)
	}

	basic := &captureBasic{output: "basic output", err: sentinel}
	got, err = New(Dependencies{Basic: basic}).Execute(context.Background(), Request{
		Call: Call{Name: "echo", Arguments: " raw "},
	})
	if !errors.Is(err, sentinel) || got.Output != "basic output" || basic.calls != 1 {
		t.Fatalf("basic result = %#v, %v, calls=%d", got, err, basic.calls)
	}
	if basic.call.Name != "echo" || basic.call.Arguments != " raw " {
		t.Fatalf("basic call = %#v", basic.call)
	}
}

func TestCoordinatorPropagatesCancelledContextWithoutOverridingResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	basic := &captureBasic{output: "dependency decides"}
	got, err := New(Dependencies{Basic: basic}).Execute(ctx, Request{})
	if err != nil || got.Output != "dependency decides" {
		t.Fatalf("Execute() = %#v, %v", got, err)
	}
	if !errors.Is(basic.contextErr, context.Canceled) {
		t.Fatalf("executor context error = %v, want context.Canceled", basic.contextErr)
	}
}

func TestCoordinatorKeepsOriginalContextWhenBinderReturnsNil(t *testing.T) {
	original := context.WithValue(context.Background(), contextKey("bound"), "original")
	basic := &captureBasic{}
	coordinator := New(Dependencies{
		Basic: basic,
		BindContext: func(context.Context, string, string) context.Context {
			return nil
		},
	})
	if _, err := coordinator.Execute(original, Request{}); err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if basic.contextValue != "original" {
		t.Fatalf("context value = %q, want original", basic.contextValue)
	}
}

func TestPortableOutcomeJSONPreservesExistingProjectionShape(t *testing.T) {
	payload, err := json.Marshal(Outcome{
		Output: "ready",
		DelegatedExecution: &DelegatedExecution{
			Rounds: []DelegatedRound{{NodeExecID: "round-1"}},
		},
		ChildNodeExecutions: []NodeExecutionProjection{{
			NodeExecID:       "child-1",
			ParentNodeExecID: "parent-1",
		}},
	})
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	want := `{"output":"ready","delegated_execution":{"rounds":[{"node_exec_id":"round-1"}]},"child_node_executions":[{"node_exec_id":"child-1","parent_node_exec_id":"parent-1"}]}`
	if string(payload) != want {
		t.Fatalf("JSON = %s, want %s", payload, want)
	}
}

func TestCoordinatorFailsClosedWithoutBasicExecutor(t *testing.T) {
	_, err := New(Dependencies{}).Execute(context.Background(), Request{})
	if got, want := err.Error(), "workflow nodeexec: basic executor is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

type captureBasic struct {
	calls        int
	call         Call
	output       string
	err          error
	contextValue string
	contextErr   error
}

func (c *captureBasic) Execute(ctx context.Context, call Call) (string, error) {
	c.calls++
	c.call = call
	c.contextValue, _ = ctx.Value(contextKey("bound")).(string)
	c.contextErr = ctx.Err()
	return c.output, c.err
}

type captureNode struct {
	calls  int
	output string
	err    error
}

func (c *captureNode) ExecuteNode(context.Context, Request) (string, error) {
	c.calls++
	return c.output, c.err
}

type captureOutcome struct {
	calls        int
	request      Request
	outcome      Outcome
	err          error
	contextValue string
}

func (c *captureOutcome) ExecuteNodeWithOutcome(ctx context.Context, request Request) (Outcome, error) {
	c.calls++
	c.request = request
	c.contextValue, _ = ctx.Value(contextKey("bound")).(string)
	return c.outcome, c.err
}
