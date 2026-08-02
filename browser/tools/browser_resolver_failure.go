package tools

import (
	"errors"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// BrowserResolverFailure exposes portable resolver failure details returned by
// a Host backend. Concrete transport errors remain owned by the Host.
type BrowserResolverFailure interface {
	error
	BrowserResolverOutcome() *agentxbrowserruntime.BrowserElementResolverOutcome
	BrowserResolverMessage() string
}

func browserManagedResolverFailure(err error) (string, *agentxbrowserruntime.BrowserElementResolverOutcome, string, bool) {
	var resolverErr BrowserResolverFailure
	if !errors.As(err, &resolverErr) || resolverErr == nil {
		return "", nil, "", false
	}
	rawOutcome := resolverErr.BrowserResolverOutcome()
	if rawOutcome == nil {
		return "", nil, "", false
	}
	normalized := rawOutcome.Normalized()
	if normalized == nil {
		return "", nil, "", false
	}
	outcome := *normalized
	status := firstNonEmpty(strings.TrimSpace(outcome.Status), "resolution_failed")
	note := browserManagedResolverFailureNote(status, firstNonEmpty(strings.TrimSpace(outcome.Note), strings.TrimSpace(resolverErr.BrowserResolverMessage())))
	return status, &outcome, note, true
}

func browserResolverRecoveryAction(outcome *agentxbrowserruntime.BrowserElementResolverOutcome) string {
	if outcome == nil {
		return ""
	}
	normalized := outcome.Normalized()
	if normalized == nil {
		return ""
	}
	return strings.TrimSpace(normalized.RecoveryAction)
}

func browserManagedResolverFailureNote(status string, upstream string) string {
	upstream = strings.TrimSpace(upstream)
	switch strings.TrimSpace(status) {
	case "page_binding_blocked":
		return firstNonEmpty(upstream, "resolver blocked by page binding; refresh the page or capture a fresh snapshot before retrying")
	case "unresolved":
		return firstNonEmpty(upstream, "element no longer matches resolver plan; capture a fresh snapshot and retry")
	case "resolution_failed":
		return firstNonEmpty(upstream, "resolver failed before the action could run; refresh the page or retry with a fresh target")
	default:
		return firstNonEmpty(upstream, "resolver failed before the action could run")
	}
}

func browserResolverOutcomeAllowsTargetTracking(outcome *agentxbrowserruntime.BrowserElementResolverOutcome) bool {
	if outcome == nil {
		return true
	}
	return strings.TrimSpace(outcome.Status) == "" || strings.TrimSpace(outcome.Status) == "matched"
}
