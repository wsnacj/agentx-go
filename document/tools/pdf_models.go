package tools

import (
	"sort"
	"strings"
)

type pdfModelResolverConfig struct {
	PreferredModel string
	FallbackModels []string
	PreferNative   *bool
	Candidates     []pdfVisionModelCandidate
}

func newPDFModelResolverConfig(opts PDFToolOptions) pdfModelResolverConfig {
	return pdfModelResolverConfig{
		PreferredModel: strings.TrimSpace(opts.PreferredModel),
		FallbackModels: append([]string(nil), opts.FallbackModels...),
		PreferNative:   opts.PreferNative,
		Candidates:     append([]pdfVisionModelCandidate(nil), opts.Models...),
	}
}

func rankPDFPromptCandidates(mode string, candidates []pdfVisionModelCandidate, resolver pdfModelResolverConfig) []pdfVisionModelCandidate {
	if len(candidates) <= 1 {
		return append([]pdfVisionModelCandidate(nil), candidates...)
	}
	preferred := preferredPDFVisionClients(mode)
	order := make(map[string]int, len(preferred))
	for idx, client := range preferred {
		order[client] = idx
	}
	ranked := append([]pdfVisionModelCandidate(nil), candidates...)
	sortByPromptRank := func(candidate pdfVisionModelCandidate) int {
		score := 0
		if idx, ok := order[pdfVisionCandidateRoutingKey(candidate)]; ok {
			score += 100 - idx
		}
		if candidate.NativePDF {
			score += 50
		}
		if candidate.SupportsVision {
			score += 20
		}
		return score
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		left := sortByPromptRank(ranked[i])
		right := sortByPromptRank(ranked[j])
		if left != right {
			return left > right
		}
		if ranked[i].Client != ranked[j].Client {
			return ranked[i].Client < ranked[j].Client
		}
		return ranked[i].Name < ranked[j].Name
	})
	return applyPDFConfiguredCandidateOrder(ranked, resolver)
}

func rankPDFVisionCandidates(mode string, candidates []pdfVisionModelCandidate, resolver pdfModelResolverConfig) []pdfVisionModelCandidate {
	if len(candidates) <= 1 {
		return append([]pdfVisionModelCandidate(nil), candidates...)
	}
	preferred := preferredPDFVisionClients(mode)
	order := make(map[string]int, len(preferred))
	for i, client := range preferred {
		order[client] = i
	}
	ranked := append([]pdfVisionModelCandidate(nil), candidates...)
	sort.SliceStable(ranked, func(i, j int) bool {
		left := pdfVisionCandidateRoutingKey(ranked[i])
		right := pdfVisionCandidateRoutingKey(ranked[j])
		li, lok := order[left]
		ri, rok := order[right]
		switch {
		case lok && rok && li != ri:
			return li < ri
		case lok != rok:
			return lok
		case left != right:
			return left < right
		default:
			return ranked[i].Name < ranked[j].Name
		}
	})
	return applyPDFConfiguredCandidateOrder(ranked, resolver)
}

func applyPDFConfiguredCandidateOrder(candidates []pdfVisionModelCandidate, resolver pdfModelResolverConfig) []pdfVisionModelCandidate {
	ordered := append([]pdfVisionModelCandidate(nil), candidates...)
	if resolver.PreferNative != nil {
		native := make([]pdfVisionModelCandidate, 0, len(ordered))
		nonNative := make([]pdfVisionModelCandidate, 0, len(ordered))
		for _, candidate := range ordered {
			if candidate.NativePDF {
				native = append(native, candidate)
			} else {
				nonNative = append(nonNative, candidate)
			}
		}
		if *resolver.PreferNative {
			ordered = append(native, nonNative...)
		} else {
			ordered = append(nonNative, native...)
		}
	}
	explicit := normalizedPDFConfiguredModelOrder(resolver)
	if len(explicit) == 0 {
		return ordered
	}
	selected := make([]pdfVisionModelCandidate, 0, len(explicit))
	used := make([]bool, len(ordered))
	for _, item := range explicit {
		for idx, candidate := range ordered {
			if used[idx] || !pdfVisionCandidateMatchesConfiguredName(candidate, item) {
				continue
			}
			selected = append(selected, candidate)
			used[idx] = true
			break
		}
	}
	return selected
}

func normalizedPDFConfiguredModelOrder(resolver pdfModelResolverConfig) []string {
	items := make([]string, 0, 1+len(resolver.FallbackModels))
	if trimmed := strings.TrimSpace(resolver.PreferredModel); trimmed != "" {
		items = append(items, trimmed)
	}
	items = append(items, resolver.FallbackModels...)
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		key := strings.ToLower(trimmed)
		if trimmed == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func pdfConfiguredModelRoutePresent(resolver pdfModelResolverConfig) bool {
	return len(normalizedPDFConfiguredModelOrder(resolver)) > 0
}

func pdfVisionCandidateMatchesConfiguredName(candidate pdfVisionModelCandidate, target string) bool {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(candidate.ConfigKey), trimmed) ||
		strings.EqualFold(strings.TrimSpace(candidate.Name), trimmed) ||
		strings.EqualFold(strings.TrimSpace(candidate.Model), trimmed)
}

func applyPDFModelResolverToPlan(plan pdfAnalysisPlan, resolver pdfModelResolverConfig) pdfAnalysisPlan {
	if !plan.NeedsVision {
		return plan
	}
	plan.PreferredClients = preferredPDFVisionClients(plan.Mode)
	plan.CandidateModels = rankPDFVisionCandidates(plan.Mode, resolver.Candidates, resolver)
	plan.ProviderRouting = buildPDFProviderRouting(plan.Mode, plan.PreferredClients, plan.CandidateModels)
	plan.NativeProviderRouting = buildPDFNativeProviderRouting(plan.Mode, plan.CandidateModels)
	if len(plan.CandidateModels) == 0 {
		plan.Warning = "no vision-capable llmx submodel is currently configured; fall back to text extraction or add a vision model"
	} else if strings.TrimSpace(plan.PreferredBackend) == "" {
		plan.PreferredBackend = plan.CandidateModels[0].Name
	}
	return plan
}
