package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Catalog is an immutable, concurrency-safe capability discovery index.
type Catalog struct {
	policy      Policy
	fingerprint string
	assets      []Asset
}

// Build validates, normalizes and deterministically indexes assets.
func Build(policy Policy, assets []Asset) (*Catalog, error) {
	if !validPolicy(policy) {
		return nil, &Error{Code: ErrorCodeInvalidPolicy}
	}
	if len(assets) > policy.MaxAssets {
		return nil, &Error{Code: ErrorCodeInvalidAsset, Cause: fmt.Errorf("asset count exceeds limit")}
	}
	normalized := make([]Asset, 0, len(assets))
	seen := make(map[Identity]struct{}, len(assets))
	for _, asset := range assets {
		item, err := normalizeAsset(policy, asset)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[item.Identity]; exists {
			return nil, &Error{Code: ErrorCodeDuplicateAsset}
		}
		seen[item.Identity] = struct{}{}
		normalized = append(normalized, item)
	}
	sortAssets(normalized)
	fingerprint, err := fingerprintAssets(normalized)
	if err != nil {
		return nil, &Error{Code: ErrorCodeInvalidAsset, Cause: err}
	}
	return &Catalog{policy: policy, fingerprint: fingerprint, assets: normalized}, nil
}

// Len returns the number of indexed assets.
func (c *Catalog) Len() int {
	if c == nil {
		return 0
	}
	return len(c.assets)
}

// Snapshot returns a detached stable-order catalog view.
func (c *Catalog) Snapshot() Snapshot {
	if c == nil {
		return Snapshot{}
	}
	return Snapshot{Fingerprint: c.fingerprint, Assets: cloneAssets(c.assets)}
}

// Search performs bounded lexical discovery over the immutable catalog.
func (c *Catalog) Search(query Query) (SearchResult, error) {
	if c == nil || !validPolicy(c.policy) {
		return SearchResult{}, &Error{Code: ErrorCodeInvalidPolicy}
	}
	normalized, err := normalizeQuery(c.policy, query)
	if err != nil {
		return SearchResult{}, err
	}
	hits := make([]SearchHit, 0, min(len(c.assets), normalized.Limit))
	matched := 0
	for _, asset := range c.assets {
		if !matchesKinds(asset, normalized.Kinds) || !matchesAnyTag(asset, normalized.AnyTags) {
			continue
		}
		score, ok := lexicalScore(asset, normalized.Text)
		if !ok {
			continue
		}
		matched++
		hits = append(hits, SearchHit{Asset: cloneAsset(asset), Score: score})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return identityLess(hits[i].Asset.Identity, hits[j].Asset.Identity)
	})
	limited := len(hits) > normalized.Limit
	if limited {
		hits = hits[:normalized.Limit]
	}
	return SearchResult{
		Fingerprint: c.fingerprint,
		Hits:        hits,
		Matched:     matched,
		Limited:     limited,
	}, nil
}

// Diff compares detached snapshots without mutating either input.
func Diff(before, after Snapshot) ChangeSet {
	oldItems := make(map[Identity]Asset, len(before.Assets))
	newItems := make(map[Identity]Asset, len(after.Assets))
	for _, asset := range before.Assets {
		oldItems[asset.Identity] = asset
	}
	for _, asset := range after.Assets {
		newItems[asset.Identity] = asset
	}
	var out ChangeSet
	for identity, oldAsset := range oldItems {
		newAsset, exists := newItems[identity]
		if !exists {
			out.Removed = append(out.Removed, identity)
			continue
		}
		if !assetEqual(oldAsset, newAsset) {
			out.Changed = append(out.Changed, identity)
		}
	}
	for identity := range newItems {
		if _, exists := oldItems[identity]; !exists {
			out.Added = append(out.Added, identity)
		}
	}
	sortIdentities(out.Added)
	sortIdentities(out.Removed)
	sortIdentities(out.Changed)
	return out
}

func validPolicy(policy Policy) bool {
	return policy.MaxAssets > 0 && policy.MaxSearchLimit > 0 &&
		policy.MaxIdentityBytes > 0 && policy.MaxTextBytes > 0 &&
		policy.MaxTags >= 0 && policy.MaxKeywords >= 0
}

func normalizeAsset(policy Policy, asset Asset) (Asset, error) {
	asset.Identity.Kind = Kind(strings.ToLower(strings.TrimSpace(string(asset.Identity.Kind))))
	asset.Identity.ID = strings.ToLower(strings.TrimSpace(asset.Identity.ID))
	asset.Name = strings.TrimSpace(asset.Name)
	asset.Description = strings.TrimSpace(asset.Description)
	asset.Version = strings.TrimSpace(asset.Version)
	asset.SourceRef = strings.TrimSpace(asset.SourceRef)
	if !validKind(asset.Identity.Kind) || !validIdentityText(asset.Identity.ID, policy.MaxIdentityBytes) ||
		!validIdentityText(asset.SourceRef, policy.MaxIdentityBytes) {
		return Asset{}, &Error{Code: ErrorCodeInvalidAsset}
	}
	if asset.Name == "" {
		asset.Name = asset.Identity.ID
	}
	if !validText(asset.Name, policy.MaxTextBytes) || !validText(asset.Description, policy.MaxTextBytes) ||
		!validText(asset.Version, policy.MaxTextBytes) {
		return Asset{}, &Error{Code: ErrorCodeInvalidAsset}
	}
	var err error
	asset.Tags, err = normalizeTerms(asset.Tags, policy.MaxTags, policy.MaxTextBytes)
	if err != nil {
		return Asset{}, err
	}
	asset.Keywords, err = normalizeTerms(asset.Keywords, policy.MaxKeywords, policy.MaxTextBytes)
	if err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func normalizeQuery(policy Policy, query Query) (Query, error) {
	query.Text = strings.ToLower(strings.TrimSpace(query.Text))
	if query.Limit < 0 || len(query.Text) > policy.MaxTextBytes {
		return Query{}, &Error{Code: ErrorCodeInvalidQuery}
	}
	if query.Limit == 0 || query.Limit > policy.MaxSearchLimit {
		query.Limit = policy.MaxSearchLimit
	}
	seenKinds := make(map[Kind]struct{}, len(query.Kinds))
	query.Kinds = slices.Clone(query.Kinds)
	for index, kind := range query.Kinds {
		kind = Kind(strings.ToLower(strings.TrimSpace(string(kind))))
		if !validKind(kind) {
			return Query{}, &Error{Code: ErrorCodeInvalidQuery}
		}
		if _, exists := seenKinds[kind]; exists {
			return Query{}, &Error{Code: ErrorCodeInvalidQuery}
		}
		seenKinds[kind] = struct{}{}
		query.Kinds[index] = kind
	}
	var err error
	query.AnyTags, err = normalizeTerms(query.AnyTags, policy.MaxTags, policy.MaxTextBytes)
	if err != nil {
		return Query{}, &Error{Code: ErrorCodeInvalidQuery, Cause: err}
	}
	return query, nil
}

func normalizeTerms(values []string, limit, textLimit int) ([]string, error) {
	if len(values) > limit {
		return nil, &Error{Code: ErrorCodeInvalidAsset}
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if !validText(value, textLimit) {
			return nil, &Error{Code: ErrorCodeInvalidAsset}
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

func validKind(kind Kind) bool {
	switch kind {
	case KindTool, KindSkill, KindPlugin, KindConnector, KindExpert, KindTeam:
		return true
	default:
		return false
	}
}

func validIdentityText(value string, limit int) bool {
	if value == "" || len(value) > limit || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func validText(value string, limit int) bool {
	if len(value) > limit || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func matchesKinds(asset Asset, kinds []Kind) bool {
	return len(kinds) == 0 || slices.Contains(kinds, asset.Identity.Kind)
}

func matchesAnyTag(asset Asset, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	for _, tag := range tags {
		if slices.Contains(asset.Tags, tag) {
			return true
		}
	}
	return false
}

func lexicalScore(asset Asset, text string) (int, bool) {
	tokens := strings.Fields(text)
	if len(tokens) == 0 {
		return 0, true
	}
	id := strings.ToLower(asset.Identity.ID)
	name := strings.ToLower(asset.Name)
	description := strings.ToLower(asset.Description)
	score := 0
	for _, token := range tokens {
		best := 0
		switch {
		case token == id || token == name:
			best = 100
		case strings.HasPrefix(id, token) || strings.HasPrefix(name, token):
			best = 80
		case slices.Contains(asset.Keywords, token):
			best = 70
		case slices.Contains(asset.Tags, token):
			best = 60
		case strings.Contains(id, token) || strings.Contains(name, token):
			best = 40
		case containsTerm(asset.Keywords, token) || containsTerm(asset.Tags, token):
			best = 30
		case strings.Contains(description, token):
			best = 10
		}
		if best == 0 {
			return 0, false
		}
		score += best
	}
	return score, true
}

func containsTerm(values []string, token string) bool {
	for _, value := range values {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func fingerprintAssets(assets []Asset) (string, error) {
	content, err := json.Marshal(assets)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func cloneAssets(assets []Asset) []Asset {
	out := make([]Asset, len(assets))
	for index, asset := range assets {
		out[index] = cloneAsset(asset)
	}
	return out
}

func cloneAsset(asset Asset) Asset {
	asset.Tags = slices.Clone(asset.Tags)
	asset.Keywords = slices.Clone(asset.Keywords)
	return asset
}

func sortAssets(assets []Asset) {
	sort.Slice(assets, func(i, j int) bool { return identityLess(assets[i].Identity, assets[j].Identity) })
}

func sortIdentities(identities []Identity) {
	sort.Slice(identities, func(i, j int) bool { return identityLess(identities[i], identities[j]) })
}

func identityLess(left, right Identity) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return left.ID < right.ID
}

func assetEqual(left, right Asset) bool {
	return left.Identity == right.Identity && left.Name == right.Name &&
		left.Description == right.Description && left.Version == right.Version &&
		left.SourceRef == right.SourceRef && slices.Equal(left.Tags, right.Tags) &&
		slices.Equal(left.Keywords, right.Keywords)
}
