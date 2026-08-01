package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

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

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
