package contracts

// AssessmentKind describes formula-derived or model-interpreted outputs.
type AssessmentKind string

const (
	AssessmentKindNone       AssessmentKind = "none"
	AssessmentKindValuation  AssessmentKind = "valuation"
	AssessmentKindInvestment AssessmentKind = "investment_risk"
	AssessmentKindBusiness   AssessmentKind = "business_performance"
)

// AssessmentBoundary separates verified facts from derived observations and advice limits.
type AssessmentBoundary struct {
	Kind                 AssessmentKind `json:"kind,omitempty"`
	Scope                string         `json:"scope,omitempty"`
	VerifiedFacts        []string       `json:"verified_facts,omitempty"`
	FormulaObservations  []string       `json:"formula_observations,omitempty"`
	ModelInterpretations []string       `json:"model_interpretations,omitempty"`
	MissingInputs        []string       `json:"missing_inputs,omitempty"`
	RiskFactors          []string       `json:"risk_factors,omitempty"`
	AdviceBoundary       string         `json:"advice_boundary,omitempty"`
	NotInvestmentAdvice  bool           `json:"not_investment_advice,omitempty"`
	Evidence             SourceEvidence `json:"evidence,omitempty"`
	Readiness            Readiness      `json:"readiness,omitempty"`
}

// BuildAssessmentBoundary creates a conservative assessment contract from evidence readiness.
func BuildAssessmentBoundary(kind AssessmentKind, scope string, readiness Readiness, verifiedFacts []string, missingInputs []string) AssessmentBoundary {
	out := AssessmentBoundary{
		Kind:                kind,
		Scope:               scope,
		VerifiedFacts:       append([]string(nil), verifiedFacts...),
		MissingInputs:       append([]string(nil), missingInputs...),
		NotInvestmentAdvice: kind == AssessmentKindInvestment || kind == AssessmentKindValuation,
		Readiness:           readiness,
	}
	if !readiness.AnswerReady {
		out.AdviceBoundary = "Evidence is incomplete; provide a limited, source-scoped assessment only."
		return out
	}
	if out.NotInvestmentAdvice {
		out.AdviceBoundary = "This is source-backed analysis, not personalized investment advice."
	}
	return out
}
