// Package profile defines document extraction profile matching primitives for
// AgentX docparse.
//
// It intentionally ships no business profiles. Hosts or future scene packages
// may register invoice, contract, statement, or report profiles explicitly.
package profile

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	MatchStatusVerified  = "verified_profile"
	MatchStatusCandidate = "candidate_profile"
	MatchStatusUnknown   = "unknown_profile"
)

// ExtractionProfile describes an explicitly registered document extraction
// profile. Business semantics and private schemas are host-owned.
type ExtractionProfile struct {
	ID           string         `json:"id"`
	DocumentType string         `json:"document_type,omitempty"`
	Version      string         `json:"version,omitempty"`
	Description  string         `json:"description,omitempty"`
	SpecPath     string         `json:"spec_path,omitempty"`
	FieldKeys    []string       `json:"field_keys,omitempty"`
	TableKeys    []string       `json:"table_keys,omitempty"`
	RouteHints   []string       `json:"route_hints,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Registry stores explicitly registered extraction profiles.
type Registry struct {
	profiles []ExtractionProfile
}

// NewRegistry builds a profile registry.
func NewRegistry(profiles ...ExtractionProfile) *Registry {
	reg := &Registry{}
	for _, p := range profiles {
		reg.Add(p)
	}
	return reg
}

// Add registers a profile when it has an ID. Later profiles with the same ID
// replace earlier entries.
func (r *Registry) Add(profile ExtractionProfile) {
	if r == nil {
		return
	}
	profile.ID = strings.TrimSpace(profile.ID)
	if profile.ID == "" {
		return
	}
	for idx, existing := range r.profiles {
		if normalizeKey(existing.ID) == normalizeKey(profile.ID) {
			r.profiles[idx] = normalizeProfile(profile)
			return
		}
	}
	r.profiles = append(r.profiles, normalizeProfile(profile))
	sort.SliceStable(r.profiles, func(i, j int) bool {
		return strings.Compare(r.profiles[i].ID, r.profiles[j].ID) < 0
	})
}

// Profiles returns a copy of registered profiles.
func (r *Registry) Profiles() []ExtractionProfile {
	if r == nil {
		return nil
	}
	out := make([]ExtractionProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		out = append(out, cloneProfile(p))
	}
	return out
}

// MatchInput is the source-neutral profile matching request.
type MatchInput struct {
	ExplicitProfileID    string   `json:"explicit_profile_id,omitempty"`
	ExplicitDocumentType string   `json:"explicit_document_type,omitempty"`
	SpecPath             string   `json:"spec_path,omitempty"`
	RequestedFields      []string `json:"requested_fields,omitempty"`
}

// MatchResult is a conservative match result. Unknown documents should not be
// treated as answer-ready just because text was extractable.
type MatchResult struct {
	Status     string             `json:"status"`
	Profile    *ExtractionProfile `json:"profile,omitempty"`
	Candidates []Candidate        `json:"candidates,omitempty"`
	Proposal   *Proposal          `json:"proposal,omitempty"`
	Reasons    []string           `json:"reasons,omitempty"`
}

// Candidate records a possible profile match and the transparent signals used.
type Candidate struct {
	Profile ExtractionProfile `json:"profile"`
	Score   int               `json:"score"`
	Reasons []string          `json:"reasons,omitempty"`
}

// Proposal describes an unknown-profile follow-up for host/project review.
type Proposal struct {
	Reason                 string                  `json:"reason"`
	Source                 string                  `json:"source,omitempty"`
	RequestedFields        []string                `json:"requested_fields,omitempty"`
	SuggestedFields        []string                `json:"suggested_fields,omitempty"`
	SuggestedRouteHints    []string                `json:"suggested_route_hints,omitempty"`
	CandidateDocumentTypes []CandidateDocumentType `json:"candidate_document_types,omitempty"`
	SuggestedProfileID     string                  `json:"suggested_profile_id,omitempty"`
	EvidenceSnippets       []string                `json:"evidence_snippets,omitempty"`
	DiscoveryQueries       []string                `json:"discovery_queries,omitempty"`
	DiscoveryResults       []DiscoveryResult       `json:"discovery_results,omitempty"`
	SpecPath               string                  `json:"spec_path,omitempty"`
	ReviewRequired         bool                    `json:"review_required,omitempty"`
}

// CandidateDocumentType records a read-only model or host-provided type
// suggestion. It is not a verified profile.
type CandidateDocumentType struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}

// DiscoveryResult records host-approved public discovery context used to
// propose an unknown document profile. It is review context, not verification.
type DiscoveryResult struct {
	Query   string `json:"query,omitempty"`
	Title   string `json:"title,omitempty"`
	URL     string `json:"url,omitempty"`
	Snippet string `json:"snippet,omitempty"`
	Source  string `json:"source,omitempty"`
}

// Matcher matches only against explicitly registered profiles and explicit
// caller hints.
type Matcher struct {
	registry *Registry
}

// NewMatcher creates a matcher over a registry.
func NewMatcher(registry *Registry) *Matcher {
	if registry == nil {
		registry = NewRegistry()
	}
	return &Matcher{registry: registry}
}

// Match finds the best profile without using document-specific heuristics.
func (m *Matcher) Match(input MatchInput) MatchResult {
	profiles := m.registry.Profiles()
	if profile, ok := matchExplicitProfileID(profiles, input.ExplicitProfileID); ok {
		return MatchResult{
			Status:  MatchStatusVerified,
			Profile: &profile,
			Reasons: []string{"explicit_profile_id"},
		}
	}
	if profile, ok := matchSpecPath(profiles, input.SpecPath); ok {
		return MatchResult{
			Status:  MatchStatusVerified,
			Profile: &profile,
			Reasons: []string{"explicit_spec_path"},
		}
	}
	candidates := matchDocumentTypeCandidates(profiles, input.ExplicitDocumentType)
	if len(candidates) > 0 {
		return MatchResult{
			Status:     MatchStatusCandidate,
			Candidates: candidates,
			Reasons:    []string{"explicit_document_type"},
		}
	}
	return MatchResult{
		Status: MatchStatusUnknown,
		Proposal: &Proposal{
			Reason:          "no explicit registered profile matched",
			Source:          "profile_matcher",
			RequestedFields: uniqueStrings(input.RequestedFields),
			SpecPath:        strings.TrimSpace(input.SpecPath),
			ReviewRequired:  true,
		},
		Reasons: []string{"profile_not_registered"},
	}
}

func matchExplicitProfileID(profiles []ExtractionProfile, profileID string) (ExtractionProfile, bool) {
	profileID = normalizeKey(profileID)
	if profileID == "" {
		return ExtractionProfile{}, false
	}
	for _, profile := range profiles {
		if normalizeKey(profile.ID) == profileID {
			return cloneProfile(profile), true
		}
	}
	return ExtractionProfile{}, false
}

func matchSpecPath(profiles []ExtractionProfile, specPath string) (ExtractionProfile, bool) {
	specPath = normalizePath(specPath)
	if specPath == "" {
		return ExtractionProfile{}, false
	}
	for _, profile := range profiles {
		if normalizePath(profile.SpecPath) == specPath {
			return cloneProfile(profile), true
		}
	}
	return ExtractionProfile{}, false
}

func matchDocumentTypeCandidates(profiles []ExtractionProfile, documentType string) []Candidate {
	documentType = normalizeKey(documentType)
	if documentType == "" {
		return nil
	}
	out := []Candidate{}
	for _, profile := range profiles {
		if normalizeKey(profile.DocumentType) != documentType {
			continue
		}
		out = append(out, Candidate{
			Profile: cloneProfile(profile),
			Score:   50,
			Reasons: []string{"document_type"},
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Profile.ID < out[j].Profile.ID
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func normalizeProfile(profile ExtractionProfile) ExtractionProfile {
	profile.ID = strings.TrimSpace(profile.ID)
	profile.DocumentType = strings.TrimSpace(profile.DocumentType)
	profile.Version = strings.TrimSpace(profile.Version)
	profile.Description = strings.TrimSpace(profile.Description)
	profile.SpecPath = strings.TrimSpace(profile.SpecPath)
	profile.FieldKeys = uniqueStrings(profile.FieldKeys)
	profile.TableKeys = uniqueStrings(profile.TableKeys)
	profile.RouteHints = uniqueStrings(profile.RouteHints)
	profile.Tags = uniqueStrings(profile.Tags)
	if profile.Metadata != nil {
		metadata := make(map[string]any, len(profile.Metadata))
		for key, value := range profile.Metadata {
			key = strings.TrimSpace(key)
			if key != "" {
				metadata[key] = value
			}
		}
		profile.Metadata = metadata
	}
	return profile
}

func cloneProfile(profile ExtractionProfile) ExtractionProfile {
	profile.FieldKeys = append([]string(nil), profile.FieldKeys...)
	profile.TableKeys = append([]string(nil), profile.TableKeys...)
	profile.RouteHints = append([]string(nil), profile.RouteHints...)
	profile.Tags = append([]string(nil), profile.Tags...)
	if profile.Metadata != nil {
		metadata := make(map[string]any, len(profile.Metadata))
		for key, value := range profile.Metadata {
			metadata[key] = value
		}
		profile.Metadata = metadata
	}
	return profile
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	underscore := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			underscore = false
		default:
			if !underscore && b.Len() > 0 {
				b.WriteByte('_')
				underscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func normalizePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.Clean(value)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
