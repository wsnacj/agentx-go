package publicnews

import (
	"regexp"
	"strings"
	"unicode"
)

var latestNewsTopicPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:帮我找下|帮我查下|帮我看下|帮我找|帮我看|找下|查下|看下|看一下|查询|了解|看看|关注)([^，。,.；;：:\n]+?)(?:的?最新新闻|的?最新进展|的?最新消息|新闻|进展|消息)`),
	regexp.MustCompile(`([^，。,.；;：:\n]+?)(?:的?最新新闻|的?最新进展|的?最新消息)`),
}

// BuildLatestNewsBriefCaseInput is a legacy/explicit pack-workflow fallback for
// materializing case inputs after a host has already selected this pack. It is
// not the default natural-language routing path. Normal AgentX short-query turns
// should use latest_news_lookup so the model supplies structured intent and host
// adapters verify freshness, sources, and grounded facts.
//
// Keep this helper source-neutral: search providers, news-site preference, and
// interactive fallback remain in host/plugin adapters and runtime tools.
func BuildLatestNewsBriefCaseInput(userMessage string) (map[string]any, bool) {
	message := strings.TrimSpace(userMessage)
	if message == "" {
		return nil, false
	}
	topicName := latestNewsTopicName(message)
	if topicName == "" {
		return nil, false
	}
	fields := latestNewsRequestedFields(message)
	if len(fields) == 0 {
		return nil, false
	}
	return map[string]any{
		"user_message": message,
		"topic": map[string]any{
			"name":     topicName,
			"entities": []any{},
		},
		"requested_fields":   fields,
		"source_policy":      "public_web_prefer_official_or_authoritative_news_source",
		"freshness":          "live",
		"cross_check_policy": "at_least_two_independent_source_sites_for_key_facts",
		"stop_condition":     "guard_passed",
	}, true
}

func latestNewsTopicName(message string) string {
	for _, pattern := range latestNewsTopicPatterns {
		match := pattern.FindStringSubmatch(message)
		if len(match) < 2 {
			continue
		}
		if cleaned := cleanLatestNewsTopicName(match[1]); cleaned != "" {
			return cleaned
		}
	}
	cleaned := cleanLatestNewsTopicName(message)
	if !latestNewsLooksLikeFreshQuery(message) {
		return ""
	}
	return cleaned
}

func cleanLatestNewsTopicName(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	replacers := []string{
		"请", "",
		"帮我", "",
		"去", "",
		"查一下", "",
		"查下", "",
		"找一下", "",
		"找下", "",
		"看一下", "",
		"看下", "",
		"看看", "",
		"了解", "",
		"查询", "",
		"最新新闻", "",
		"最新进展", "",
		"最新消息", "",
		"新闻", "",
		"进展", "",
		"消息", "",
		"的", "",
	}
	for i := 0; i < len(replacers); i += 2 {
		value = strings.ReplaceAll(value, replacers[i], replacers[i+1])
	}
	value = strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。,.；;：:()（）[]【】<>《》\"'", r)
	})
	if value == "" || value == "最新" || value == "某个" || value == "某" {
		return ""
	}
	return value
}

func latestNewsRequestedFields(message string) []any {
	fields := []string{"headline", "published_at", "key_update", "source_url"}
	if containsAnyLatestNews(message, "来源", "出处", "source") {
		fields = append(fields, "source_site")
	}
	if containsAnyLatestNews(message, "影响", "后续", "下一步", "风险") {
		fields = append(fields, "implications")
	}
	out := make([]any, 0, len(fields))
	seen := map[string]bool{}
	for _, field := range fields {
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}

func latestNewsLooksLikeFreshQuery(message string) bool {
	return containsAnyLatestNews(message, "最新新闻", "最新进展", "最新消息", "最新", "news", "latest", "breaking")
}

func containsAnyLatestNews(message string, needles ...string) bool {
	lower := strings.ToLower(message)
	for _, needle := range needles {
		needle = strings.TrimSpace(needle)
		if needle == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
