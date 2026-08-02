package pipeline

import (
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"sort"
	"strings"
	"time"
)

func newDocumentDiagnostics() *types.DocumentDiagnostics {
	return &types.DocumentDiagnostics{
		Status:   "running",
		Sections: map[string]types.SectionDiagnostic{},
		Fields:   map[string]types.FieldDiagnostic{},
	}
}

func startDiagnosticStage(diag *types.DocumentDiagnostics, name string) func(status string, err error, warnings ...string) {
	start := time.Now()
	return func(status string, err error, warnings ...string) {
		if diag == nil {
			return
		}
		entry := types.StageDiagnostic{
			Name:       name,
			Status:     firstNonEmptyString(status, "completed"),
			DurationMS: time.Since(start).Milliseconds(),
			Warnings:   compactStrings(warnings),
		}
		if err != nil {
			entry.Error = err.Error()
			if status == "" {
				entry.Status = "failed"
			}
		}
		diag.Stages = append(diag.Stages, entry)
	}
}

func documentTextQuality(pages []string) string {
	if len(pages) == 0 {
		return "no_pages"
	}
	total := 0
	nonEmpty := 0
	for _, page := range pages {
		n := len([]rune(strings.TrimSpace(page)))
		total += n
		if n > 0 {
			nonEmpty++
		}
	}
	if total == 0 {
		return "no_text"
	}
	if nonEmpty == 0 || total/len(pages) < 30 {
		return "sparse_text"
	}
	return "ok"
}

func setSectionDiagnostic(diag *types.DocumentDiagnostics, entry types.SectionDiagnostic) {
	if diag == nil || strings.TrimSpace(entry.Key) == "" {
		return
	}
	if diag.Sections == nil {
		diag.Sections = map[string]types.SectionDiagnostic{}
	}
	if entry.Status == "" {
		entry.Status = "unknown"
	}
	diag.Sections[entry.Key] = entry
}

func setFieldDiagnostic(diag *types.DocumentDiagnostics, entry types.FieldDiagnostic) {
	if diag == nil || strings.TrimSpace(entry.Chapter) == "" || strings.TrimSpace(entry.Field) == "" {
		return
	}
	if diag.Fields == nil {
		diag.Fields = map[string]types.FieldDiagnostic{}
	}
	if entry.Status == "" {
		entry.Status = "unknown"
	}
	diag.Fields[fieldDiagnosticID(entry.Chapter, entry.Field)] = entry
}

func fieldDiagnosticID(chapter string, field string) string {
	return fmt.Sprintf("%s.%s", strings.TrimSpace(chapter), strings.TrimSpace(field))
}

func matchedFieldDiagnostic(chapter string, field string, result types.FieldResult, normalizationWarning string, warnings []string) types.FieldDiagnostic {
	selected, hasSelected := selectedFieldCandidate(result.Candidates)
	agreementSourceCount := 0
	if hasSelected {
		agreementSourceCount = fieldCandidateAgreementSourceCount(selected, result.Candidates)
	}
	return types.FieldDiagnostic{
		Chapter:              chapter,
		Field:                field,
		Status:               "matched",
		Source:               result.Source,
		NormalizationWarning: strings.TrimSpace(normalizationWarning),
		CandidateCount:       len(result.Candidates),
		CandidateValueCount:  fieldCandidateValueCount(result.Candidates),
		CandidateSources:     fieldCandidateSources(result.Candidates),
		AgreementSourceCount: agreementSourceCount,
		SelectionReason:      result.SelectionReason,
		ReviewRequired:       result.ReviewRequired,
		Warnings:             compactStrings(warnings),
	}
}

func selectedFieldCandidate(candidates []types.FieldCandidate) (types.FieldCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.Selected {
			return candidate, true
		}
	}
	return types.FieldCandidate{}, false
}

func fieldCandidateSources(candidates []types.FieldCandidate) []string {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		source := fieldCandidateSourceKey(candidate)
		if source == "" {
			continue
		}
		seen[source] = true
	}
	out := make([]string, 0, len(seen))
	for source := range seen {
		out = append(out, source)
	}
	sort.Strings(out)
	return out
}

func fieldCandidateValueCount(candidates []types.FieldCandidate) int {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		value := fieldCandidateValueKey(candidate)
		if value == "" {
			continue
		}
		seen[value] = true
	}
	return len(seen)
}

func finishDocumentDiagnostics(diag *types.DocumentDiagnostics) {
	if diag == nil {
		return
	}
	status := "completed"
	for _, stage := range diag.Stages {
		if stage.Status == "failed" {
			status = "completed_with_warnings"
			break
		}
	}
	for _, section := range diag.Sections {
		if section.Status == "missing" {
			status = "completed_with_warnings"
			break
		}
	}
	for _, field := range diag.Fields {
		if field.Status == "missing" || field.ReviewRequired {
			status = "completed_with_warnings"
			break
		}
	}
	diag.Status = status
}

func compactStrings(values []string) []string {
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

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
