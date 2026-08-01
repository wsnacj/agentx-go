package executionpolicy

import "strings"

const (
	RuntimeControlSourceExecutionContract = "execution_contract"

	RuntimeEnforcementSurfaceRuntimePreflight = "runtime_preflight"

	RuntimePolicySourceToolCallPolicy = "tool_call_policy"
	RuntimePolicySourceRuntimeGuard   = "runtime_guard"
	RuntimePolicySourceApprovalHook   = "approval_hook"
)

type RuntimeDecision struct {
	Checked            bool   `json:"checked,omitempty"`
	Tool               string `json:"tool,omitempty"`
	Action             string `json:"action,omitempty"`
	Allowed            bool   `json:"allowed,omitempty"`
	Denied             bool   `json:"denied,omitempty"`
	RequiresConfirm    bool   `json:"requires_confirm,omitempty"`
	Degraded           bool   `json:"degraded,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Detail             string `json:"detail,omitempty"`
	DecisionSubject    string `json:"decision_subject,omitempty"`
	TargetKind         string `json:"target_kind,omitempty"`
	PolicySource       string `json:"policy_source,omitempty"`
	ControlSource      string `json:"control_source,omitempty"`
	EnforcementSurface string `json:"enforcement_surface,omitempty"`
}

type RuntimeEnforcementResult struct {
	Decision RuntimeDecision `json:"decision,omitempty"`
	Err      error           `json:"-"`
}

type RuntimeDecisionSummary struct {
	Checks          int            `json:"checks,omitempty"`
	Allowed         int            `json:"allowed,omitempty"`
	Denied          int            `json:"denied,omitempty"`
	RequiresConfirm int            `json:"requires_confirm,omitempty"`
	Degraded        int            `json:"degraded,omitempty"`
	ByPolicySource  map[string]int `json:"by_policy_source,omitempty"`
	ByReason        map[string]int `json:"by_reason,omitempty"`
	BySubject       map[string]int `json:"by_subject,omitempty"`
	ByTargetKind    map[string]int `json:"by_target_kind,omitempty"`
	ByTool          map[string]int `json:"by_tool,omitempty"`
}

func RuntimeDecisionSummaryFromDecision(decision RuntimeDecision) RuntimeDecisionSummary {
	if !decision.Checked {
		return RuntimeDecisionSummary{}
	}
	summary := RuntimeDecisionSummary{
		Checks: 1,
	}
	if decision.Allowed {
		summary.Allowed = 1
	}
	if decision.Denied {
		summary.Denied = 1
	}
	if decision.RequiresConfirm {
		summary.RequiresConfirm = 1
	}
	if decision.Degraded {
		summary.Degraded = 1
	}
	if source := strings.TrimSpace(decision.PolicySource); source != "" {
		summary.ByPolicySource = map[string]int{source: 1}
	}
	if reason := strings.TrimSpace(decision.Reason); reason != "" {
		summary.ByReason = map[string]int{reason: 1}
	}
	if subject := strings.TrimSpace(decision.DecisionSubject); subject != "" {
		summary.BySubject = map[string]int{subject: 1}
	}
	if targetKind := strings.TrimSpace(decision.TargetKind); targetKind != "" {
		summary.ByTargetKind = map[string]int{targetKind: 1}
	}
	if tool := strings.TrimSpace(decision.Tool); tool != "" {
		summary.ByTool = map[string]int{tool: 1}
	}
	return summary
}

func MergeRuntimeDecisionSummary(base RuntimeDecisionSummary, delta RuntimeDecisionSummary) RuntimeDecisionSummary {
	base.Checks += delta.Checks
	base.Allowed += delta.Allowed
	base.Denied += delta.Denied
	base.RequiresConfirm += delta.RequiresConfirm
	base.Degraded += delta.Degraded
	base.ByPolicySource = mergeRuntimeDecisionCountMap(base.ByPolicySource, delta.ByPolicySource)
	base.ByReason = mergeRuntimeDecisionCountMap(base.ByReason, delta.ByReason)
	base.BySubject = mergeRuntimeDecisionCountMap(base.BySubject, delta.BySubject)
	base.ByTargetKind = mergeRuntimeDecisionCountMap(base.ByTargetKind, delta.ByTargetKind)
	base.ByTool = mergeRuntimeDecisionCountMap(base.ByTool, delta.ByTool)
	return base
}

func allowRuntimeDecision(tool string, action string, policySource string, reason string, detail string) RuntimeEnforcementResult {
	return RuntimeEnforcementResult{
		Decision: RuntimeDecision{
			Checked:            true,
			Tool:               strings.TrimSpace(tool),
			Action:             strings.TrimSpace(action),
			Allowed:            true,
			Reason:             strings.TrimSpace(reason),
			Detail:             strings.TrimSpace(detail),
			PolicySource:       strings.TrimSpace(policySource),
			ControlSource:      RuntimeControlSourceExecutionContract,
			EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
		},
	}
}

func degradeRuntimeDecision(tool string, action string, policySource string, reason string, detail string) RuntimeEnforcementResult {
	result := allowRuntimeDecision(tool, action, policySource, reason, detail)
	result.Decision.Degraded = true
	return result
}

func denyRuntimeDecision(tool string, action string, policySource string, reason string, err error) RuntimeEnforcementResult {
	result := RuntimeEnforcementResult{
		Decision: RuntimeDecision{
			Checked:            true,
			Tool:               strings.TrimSpace(tool),
			Action:             strings.TrimSpace(action),
			Denied:             true,
			Reason:             strings.TrimSpace(reason),
			PolicySource:       strings.TrimSpace(policySource),
			ControlSource:      RuntimeControlSourceExecutionContract,
			EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
		},
		Err: err,
	}
	if err != nil {
		result.Decision.Detail = strings.TrimSpace(err.Error())
	}
	return result
}

func annotateRuntimeDecisionTarget(result RuntimeEnforcementResult, subject string, targetKind string) RuntimeEnforcementResult {
	if !result.Decision.Checked {
		return result
	}
	result.Decision.DecisionSubject = strings.TrimSpace(subject)
	result.Decision.TargetKind = strings.TrimSpace(targetKind)
	return result
}

func mergeRuntimeDecisionCountMap(base map[string]int, delta map[string]int) map[string]int {
	if len(delta) == 0 {
		if len(base) == 0 {
			return nil
		}
		return base
	}
	if base == nil {
		base = map[string]int{}
	}
	for key, value := range delta {
		key = strings.TrimSpace(key)
		if key == "" || value == 0 {
			continue
		}
		base[key] += value
	}
	if len(base) == 0 {
		return nil
	}
	return base
}
