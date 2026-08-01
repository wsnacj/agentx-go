package executionpolicy

import "strings"

const (
	SoftRejectionActionAllow         = "allow"
	SoftRejectionActionRejectContent = "reject_content"
	SoftRejectionActionHalt          = "halt"

	SoftRejectionSourceArgumentRepair  = "argument_repair"
	SoftRejectionSourceToolOutputGuard = "tool_output_guard"
	SoftRejectionSourceApproval        = "approval"
)

type SoftRejectionDecision struct {
	Action       string `json:"action,omitempty"`
	Source       string `json:"source,omitempty"`
	Surface      string `json:"surface,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Detail       string `json:"detail,omitempty"`
	PolicySource string `json:"policy_source,omitempty"`
}

func NewSoftRejectionDecision(action string, source string, surface string, reason string, detail string) SoftRejectionDecision {
	return NormalizeSoftRejectionDecision(SoftRejectionDecision{
		Action:  action,
		Source:  source,
		Surface: surface,
		Reason:  reason,
		Detail:  detail,
	})
}

func NormalizeSoftRejectionDecision(in SoftRejectionDecision) SoftRejectionDecision {
	out := in
	out.Action = NormalizeSoftRejectionAction(out.Action)
	out.Source = strings.ToLower(strings.TrimSpace(out.Source))
	out.Surface = strings.TrimSpace(out.Surface)
	out.Reason = strings.TrimSpace(out.Reason)
	out.Detail = strings.TrimSpace(out.Detail)
	out.PolicySource = strings.TrimSpace(out.PolicySource)
	return out
}

func NormalizeSoftRejectionAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", SoftRejectionActionAllow, "allowed":
		return SoftRejectionActionAllow
	case SoftRejectionActionRejectContent, "reject", "content_rejected", "redact", "truncate":
		return SoftRejectionActionRejectContent
	case SoftRejectionActionHalt, "deny", "denied", "block", "blocked":
		return SoftRejectionActionHalt
	default:
		return ""
	}
}

func PrimarySoftRejectionDecision(decisions []SoftRejectionDecision) (SoftRejectionDecision, bool) {
	var primary SoftRejectionDecision
	found := false
	for _, decision := range decisions {
		decision = NormalizeSoftRejectionDecision(decision)
		if decision.Action == "" {
			continue
		}
		if !found || softRejectionActionSeverity(decision.Action) > softRejectionActionSeverity(primary.Action) {
			primary = decision
			found = true
		}
	}
	return primary, found
}

func softRejectionActionSeverity(action string) int {
	switch NormalizeSoftRejectionAction(action) {
	case SoftRejectionActionHalt:
		return 3
	case SoftRejectionActionRejectContent:
		return 2
	case SoftRejectionActionAllow:
		return 1
	default:
		return 0
	}
}
