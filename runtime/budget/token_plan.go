package budget

import llm "github.com/wsnacj/agentx-go/components/llm"

const (
	// ReasonContextWindowTokens marks a request whose reserved input and output
	// cannot fit in the configured model context window.
	ReasonContextWindowTokens = "context_window_tokens"
	// ReasonInvalidTokenPlan marks negative request-owned token values.
	ReasonInvalidTokenPlan = "invalid_token_plan"
)

// TokenPlanRequest combines a model's declared limits with one request-time
// input count and requested output reserve.
type TokenPlanRequest struct {
	Limits                llm.ModelLimits
	Input                 llm.TokenCount
	RequestedOutputTokens int64
	ReservedOutputTokens  int64
}

// TokenPlan is the deterministic request-time budget decision. An input limit
// of zero means no limit was declared by the Host.
type TokenPlan struct {
	Allowed              bool
	Reason               string
	Input                llm.TokenCount
	InputLimitTokens     int64
	OutputLimitTokens    int64
	PlannedOutputTokens  int64
	ReservedOutputTokens int64
	ContextWindowTokens  int64
}

// PlanTokens computes a provider-neutral token window without performing I/O
// or choosing a model. It never treats an estimated input as reported usage.
func PlanTokens(request TokenPlanRequest) TokenPlan {
	plan := TokenPlan{Input: request.Input}
	if request.Input.Tokens < 0 || request.RequestedOutputTokens < 0 || request.ReservedOutputTokens < 0 {
		plan.Reason = ReasonInvalidTokenPlan
		return plan
	}
	limits := request.Limits.Normalize()
	plan.ContextWindowTokens = limits.ContextWindowTokens
	plan.OutputLimitTokens = limits.MaxOutputTokens

	plannedOutput := request.RequestedOutputTokens
	if plannedOutput == 0 {
		plannedOutput = limits.MaxOutputTokens
	}
	if limits.MaxOutputTokens > 0 && plannedOutput > limits.MaxOutputTokens {
		plan.Reason = ReasonMaxOutputTokens
		return plan
	}
	plan.PlannedOutputTokens = plannedOutput
	reservedOutput := maxInt64(request.ReservedOutputTokens, plannedOutput)
	plan.ReservedOutputTokens = reservedOutput

	inputLimit := limits.MaxInputTokens
	if limits.ContextWindowTokens > 0 {
		if reservedOutput >= limits.ContextWindowTokens {
			plan.Reason = ReasonContextWindowTokens
			return plan
		}
		windowInput := limits.ContextWindowTokens - reservedOutput
		inputLimit = minPositiveInt64(inputLimit, windowInput)
	}
	plan.InputLimitTokens = inputLimit
	if limits.MaxInputTokens > 0 && request.Input.Tokens > limits.MaxInputTokens {
		plan.Reason = ReasonMaxInputTokens
		return plan
	}
	if inputLimit > 0 && request.Input.Tokens > inputLimit {
		plan.Reason = ReasonContextWindowTokens
		return plan
	}
	plan.Allowed = true
	return plan
}

func minPositiveInt64(left, right int64) int64 {
	if left <= 0 {
		return right
	}
	if right <= 0 || left < right {
		return left
	}
	return right
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
