package pack

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// ValidateDefinition使用构造 Coordinator时注入的 Workflow validator，按稳定顺序
// 校验完整 Pack定义。validator返回的错误会保留 identity并按既有上下文包装。
func (c *Coordinator) ValidateDefinition(def Definition) error {
	validator, err := c.workflowValidator()
	if err != nil {
		return err
	}
	return validateDefinitionWithWorkflowValidator(def, validator)
}

func validateDefinitionWithWorkflowValidator(def Definition, validator agentxworkflow.Validator) error {
	manifest := def.Manifest
	if err := validateRequiredPackField(manifest.ID, "manifest id"); err != nil {
		return err
	}
	if err := validateRequiredPackField(manifest.Version, "manifest version"); err != nil {
		return err
	}
	if err := validateRequiredPackField(manifest.Domain, "manifest domain"); err != nil {
		return err
	}
	if len(manifest.SupportedCaseTypes) == 0 {
		return fmt.Errorf("pack: at least one supported case type is required")
	}
	if err := validateRequiredPackField(manifest.DefaultWorkflow, "default workflow"); err != nil {
		return err
	}
	if len(def.Workflows) == 0 {
		return fmt.Errorf("pack: at least one workflow is required")
	}
	if err := validateUniqueStrings(manifest.SupportedCaseTypes, "supported case type"); err != nil {
		return err
	}
	if err := validateOptionalUniqueStrings(manifest.RouteHints, "manifest route hint"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.RequiredPlugins, "required plugin"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.OptionalSkills, "optional skill"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.PolicyProfiles, "manifest policy profile"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.ArtifactTypes, "artifact type"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.Evaluators, "manifest evaluator"); err != nil {
		return err
	}
	if err := validateUniqueStrings(manifest.EvalSuites, "manifest eval suite"); err != nil {
		return err
	}
	if err := validateCaseSchemas(def.CaseSchemas, manifest); err != nil {
		return err
	}
	if err := validateCaseLibrary(def.CaseLibrary, manifest); err != nil {
		return err
	}
	if err := validatePromptTemplates(def.PromptTemplates, manifest); err != nil {
		return err
	}
	if err := validatePackMediaArtifacts(def.MediaArtifacts, manifest); err != nil {
		return err
	}
	if err := validateWorkflows(def.Workflows, manifest, def, validator); err != nil {
		return err
	}
	if err := validateSupportedCaseTypeWorkflowCoverage(manifest, def.Workflows); err != nil {
		return err
	}
	if _, ok := def.DefaultWorkflow(); !ok {
		return fmt.Errorf("pack: default workflow %q not found", strings.TrimSpace(manifest.DefaultWorkflow))
	}
	if err := validateSemanticTools(def.Tools, manifest); err != nil {
		return err
	}
	if err := validateEvaluators(def.Evaluators, manifest); err != nil {
		return err
	}
	if err := validateEvalSuites(def.EvalSuites, manifest, def.Workflows); err != nil {
		return err
	}
	if err := validatePackWorkflowEvalSuiteCoverage(def); err != nil {
		return err
	}
	if err := validatePolicyProfiles(def.PolicyProfiles, manifest); err != nil {
		return err
	}
	if err := validateMemorySchemas(def.MemorySchemas); err != nil {
		return err
	}
	if err := validateMemoryRecallPolicy(def.MemoryRecallPolicy); err != nil {
		return err
	}
	return nil
}

func validateUniqueStrings(items []string, label string) error {
	seen := map[string]bool{}
	for _, raw := range items {
		if err := validateRequiredPackField(raw, label); err != nil {
			return err
		}
		value := raw
		if seen[value] {
			return fmt.Errorf("pack: duplicate %s %q", label, value)
		}
		seen[value] = true
	}
	return nil
}

func validateCaseSchemas(schemas []CaseSchema, manifest Manifest) error {
	seen := map[string]bool{}
	for _, schema := range schemas {
		if err := validateRequiredPackField(schema.CaseType, "case schema case_type"); err != nil {
			return err
		}
		caseType := schema.CaseType
		if seen[caseType] {
			return fmt.Errorf("pack: duplicate case schema %q", caseType)
		}
		if !manifest.SupportsCaseType(caseType) {
			return fmt.Errorf("pack: case schema %q is not declared in manifest.supported_case_types", caseType)
		}
		if err := validateOptionalPackField(schema.Description, fmt.Sprintf("case schema %q description", caseType)); err != nil {
			return err
		}
		if err := validateOptionalPackSchema(schema.Schema, fmt.Sprintf("case schema %q schema", caseType)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(schema.RouteHints, fmt.Sprintf("case schema %q route hint", caseType)); err != nil {
			return err
		}
		seen[caseType] = true
	}
	return nil
}

func validateCaseLibrary(items []CaseLibraryCase, manifest Manifest) error {
	seen := map[string]bool{}
	for _, item := range items {
		if err := validateRequiredPackField(item.ID, "case library id"); err != nil {
			return err
		}
		id := item.ID
		if seen[id] {
			return fmt.Errorf("pack: duplicate case library id %q", id)
		}
		if err := validateRequiredPackField(item.CaseType, fmt.Sprintf("case library %q case_type", id)); err != nil {
			return err
		}
		if !manifest.SupportsCaseType(item.CaseType) {
			return fmt.Errorf("pack: case library %q case_type %q is not declared in manifest.supported_case_types", id, item.CaseType)
		}
		if err := validateRequiredPackField(item.Locale, fmt.Sprintf("case library %q locale", id)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Description, fmt.Sprintf("case library %q description", id)); err != nil {
			return err
		}
		if len(item.Input) == 0 {
			return fmt.Errorf("pack: case library %q input is required", id)
		}
		if err := validateCaseInputPlaceholders(item.InputPlaceholders, id); err != nil {
			return err
		}
		if len(item.ExpectedOutput) == 0 {
			return fmt.Errorf("pack: case library %q expected_output is required", id)
		}
		if err := validateCaseLibraryExpectedOutput(item.ExpectedOutput, id); err != nil {
			return err
		}
		if err := validateRequiredPackField(item.ReviewStatus, fmt.Sprintf("case library %q review_status", id)); err != nil {
			return err
		}
		if NormalizeCaseReviewStatus(item.ReviewStatus) == "" {
			return fmt.Errorf("pack: case library %q review_status %q is not supported", id, item.ReviewStatus)
		}
		if err := validateOptionalUniqueStrings(item.Tags, fmt.Sprintf("case library %q tag", id)); err != nil {
			return err
		}
		seen[id] = true
	}
	return nil
}

func validateCaseLibraryExpectedOutput(items map[string]any, caseID string) error {
	for path := range items {
		if err := validateRequiredPackField(path, fmt.Sprintf("case library %q expected_output path", caseID)); err != nil {
			return err
		}
	}
	return nil
}

func validateCaseInputPlaceholders(items []CaseInputPlaceholder, caseID string) error {
	seen := map[string]bool{}
	for _, item := range items {
		if err := validateRequiredPackField(item.Name, fmt.Sprintf("case library %q input placeholder name", caseID)); err != nil {
			return err
		}
		name := item.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate case library %q input placeholder %q", caseID, name)
		}
		if err := validateOptionalPackField(item.Path, fmt.Sprintf("case library %q input placeholder %q path", caseID, name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Description, fmt.Sprintf("case library %q input placeholder %q description", caseID, name)); err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

func validatePromptTemplates(items []PromptTemplate, manifest Manifest) error {
	seen := map[string]bool{}
	for _, item := range items {
		if err := validateRequiredPackField(item.Name, "prompt template name"); err != nil {
			return err
		}
		name := item.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate prompt template %q", name)
		}
		if err := validateOptionalPackField(item.Description, fmt.Sprintf("prompt template %q description", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Locale, fmt.Sprintf("prompt template %q locale", name)); err != nil {
			return err
		}
		if err := validateRequiredPackField(item.Template, fmt.Sprintf("prompt template %q template", name)); err != nil {
			return err
		}
		if err := validatePromptTemplateVariables(item.Variables, name); err != nil {
			return err
		}
		if err := validateSourceAttributions(item.SourceAttributions, fmt.Sprintf("prompt template %q", name)); err != nil {
			return err
		}
		if err := validateCaseTypeRefs(item.CaseTypes, manifest, fmt.Sprintf("prompt template %q case type", name)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(item.Tags, fmt.Sprintf("prompt template %q tag", name)); err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

func validatePromptTemplateVariables(items []PromptTemplateVariable, templateName string) error {
	seen := map[string]bool{}
	for _, item := range items {
		if err := validateRequiredPackField(item.Name, fmt.Sprintf("prompt template %q variable name", templateName)); err != nil {
			return err
		}
		name := item.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate prompt template %q variable %q", templateName, name)
		}
		if err := validateOptionalPackField(item.Description, fmt.Sprintf("prompt template %q variable %q description", templateName, name)); err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

func validatePackMediaArtifacts(items []PackMediaArtifact, manifest Manifest) error {
	seen := map[string]bool{}
	artifactTypes := map[string]bool{}
	for _, item := range manifest.ArtifactTypes {
		if artifactType := strings.TrimSpace(item); artifactType != "" {
			artifactTypes[artifactType] = true
		}
	}
	for _, item := range items {
		if err := validateRequiredPackField(item.ID, "media artifact id"); err != nil {
			return err
		}
		id := item.ID
		if seen[id] {
			return fmt.Errorf("pack: duplicate media artifact %q", id)
		}
		if err := validateOptionalPackField(item.ArtifactType, fmt.Sprintf("media artifact %q artifact_type", id)); err != nil {
			return err
		}
		if item.ArtifactType != "" && !artifactTypes[item.ArtifactType] {
			return fmt.Errorf("pack: media artifact %q references undeclared artifact type %q", id, item.ArtifactType)
		}
		if err := validateRequiredPackField(item.Kind, fmt.Sprintf("media artifact %q kind", id)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Description, fmt.Sprintf("media artifact %q description", id)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Path, fmt.Sprintf("media artifact %q path", id)); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.URL, fmt.Sprintf("media artifact %q url", id)); err != nil {
			return err
		}
		if item.Path == "" && item.URL == "" {
			return fmt.Errorf("pack: media artifact %q path or url is required", id)
		}
		if err := validateOptionalPackField(item.MIMEType, fmt.Sprintf("media artifact %q mime_type", id)); err != nil {
			return err
		}
		if err := validateSourceAttributions(item.SourceAttributions, fmt.Sprintf("media artifact %q", id)); err != nil {
			return err
		}
		if err := validateCaseTypeRefs(item.CaseTypes, manifest, fmt.Sprintf("media artifact %q case type", id)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(item.Tags, fmt.Sprintf("media artifact %q tag", id)); err != nil {
			return err
		}
		seen[id] = true
	}
	return nil
}

func validateSourceAttributions(items []SourceAttribution, label string) error {
	if len(items) == 0 {
		return fmt.Errorf("pack: %s source_attributions is required", label)
	}
	seen := map[string]bool{}
	for idx, item := range items {
		itemLabel := fmt.Sprintf("%s source_attributions[%d]", label, idx)
		if err := validateRequiredPackField(item.SourceType, itemLabel+" source_type"); err != nil {
			return err
		}
		if NormalizeSourceAttributionType(item.SourceType) == "" {
			return fmt.Errorf("pack: %s source_type %q is not supported", itemLabel, item.SourceType)
		}
		if err := validateRequiredPackField(item.SourceID, itemLabel+" source_id"); err != nil {
			return err
		}
		if err := validateRequiredPackField(item.Title, itemLabel+" title"); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.URL, itemLabel+" url"); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Path, itemLabel+" path"); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.License, itemLabel+" license"); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.RetrievedAt, itemLabel+" retrieved_at"); err != nil {
			return err
		}
		if err := validateOptionalPackField(item.Notes, itemLabel+" notes"); err != nil {
			return err
		}
		key := item.SourceType + "\x00" + item.SourceID
		if seen[key] {
			return fmt.Errorf("pack: duplicate %s source attribution %q/%q", label, item.SourceType, item.SourceID)
		}
		seen[key] = true
	}
	return nil
}

func validateCaseTypeRefs(items []string, manifest Manifest, label string) error {
	if err := validateOptionalUniqueStrings(items, label); err != nil {
		return err
	}
	for _, item := range items {
		if !manifest.SupportsCaseType(item) {
			return fmt.Errorf("pack: %s %q is not declared in manifest.supported_case_types", label, item)
		}
	}
	return nil
}

func validateWorkflows(
	workflows []agentxworkflow.Spec,
	manifest Manifest,
	def Definition,
	validator agentxworkflow.Validator,
) error {
	seen := map[string]bool{}
	policyProfiles := make(map[string]bool, len(def.PolicyProfiles))
	evaluators := make(map[string]bool, len(def.Evaluators))
	for _, profile := range def.PolicyProfiles {
		if name := strings.TrimSpace(profile.Name); name != "" {
			policyProfiles[name] = true
		}
	}
	for _, evaluator := range def.Evaluators {
		if name := strings.TrimSpace(evaluator.Name); name != "" {
			evaluators[name] = true
		}
	}
	for _, spec := range workflows {
		workflowID := strings.TrimSpace(spec.ID)
		if workflowID == "" {
			return fmt.Errorf("pack: workflow id is required")
		}
		if seen[workflowID] {
			return fmt.Errorf("pack: duplicate workflow %q", workflowID)
		}
		if packID := strings.TrimSpace(spec.Pack); packID != "" && packID != strings.TrimSpace(manifest.ID) {
			return fmt.Errorf("pack: workflow %q declares pack %q, expected %q", workflowID, packID, strings.TrimSpace(manifest.ID))
		}
		if err := validateOptionalUniqueStrings(spec.RouteHints, fmt.Sprintf("workflow %q route hint", workflowID)); err != nil {
			return err
		}
		if err := validateWorkflowArtifactSchema(spec, manifest); err != nil {
			return fmt.Errorf("pack: workflow %q invalid: %w", workflowID, err)
		}
		if contractRef := strings.TrimSpace(spec.DefaultContract); contractRef != "" && !policyProfiles[contractRef] {
			return fmt.Errorf("pack: workflow %q references unknown default contract %q", workflowID, contractRef)
		}
		for _, ref := range spec.EvaluatorSchema {
			name := strings.TrimSpace(ref.Name)
			if name == "" {
				continue
			}
			if !evaluators[name] {
				return fmt.Errorf("pack: workflow %q references unknown evaluator %q", workflowID, name)
			}
		}
		for _, node := range spec.Nodes {
			if ref := strings.TrimSpace(node.ContractRef); ref != "" && !policyProfiles[ref] {
				return fmt.Errorf("pack: workflow %q node %q references unknown contract_ref %q", workflowID, strings.TrimSpace(node.ID), ref)
			}
			if node.Kind != agentxworkflow.NodeEvaluate {
				continue
			}
			if evaluatorName := strings.TrimSpace(readBindingConfigString(node.Config, "evaluator")); evaluatorName != "" && !evaluators[evaluatorName] {
				return fmt.Errorf("pack: workflow %q evaluate node %q references unknown evaluator %q", workflowID, strings.TrimSpace(node.ID), evaluatorName)
			}
		}
		for _, caseType := range spec.CaseTypes {
			caseType = strings.TrimSpace(caseType)
			if caseType == "" {
				return fmt.Errorf("pack: workflow %q contains empty case type", workflowID)
			}
			if !manifest.SupportsCaseType(caseType) {
				return fmt.Errorf("pack: workflow %q references unsupported case type %q", workflowID, caseType)
			}
		}
		if err := validator.ValidateSpec(spec); err != nil {
			return fmt.Errorf("pack: workflow %q invalid: %w", workflowID, err)
		}
		seen[workflowID] = true
	}
	return nil
}

func validateOptionalUniqueStrings(items []string, label string) error {
	seen := map[string]bool{}
	for _, raw := range items {
		if err := validateRequiredPackField(raw, label); err != nil {
			return err
		}
		value := raw
		if seen[value] {
			return fmt.Errorf("pack: duplicate %s %q", label, value)
		}
		seen[value] = true
	}
	return nil
}

func validateRequiredPackField(value string, label string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("pack: %s is required", label)
	}
	if value != trimmed {
		return fmt.Errorf("pack: %s %q must not include surrounding whitespace", label, value)
	}
	return nil
}

func validateOptionalPackField(value string, label string) error {
	if value == "" {
		return nil
	}
	return validateRequiredPackField(value, label)
}

func validateSupportedCaseTypeWorkflowCoverage(manifest Manifest, workflows []agentxworkflow.Spec) error {
	for _, caseType := range manifest.SupportedCaseTypes {
		caseType = strings.TrimSpace(caseType)
		if caseType == "" {
			continue
		}
		covered := false
		for _, spec := range workflows {
			if workflowSupportsCaseType(spec, caseType) {
				covered = true
				break
			}
		}
		if !covered {
			return fmt.Errorf("pack: supported case type %q is not covered by any workflow", caseType)
		}
	}
	return nil
}

func validateSemanticTools(tools []SemanticTool, manifest Manifest) error {
	seen := map[string]bool{}
	manifestArtifactTypes := make(map[string]bool, len(manifest.ArtifactTypes))
	for _, item := range manifest.ArtifactTypes {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			manifestArtifactTypes[trimmed] = true
		}
	}
	for _, tool := range tools {
		if err := validateRequiredPackField(tool.Name, "semantic tool name"); err != nil {
			return err
		}
		name := tool.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate semantic tool %q", name)
		}
		if err := validateOptionalPackField(tool.Description, fmt.Sprintf("semantic tool %q description", name)); err != nil {
			return err
		}
		if err := validateOptionalPackSchema(tool.InputSchema, fmt.Sprintf("semantic tool %q input_schema", name)); err != nil {
			return err
		}
		if err := validateOptionalPackSchema(tool.OutputSchema, fmt.Sprintf("semantic tool %q output_schema", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(tool.RuntimeTool, fmt.Sprintf("semantic tool %q runtime_tool", name)); err != nil {
			return err
		}
		if tool.RuntimeTool == "" && len(tool.RuntimeArgs) > 0 {
			return fmt.Errorf("pack: semantic tool %q declares runtime_args without runtime_tool", name)
		}
		if err := validateSemanticToolRuntimeArgs(tool.RuntimeArgs, name); err != nil {
			return err
		}
		if err := validateSemanticToolRuntimeArgsAgainstInputSchema(tool.RuntimeArgs, tool.InputSchema, name); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(tool.ArtifactTypes, fmt.Sprintf("semantic tool %q artifact type", name)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(tool.Tags, fmt.Sprintf("semantic tool %q tag", name)); err != nil {
			return err
		}
		for _, artifactType := range tool.ArtifactTypes {
			artifactType = strings.TrimSpace(artifactType)
			if artifactType == "" {
				continue
			}
			if !manifestArtifactTypes[artifactType] {
				return fmt.Errorf("pack: semantic tool %q references undeclared artifact type %q", name, artifactType)
			}
		}
		seen[name] = true
	}
	return nil
}

func validateSemanticToolRuntimeArgs(runtimeArgs map[string]any, toolName string) error {
	if len(runtimeArgs) == 0 {
		return nil
	}
	for _, key := range []string{
		"arguments_json", "args", "arguments", "payload", "input",
		"tool", "tool_name",
		semanticToolNameConfigKey, artifactTypesConfigKey, semanticToolTagsConfigKey,
	} {
		if _, exists := runtimeArgs[key]; exists {
			return fmt.Errorf("pack: semantic tool %q runtime_args.%s is not allowed; runtime_args must use canonical payload keys only", toolName, key)
		}
	}
	return nil
}

func validateSemanticToolRuntimeArgsAgainstInputSchema(runtimeArgs map[string]any, schema map[string]any, toolName string) error {
	if len(runtimeArgs) == 0 || len(schema) == 0 {
		return nil
	}
	if err := validateSchemaProvidedValue(schema, runtimeArgs, "semantic_tool_runtime_args"); err != nil {
		return fmt.Errorf("pack: semantic tool %q runtime_args invalid against input_schema: %w", toolName, err)
	}
	return nil
}

func validateWorkflowArtifactSchema(spec agentxworkflow.Spec, manifest Manifest) error {
	seen := map[string]bool{}
	manifestArtifactTypes := make(map[string]bool, len(manifest.ArtifactTypes))
	for _, item := range manifest.ArtifactTypes {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			manifestArtifactTypes[trimmed] = true
		}
	}
	for _, ref := range spec.ArtifactSchema {
		artifactType := strings.TrimSpace(ref.Type)
		if artifactType == "" {
			return fmt.Errorf("workflow artifact_schema type is required")
		}
		if seen[artifactType] {
			return fmt.Errorf("duplicate workflow artifact_schema type %q", artifactType)
		}
		if !manifestArtifactTypes[artifactType] {
			return fmt.Errorf("workflow artifact_schema type %q is not declared in manifest.artifact_types", artifactType)
		}
		seen[artifactType] = true
	}
	return nil
}

func validateEvaluators(items []Evaluator, manifest Manifest) error {
	seen := map[string]bool{}
	for _, evaluator := range items {
		if err := validateRequiredPackField(evaluator.Name, "evaluator name"); err != nil {
			return err
		}
		name := evaluator.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate evaluator %q", name)
		}
		if err := validateOptionalPackField(evaluator.Description, fmt.Sprintf("evaluator %q description", name)); err != nil {
			return err
		}
		if err := validateOptionalPackSchema(evaluator.OutputSchema, fmt.Sprintf("evaluator %q output_schema", name)); err != nil {
			return err
		}
		seen[name] = true
	}
	for _, name := range manifest.Evaluators {
		if !seen[strings.TrimSpace(name)] {
			return fmt.Errorf("pack: manifest evaluator %q is not defined", strings.TrimSpace(name))
		}
	}
	return nil
}

func validateEvalSuites(items []EvalSuite, manifest Manifest, workflows []agentxworkflow.Spec) error {
	seen := map[string]bool{}
	workflowIDs := map[string]bool{}
	artifactTypes := map[string]bool{}
	defaultCount := 0
	for _, spec := range workflows {
		if workflowID := strings.TrimSpace(spec.ID); workflowID != "" {
			workflowIDs[workflowID] = true
		}
	}
	for _, item := range manifest.ArtifactTypes {
		if artifactType := strings.TrimSpace(item); artifactType != "" {
			artifactTypes[artifactType] = true
		}
	}
	for _, suite := range items {
		if err := validateRequiredPackField(suite.Name, "eval suite name"); err != nil {
			return err
		}
		name := suite.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate eval suite %q", name)
		}
		if err := validateOptionalPackField(suite.Description, fmt.Sprintf("eval suite %q description", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(suite.Mode, fmt.Sprintf("eval suite %q mode", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(suite.PassPath, fmt.Sprintf("eval suite %q pass_path", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(suite.ScorePath, fmt.Sprintf("eval suite %q score_path", name)); err != nil {
			return err
		}
		if err := validateOptionalPackField(suite.SummaryPath, fmt.Sprintf("eval suite %q summary_path", name)); err != nil {
			return err
		}
		if suite.Default {
			defaultCount++
		}
		if mode := NormalizeEvalSuiteMode(suite.Mode); mode == "" {
			return fmt.Errorf("pack: eval suite %q has unsupported mode %q", name, strings.TrimSpace(suite.Mode))
		} else if suite.Mode != "" && suite.Mode != mode {
			return fmt.Errorf("pack: eval suite %q mode %q must use canonical lowercase", name, suite.Mode)
		}
		scorePath := strings.TrimSpace(suite.ScorePath)
		switch {
		case scorePath == "" && suite.MinScore != nil:
			return fmt.Errorf("pack: eval suite %q declares min_score without score_path", name)
		case scorePath != "" && suite.MinScore == nil:
			return fmt.Errorf("pack: eval suite %q declares score_path without min_score", name)
		}
		if suite.MinScore != nil && (math.IsNaN(*suite.MinScore) || math.IsInf(*suite.MinScore, 0)) {
			return fmt.Errorf("pack: eval suite %q min_score must be finite", name)
		}
		if err := validateOptionalUniqueStrings(suite.WorkflowIDs, fmt.Sprintf("eval suite %q workflow_id", name)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(suite.RequiredArtifacts, fmt.Sprintf("eval suite %q required artifact", name)); err != nil {
			return err
		}
		if err := validateOptionalUniqueStrings(suite.RequiredState, fmt.Sprintf("eval suite %q required state", name)); err != nil {
			return err
		}
		for _, workflowID := range suite.WorkflowIDs {
			if !workflowIDs[strings.TrimSpace(workflowID)] {
				return fmt.Errorf("pack: eval suite %q references unknown workflow %q", name, strings.TrimSpace(workflowID))
			}
		}
		for _, artifactType := range suite.RequiredArtifacts {
			artifactType = strings.TrimSpace(artifactType)
			if artifactType == "" {
				continue
			}
			if !artifactTypes[artifactType] {
				return fmt.Errorf("pack: eval suite %q references undeclared artifact type %q", name, artifactType)
			}
		}
		seen[name] = true
	}
	if defaultCount > 1 {
		return fmt.Errorf("pack: only one eval suite can be default")
	}
	for _, name := range manifest.EvalSuites {
		if !seen[strings.TrimSpace(name)] {
			return fmt.Errorf("pack: manifest eval suite %q is not defined", strings.TrimSpace(name))
		}
	}
	return nil
}

func validatePackWorkflowEvalSuiteCoverage(def Definition) error {
	semanticTools := def.semanticToolIndex()
	for _, spec := range def.Workflows {
		workflowID := strings.TrimSpace(spec.ID)
		if workflowID == "" {
			continue
		}
		suites := def.EvalSuitesForWorkflow(workflowID)
		if len(suites) == 0 {
			continue
		}
		produced := workflowProducedArtifactTypesForDefinition(spec, semanticTools)
		if err := validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec, suites, produced); err != nil {
			return err
		}
	}
	return nil
}

func workflowProducedArtifactTypesForDefinition(spec agentxworkflow.Spec, semanticTools map[string]SemanticTool) map[string]bool {
	produced := workflowProducedArtifactTypes(spec)
	if len(semanticTools) == 0 {
		return produced
	}
	for _, node := range spec.Nodes {
		toolName := workflowToolName(node.Config)
		if toolName == "" {
			continue
		}
		semantic, ok := semanticTools[toolName]
		if !ok {
			continue
		}
		for _, artifactType := range NormalizeArtifactTypes(semantic.ArtifactTypes) {
			produced[artifactType] = true
		}
	}
	return produced
}

func validatePolicyProfiles(items []PolicyProfile, manifest Manifest) error {
	seen := map[string]bool{}
	defaultCount := 0
	for _, profile := range items {
		if err := validateRequiredPackField(profile.Name, "policy profile name"); err != nil {
			return err
		}
		name := profile.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate policy profile %q", name)
		}
		if err := validateRequiredPackField(profile.Contract.ID, fmt.Sprintf("policy profile %q contract id", name)); err != nil {
			return err
		}
		if profile.Default {
			defaultCount++
		}
		seen[name] = true
	}
	if defaultCount > 1 {
		return fmt.Errorf("pack: only one policy profile can be default")
	}
	for _, name := range manifest.PolicyProfiles {
		if !seen[strings.TrimSpace(name)] {
			return fmt.Errorf("pack: manifest policy profile %q is not defined", strings.TrimSpace(name))
		}
	}
	return nil
}

func validateMemorySchemas(items []MemorySchema) error {
	seen := map[string]bool{}
	defaultCount := 0
	for _, schema := range items {
		if err := validateRequiredPackField(schema.Name, "memory schema name"); err != nil {
			return err
		}
		name := schema.Name
		if seen[name] {
			return fmt.Errorf("pack: duplicate memory schema %q", name)
		}
		if err := validateOptionalPackField(schema.Description, fmt.Sprintf("memory schema %q description", name)); err != nil {
			return err
		}
		if err := validateOptionalPackSchema(schema.Schema, fmt.Sprintf("memory schema %q schema", name)); err != nil {
			return err
		}
		if schema.Default {
			defaultCount++
		}
		seen[name] = true
	}
	if defaultCount > 1 {
		return fmt.Errorf("pack: only one memory schema can be default")
	}
	return nil
}

func validateMemoryRecallPolicy(policy *MemoryRecallPolicy) error {
	if policy == nil {
		return nil
	}
	if policy.Limit < 0 {
		return fmt.Errorf("pack: memory recall policy limit must be >= 0")
	}
	if policy.MaxChars < 0 {
		return fmt.Errorf("pack: memory recall policy max_chars must be >= 0")
	}
	return validateOptionalUniqueStrings(policy.QueryHints, "memory recall query hint")
}

func sortDefinitions(items []Definition) {
	sort.Slice(items, func(i, j int) bool {
		return strings.TrimSpace(items[i].Manifest.ID) < strings.TrimSpace(items[j].Manifest.ID)
	})
}

func validateOptionalPackSchema(schema map[string]any, label string) error {
	if len(schema) == 0 {
		return nil
	}
	return validatePackSchemaDefinition(schema, label)
}

func validatePackSchemaDefinition(schema map[string]any, path string) error {
	types, err := readPackSchemaTypes(schema["type"])
	if err != nil {
		return fmt.Errorf("pack: %s.type: %w", path, err)
	}
	for _, item := range types {
		if !isSupportedPackSchemaType(item) {
			return fmt.Errorf("pack: %s.type: unsupported type %q", path, item)
		}
	}
	if err := validatePackSchemaKeywordTypeApplicability(schema, types, path); err != nil {
		return err
	}
	if rawConst, exists := schema["const"]; exists {
		if err := validatePackSchemaConstCompatibility(rawConst, schema, types); err != nil {
			return fmt.Errorf("pack: %s.const: %w", path, err)
		}
	}
	if rawEnum, exists := schema["enum"]; exists {
		if err := validatePackSchemaEnum(rawEnum, schema, types); err != nil {
			return fmt.Errorf("pack: %s.enum: %w", path, err)
		}
	}
	if required, exists := schema["required"]; exists {
		if _, err := readPackSchemaRequired(required); err != nil {
			return fmt.Errorf("pack: %s.required: %w", path, err)
		}
	}
	if rawProps, exists := schema["properties"]; exists {
		props, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("pack: %s.properties: must be an object", path)
		}
		for key, rawChild := range props {
			child, ok := rawChild.(map[string]any)
			if !ok {
				return fmt.Errorf("pack: %s.properties.%s: must be an object", path, key)
			}
			if err := validatePackSchemaDefinition(child, path+".properties."+key); err != nil {
				return err
			}
		}
	}
	if rawItems, exists := schema["items"]; exists {
		child, ok := rawItems.(map[string]any)
		if !ok {
			return fmt.Errorf("pack: %s.items: must be an object", path)
		}
		if err := validatePackSchemaDefinition(child, path+".items"); err != nil {
			return err
		}
	}
	if rawAdditional, exists := schema["additionalProperties"]; exists {
		switch value := rawAdditional.(type) {
		case bool:
			_ = value
		case map[string]any:
			if err := validatePackSchemaDefinition(value, path+".additionalProperties"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("pack: %s.additionalProperties: must be boolean or object", path)
		}
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("pack: %s.pattern: must be a string", path)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("pack: %s.pattern: %w", path, err)
		}
	}
	for _, keyword := range []string{"minProperties", "maxProperties", "minItems", "maxItems", "minLength", "maxLength"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validatePackSchemaNonNegativeIntegerKeyword(rawValue); err != nil {
				return fmt.Errorf("pack: %s.%s: %w", path, keyword, err)
			}
		}
	}
	for _, keyword := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validatePackSchemaFiniteNumberKeyword(rawValue); err != nil {
				return fmt.Errorf("pack: %s.%s: %w", path, keyword, err)
			}
		}
	}
	if err := validatePackSchemaKeywordRanges(schema, path); err != nil {
		return err
	}
	return nil
}

func validatePackSchemaKeywordTypeApplicability(schema map[string]any, types []string, path string) error {
	if len(types) == 0 {
		return nil
	}
	if err := validatePackSchemaKeywordsRequireType(schema, types, path, []string{
		"properties", "required", "additionalProperties", "minProperties", "maxProperties",
	}, "object"); err != nil {
		return err
	}
	if err := validatePackSchemaKeywordsRequireType(schema, types, path, []string{
		"items", "minItems", "maxItems",
	}, "array"); err != nil {
		return err
	}
	if err := validatePackSchemaKeywordsRequireType(schema, types, path, []string{
		"pattern", "minLength", "maxLength",
	}, "string"); err != nil {
		return err
	}
	if err := validatePackSchemaKeywordsRequireAnyType(schema, types, path, []string{
		"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum",
	}, []string{"number", "integer"}); err != nil {
		return err
	}
	return nil
}

func validatePackSchemaKeywordsRequireType(schema map[string]any, types []string, path string, keywords []string, requiredType string) error {
	return validatePackSchemaKeywordsRequireAnyType(schema, types, path, keywords, []string{requiredType})
}

func validatePackSchemaKeywordsRequireAnyType(schema map[string]any, types []string, path string, keywords []string, requiredTypes []string) error {
	for _, keyword := range keywords {
		if _, exists := schema[keyword]; !exists {
			continue
		}
		if packSchemaTypesIncludeAny(types, requiredTypes...) {
			continue
		}
		return fmt.Errorf("pack: %s.%s requires declared type to include %s", path, keyword, strings.Join(requiredTypes, " or "))
	}
	return nil
}

func packSchemaTypesIncludeAny(types []string, candidates ...string) bool {
	for _, item := range types {
		for _, candidate := range candidates {
			if item == candidate {
				return true
			}
		}
	}
	return false
}

func readPackSchemaTypes(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case string:
		value, err := normalizePackSchemaTypeEntry(typed)
		if err != nil {
			return nil, err
		}
		return []string{value}, nil
	case []string:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			value, err := normalizePackSchemaTypeEntry(item)
			if err != nil {
				return nil, err
			}
			if seen[value] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[value] = true
			out = append(out, value)
		}
		return out, nil
	case []any:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must contain strings")
			}
			value, err := normalizePackSchemaTypeEntry(text)
			if err != nil {
				return nil, err
			}
			if seen[value] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[value] = true
			out = append(out, value)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be string or array")
	}
}

func readPackSchemaRequired(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	var out []string
	switch typed := raw.(type) {
	case []string:
		seen := map[string]bool{}
		out = make([]string, 0, len(typed))
		for _, item := range typed {
			name, err := normalizePackSchemaRequiredEntry(item)
			if err != nil {
				return nil, err
			}
			if seen[name] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[name] = true
			out = append(out, name)
		}
	case []any:
		seen := map[string]bool{}
		out = make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must contain strings")
			}
			name, err := normalizePackSchemaRequiredEntry(text)
			if err != nil {
				return nil, err
			}
			if seen[name] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[name] = true
			out = append(out, name)
		}
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
	return out, nil
}

func normalizePackSchemaTypeEntry(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must not contain empty entries")
	}
	if raw != trimmed {
		return "", fmt.Errorf("must not include surrounding whitespace")
	}
	lower := strings.ToLower(trimmed)
	if trimmed != lower {
		return "", fmt.Errorf("must use canonical lowercase")
	}
	return lower, nil
}

func normalizePackSchemaRequiredEntry(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must not contain empty entries")
	}
	if raw != trimmed {
		return "", fmt.Errorf("must not include surrounding whitespace")
	}
	return trimmed, nil
}

func validatePackSchemaEnum(raw any, schema map[string]any, types []string) error {
	items, ok := schemaSlice(raw)
	if !ok || len(items) == 0 {
		return fmt.Errorf("must be a non-empty array")
	}
	for idx, item := range items {
		for prev := 0; prev < idx; prev++ {
			if schemaValuesEqual(item, items[prev]) {
				return fmt.Errorf("[%d]: must not contain duplicate entries", idx)
			}
		}
		if err := validatePackSchemaLiteralMatchesTypes(item, types); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
		if err := validatePackSchemaLiteralAgainstDefinition(item, schema); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
	}
	return nil
}

func validatePackSchemaConstCompatibility(raw any, schema map[string]any, types []string) error {
	if err := validatePackSchemaLiteralMatchesTypes(raw, types); err != nil {
		return err
	}
	return validatePackSchemaLiteralAgainstDefinition(raw, schema)
}

func validatePackSchemaLiteralMatchesTypes(raw any, types []string) error {
	if len(types) == 0 {
		return nil
	}
	for _, schemaType := range types {
		if packSchemaValueMatchesType(raw, schemaType) {
			return nil
		}
	}
	return fmt.Errorf("does not match declared type constraint")
}

func validatePackSchemaLiteralAgainstDefinition(raw any, schema map[string]any) error {
	if err := validatePackSchemaLiteralMatchesTypes(raw, readPackSchemaTypesOrNil(schema["type"])); err != nil {
		return err
	}
	switch value := raw.(type) {
	case string:
		return validatePackSchemaLiteralStringConstraints(value, schema)
	case int:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case int8:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case int16:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case int32:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case int64:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case uint:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case uint8:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case uint16:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case uint32:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case uint64:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case float32:
		return validatePackSchemaLiteralNumberConstraints(float64(value), schema)
	case float64:
		return validatePackSchemaLiteralNumberConstraints(value, schema)
	}
	if object, ok := packSchemaLiteralMap(raw); ok {
		return validatePackSchemaLiteralObjectConstraints(object, schema)
	}
	if items, ok := packSchemaLiteralSlice(raw); ok {
		return validatePackSchemaLiteralArrayConstraints(items, schema)
	}
	return nil
}

func readPackSchemaTypesOrNil(raw any) []string {
	types, err := readPackSchemaTypes(raw)
	if err != nil {
		return nil
	}
	return types
}

func validatePackSchemaLiteralObjectConstraints(object map[string]any, schema map[string]any) error {
	if minProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err == nil && ok && float64(len(object)) < minProps {
		return fmt.Errorf("violates minProperties")
	} else if err != nil {
		return err
	}
	if maxProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err == nil && ok && float64(len(object)) > maxProps {
		return fmt.Errorf("violates maxProperties")
	} else if err != nil {
		return err
	}
	required, err := readPackSchemaRequired(schema["required"])
	if err != nil {
		return err
	}
	for _, key := range required {
		if _, exists := object[key]; !exists {
			return fmt.Errorf("violates required")
		}
	}
	properties := map[string]any{}
	if rawProps, exists := schema["properties"]; exists {
		typed, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("properties: must be an object")
		}
		properties = typed
	}
	additional := schema["additionalProperties"]
	for key, item := range object {
		if childSchemaRaw, exists := properties[key]; exists {
			childSchema, ok := childSchemaRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("property schema must be object")
			}
			if err := validatePackSchemaLiteralAgainstDefinition(item, childSchema); err != nil {
				return err
			}
			continue
		}
		switch typed := additional.(type) {
		case nil:
			continue
		case bool:
			if !typed {
				return fmt.Errorf("violates additionalProperties")
			}
		case map[string]any:
			if err := validatePackSchemaLiteralAgainstDefinition(item, typed); err != nil {
				return err
			}
		default:
			return fmt.Errorf("additionalProperties: must be boolean or object")
		}
	}
	return nil
}

func validatePackSchemaLiteralArrayConstraints(items []any, schema map[string]any) error {
	if minItems, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minItems"]); err == nil && ok && float64(len(items)) < minItems {
		return fmt.Errorf("violates minItems")
	} else if err != nil {
		return err
	}
	if maxItems, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err == nil && ok && float64(len(items)) > maxItems {
		return fmt.Errorf("violates maxItems")
	} else if err != nil {
		return err
	}
	rawItems, exists := schema["items"]
	if !exists {
		return nil
	}
	itemSchema, ok := rawItems.(map[string]any)
	if !ok {
		return fmt.Errorf("items: must be an object")
	}
	for _, item := range items {
		if err := validatePackSchemaLiteralAgainstDefinition(item, itemSchema); err != nil {
			return err
		}
	}
	return nil
}

func validatePackSchemaLiteralStringConstraints(value string, schema map[string]any) error {
	if minLength, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minLength"]); err == nil && ok && float64(len(value)) < minLength {
		return fmt.Errorf("violates minLength")
	} else if err != nil {
		return err
	}
	if maxLength, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err == nil && ok && float64(len(value)) > maxLength {
		return fmt.Errorf("violates maxLength")
	} else if err != nil {
		return err
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("pattern: must be a string")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		if !re.MatchString(value) {
			return fmt.Errorf("violates pattern")
		}
	}
	return nil
}

func validatePackSchemaLiteralNumberConstraints(value float64, schema map[string]any) error {
	if minimum, ok, err := readPackSchemaFiniteNumberKeyword(schema["minimum"]); err == nil && ok && value < minimum {
		return fmt.Errorf("violates minimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readPackSchemaFiniteNumberKeyword(schema["maximum"]); err == nil && ok && value > maximum {
		return fmt.Errorf("violates maximum")
	} else if err != nil {
		return err
	}
	if minimum, ok, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err == nil && ok && value <= minimum {
		return fmt.Errorf("violates exclusiveMinimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err == nil && ok && value >= maximum {
		return fmt.Errorf("violates exclusiveMaximum")
	} else if err != nil {
		return err
	}
	return nil
}

func packSchemaLiteralMap(raw any) (map[string]any, bool) {
	if raw == nil {
		return nil, false
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		out[iter.Key().String()] = iter.Value().Interface()
	}
	return out, true
}

func packSchemaLiteralSlice(raw any) ([]any, bool) {
	if raw == nil {
		return nil, false
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	out := make([]any, rv.Len())
	for idx := 0; idx < rv.Len(); idx++ {
		out[idx] = rv.Index(idx).Interface()
	}
	return out, true
}

func packSchemaValueMatchesType(value any, schemaType string) bool {
	switch schemaType {
	case "object":
		if value == nil {
			return false
		}
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String
	case "array":
		if value == nil {
			return false
		}
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		return packSchemaValueAsFiniteNumber(value, false)
	case "integer":
		return packSchemaValueAsFiniteNumber(value, true)
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

func packSchemaValueAsFiniteNumber(value any, integerOnly bool) bool {
	switch typed := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		num := float64(typed)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return false
		}
		return !integerOnly || math.Trunc(num) == num
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return false
		}
		return !integerOnly || math.Trunc(typed) == typed
	default:
		return false
	}
}

func validatePackSchemaNonNegativeIntegerKeyword(raw any) error {
	switch value := raw.(type) {
	case int:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int8:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int16:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int32:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int64:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return fmt.Errorf("must be a finite integer")
		}
		if math.Trunc(num) != num {
			return fmt.Errorf("must be an integer")
		}
		if num < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("must be a finite integer")
		}
		if math.Trunc(value) != value {
			return fmt.Errorf("must be an integer")
		}
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	default:
		return fmt.Errorf("must be a non-negative integer")
	}
}

func validatePackSchemaFiniteNumberKeyword(raw any) error {
	switch value := raw.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return fmt.Errorf("must be a finite number")
		}
		return nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("must be a finite number")
		}
		return nil
	default:
		return fmt.Errorf("must be a number")
	}
}

func validatePackSchemaKeywordRanges(schema map[string]any, path string) error {
	if min, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err != nil {
		return fmt.Errorf("pack: %s.minProperties: %w", path, err)
	} else if max, okMax, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err != nil {
		return fmt.Errorf("pack: %s.maxProperties: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("pack: %s.minProperties: must be <= maxProperties", path)
	}
	if min, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minItems"]); err != nil {
		return fmt.Errorf("pack: %s.minItems: %w", path, err)
	} else if max, okMax, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err != nil {
		return fmt.Errorf("pack: %s.maxItems: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("pack: %s.minItems: must be <= maxItems", path)
	}
	if min, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minLength"]); err != nil {
		return fmt.Errorf("pack: %s.minLength: %w", path, err)
	} else if max, okMax, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err != nil {
		return fmt.Errorf("pack: %s.maxLength: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("pack: %s.minLength: must be <= maxLength", path)
	}
	if min, ok, err := readPackSchemaFiniteNumberKeyword(schema["minimum"]); err != nil {
		return fmt.Errorf("pack: %s.minimum: %w", path, err)
	} else if max, okMax, err := readPackSchemaFiniteNumberKeyword(schema["maximum"]); err != nil {
		return fmt.Errorf("pack: %s.maximum: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("pack: %s.minimum: must be <= maximum", path)
	}
	if min, ok, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err != nil {
		return fmt.Errorf("pack: %s.exclusiveMinimum: %w", path, err)
	} else if max, okMax, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err != nil {
		return fmt.Errorf("pack: %s.exclusiveMaximum: %w", path, err)
	} else if ok && okMax && min >= max {
		return fmt.Errorf("pack: %s.exclusiveMinimum: must be < exclusiveMaximum", path)
	}
	return nil
}

func readPackSchemaNonNegativeIntegerKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validatePackSchemaNonNegativeIntegerKeyword(raw); err != nil {
		return 0, false, err
	}
	switch value := raw.(type) {
	case int:
		return float64(value), true, nil
	case int8:
		return float64(value), true, nil
	case int16:
		return float64(value), true, nil
	case int32:
		return float64(value), true, nil
	case int64:
		return float64(value), true, nil
	case uint:
		return float64(value), true, nil
	case uint8:
		return float64(value), true, nil
	case uint16:
		return float64(value), true, nil
	case uint32:
		return float64(value), true, nil
	case uint64:
		return float64(value), true, nil
	case float32:
		return float64(value), true, nil
	case float64:
		return value, true, nil
	default:
		return 0, false, nil
	}
}

func readPackSchemaFiniteNumberKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validatePackSchemaFiniteNumberKeyword(raw); err != nil {
		return 0, false, err
	}
	switch value := raw.(type) {
	case int:
		return float64(value), true, nil
	case int8:
		return float64(value), true, nil
	case int16:
		return float64(value), true, nil
	case int32:
		return float64(value), true, nil
	case int64:
		return float64(value), true, nil
	case uint:
		return float64(value), true, nil
	case uint8:
		return float64(value), true, nil
	case uint16:
		return float64(value), true, nil
	case uint32:
		return float64(value), true, nil
	case uint64:
		return float64(value), true, nil
	case float32:
		return float64(value), true, nil
	case float64:
		return value, true, nil
	default:
		return 0, false, nil
	}
}

func isSupportedPackSchemaType(schemaType string) bool {
	switch schemaType {
	case "object", "array", "string", "number", "integer", "boolean", "null":
		return true
	default:
		return false
	}
}
