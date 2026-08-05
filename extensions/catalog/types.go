package catalog

// Kind identifies one discoverable capability family. Kind is discovery
// metadata only; it does not select an executor or lifecycle owner.
type Kind string

const (
	KindTool      Kind = "tool"
	KindSkill     Kind = "skill"
	KindPlugin    Kind = "plugin"
	KindConnector Kind = "connector"
	KindExpert    Kind = "expert"
	KindTeam      Kind = "team"
)

// Identity is the stable catalog key for one asset.
type Identity struct {
	Kind Kind   `json:"kind"`
	ID   string `json:"id"`
}

// Asset is the common, display-safe discovery envelope. SourceRef is an opaque
// Host-provided provenance reference; it must never contain credentials.
type Asset struct {
	Identity    Identity `json:"identity"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	SourceRef   string   `json:"source_ref"`
	Tags        []string `json:"tags,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// Policy provides explicit safety bounds for catalog construction and search.
type Policy struct {
	MaxAssets        int
	MaxSearchLimit   int
	MaxIdentityBytes int
	MaxTextBytes     int
	MaxTags          int
	MaxKeywords      int
}

// DefaultPolicy returns conservative in-process discovery bounds. Callers may
// copy and narrow them; Catalog never reads product configuration implicitly.
func DefaultPolicy() Policy {
	return Policy{
		MaxAssets:        4096,
		MaxSearchLimit:   100,
		MaxIdentityBytes: 256,
		MaxTextBytes:     16 << 10,
		MaxTags:          64,
		MaxKeywords:      128,
	}
}

// Query filters one immutable catalog snapshot. AnyTags uses OR semantics;
// text tokens use AND semantics.
type Query struct {
	Text    string
	Kinds   []Kind
	AnyTags []string
	Limit   int
}

// SearchHit contains a detached asset and an explainable lexical score. Score
// is catalog ordering metadata, not an execution-routing decision.
type SearchHit struct {
	Asset Asset `json:"asset"`
	Score int   `json:"score"`
}

// SearchResult reports the immutable snapshot searched and whether results
// were bounded by Query.Limit.
type SearchResult struct {
	Fingerprint string      `json:"fingerprint"`
	Hits        []SearchHit `json:"hits,omitempty"`
	Matched     int         `json:"matched"`
	Limited     bool        `json:"limited,omitempty"`
}

// Snapshot is a detached, stable-order view of a Catalog.
type Snapshot struct {
	Fingerprint string  `json:"fingerprint"`
	Assets      []Asset `json:"assets,omitempty"`
}

// ChangeSet describes identity-level changes between two snapshots.
type ChangeSet struct {
	Added   []Identity `json:"added,omitempty"`
	Removed []Identity `json:"removed,omitempty"`
	Changed []Identity `json:"changed,omitempty"`
}
