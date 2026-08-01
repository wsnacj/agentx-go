package executionpolicy

import "testing"

func TestNormalizeSoftRejectionAction(t *testing.T) {
	for raw, want := range map[string]string{
		"":               SoftRejectionActionAllow,
		"allowed":        SoftRejectionActionAllow,
		"reject":         SoftRejectionActionRejectContent,
		"truncate":       SoftRejectionActionRejectContent,
		"blocked":        SoftRejectionActionHalt,
		"reject_content": SoftRejectionActionRejectContent,
		"halt":           SoftRejectionActionHalt,
	} {
		if got := NormalizeSoftRejectionAction(raw); got != want {
			t.Fatalf("NormalizeSoftRejectionAction(%q)=%q want %q", raw, got, want)
		}
	}
}

func TestPrimarySoftRejectionDecisionPrefersHalt(t *testing.T) {
	primary, ok := PrimarySoftRejectionDecision([]SoftRejectionDecision{
		NewSoftRejectionDecision(SoftRejectionActionRejectContent, SoftRejectionSourceToolOutputGuard, "open_page", "runtime_guard_truncated", ""),
		NewSoftRejectionDecision(SoftRejectionActionHalt, SoftRejectionSourceApproval, "authorize", "policy_deny", ""),
	})
	if !ok {
		t.Fatal("expected primary decision")
	}
	if primary.Action != SoftRejectionActionHalt ||
		primary.Source != SoftRejectionSourceApproval ||
		primary.Reason != "policy_deny" {
		t.Fatalf("unexpected primary soft rejection: %#v", primary)
	}
}
