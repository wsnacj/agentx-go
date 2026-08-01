package controlcontract_test

import (
	"encoding/json"
	"reflect"
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestExternalConsumerBuildsPortableControlProjection(t *testing.T) {
	projection := controlcontract.BuildManagedObjectiveProjection(controlcontract.ManagedObjectiveProjectionInput{
		Activation: controlcontract.ActivationManaged,
		Frame: controlcontract.ObjectiveFrame{
			ID:              "objective:example",
			UserGoalDigest:  "sha256:goal",
			SuccessCriteria: []string{"result is verified"},
		},
		LedgerRef:           "ledger:example",
		Approved:            true,
		ApprovalRefs:        []controlcontract.DisplaySafeRef{"approval:example"},
		PolicyRefs:          []controlcontract.DisplaySafeRef{"contract:intensity_gate", "contract:budget", "contract:approval_policy", "contract:strategy_scope", "contract:redaction_policy"},
		AllowedStrategyRefs: []controlcontract.DisplaySafeRef{"strategy:example"},
	})
	if !projection.Ready || projection.Status != controlcontract.HostActionReady {
		t.Fatalf("projection = %#v", projection)
	}
	if projection.NextHostAction != "host_may_plan_managed_objective" || projection.RunnerEffect != "none" || projection.PromptEffect != "none" {
		t.Fatalf("unexpected portable effects: %#v", projection)
	}

	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var roundTrip controlcontract.ManagedObjectiveProjection
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	roundTrip = roundTrip.Normalize()
	if !reflect.DeepEqual(projection, roundTrip) {
		t.Fatalf("round trip mismatch:\nwant %#v\n got %#v", projection, roundTrip)
	}
}

func TestExternalConsumerUsesDeterministicGates(t *testing.T) {
	budget := controlcontract.EvaluateRetryBudgetGate(controlcontract.BudgetGateInput{
		Limit:     3,
		Used:      1,
		Increment: 1,
		Scope:     "objective:example",
	})
	if !budget.Allowed || budget.RetryBudgetRemaining != 1 || budget.Status != controlcontract.VerificationSatisfied {
		t.Fatalf("budget = %#v", budget)
	}

	transition := controlcontract.CheckLifecycleTransition(controlcontract.LifecycleStageReady, controlcontract.LifecycleStageApplied)
	if !transition.Allowed || transition.Status != controlcontract.HostActionReady {
		t.Fatalf("transition = %#v", transition)
	}

	unsafe := controlcontract.VerifyDisplaySafeOnly(false, []string{"https://example.invalid/raw"})
	if unsafe.Satisfied || unsafe.Status != controlcontract.VerificationBlocked || unsafe.FailureClass != controlcontract.FailureEvidenceWeak {
		t.Fatalf("unsafe projection = %#v", unsafe)
	}
}
