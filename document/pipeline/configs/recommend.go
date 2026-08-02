package configs

import (
	"sort"
	"strings"
	"unicode"
)

type SpecRecommendation struct {
	DocType         string   `json:"doc_type,omitempty"`
	Version         string   `json:"version,omitempty"`
	ConfigDir       string   `json:"config_dir,omitempty"`
	Score           float64  `json:"score"`
	MatchedChapters []string `json:"matched_chapters,omitempty"`
	MatchedFields   []string `json:"matched_fields,omitempty"`
	MatchedKeywords []string `json:"matched_keywords,omitempty"`
}

func RecommendSpecsForText(text string, specs []*DocSpec) []SpecRecommendation {
	profile := normalizedRecommendationText(text)
	if profile.raw == "" || len(specs) == 0 {
		return nil
	}
	out := make([]SpecRecommendation, 0, len(specs))
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		rec := scoreSpecRecommendation(profile, spec)
		if rec.Score <= 0 {
			continue
		}
		out = append(out, rec)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].DocType < out[j].DocType
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func scoreSpecRecommendation(profile recommendationText, spec *DocSpec) SpecRecommendation {
	rec := SpecRecommendation{
		DocType:   strings.TrimSpace(spec.Meta.DocType),
		Version:   strings.TrimSpace(spec.Meta.Version),
		ConfigDir: strings.TrimSpace(spec.ConfigDir),
	}
	seenChapters := map[string]bool{}
	seenFields := map[string]bool{}
	seenKeywords := map[string]bool{}
	for _, chapter := range spec.Chapters {
		chapterMatched := false
		for _, keyword := range chapter.TitleKeywords {
			if recommendationTextContains(profile, keyword) {
				rec.Score += 4
				appendUniqueString(&rec.MatchedKeywords, seenKeywords, keyword)
				chapterMatched = true
			}
		}
		if chapterMatched {
			appendUniqueString(&rec.MatchedChapters, seenChapters, chapter.Key)
		}
		for _, field := range chapter.Fields {
			fieldMatched := false
			if recommendationTextContains(profile, field.Key) {
				rec.Score += 1.5
				appendUniqueString(&rec.MatchedKeywords, seenKeywords, field.Key)
				fieldMatched = true
			}
			for _, alias := range field.Aliases {
				if recommendationTextContains(profile, alias) {
					rec.Score += 1
					appendUniqueString(&rec.MatchedKeywords, seenKeywords, alias)
					fieldMatched = true
				}
			}
			if fieldMatched {
				appendUniqueString(&rec.MatchedFields, seenFields, field.Key)
			}
		}
	}
	return rec
}

type recommendationText struct {
	raw     string
	compact string
}

func normalizedRecommendationText(text string) recommendationText {
	raw := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
	return recommendationText{raw: raw, compact: compactRecommendationText(raw)}
}

func recommendationTextContains(profile recommendationText, keyword string) bool {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return false
	}
	rawKeyword := strings.ToLower(strings.Join(strings.Fields(keyword), " "))
	if rawKeyword != "" && strings.Contains(profile.raw, rawKeyword) {
		return true
	}
	compactKeyword := compactRecommendationText(rawKeyword)
	return compactKeyword != "" && strings.Contains(profile.compact, compactKeyword)
}

func compactRecommendationText(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func appendUniqueString(out *[]string, seen map[string]bool, value string) {
	value = strings.TrimSpace(value)
	if value == "" || seen[value] {
		return
	}
	seen[value] = true
	*out = append(*out, value)
}
