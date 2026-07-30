// Package bindingstate 提供与执行 substrate 无关的 Workflow binding 和内存状态转换机制。
//
// 当前 package 处于 Experimental/private validation。它不拥有 lowering、
// executor、RunStore、durable lifecycle、retry、cancellation、provider 或
// Scene policy。
package bindingstate

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// Inputs contains the three portable value roots available to input bindings.
type Inputs struct {
	InitialState map[string]any
	SessionInput map[string]any
	CaseInput    map[string]any
}

// NodeResult is an opaque substrate-neutral node result.
//
// Construct values with NewNodeResult so structured output decoding follows
// the documented raw-JSON contract.
type NodeResult struct {
	status     string
	output     string
	errorText  string
	structured any
}

// Runtime owns one in-memory workflow binding/state execution slice.
type Runtime struct {
	state        map[string]any
	sessionInput map[string]any
	caseInput    map[string]any
	nodes        map[string]NodeResult
}

// New returns an isolated binding/state runtime. Input maps are deep-cloned so
// callers cannot mutate runtime state through construction values.
func New(inputs Inputs) *Runtime {
	return &Runtime{
		state:        cloneMapValue(inputs.InitialState),
		sessionInput: cloneMapValue(inputs.SessionInput),
		caseInput:    cloneMapValue(inputs.CaseInput),
		nodes:        map[string]NodeResult{},
	}
}

// NewNodeResult constructs a portable node result and decodes structured JSON
// output only when the raw output has no surrounding whitespace.
func NewNodeResult(status string, output string, errorText string) NodeResult {
	return NodeResult{
		status:     status,
		output:     output,
		errorText:  errorText,
		structured: decodeStructuredOutput(output),
	}
}

// State returns a deep-cloned snapshot of the current workflow state.
func (r *Runtime) State() map[string]any {
	if r == nil {
		return map[string]any{}
	}
	return cloneMapValue(r.state)
}

// MaterializeArguments applies input bindings to a JSON object argument value.
//
// Bindings are evaluated in input order. Optional bindings only suppress a
// missing-value error; malformed source/target paths remain errors.
func (r *Runtime) MaterializeArguments(nodeID string, argumentsJSON string, bindings []workflow.BindingSpec) (string, error) {
	if len(bindings) == 0 {
		return argumentsJSON, nil
	}
	args, ok := decodeArgumentsObject(argumentsJSON)
	if !ok {
		return argumentsJSON, fmt.Errorf("workflow: node %q arguments must be a JSON object for input bindings", nodeID)
	}
	for _, binding := range bindings {
		value, err := r.resolveInputBinding(binding.From)
		if err != nil {
			if binding.Optional && isMissingRuntimeBindingValue(err) {
				continue
			}
			return argumentsJSON, fmt.Errorf("workflow: node %q input binding %q -> %q: %w", nodeID, binding.From, binding.To, err)
		}
		targetPath, err := normalizeInputBindingTarget(binding.To)
		if err != nil {
			return argumentsJSON, fmt.Errorf("workflow: node %q input binding %q -> %q: %w", nodeID, binding.From, binding.To, err)
		}
		setObjectPath(args, targetPath, cloneValue(value))
	}
	raw, err := json.Marshal(args)
	if err != nil {
		return argumentsJSON, fmt.Errorf("workflow: node %q remarshal bound arguments: %w", nodeID, err)
	}
	return string(raw), nil
}

// ApplyNodeOutputs records the node result before applying output bindings.
//
// Bindings are applied in input order. A later error does not roll back the
// already recorded result or earlier state writes.
func (r *Runtime) ApplyNodeOutputs(nodeID string, bindings []workflow.BindingSpec, node NodeResult) error {
	if r.nodes == nil {
		r.nodes = map[string]NodeResult{}
	}
	if r.state == nil {
		r.state = map[string]any{}
	}
	r.nodes[nodeID] = cloneNodeResult(node)
	for _, binding := range bindings {
		value, err := resolveOutputBinding(node, binding.From)
		if err != nil {
			if binding.Optional && isMissingRuntimeBindingValue(err) {
				continue
			}
			return fmt.Errorf("workflow: node %q output binding %q -> %q: %w", nodeID, binding.From, binding.To, err)
		}
		targetPath, err := normalizeOutputBindingTarget(binding.To)
		if err != nil {
			return fmt.Errorf("workflow: node %q output binding %q -> %q: %w", nodeID, binding.From, binding.To, err)
		}
		setObjectPath(r.state, targetPath, cloneValue(value))
	}
	return nil
}

// ValidateRequiredSlots verifies that every required workflow state slot has
// been populated.
func (r *Runtime) ValidateRequiredSlots(slots []workflow.StateSlotSpec) error {
	state := map[string]any{}
	if r != nil {
		state = r.state
	}
	for _, slot := range slots {
		if err := ValidateSlotName(slot.Name); err != nil {
			return err
		}
		if !slot.Required {
			continue
		}
		if _, ok := lookupMapPath(state, splitBindingPath(slot.Name)); ok {
			continue
		}
		return fmt.Errorf("workflow: required state slot %q was not populated", slot.Name)
	}
	return nil
}

// ValidateSlotName validates the canonical rooted-at-state slot name grammar.
func ValidateSlotName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("workflow: state slot name is required")
	}
	if name != trimmed {
		return fmt.Errorf("workflow: state slot %q must not include surrounding whitespace", name)
	}
	if strings.HasPrefix(trimmed, "state.") {
		return fmt.Errorf("workflow: state slot %q must not include \"state.\" prefix; state_schema names are already rooted at workflow state", trimmed)
	}
	for _, segment := range strings.Split(trimmed, ".") {
		segmentTrimmed := strings.TrimSpace(segment)
		if segmentTrimmed == "" {
			return fmt.Errorf("workflow: state slot %q contains empty path segment", trimmed)
		}
		if segment != segmentTrimmed {
			return fmt.Errorf("workflow: state slot %q contains path segment with surrounding whitespace", trimmed)
		}
	}
	return nil
}

func isMissingRuntimeBindingValue(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(err.Error()), "value ") && strings.HasSuffix(strings.TrimSpace(err.Error()), " not found")
}

func (r *Runtime) resolveInputBinding(source string) (any, error) {
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	if err := validateRuntimeBindingPath(source, "input binding source"); err != nil {
		return nil, err
	}
	switch {
	case strings.HasPrefix(source, "state."):
		return lookupPathRequired(r.state, strings.TrimPrefix(source, "state."))
	case strings.HasPrefix(source, "session.input."):
		return lookupPathRequired(r.sessionInput, strings.TrimPrefix(source, "session.input."))
	case strings.HasPrefix(source, "case.input."):
		return lookupPathRequired(r.caseInput, strings.TrimPrefix(source, "case.input."))
	case strings.HasPrefix(source, "node."):
		return r.resolveNodeBinding(source)
	default:
		return nil, fmt.Errorf("unsupported input binding source %q", source)
	}
}

func (r *Runtime) resolveNodeBinding(source string) (any, error) {
	if err := validateRuntimeBindingPath(source, "node binding source"); err != nil {
		return nil, err
	}
	parts := splitBindingPath(source)
	if len(parts) < 3 || parts[0] != "node" {
		return nil, fmt.Errorf("unsupported node binding source %q", source)
	}
	nodeID := parts[1]
	if nodeID == "" {
		return nil, fmt.Errorf("node id is required in source %q", source)
	}
	node, ok := r.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %q has no recorded output", nodeID)
	}
	switch parts[2] {
	case "output":
		if len(parts) == 3 {
			return node.output, nil
		}
		if node.structured == nil {
			return nil, fmt.Errorf("node %q output is not structured", nodeID)
		}
		return lookupAnyPathRequired(node.structured, parts[3:])
	case "result":
		if len(parts) == 3 {
			if node.structured != nil {
				return cloneValue(node.structured), nil
			}
			return node.output, nil
		}
		if node.structured == nil {
			return nil, fmt.Errorf("node %q result is not structured", nodeID)
		}
		return lookupAnyPathRequired(node.structured, parts[3:])
	case "status":
		if len(parts) > 3 {
			return nil, fmt.Errorf("node %q status is scalar and cannot be dereferenced", nodeID)
		}
		return node.status, nil
	case "error":
		if len(parts) > 3 {
			return nil, fmt.Errorf("node %q error is scalar and cannot be dereferenced", nodeID)
		}
		return node.errorText, nil
	default:
		return nil, fmt.Errorf("unsupported node binding source %q", source)
	}
}

func resolveOutputBinding(node NodeResult, source string) (any, error) {
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	if err := validateRuntimeBindingPath(source, "output binding source"); err != nil {
		return nil, err
	}
	switch {
	case source == "output":
		return node.output, nil
	case strings.HasPrefix(source, "output."):
		if node.structured == nil {
			return nil, fmt.Errorf("output is not structured")
		}
		return lookupAnyPathRequired(node.structured, splitBindingPath(strings.TrimPrefix(source, "output.")))
	case source == "result":
		if node.structured != nil {
			return cloneValue(node.structured), nil
		}
		return node.output, nil
	case strings.HasPrefix(source, "result."):
		if node.structured == nil {
			return nil, fmt.Errorf("result is not structured")
		}
		return lookupAnyPathRequired(node.structured, splitBindingPath(strings.TrimPrefix(source, "result.")))
	case source == "error":
		return node.errorText, nil
	case source == "status":
		return node.status, nil
	case strings.HasPrefix(source, "error."):
		return nil, fmt.Errorf("error is scalar and cannot be dereferenced")
	case strings.HasPrefix(source, "status."):
		return nil, fmt.Errorf("status is scalar and cannot be dereferenced")
	default:
		return nil, fmt.Errorf("unsupported output binding source %q", source)
	}
}

func cloneNodeResult(node NodeResult) NodeResult {
	node.structured = cloneValue(node.structured)
	return node
}

func decodeArgumentsObject(raw string) (map[string]any, bool) {
	if raw == "" {
		return map[string]any{}, true
	}
	if raw != strings.TrimSpace(raw) {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return nil, false
	}
	return out, true
}

func decodeStructuredOutput(raw string) any {
	if raw == "" {
		return nil
	}
	if raw != strings.TrimSpace(raw) {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil
	}
	return value
}

func normalizeInputBindingTarget(path string) (string, error) {
	if err := validateRuntimeBindingPath(path, "input binding target"); err != nil {
		return "", err
	}
	if !strings.HasPrefix(path, "args.") {
		return "", fmt.Errorf("unsupported input binding target %q", path)
	}
	original := path
	path = strings.TrimPrefix(path, "args.")
	if path == "" {
		return "", fmt.Errorf("input binding target %q must bind a concrete tool argument path", original)
	}
	return path, nil
}

func normalizeOutputBindingTarget(path string) (string, error) {
	if err := validateRuntimeBindingPath(path, "output binding target"); err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "state.") {
		return strings.TrimPrefix(path, "state."), nil
	}
	return "", fmt.Errorf("unsupported output binding target %q", path)
}

func lookupPathRequired(root map[string]any, path string) (any, error) {
	value, ok := lookupMapPath(root, splitBindingPath(path))
	if !ok {
		return nil, fmt.Errorf("value %q not found", path)
	}
	return value, nil
}

func lookupAnyPathRequired(root any, parts []string) (any, error) {
	value, ok := lookupAnyPath(root, parts)
	if !ok {
		return nil, fmt.Errorf("value %q not found", strings.Join(parts, "."))
	}
	return value, nil
}

func lookupMapPath(root map[string]any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return root, true
	}
	return lookupAnyPath(root, parts)
}

func lookupAnyPath(root any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return root, true
	}
	current := root
	for _, part := range parts {
		switch obj := current.(type) {
		case map[string]any:
			next, ok := obj[part]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			idx, ok := bindingPathIndex(part)
			if !ok || idx >= len(obj) {
				return nil, false
			}
			current = obj[idx]
		default:
			return nil, false
		}
	}
	return current, true
}

func setObjectPath(root map[string]any, path string, value any) {
	parts := splitBindingPath(path)
	if len(parts) == 0 {
		return
	}
	setAnyObjectPath(root, parts, value)
}

func setAnyObjectPath(current any, parts []string, value any) any {
	if len(parts) == 0 {
		return value
	}
	part := parts[0]
	if idx, ok := bindingPathIndex(part); ok {
		items, _ := current.([]any)
		if len(items) <= idx {
			nextItems := make([]any, idx+1)
			copy(nextItems, items)
			items = nextItems
		}
		if len(parts) == 1 {
			items[idx] = value
			return items
		}
		next := items[idx]
		if !bindingContainerMatches(next, parts[1]) {
			next = newBindingContainer(parts[1])
		}
		items[idx] = setAnyObjectPath(next, parts[1:], value)
		return items
	}
	obj, _ := current.(map[string]any)
	if obj == nil {
		obj = map[string]any{}
	}
	if len(parts) == 1 {
		obj[part] = value
		return obj
	}
	next := obj[part]
	if !bindingContainerMatches(next, parts[1]) {
		next = newBindingContainer(parts[1])
	}
	obj[part] = setAnyObjectPath(next, parts[1:], value)
	return obj
}

func bindingContainerMatches(value any, nextPart string) bool {
	if _, wantArray := bindingPathIndex(nextPart); wantArray {
		_, ok := value.([]any)
		return ok
	}
	_, ok := value.(map[string]any)
	return ok
}

func newBindingContainer(nextPart string) any {
	if _, ok := bindingPathIndex(nextPart); ok {
		return []any{}
	}
	return map[string]any{}
}

func bindingPathIndex(part string) (int, bool) {
	if part == "" {
		return 0, false
	}
	for _, ch := range part {
		if ch < '0' || ch > '9' {
			return 0, false
		}
	}
	idx, err := strconv.Atoi(part)
	if err != nil || idx < 0 {
		return 0, false
	}
	return idx, true
}

func validateRuntimeBindingPath(path string, label string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		if path != "" {
			return fmt.Errorf("%s %q must not include surrounding whitespace", label, path)
		}
		return nil
	}
	if path != trimmed {
		return fmt.Errorf("%s %q must not include surrounding whitespace", label, path)
	}
	for _, part := range strings.Split(trimmed, ".") {
		partTrimmed := strings.TrimSpace(part)
		if partTrimmed == "" {
			return fmt.Errorf("%s %q contains empty path segment", label, trimmed)
		}
		if part != partTrimmed {
			return fmt.Errorf("%s %q contains path segment with surrounding whitespace", label, trimmed)
		}
	}
	return nil
}

func splitBindingPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

func cloneMapValue(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMapValue(typed)
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = cloneValue(typed[i])
		}
		return out
	default:
		return typed
	}
}
