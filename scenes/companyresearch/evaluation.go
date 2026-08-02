package companyresearch

import "strings"

type CompanyResearchEvaluationInput struct {
	UserMessage                 string
	CaseType                    string
	SubjectCount                int
	ExpectedSubjectCount        int
	SubjectResolved             bool
	AnswerReady                 bool
	GuardStatus                 string
	RequestedDimensions         []string
	ReadyDimensions             []string
	MissingDimensions           []string
	SourceBacked                bool
	FreshnessConfirmed          bool
	AnswerContractRecommended   bool
	FinalAnswerBoundaryObserved bool
	OverClaimDetected           bool
	TaskConflictCount           int
	SubjectResolutionDrift      bool
	UnguardedSynthesisDetected  bool
}

type CompanyResearchEvaluation struct {
	Passed                      bool     `json:"passed"`
	SubjectCorrect              bool     `json:"subject_correct"`
	EvidenceComplete            bool     `json:"evidence_complete"`
	SourceBacked                bool     `json:"source_backed"`
	FreshnessConfirmed          bool     `json:"freshness_confirmed"`
	AnswerReady                 bool     `json:"answer_ready"`
	GuardPassed                 bool     `json:"guard_passed"`
	BoundaryOK                  bool     `json:"boundary_ok"`
	OverClaimDetected           bool     `json:"over_claim_detected"`
	AnswerContractRecommended   bool     `json:"answer_contract_recommended"`
	FinalAnswerBoundaryObserved bool     `json:"final_answer_boundary_observed"`
	TaskConflictFree            bool     `json:"task_conflict_free"`
	SubjectResolutionDrift      bool     `json:"subject_resolution_drift"`
	UnguardedSynthesisDetected  bool     `json:"unguarded_synthesis_detected"`
	ReadyDimensions             []string `json:"ready_dimensions,omitempty"`
	MissingDimensions           []string `json:"missing_dimensions,omitempty"`
	FailureReason               string   `json:"failure_reason,omitempty"`
}

func EvaluateCompanyResearchEvidence(input CompanyResearchEvaluationInput) CompanyResearchEvaluation {
	missing := normalizeEvaluationStrings(input.MissingDimensions)
	ready := normalizeEvaluationStrings(input.ReadyDimensions)
	subjectCorrect := input.SubjectResolved
	if input.ExpectedSubjectCount > 0 {
		subjectCorrect = subjectCorrect && input.SubjectCount >= input.ExpectedSubjectCount
	}
	guardPassed := strings.EqualFold(strings.TrimSpace(input.GuardStatus), "passed")
	evidenceComplete := input.AnswerReady && len(missing) == 0 && requestedDimensionsReady(input.RequestedDimensions, ready)
	boundaryOK := !input.OverClaimDetected && (input.AnswerReady || (input.AnswerContractRecommended && input.FinalAnswerBoundaryObserved))
	taskConflictFree := input.TaskConflictCount == 0 && !input.SubjectResolutionDrift
	out := CompanyResearchEvaluation{
		SubjectCorrect:              subjectCorrect,
		EvidenceComplete:            evidenceComplete,
		SourceBacked:                input.SourceBacked,
		FreshnessConfirmed:          input.FreshnessConfirmed,
		AnswerReady:                 input.AnswerReady,
		GuardPassed:                 guardPassed,
		BoundaryOK:                  boundaryOK,
		OverClaimDetected:           input.OverClaimDetected,
		AnswerContractRecommended:   input.AnswerContractRecommended,
		FinalAnswerBoundaryObserved: input.FinalAnswerBoundaryObserved,
		TaskConflictFree:            taskConflictFree,
		SubjectResolutionDrift:      input.SubjectResolutionDrift,
		UnguardedSynthesisDetected:  input.UnguardedSynthesisDetected,
		ReadyDimensions:             ready,
		MissingDimensions:           missing,
	}
	reasons := []string{}
	if !out.SubjectCorrect {
		reasons = append(reasons, "subject_not_verified")
	}
	if !out.EvidenceComplete {
		reasons = append(reasons, "evidence_incomplete")
	}
	if !out.SourceBacked {
		reasons = append(reasons, "source_not_backed")
	}
	if !out.FreshnessConfirmed {
		reasons = append(reasons, "freshness_not_confirmed")
	}
	if !out.GuardPassed {
		reasons = append(reasons, "guard_not_passed")
	}
	if !out.BoundaryOK {
		reasons = append(reasons, "answer_boundary_not_ok")
	}
	if !out.TaskConflictFree {
		reasons = append(reasons, "task_conflict_detected")
	}
	if out.SubjectResolutionDrift {
		reasons = append(reasons, "subject_resolution_drift")
	}
	if out.UnguardedSynthesisDetected {
		reasons = append(reasons, "unguarded_synthesis_detected")
	}
	if out.OverClaimDetected {
		reasons = append(reasons, "over_claim_detected")
	}
	out.Passed = len(reasons) == 0
	out.FailureReason = strings.Join(reasons, ",")
	return out
}

func requestedDimensionsReady(requested []string, ready []string) bool {
	requested = normalizeEvaluationStrings(requested)
	if len(requested) == 0 {
		return true
	}
	readySet := map[string]bool{}
	for _, item := range ready {
		readySet[strings.ToLower(item)] = true
	}
	for _, item := range requested {
		if !readySet[strings.ToLower(item)] {
			return false
		}
	}
	return true
}

func normalizeEvaluationStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
