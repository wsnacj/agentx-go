package tools

import (
	"fmt"
	"strings"
)

const browserCDPEscapeHatchPolicyName = "browser_cdp_escape_hatch"

type browserCDPEscapeHatchPolicyDecision struct {
	Allowed    bool
	Configured bool
	Reason     string
}

type browserResultSafetyEvent struct {
	EventCode        string `json:"event_code"`
	Severity         string `json:"severity"`
	Source           string `json:"source,omitempty"`
	Action           string `json:"action,omitempty"`
	Backend          string `json:"backend,omitempty"`
	RuntimeTarget    string `json:"runtime_target,omitempty"`
	Policy           string `json:"policy,omitempty"`
	Decision         string `json:"decision,omitempty"`
	Reason           string `json:"reason,omitempty"`
	PolicyConfigured bool   `json:"policy_configured,omitempty"`
}

func browserCDPEscapeHatchPolicyDecisionForOptions(opts BrowserToolOptions) browserCDPEscapeHatchPolicyDecision {
	if opts.BrowserCDPEscapeHatchAllowed == nil {
		return browserCDPEscapeHatchPolicyDecision{
			Allowed: true,
			Reason:  "default_allow",
		}
	}
	if !*opts.BrowserCDPEscapeHatchAllowed {
		return browserCDPEscapeHatchPolicyDecision{
			Allowed:    false,
			Configured: true,
			Reason:     "disabled_by_policy",
		}
	}
	return browserCDPEscapeHatchPolicyDecision{
		Allowed:    true,
		Configured: true,
		Reason:     "enabled_by_policy",
	}
}

func browserCheckCDPEscapeHatchPolicy(opts BrowserToolOptions) (browserCDPEscapeHatchPolicyDecision, error) {
	decision := browserCDPEscapeHatchPolicyDecisionForOptions(opts)
	if decision.Allowed {
		return decision, nil
	}
	return decision, fmt.Errorf("cdp_escape_hatch_blocked: browser JavaScript evaluation is disabled by policy")
}

func browserResultSafetyEventForCDPEscapeHatch(source string, backend string, runtimeTarget string, decision browserCDPEscapeHatchPolicyDecision) *browserResultSafetyEvent {
	eventCode := "browser_cdp_escape_hatch_allowed"
	severity := "info"
	decisionText := "allowed"
	if !decision.Allowed {
		eventCode = "browser_cdp_escape_hatch_blocked"
		severity = "warning"
		decisionText = "denied"
	}
	source = strings.TrimSpace(source)
	action := "evaluate"
	if source == "browser_act" {
		action = "browser_act kind=evaluate"
	}
	return &browserResultSafetyEvent{
		EventCode:        eventCode,
		Severity:         severity,
		Source:           source,
		Action:           action,
		Backend:          strings.TrimSpace(backend),
		RuntimeTarget:    strings.TrimSpace(runtimeTarget),
		Policy:           browserCDPEscapeHatchPolicyName,
		Decision:         decisionText,
		Reason:           strings.TrimSpace(decision.Reason),
		PolicyConfigured: decision.Configured,
	}
}

func browserActResultSafetyEvent(opts BrowserToolOptions, result BrowserActResult) *browserResultSafetyEvent {
	if browserNormalizeToolToken(result.Kind) != "evaluate" {
		return nil
	}
	return browserResultSafetyEventForCDPEscapeHatch(
		"browser_act",
		result.Backend,
		result.RuntimeTarget,
		browserCDPEscapeHatchPolicyDecisionForOptions(opts),
	)
}
