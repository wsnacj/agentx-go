package hostkit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

func TestCompanyResearchAnswerContractBudgetsLongMultiSubjectSummary(t *testing.T) {
	subjects := make([]research.CompanyResearchPayload, 0, defaultCompanyResearchMaxSubjects+2)
	for i := 0; i < defaultCompanyResearchMaxSubjects+2; i++ {
		name := fmt.Sprintf("主体%02d", i+1)
		subjects = append(subjects, research.CompanyResearchPayload{
			Intent: research.CompanyResearchIntent{EntityName: name},
			AnswerReadiness: research.CompanyResearchAnswerReadiness{
				AnswerReady:     true,
				SafeToAnswer:    true,
				ReadyDimensions: []string{"financials", "news"},
			},
			AnswerContract: &research.CompanyResearchAnswerContract{
				FinalAnswerRecommended: true,
				FinalAnswerDraft:       name + "：" + strings.Repeat("已核验公开证据", 120),
			},
		})
	}
	contract := CompanyResearchAnswerContract(research.CompanyResearchPayload{
		Intent:   research.CompanyResearchIntent{EntityName: "多主体比较"},
		Subjects: subjects,
		AnswerReadiness: research.CompanyResearchAnswerReadiness{
			Degraded:          true,
			SafeToAnswer:      true,
			FailureCode:       "company_compare_partial",
			ReadyDimensions:   []string{"financials"},
			MissingDimensions: []string{"news"},
		},
	})
	if contract == nil || contract.SubjectBudget == nil {
		t.Fatalf("expected structured subject budget, got %#v", contract)
	}
	if contract.SubjectBudget.TotalSubjects != defaultCompanyResearchMaxSubjects+2 ||
		contract.SubjectBudget.ReturnedSubjects != defaultCompanyResearchMaxSubjects ||
		len(contract.SubjectSummaries) != defaultCompanyResearchMaxSubjects {
		t.Fatalf("unexpected subject budget readback: %#v", contract.SubjectBudget)
	}
	if !contract.Truncated {
		t.Fatalf("expected aggregate truncated readback")
	}
	for index, summary := range contract.SubjectSummaries {
		wantName := fmt.Sprintf("主体%02d", index+1)
		if summary.EntityName != wantName || !summary.Truncated || !contractSummaryMarkedTruncated(summary.Summary) {
			t.Fatalf("unexpected subject summary %d: %#v", index, summary)
		}
		content := strings.TrimSuffix(summary.Summary, contractSummaryTruncationMarker)
		if utf8.RuneCountInString(content) > defaultCompanyResearchMaxRunesPerSubject {
			t.Fatalf("subject %q exceeded rune budget: %d", summary.EntityName, utf8.RuneCountInString(content))
		}
	}
	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("marshal contract: %v", err)
	}
	text := string(encoded)
	for _, want := range []string{`"subject_budget"`, `"subject_summaries"`, `"truncated":true`} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected JSON readback %q, got %s", want, text)
		}
	}
}

func TestSingleSubjectContractDoesNotAddComparisonBudget(t *testing.T) {
	contract := readyCompanyResearchAnswerContract(research.CompanyResearchPayload{
		Intent: research.CompanyResearchIntent{EntityName: "Example"},
		AnswerReadiness: research.CompanyResearchAnswerReadiness{
			AnswerReady:     true,
			SafeToAnswer:    true,
			ReadyDimensions: []string{"financials"},
		},
	})
	if contract.SubjectBudget != nil || len(contract.SubjectSummaries) != 0 || contract.Truncated {
		t.Fatalf("expected single-subject compatibility contract, got %#v", contract)
	}
}
