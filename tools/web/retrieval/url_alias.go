package retrieval

import (
	"net/url"
	"strings"
)

var retrievalURLAliasFields = []string{
	"source_url",
	"sourceUrl",
	"href",
	"link",
	"page_url",
	"pageUrl",
}

// FirstHTTPURLAlias returns the first safe retrieval URL alias when no canonical url is present.
func FirstHTTPURLAlias(params map[string]any) (field string, rawURL string, ok bool) {
	if params == nil {
		return "", "", false
	}
	if _, hasExplicitURL := params["url"]; hasExplicitURL {
		return "", "", false
	}
	for _, candidateField := range retrievalURLAliasFields {
		value := strings.TrimSpace(stringParam(params, candidateField))
		if !IsHTTPURLAliasCandidate(value) {
			continue
		}
		return candidateField, value, true
	}
	return "", "", false
}

// IsHTTPURLAliasCandidate reports whether value is an absolute http(s) URL.
func IsHTTPURLAliasCandidate(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return false
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	return (scheme == "http" || scheme == "https") && strings.TrimSpace(parsed.Host) != ""
}

func stringParam(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	value, ok := params[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
