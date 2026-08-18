package retrieval

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// SearchProviderError is a display-safe provider failure. It preserves a
// stable health class and retryability without retaining provider response
// bodies, credentials or user queries.
type SearchProviderError struct {
	Provider   string
	Code       string
	Health     string
	HTTPStatus int
	Retryable  bool
}

func (e *SearchProviderError) Error() string {
	if e == nil {
		return "web_search: provider_error"
	}
	parts := []string{"web_search: provider_error", "provider=" + NormalizeSearchProvider(e.Provider)}
	if code := strings.TrimSpace(e.Code); code != "" {
		parts = append(parts, "code="+code)
	}
	if health := NormalizeProviderHealthStatus(e.Health); health != "" {
		parts = append(parts, "health="+health)
	}
	if e.HTTPStatus > 0 {
		parts = append(parts, "http_status="+strconv.Itoa(e.HTTPStatus))
	}
	parts = append(parts, fmt.Sprintf("retryable=%t", e.Retryable))
	return strings.Join(parts, " ")
}

func newDoubaoProviderError(provider, code, message string, httpStatus int) error {
	code = strings.TrimSpace(code)
	normalizedMessage := strings.ToLower(strings.TrimSpace(message))
	health, retryable := ProviderHealthStatusError, false
	switch code {
	case "10403", "700901":
		health = ProviderHealthCredentialInvalid
	case "10406", "10410", "10412":
		health = ProviderHealthQuotaLimited
	case "10408", "10409":
		health = ProviderHealthUnsupportedFeature
	case "700429":
		health, retryable = ProviderHealthRateLimited, true
	case "10500", "10501":
		health, retryable = ProviderHealthRequestFailed, true
	default:
		switch {
		case strings.Contains(normalizedMessage, "invalid_api_key"), strings.Contains(normalizedMessage, "invalid api key"):
			health = ProviderHealthCredentialInvalid
		case httpStatus == http.StatusUnauthorized || httpStatus == http.StatusForbidden:
			health = ProviderHealthCredentialInvalid
		case httpStatus == http.StatusTooManyRequests:
			health, retryable = ProviderHealthRateLimited, true
		case httpStatus >= 500:
			health, retryable = ProviderHealthRequestFailed, true
		}
	}
	return &SearchProviderError{Provider: provider, Code: code, Health: health, HTTPStatus: httpStatus, Retryable: retryable}
}
