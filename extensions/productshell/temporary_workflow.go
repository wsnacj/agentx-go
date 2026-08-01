package productshell

import (
	"context"
	"errors"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// TemporaryWorkflowBinding describes one portable workflow data binding.
type TemporaryWorkflowBinding struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// TemporaryWorkflowStep is one generated linear workflow step.
type TemporaryWorkflowStep struct {
	Kind                 string                     `json:"kind,omitempty"`
	Title                string                     `json:"title,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Tool                 string                     `json:"tool,omitempty"`
	Args                 map[string]any             `json:"args,omitempty"`
	Instruction          string                     `json:"instruction,omitempty"`
	Input                string                     `json:"input,omitempty"`
	Context              string                     `json:"context,omitempty"`
	Schema               map[string]any             `json:"schema,omitempty"`
	UsePreviousOutput    bool                       `json:"use_previous_output,omitempty"`
	BindPreviousOutputTo string                     `json:"bind_previous_output_to,omitempty"`
	InputBindings        []TemporaryWorkflowBinding `json:"input_bindings,omitempty"`
	OutputBindings       []TemporaryWorkflowBinding `json:"output_bindings,omitempty"`
}

// TemporaryWorkflowPlan is the typed result produced by a host plan generator.
type TemporaryWorkflowPlan struct {
	Title       string                  `json:"title,omitempty"`
	Description string                  `json:"description,omitempty"`
	Steps       []TemporaryWorkflowStep `json:"steps,omitempty"`
}

// TemporaryWorkflowPlanningMetrics records the deterministic planning result.
type TemporaryWorkflowPlanningMetrics struct {
	Attempted          bool     `json:"attempted,omitempty"`
	Extracted          bool     `json:"extracted,omitempty"`
	Applied            bool     `json:"applied,omitempty"`
	Source             string   `json:"source,omitempty"`
	WorkflowID         string   `json:"workflow_id,omitempty"`
	WorkflowTitle      string   `json:"workflow_title,omitempty"`
	PlanningMode       string   `json:"planning_mode,omitempty"`
	NodeCount          int      `json:"node_count,omitempty"`
	NodeKinds          []string `json:"node_kinds,omitempty"`
	ToolNames          []string `json:"tool_names,omitempty"`
	InputBindingCount  int      `json:"input_binding_count,omitempty"`
	OutputBindingCount int      `json:"output_binding_count,omitempty"`
	VisibleToolCount   int      `json:"visible_tool_count,omitempty"`
	PlannableTools     []string `json:"plannable_tools,omitempty"`
	LLMStepsSupported  bool     `json:"llm_steps_supported,omitempty"`
	SkipReason         string   `json:"skip_reason,omitempty"`
}

// PreparedTemporaryWorkflowPlan carries a generated plan and its validated
// canonical Workflow Spec.
type PreparedTemporaryWorkflowPlan struct {
	Plan       TemporaryWorkflowPlan
	Workflow   *agentxworkflow.Spec
	Applied    bool
	SkipReason string
	Metrics    TemporaryWorkflowPlanningMetrics
}

// TemporaryWorkflowPlanningError reports a failure after the shell has
// committed to workflow-first execution.
type TemporaryWorkflowPlanningError struct {
	Reason  string
	Metrics TemporaryWorkflowPlanningMetrics
	Cause   error
}

func (e *TemporaryWorkflowPlanningError) Error() string {
	if e == nil {
		return "temporary workflow planning failed"
	}
	reason := strings.TrimSpace(e.Reason)
	if reason == "" && e.Cause != nil {
		reason = strings.TrimSpace(e.Cause.Error())
	}
	if reason == "" {
		return "temporary workflow planning failed"
	}
	return "temporary workflow planning failed: " + reason
}

func (e *TemporaryWorkflowPlanningError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// AsTemporaryWorkflowPlanningError resolves a typed planning error through an
// arbitrary wrapped error chain.
func AsTemporaryWorkflowPlanningError(err error) (*TemporaryWorkflowPlanningError, bool) {
	var target *TemporaryWorkflowPlanningError
	if !errors.As(err, &target) {
		return nil, false
	}
	return target, true
}

// NewTemporaryWorkflowPlanningError creates a planning error while preserving
// the original cause identity.
func NewTemporaryWorkflowPlanningError(reason string, metrics TemporaryWorkflowPlanningMetrics, cause error) error {
	metrics = normalizeTemporaryWorkflowPlanningMetrics(metrics, reason)
	return &TemporaryWorkflowPlanningError{
		Reason:  strings.TrimSpace(reason),
		Metrics: metrics,
		Cause:   cause,
	}
}

func normalizeTemporaryWorkflowPlanningMetrics(metrics TemporaryWorkflowPlanningMetrics, reason string) TemporaryWorkflowPlanningMetrics {
	metrics.Attempted = true
	if strings.TrimSpace(metrics.SkipReason) == "" {
		metrics.SkipReason = strings.TrimSpace(reason)
	}
	return metrics
}

// TemporaryWorkflowPlanningInput is the portable planning-stage input. Host
// execution-policy and capability state intentionally stay outside this type.
type TemporaryWorkflowPlanningInput struct {
	Input            Input
	UserMessage      string
	WorkflowSpec     *agentxworkflow.Spec
	HasPackBinding   bool
	VisibleTools     []types.Tool
	LLMTaskTimeoutMs int
}

// TemporaryWorkflowPlanningResult is the portable result consumed by a host
// before it compiles execution-policy and capability state.
type TemporaryWorkflowPlanningResult struct {
	Input        Input
	WorkflowSpec *agentxworkflow.Spec
	Metrics      TemporaryWorkflowPlanningMetrics
}

// TemporaryWorkflowPlanningRuntime is the narrow host/mechanism port used by
// TemporaryWorkflowPlanningPipeline.
type TemporaryWorkflowPlanningRuntime interface {
	ShouldAttemptTemporaryWorkflowPlanning(Input, *agentxworkflow.Spec, bool) bool
	ResolveTemporaryWorkflowPlan(context.Context, Input, string, []types.Tool, int) (*PreparedTemporaryWorkflowPlan, error)
	ApplyTemporaryWorkflowPlan(Input, *PreparedTemporaryWorkflowPlan) Input
}

// TemporaryWorkflowPlanningPipeline owns the fixed planning-stage order while
// leaving product policy and concrete model execution behind host ports.
type TemporaryWorkflowPlanningPipeline struct {
	runtime TemporaryWorkflowPlanningRuntime
}

// NewTemporaryWorkflowPlanningPipeline creates an Experimental planning-stage
// coordinator. A nil runtime produces a pass-through result.
func NewTemporaryWorkflowPlanningPipeline(runtime TemporaryWorkflowPlanningRuntime) *TemporaryWorkflowPlanningPipeline {
	return &TemporaryWorkflowPlanningPipeline{runtime: runtime}
}

// Plan executes attempt, resolve, typed-error projection and apply in a stable
// order. Execution snapshot compilation remains a host continuation.
func (p *TemporaryWorkflowPlanningPipeline) Plan(ctx context.Context, request TemporaryWorkflowPlanningInput) (TemporaryWorkflowPlanningResult, error) {
	out := TemporaryWorkflowPlanningResult{
		Input:        request.Input,
		WorkflowSpec: request.WorkflowSpec,
	}
	if p == nil || p.runtime == nil {
		return out, nil
	}
	rt := p.runtime
	if !rt.ShouldAttemptTemporaryWorkflowPlanning(request.Input, request.WorkflowSpec, request.HasPackBinding) {
		return out, nil
	}
	prepared, err := rt.ResolveTemporaryWorkflowPlan(
		ctx,
		request.Input,
		request.UserMessage,
		request.VisibleTools,
		request.LLMTaskTimeoutMs,
	)
	if err != nil {
		if prepared != nil {
			out.Metrics = prepared.Metrics
		}
		if planningErr, ok := AsTemporaryWorkflowPlanningError(err); ok {
			out.Metrics = planningErr.Metrics
		} else {
			out.Metrics = normalizeTemporaryWorkflowPlanningMetrics(out.Metrics, err.Error())
		}
		return out, err
	}
	if prepared == nil {
		out.Metrics = normalizeTemporaryWorkflowPlanningMetrics(out.Metrics, "planner_returned_nil")
		return out, NewTemporaryWorkflowPlanningError(out.Metrics.SkipReason, out.Metrics, nil)
	}
	out.Metrics = prepared.Metrics
	if !prepared.Applied {
		reason := strings.TrimSpace(prepared.SkipReason)
		if reason == "" {
			reason = "planner_not_applied"
		}
		out.Metrics = normalizeTemporaryWorkflowPlanningMetrics(out.Metrics, reason)
		return out, NewTemporaryWorkflowPlanningError(reason, out.Metrics, nil)
	}
	out.Input = rt.ApplyTemporaryWorkflowPlan(request.Input, prepared)
	out.WorkflowSpec = prepared.Workflow
	return out, nil
}

// ShouldAttemptTemporaryWorkflowPlanning applies the portable explicit-option
// rules, then uses defaultAttempt for host-owned shell heuristics.
func ShouldAttemptTemporaryWorkflowPlanning(input Input, workflowSpec *agentxworkflow.Spec, hasPackBinding bool, defaultAttempt bool) bool {
	if workflowSpec != nil || hasPackBinding {
		return false
	}
	if input.ShellOptions.AutoWorkflowPlanning != nil {
		return *input.ShellOptions.AutoWorkflowPlanning
	}
	value := parseInputBoolOption(input.Options,
		"auto_workflow_planning",
		"autoWorkflowPlanning",
		"workflow_planning_auto",
		"workflowPlanningAuto",
	)
	if value != nil {
		return *value
	}
	return defaultAttempt
}

// ApplyTemporaryWorkflowPlan projects an applied plan onto typed shell input.
func ApplyTemporaryWorkflowPlan(input Input, prepared *PreparedTemporaryWorkflowPlan) Input {
	if prepared == nil || !prepared.Applied || prepared.Workflow == nil {
		return input
	}
	spec := *prepared.Workflow
	input.WorkflowSpec = &spec
	input.RawWorkflowOptIn = true
	return input
}
