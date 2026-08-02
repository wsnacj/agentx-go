// Package credentialhttp provides the module-internal HTTP redirect policy for
// clients that attach credentials or signed request bodies.
package credentialhttp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultMaxRedirects = 5

var (
	// ErrCrossOriginRedirect reports a redirect that would move a credentialed
	// request to a different origin.
	ErrCrossOriginRedirect = errors.New("credential HTTP client rejected a cross-origin redirect")
	// ErrRedirectLimit reports a redirect chain beyond the bounded policy.
	ErrRedirectLimit = errors.New("credential HTTP client exceeded its redirect limit")
)

// NewClient returns an HTTP client that follows only bounded same-origin
// redirects. It prevents custom authentication headers and signed request
// bodies from being replayed to a different scheme, host, or effective port.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: CheckRedirect,
	}
}

// CheckRedirect is the shared redirect callback for credential-bearing HTTP
// clients in this module.
func CheckRedirect(request *http.Request, via []*http.Request) error {
	if request == nil || request.URL == nil || len(via) == 0 || via[0] == nil || via[0].URL == nil {
		return fmt.Errorf("%w: redirect URL context is incomplete", ErrCrossOriginRedirect)
	}
	if len(via) > defaultMaxRedirects {
		return ErrRedirectLimit
	}
	if !sameOrigin(via[0].URL, request.URL) {
		return ErrCrossOriginRedirect
	}
	return nil
}

func sameOrigin(left, right *url.URL) bool {
	if left == nil || right == nil {
		return false
	}
	return strings.EqualFold(left.Scheme, right.Scheme) &&
		strings.EqualFold(left.Hostname(), right.Hostname()) &&
		effectivePort(left) == effectivePort(right)
}

func effectivePort(value *url.URL) string {
	if value == nil {
		return ""
	}
	if port := value.Port(); port != "" {
		return port
	}
	switch strings.ToLower(value.Scheme) {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}
