package pack

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type Binding struct {
	PackID        string              `json:"pack_id,omitempty"`
	CaseType      string              `json:"case_type,omitempty"`
	WorkflowID    string              `json:"workflow_id,omitempty"`
	Definition    Definition          `json:"definition"`
	Workflow      agentxworkflow.Spec `json:"workflow"`
	CaseSchema    *CaseSchema         `json:"case_schema,omitempty"`
	PolicyProfile *PolicyProfile      `json:"policy_profile,omitempty"`
	EvalSuite     *EvalSuite          `json:"eval_suite,omitempty"`
	EvalSuites    []EvalSuite         `json:"eval_suites,omitempty"`
	MemorySchema  *MemorySchema       `json:"memory_schema,omitempty"`
	MemoryRecall  *MemoryRecallPolicy `json:"memory_recall,omitempty"`
}

// ResolveBinding从 Registry解析一个完整、可执行前检查的 Pack Binding。
func (c *Coordinator) ResolveBinding(reg Registry, packID string, caseType string, workflowID string) (Binding, bool, error) {
	if _, err := c.workflowValidator(); err != nil {
		return Binding{}, false, err
	}
	if reg == nil {
		return Binding{}, false, nil
	}
	def, ok := reg.Get(packID)
	if !ok {
		return Binding{}, false, nil
	}
	binding, err := c.resolveBindingFromDefinition(def, caseType, workflowID)
	if err != nil {
		return Binding{}, false, err
	}
	return binding, true, nil
}

func (c *Coordinator) resolveBindingFromDefinition(def Definition, caseType string, workflowID string) (Binding, error) {
	if caseType == "" {
		return Binding{}, fmt.Errorf("pack: case_type is required")
	}
	if !def.Manifest.SupportsCaseType(caseType) {
		return Binding{}, fmt.Errorf("pack: definition %q does not support case type %q", def.Manifest.ID, caseType)
	}
	selected, err := def.ResolveWorkflowForCaseType(caseType, workflowID)
	if err != nil {
		return Binding{}, err
	}
	spec, err := c.MaterializeWorkflow(def, selected.ID)
	if err != nil {
		return Binding{}, err
	}
	binding := Binding{
		PackID:     def.Manifest.ID,
		CaseType:   caseType,
		WorkflowID: spec.ID,
		Definition: cloneDefinition(def),
		Workflow:   spec,
	}
	if schema, ok := def.CaseSchemaByType(caseType); ok {
		schemaCopy := schema
		binding.CaseSchema = &schemaCopy
	}
	profileName := spec.DefaultContract
	switch {
	case profileName != "":
		if profile, ok := def.PolicyProfileByName(profileName); ok {
			profileCopy := profile
			binding.PolicyProfile = &profileCopy
		} else {
			return Binding{}, fmt.Errorf("pack: workflow %q references unknown policy profile %q", spec.ID, profileName)
		}
	default:
		if profile, ok := def.DefaultPolicyProfile(); ok {
			profileCopy := profile
			binding.PolicyProfile = &profileCopy
		}
	}
	if schema, ok := def.DefaultMemorySchema(); ok {
		schemaCopy := schema
		binding.MemorySchema = &schemaCopy
	}
	if suites := def.EvalSuitesForWorkflow(spec.ID); len(suites) > 0 {
		binding.EvalSuites = cloneEvalSuites(suites)
		suiteCopy := binding.EvalSuites[0]
		binding.EvalSuite = &suiteCopy
	}
	if def.MemoryRecallPolicy != nil {
		binding.MemoryRecall = cloneMemoryRecallPolicy(def.MemoryRecallPolicy)
	}
	runtimeWorkflow, err := binding.prepareRuntimeWorkflow()
	if err != nil {
		return Binding{}, err
	}
	binding.Workflow = runtimeWorkflow
	return binding, nil
}

func (b Binding) ValidateCaseInput(input map[string]any) error {
	if b.CaseSchema == nil || len(b.CaseSchema.Schema) == 0 {
		return nil
	}
	if input == nil {
		input = map[string]any{}
	}
	if err := validatePackSchemaDefinition(b.CaseSchema.Schema, "case_input"); err != nil {
		return err
	}
	return validateSchemaValue(b.CaseSchema.Schema, input, "case_input")
}

func (b Binding) prepareRuntimeWorkflow() (agentxworkflow.Spec, error) {
	spec := cloneWorkflowSpec(b.Workflow)
	declaredEvaluators := map[string]bool{}
	defaultEvaluator := ""
	declaredCount := 0
	for _, ref := range spec.EvaluatorSchema {
		name := ref.Name
		if name == "" {
			continue
		}
		if _, ok := b.Definition.EvaluatorByName(name); !ok {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: workflow %q references unknown evaluator %q", spec.ID, name)
		}
		declaredEvaluators[name] = true
		declaredCount++
		if declaredCount == 1 {
			defaultEvaluator = name
		} else {
			defaultEvaluator = ""
		}
	}
	if defaultEvaluator == "" && declaredCount == 0 {
		if evaluator, ok := b.Definition.DefaultEvaluator(); ok {
			defaultEvaluator = evaluator.Name
		}
	}
	for idx := range spec.Nodes {
		node := spec.Nodes[idx]
		if ref := node.ContractRef; ref != "" {
			if _, ok := b.Definition.PolicyProfileByName(ref); !ok {
				return agentxworkflow.Spec{}, fmt.Errorf("pack: workflow %q node %q references unknown contract_ref %q",
					spec.ID, node.ID, ref)
			}
		}
		if node.Kind != agentxworkflow.NodeEvaluate {
			continue
		}
		config := cloneStringAnyMap(node.Config)
		if config == nil {
			config = map[string]any{}
		}
		evaluatorName := readBindingConfigString(config, "evaluator")
		if evaluatorName == "" && defaultEvaluator != "" {
			evaluatorName = defaultEvaluator
			config["evaluator"] = evaluatorName
		}
		if evaluatorName == "" {
			spec.Nodes[idx].Config = config
			continue
		}
		if len(declaredEvaluators) > 0 && !declaredEvaluators[evaluatorName] {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: workflow %q evaluate node %q references evaluator %q outside evaluator_schema",
				spec.ID, node.ID, evaluatorName)
		}
		evaluator, ok := b.Definition.EvaluatorByName(evaluatorName)
		if !ok {
			return agentxworkflow.Spec{}, fmt.Errorf("pack: workflow %q evaluate node %q references unknown evaluator %q",
				spec.ID, node.ID, evaluatorName)
		}
		if readBindingConfigString(config, "instruction", "task", "prompt", "goal", "request", "query") == "" {
			config["instruction"] = buildPackEvaluatorInstruction(evaluator)
		}
		if !bindingConfigHasSchema(config) && len(evaluator.OutputSchema) > 0 {
			config["schema"] = cloneStringAnyMap(evaluator.OutputSchema)
		}
		spec.Nodes[idx].Config = config
	}
	if err := validateWorkflowArtifactCoverage(spec); err != nil {
		return agentxworkflow.Spec{}, err
	}
	if err := validateWorkflowEvalSuiteCoverage(spec, b.EvalSuites); err != nil {
		return agentxworkflow.Spec{}, err
	}
	return spec, nil
}

func (b Binding) ContractByRef(name string) (agentxexecution.Contract, bool) {
	if name == "" {
		return agentxexecution.Contract{}, false
	}
	profile, ok := b.Definition.PolicyProfileByName(name)
	if !ok {
		return agentxexecution.Contract{}, false
	}
	return profile.Contract, true
}

func (b Binding) ResolveMemorySchema(name string) (MemorySchema, bool, error) {
	if name != "" {
		schema, ok := b.Definition.MemorySchemaByName(name)
		if !ok {
			return MemorySchema{}, false, fmt.Errorf("pack: binding %q references unknown memory schema %q", b.PackID, name)
		}
		return schema, true, nil
	}
	if b.MemorySchema != nil {
		return *b.MemorySchema, true, nil
	}
	schema, ok := b.Definition.DefaultMemorySchema()
	return schema, ok, nil
}

func (b Binding) ValidateMemoryRecord(schemaName string, value map[string]any) error {
	schema, ok, err := b.ResolveMemorySchema(schemaName)
	if err != nil {
		return err
	}
	if !ok || len(schema.Schema) == 0 {
		return nil
	}
	if value == nil {
		value = map[string]any{}
	}
	if err := validatePackSchemaDefinition(schema.Schema, "memory_record"); err != nil {
		return err
	}
	return validateSchemaValue(schema.Schema, value, "memory_record")
}

func buildPackEvaluatorInstruction(evaluator Evaluator) string {
	description := evaluator.Description
	if description == "" {
		description = "Evaluate the provided input and return a structured decision."
	}
	return description
}

func validateWorkflowEvalSuiteCoverage(spec agentxworkflow.Spec, suites []EvalSuite) error {
	return validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec, suites, workflowProducedArtifactTypes(spec))
}

func validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec agentxworkflow.Spec, suites []EvalSuite, produced map[string]bool) error {
	if len(suites) == 0 {
		return nil
	}
	stateSlots := workflowStateSlotSet(spec)
	for _, suite := range suites {
		for _, path := range suite.RequiredState {
			if path == "" {
				continue
			}
			if !stateSlots[path] {
				return fmt.Errorf("pack: workflow %q eval suite %q requires undeclared state path %q", spec.ID, suite.Name, path)
			}
		}
		for _, path := range []string{suite.PassPath, suite.ScorePath, suite.SummaryPath} {
			if path == "" {
				continue
			}
			if !stateSlots[path] {
				return fmt.Errorf("pack: workflow %q eval suite %q references undeclared state path %q", spec.ID, suite.Name, path)
			}
		}
		for _, artifactType := range suite.RequiredArtifacts {
			if produced[artifactType] {
				continue
			}
			return fmt.Errorf("pack: workflow %q eval suite %q requires unproduced artifact type %q", spec.ID, suite.Name, artifactType)
		}
	}
	return nil
}

func validateWorkflowArtifactCoverage(spec agentxworkflow.Spec) error {
	if len(spec.ArtifactSchema) == 0 {
		return nil
	}
	produced := workflowProducedArtifactTypes(spec)
	missing := make([]string, 0)
	for _, ref := range spec.ArtifactSchema {
		artifactType := ref.Type
		if artifactType == "" || produced[artifactType] {
			continue
		}
		missing = append(missing, artifactType)
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("pack: workflow %q artifact_schema requires unproduced artifact types %q", spec.ID, strings.Join(missing, ", "))
}

func workflowProducedArtifactTypes(spec agentxworkflow.Spec) map[string]bool {
	produced := map[string]bool{}
	for _, node := range spec.Nodes {
		for _, artifactType := range ArtifactTypesFromConfig(node.Config) {
			produced[artifactType] = true
		}
	}
	return produced
}

func workflowStateSlotSet(spec agentxworkflow.Spec) map[string]bool {
	out := map[string]bool{}
	for _, slot := range spec.StateSchema {
		name := slot.Name
		if name != "" {
			out[name] = true
		}
	}
	return out
}

func bindingConfigHasSchema(config map[string]any) bool {
	if len(config) == 0 {
		return false
	}
	for _, key := range []string{"schema", "output_schema"} {
		raw, ok := config[key]
		if ok && raw != nil {
			return true
		}
	}
	return false
}

func readBindingConfigString(config map[string]any, keys ...string) string {
	if len(config) == 0 {
		return ""
	}
	for _, key := range keys {
		raw, ok := config[key]
		if !ok || raw == nil {
			continue
		}
		if value, ok := raw.(string); ok {
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func validateSchemaValue(schema map[string]any, value any, path string) error {
	if len(schema) == 0 {
		return nil
	}
	if err := validateSchemaConstAndEnumValue(schema, value, path); err != nil {
		return err
	}
	schemaTypes := runtimePackSchemaTypes(schema)
	if len(schemaTypes) == 0 {
		return nil
	}
	if objectValue, ok := value.(map[string]any); ok && runtimePackSchemaAllowsType(schemaTypes, "object") {
		if minProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err != nil {
			return fmt.Errorf("pack: %s minProperties: %w", path, err)
		} else if ok && float64(len(objectValue)) < minProps {
			return fmt.Errorf("pack: %s violates minProperties", path)
		}
		if maxProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err != nil {
			return fmt.Errorf("pack: %s maxProperties: %w", path, err)
		} else if ok && float64(len(objectValue)) > maxProps {
			return fmt.Errorf("pack: %s violates maxProperties", path)
		}
		required, err := readPackSchemaRequired(schema["required"])
		if err != nil {
			return fmt.Errorf("pack: %s.required: %w", path, err)
		}
		for _, name := range required {
			if _, exists := objectValue[name]; !exists {
				return fmt.Errorf("pack: %s.%s is required", path, name)
			}
		}
		properties := readSchemaMapMap(schema["properties"])
		rawAdditional, additionalExists := schema["additionalProperties"]
		for key, childValue := range objectValue {
			if childSchema, exists := properties[key]; exists {
				if err := validateSchemaValue(childSchema, childValue, path+"."+key); err != nil {
					return err
				}
				continue
			}
			if !additionalExists {
				continue
			}
			switch typed := rawAdditional.(type) {
			case bool:
				if !typed {
					return fmt.Errorf("pack: %s.%s is not allowed", path, key)
				}
			case map[string]any:
				if err := validateSchemaValue(typed, childValue, path+"."+key); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if sliceValue, ok := schemaSlice(value); ok && runtimePackSchemaAllowsType(schemaTypes, "array") {
		items := readSchemaMap(schema["items"])
		if minItems, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minItems"]); err != nil {
			return fmt.Errorf("pack: %s minItems: %w", path, err)
		} else if ok && float64(len(sliceValue)) < minItems {
			return fmt.Errorf("pack: %s violates minItems", path)
		}
		if maxItems, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err != nil {
			return fmt.Errorf("pack: %s maxItems: %w", path, err)
		} else if ok && float64(len(sliceValue)) > maxItems {
			return fmt.Errorf("pack: %s violates maxItems", path)
		}
		if len(items) == 0 {
			return nil
		}
		for idx, item := range sliceValue {
			if err := validateSchemaValue(items, item, fmt.Sprintf("%s[%d]", path, idx)); err != nil {
				return err
			}
		}
		return nil
	}
	if text, ok := value.(string); ok && runtimePackSchemaAllowsType(schemaTypes, "string") {
		if minLength, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minLength"]); err != nil {
			return fmt.Errorf("pack: %s minLength: %w", path, err)
		} else if ok && float64(len(text)) < minLength {
			return fmt.Errorf("pack: %s violates minLength", path)
		}
		if maxLength, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err != nil {
			return fmt.Errorf("pack: %s maxLength: %w", path, err)
		} else if ok && float64(len(text)) > maxLength {
			return fmt.Errorf("pack: %s violates maxLength", path)
		}
		if rawPattern, exists := schema["pattern"]; exists {
			pattern, ok := rawPattern.(string)
			if !ok {
				return fmt.Errorf("pack: %s pattern must be string", path)
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("pack: %s pattern: %w", path, err)
			}
			if !re.MatchString(text) {
				return fmt.Errorf("pack: %s violates pattern", path)
			}
		}
		return nil
	}
	if _, ok := value.(bool); ok && runtimePackSchemaAllowsType(schemaTypes, "boolean") {
		return nil
	}
	if isSchemaInteger(value) && runtimePackSchemaAllowsType(schemaTypes, "integer") {
		if err := validateSchemaNumericValue(schema, value, path); err != nil {
			return err
		}
		return nil
	}
	if isSchemaNumber(value) && runtimePackSchemaAllowsType(schemaTypes, "number") {
		if err := validateSchemaNumericValue(schema, value, path); err != nil {
			return err
		}
		return nil
	}
	if value == nil && runtimePackSchemaAllowsType(schemaTypes, "null") {
		return nil
	}
	return fmt.Errorf("pack: %s must match declared type", path)
}

func validateSchemaProvidedValue(schema map[string]any, value any, path string) error {
	if len(schema) == 0 {
		return nil
	}
	if err := validateSchemaConstAndEnumValue(schema, value, path); err != nil {
		return err
	}
	schemaTypes := runtimePackSchemaTypes(schema)
	if len(schemaTypes) == 0 {
		return nil
	}
	if objectValue, ok := value.(map[string]any); ok && runtimePackSchemaAllowsType(schemaTypes, "object") {
		if minProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err != nil {
			return fmt.Errorf("pack: %s minProperties: %w", path, err)
		} else if ok && float64(len(objectValue)) < minProps {
			return fmt.Errorf("pack: %s violates minProperties", path)
		}
		if maxProps, ok, err := readPackSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err != nil {
			return fmt.Errorf("pack: %s maxProperties: %w", path, err)
		} else if ok && float64(len(objectValue)) > maxProps {
			return fmt.Errorf("pack: %s violates maxProperties", path)
		}
		properties := readSchemaMapMap(schema["properties"])
		rawAdditional, additionalExists := schema["additionalProperties"]
		for key, childValue := range objectValue {
			if childSchema, exists := properties[key]; exists {
				if err := validateSchemaProvidedValue(childSchema, childValue, path+"."+key); err != nil {
					return err
				}
				continue
			}
			if !additionalExists {
				continue
			}
			switch typed := rawAdditional.(type) {
			case bool:
				if !typed {
					return fmt.Errorf("pack: %s.%s is not allowed", path, key)
				}
			case map[string]any:
				if err := validateSchemaProvidedValue(typed, childValue, path+"."+key); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return validateSchemaValue(schema, value, path)
}

func validateSchemaConstAndEnumValue(schema map[string]any, value any, path string) error {
	if rawConst, exists := schema["const"]; exists && !schemaValuesEqual(value, rawConst) {
		return fmt.Errorf("pack: %s must equal const", path)
	}
	if rawEnum, exists := schema["enum"]; exists {
		items, ok := schemaSlice(rawEnum)
		if ok && len(items) > 0 {
			for _, item := range items {
				if schemaValuesEqual(value, item) {
					return nil
				}
			}
			return fmt.Errorf("pack: %s must match enum", path)
		}
	}
	return nil
}

func validateSchemaNumericValue(schema map[string]any, value any, path string) error {
	num, ok := schemaValueAsFiniteFloat64(value)
	if !ok {
		return fmt.Errorf("pack: %s must be finite number", path)
	}
	if minimum, ok, err := readPackSchemaFiniteNumberKeyword(schema["minimum"]); err != nil {
		return fmt.Errorf("pack: %s minimum: %w", path, err)
	} else if ok && num < minimum {
		return fmt.Errorf("pack: %s violates minimum", path)
	}
	if maximum, ok, err := readPackSchemaFiniteNumberKeyword(schema["maximum"]); err != nil {
		return fmt.Errorf("pack: %s maximum: %w", path, err)
	} else if ok && num > maximum {
		return fmt.Errorf("pack: %s violates maximum", path)
	}
	if minimum, ok, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err != nil {
		return fmt.Errorf("pack: %s exclusiveMinimum: %w", path, err)
	} else if ok && num <= minimum {
		return fmt.Errorf("pack: %s violates exclusiveMinimum", path)
	}
	if maximum, ok, err := readPackSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err != nil {
		return fmt.Errorf("pack: %s exclusiveMaximum: %w", path, err)
	} else if ok && num >= maximum {
		return fmt.Errorf("pack: %s violates exclusiveMaximum", path)
	}
	return nil
}

func schemaValueAsFiniteFloat64(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case float32:
		num := float64(typed)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return 0, false
		}
		return num, true
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return 0, false
		}
		return typed, true
	default:
		return 0, false
	}
}

func schemaValuesEqual(left any, right any) bool {
	return reflect.DeepEqual(normalizeSchemaComparable(left), normalizeSchemaComparable(right))
}

func normalizeSchemaComparable(value any) any {
	if num, ok := schemaValueAsFiniteFloat64(value); ok {
		return num
	}
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return value
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out[iter.Key().String()] = normalizeSchemaComparable(iter.Value().Interface())
		}
		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for idx := 0; idx < rv.Len(); idx++ {
			out[idx] = normalizeSchemaComparable(rv.Index(idx).Interface())
		}
		return out
	default:
		return value
	}
}

func schemaHasObjectShape(schema map[string]any) bool {
	return len(readSchemaMapMap(schema["properties"])) > 0 || len(readPackSchemaRequiredOrNil(schema["required"])) > 0
}

func schemaHasArrayShape(schema map[string]any) bool {
	return len(readSchemaMap(schema["items"])) > 0
}

func runtimePackSchemaTypes(schema map[string]any) []string {
	if types := readPackSchemaTypesOrNil(schema["type"]); len(types) > 0 {
		return types
	}
	switch {
	case schemaHasObjectShape(schema):
		return []string{"object"}
	case schemaHasArrayShape(schema):
		return []string{"array"}
	default:
		return nil
	}
}

func runtimePackSchemaAllowsType(schemaTypes []string, candidate string) bool {
	for _, schemaType := range schemaTypes {
		if schemaType == candidate {
			return true
		}
	}
	return false
}

func readSchemaMap(value any) map[string]any {
	out, _ := value.(map[string]any)
	return out
}

func readPackSchemaRequiredOrNil(raw any) []string {
	required, err := readPackSchemaRequired(raw)
	if err != nil {
		return nil
	}
	return required
}

func readSchemaMapMap(value any) map[string]map[string]any {
	raw, _ := value.(map[string]any)
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]map[string]any, len(raw))
	for key, item := range raw {
		child, _ := item.(map[string]any)
		if len(child) == 0 {
			continue
		}
		out[key] = child
	}
	return out
}

func schemaSlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil, false
	}
	out := make([]any, 0, rv.Len())
	for idx := 0; idx < rv.Len(); idx++ {
		out = append(out, rv.Index(idx).Interface())
	}
	return out, true
}

func isSchemaInteger(value any) bool {
	switch typed := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return float64(int64(typed)) == float64(typed)
	case float64:
		return float64(int64(typed)) == typed
	default:
		return false
	}
}

func isSchemaNumber(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}
