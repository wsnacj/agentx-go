package publicsource

import (
	"net/url"
	"strings"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

// SourcePolicy applies Host-supplied HTTPS and allowlist rules. The mechanism
// is canonical; the actual hosts and URL prefixes remain Host policy.
type SourcePolicy struct {
	PolicyRef           control.DisplaySafeRef `json:"policy_ref,omitempty"`
	AllowedHosts        []string               `json:"allowed_hosts,omitempty"`
	AllowedHostSuffixes []string               `json:"allowed_host_suffixes,omitempty"`
	AllowedURLPrefixes  []string               `json:"allowed_url_prefixes,omitempty"`
	RequireHTTPS        bool                   `json:"require_https"`
	Boundaries          []control.Boundary     `json:"boundaries,omitempty"`
}

type SourcePolicyDecision struct {
	Applied       bool                   `json:"applied"`
	Allowed       bool                   `json:"allowed"`
	FailureClass  control.FailureClass   `json:"failure_class,omitempty"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	PolicyRef     control.DisplaySafeRef `json:"policy_ref,omitempty"`
	AllowedCount  int                    `json:"allowed_count,omitempty"`
	RejectedCount int                    `json:"rejected_count,omitempty"`
	Boundaries    []control.Boundary     `json:"boundaries,omitempty"`
}

func (policy SourcePolicy) Enabled() bool { return policy.normalize().enabled() }

func (policy SourcePolicy) CheckURL(rawURL string) SourcePolicyDecision {
	policy = policy.normalize()
	decision := SourcePolicyDecision{
		Applied: policy.enabled(), Allowed: true, FailureClass: control.FailureNone, PolicyRef: policy.PolicyRef,
		Boundaries: control.AppendBoundaries(policy.Boundaries, "host_owned_public_source_source_policy", "public_source_source_policy_applied", "source_allowlist_supplied_by_host"),
	}
	if !decision.Applied {
		decision.Boundaries = nil
		return decision
	}
	trimmed := strings.TrimSpace(rawURL)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed == nil || strings.TrimSpace(parsed.Hostname()) == "" {
		return policy.block(decision, "public_source_source_policy_url_invalid")
	}
	if policy.RequireHTTPS && strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "https" {
		return policy.block(decision, "public_source_source_policy_https_required")
	}
	host := normalizeHost(parsed.Hostname())
	if policy.hasRules() && !policy.matches(trimmed, host) {
		return policy.block(decision, "public_source_source_not_allowlisted")
	}
	decision.Boundaries = control.AppendBoundaries(decision.Boundaries, "public_source_source_policy_allowed")
	return decision
}

// Filter removes disallowed search results and position-matched summaries.
func (policy SourcePolicy) Filter(payload SearchPayload, summaries []DisplaySummary) (SearchPayload, []DisplaySummary, SourcePolicyDecision) {
	policy = policy.normalize()
	decision := SourcePolicyDecision{Applied: policy.enabled(), Allowed: true, FailureClass: control.FailureNone, PolicyRef: policy.PolicyRef}
	if !decision.Applied {
		return payload, append([]DisplaySummary(nil), summaries...), decision
	}
	decision.Boundaries = control.AppendBoundaries(policy.Boundaries, "host_owned_public_source_source_policy", "public_source_source_policy_applied", "source_allowlist_supplied_by_host")
	results := make([]SearchResult, 0, len(payload.Results))
	filteredSummaries := make([]DisplaySummary, 0, len(summaries))
	for index, result := range payload.Results {
		if item := policy.CheckURL(result.URL); item.Allowed {
			results = append(results, result)
			if index < len(summaries) {
				filteredSummaries = append(filteredSummaries, summaries[index])
			}
			decision.AllowedCount++
		} else {
			decision.RejectedCount++
			decision.Boundaries = control.AppendBoundaries(decision.Boundaries, "public_source_source_policy_filtered_result")
		}
	}
	if decision.AllowedCount == 0 && decision.RejectedCount > 0 {
		decision.Allowed, decision.FailureClass, decision.FailureReason = false, control.FailurePolicyBlocked, "public_source_source_policy_no_allowed_sources"
		decision.Boundaries = control.AppendBoundaries(decision.Boundaries, "public_source_source_policy_blocked", control.Boundary(decision.FailureReason))
	} else {
		decision.Boundaries = control.AppendBoundaries(decision.Boundaries, "public_source_source_policy_allowed")
	}
	payload.Results, payload.Count = results, len(results)
	return payload, filteredSummaries, decision
}

func (policy SourcePolicy) block(decision SourcePolicyDecision, reason string) SourcePolicyDecision {
	decision.Allowed, decision.FailureClass, decision.FailureReason = false, control.FailurePolicyBlocked, controlToken(reason)
	decision.Boundaries = control.AppendBoundaries(decision.Boundaries, "public_source_source_policy_blocked", control.Boundary(decision.FailureReason))
	return decision
}

func (policy SourcePolicy) normalize() SourcePolicy {
	out := policy
	out.PolicyRef = normalizeRef(out.PolicyRef)
	if out.PolicyRef == "" && (len(out.AllowedHosts) > 0 || len(out.AllowedHostSuffixes) > 0 || len(out.AllowedURLPrefixes) > 0 || out.RequireHTTPS) {
		out.PolicyRef = "policy:host_public_source_allowlist"
	}
	out.AllowedHosts = normalizePolicyValues(out.AllowedHosts, normalizeHost)
	out.AllowedHostSuffixes = normalizePolicyValues(out.AllowedHostSuffixes, func(value string) string { return strings.TrimLeft(normalizeHost(value), ".") })
	out.AllowedURLPrefixes = normalizePolicyValues(out.AllowedURLPrefixes, func(value string) string { return strings.ToLower(strings.TrimSpace(value)) })
	out.Boundaries = control.AppendBoundaries(nil, out.Boundaries...)
	return out
}

func (policy SourcePolicy) enabled() bool {
	return policy.RequireHTTPS || len(policy.AllowedHosts)+len(policy.AllowedHostSuffixes)+len(policy.AllowedURLPrefixes) > 0
}
func (policy SourcePolicy) hasRules() bool {
	return len(policy.AllowedHosts)+len(policy.AllowedHostSuffixes)+len(policy.AllowedURLPrefixes) > 0
}
func (policy SourcePolicy) matches(rawURL, host string) bool {
	for _, candidate := range policy.AllowedHosts {
		if host == candidate {
			return true
		}
	}
	for _, suffix := range policy.AllowedHostSuffixes {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	normalizedURL := strings.ToLower(strings.TrimSpace(rawURL))
	for _, prefix := range policy.AllowedURLPrefixes {
		if strings.HasPrefix(normalizedURL, prefix) {
			return true
		}
	}
	return false
}

func normalizePolicyValues(values []string, normalize func(string) string) []string {
	out, seen := []string{}, map[string]bool{}
	for _, value := range values {
		value = normalize(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func normalizeHost(value string) string {
	host := strings.Trim(strings.ToLower(strings.TrimSpace(value)), ".")
	if parsed, err := url.Parse(host); err == nil && parsed != nil && parsed.Hostname() != "" {
		host = strings.Trim(strings.ToLower(strings.TrimSpace(parsed.Hostname())), ".")
	}
	return host
}
