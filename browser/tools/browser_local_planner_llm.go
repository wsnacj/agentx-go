package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const browserLocalPlannerDefaultSystemPrompt = "You are a browser-local constrained planner. Use only the provided browser-local context. Do not reinterpret the full user task. Return exactly one JSON object and nothing else. Allowed decisions: retry_one_step, refresh_then_retry, surface_review, stop_and_return_current_result. Allowed action kinds: snapshot, click, wait_download, extract, wait. Only use allowlisted params. If the safest action is unclear, return surface_review or stop_and_return_current_result."

func invokeBrowserLocalPlanner(
	ctx context.Context,
	opts BrowserToolOptions,
	eligibility BrowserLocalPlannerEligibility,
	plannerContext BrowserLocalPlannerContext,
) (BrowserLocalPlannerDecision, BrowserLocalPlannerTelemetry, error) {
	telemetry := BrowserLocalPlannerTelemetry{
		Invoked:    true,
		ReasonCode: eligibility.ReasonCode,
		Model:      strings.TrimSpace(opts.BrowserLocalPlannerModel),
	}
	model := strings.TrimSpace(opts.BrowserLocalPlannerModel)
	if model == "" {
		return BrowserLocalPlannerDecision{}, telemetry, fmt.Errorf("browser local planner: model is not configured")
	}

	payload := map[string]any{
		"eligibility": eligibility,
		"context":     plannerContext,
		"contract": map[string]any{
			"allowed_decisions":    BrowserLocalPlannerAllowedDecisions(),
			"allowed_action_kinds": BrowserLocalPlannerAllowedActionKinds(),
		},
		"execution_constraints": map[string]any{
			"single_step_only":                     true,
			"prefer_surface_review_when_uncertain": true,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return BrowserLocalPlannerDecision{}, telemetry, fmt.Errorf("browser local planner: marshal input: %w", err)
	}

	plannerCtx := ctx
	timeoutMs := opts.BrowserLocalPlannerTimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = opts.TimeoutMs
	}
	if timeoutMs <= 0 {
		timeoutMs = defaultLLMTaskTimeoutMs
	}
	var cancel context.CancelFunc
	if timeoutMs > 0 {
		plannerCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	started := time.Now()
	content, err := invokeBrowserLocalPlannerChat(plannerCtx, opts.LocalPlannerChat, model, string(body))
	telemetry.LatencyMs = time.Since(started).Milliseconds()
	if err != nil {
		return BrowserLocalPlannerDecision{}, telemetry, err
	}
	if strings.TrimSpace(content) == "" {
		return BrowserLocalPlannerDecision{}, telemetry, fmt.Errorf("browser local planner: empty chat response")
	}
	decision, err := DecodeBrowserLocalPlannerDecision(browserLocalPlannerResponseContent(content))
	if err != nil {
		telemetry.DiscardedInvalidOutput = true
		return BrowserLocalPlannerDecision{}, telemetry, err
	}
	telemetry.Decision = decision.Decision
	if decision.Action != nil {
		telemetry.ActionKind = decision.Action.Kind
	}
	return decision, telemetry, nil
}

func invokeBrowserLocalPlannerChat(ctx context.Context, chat BrowserLocalPlannerChat, model string, payload string) (string, error) {
	if chat != nil {
		return chat(ctx, model, browserLocalPlannerDefaultSystemPrompt, payload)
	}
	return "", fmt.Errorf("browser local planner: typed chat is not configured")
}

func browserLocalPlannerResponseContent(raw string) string {
	content := strings.TrimSpace(raw)
	if content == "" {
		return ""
	}
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) >= 3 {
			content = strings.Join(lines[1:len(lines)-1], "\n")
		}
		content = strings.TrimSpace(content)
	}
	return content
}
