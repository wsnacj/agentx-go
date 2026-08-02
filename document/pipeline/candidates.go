package pipeline

import (
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"regexp"
	"strings"
	"unicode"
)

const (
	fieldWarningCandidateConflict = "candidate_conflict"
	fieldWarningLowConfidence     = "low_confidence"
)

var policyYearRe = regexp.MustCompile(`(?:19|20)\d{2}`)

func selectFieldCandidate(field configs.FieldSpec, chapterKey string, candidates []types.FieldCandidate) (types.FieldResult, bool) {
	out := types.FieldResult{Key: field.Key}
	if len(candidates) == 0 {
		return out, false
	}
	normalized := make([]types.FieldCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Value == nil {
			continue
		}
		if candidate.Score == 0 {
			candidate.Score = candidate.Confidence
		}
		if candidate.NormalizedValue == nil {
			candidate.NormalizedValue = candidate.Value
		}
		if strings.TrimSpace(candidate.Source) == "" {
			candidate.Source = candidate.Extractor
		}
		if strings.TrimSpace(candidate.Chapter) == "" {
			candidate.Chapter = chapterKey
		}
		if !applyFieldPolicy(&candidate, field, chapterKey, candidates) {
			continue
		}
		if strings.TrimSpace(candidate.Unit) == "" {
			candidate.Unit = strings.TrimSpace(field.Unit)
		}
		normalized = append(normalized, candidate)
	}
	if len(normalized) == 0 {
		return out, false
	}
	calibrateFieldCandidateEnsemble(normalized)
	selectedIndex := 0
	for i := 1; i < len(normalized); i++ {
		if normalized[i].Score > normalized[selectedIndex].Score {
			selectedIndex = i
		}
	}
	conflict := fieldCandidatesConflict(normalized)
	selectionReason := "selected_only_candidate"
	if len(normalized) > 1 {
		selectionReason = "selected_highest_score"
		if fieldCandidatesShareTopScore(normalized, normalized[selectedIndex].Score) {
			selectionReason = "selected_stable_tie"
		}
	}
	if fieldCandidateAgreementSourceCount(normalized[selectedIndex], normalized) > 1 {
		selectionReason += "_with_ensemble_agreement"
	}
	warnings := compactStrings(normalized[selectedIndex].Warnings)
	if conflict {
		warnings = append(warnings, fieldWarningCandidateConflict)
		selectionReason += "_with_conflict"
	}
	if normalized[selectedIndex].Confidence > 0 && normalized[selectedIndex].Confidence < 0.5 {
		warnings = append(warnings, fieldWarningLowConfidence)
	}
	warnings = compactStrings(warnings)
	for i := range normalized {
		normalized[i].Selected = i == selectedIndex
		if i == selectedIndex {
			normalized[i].SelectionReason = selectionReason
		}
	}
	selected := normalized[selectedIndex]
	out.Chapter = chapterKey
	out.Value = selected.Value
	out.RawValue = selected.RawValue
	out.NormalizedValue = selected.NormalizedValue
	out.Source = selected.Source
	out.Confidence = selected.Confidence
	out.Evidence = selected.Evidence
	out.Unit = selected.Unit
	out.Currency = selected.Currency
	out.Period = selected.Period
	out.PageRefs = append([]int{}, selected.PageRefs...)
	out.BoundingBoxes = append([]types.BoundingBox{}, selected.BoundingBoxes...)
	out.TableCells = append([]types.TableCellRef{}, selected.TableCells...)
	out.Warnings = warnings
	out.ReviewRequired = len(warnings) > 0
	out.SelectionReason = selectionReason
	out.Candidates = normalized
	return out, true
}

func calibrateFieldCandidateEnsemble(candidates []types.FieldCandidate) {
	sourceCounts := map[string]map[string]bool{}
	for _, candidate := range candidates {
		key := fieldCandidateValueKey(candidate)
		if key == "" {
			continue
		}
		source := fieldCandidateSourceKey(candidate)
		if source == "" {
			continue
		}
		if sourceCounts[key] == nil {
			sourceCounts[key] = map[string]bool{}
		}
		sourceCounts[key][source] = true
	}
	for i := range candidates {
		key := fieldCandidateValueKey(candidates[i])
		sources := sourceCounts[key]
		if len(sources) <= 1 {
			continue
		}
		boost := 0.10 * float64(len(sources)-1)
		if boost > 0.20 {
			boost = 0.20
		}
		candidates[i].Score += boost
	}
}

func fieldCandidateAgreementSourceCount(candidate types.FieldCandidate, candidates []types.FieldCandidate) int {
	key := fieldCandidateValueKey(candidate)
	if key == "" {
		return 0
	}
	seen := map[string]bool{}
	for _, other := range candidates {
		if fieldCandidateValueKey(other) != key {
			continue
		}
		source := fieldCandidateSourceKey(other)
		if source == "" {
			continue
		}
		seen[source] = true
	}
	return len(seen)
}

func fieldCandidateValueKey(candidate types.FieldCandidate) string {
	value := candidate.NormalizedValue
	if value == nil {
		value = candidate.Value
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func fieldCandidateSourceKey(candidate types.FieldCandidate) string {
	source := strings.TrimSpace(candidate.Source)
	extractor := strings.TrimSpace(candidate.Extractor)
	switch {
	case source != "" && extractor != "":
		return source + "/" + extractor
	case source != "":
		return source
	default:
		return extractor
	}
}

func applyFieldPolicy(candidate *types.FieldCandidate, field configs.FieldSpec, chapterKey string, candidates []types.FieldCandidate) bool {
	if candidate == nil {
		return false
	}
	text := fieldCandidatePolicyText(*candidate)
	if matchesAnyPolicyPattern(field.DisallowPatterns, text) {
		return false
	}
	sourceUnit := strings.TrimSpace(candidate.Unit)
	switch strings.ToLower(strings.TrimSpace(field.UnitPolicy)) {
	case "required":
		if sourceUnit == "" {
			return false
		}
	case "prefer":
		if sourceUnit != "" {
			candidate.Score += 0.05
		}
	}
	if matchesAnyPolicyLabel(candidate.RowLabel, field.Aliases) ||
		matchesAnyPolicyLabel(candidate.Evidence, field.Aliases) ||
		matchesAnyPolicyLabel(fmt.Sprint(candidate.RawValue), field.Aliases) {
		candidate.Score += 0.08
	}
	if matchesAnyPolicyPattern(field.PreferPatterns, text) {
		candidate.Score += 0.12
	}
	if strings.EqualFold(strings.TrimSpace(field.PeriodPolicy), "current") {
		scorePeriodPolicy(candidate, candidates)
	}
	if matchesAnyPolicyLabel(chapterKey, field.PreferredChapters) {
		candidate.Score += 0.05
	}
	return true
}

func scorePeriodPolicy(candidate *types.FieldCandidate, candidates []types.FieldCandidate) {
	year, hasYear := candidatePolicyYear(*candidate)
	maxYear := 0
	for _, other := range candidates {
		if y, ok := candidatePolicyYear(other); ok && y > maxYear {
			maxYear = y
		}
	}
	if hasYear && maxYear > 0 {
		if year == maxYear {
			candidate.Score += 0.16
		} else if year < maxYear {
			candidate.Score -= 0.08
		}
		return
	}
	text := strings.ToLower(fieldCandidatePolicyText(*candidate))
	if strings.Contains(text, "current") ||
		strings.Contains(text, "本期") ||
		strings.Contains(text, "本年") ||
		strings.Contains(text, "本年度") ||
		strings.Contains(text, "报告期") ||
		strings.Contains(text, "報告期") {
		candidate.Score += 0.16
	}
}

func candidatePolicyYear(candidate types.FieldCandidate) (int, bool) {
	for _, value := range []string{
		candidate.Period,
		candidate.ColumnLabel,
		candidate.Evidence,
	} {
		if year, ok := firstPolicyYear(value); ok {
			return year, true
		}
	}
	return 0, false
}

func firstPolicyYear(value string) (int, bool) {
	m := policyYearRe.FindString(value)
	if m == "" {
		return 0, false
	}
	year := 0
	for _, r := range m {
		year = year*10 + int(r-'0')
	}
	return year, true
}

func fieldCandidatePolicyText(candidate types.FieldCandidate) string {
	return strings.Join([]string{
		fmt.Sprint(candidate.Value),
		fmt.Sprint(candidate.RawValue),
		candidate.Evidence,
		candidate.RowLabel,
		candidate.ColumnLabel,
		candidate.Unit,
		candidate.UnitSource,
		candidate.Period,
		candidate.Source,
		candidate.Extractor,
	}, "\n")
}

func matchesAnyPolicyPattern(patterns []string, text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err == nil {
			if re.MatchString(text) {
				return true
			}
			continue
		}
		if strings.Contains(strings.ToLower(text), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func matchesAnyPolicyLabel(value string, labels []string) bool {
	valueKey := compactPolicyLabel(value)
	if valueKey == "" {
		return false
	}
	for _, label := range labels {
		labelKey := compactPolicyLabel(label)
		if labelKey == "" {
			continue
		}
		if valueKey == labelKey {
			return true
		}
		if containsCJKPolicyLabel(label) || strings.ContainsAny(label, " \t-_/") {
			if strings.Contains(valueKey, labelKey) {
				return true
			}
		}
	}
	return false
}

func compactPolicyLabel(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func containsCJKPolicyLabel(value string) bool {
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func fieldCandidatesShareTopScore(candidates []types.FieldCandidate, score float64) bool {
	matches := 0
	for _, candidate := range candidates {
		if candidate.Score == score {
			matches++
		}
	}
	return matches > 1
}

func fieldCandidatesConflict(candidates []types.FieldCandidate) bool {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		value := candidate.NormalizedValue
		if value == nil {
			value = candidate.Value
		}
		key := strings.TrimSpace(fmt.Sprint(value))
		if key == "" {
			continue
		}
		seen[key] = true
		if len(seen) > 1 {
			return true
		}
	}
	return false
}

func hasString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
