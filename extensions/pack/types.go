package pack

import (
	"fmt"
	"sort"

	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	EvalSuiteModeGate   = "gate"
	EvalSuiteModeShadow = "shadow"
)

const (
	CaseReviewStatusDraft          = "draft"
	CaseReviewStatusReviewRequired = "review_required"
	CaseReviewStatusApproved       = "approved"
	CaseReviewStatusDeprecated     = "deprecated"
)

const (
	SourceAttributionTypePack             = "pack"
	SourceAttributionTypeReferenceProject = "reference_project"
	SourceAttributionTypePublicURL        = "public_url"
	SourceAttributionTypeLocalFile        = "local_file"
	SourceAttributionTypeUserProvided     = "user_provided"
	SourceAttributionTypeGenerated        = "generated"
)

type Manifest struct {
	ID                 string   `json:"id,omitempty"`
	Version            string   `json:"version,omitempty"`
	Domain             string   `json:"domain,omitempty"`
	RouteHints         []string `json:"route_hints,omitempty"`
	SupportedCaseTypes []string `json:"supported_case_types,omitempty"`
	DefaultWorkflow    string   `json:"default_workflow,omitempty"`
	RequiredPlugins    []string `json:"required_plugins,omitempty"`
	OptionalSkills     []string `json:"optional_skills,omitempty"`
	PolicyProfiles     []string `json:"policy_profiles,omitempty"`
	ArtifactTypes      []string `json:"artifact_types,omitempty"`
	Evaluators         []string `json:"evaluators,omitempty"`
	EvalSuites         []string `json:"eval_suites,omitempty"`
}

type CaseSchema struct {
	CaseType    string         `json:"case_type,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
	Description string         `json:"description,omitempty"`
	RouteHints  []string       `json:"route_hints,omitempty"`
}

type CaseLibraryCase struct {
	ID                string                 `json:"id,omitempty"`
	CaseType          string                 `json:"case_type,omitempty"`
	Locale            string                 `json:"locale,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Input             map[string]any         `json:"input,omitempty"`
	InputPlaceholders []CaseInputPlaceholder `json:"input_placeholders,omitempty"`
	ExpectedOutput    map[string]any         `json:"expected_output,omitempty"`
	ReviewStatus      string                 `json:"review_status,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
}

type CaseInputPlaceholder struct {
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Example     any    `json:"example,omitempty"`
}

type SourceAttribution struct {
	SourceType  string `json:"source_type,omitempty"`
	SourceID    string `json:"source_id,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Path        string `json:"path,omitempty"`
	License     string `json:"license,omitempty"`
	RetrievedAt string `json:"retrieved_at,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type PromptTemplate struct {
	Name               string                   `json:"name,omitempty"`
	Description        string                   `json:"description,omitempty"`
	Locale             string                   `json:"locale,omitempty"`
	Template           string                   `json:"template,omitempty"`
	Variables          []PromptTemplateVariable `json:"variables,omitempty"`
	SourceAttributions []SourceAttribution      `json:"source_attributions,omitempty"`
	CaseTypes          []string                 `json:"case_types,omitempty"`
	Tags               []string                 `json:"tags,omitempty"`
}

type PromptTemplateVariable struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Example     any    `json:"example,omitempty"`
}

type PackMediaArtifact struct {
	ID                 string              `json:"id,omitempty"`
	ArtifactType       string              `json:"artifact_type,omitempty"`
	Kind               string              `json:"kind,omitempty"`
	Description        string              `json:"description,omitempty"`
	Path               string              `json:"path,omitempty"`
	URL                string              `json:"url,omitempty"`
	MIMEType           string              `json:"mime_type,omitempty"`
	SourceAttributions []SourceAttribution `json:"source_attributions,omitempty"`
	CaseTypes          []string            `json:"case_types,omitempty"`
	Tags               []string            `json:"tags,omitempty"`
}

type SemanticTool struct {
	Name          string         `json:"name,omitempty"`
	Description   string         `json:"description,omitempty"`
	InputSchema   map[string]any `json:"input_schema,omitempty"`
	OutputSchema  map[string]any `json:"output_schema,omitempty"`
	RuntimeTool   string         `json:"runtime_tool,omitempty"`
	RuntimeArgs   map[string]any `json:"runtime_args,omitempty"`
	ArtifactTypes []string       `json:"artifact_types,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
}

type Evaluator struct {
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

type PolicyProfile struct {
	Name     string                   `json:"name,omitempty"`
	Contract agentxexecution.Contract `json:"contract,omitempty"`
	Default  bool                     `json:"default,omitempty"`
}

type MemorySchema struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
	Default     bool           `json:"default,omitempty"`
}

type MemoryRecallPolicy struct {
	QueryHints []string `json:"query_hints,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	MaxChars   int      `json:"max_chars,omitempty"`
	ScopedOnly bool     `json:"scoped_only,omitempty"`
}

type EvalSuite struct {
	Name              string   `json:"name,omitempty"`
	Description       string   `json:"description,omitempty"`
	Mode              string   `json:"mode,omitempty"`
	WorkflowIDs       []string `json:"workflow_ids,omitempty"`
	RequiredArtifacts []string `json:"required_artifacts,omitempty"`
	RequiredState     []string `json:"required_state,omitempty"`
	PassPath          string   `json:"pass_path,omitempty"`
	ScorePath         string   `json:"score_path,omitempty"`
	MinScore          *float64 `json:"min_score,omitempty"`
	SummaryPath       string   `json:"summary_path,omitempty"`
	Default           bool     `json:"default,omitempty"`
}

type Definition struct {
	Manifest           Manifest              `json:"manifest"`
	CaseSchemas        []CaseSchema          `json:"case_schemas,omitempty"`
	CaseLibrary        []CaseLibraryCase     `json:"case_library,omitempty"`
	PromptTemplates    []PromptTemplate      `json:"prompt_templates,omitempty"`
	MediaArtifacts     []PackMediaArtifact   `json:"media_artifacts,omitempty"`
	Workflows          []agentxworkflow.Spec `json:"workflows,omitempty"`
	Tools              []SemanticTool        `json:"tools,omitempty"`
	Evaluators         []Evaluator           `json:"evaluators,omitempty"`
	EvalSuites         []EvalSuite           `json:"eval_suites,omitempty"`
	PolicyProfiles     []PolicyProfile       `json:"policy_profiles,omitempty"`
	MemorySchemas      []MemorySchema        `json:"memory_schemas,omitempty"`
	MemoryRecallPolicy *MemoryRecallPolicy   `json:"memory_recall_policy,omitempty"`
}

func (m Manifest) SupportsCaseType(caseType string) bool {
	if caseType == "" {
		return true
	}
	for _, item := range m.SupportedCaseTypes {
		if item == caseType {
			return true
		}
	}
	return false
}

func (d Definition) WorkflowByID(id string) (agentxworkflow.Spec, bool) {
	if id == "" {
		return agentxworkflow.Spec{}, false
	}
	for _, spec := range d.Workflows {
		if spec.ID == id {
			return spec, true
		}
	}
	return agentxworkflow.Spec{}, false
}

func (d Definition) DefaultWorkflow() (agentxworkflow.Spec, bool) {
	return d.WorkflowByID(d.Manifest.DefaultWorkflow)
}

func (d Definition) ResolveWorkflowForCaseType(caseType string, workflowID string) (agentxworkflow.Spec, error) {
	if workflowID != "" {
		spec, ok := d.WorkflowByID(workflowID)
		if !ok {
			return agentxworkflow.Spec{}, workflowSelectionError(d.Manifest.ID, caseType, workflowID, "workflow_not_found")
		}
		if !workflowSupportsCaseType(spec, caseType) {
			return agentxworkflow.Spec{}, workflowSelectionError(d.Manifest.ID, caseType, workflowID, "workflow_case_type_mismatch")
		}
		return spec, nil
	}
	if spec, ok := d.DefaultWorkflow(); ok && workflowSupportsCaseType(spec, caseType) {
		return spec, nil
	}
	matches := d.WorkflowsForCaseType(caseType)
	switch len(matches) {
	case 0:
		return agentxworkflow.Spec{}, workflowSelectionError(d.Manifest.ID, caseType, "", "case_type_unbound")
	case 1:
		return matches[0], nil
	default:
		return agentxworkflow.Spec{}, workflowSelectionError(d.Manifest.ID, caseType, "", "case_type_ambiguous")
	}
}

func (d Definition) WorkflowsForCaseType(caseType string) []agentxworkflow.Spec {
	if caseType == "" {
		return nil
	}
	out := make([]agentxworkflow.Spec, 0)
	for _, spec := range d.Workflows {
		if workflowSupportsCaseType(spec, caseType) {
			out = append(out, spec)
		}
	}
	return out
}

func (d Definition) SemanticToolByName(name string) (SemanticTool, bool) {
	if name == "" {
		return SemanticTool{}, false
	}
	for _, tool := range d.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return SemanticTool{}, false
}

func (d Definition) CaseSchemaByType(caseType string) (CaseSchema, bool) {
	if caseType == "" {
		return CaseSchema{}, false
	}
	for _, schema := range d.CaseSchemas {
		if schema.CaseType == caseType {
			return schema, true
		}
	}
	return CaseSchema{}, false
}

func (d Definition) CaseLibraryCaseByID(id string) (CaseLibraryCase, bool) {
	if id == "" {
		return CaseLibraryCase{}, false
	}
	for _, item := range d.CaseLibrary {
		if item.ID == id {
			return cloneCaseLibraryCase(item), true
		}
	}
	return CaseLibraryCase{}, false
}

func (d Definition) CaseLibraryCasesForType(caseType string) []CaseLibraryCase {
	if caseType == "" {
		return nil
	}
	out := make([]CaseLibraryCase, 0)
	for _, item := range d.CaseLibrary {
		if item.CaseType == caseType {
			out = append(out, cloneCaseLibraryCase(item))
		}
	}
	return out
}

func (d Definition) PromptTemplateByName(name string) (PromptTemplate, bool) {
	if name == "" {
		return PromptTemplate{}, false
	}
	for _, item := range d.PromptTemplates {
		if item.Name == name {
			return clonePromptTemplate(item), true
		}
	}
	return PromptTemplate{}, false
}

func (d Definition) MediaArtifactByID(id string) (PackMediaArtifact, bool) {
	if id == "" {
		return PackMediaArtifact{}, false
	}
	for _, item := range d.MediaArtifacts {
		if item.ID == id {
			return clonePackMediaArtifact(item), true
		}
	}
	return PackMediaArtifact{}, false
}

func (d Definition) PolicyProfileByName(name string) (PolicyProfile, bool) {
	if name == "" {
		return PolicyProfile{}, false
	}
	for _, profile := range d.PolicyProfiles {
		if profile.Name == name {
			return profile, true
		}
	}
	return PolicyProfile{}, false
}

func (d Definition) EvaluatorByName(name string) (Evaluator, bool) {
	if name == "" {
		return Evaluator{}, false
	}
	for _, evaluator := range d.Evaluators {
		if evaluator.Name == name {
			return evaluator, true
		}
	}
	return Evaluator{}, false
}

func (d Definition) EvalSuiteByName(name string) (EvalSuite, bool) {
	if name == "" {
		return EvalSuite{}, false
	}
	for _, suite := range d.EvalSuites {
		if suite.Name == name {
			return suite, true
		}
	}
	return EvalSuite{}, false
}

func workflowSupportsCaseType(spec agentxworkflow.Spec, caseType string) bool {
	if caseType == "" || len(spec.CaseTypes) == 0 {
		return true
	}
	for _, item := range spec.CaseTypes {
		if item == caseType {
			return true
		}
	}
	return false
}

func workflowSelectionError(packID string, caseType string, workflowID string, reason string) error {
	switch reason {
	case "workflow_not_found":
		return joinWorkflowSelectionError(packID, caseType, workflowID, "workflow not found")
	case "workflow_case_type_mismatch":
		return joinWorkflowSelectionError(packID, caseType, workflowID, "workflow does not support case type")
	case "case_type_unbound":
		return joinWorkflowSelectionError(packID, caseType, "", "no workflow supports case type")
	case "case_type_ambiguous":
		return joinWorkflowSelectionError(packID, caseType, "", "multiple workflows support case type; explicit workflow_id is required")
	default:
		return joinWorkflowSelectionError(packID, caseType, workflowID, "workflow selection failed")
	}
}

func joinWorkflowSelectionError(packID string, caseType string, workflowID string, detail string) error {
	message := "pack: definition"
	if packID != "" {
		message = fmt.Sprintf("pack: definition %q", packID)
	}
	if caseType != "" {
		message += fmt.Sprintf(" case type %q", caseType)
	}
	if workflowID != "" {
		message += fmt.Sprintf(" workflow %q", workflowID)
	}
	return fmt.Errorf("%s %s", message, detail)
}

func (d Definition) DefaultPolicyProfile() (PolicyProfile, bool) {
	for _, profile := range d.PolicyProfiles {
		if profile.Default {
			return profile, true
		}
	}
	if len(d.PolicyProfiles) == 1 {
		return d.PolicyProfiles[0], true
	}
	return PolicyProfile{}, false
}

func (d Definition) DefaultEvaluator() (Evaluator, bool) {
	if len(d.Evaluators) == 1 {
		return d.Evaluators[0], true
	}
	return Evaluator{}, false
}

func (d Definition) DefaultEvalSuite() (EvalSuite, bool) {
	for _, suite := range d.EvalSuites {
		if suite.Default {
			return suite, true
		}
	}
	if len(d.EvalSuites) == 1 {
		return d.EvalSuites[0], true
	}
	return EvalSuite{}, false
}

func (d Definition) EvalSuiteForWorkflow(workflowID string) (EvalSuite, bool) {
	suites := d.EvalSuitesForWorkflow(workflowID)
	if len(suites) == 0 {
		return EvalSuite{}, false
	}
	return suites[0], true
}

func (d Definition) EvalSuitesForWorkflow(workflowID string) []EvalSuite {
	out := make([]EvalSuite, 0)
	if workflowID != "" {
		for _, suite := range d.EvalSuites {
			for _, candidate := range suite.WorkflowIDs {
				if candidate == workflowID {
					out = append(out, suite)
					break
				}
			}
		}
	}
	if len(out) > 0 {
		return canonicalEvalSuites(out)
	}
	if suite, ok := d.DefaultEvalSuite(); ok {
		return canonicalEvalSuites([]EvalSuite{suite})
	}
	return nil
}

func canonicalEvalSuites(in []EvalSuite) []EvalSuite {
	out := cloneEvalSuites(in)
	sort.SliceStable(out, func(i, j int) bool {
		if evalSuiteModeRank(out[i].Mode) != evalSuiteModeRank(out[j].Mode) {
			return evalSuiteModeRank(out[i].Mode) < evalSuiteModeRank(out[j].Mode)
		}
		if out[i].Default != out[j].Default {
			return out[i].Default
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func evalSuiteModeRank(mode string) int {
	switch exactEvalSuiteMode(mode) {
	case EvalSuiteModeShadow:
		return 1
	default:
		return 0
	}
}

func exactEvalSuiteMode(mode string) string {
	switch mode {
	case "", EvalSuiteModeGate:
		return EvalSuiteModeGate
	case EvalSuiteModeShadow:
		return EvalSuiteModeShadow
	default:
		return mode
	}
}

func NormalizeEvalSuiteMode(mode string) string {
	switch mode {
	case "", EvalSuiteModeGate:
		return EvalSuiteModeGate
	case EvalSuiteModeShadow:
		return EvalSuiteModeShadow
	default:
		return ""
	}
}

func NormalizeCaseReviewStatus(status string) string {
	switch status {
	case CaseReviewStatusDraft, CaseReviewStatusReviewRequired, CaseReviewStatusApproved, CaseReviewStatusDeprecated:
		return status
	default:
		return ""
	}
}

func NormalizeSourceAttributionType(sourceType string) string {
	switch sourceType {
	case SourceAttributionTypePack,
		SourceAttributionTypeReferenceProject,
		SourceAttributionTypePublicURL,
		SourceAttributionTypeLocalFile,
		SourceAttributionTypeUserProvided,
		SourceAttributionTypeGenerated:
		return sourceType
	default:
		return ""
	}
}

func (d Definition) MemorySchemaByName(name string) (MemorySchema, bool) {
	if name == "" {
		return MemorySchema{}, false
	}
	for _, schema := range d.MemorySchemas {
		if schema.Name == name {
			return schema, true
		}
	}
	return MemorySchema{}, false
}

func (d Definition) DefaultMemorySchema() (MemorySchema, bool) {
	for _, schema := range d.MemorySchemas {
		if schema.Default {
			return schema, true
		}
	}
	if len(d.MemorySchemas) == 1 {
		return d.MemorySchemas[0], true
	}
	return MemorySchema{}, false
}
