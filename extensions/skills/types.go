package skills

import (
	"io/fs"
	"strings"
)

type Source string

const (
	SourceCustom    Source = "custom"
	SourceExtra     Source = "extra"
	SourceBundled   Source = "bundled"
	SourceManaged   Source = "managed"
	SourceWorkspace Source = "workspace"
)

type InvocationPolicy struct {
	UserInvocable          bool
	DisableModelInvocation bool
}

const (
	SkillExecutionContextInline = "inline"
	SkillExecutionContextFork   = "fork"
)

type Requires struct {
	Bins    []string
	AnyBins []string
	Env     []string
	Config  []string
}

type InstallSpec struct {
	ID              string
	DependsOn       []string
	Kind            string
	Label           string
	Bins            []string
	OS              []string
	Tap             string
	Cask            string
	Formula         string
	Package         string
	Module          string
	URL             string
	Command         []string
	Rollback        []string
	Archive         string
	Extract         bool
	StripComponents int
	TargetDir       string
}

type Resources struct {
	Scripts    []string
	References []string
	Assets     []string
}

type DispatchSpec struct {
	Kind    string
	Tool    string
	ArgMode string
	Command string
	Aliases []string
}

type Skill struct {
	Name             string
	Description      string
	Keywords         []string
	Tags             []string
	WhenToUse        []string
	WhenNotToUse     []string
	NegativeExamples []string
	Steps            []string
	Paths            []string
	ToolHints        []string
	ToolHintsMatch   string
	Examples         []string
	EvalAssertions   []string
	Content          string
	Location         string
	BaseDir          string
	Source           Source
	ExecutionContext string
	AllowedTools     []string
	Effort           string
	Invocation       InvocationPolicy
	Dispatch         *DispatchSpec
	OS               []string
	Requires         Requires
	Install          []InstallSpec
	Resources        Resources
	Metadata         map[string]string
}

type SkillConfig struct {
	Enabled *bool
	Env     map[string]string
}

type Eligibility struct {
	Allowed             map[string]bool
	AllowBundled        map[string]bool
	Denied              map[string]bool
	Config              map[string]SkillConfig
	ResourceConsistency string
	Runtime             RuntimeEligibility
	OnDecision          func(Decision)
}

type RuntimeEligibility struct {
	Platform     string
	HasBin       func(string) bool
	HasAnyBin    func([]string) bool
	HasEnv       func(string) bool
	ConfigTruthy func(string) bool
}

type Decision struct {
	Name    string
	Source  Source
	Include bool
	Reason  string
	Hints   []string
}

func ResolveSkillKey(s Skill) string {
	if s.Metadata != nil {
		if key := strings.ToLower(strings.TrimSpace(s.Metadata["skill_key"])); key != "" {
			return key
		}
	}
	return strings.ToLower(strings.TrimSpace(s.Name))
}

type LoadIssue struct {
	Code    string `json:"code"`
	Source  Source `json:"source,omitempty"`
	Path    string `json:"path,omitempty"`
	Stage   string `json:"stage,omitempty"`
	Message string `json:"message"`
}

type LoadReport struct {
	Loaded      int         `json:"loaded"`
	Skipped     int         `json:"skipped"`
	ParseFailed int         `json:"parse_failed"`
	CacheHit    bool        `json:"cache_hit,omitempty"`
	Generation  uint64      `json:"-"`
	Issues      []LoadIssue `json:"issues,omitempty"`
}

func (r LoadReport) HasIssues() bool {
	return len(r.Issues) > 0 || r.ParseFailed > 0
}

const (
	ToolHintsMatchAny = "any"
	ToolHintsMatchAll = "all"
)

func NormalizeToolHintsMatch(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ToolHintsMatchAny:
		return ToolHintsMatchAny
	case ToolHintsMatchAll:
		return ToolHintsMatchAll
	default:
		return ""
	}
}

func EffectiveToolHintsMatch(raw string) string {
	if mode := NormalizeToolHintsMatch(raw); mode != "" {
		return mode
	}
	return ToolHintsMatchAny
}

func NormalizeSkillExecutionContext(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case SkillExecutionContextInline:
		return SkillExecutionContextInline
	case SkillExecutionContextFork:
		return SkillExecutionContextFork
	default:
		return ""
	}
}

func NormalizeSkillAllowedTools(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, item := range raw {
		name := strings.ToLower(strings.TrimSpace(item))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func NormalizeSkillExecutionEffort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "none":
		return "none"
	case "minimal":
		return "minimal"
	case "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	case "xhigh":
		return "xhigh"
	default:
		return ""
	}
}

// FSSource describes a read-only skill root. ID and Fingerprint jointly form
// its declared identity. The loader enables immutable caching only when FS is
// an assetfs.Provider view that attests the same identity; other valid sources
// remain loadable but intentionally uncached.
type FSSource struct {
	ID          string
	FS          fs.FS
	Fingerprint string
}

// Valid reports whether the source has the fields required for loading. It
// does not by itself attest immutability or cache eligibility.
func (s FSSource) Valid() bool {
	return strings.TrimSpace(s.ID) != "" &&
		s.FS != nil &&
		strings.TrimSpace(s.Fingerprint) != ""
}

type LoadOptions struct {
	CustomDirs               []string
	ExtraDirs                []string
	ExtraFS                  []FSSource
	BundledFS                FSSource
	BundledDir               string
	ManagedDir               string
	WorkspaceDir             string
	MaxCandidatesPerRoot     int
	MaxSkillsLoadedPerSource int
	MaxSkillFileBytes        int
	FailFast                 bool
	StrictFrontmatter        bool
}
