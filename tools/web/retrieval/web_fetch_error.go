package retrieval

import (
	"fmt"
	"strings"
)

type WebFetchErrorKind string

const (
	WebFetchErrorRequest       WebFetchErrorKind = "request_failed"
	WebFetchErrorTimeout       WebFetchErrorKind = "timeout"
	WebFetchErrorRedirectLimit WebFetchErrorKind = "redirect_limit"
	WebFetchErrorProxyConfig   WebFetchErrorKind = "proxy_config_invalid"
	WebFetchErrorPolicyBlocked WebFetchErrorKind = "policy_blocked"
	WebFetchErrorTLS           WebFetchErrorKind = "tls_error"
	WebFetchErrorStatus        WebFetchErrorKind = "status_error"
)

type WebFetchError struct {
	Kind        WebFetchErrorKind
	ToolName    string
	URL         string
	TimeoutMs   int
	Status      int
	StatusClass string
	ContentType string
	Challenge   bool
	fallback    string
	cause       error
}

func (e *WebFetchError) Error() string {
	if e == nil {
		return "web_fetch: request_failed"
	}
	toolName := strings.TrimSpace(e.ToolName)
	if toolName == "" {
		toolName = "web_fetch"
	}
	urlValue := strings.TrimSpace(e.URL)
	switch e.Kind {
	case WebFetchErrorTimeout:
		return fmt.Sprintf("%s: timeout url=%s timeout_ms=%d hint=try reducing response size or increasing timeout_ms%s", toolName, urlValue, e.TimeoutMs, e.fallback)
	case WebFetchErrorRedirectLimit:
		return fmt.Sprintf("%s: redirect_limit url=%s hint=check redirect chain or max_redirects setting%s", toolName, urlValue, e.fallback)
	case WebFetchErrorProxyConfig:
		return fmt.Sprintf("%s: proxy_config_invalid url=%s hint=check tools.webFetchTrustedEnvProxy and HTTP_PROXY/HTTPS_PROXY/NO_PROXY%s", toolName, urlValue, e.fallback)
	case WebFetchErrorPolicyBlocked:
		return fmt.Sprintf("%s: policy_blocked url=%s reason=outbound_policy%s", toolName, urlValue, e.fallback)
	case WebFetchErrorTLS:
		return fmt.Sprintf("%s: tls_error url=%s hint=verify certificate and endpoint trust chain%s", toolName, urlValue, e.fallback)
	case WebFetchErrorStatus:
		return fmt.Sprintf(
			"%s: status_error class=%s status=%d url=%s content_type=%s hint=check endpoint path/authentication and upstream availability%s",
			toolName,
			strings.TrimSpace(e.StatusClass),
			e.Status,
			urlValue,
			strings.TrimSpace(e.ContentType),
			e.fallback,
		)
	default:
		return fmt.Sprintf("%s: request_failed url=%s hint=check network connectivity and endpoint availability%s", toolName, urlValue, e.fallback)
	}
}

func (e *WebFetchError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}
