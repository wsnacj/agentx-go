package productshell

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type temporaryPlanningRuntimeStub struct {
	calls    []string
	should   bool
	prepared *PreparedTemporaryWorkflowPlan
	err      error
}

func (s *temporaryPlanningRuntimeStub) ShouldAttemptTemporaryWorkflowPlanning(Input, *agentxworkflow.Spec, bool) bool {
	s.calls = append(s.calls, "should")
	return s.should
}

func (s *temporaryPlanningRuntimeStub) ResolveTemporaryWorkflowPlan(context.Context, Input, string, []types.Tool, int) (*PreparedTemporaryWorkflowPlan, error) {
	s.calls = append(s.calls, "resolve")
	return s.prepared, s.err
}

func (s *temporaryPlanningRuntimeStub) ApplyTemporaryWorkflowPlan(input Input, prepared *PreparedTemporaryWorkflowPlan) Input {
	s.calls = append(s.calls, "apply")
	return ApplyTemporaryWorkflowPlan(input, prepared)
}

func TestTemporaryWorkflowPlanningPipelineOrder(t *testing.T) {
	spec := agentxworkflow.Spec{ID: "temp_workflow_fixed"}
	runtime := &temporaryPlanningRuntimeStub{
		should: true,
		prepared: &PreparedTemporaryWorkflowPlan{
			Applied:  true,
			Workflow: &spec,
			Metrics:  TemporaryWorkflowPlanningMetrics{Attempted: true, Applied: true},
		},
	}
	result, err := NewTemporaryWorkflowPlanningPipeline(runtime).Plan(context.Background(), TemporaryWorkflowPlanningInput{
		Input:        Input{UserMessage: "plan"},
		UserMessage:  "plan",
		VisibleTools: []types.Tool{{Type: "function", Function: types.Function{Name: "lookup"}}},
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if got, want := strings.Join(runtime.calls, ","), "should,resolve,apply"; got != want {
		t.Fatalf("calls = %q, want %q", got, want)
	}
	if !result.Applied || result.WorkflowSpec != &spec || result.Input.WorkflowSpec == nil || !result.Input.RawWorkflowOptIn {
		t.Fatalf("result = %#v", result)
	}
}

func TestTemporaryWorkflowPlanningPipelinePreservesTypedErrorMetrics(t *testing.T) {
	sentinel := errors.New("generator failed")
	metrics := TemporaryWorkflowPlanningMetrics{Source: "llm_task", SkipReason: "generation_failed"}
	typed := NewTemporaryWorkflowPlanningError("generation_failed", metrics, sentinel)
	runtime := &temporaryPlanningRuntimeStub{should: true, prepared: &PreparedTemporaryWorkflowPlan{}, err: typed}
	result, err := NewTemporaryWorkflowPlanningPipeline(runtime).Plan(context.Background(), TemporaryWorkflowPlanningInput{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Plan() error = %v, want sentinel", err)
	}
	if _, ok := AsTemporaryWorkflowPlanningError(err); !ok {
		t.Fatalf("Plan() error = %T, want typed planning error", err)
	}
	if result.Metrics.SkipReason != "generation_failed" || !result.Metrics.Attempted {
		t.Fatalf("metrics = %#v", result.Metrics)
	}
}

func TestTemporaryWorkflowPlanningErrorContract(t *testing.T) {
	var nilError *TemporaryWorkflowPlanningError
	if got := nilError.Error(); got != "temporary workflow planning failed" {
		t.Fatalf("nil Error() = %q", got)
	}
	sentinel := errors.New("sentinel")
	err := NewTemporaryWorkflowPlanningError("", TemporaryWorkflowPlanningMetrics{}, sentinel)
	if got, want := err.Error(), "temporary workflow planning failed: sentinel"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is(%v, sentinel) = false", err)
	}
	typed, ok := AsTemporaryWorkflowPlanningError(err)
	if !ok || typed.Reason != "" || !typed.Metrics.Attempted || typed.Metrics.SkipReason != "" {
		t.Fatalf("typed = %#v, ok=%v", typed, ok)
	}
}

func TestTemporaryWorkflowPlanningPipelineNilAndNotApplied(t *testing.T) {
	tests := []struct {
		name     string
		prepared *PreparedTemporaryWorkflowPlan
		want     string
	}{
		{name: "nil", want: "planner_returned_nil"},
		{name: "not applied", prepared: &PreparedTemporaryWorkflowPlan{}, want: "planner_not_applied"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := &temporaryPlanningRuntimeStub{should: true, prepared: tt.prepared}
			result, err := NewTemporaryWorkflowPlanningPipeline(runtime).Plan(context.Background(), TemporaryWorkflowPlanningInput{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Plan() error = %v, want %q", err, tt.want)
			}
			if result.Metrics.SkipReason != tt.want {
				t.Fatalf("metrics = %#v", result.Metrics)
			}
		})
	}
}

func TestTemporaryWorkflowContractsJSON(t *testing.T) {
	value := TemporaryWorkflowPlan{
		Title: "Research",
		Steps: []TemporaryWorkflowStep{{
			Kind: "tool", Title: "Lookup", Tool: "lookup",
			InputBindings: []TemporaryWorkflowBinding{{From: "session.input.request", To: "args.query"}},
		}},
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	want := `{"title":"Research","steps":[{"kind":"tool","title":"Lookup","tool":"lookup","input_bindings":[{"from":"session.input.request","to":"args.query"}]}]}`
	if string(payload) != want {
		t.Fatalf("json = %s, want %s", payload, want)
	}
}

type temporaryGeneratorStub struct {
	requests []TemporaryWorkflowPlanGenerationRequest
	results  []TemporaryWorkflowPlan
	errors   []error
}

func (s *temporaryGeneratorStub) GenerateTemporaryWorkflowPlan(_ context.Context, request TemporaryWorkflowPlanGenerationRequest) (TemporaryWorkflowPlan, error) {
	s.requests = append(s.requests, request)
	index := len(s.requests) - 1
	var result TemporaryWorkflowPlan
	if index < len(s.results) {
		result = s.results[index]
	}
	if index < len(s.errors) {
		return result, s.errors[index]
	}
	return result, nil
}

type temporaryValidatorStub struct {
	calls int
	err   error
}

func (s *temporaryValidatorStub) ValidateSpec(agentxworkflow.Spec) error {
	s.calls++
	return s.err
}

func newTemporaryPlannerForTest(generator TemporaryWorkflowPlanGenerator, validator agentxworkflow.Validator) *TemporaryWorkflowPlanner {
	return NewTemporaryWorkflowPlanner(TemporaryWorkflowPlannerConfig{
		Generator:         generator,
		Validator:         validator,
		WorkflowIDFactory: func() string { return "temp_workflow_00000000-0000-0000-0000-000000000001" },
		NormalizeToolName: func(value string) string { return strings.ToLower(strings.TrimSpace(value)) },
	})
}

func TestTemporaryWorkflowPlannerBuildsLinearWorkflow(t *testing.T) {
	generator := &temporaryGeneratorStub{results: []TemporaryWorkflowPlan{{
		Title: "Lookup And Summarize",
		Steps: []TemporaryWorkflowStep{
			{Kind: "tool", Title: "Lookup", Tool: "LOOKUP", Args: map[string]any{"query": "alpha"}},
			{Kind: "llm", Title: "Summarize", Instruction: "Summarize", UsePreviousOutput: true},
		},
	}}}
	validator := &temporaryValidatorStub{}
	prepared, err := newTemporaryPlannerForTest(generator, validator).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input:            Input{UserMessage: "lookup alpha"},
		PlanningTools:    []TemporaryWorkflowPlanningTool{{Name: "lookup", Description: "Lookup", ArgumentKeys: []string{"query"}}},
		AllowLLMSteps:    true,
		VisibleToolCount: 2,
	})
	if err != nil {
		t.Fatalf("ResolveTemporaryWorkflowPlan() error = %v", err)
	}
	if prepared == nil || prepared.Workflow == nil || !prepared.Applied {
		t.Fatalf("prepared = %#v", prepared)
	}
	if prepared.Workflow.ID != "temp_workflow_00000000-0000-0000-0000-000000000001" || prepared.Workflow.EntryNode != "step_01" {
		t.Fatalf("workflow = %#v", prepared.Workflow)
	}
	if got := prepared.Workflow.Nodes[1].Inputs; len(got) != 1 || got[0].From != "node.step_01.output" || got[0].To != "args.input" {
		t.Fatalf("bindings = %#v", got)
	}
	if validator.calls != 1 || prepared.Metrics.NodeCount != 2 || prepared.Metrics.VisibleToolCount != 2 {
		t.Fatalf("validator=%d metrics=%#v", validator.calls, prepared.Metrics)
	}
	if len(generator.requests) != 1 || !generator.requests[0].Strict || generator.requests[0].TimeoutMs != 45_000 {
		t.Fatalf("request = %#v", generator.requests)
	}
	if !strings.Contains(generator.requests[0].Instruction, "Return 1 to 5 linear steps only.") ||
		!strings.Contains(generator.requests[0].Input, `"available_tools":[{"arg_keys":["query"],"description":"Lookup","name":"lookup"}]`) {
		t.Fatalf("generation request = %#v", generator.requests[0])
	}
}

func TestTemporaryWorkflowPlannerNormalizesToolOrderAndArgumentKeys(t *testing.T) {
	generator := &temporaryGeneratorStub{results: []TemporaryWorkflowPlan{{
		Title: "Order",
		Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "A", Tool: "A"}},
	}}}
	prepared, err := newTemporaryPlannerForTest(generator, &temporaryValidatorStub{}).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input: Input{UserMessage: "order"},
		PlanningTools: []TemporaryWorkflowPlanningTool{
			{Name: " b ", Description: " B ", ArgumentKeys: []string{"z", "a"}},
			{Name: "A", Description: "A", ArgumentKeys: []string{"y", "x"}},
			{Name: "a", Description: "duplicate", ArgumentKeys: []string{"d"}},
		},
		VisibleToolCount: 4,
	})
	if err != nil || prepared == nil {
		t.Fatalf("prepared=%#v error=%v", prepared, err)
	}
	if got, want := prepared.Metrics.PlannableTools, []string{"a", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("plannable tools = %#v, want %#v", got, want)
	}
	var payload struct {
		AvailableTools []struct {
			Name    string   `json:"name"`
			ArgKeys []string `json:"arg_keys"`
		} `json:"available_tools"`
	}
	if err := json.Unmarshal([]byte(generator.requests[0].Input), &payload); err != nil {
		t.Fatalf("decode planner input: %v", err)
	}
	if got := payload.AvailableTools; len(got) != 3 || got[0].Name != "a" || got[1].Name != "a" || got[2].Name != "b" || !reflect.DeepEqual(got[2].ArgKeys, []string{"a", "z"}) {
		t.Fatalf("available tools = %#v", got)
	} else {
		for _, item := range got[:2] {
			if !reflect.DeepEqual(item.ArgKeys, []string{"d"}) && !reflect.DeepEqual(item.ArgKeys, []string{"x", "y"}) {
				t.Fatalf("normalized duplicate args = %#v", got[:2])
			}
		}
	}
}

func TestTemporaryWorkflowPlannerRetriesTimeoutOnce(t *testing.T) {
	generator := &temporaryGeneratorStub{
		results: []TemporaryWorkflowPlan{{}, {Title: "Retry", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup"}}}},
		errors:  []error{context.DeadlineExceeded, nil},
	}
	prepared, err := newTemporaryPlannerForTest(generator, &temporaryValidatorStub{}).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input: Input{UserMessage: "lookup"}, PlanningTools: []TemporaryWorkflowPlanningTool{{Name: "lookup"}}, LLMTaskTimeoutMs: 50_000,
	})
	if err != nil || prepared == nil || !prepared.Applied {
		t.Fatalf("prepared=%#v error=%v", prepared, err)
	}
	if got := []int{generator.requests[0].TimeoutMs, generator.requests[1].TimeoutMs}; !reflect.DeepEqual(got, []int{50_000, 90_000}) {
		t.Fatalf("timeouts = %#v", got)
	}
}

func TestTemporaryWorkflowPlannerDoesNotRetryCancellation(t *testing.T) {
	generator := &temporaryGeneratorStub{errors: []error{context.Canceled}}
	_, err := newTemporaryPlannerForTest(generator, &temporaryValidatorStub{}).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input: Input{UserMessage: "lookup"}, PlanningTools: []TemporaryWorkflowPlanningTool{{Name: "lookup"}},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if len(generator.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(generator.requests))
	}
}

func TestTemporaryWorkflowPlannerCombinedRetryCeiling(t *testing.T) {
	generator := &temporaryGeneratorStub{
		results: []TemporaryWorkflowPlan{
			{},
			{Title: "Bad", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup", InputBindings: []TemporaryWorkflowBinding{{From: "session.input.missing", To: "args.query"}}}}},
			{},
			{Title: "Good", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup", InputBindings: []TemporaryWorkflowBinding{{From: "session.input.request", To: "args.query"}}}}},
		},
		errors: []error{context.DeadlineExceeded, nil, context.DeadlineExceeded, nil},
	}
	prepared, err := newTemporaryPlannerForTest(generator, &temporaryValidatorStub{}).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input:         Input{UserMessage: "lookup", SessionInput: map[string]any{"request": "lookup"}},
		PlanningTools: []TemporaryWorkflowPlanningTool{{Name: "lookup"}},
	})
	if err != nil || prepared == nil || !prepared.Applied {
		t.Fatalf("prepared=%#v error=%v", prepared, err)
	}
	if len(generator.requests) != 4 {
		t.Fatalf("requests = %d, want 4", len(generator.requests))
	}
	for index, request := range generator.requests {
		want := 45_000
		if index%2 == 1 {
			want = 90_000
		}
		if request.TimeoutMs != want {
			t.Fatalf("request %d timeout = %d, want %d", index, request.TimeoutMs, want)
		}
	}
}

func TestTemporaryWorkflowPlannerPreservesValidatorErrorAndIDTiming(t *testing.T) {
	sentinel := errors.New("validator sentinel")
	generator := &temporaryGeneratorStub{results: []TemporaryWorkflowPlan{{
		Title: "Validate", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup"}},
	}}}
	validator := &temporaryValidatorStub{err: sentinel}
	idCalls := 0
	planner := NewTemporaryWorkflowPlanner(TemporaryWorkflowPlannerConfig{
		Generator: generator,
		Validator: validator,
		WorkflowIDFactory: func() string {
			idCalls++
			return "temp_workflow_fixed"
		},
		NormalizeToolName: strings.TrimSpace,
	})
	_, err := planner.ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input: Input{UserMessage: "lookup"}, PlanningTools: []TemporaryWorkflowPlanningTool{{Name: "lookup"}},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want validator sentinel", err)
	}
	if idCalls != 1 || validator.calls != 1 || len(generator.requests) != 1 {
		t.Fatalf("idCalls=%d validator=%d generator=%d", idCalls, validator.calls, len(generator.requests))
	}
}

func TestTemporaryWorkflowPlannerRetriesUnavailableSessionBinding(t *testing.T) {
	generator := &temporaryGeneratorStub{results: []TemporaryWorkflowPlan{
		{Title: "Bad", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup", InputBindings: []TemporaryWorkflowBinding{{From: "session.input.missing", To: "args.query"}}}}},
		{Title: "Good", Steps: []TemporaryWorkflowStep{{Kind: "tool", Title: "Lookup", Tool: "lookup", InputBindings: []TemporaryWorkflowBinding{{From: "session.input.request", To: "args.query"}}}}},
	}}
	prepared, err := newTemporaryPlannerForTest(generator, &temporaryValidatorStub{}).ResolveTemporaryWorkflowPlan(context.Background(), TemporaryWorkflowPlannerInput{
		Input:         Input{UserMessage: "lookup", SessionInput: map[string]any{"request": "lookup"}},
		PlanningTools: []TemporaryWorkflowPlanningTool{{Name: "lookup"}},
	})
	if err != nil || prepared == nil || prepared.Workflow == nil {
		t.Fatalf("prepared=%#v error=%v", prepared, err)
	}
	if len(generator.requests) != 2 || !strings.Contains(generator.requests[1].Input, "planning_feedback") {
		t.Fatalf("requests = %#v", generator.requests)
	}
}

func TestApplyTemporaryWorkflowPlanUsesShallowWorkflowCopy(t *testing.T) {
	spec := &agentxworkflow.Spec{ID: "temp_workflow_copy", Nodes: []agentxworkflow.NodeSpec{{ID: "step_01"}}}
	result := ApplyTemporaryWorkflowPlan(Input{}, &PreparedTemporaryWorkflowPlan{Applied: true, Workflow: spec})
	if result.WorkflowSpec == spec || result.WorkflowSpec == nil || !result.RawWorkflowOptIn {
		t.Fatalf("result = %#v", result)
	}
	spec.Nodes[0].Title = "mutated"
	if result.WorkflowSpec.Nodes[0].Title != "mutated" {
		t.Fatal("expected legacy shallow slice aliasing")
	}
}
