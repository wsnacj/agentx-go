package retrieval

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/wsnacj/agentx-go/tools/httprequest"
)

// Preparer is the narrow Host-owned network construction port used by search
// and page retrieval. The Host owns URL policy, proxy selection, redirects and
// transport lifecycle; this package owns only provider and extraction logic.
type Preparer = httprequest.Preparer

// ExternalContentMeta identifies untrusted fields returned by retrieval.
type ExternalContentMeta struct {
	Untrusted bool     `json:"untrusted"`
	Source    string   `json:"source,omitempty"`
	Wrapped   bool     `json:"wrapped,omitempty"`
	Fields    []string `json:"fields,omitempty"`
}

// NetworkErrorClass lets a Host retain its own network error identities while
// mapping them into the stable retrieval error vocabulary.
type NetworkErrorClass string

const (
	NetworkErrorUnknown       NetworkErrorClass = ""
	NetworkErrorRedirectLimit NetworkErrorClass = "redirect_limit"
	NetworkErrorProxyConfig   NetworkErrorClass = "proxy_config"
	NetworkErrorPolicyBlocked NetworkErrorClass = "policy_blocked"
)

// NetworkErrorClassifier maps a Host-specific error without transferring
// policy ownership into this package.
type NetworkErrorClassifier func(error) NetworkErrorClass

// URLPrepareError reports a Host rejection without exposing URL credentials or
// losing the Host error identity for errors.Is/errors.As.
type URLPrepareError struct {
	Tool  string
	URL   string
	Cause error
}

func (e *URLPrepareError) Error() string {
	if e == nil {
		return "web retrieval: invalid_or_blocked_url"
	}
	return fmt.Sprintf("%s: invalid_or_blocked_url url=%s reason=%s", strings.TrimSpace(e.Tool), evidenceURL(e.URL), strings.TrimSpace(e.Cause.Error()))
}

func (e *URLPrepareError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func evidenceURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil {
		return "[invalid-url]"
	}
	if parsed.IsAbs() {
		scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
		if (scheme != "http" && scheme != "https") || strings.TrimSpace(parsed.Hostname()) == "" {
			return "[invalid-url]"
		}
	}
	parsed.User = nil
	parsed.Fragment = ""
	parsed.ForceQuery = false
	parsed.RawQuery = ""
	return parsed.String()
}

func requestIdentitySHA256(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func runeLen(text string) int { return utf8.RuneCountInString(text) }

func trimToMaxChars(text string, max int) (string, bool) {
	if max <= 0 || text == "" || runeLen(text) <= max {
		return text, false
	}
	return string([]rune(text)[:max]), true
}

func truncateWithEllipsis(text string, max int) string {
	if max <= 0 || text == "" || runeLen(text) <= max {
		return text
	}
	if max == 1 {
		return "…"
	}
	return string([]rune(text)[:max-1]) + "…"
}
