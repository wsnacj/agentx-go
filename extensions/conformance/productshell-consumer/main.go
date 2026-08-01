package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/extensions/productshell"
	"github.com/wsnacj/agentx-go/runtime/cases"
	"github.com/wsnacj/agentx-go/runtime/workflow"
)

type validator struct{}

func (validator) ValidateSpec(spec workflow.Spec) error {
	if strings.TrimSpace(spec.ID) == "" {
		return fmt.Errorf("workflow id is required")
	}
	return nil
}

type deterministicTemporaryWorkflowGenerator struct {
	requests []productshell.TemporaryWorkflowPlanGenerationRequest
}

func (g *deterministicTemporaryWorkflowGenerator) GenerateTemporaryWorkflowPlan(
	_ context.Context,
	request productshell.TemporaryWorkflowPlanGenerationRequest,
) (productshell.TemporaryWorkflowPlan, error) {
	g.requests = append(g.requests, request)
	return productshell.TemporaryWorkflowPlan{
		Title:       "Lookup AgentX",
		Description: "Deterministic external-consumer fixture",
		Steps: []productshell.TemporaryWorkflowStep{{
			Kind:  "tool",
			Title: "Lookup",
			Tool:  "LOOKUP",
			Args:  map[string]any{"limit": 1},
			InputBindings: []productshell.TemporaryWorkflowBinding{{
				From: "session.input.request",
				To:   "args.query",
			}},
			OutputBindings: []productshell.TemporaryWorkflowBinding{{
				From: "result.summary",
				To:   "state.summary",
			}},
		}},
	}, nil
}

type deterministicTemporaryWorkflowHost struct {
	planner *productshell.TemporaryWorkflowPlanner
	calls   []string
}

func (h *deterministicTemporaryWorkflowHost) ShouldAttemptTemporaryWorkflowPlanning(
	input productshell.Input,
	spec *workflow.Spec,
	hasPackBinding bool,
) bool {
	h.calls = append(h.calls, "should")
	return productshell.ShouldAttemptTemporaryWorkflowPlanning(input, spec, hasPackBinding, false)
}

func (h *deterministicTemporaryWorkflowHost) ResolveTemporaryWorkflowPlan(
	ctx context.Context,
	input productshell.Input,
	userMessage string,
	visibleTools []llm.Tool,
	timeoutMs int,
) (*productshell.PreparedTemporaryWorkflowPlan, error) {
	h.calls = append(h.calls, "resolve")
	planningTools, allowLLMSteps := externalPlanningTools(visibleTools)
	return h.planner.ResolveTemporaryWorkflowPlan(ctx, productshell.TemporaryWorkflowPlannerInput{
		Input:            input,
		UserMessage:      userMessage,
		PlanningTools:    planningTools,
		AllowLLMSteps:    allowLLMSteps,
		VisibleToolCount: len(visibleTools),
		LLMTaskTimeoutMs: timeoutMs,
	})
}

func (h *deterministicTemporaryWorkflowHost) ApplyTemporaryWorkflowPlan(
	input productshell.Input,
	prepared *productshell.PreparedTemporaryWorkflowPlan,
) productshell.Input {
	h.calls = append(h.calls, "apply")
	return productshell.ApplyTemporaryWorkflowPlan(input, prepared)
}

func externalPlanningTools(visibleTools []llm.Tool) ([]productshell.TemporaryWorkflowPlanningTool, bool) {
	planningTools := make([]productshell.TemporaryWorkflowPlanningTool, 0, len(visibleTools))
	allowLLMSteps := false
	for _, tool := range visibleTools {
		name := strings.ToLower(strings.TrimSpace(tool.Function.Name))
		if name == "llm_task" {
			allowLLMSteps = true
			continue
		}
		// Tool visibility and deny/allow policy remain Host-owned. This fixture
		// deliberately exposes only one harmless deterministic tool.
		if name != "lookup" {
			continue
		}
		argumentKeys := make([]string, 0)
		if properties, ok := tool.Function.Parameters["properties"].(map[string]any); ok {
			for key := range properties {
				if key = strings.TrimSpace(key); key != "" {
					argumentKeys = append(argumentKeys, key)
				}
			}
		}
		sort.Strings(argumentKeys)
		planningTools = append(planningTools, productshell.TemporaryWorkflowPlanningTool{
			Name:         tool.Function.Name,
			Description:  strings.TrimSpace(tool.Function.Description),
			ArgumentKeys: argumentKeys,
		})
	}
	sort.Slice(planningTools, func(i, j int) bool {
		return strings.ToLower(planningTools[i].Name) < strings.ToLower(planningTools[j].Name)
	})
	return planningTools, allowLLMSteps
}

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

type deterministicHost struct {
	coordinator *pack.Coordinator
	registry    *pack.MemoryRegistry
	calls       []string
}

func newDeterministicHost() (*deterministicHost, error) {
	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		return nil, err
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		return nil, err
	}
	definition := pack.Definition{
		Manifest: pack.Manifest{
			ID:                 "portable-research",
			Version:            "1.0.0",
			Domain:             "research",
			RouteHints:         []string{"研究资料"},
			SupportedCaseTypes: []string{"research.lookup"},
			DefaultWorkflow:    "collect-v1",
		},
		Workflows: []workflow.Spec{{
			ID:         "collect-v1",
			Pack:       "portable-research",
			CaseTypes:  []string{"research.lookup"},
			RouteHints: []string{"收集资料"},
			EntryNode:  "collect",
			Nodes: []workflow.NodeSpec{{
				ID:     "collect",
				Kind:   workflow.NodeTool,
				Config: map[string]any{"tool_name": "collect"},
			}},
		}},
		Tools: []pack.SemanticTool{{Name: "collect", RuntimeTool: "host_collect"}},
	}
	if err := registry.Register(definition); err != nil {
		return nil, err
	}
	return &deterministicHost{coordinator: coordinator, registry: registry}, nil
}

func (h *deterministicHost) record(name string) {
	h.calls = append(h.calls, name)
}

func (h *deterministicHost) runtime() productshell.PreparationRuntime {
	return productshell.PreparationRuntimeFuncs{
		ApplyInputCaseFn: func(input productshell.Input) productshell.Input {
			h.record("input")
			return productshell.ApplyInputCase(input)
		},
		ResolveShellBindingFn: func(ctx context.Context, _ string, _ productshell.Input) (*productshell.PreparedShellBinding, error) {
			h.record("shell")
			return nil, ctx.Err()
		},
		ResolveCommandDispatchFn: func(ctx context.Context, _ productshell.Input) (*productshell.PreparedCommandDispatch, error) {
			h.record("command")
			return nil, ctx.Err()
		},
		ParseRequestedSkillsFn: func(input productshell.Input) ([]string, string) {
			h.record("skills")
			return productshell.ParseRequestedSkills(input)
		},
		ShouldAttemptPackSelectionFn: func(productshell.Input) bool {
			h.record("select-policy")
			return true
		},
		ResolvePackSelectionFn: func(ctx context.Context, _ productshell.Input, message string) (*productshell.PreparedPackSelection, error) {
			h.record("select-pack")
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			selection, matched := pack.SelectBinding(h.registry, message, pack.SelectOptions{})
			if !matched {
				return nil, fmt.Errorf("pack selection did not match: %#v", selection)
			}
			return &productshell.PreparedPackSelection{
				Selection: selection,
				Binding: productshell.WorkflowBinding{
					PackID: selection.Selected.PackID, CaseType: selection.Selected.CaseType,
					WorkflowID: selection.Selected.WorkflowID,
				},
				Matched: true,
				Applied: true,
			}, nil
		},
		ApplyPackSelectionFn: func(input productshell.Input, prepared *productshell.PreparedPackSelection) productshell.Input {
			h.record("apply-pack")
			return productshell.ApplyPackSelection(input, prepared)
		},
		ShouldAttemptCaseBindingFn: func(productshell.Input) bool {
			h.record("case-policy")
			return true
		},
		ResolveCandidateCaseBindingFn: func(input productshell.Input, _ *productshell.PreparedPackSelection) (pack.Binding, bool, error) {
			h.record("case-binding")
			return h.coordinator.ResolveBinding(h.registry, input.PackID, input.CaseType, input.PackWorkflow)
		},
		ResolveCaseBindingDraftFn: func(ctx context.Context, input productshell.Input, _ string, binding pack.Binding, _ int) (*productshell.PreparedCaseBinding, error) {
			h.record("case-draft")
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			merged := map[string]any{}
			if input.Case != nil {
				for key, value := range input.Case.Inputs {
					merged[key] = value
				}
			}
			return &productshell.PreparedCaseBinding{
				Binding: binding,
				Merged:  merged,
				Applied: true,
				Metrics: productshell.CaseBindingMetrics{
					Attempted: true, Matched: true, Applied: true, Source: "deterministic-host",
					PackID: binding.PackID, CaseType: binding.CaseType, WorkflowID: binding.WorkflowID,
				},
			}, nil
		},
		MergeCaseBindingMetricsFn: func(_ productshell.CaseBindingMetrics, prepared *productshell.PreparedCaseBinding) productshell.CaseBindingMetrics {
			h.record("case-metrics")
			return prepared.Metrics
		},
		ApplyCaseBindingFn: func(input productshell.Input, prepared *productshell.PreparedCaseBinding) productshell.Input {
			h.record("apply-case-binding")
			return productshell.ApplyCaseBinding(input, prepared)
		},
		ResolveWorkflowFn: func(input productshell.Input) (productshell.ResolvedWorkflow, error) {
			h.record("workflow")
			binding, ok, err := h.coordinator.ResolveBinding(h.registry, input.PackID, input.CaseType, input.PackWorkflow)
			if err != nil || !ok {
				return productshell.ResolvedWorkflow{}, err
			}
			spec := binding.Workflow
			return productshell.ResolvedWorkflow{Spec: &spec, PackBinding: &binding}, nil
		},
		ResolveEffectiveCaseFn: func(sessionID string, input productshell.Input, message string, _ *workflow.Spec, _ *pack.Binding) (*cases.Case, error) {
			h.record("effective-case")
			value := cases.Clone(input.Case)
			if value == nil {
				value = &cases.Case{}
			}
			value.SessionID = sessionID
			value.Intent = message
			value.Source = "external-consumer"
			value.Status = "prepared"
			return value, nil
		},
		ValidateEffectiveCaseFn: func(binding *pack.Binding, value *cases.Case) error {
			h.record("validate-case")
			if binding == nil || value == nil {
				return fmt.Errorf("resolved binding and case are required")
			}
			if value.PackID != binding.PackID || value.Type != binding.CaseType || value.WorkflowID != binding.WorkflowID {
				return fmt.Errorf("case binding does not match workflow binding")
			}
			return binding.ValidateCaseInput(value.Inputs)
		},
		ApplyEffectiveCaseFn: func(input productshell.Input, value cases.Case) productshell.Input {
			h.record("apply-effective-case")
			return productshell.ApplyEffectiveCase(input, value)
		},
		PackSelectionMetricsFromFn: func(prepared *productshell.PreparedPackSelection) productshell.PackSelectionMetrics {
			h.record("pack-metrics")
			return productshell.PackSelectionMetricsFromPrepared(prepared)
		},
	}
}

func run() (string, error) {
	host, err := newDeterministicHost()
	if err != nil {
		return "", err
	}
	result, err := productshell.NewPreparationPipeline(host.runtime()).Prepare(
		context.Background(),
		"session-001",
		productshell.Input{
			UserMessage:        "[skill:portable-review] 请研究资料并收集资料",
			RequestedCaseID:    "case-001",
			RequestedCaseInput: map[string]any{"topic": "AgentX"},
		},
	)
	if err != nil {
		return "", err
	}
	if result.Workflow == nil || result.Workflow.Spec == nil || result.EffectiveCase == nil {
		return "", fmt.Errorf("preparation did not produce workflow and case")
	}
	if !slices.Equal(result.RequestedSkills, []string{"portable-review"}) {
		return "", fmt.Errorf("unexpected requested skills: %v", result.RequestedSkills)
	}
	wantStages := []string{
		"input", "shell", "command", "skills", "select-policy", "select-pack", "apply-pack",
		"case-policy", "case-binding", "case-draft", "case-metrics", "apply-case-binding",
		"workflow", "effective-case", "validate-case", "apply-effective-case", "pack-metrics",
	}
	if !slices.Equal(host.calls, wantStages) {
		return "", fmt.Errorf("unexpected stage order: %v", host.calls)
	}
	topic, _ := result.EffectiveCase.Inputs["topic"].(string)
	return fmt.Sprintf(
		"agentx-productshell-ok:%s:%s:%s:%s:%s:%s",
		result.EffectiveCase.PackID,
		result.EffectiveCase.Type,
		result.Workflow.Spec.ID,
		result.RequestedSkills[0],
		result.EffectiveCase.ID,
		topic,
	), nil
}

func runTemporaryWorkflowPlanning() (string, error) {
	generator := &deterministicTemporaryWorkflowGenerator{}
	planner := productshell.NewTemporaryWorkflowPlanner(productshell.TemporaryWorkflowPlannerConfig{
		Generator:               generator,
		Validator:               validator{},
		WorkflowIDFactory:       func() string { return "temp_workflow_external_consumer" },
		NormalizeToolName:       func(value string) string { return strings.ToLower(strings.TrimSpace(value)) },
		DefaultLLMTaskTimeoutMs: 30_000,
	})
	host := &deterministicTemporaryWorkflowHost{planner: planner}
	autoPlanning := true
	result, err := productshell.NewTemporaryWorkflowPlanningPipeline(host).Plan(
		context.Background(),
		productshell.TemporaryWorkflowPlanningInput{
			Input: productshell.Input{
				UserMessage:  "lookup AgentX",
				SessionInput: map[string]any{"request": "lookup AgentX"},
				ShellOptions: productshell.InputShellOptions{
					AutoWorkflowPlanning: &autoPlanning,
				},
			},
			UserMessage: "lookup AgentX",
			VisibleTools: []llm.Tool{{
				Type: "function",
				Function: llm.Function{
					Name:        "lookup",
					Description: "Lookup a topic",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{"type": "string"},
							"limit": map[string]any{"type": "integer"},
						},
					},
				},
			}},
			LLMTaskTimeoutMs: 25_000,
		},
	)
	if err != nil {
		return "", err
	}
	if !slices.Equal(host.calls, []string{"should", "resolve", "apply"}) {
		return "", fmt.Errorf("unexpected planning stage order: %v", host.calls)
	}
	if len(generator.requests) != 1 {
		return "", fmt.Errorf("unexpected generation attempts: %d", len(generator.requests))
	}
	generationRequest := generator.requests[0]
	if !generationRequest.Strict || generationRequest.TimeoutMs != 25_000 ||
		!strings.Contains(generationRequest.Instruction, "Return 1 to 5 linear steps only.") ||
		!strings.Contains(generationRequest.Input, `"available_tools":[{"arg_keys":["limit","query"]`) ||
		len(generationRequest.Schema) == 0 {
		return "", fmt.Errorf("unexpected generation request: %#v", generationRequest)
	}
	if !result.Applied || result.WorkflowSpec == nil || result.Input.WorkflowSpec == nil ||
		!result.Input.RawWorkflowOptIn {
		return "", fmt.Errorf("temporary workflow was not applied: %#v", result)
	}
	if result.WorkflowSpec.ID != "temp_workflow_external_consumer" ||
		result.WorkflowSpec.EntryNode != "step_01" || len(result.WorkflowSpec.Nodes) != 1 {
		return "", fmt.Errorf("unexpected workflow: %#v", result.WorkflowSpec)
	}
	metrics := result.Metrics
	if !metrics.Attempted || !metrics.Extracted || !metrics.Applied ||
		metrics.NodeCount != 1 || metrics.InputBindingCount != 1 || metrics.OutputBindingCount != 1 ||
		!slices.Equal(metrics.ToolNames, []string{"lookup"}) ||
		!slices.Equal(metrics.PlannableTools, []string{"lookup"}) {
		return "", fmt.Errorf("unexpected planning metrics: %#v", metrics)
	}
	return fmt.Sprintf(
		"agentx-productshell-planning-ok:%s:%d:%s:%t",
		result.WorkflowSpec.ID,
		metrics.NodeCount,
		metrics.ToolNames[0],
		result.Applied,
	), nil
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
	planningResult, err := runTemporaryWorkflowPlanning()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(planningResult)
}
