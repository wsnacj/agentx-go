package pack

import (
	"encoding/json"
	"fmt"
	"strings"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// MaterializeWorkflow校验定义、选择 Workflow并把 semantic tool确定性映射为
// runtime tool。它不执行工具，也不拥有 provider或具体 backend。
func (c *Coordinator) MaterializeWorkflow(def Definition, workflowID string) (agentxworkflow.Spec, error) {
	if err := c.ValidateDefinition(def); err != nil {
		return agentxworkflow.Spec{}, err
	}
	var (
		spec agentxworkflow.Spec
		ok   bool
	)
	if workflowID == "" {
		spec, ok = def.DefaultWorkflow()
		if !ok {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: default workflow %q not found", def.Manifest.DefaultWorkflow)
		}
	} else {
		spec, ok = def.WorkflowByID(workflowID)
		if !ok {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: workflow %q not found", workflowID)
		}
	}
	validator, err := c.workflowValidator()
	if err != nil {
		return agentxworkflow.Spec{}, err
	}
	lowerer, err := c.toolArgumentLowerer()
	if err != nil {
		return agentxworkflow.Spec{}, err
	}
	return materializeWorkflowSpecWithWorkflowValidator(spec, def.semanticToolIndex(), validator, lowerer)
}

func (r *MemoryRegistry) ResolveMaterializedWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool, error) {
	def, ok := r.Get(packID)
	if !ok {
		return agentxworkflow.Spec{}, false, nil
	}
	if !def.Manifest.SupportsCaseType(caseType) {
		return agentxworkflow.Spec{}, false, nil
	}
	selected, err := def.ResolveWorkflowForCaseType(caseType, "")
	if err != nil {
		return agentxworkflow.Spec{}, false, err
	}
	if r == nil || r.coordinator == nil {
		return agentxworkflow.Spec{}, false, fmt.Errorf("agentx pack: registry is required")
	}
	spec, err := r.coordinator.MaterializeWorkflow(def, selected.ID)
	if err != nil {
		return agentxworkflow.Spec{}, false, err
	}
	return spec, true, nil
}

func materializeWorkflowSpecWithWorkflowValidator(
	spec agentxworkflow.Spec,
	semanticTools map[string]SemanticTool,
	validator agentxworkflow.Validator,
	lowerer ToolArgumentLowerer,
) (agentxworkflow.Spec, error) {
	out := cloneWorkflowSpec(spec)
	for idx := range out.Nodes {
		node := out.Nodes[idx]
		if node.Kind != agentxworkflow.NodeTool {
			continue
		}
		toolName := workflowToolName(node.Config)
		if toolName == "" {
			continue
		}
		semantic, ok := semanticTools[toolName]
		if !ok {
			continue
		}
		if semantic.RuntimeTool == "" {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: semantic tool %q does not declare runtime_tool", toolName)
		}
		config, err := materializeToolConfig(node, semantic, lowerer)
		if err != nil {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: materialize semantic tool %q for node %q: %w", toolName, node.ID, err)
		}
		out.Nodes[idx].Config = config
	}
	if err := validator.ValidateSpec(out); err != nil {
		return agentxworkflow.Spec{}, err
	}
	return out, nil
}

func materializeToolConfig(node agentxworkflow.NodeSpec, semantic SemanticTool, lowerer ToolArgumentLowerer) (map[string]any, error) {
	out := cloneStringAnyMap(node.Config)
	if out == nil {
		out = map[string]any{}
	}
	applySemanticToolRuntimeMetadata(out, semantic)
	runtimeTool := semantic.RuntimeTool
	delete(out, "tool_name")
	out["tool"] = runtimeTool
	defaults := cloneStringAnyMap(semantic.RuntimeArgs)
	if len(defaults) == 0 {
		return out, validateSemanticToolInputSchema(node, out, semantic.InputSchema, lowerer)
	}
	key, args, ok, err := extractArgumentContainer(out)
	if err != nil {
		return nil, err
	}
	if ok {
		if err := writeArgumentContainer(out, key, mergeMapsDefaultFirst(defaults, args)); err != nil {
			return nil, err
		}
		return out, validateSemanticToolInputSchema(node, out, semantic.InputSchema, lowerer)
	}
	for key, value := range defaults {
		if _, exists := out[key]; exists {
			continue
		}
		out[key] = cloneValue(value)
	}
	return out, validateSemanticToolInputSchema(node, out, semantic.InputSchema, lowerer)
}

func validateSemanticToolInputSchema(node agentxworkflow.NodeSpec, config map[string]any, schema map[string]any, lowerer ToolArgumentLowerer) error {
	if len(schema) == 0 {
		return nil
	}
	if err := validatePackSchemaDefinition(schema, "semantic_tool_input"); err != nil {
		return err
	}
	if lowerer == nil {
		return fmt.Errorf("agentx pack: tool argument lowerer is required")
	}
	argumentsJSON, err := lowerer.LowerToolArguments(agentxworkflow.NodeSpec{
		ID:     node.ID,
		Kind:   agentxworkflow.NodeTool,
		Config: cloneStringAnyMap(config),
	})
	if err != nil {
		return fmt.Errorf("lower semantic tool payload: %w", err)
	}
	payload := map[string]any{}
	if argumentsJSON != "" {
		if err := json.Unmarshal([]byte(argumentsJSON), &payload); err != nil {
			return fmt.Errorf("decode semantic tool payload: %w", err)
		}
	}
	for _, binding := range node.Inputs {
		injectSemanticToolInputBindingPlaceholder(payload, schema, binding.To)
	}
	if err := validateSchemaValue(schema, payload, "semantic_tool_input"); err != nil {
		return err
	}
	return nil
}

func injectSemanticToolInputBindingPlaceholder(payload map[string]any, schema map[string]any, target string) {
	target = normalizeSemanticToolInputBindingTarget(target)
	if target == "" {
		return
	}
	setMaterializedPayloadPath(payload, target, semanticToolSchemaPlaceholder(semanticToolSchemaForTarget(schema, target)))
}

func normalizeSemanticToolInputBindingTarget(path string) string {
	if strings.HasPrefix(path, "args.") {
		return strings.TrimPrefix(path, "args.")
	}
	return path
}

func semanticToolSchemaForTarget(schema map[string]any, target string) map[string]any {
	if len(schema) == 0 {
		return nil
	}
	current := schema
	for _, part := range splitMaterializedPayloadPath(target) {
		if !runtimePackSchemaAllowsType(runtimePackSchemaTypes(current), "object") {
			return nil
		}
		next, ok := readSchemaMapMap(current["properties"])[part]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func semanticToolSchemaPlaceholder(schema map[string]any) any {
	if len(schema) == 0 {
		return ""
	}
	schemaTypes := runtimePackSchemaTypes(schema)
	switch {
	case runtimePackSchemaAllowsType(schemaTypes, "object"):
		out := map[string]any{}
		properties := readSchemaMapMap(schema["properties"])
		required, _ := readPackSchemaRequired(schema["required"])
		for _, name := range required {
			childSchema, ok := properties[name]
			if !ok {
				out[name] = ""
				continue
			}
			out[name] = semanticToolSchemaPlaceholder(childSchema)
		}
		return out
	case runtimePackSchemaAllowsType(schemaTypes, "array"):
		return []any{}
	case runtimePackSchemaAllowsType(schemaTypes, "boolean"):
		return false
	case runtimePackSchemaAllowsType(schemaTypes, "integer"):
		return 0
	case runtimePackSchemaAllowsType(schemaTypes, "number"):
		return 0.0
	default:
		return ""
	}
}

func setMaterializedPayloadPath(root map[string]any, path string, value any) {
	parts := splitMaterializedPayloadPath(path)
	if len(parts) == 0 {
		return
	}
	current := root
	for idx, part := range parts {
		if idx == len(parts)-1 {
			current[part] = cloneValue(value)
			return
		}
		next, ok := current[part]
		if !ok {
			child := map[string]any{}
			current[part] = child
			current = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok || child == nil {
			child = map[string]any{}
			current[part] = child
		}
		current = child
	}
}

func splitMaterializedPayloadPath(path string) []string {
	if path == "" {
		return nil
	}
	if path != strings.TrimSpace(path) {
		return nil
	}
	raw := strings.Split(path, ".")
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if part == "" || part != strings.TrimSpace(part) {
			return nil
		}
		out = append(out, part)
	}
	return out
}

func extractArgumentContainer(config map[string]any) (string, map[string]any, bool, error) {
	if len(config) == 0 {
		return "", nil, false, nil
	}
	for _, key := range []string{"arguments_json", "args", "arguments", "payload"} {
		raw, ok := config[key]
		if !ok || raw == nil {
			continue
		}
		switch typed := raw.(type) {
		case map[string]any:
			if key == "arguments_json" {
				return "", nil, false, fmt.Errorf("pack: arguments_json must be a JSON object string")
			}
			return key, cloneStringAnyMap(typed), true, nil
		case string:
			if typed == "" {
				return key, map[string]any{}, true, nil
			}
			trimmed := strings.TrimSpace(typed)
			if trimmed == "" {
				return "", nil, false, fmt.Errorf("pack: %s must not be whitespace-only", key)
			}
			if trimmed != typed && (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) {
				return "", nil, false, fmt.Errorf("pack: %s must not include surrounding whitespace", key)
			}
			if strings.HasPrefix(trimmed, "{") {
				var decoded map[string]any
				if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
					return "", nil, false, fmt.Errorf("decode %s as JSON object: %w", key, err)
				}
				if decoded == nil {
					decoded = map[string]any{}
				}
				return key, decoded, true, nil
			}
			if strings.HasPrefix(trimmed, "[") {
				var decoded any
				if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
					return "", nil, false, fmt.Errorf("decode %s as JSON array: %w", key, err)
				}
				return "", nil, false, fmt.Errorf("pack: %s must be a plain input string or JSON object string, not a JSON array", key)
			}
			if key == "arguments_json" {
				return "", nil, false, fmt.Errorf("pack: arguments_json must be a JSON object string")
			}
			return key, map[string]any{"input": typed}, true, nil
		default:
			if key == "arguments_json" {
				return "", nil, false, fmt.Errorf("pack: arguments_json must be a JSON object string")
			}
			return "", nil, false, fmt.Errorf("pack: %s must be a string or object", key)
		}
	}
	return "", nil, false, nil
}

func writeArgumentContainer(config map[string]any, key string, args map[string]any) error {
	if config == nil {
		return nil
	}
	if key == "arguments_json" {
		raw, err := json.Marshal(args)
		if err != nil {
			return fmt.Errorf("encode arguments_json: %w", err)
		}
		config[key] = string(raw)
		return nil
	}
	config[key] = cloneStringAnyMap(args)
	return nil
}

func mergeMapsDefaultFirst(defaults map[string]any, explicit map[string]any) map[string]any {
	if len(defaults) == 0 && len(explicit) == 0 {
		return map[string]any{}
	}
	out := cloneStringAnyMap(defaults)
	if out == nil {
		out = map[string]any{}
	}
	for key, value := range explicit {
		out[key] = cloneValue(value)
	}
	return out
}

func workflowToolName(config map[string]any) string {
	if len(config) == 0 {
		return ""
	}
	for _, key := range []string{"tool", "tool_name"} {
		raw, ok := config[key]
		if !ok || raw == nil {
			continue
		}
		if text, ok := raw.(string); ok {
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func (d Definition) semanticToolIndex() map[string]SemanticTool {
	if len(d.Tools) == 0 {
		return nil
	}
	out := make(map[string]SemanticTool, len(d.Tools))
	for _, tool := range d.Tools {
		name := tool.Name
		if name == "" {
			continue
		}
		out[name] = cloneSemanticTool(tool)
	}
	return out
}

func cloneDefinition(in Definition) Definition {
	out := in
	out.Manifest.RouteHints = append([]string(nil), in.Manifest.RouteHints...)
	out.Manifest.SupportedCaseTypes = append([]string(nil), in.Manifest.SupportedCaseTypes...)
	out.Manifest.RequiredPlugins = append([]string(nil), in.Manifest.RequiredPlugins...)
	out.Manifest.OptionalSkills = append([]string(nil), in.Manifest.OptionalSkills...)
	out.Manifest.PolicyProfiles = append([]string(nil), in.Manifest.PolicyProfiles...)
	out.Manifest.ArtifactTypes = append([]string(nil), in.Manifest.ArtifactTypes...)
	out.Manifest.Evaluators = append([]string(nil), in.Manifest.Evaluators...)
	out.Manifest.EvalSuites = append([]string(nil), in.Manifest.EvalSuites...)
	out.CaseSchemas = cloneCaseSchemas(in.CaseSchemas)
	out.CaseLibrary = cloneCaseLibrary(in.CaseLibrary)
	out.PromptTemplates = clonePromptTemplates(in.PromptTemplates)
	out.MediaArtifacts = clonePackMediaArtifacts(in.MediaArtifacts)
	out.Workflows = cloneWorkflowSpecs(in.Workflows)
	out.Tools = cloneSemanticTools(in.Tools)
	out.Evaluators = cloneEvaluators(in.Evaluators)
	out.EvalSuites = cloneEvalSuites(in.EvalSuites)
	out.PolicyProfiles = clonePolicyProfiles(in.PolicyProfiles)
	out.MemorySchemas = cloneMemorySchemas(in.MemorySchemas)
	out.MemoryRecallPolicy = cloneMemoryRecallPolicy(in.MemoryRecallPolicy)
	return out
}

func cloneCaseSchemas(in []CaseSchema) []CaseSchema {
	if len(in) == 0 {
		return nil
	}
	out := make([]CaseSchema, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Schema = cloneStringAnyMap(item.Schema)
		out[idx].RouteHints = append([]string(nil), item.RouteHints...)
	}
	return out
}

func cloneCaseLibrary(in []CaseLibraryCase) []CaseLibraryCase {
	if len(in) == 0 {
		return nil
	}
	out := make([]CaseLibraryCase, len(in))
	for idx, item := range in {
		out[idx] = cloneCaseLibraryCase(item)
	}
	return out
}

func cloneCaseLibraryCase(in CaseLibraryCase) CaseLibraryCase {
	out := in
	out.Input = cloneStringAnyMap(in.Input)
	out.InputPlaceholders = cloneCaseInputPlaceholders(in.InputPlaceholders)
	out.ExpectedOutput = cloneStringAnyMap(in.ExpectedOutput)
	out.Tags = append([]string(nil), in.Tags...)
	return out
}

func cloneCaseInputPlaceholders(in []CaseInputPlaceholder) []CaseInputPlaceholder {
	if len(in) == 0 {
		return nil
	}
	out := make([]CaseInputPlaceholder, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Example = cloneValue(item.Example)
	}
	return out
}

func clonePromptTemplates(in []PromptTemplate) []PromptTemplate {
	if len(in) == 0 {
		return nil
	}
	out := make([]PromptTemplate, len(in))
	for idx, item := range in {
		out[idx] = clonePromptTemplate(item)
	}
	return out
}

func clonePromptTemplate(in PromptTemplate) PromptTemplate {
	out := in
	out.Variables = clonePromptTemplateVariables(in.Variables)
	out.SourceAttributions = cloneSourceAttributions(in.SourceAttributions)
	out.CaseTypes = append([]string(nil), in.CaseTypes...)
	out.Tags = append([]string(nil), in.Tags...)
	return out
}

func clonePromptTemplateVariables(in []PromptTemplateVariable) []PromptTemplateVariable {
	if len(in) == 0 {
		return nil
	}
	out := make([]PromptTemplateVariable, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Example = cloneValue(item.Example)
	}
	return out
}

func clonePackMediaArtifacts(in []PackMediaArtifact) []PackMediaArtifact {
	if len(in) == 0 {
		return nil
	}
	out := make([]PackMediaArtifact, len(in))
	for idx, item := range in {
		out[idx] = clonePackMediaArtifact(item)
	}
	return out
}

func clonePackMediaArtifact(in PackMediaArtifact) PackMediaArtifact {
	out := in
	out.SourceAttributions = cloneSourceAttributions(in.SourceAttributions)
	out.CaseTypes = append([]string(nil), in.CaseTypes...)
	out.Tags = append([]string(nil), in.Tags...)
	return out
}

func cloneSourceAttributions(in []SourceAttribution) []SourceAttribution {
	if len(in) == 0 {
		return nil
	}
	return append([]SourceAttribution(nil), in...)
}

func cloneWorkflowSpecs(in []agentxworkflow.Spec) []agentxworkflow.Spec {
	if len(in) == 0 {
		return nil
	}
	out := make([]agentxworkflow.Spec, len(in))
	for idx, item := range in {
		out[idx] = cloneWorkflowSpec(item)
	}
	return out
}

func cloneWorkflowSpec(in agentxworkflow.Spec) agentxworkflow.Spec {
	out := in
	out.CaseTypes = append([]string(nil), in.CaseTypes...)
	out.RouteHints = append([]string(nil), in.RouteHints...)
	out.Nodes = cloneNodeSpecs(in.Nodes)
	out.Edges = append([]agentxworkflow.EdgeSpec(nil), in.Edges...)
	out.StateSchema = append([]agentxworkflow.StateSlotSpec(nil), in.StateSchema...)
	out.ArtifactSchema = append([]agentxworkflow.ArtifactTypeRef(nil), in.ArtifactSchema...)
	out.EvaluatorSchema = append([]agentxworkflow.EvaluatorRef(nil), in.EvaluatorSchema...)
	return out
}

func cloneNodeSpecs(in []agentxworkflow.NodeSpec) []agentxworkflow.NodeSpec {
	if len(in) == 0 {
		return nil
	}
	out := make([]agentxworkflow.NodeSpec, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Inputs = append([]agentxworkflow.BindingSpec(nil), item.Inputs...)
		out[idx].Outputs = append([]agentxworkflow.BindingSpec(nil), item.Outputs...)
		out[idx].Retry.BackoffMs = append([]int(nil), item.Retry.BackoffMs...)
		out[idx].Config = cloneStringAnyMap(item.Config)
	}
	return out
}

func cloneSemanticTools(in []SemanticTool) []SemanticTool {
	if len(in) == 0 {
		return nil
	}
	out := make([]SemanticTool, len(in))
	for idx, item := range in {
		out[idx] = cloneSemanticTool(item)
	}
	return out
}

func cloneSemanticTool(in SemanticTool) SemanticTool {
	out := in
	out.InputSchema = cloneStringAnyMap(in.InputSchema)
	out.OutputSchema = cloneStringAnyMap(in.OutputSchema)
	out.RuntimeArgs = cloneStringAnyMap(in.RuntimeArgs)
	out.ArtifactTypes = append([]string(nil), in.ArtifactTypes...)
	out.Tags = append([]string(nil), in.Tags...)
	return out
}

func cloneEvaluators(in []Evaluator) []Evaluator {
	if len(in) == 0 {
		return nil
	}
	out := make([]Evaluator, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].OutputSchema = cloneStringAnyMap(item.OutputSchema)
	}
	return out
}

func cloneEvalSuites(in []EvalSuite) []EvalSuite {
	if len(in) == 0 {
		return nil
	}
	out := make([]EvalSuite, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Mode = exactEvalSuiteMode(item.Mode)
		if item.MinScore != nil {
			value := *item.MinScore
			out[idx].MinScore = &value
		}
		out[idx].WorkflowIDs = append([]string(nil), item.WorkflowIDs...)
		out[idx].RequiredArtifacts = append([]string(nil), item.RequiredArtifacts...)
		out[idx].RequiredState = append([]string(nil), item.RequiredState...)
	}
	return out
}

func clonePolicyProfiles(in []PolicyProfile) []PolicyProfile {
	if len(in) == 0 {
		return nil
	}
	out := make([]PolicyProfile, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Contract.Visibility.AllowTools = append([]string(nil), item.Contract.Visibility.AllowTools...)
		out[idx].Contract.Visibility.DenyTools = append([]string(nil), item.Contract.Visibility.DenyTools...)
		out[idx].Contract.Visibility.DeclaredTools = append([]string(nil), item.Contract.Visibility.DeclaredTools...)
		out[idx].Contract.Approval.AllowTools = append([]string(nil), item.Contract.Approval.AllowTools...)
		out[idx].Contract.Approval.DenyTools = append([]string(nil), item.Contract.Approval.DenyTools...)
		out[idx].Contract.Replay.AllowTools = append([]string(nil), item.Contract.Replay.AllowTools...)
		out[idx].Contract.Replay.DenyTools = append([]string(nil), item.Contract.Replay.DenyTools...)
		out[idx].Contract.Replay.AllowIdempotencyLevels = append([]string(nil), item.Contract.Replay.AllowIdempotencyLevels...)
		out[idx].Contract.Replay.AllowIdempotencyByEnv = cloneStringSliceMap(item.Contract.Replay.AllowIdempotencyByEnv)
		out[idx].Contract.Replay.ApprovalAllowTools = append([]string(nil), item.Contract.Replay.ApprovalAllowTools...)
		out[idx].Contract.Replay.ApprovalDenyTools = append([]string(nil), item.Contract.Replay.ApprovalDenyTools...)
		out[idx].Contract.Replay.AutoAllowTools = append([]string(nil), item.Contract.Replay.AutoAllowTools...)
		out[idx].Contract.Replay.AutoAllowByEnv = cloneStringSliceMap(item.Contract.Replay.AutoAllowByEnv)
		out[idx].Contract.SideEffects.CrossSystemConfirmTools = append([]string(nil), item.Contract.SideEffects.CrossSystemConfirmTools...)
		out[idx].Contract.Sandbox.ExecAllowlist = append([]string(nil), item.Contract.Sandbox.ExecAllowlist...)
		out[idx].Contract.Sandbox.ExecDenyPatterns = append([]string(nil), item.Contract.Sandbox.ExecDenyPatterns...)
		out[idx].Contract.Sandbox.ProcessSignals = append([]string(nil), item.Contract.Sandbox.ProcessSignals...)
		out[idx].Contract.Sandbox.BrowserProxyActKinds = append([]string(nil), item.Contract.Sandbox.BrowserProxyActKinds...)
		out[idx].Contract.Evidence.RequiredArtifacts = append([]string(nil), item.Contract.Evidence.RequiredArtifacts...)
	}
	return out
}

func cloneMemorySchemas(in []MemorySchema) []MemorySchema {
	if len(in) == 0 {
		return nil
	}
	out := make([]MemorySchema, len(in))
	for idx, item := range in {
		out[idx] = item
		out[idx].Schema = cloneStringAnyMap(item.Schema)
	}
	return out
}

func cloneMemoryRecallPolicy(in *MemoryRecallPolicy) *MemoryRecallPolicy {
	if in == nil {
		return nil
	}
	out := *in
	out.QueryHints = append([]string(nil), in.QueryHints...)
	return &out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func cloneStringAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(in any) any {
	switch typed := in.(type) {
	case map[string]any:
		return cloneStringAnyMap(typed)
	case []any:
		out := make([]any, len(typed))
		for idx, item := range typed {
			out[idx] = cloneValue(item)
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	case []int:
		return append([]int(nil), typed...)
	default:
		return typed
	}
}
