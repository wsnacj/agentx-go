package budget

import (
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestPlanTokensReservesOutputInsideContextWindow(t *testing.T) {
	plan := PlanTokens(TokenPlanRequest{
		Limits:                llm.ModelLimits{ContextWindowTokens: 100, MaxInputTokens: 90, MaxOutputTokens: 40},
		Input:                 llm.TokenCount{Tokens: 60, Exact: true, Source: "fixture"},
		RequestedOutputTokens: 30,
	})
	if !plan.Allowed || plan.InputLimitTokens != 70 || plan.PlannedOutputTokens != 30 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestPlanTokensRejectsContextOverflow(t *testing.T) {
	plan := PlanTokens(TokenPlanRequest{
		Limits:                llm.ModelLimits{ContextWindowTokens: 100, MaxInputTokens: 90, MaxOutputTokens: 40},
		Input:                 llm.TokenCount{Tokens: 71},
		RequestedOutputTokens: 30,
	})
	if plan.Allowed || plan.Reason != ReasonContextWindowTokens {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestPlanTokensRejectsOutputAboveModelLimit(t *testing.T) {
	plan := PlanTokens(TokenPlanRequest{
		Limits:                llm.ModelLimits{ContextWindowTokens: 100, MaxOutputTokens: 20},
		Input:                 llm.TokenCount{Tokens: 10},
		RequestedOutputTokens: 21,
	})
	if plan.Allowed || plan.Reason != ReasonMaxOutputTokens {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
