package pipeline

import (
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"sort"
	"strings"
)

func llmFieldEvidence(chapterKey string, rawValue any) string {
	parts := []string{}
	if chapterKey = strings.TrimSpace(chapterKey); chapterKey != "" {
		parts = append(parts, "chapter="+chapterKey)
	}
	value := strings.TrimSpace(fmt.Sprint(rawValue))
	if value != "" && value != "<nil>" {
		if len([]rune(value)) > 240 {
			value = string([]rune(value)[:240]) + "..."
		}
		parts = append(parts, "value="+value)
	}
	return strings.Join(parts, "\n")
}

func pageRefsForChapter(diag *types.DocumentDiagnostics, chapterKey string, pages []string, chapterText string) []int {
	if diag != nil {
		if section, ok := diag.Sections[strings.TrimSpace(chapterKey)]; ok && len(section.CandidatePages) > 0 {
			return append([]int{}, section.CandidatePages...)
		}
	}
	return pageRefsForEvidence(pages, chapterText)
}

func pageRefsForSectionPages(pages []string, sectionPages []string) []int {
	out := []int{}
	seen := map[int]bool{}
	for _, sectionPage := range sectionPages {
		for _, pageRef := range pageRefsForEvidence(pages, sectionPage) {
			if pageRef <= 0 || seen[pageRef] {
				continue
			}
			seen[pageRef] = true
			out = append(out, pageRef)
			if len(out) >= 20 {
				return out
			}
		}
	}
	return out
}

func pageRefsForEvidence(pages []string, evidence string) []int {
	needles := evidenceNeedles(evidence)
	if len(needles) == 0 {
		return nil
	}
	seen := map[int]bool{}
	out := []int{}
	for _, needle := range needles {
		needle = normalizeEvidenceWhitespace(needle)
		if needle == "" {
			continue
		}
		for i, page := range pages {
			pageRef := i + 1
			if seen[pageRef] {
				continue
			}
			if strings.Contains(normalizeEvidenceWhitespace(page), needle) {
				seen[pageRef] = true
				out = append(out, pageRef)
				if len(out) >= 5 {
					sort.Ints(out)
					return out
				}
			}
		}
	}
	sort.Ints(out)
	return out
}

func bestEvidenceNeedle(evidence string) string {
	best := ""
	for _, line := range strings.Split(evidence, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len([]rune(line)) > len([]rune(best)) {
			best = line
		}
	}
	return best
}

func evidenceNeedles(evidence string) []string {
	best := bestEvidenceNeedle(evidence)
	if best == "" {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, line := range strings.Split(evidence, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		needle := normalizeEvidenceWhitespace(line)
		if needle == "" || seen[needle] {
			continue
		}
		seen[needle] = true
		out = append(out, needle)
	}
	if len(out) == 0 {
		return []string{best}
	}
	return out
}

func normalizeEvidenceWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
