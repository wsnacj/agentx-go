package publicsource

import "strings"

// Evaluation is the deterministic read-only evidence gate for the Domain Kit.
type Evaluation struct {
	Passed                  bool     `json:"passed"`
	EvidenceObserved        bool     `json:"evidence_observed"`
	DisplaySummaryObserved  bool     `json:"display_summary_observed"`
	AttestedSummaryObserved bool     `json:"attested_summary_observed"`
	RawOutputAbsent         bool     `json:"raw_output_absent"`
	FailureReasons          []string `json:"failure_reasons,omitempty"`
	Summary                 string   `json:"summary,omitempty"`
}

func Evaluate(report Report, requireAttestation bool) Evaluation {
	report = report.Normalize()
	result := Evaluation{EvidenceObserved: len(report.Evidence) > 0, DisplaySummaryObserved: len(report.DisplaySummaries) > 0, RawOutputAbsent: !report.RawOutputLoaded}
	for _, summary := range report.DisplaySummaries {
		if summaryStrength(summary) != "weak" {
			result.AttestedSummaryObserved = true
			break
		}
	}
	if !result.EvidenceObserved {
		result.FailureReasons = append(result.FailureReasons, "public_source_evidence_missing")
	}
	if !result.DisplaySummaryObserved {
		result.FailureReasons = append(result.FailureReasons, "public_source_display_summary_missing")
	}
	if requireAttestation && !result.AttestedSummaryObserved {
		result.FailureReasons = append(result.FailureReasons, "public_source_display_summary_attestation_missing")
	}
	if !result.RawOutputAbsent {
		result.FailureReasons = append(result.FailureReasons, "public_source_raw_output_not_allowed")
	}
	result.Passed = report.Status == "satisfied" && len(result.FailureReasons) == 0
	if result.Passed {
		result.Summary = "public source evidence verified"
	} else {
		result.Summary = strings.Join(result.FailureReasons, ",")
	}
	return result
}
