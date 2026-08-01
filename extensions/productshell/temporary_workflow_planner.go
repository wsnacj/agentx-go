package productshell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const defaultTemporaryWorkflowPlanningTimeoutMs = 45_000

// TemporaryWorkflowPlanningTool is the policy-filtered tool description that
// the host permits the portable planner to use.
type TemporaryWorkflowPlanningTool struct {
	Name         string
	Description  string
	ArgumentKeys []string
}

// TemporaryWorkflowPlanGenerationRequest is the structured-generation request
// passed to a host-owned model adapter.
type TemporaryWorkflowPlanGenerationRequest struct {
	Instruction string
	Input       string
	Schema      map[string]any
	Strict      bool
	TimeoutMs   int
}

// TemporaryWorkflowPlanGenerator owns concrete model/provider execution. It
// must preserve its returned error identity.
type TemporaryWorkflowPlanGenerator interface {
	GenerateTemporaryWorkflowPlan(context.Context, TemporaryWorkflowPlanGenerationRequest) (TemporaryWorkflowPlan, error)
}

// TemporaryWorkflowPlannerInput contains the already policy-filtered planning
// surface supplied by a host.
type TemporaryWorkflowPlannerInput struct {
	Input            Input
	UserMessage      string
	PlanningTools    []TemporaryWorkflowPlanningTool
	AllowLLMSteps    bool
	VisibleToolCount int
	LLMTaskTimeoutMs int
}

// TemporaryWorkflowPlannerConfig supplies the narrow host ports used by the
// portable planning mechanism.
type TemporaryWorkflowPlannerConfig struct {
	Generator               TemporaryWorkflowPlanGenerator
	Validator               agentxworkflow.Validator
	WorkflowIDFactory       func() string
	NormalizeToolName       func(string) string
	DefaultLLMTaskTimeoutMs int
}

// TemporaryWorkflowPlanner owns deterministic prompt/schema construction,
// bounded retry, binding lowering, Workflow Spec construction and metrics.
type TemporaryWorkflowPlanner struct {
	config TemporaryWorkflowPlannerConfig
}

// NewTemporaryWorkflowPlanner creates an Experimental portable planner.
func NewTemporaryWorkflowPlanner(config TemporaryWorkflowPlannerConfig) *TemporaryWorkflowPlanner {
	return &TemporaryWorkflowPlanner{config: config}
}

type temporaryWorkflowPlanningContext struct {
	SessionInputs    map[string]any
	SessionInputSet  map[string]struct{}
	SessionInputList []string
}

type temporaryWorkflowUnavailableSessionInputError struct {
	StepIndex      int
	Source         string
	AvailablePaths []string
}

func (e *temporaryWorkflowUnavailableSessionInputError) Error() string {
	if e == nil {
		return ""
	}
	message := fmt.Sprintf("temporary workflow step %d references unavailable session input %q", e.StepIndex, e.Source)
	if len(e.AvailablePaths) == 0 {
		return message
	}
	return fmt.Sprintf("%s (available: %s)", message, strings.Join(e.AvailablePaths, ", "))
}

// ResolveTemporaryWorkflowPlan generates and validates one temporary Workflow
// while preserving the legacy error, retry and metrics order.
func (p *TemporaryWorkflowPlanner) ResolveTemporaryWorkflowPlan(ctx context.Context, request TemporaryWorkflowPlannerInput) (*PreparedTemporaryWorkflowPlan, error) {
	metrics := TemporaryWorkflowPlanningMetrics{
		Attempted:        true,
		Source:           "llm_task",
		VisibleToolCount: request.VisibleToolCount,
	}
	prepared := &PreparedTemporaryWorkflowPlan{Metrics: metrics}
	if p == nil || p.config.Generator == nil {
		prepared.SkipReason = "chat_unavailable"
		prepared.Metrics.SkipReason = prepared.SkipReason
		return prepared, NewTemporaryWorkflowPlanningError(prepared.SkipReason, prepared.Metrics, nil)
	}
	userMessage := firstNonEmptyTemporaryWorkflowString(request.UserMessage, request.Input.UserMessage)
	if userMessage == "" {
		prepared.SkipReason = "empty_user_message"
		prepared.Metrics.SkipReason = prepared.SkipReason
		return prepared, NewTemporaryWorkflowPlanningError(prepared.SkipReason, prepared.Metrics, nil)
	}
	planningTools := normalizeTemporaryWorkflowPlanningTools(request.PlanningTools, p.normalizeToolName)
	if len(planningTools) == 0 && !request.AllowLLMSteps {
		prepared.SkipReason = "no_plannable_tools"
		prepared.Metrics.SkipReason = prepared.SkipReason
		return prepared, NewTemporaryWorkflowPlanningError(prepared.SkipReason, prepared.Metrics, nil)
	}
	prepared.Metrics.LLMStepsSupported = request.AllowLLMSteps
	prepared.Metrics.PlannableTools = temporaryWorkflowPlanningToolNames(planningTools)
	planningContext := buildTemporaryWorkflowPlanningContext(request.Input, userMessage)

	plan, err := p.extractTemporaryWorkflowPlan(ctx, userMessage, planningTools, request.AllowLLMSteps, request.LLMTaskTimeoutMs, planningContext, "")
	if err != nil {
		prepared.SkipReason = err.Error()
		prepared.Metrics.SkipReason = prepared.SkipReason
		return prepared, NewTemporaryWorkflowPlanningError(prepared.SkipReason, prepared.Metrics, err)
	}
	prepared.Plan = plan
	prepared.Metrics.Extracted = len(plan.Steps) > 0

	spec, err := p.buildTemporaryWorkflowSpec(plan, planningTools, request.AllowLLMSteps, planningContext.SessionInputSet)
	if err != nil {
		if feedback, ok := temporaryWorkflowPlanningRetryFeedback(err); ok {
			retryPlan, retryErr := p.extractTemporaryWorkflowPlan(ctx, userMessage, planningTools, request.AllowLLMSteps, request.LLMTaskTimeoutMs, planningContext, feedback)
			if retryErr == nil {
				plan = retryPlan
				prepared.Plan = plan
				prepared.Metrics.Extracted = len(plan.Steps) > 0
				spec, err = p.buildTemporaryWorkflowSpec(plan, planningTools, request.AllowLLMSteps, planningContext.SessionInputSet)
			}
		}
	}
	if err != nil {
		prepared.SkipReason = err.Error()
		prepared.Metrics.SkipReason = prepared.SkipReason
		return prepared, NewTemporaryWorkflowPlanningError(prepared.SkipReason, prepared.Metrics, err)
	}
	prepared.Workflow = &spec
	prepared.Applied = true
	prepared.Metrics.Applied = true
	prepared.Metrics.WorkflowID = strings.TrimSpace(spec.ID)
	prepared.Metrics.WorkflowTitle = strings.TrimSpace(spec.Title)
	prepared.Metrics.PlanningMode = string(spec.PlanningMode)
	prepared.Metrics.NodeCount = len(spec.Nodes)
	prepared.Metrics.NodeKinds = temporaryWorkflowNodeKinds(spec)
	prepared.Metrics.ToolNames = temporaryWorkflowToolNames(spec)
	prepared.Metrics.InputBindingCount, prepared.Metrics.OutputBindingCount = temporaryWorkflowBindingCounts(spec)
	return prepared, nil
}

func (p *TemporaryWorkflowPlanner) extractTemporaryWorkflowPlan(
	ctx context.Context,
	userMessage string,
	planningTools []TemporaryWorkflowPlanningTool,
	allowLLMSteps bool,
	timeoutMs int,
	planningContext temporaryWorkflowPlanningContext,
	feedback string,
) (TemporaryWorkflowPlan, error) {
	defaultTimeoutMs := p.runtimeTimeoutMs(timeoutMs)
	runAttempt := func(attemptTimeoutMs int) (TemporaryWorkflowPlan, error) {
		return p.config.Generator.GenerateTemporaryWorkflowPlan(ctx, TemporaryWorkflowPlanGenerationRequest{
			Instruction: buildTemporaryWorkflowPlanningInstruction(allowLLMSteps, planningContext, feedback),
			Input:       buildTemporaryWorkflowPlanningInput(userMessage, planningTools, allowLLMSteps, planningContext, feedback),
			Schema:      temporaryWorkflowPlanSchema(),
			Strict:      true,
			TimeoutMs:   attemptTimeoutMs,
		})
	}
	plan, err := runAttempt(defaultTimeoutMs)
	if err == nil {
		return plan, nil
	}
	if !isTemporaryWorkflowPlanningTimeout(err) {
		return TemporaryWorkflowPlan{}, err
	}
	retryTimeoutMs := relaxedTemporaryWorkflowPlanningTimeoutMs(defaultTimeoutMs)
	if retryTimeoutMs <= defaultTimeoutMs {
		return TemporaryWorkflowPlan{}, err
	}
	return runAttempt(retryTimeoutMs)
}

func (p *TemporaryWorkflowPlanner) runtimeTimeoutMs(timeoutMs int) int {
	if timeoutMs > 0 {
		return timeoutMs
	}
	if p != nil && p.config.DefaultLLMTaskTimeoutMs > 0 {
		return p.config.DefaultLLMTaskTimeoutMs
	}
	return defaultTemporaryWorkflowPlanningTimeoutMs
}

func (p *TemporaryWorkflowPlanner) normalizeToolName(value string) string {
	if p != nil && p.config.NormalizeToolName != nil {
		return p.config.NormalizeToolName(value)
	}
	return strings.TrimSpace(value)
}

func relaxedTemporaryWorkflowPlanningTimeoutMs(timeoutMs int) int {
	if timeoutMs <= 0 {
		return 0
	}
	const maxTimeoutMs = 90_000
	if timeoutMs >= maxTimeoutMs {
		return timeoutMs
	}
	next := timeoutMs * 2
	if next < timeoutMs {
		return timeoutMs
	}
	if next > maxTimeoutMs {
		next = maxTimeoutMs
	}
	return next
}

func isTemporaryWorkflowPlanningTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "context deadline exceeded")
}

func buildTemporaryWorkflowPlanningInstruction(allowLLMSteps bool, planningContext temporaryWorkflowPlanningContext, feedback string) string {
	lines := []string{
		"Plan a short temporary workflow for the user request.",
		"Return 1 to 5 linear steps only.",
		"Use kind=tool only when a listed runtime tool should be called directly.",
		"Use tool names exactly as listed.",
		"Do not invent URLs, paths, selectors, IDs, or arguments that are not directly implied by the user request.",
		"When a step needs the previous step output, set use_previous_output=true.",
		"When binding previous output, bind_previous_output_to should usually be args.input unless the tool needs a different field.",
		"Use input_bindings when a step needs values from session.input.*, case.input.*, state.*, or previous.output / previous.result.*.",
		"Use output_bindings to store step output/result fields into workflow state so later steps can reuse them.",
		"Only bind session.input.* paths that are explicitly listed in available_session_input_paths. Do not invent new session.input fields.",
	}
	if len(planningContext.SessionInputList) == 0 {
		lines = append(lines, "Do not use session.input.* bindings because no session input values are available.")
	} else {
		lines = append(lines, fmt.Sprintf("Available session.input paths for this request: %s.", strings.Join(planningContext.SessionInputList, ", ")))
		if _, ok := planningContext.SessionInputSet["session.input.last_assistant_message"]; ok {
			lines = append(lines, "For follow-up questions about the immediately previous answer, prefer session.input.last_assistant_message when it contains the needed evidence.")
		}
		if _, ok := planningContext.SessionInputSet["session.input.last_tool_message"]; ok {
			lines = append(lines, "For follow-up questions about the immediately previous tool result, prefer session.input.last_tool_message when it contains the needed evidence.")
		}
	}
	if strings.TrimSpace(feedback) != "" {
		lines = append(lines,
			"Previous planning attempt failed.",
			feedback,
			"Return a corrected workflow that only uses available session.input bindings.",
		)
	}
	if allowLLMSteps {
		lines = append(lines,
			"Use kind=llm only for focused synthesis, extraction, or summarization steps.",
			"llm steps must include a concrete instruction.",
		)
	} else {
		lines = append(lines, "Do not use kind=llm in this plan.")
	}
	return strings.Join(lines, "\n")
}

func buildTemporaryWorkflowPlanningInput(
	userMessage string,
	planningTools []TemporaryWorkflowPlanningTool,
	allowLLMSteps bool,
	planningContext temporaryWorkflowPlanningContext,
	feedback string,
) string {
	payload := map[string]any{
		"user_request":                  strings.TrimSpace(userMessage),
		"llm_steps_supported":           allowLLMSteps,
		"available_tools":               temporaryWorkflowPlanningToolDescriptors(planningTools),
		"available_session_input":       temporaryWorkflowPlanningSessionInputPreview(planningContext.SessionInputs),
		"available_session_input_paths": append([]string(nil), planningContext.SessionInputList...),
	}
	if strings.TrimSpace(feedback) != "" {
		payload["planning_feedback"] = strings.TrimSpace(feedback)
	}
	return temporaryWorkflowMustJSON(payload)
}

func temporaryWorkflowPlanSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"required":             []any{"title", "steps"},
		"additionalProperties": false,
		"properties": map[string]any{
			"title":       map[string]any{"type": "string", "minLength": 1, "maxLength": 120},
			"description": map[string]any{"type": "string", "maxLength": 400},
			"steps": map[string]any{
				"type": "array", "minItems": 1, "maxItems": 5,
				"items": map[string]any{
					"type": "object", "required": []any{"kind", "title"}, "additionalProperties": false,
					"properties": map[string]any{
						"kind":                    map[string]any{"type": "string", "enum": []any{"tool", "llm"}},
						"title":                   map[string]any{"type": "string", "minLength": 1, "maxLength": 120},
						"description":             map[string]any{"type": "string", "maxLength": 240},
						"tool":                    map[string]any{"type": "string", "maxLength": 120},
						"args":                    map[string]any{"type": "object", "additionalProperties": true},
						"instruction":             map[string]any{"type": "string", "maxLength": 4000},
						"input":                   map[string]any{"type": "string", "maxLength": 4000},
						"context":                 map[string]any{"type": "string", "maxLength": 4000},
						"schema":                  map[string]any{"type": "object", "additionalProperties": true},
						"use_previous_output":     map[string]any{"type": "boolean"},
						"bind_previous_output_to": map[string]any{"type": "string", "maxLength": 120},
						"input_bindings": temporaryWorkflowBindingArraySchema(
							[]any{"previous.output", "previous.result.summary", "session.input.request", "state.evidence_path"},
							[]any{"args.input", "args.context.evidence_path", "args.payload.target_url"},
						),
						"output_bindings": temporaryWorkflowBindingArraySchema(
							[]any{"output", "result.summary", "result.path"},
							[]any{"state.summary", "state.evidence_path"},
						),
					},
				},
			},
		},
	}
}

func temporaryWorkflowBindingArraySchema(exampleFrom []any, exampleTo []any) map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object", "required": []any{"from", "to"}, "additionalProperties": false,
			"properties": map[string]any{
				"from": map[string]any{"type": "string", "minLength": 1, "maxLength": 160, "examples": exampleFrom},
				"to":   map[string]any{"type": "string", "minLength": 1, "maxLength": 160, "examples": exampleTo},
			},
		},
	}
}

func (p *TemporaryWorkflowPlanner) buildTemporaryWorkflowSpec(
	plan TemporaryWorkflowPlan,
	planningTools []TemporaryWorkflowPlanningTool,
	allowLLMSteps bool,
	availableSessionInputs map[string]struct{},
) (agentxworkflow.Spec, error) {
	title := strings.TrimSpace(plan.Title)
	if title == "" {
		return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow plan title is required")
	}
	if len(plan.Steps) == 0 {
		return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow plan requires at least one step")
	}
	allowedTools := make(map[string]TemporaryWorkflowPlanningTool, len(planningTools))
	for _, tool := range planningTools {
		name := p.normalizeToolName(tool.Name)
		if name != "" {
			allowedTools[name] = tool
		}
	}
	nodes := make([]agentxworkflow.NodeSpec, 0, len(plan.Steps))
	edges := make([]agentxworkflow.EdgeSpec, 0, temporaryWorkflowMaxInt(len(plan.Steps)-1, 0))
	prevNodeID := ""
	for idx, rawStep := range plan.Steps {
		step, err := p.normalizeTemporaryWorkflowStep(rawStep, idx, allowedTools, allowLLMSteps)
		if err != nil {
			return agentxworkflow.Spec{}, err
		}
		nodeID := fmt.Sprintf("step_%02d", idx+1)
		node := agentxworkflow.NodeSpec{ID: nodeID, Title: step.Title, Description: step.Description, Config: map[string]any{}}
		switch step.Kind {
		case "tool":
			node.Kind = agentxworkflow.NodeTool
			node.Config["tool"] = step.Tool
			if len(step.Args) > 0 {
				node.Config["args"] = cloneTemporaryWorkflowMap(step.Args)
			}
		case "llm":
			node.Kind = agentxworkflow.NodeLLM
			node.Config["instruction"] = step.Instruction
			if strings.TrimSpace(step.Input) != "" {
				node.Config["input"] = strings.TrimSpace(step.Input)
			}
			if strings.TrimSpace(step.Context) != "" {
				node.Config["context"] = strings.TrimSpace(step.Context)
			}
			if len(step.Schema) > 0 {
				node.Config["schema"] = cloneTemporaryWorkflowMap(step.Schema)
				node.Config["strict"] = true
			}
		default:
			return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow step %d has unsupported kind %q", idx+1, step.Kind)
		}
		if step.UsePreviousOutput {
			if prevNodeID == "" {
				return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow step %d cannot use previous output without a previous step", idx+1)
			}
			target := strings.TrimSpace(step.BindPreviousOutputTo)
			if target == "" {
				target = "args.input"
			} else {
				target = normalizeTemporaryWorkflowInputBindingTarget(target)
			}
			node.Inputs = append(node.Inputs, agentxworkflow.BindingSpec{From: "node." + prevNodeID + ".output", To: target})
		}
		inputBindings, err := compileTemporaryWorkflowInputBindings(step.InputBindings, prevNodeID, idx, availableSessionInputs)
		if err != nil {
			return agentxworkflow.Spec{}, err
		}
		node.Inputs = append(node.Inputs, inputBindings...)
		outputBindings, err := compileTemporaryWorkflowOutputBindings(step.OutputBindings, idx)
		if err != nil {
			return agentxworkflow.Spec{}, err
		}
		node.Outputs = append(node.Outputs, outputBindings...)
		nodes = append(nodes, node)
		if prevNodeID != "" {
			edges = append(edges, agentxworkflow.EdgeSpec{From: prevNodeID, To: nodeID, On: "success"})
		}
		prevNodeID = nodeID
	}
	if p.config.WorkflowIDFactory == nil {
		return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow id factory is required")
	}
	spec := agentxworkflow.Spec{
		ID:           strings.TrimSpace(p.config.WorkflowIDFactory()),
		Title:        title,
		Description:  strings.TrimSpace(plan.Description),
		Version:      "temp.v1",
		PlanningMode: agentxworkflow.PlanningOpen,
		EntryNode:    nodes[0].ID,
		Nodes:        nodes,
		Edges:        edges,
	}
	if p.config.Validator == nil {
		return agentxworkflow.Spec{}, fmt.Errorf("temporary workflow validator is required")
	}
	if err := p.config.Validator.ValidateSpec(spec); err != nil {
		return agentxworkflow.Spec{}, err
	}
	return spec, nil
}

func (p *TemporaryWorkflowPlanner) normalizeTemporaryWorkflowStep(
	raw TemporaryWorkflowStep,
	index int,
	allowedTools map[string]TemporaryWorkflowPlanningTool,
	allowLLMSteps bool,
) (TemporaryWorkflowStep, error) {
	step := TemporaryWorkflowStep{
		Kind: strings.ToLower(strings.TrimSpace(raw.Kind)), Title: strings.TrimSpace(raw.Title),
		Description: strings.TrimSpace(raw.Description), Tool: p.normalizeToolName(raw.Tool),
		Args: cloneTemporaryWorkflowMap(raw.Args), Instruction: strings.TrimSpace(raw.Instruction),
		Input: strings.TrimSpace(raw.Input), Context: strings.TrimSpace(raw.Context),
		Schema: cloneTemporaryWorkflowMap(raw.Schema), UsePreviousOutput: raw.UsePreviousOutput,
		BindPreviousOutputTo: strings.TrimSpace(raw.BindPreviousOutputTo),
		InputBindings:        normalizeTemporaryWorkflowBindings(raw.InputBindings),
		OutputBindings:       normalizeTemporaryWorkflowBindings(raw.OutputBindings),
	}
	if step.Title == "" {
		return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d title is required", index+1)
	}
	switch step.Kind {
	case "tool":
		if step.Tool == "" {
			return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d tool is required", index+1)
		}
		if _, ok := allowedTools[step.Tool]; !ok {
			return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d references unavailable tool %q", index+1, step.Tool)
		}
	case "llm":
		if !allowLLMSteps {
			return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d cannot use llm when llm_task is not visible", index+1)
		}
		if step.Instruction == "" {
			return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d instruction is required", index+1)
		}
	default:
		return TemporaryWorkflowStep{}, fmt.Errorf("temporary workflow step %d kind %q is not supported", index+1, step.Kind)
	}
	return step, nil
}

func normalizeTemporaryWorkflowBindings(raw []TemporaryWorkflowBinding) []TemporaryWorkflowBinding {
	if len(raw) == 0 {
		return nil
	}
	out := make([]TemporaryWorkflowBinding, 0, len(raw))
	for _, binding := range raw {
		from, to := strings.TrimSpace(binding.From), strings.TrimSpace(binding.To)
		if from != "" && to != "" {
			out = append(out, TemporaryWorkflowBinding{From: from, To: to})
		}
	}
	return out
}

func compileTemporaryWorkflowInputBindings(bindings []TemporaryWorkflowBinding, prevNodeID string, stepIndex int, availableSessionInputs map[string]struct{}) ([]agentxworkflow.BindingSpec, error) {
	if len(bindings) == 0 {
		return nil, nil
	}
	out := make([]agentxworkflow.BindingSpec, 0, len(bindings))
	for _, binding := range bindings {
		source, err := resolveTemporaryWorkflowInputBindingSource(binding.From, prevNodeID, stepIndex, availableSessionInputs)
		if err != nil {
			return nil, err
		}
		out = append(out, agentxworkflow.BindingSpec{From: source, To: normalizeTemporaryWorkflowInputBindingTarget(binding.To)})
	}
	return out, nil
}

func normalizeTemporaryWorkflowInputBindingTarget(target string) string {
	target = strings.TrimSpace(target)
	switch {
	case target == "", target == "args", strings.HasPrefix(target, "args."), strings.HasPrefix(target, "state."),
		strings.HasPrefix(target, "session."), strings.HasPrefix(target, "case."), strings.HasPrefix(target, "node."):
		return target
	default:
		return "args." + target
	}
}

func resolveTemporaryWorkflowInputBindingSource(source, prevNodeID string, stepIndex int, availableSessionInputs map[string]struct{}) (string, error) {
	source = strings.TrimSpace(source)
	switch {
	case strings.HasPrefix(source, "previous."):
		if prevNodeID == "" {
			return "", fmt.Errorf("temporary workflow step %d cannot bind %q without a previous step", stepIndex+1, source)
		}
		return "node." + prevNodeID + "." + strings.TrimPrefix(source, "previous."), nil
	case strings.HasPrefix(source, "session.input."):
		if len(availableSessionInputs) > 0 {
			if _, ok := availableSessionInputs[source]; !ok {
				available := make([]string, 0, len(availableSessionInputs))
				for item := range availableSessionInputs {
					available = append(available, item)
				}
				sort.Strings(available)
				return "", &temporaryWorkflowUnavailableSessionInputError{StepIndex: stepIndex + 1, Source: source, AvailablePaths: available}
			}
		}
		return source, nil
	case strings.HasPrefix(source, "case.input."), strings.HasPrefix(source, "state."):
		return source, nil
	default:
		return "", fmt.Errorf("temporary workflow step %d input binding source %q is not supported", stepIndex+1, source)
	}
}

func compileTemporaryWorkflowOutputBindings(bindings []TemporaryWorkflowBinding, stepIndex int) ([]agentxworkflow.BindingSpec, error) {
	if len(bindings) == 0 {
		return nil, nil
	}
	out := make([]agentxworkflow.BindingSpec, 0, len(bindings))
	for _, binding := range bindings {
		source, target := strings.TrimSpace(binding.From), strings.TrimSpace(binding.To)
		if !temporaryWorkflowOutputBindingSourceSupported(source) {
			return nil, fmt.Errorf("temporary workflow step %d output binding source %q is not supported", stepIndex+1, source)
		}
		out = append(out, agentxworkflow.BindingSpec{From: source, To: target})
	}
	return out, nil
}

func temporaryWorkflowOutputBindingSourceSupported(source string) bool {
	source = strings.TrimSpace(source)
	return source == "output" || source == "result" || source == "status" || source == "error" || strings.HasPrefix(source, "result.")
}

func buildTemporaryWorkflowPlanningContext(input Input, userMessage string) temporaryWorkflowPlanningContext {
	sessionInputs := CloneShellBindingMap(input.SessionInput)
	userMessage = firstNonEmptyTemporaryWorkflowString(userMessage, input.UserMessage)
	if userMessage != "" {
		if _, ok := sessionInputs["request"]; !ok {
			sessionInputs["request"] = userMessage
		}
		if _, ok := sessionInputs["user_message"]; !ok {
			sessionInputs["user_message"] = userMessage
		}
	}
	if len(sessionInputs) == 0 {
		return temporaryWorkflowPlanningContext{}
	}
	paths := make([]string, 0, len(sessionInputs))
	pathSet := make(map[string]struct{}, len(sessionInputs))
	for key := range sessionInputs {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		path := "session.input." + key
		if _, exists := pathSet[path]; exists {
			continue
		}
		pathSet[path] = struct{}{}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return temporaryWorkflowPlanningContext{SessionInputs: sessionInputs, SessionInputSet: pathSet, SessionInputList: paths}
}

func temporaryWorkflowPlanningSessionInputPreview(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey != "" {
			out[trimmedKey] = temporaryWorkflowPlanningPreviewValue(value)
		}
	}
	return out
}

func temporaryWorkflowPlanningPreviewValue(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	text = strings.TrimSpace(text)
	if len(text) <= 1200 {
		return text
	}
	return text[:1200] + "...(truncated)"
}

func temporaryWorkflowPlanningRetryFeedback(err error) (string, bool) {
	var unavailable *temporaryWorkflowUnavailableSessionInputError
	if !errors.As(err, &unavailable) || unavailable == nil {
		return "", false
	}
	feedback := fmt.Sprintf("The plan referenced unavailable session input %q.", unavailable.Source)
	if len(unavailable.AvailablePaths) == 0 {
		return feedback + " No session.input bindings are available for this request.", true
	}
	return feedback + " Available session.input paths are: " + strings.Join(unavailable.AvailablePaths, ", ") + ".", true
}

func normalizeTemporaryWorkflowPlanningTools(tools []TemporaryWorkflowPlanningTool, normalize func(string) string) []TemporaryWorkflowPlanningTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]TemporaryWorkflowPlanningTool, 0, len(tools))
	for _, tool := range tools {
		name := normalize(tool.Name)
		if name == "" {
			continue
		}
		keys := append([]string(nil), tool.ArgumentKeys...)
		sort.Strings(keys)
		out = append(out, TemporaryWorkflowPlanningTool{Name: name, Description: strings.TrimSpace(tool.Description), ArgumentKeys: keys})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func temporaryWorkflowPlanningToolDescriptors(tools []TemporaryWorkflowPlanningTool) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{"name": tool.Name, "description": strings.TrimSpace(tool.Description), "arg_keys": append([]string(nil), tool.ArgumentKeys...)})
	}
	return out
}

func temporaryWorkflowPlanningToolNames(tools []TemporaryWorkflowPlanningTool) []string {
	out := make([]string, 0, len(tools))
	for _, tool := range tools {
		if name := strings.TrimSpace(tool.Name); name != "" {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func temporaryWorkflowNodeKinds(spec agentxworkflow.Spec) []string {
	if len(spec.Nodes) == 0 {
		return nil
	}
	out := make([]string, 0, len(spec.Nodes))
	for _, node := range spec.Nodes {
		out = append(out, string(node.Kind))
	}
	return out
}

func temporaryWorkflowToolNames(spec agentxworkflow.Spec) []string {
	if len(spec.Nodes) == 0 {
		return nil
	}
	out := make([]string, 0, len(spec.Nodes))
	for _, node := range spec.Nodes {
		if node.Kind == agentxworkflow.NodeTool {
			if toolName := strings.TrimSpace(temporaryWorkflowConfigString(node.Config, "tool", "tool_name")); toolName != "" {
				out = append(out, toolName)
			}
		}
	}
	sort.Strings(out)
	return out
}

func temporaryWorkflowBindingCounts(spec agentxworkflow.Spec) (int, int) {
	var inputCount, outputCount int
	for _, node := range spec.Nodes {
		inputCount += len(node.Inputs)
		outputCount += len(node.Outputs)
	}
	return inputCount, outputCount
}

func temporaryWorkflowConfigString(config map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := config[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func cloneTemporaryWorkflowMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func firstNonEmptyTemporaryWorkflowString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func temporaryWorkflowMustJSON(value any) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func temporaryWorkflowMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
