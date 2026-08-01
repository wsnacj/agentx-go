package main

import (
	"fmt"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func run() (string, error) {
	projection := controlcontract.BuildManagedObjectiveProjection(controlcontract.ManagedObjectiveProjectionInput{
		Activation: controlcontract.ActivationManaged,
		Frame: controlcontract.ObjectiveFrame{
			ID:             "objective:fixed-consumer",
			UserGoalDigest: "sha256:fixed-consumer",
			SuccessCriteria: []string{"verified output"},
		},
		LedgerRef:    "ledger:fixed-consumer",
		Approved:     true,
		ApprovalRefs: []controlcontract.DisplaySafeRef{"approval:fixed-consumer"},
		PolicyRefs: []controlcontract.DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
		AllowedStrategyRefs: []controlcontract.DisplaySafeRef{"strategy:fixed-consumer"},
	})
	if !projection.Ready || projection.Status != controlcontract.HostActionReady {
		return "", fmt.Errorf("projection is not ready: %#v", projection)
	}

	budget := controlcontract.EvaluateRetryBudgetGate(controlcontract.BudgetGateInput{
		Limit:     3,
		Used:      1,
		Increment: 1,
		Scope:     "objective:fixed-consumer",
	})
	if !budget.Allowed {
		return "", fmt.Errorf("budget gate blocked: %#v", budget)
	}

	lifecycle := controlcontract.CheckLifecycleTransition(
		controlcontract.LifecycleStageReady,
		controlcontract.LifecycleStageApplied,
	)
	if !lifecycle.Allowed {
		return "", fmt.Errorf("lifecycle transition blocked: %#v", lifecycle)
	}

	unsafe := controlcontract.VerifyDisplaySafeOnly(false, []string{"https://example.invalid/raw"})
	if unsafe.Satisfied || unsafe.Status != controlcontract.VerificationBlocked {
		return "", fmt.Errorf("unsafe ref was not rejected: %#v", unsafe)
	}

	return fmt.Sprintf(
		"agentx-controlcontract-ok:%s:%d:%s:%s",
		projection.Status,
		budget.RetryBudgetRemaining,
		lifecycle.To,
		unsafe.FailureClass,
	), nil
}

func main() {
	result, err := run()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
