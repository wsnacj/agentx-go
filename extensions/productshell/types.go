package productshell

import (
	"context"

	agentxcases "github.com/wsnacj/agentx-go/runtime/cases"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"

	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/extensions/skills"
)

// Input is the typed preparation input carried between a host and the
// canonical preparation pipeline.
type Input struct {
	UserMessage           string
	ProductShell          string
	Case                  *agentxcases.Case
	WorkflowSpec          *agentxworkflow.Spec
	RawWorkflowOptIn      bool
	PackID                string
	CaseType              string
	PackWorkflow          string
	WorkspaceDir          string
	RequestedShellBinding *RequestedShellBinding
	RequestedCaseID       string
	RequestedCaseInput    map[string]any
	SessionInput          map[string]any
	WorkflowState         map[string]any
	ShellOptions          InputShellOptions
	Options               map[string]any
}

// InputShellOptions contains preparation options with stable typed ownership.
type InputShellOptions struct {
	AutoCaseBinding         *bool
	AutoWorkflowPlanning    *bool
	RequestedSkills         []string
	RequestedSkillSemantics []RequestedSkillSemantic
	SkillActivationPaths    []string
}

// RequestedSkillSemantic is the canonical skill execution request contract.
type RequestedSkillSemantic = skills.RequestedSkillSemantic

type RequestedShellBinding struct {
	Binding          ShellBinding
	PersistRequested bool
}

type ShellBinding struct {
	PackID        string         `json:"pack_id,omitempty"`
	CaseType      string         `json:"case_type,omitempty"`
	WorkflowID    string         `json:"workflow_id,omitempty"`
	CaseID        string         `json:"case_id,omitempty"`
	CaseInput     map[string]any `json:"case_input,omitempty"`
	SessionInput  map[string]any `json:"session_input,omitempty"`
	WorkflowState map[string]any `json:"workflow_state,omitempty"`
}

type ShellBindingMetrics struct {
	Source            string   `json:"source,omitempty"`
	Matched           bool     `json:"matched,omitempty"`
	PersistRequested  bool     `json:"persist_requested,omitempty"`
	Persisted         bool     `json:"persisted,omitempty"`
	PackID            string   `json:"pack_id,omitempty"`
	CaseType          string   `json:"case_type,omitempty"`
	WorkflowID        string   `json:"workflow_id,omitempty"`
	CaseID            string   `json:"case_id,omitempty"`
	HasCaseInput      bool     `json:"has_case_input,omitempty"`
	HasSessionInput   bool     `json:"has_session_input,omitempty"`
	HasWorkflowState  bool     `json:"has_workflow_state,omitempty"`
	CaseInputKeys     []string `json:"case_input_keys,omitempty"`
	SessionInputKeys  []string `json:"session_input_keys,omitempty"`
	WorkflowStateKeys []string `json:"workflow_state_keys,omitempty"`
}

type PreparedShellBinding struct {
	Binding          ShellBinding        `json:"binding"`
	Source           string              `json:"source,omitempty"`
	Matched          bool                `json:"matched,omitempty"`
	PersistRequested bool                `json:"persist_requested,omitempty"`
	Metrics          ShellBindingMetrics `json:"metrics,omitempty"`
}

type WorkflowBinding struct {
	PackID     string
	CaseType   string
	WorkflowID string
}

type PackSelectionCandidateMetrics struct {
	PackID           string   `json:"pack_id,omitempty"`
	CaseType         string   `json:"case_type,omitempty"`
	WorkflowID       string   `json:"workflow_id,omitempty"`
	WorkflowTitle    string   `json:"workflow_title,omitempty"`
	Score            int      `json:"score,omitempty"`
	MatchedHints     []string `json:"matched_hints,omitempty"`
	MatchedFragments []string `json:"matched_fragments,omitempty"`
	Reasons          []string `json:"reasons,omitempty"`
}

type PackSelectionMetrics struct {
	Attempted      bool                            `json:"attempted,omitempty"`
	Matched        bool                            `json:"matched,omitempty"`
	Applied        bool                            `json:"applied,omitempty"`
	Ambiguous      bool                            `json:"ambiguous,omitempty"`
	Threshold      int                             `json:"threshold,omitempty"`
	CandidateCount int                             `json:"candidate_count,omitempty"`
	Message        string                          `json:"message,omitempty"`
	Selected       PackSelectionCandidateMetrics   `json:"selected"`
	Candidates     []PackSelectionCandidateMetrics `json:"candidates,omitempty"`
	SkipReason     string                          `json:"skip_reason,omitempty"`
}

type CommandDispatchMetrics struct {
	RequestedCommand string              `json:"requested_command,omitempty"`
	PrimaryCommand   string              `json:"primary_command,omitempty"`
	Matched          bool                `json:"matched,omitempty"`
	CacheHit         bool                `json:"cache_hit,omitempty"`
	CandidateCount   int                 `json:"candidate_count,omitempty"`
	ConflictCount    int                 `json:"conflict_count,omitempty"`
	ConflictCommands []string            `json:"conflict_commands,omitempty"`
	ConflictOwners   map[string][]string `json:"conflict_owners,omitempty"`
	Skill            string              `json:"skill,omitempty"`
	Tool             string              `json:"tool,omitempty"`
	ExecutionContext string              `json:"execution_context,omitempty"`
	AllowedTools     []string            `json:"allowed_tools,omitempty"`
	Effort           string              `json:"effort,omitempty"`
}

type PreparedCommandDispatch struct {
	UserMessage            string                 `json:"user_message,omitempty"`
	Skill                  string                 `json:"skill,omitempty"`
	Tool                   string                 `json:"tool,omitempty"`
	RequestedSkillSemantic RequestedSkillSemantic `json:"requested_skill_semantic,omitempty"`
	Metrics                CommandDispatchMetrics `json:"metrics,omitempty"`
	Matched                bool                   `json:"matched,omitempty"`
}

type PreparedPackSelection struct {
	Selection  pack.RouteSelection
	Binding    WorkflowBinding
	Matched    bool
	Applied    bool
	SkipReason string
}

type CaseBindingMetrics struct {
	Attempted            bool     `json:"attempted,omitempty"`
	Matched              bool     `json:"matched,omitempty"`
	Extracted            bool     `json:"extracted,omitempty"`
	Applied              bool     `json:"applied,omitempty"`
	PackMemoryApplied    bool     `json:"pack_memory_applied,omitempty"`
	PackMemoryHits       int      `json:"pack_memory_hits,omitempty"`
	Source               string   `json:"source,omitempty"`
	PackID               string   `json:"pack_id,omitempty"`
	CaseType             string   `json:"case_type,omitempty"`
	WorkflowID           string   `json:"workflow_id,omitempty"`
	ExistingValid        bool     `json:"existing_valid,omitempty"`
	DraftSchema          bool     `json:"draft_schema,omitempty"`
	ExtractedKeys        []string `json:"extracted_keys,omitempty"`
	MergedKeys           []string `json:"merged_keys,omitempty"`
	PackMemorySkipReason string   `json:"pack_memory_skip_reason,omitempty"`
	SkipReason           string   `json:"skip_reason,omitempty"`
}

type PreparedCaseBinding struct {
	Binding       pack.Binding
	Extracted     map[string]any
	Merged        map[string]any
	Applied       bool
	SkipReason    string
	ExistingValid bool
	Metrics       CaseBindingMetrics
}

type ResolvedWorkflow struct {
	Spec        *agentxworkflow.Spec
	PackBinding *pack.Binding
}

type PrepareResult struct {
	Input                  Input
	UserMessage            string
	RequestedSkills        []string
	CommandDispatch        *PreparedCommandDispatch
	Workflow               *ResolvedWorkflow
	EffectiveCase          *agentxcases.Case
	ShellBinding           *PreparedShellBinding
	PackSelection          *PreparedPackSelection
	CommandDispatchMetrics CommandDispatchMetrics
	ShellBindingMetrics    ShellBindingMetrics
	PackSelectionMetrics   PackSelectionMetrics
	CaseBindingMetrics     CaseBindingMetrics
}

type PrepareInput struct {
	SessionID        string
	Input            Input
	LLMTaskTimeoutMs int
}

// PreparationRuntime is the narrow host port used by PreparationPipeline.
// Implementations retain ownership of product policy and concrete backends.
type PreparationRuntime interface {
	ApplyInputCase(Input) Input
	ResolveShellBinding(context.Context, string, Input) (*PreparedShellBinding, error)
	ApplyShellBinding(Input, *PreparedShellBinding) (Input, error)
	ResolveCommandDispatch(context.Context, Input) (*PreparedCommandDispatch, error)
	ApplyCommandDispatch(Input, *PreparedCommandDispatch) Input
	ParseRequestedSkills(Input) ([]string, string)
	ShouldAttemptPackSelection(Input) bool
	ResolvePackSelection(context.Context, Input, string) (*PreparedPackSelection, error)
	ApplyPackSelection(Input, *PreparedPackSelection) Input
	ShouldAttemptCaseBinding(Input) bool
	ResolveCandidateCaseBinding(Input, *PreparedPackSelection) (pack.Binding, bool, error)
	ResolveCaseBindingDraft(context.Context, Input, string, pack.Binding, int) (*PreparedCaseBinding, error)
	MergeCaseBindingMetrics(CaseBindingMetrics, *PreparedCaseBinding) CaseBindingMetrics
	ApplyCaseBinding(Input, *PreparedCaseBinding) Input
	ResolveWorkflow(Input) (ResolvedWorkflow, error)
	ResolveEffectiveCase(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error)
	ValidateEffectiveCase(*pack.Binding, *agentxcases.Case) error
	ApplyEffectiveCase(Input, agentxcases.Case) Input
	FinalizeShellBindingMetrics(ShellBindingMetrics, Input, *pack.Binding, *PreparedShellBinding) ShellBindingMetrics
	PackSelectionMetricsFromPrepared(*PreparedPackSelection) PackSelectionMetrics
}
