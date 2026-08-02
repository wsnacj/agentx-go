package tools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// BrowserLocalPlannerContext is the compressed browser-local input surface for
// a constrained planner. It intentionally excludes general run/session history.
type BrowserLocalPlannerContext struct {
	Tool           string                               `json:"tool,omitempty"`
	Action         string                               `json:"action,omitempty"`
	CurrentArgs    map[string]any                       `json:"current_args,omitempty"`
	CurrentResult  map[string]any                       `json:"current_result,omitempty"`
	Page           BrowserLocalPlannerPage              `json:"page,omitempty"`
	RecentToolRuns []BrowserLocalPlannerRecentToolRun   `json:"recent_tool_runs,omitempty"`
	LatestSnapshot *BrowserLocalPlannerSnapshotSummary  `json:"latest_snapshot,omitempty"`
	Policy         BrowserLocalPlannerPolicyConstraints `json:"policy,omitempty"`
}

// BrowserLocalPlannerPage captures the current page-level browser facts.
type BrowserLocalPlannerPage struct {
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

// BrowserLocalPlannerRecentToolRun summarizes a recent browser-local action.
type BrowserLocalPlannerRecentToolRun struct {
	Action          string `json:"action,omitempty"`
	Status          string `json:"status,omitempty"`
	SummaryCode     string `json:"summary_code,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

// BrowserLocalPlannerSnapshotSummary is the compressed latest snapshot surface
// that the planner can use without requiring the full DOM payload.
type BrowserLocalPlannerSnapshotSummary struct {
	Format             string   `json:"format,omitempty"`
	TopClickableLabels []string `json:"top_clickable_labels,omitempty"`
	TopRefs            []string `json:"top_refs,omitempty"`
}

// BrowserLocalPlannerPolicyConstraints describes execution-side boundaries that
// the planner must respect. It is informational and does not replace runtime
// policy enforcement.
type BrowserLocalPlannerPolicyConstraints struct {
	AllowForceRetry bool   `json:"allow_force_retry,omitempty"`
	RiskTier        string `json:"risk_tier,omitempty"`
}

// BrowserLocalPlannerAction is a single allowlisted browser follow-up action.
type BrowserLocalPlannerAction struct {
	Kind   string         `json:"kind,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

// BrowserLocalPlannerDecision is the strict JSON contract returned by the
// constrained planner.
type BrowserLocalPlannerDecision struct {
	Decision   string                     `json:"decision,omitempty"`
	ReasonCode string                     `json:"reason_code,omitempty"`
	Confidence string                     `json:"confidence,omitempty"`
	Action     *BrowserLocalPlannerAction `json:"action,omitempty"`
}

// BrowserLocalPlannerTelemetry carries planner-specific observability fields
// without requiring engine code to understand browser-local policy details.
type BrowserLocalPlannerTelemetry struct {
	Invoked                bool   `json:"invoked,omitempty"`
	ReasonCode             string `json:"reason_code,omitempty"`
	Model                  string `json:"model,omitempty"`
	LatencyMs              int64  `json:"latency_ms,omitempty"`
	Decision               string `json:"decision,omitempty"`
	ActionKind             string `json:"action_kind,omitempty"`
	FollowupOK             bool   `json:"followup_ok,omitempty"`
	FollowupRecovered      bool   `json:"followup_recovered,omitempty"`
	BlockedByPolicy        bool   `json:"blocked_by_policy,omitempty"`
	DiscardedInvalidOutput bool   `json:"discarded_invalid_output,omitempty"`
}

var browserLocalPlannerAllowedDecisions = map[string]bool{
	"retry_one_step":                 true,
	"refresh_then_retry":             true,
	"surface_review":                 true,
	"stop_and_return_current_result": true,
}

var browserLocalPlannerAllowedConfidence = map[string]bool{
	"":       true,
	"low":    true,
	"medium": true,
	"high":   true,
}

var browserLocalPlannerAllowedActionKinds = map[string]bool{
	"snapshot":      true,
	"click":         true,
	"wait_download": true,
	"extract":       true,
	"wait":          true,
}

var browserLocalPlannerAllowedParamKeys = map[string]map[string]bool{
	"snapshot": {
		"mode":    true,
		"refs":    true,
		"wait_ms": true,
		"force":   true,
	},
	"click": {
		"selector": true,
		"ref":      true,
		"element":  true,
		"wait_ms":  true,
		"force":    true,
	},
	"wait_download": {
		"wait_ms": true,
		"force":   true,
	},
	"extract": {
		"format":  true,
		"query":   true,
		"wait_ms": true,
		"force":   true,
	},
	"wait": {
		"wait_ms": true,
	},
}

// BrowserLocalPlannerAllowedDecisions returns the strict decision allowlist.
func BrowserLocalPlannerAllowedDecisions() []string {
	return browserLocalPlannerAllowlistKeys(browserLocalPlannerAllowedDecisions)
}

// BrowserLocalPlannerAllowedActionKinds returns the strict follow-up kind allowlist.
func BrowserLocalPlannerAllowedActionKinds() []string {
	return browserLocalPlannerAllowlistKeys(browserLocalPlannerAllowedActionKinds)
}

// DecodeBrowserLocalPlannerDecision parses and validates a planner decision.
func DecodeBrowserLocalPlannerDecision(raw string) (BrowserLocalPlannerDecision, error) {
	var decision BrowserLocalPlannerDecision
	if strings.TrimSpace(raw) == "" {
		return decision, fmt.Errorf("browser local planner: empty decision payload")
	}
	if err := json.Unmarshal([]byte(raw), &decision); err != nil {
		return BrowserLocalPlannerDecision{}, fmt.Errorf("browser local planner: decode decision: %w", err)
	}
	decision = normalizeBrowserLocalPlannerDecision(decision)
	if err := ValidateBrowserLocalPlannerDecision(decision); err != nil {
		return BrowserLocalPlannerDecision{}, err
	}
	return decision, nil
}

// ValidateBrowserLocalPlannerDecision enforces the fail-closed decision contract.
func ValidateBrowserLocalPlannerDecision(decision BrowserLocalPlannerDecision) error {
	decision = normalizeBrowserLocalPlannerDecision(decision)
	if !browserLocalPlannerAllowedDecisions[decision.Decision] {
		return fmt.Errorf("browser local planner: decision must be one of %s", strings.Join(BrowserLocalPlannerAllowedDecisions(), ", "))
	}
	if !browserLocalPlannerAllowedConfidence[decision.Confidence] {
		return fmt.Errorf("browser local planner: confidence must be one of low, medium, high")
	}
	requiresAction := browserLocalPlannerDecisionRequiresAction(decision.Decision)
	if requiresAction && decision.Action == nil {
		return fmt.Errorf("browser local planner: decision %q requires action", decision.Decision)
	}
	if !requiresAction && decision.Action != nil {
		return fmt.Errorf("browser local planner: decision %q must not include action", decision.Decision)
	}
	if decision.Action == nil {
		return nil
	}
	return ValidateBrowserLocalPlannerAction(*decision.Action)
}

// ValidateBrowserLocalPlannerAction enforces the allowlisted action contract.
func ValidateBrowserLocalPlannerAction(action BrowserLocalPlannerAction) error {
	kind := normalizeBrowserLocalPlannerToken(action.Kind)
	if !browserLocalPlannerAllowedActionKinds[kind] {
		return fmt.Errorf("browser local planner: action kind must be one of %s", strings.Join(BrowserLocalPlannerAllowedActionKinds(), ", "))
	}
	allowedKeys := browserLocalPlannerAllowedParamKeys[kind]
	for key, value := range action.Params {
		normalizedKey := normalizeBrowserLocalPlannerToken(key)
		if !allowedKeys[normalizedKey] {
			return fmt.Errorf("browser local planner: action kind %q does not allow param %q", kind, strings.TrimSpace(key))
		}
		if err := validateBrowserLocalPlannerParamValue(value); err != nil {
			return fmt.Errorf("browser local planner: action param %q invalid: %w", strings.TrimSpace(key), err)
		}
	}
	return nil
}

// NormalizeBrowserLocalPlannerTelemetry trims and canonicalizes telemetry fields
// before they are projected into engine diagnostics.
func NormalizeBrowserLocalPlannerTelemetry(in BrowserLocalPlannerTelemetry) BrowserLocalPlannerTelemetry {
	in.ReasonCode = normalizeBrowserLocalPlannerToken(in.ReasonCode)
	in.Decision = normalizeBrowserLocalPlannerToken(in.Decision)
	in.ActionKind = normalizeBrowserLocalPlannerToken(in.ActionKind)
	in.Model = strings.TrimSpace(in.Model)
	if in.LatencyMs < 0 {
		in.LatencyMs = 0
	}
	return in
}

func browserLocalPlannerDecisionRequiresAction(decision string) bool {
	switch normalizeBrowserLocalPlannerToken(decision) {
	case "retry_one_step", "refresh_then_retry":
		return true
	default:
		return false
	}
}

func normalizeBrowserLocalPlannerDecision(in BrowserLocalPlannerDecision) BrowserLocalPlannerDecision {
	in.Decision = normalizeBrowserLocalPlannerToken(in.Decision)
	in.ReasonCode = normalizeBrowserLocalPlannerToken(in.ReasonCode)
	in.Confidence = normalizeBrowserLocalPlannerToken(in.Confidence)
	if in.Action != nil {
		normalized := BrowserLocalPlannerAction{
			Kind:   normalizeBrowserLocalPlannerToken(in.Action.Kind),
			Params: in.Action.Params,
		}
		if len(normalized.Params) == 0 {
			normalized.Params = nil
		}
		in.Action = &normalized
	}
	return in
}

func normalizeBrowserLocalPlannerToken(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validateBrowserLocalPlannerParamValue(value any) error {
	switch typed := value.(type) {
	case nil:
		return nil
	case string, bool, float64, json.Number:
		return nil
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case []string:
		return nil
	case []any:
		for _, item := range typed {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("array values must be strings")
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported value type %T", value)
	}
}

func browserLocalPlannerAllowlistKeys(items map[string]bool) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for key := range items {
		if strings.TrimSpace(key) == "" {
			continue
		}
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
