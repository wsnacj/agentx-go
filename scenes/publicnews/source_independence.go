package publicnews

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	sourceIndependenceMinContentRunes       = 120
	sourceIndependenceBriefMaxContentRunes  = 600
	sourceIndependenceShingleRunes          = 12
	sourceIndependenceLongShingleStride     = 3
	sourceIndependenceBriefShingleStride    = 1
	sourceIndependenceStrongContainment     = 0.90
	sourceIndependenceLongContainment       = 0.60
	sourceIndependenceBriefContainment      = 0.33
	sourceIndependenceMixedContainment      = 0.30
	sourceIndependenceSameHeadlineThreshold = 0.70
)

var (
	attributedChineseReportPattern        = regexp.MustCompile(`(?:据|援引)[^，。；\n]{1,48}(?:报道|消息|披露)`)
	attributedChinesePublisherLeadPattern = regexp.MustCompile(`(?:《[^》]{2,40}》|[^，。；\n]{2,40}(?:时报|日报|通讯社|新闻社|媒体))(?:引述|援引)[^，。；\n]{0,48}(?:报道|消息|称)`)
	attributedEnglishReportPattern        = regexp.MustCompile(`(?i)(?:according to|as reported by|citing)\s+[^.,;\n]{2,64}|\b(?:reuters|bloomberg|associated press|financial times|wall street journal)\s+(?:reported|said)`)
	attributedPublisherAppLeadPattern     = regexp.MustCompile(`(?i)(?:^|[\n。！？；;])\s*([^\n。！？；;：:,，]{2,32})\s*(?:app|网)\s*(?:讯|消息)\s*[：:,，]`)
	upstreamPublisherAttributionPatterns  = map[string]*regexp.Regexp{
		"reuters":          regexp.MustCompile(`(?i)(?:据|援引|引述)\s*路透(?:社)?|路透(?:社)?[^，。；\n]{0,32}(?:报道|消息|调查|援引|引述)|(?:对|向)\s*路透(?:社)?\s*(?:表示|称|透露)|(?:according to|reported by|citing|told)\s+reuters|reuters\s+(?:reported|poll|survey|said|cited)`),
		"bloomberg":        regexp.MustCompile(`(?i)(?:据|援引|引述)\s*彭博(?:社)?|彭博(?:社)?[^，。；\n]{0,32}(?:报道|消息|调查|援引|引述)|(?:对|向)\s*彭博(?:社)?\s*(?:表示|称|透露)|(?:according to|reported by|citing|told)\s+bloomberg|bloomberg\s+(?:reported|poll|survey|said|cited)`),
		"associated_press": regexp.MustCompile(`(?i)(?:据|援引|引述)\s*美联社|美联社[^，。；\n]{0,32}(?:报道|消息|调查|援引|引述)|(?:according to|reported by|citing|told)\s+(?:the\s+)?associated press|(?:associated press|\bap\b)\s+(?:reported|said|cited)`),
		"afp":              regexp.MustCompile(`(?i)(?:据|援引|引述)\s*法新社|法新社[^，。；\n]{0,32}(?:报道|消息|调查|援引|引述)|(?:according to|reported by|citing|told)\s+(?:afp|agence france-presse)|(?:afp|agence france-presse)\s+(?:reported|said|cited)`),
		"xinhua":           regexp.MustCompile(`(?i)(?:据|援引|引述)\s*新华社|新华社[^，。；\n]{0,32}(?:报道|消息|调查|援引|引述)|(?:according to|reported by|citing)\s+xinhua|xinhua\s+(?:reported|said|cited)`),
	}
)

type crossCheckSourceStats struct {
	uniqueCount                  int
	duplicatePublisherCount      int
	syndicatedCopyCount          int
	attributedRepublicationCount int
	sharedUpstreamPublisherCount int
}

// LatestNewsSourcesLookLikeSyndicatedCopy reports whether two opened source
// bodies carry enough copied text to be treated as one editorial source.
func LatestNewsSourcesLookLikeSyndicatedCopy(left LatestNewsLookupSource, right LatestNewsLookupSource) bool {
	if evidenceLooksLikeAttributedRepublication(Evidence{groundedText: strings.TrimSpace(right.Text)}) {
		return true
	}
	return evidenceLooksLikeSyndicatedCopy(
		Evidence{
			Headline:     firstNonEmpty(left.Title, left.Headline),
			groundedText: strings.TrimSpace(left.Text),
		},
		Evidence{
			Headline:     firstNonEmpty(right.Title, right.Headline),
			groundedText: strings.TrimSpace(right.Text),
		},
	)
}

// LatestNewsSourcesShareUpstreamPublisher reports whether two pages both rely
// on the same attributed wire service strongly enough to represent one
// editorial lineage rather than two independent confirmations.
func LatestNewsSourcesShareUpstreamPublisher(left LatestNewsLookupSource, right LatestNewsLookupSource) bool {
	return evidenceSharesUpstreamPublisher(
		Evidence{groundedText: strings.TrimSpace(left.Text)},
		Evidence{groundedText: strings.TrimSpace(right.Text)},
	)
}

func evaluateCrossCheckSources(primary Evidence, supporting []Evidence) crossCheckSourceStats {
	stats := crossCheckSourceStats{}
	seenSites := map[string]bool{}
	accepted := []Evidence{}
	add := func(evidence Evidence, supportingSource bool) {
		if !EvidenceUsableForCrossCheck(evidence) {
			return
		}
		if supportingSource && evidenceLooksLikeAttributedRepublication(evidence) {
			stats.attributedRepublicationCount++
			return
		}
		siteKey := CrossCheckSourceKey(evidence)
		if siteKey == "" {
			return
		}
		if seenSites[siteKey] {
			if supportingSource {
				stats.duplicatePublisherCount++
			}
			return
		}
		for _, existing := range accepted {
			if evidenceSharesUpstreamPublisher(existing, evidence) {
				if supportingSource {
					stats.sharedUpstreamPublisherCount++
				}
				return
			}
			if evidenceLooksLikeSyndicatedCopy(existing, evidence) {
				if supportingSource {
					stats.syndicatedCopyCount++
				}
				return
			}
		}
		seenSites[siteKey] = true
		accepted = append(accepted, evidence)
		stats.uniqueCount++
	}
	add(primary, false)
	for _, evidence := range supporting {
		add(evidence, true)
	}
	return stats
}

func evidenceSharesUpstreamPublisher(left Evidence, right Evidence) bool {
	if evidenceHasOriginalReportingSignals(left.groundedText) || evidenceHasOriginalReportingSignals(right.groundedText) {
		return false
	}
	leftPublisherLeads := attributedPublisherLeadKeys(left.groundedText)
	rightPublisherLeads := attributedPublisherLeadKeys(right.groundedText)
	for publisher := range leftPublisherLeads {
		if rightPublisherLeads[publisher] {
			return true
		}
	}
	leftCounts := attributedUpstreamPublisherCounts(left.groundedText)
	rightCounts := attributedUpstreamPublisherCounts(right.groundedText)
	for publisher, leftCount := range leftCounts {
		rightCount := rightCounts[publisher]
		if leftCount > 0 && rightCount > 0 && leftCount+rightCount >= 3 {
			return true
		}
	}
	return false
}

func attributedPublisherLeadKeys(text string) map[string]bool {
	keys := map[string]bool{}
	for _, match := range attributedPublisherAppLeadPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		key := string(normalizedArticleContent(match[1]))
		if len([]rune(key)) >= 2 {
			keys[key] = true
		}
	}
	return keys
}

func attributedUpstreamPublisherCounts(text string) map[string]int {
	counts := map[string]int{}
	for publisher, pattern := range upstreamPublisherAttributionPatterns {
		if count := len(pattern.FindAllStringIndex(text, -1)); count > 0 {
			counts[publisher] = count
		}
	}
	return counts
}

func evidenceHasOriginalReportingSignals(text string) bool {
	return containsAnyFold(strings.ToLower(text),
		"本报记者", "记者采访", "记者从", "独家获悉", "记者查阅",
		"our reporter", "we interviewed", "our reporting", "documents reviewed by",
	)
}

func evidenceLooksLikeAttributedRepublication(evidence Evidence) bool {
	text := strings.TrimSpace(evidence.groundedText)
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	if containsAnyFold(lower,
		"转载来源", "转载自", "转自", "原文来源", "文章来源",
		"republished from", "reprinted from", "syndicated from",
	) {
		return true
	}
	if len(normalizedArticleContent(text)) > sourceIndependenceBriefMaxContentRunes {
		return false
	}
	if evidenceHasOriginalReportingSignals(text) {
		return false
	}
	return attributedChineseReportPattern.MatchString(text) ||
		attributedChinesePublisherLeadPattern.MatchString(text) ||
		attributedEnglishReportPattern.MatchString(text)
}

func evidenceLooksLikeSyndicatedCopy(left Evidence, right Evidence) bool {
	leftContent := normalizedArticleContent(stripSourceIndependenceBoilerplate(left.groundedText))
	rightContent := normalizedArticleContent(stripSourceIndependenceBoilerplate(right.groundedText))
	if len(leftContent) < sourceIndependenceMinContentRunes || len(rightContent) < sourceIndependenceMinContentRunes {
		return false
	}
	containment := articleShingleContainment(leftContent, rightContent)
	if containment >= sourceIndependenceStrongContainment {
		return true
	}
	if len(leftContent) > sourceIndependenceBriefMaxContentRunes &&
		len(rightContent) > sourceIndependenceBriefMaxContentRunes &&
		containment >= sourceIndependenceLongContainment {
		return true
	}
	if len(leftContent) <= sourceIndependenceBriefMaxContentRunes &&
		len(rightContent) <= sourceIndependenceBriefMaxContentRunes &&
		containment >= sourceIndependenceBriefContainment {
		return true
	}
	if (len(leftContent) <= sourceIndependenceBriefMaxContentRunes) !=
		(len(rightContent) <= sourceIndependenceBriefMaxContentRunes) &&
		containment >= sourceIndependenceMixedContainment {
		return true
	}
	return normalizedHeadline(left.Headline) == normalizedHeadline(right.Headline) &&
		normalizedHeadline(left.Headline) != "" &&
		containment >= sourceIndependenceSameHeadlineThreshold
}

func stripSourceIndependenceBoilerplate(value string) string {
	kept := make([]string, 0)
	for _, line := range strings.Split(value, "\n") {
		lower := strings.ToLower(strings.TrimSpace(line))
		if lower == "" {
			continue
		}
		if containsAnyFold(lower,
			"特别声明", "免责声明", "本平台仅提供信息存储空间", "使用前请核实",
			"notice: the content above", "disclaimer:", "provides information storage space",
		) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func normalizedArticleContent(value string) []rune {
	out := make([]rune, 0, len(value))
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out = append(out, r)
		}
	}
	return out
}

func normalizedHeadline(value string) string {
	return string(normalizedArticleContent(value))
}

func articleShingleContainment(left []rune, right []rune) float64 {
	stride := sourceIndependenceLongShingleStride
	if len(left) <= sourceIndependenceBriefMaxContentRunes || len(right) <= sourceIndependenceBriefMaxContentRunes {
		stride = sourceIndependenceBriefShingleStride
	}
	leftShingles := articleShingles(left, stride)
	rightShingles := articleShingles(right, stride)
	if len(leftShingles) == 0 || len(rightShingles) == 0 {
		return 0
	}
	intersection := 0
	for shingle := range leftShingles {
		if rightShingles[shingle] {
			intersection++
		}
	}
	denominator := len(leftShingles)
	if len(rightShingles) < denominator {
		denominator = len(rightShingles)
	}
	return float64(intersection) / float64(denominator)
}

func articleShingles(content []rune, stride int) map[string]bool {
	if len(content) < sourceIndependenceShingleRunes {
		return nil
	}
	if stride <= 0 {
		stride = 1
	}
	out := map[string]bool{}
	lastStart := len(content) - sourceIndependenceShingleRunes
	for start := 0; start <= lastStart; start += stride {
		out[string(content[start:start+sourceIndependenceShingleRunes])] = true
	}
	if lastStart%stride != 0 {
		out[string(content[lastStart:])] = true
	}
	return out
}
