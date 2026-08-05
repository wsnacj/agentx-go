package expert

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"
)

const (
	maxSpecBytes        = 256 << 10
	maxIdentityBytes    = 128
	maxNameBytes        = 512
	maxDescriptionBytes = 16 << 10
	maxInstructionBytes = 128 << 10
	maxRequirements     = 128
	maxTags             = 64
)

var forbiddenHostFields = map[string]bool{
	"approval": true, "approval_mode": true, "budget": true,
	"credential": true, "credentials": true, "memory": true, "memory_state": true,
	"model": true, "model_id": true, "provider": true, "provider_id": true,
	"runtime": true, "runtime_policy": true, "sandbox": true,
	"secret": true, "secrets": true, "session": true, "session_id": true,
	"tenant": true, "tenant_id": true, "timeout": true, "timeout_ms": true,
}

// Parse validates JSON, rejects Host-owned fields and returns a detached,
// normalized Spec. Unknown descriptive fields are ignored for forward
// compatibility.
func Parse(content []byte) (Spec, error) {
	if len(content) == 0 || len(content) > maxSpecBytes {
		return Spec{}, specError(ErrorCodeInvalidSpec, "spec size is outside bounds")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil {
		return Spec{}, specError(ErrorCodeInvalidSpec, err.Error())
	}
	for field := range fields {
		if forbiddenHostFields[strings.ToLower(strings.TrimSpace(field))] {
			return Spec{}, specError(ErrorCodeForbiddenField, field)
		}
	}
	var raw Spec
	if err := json.Unmarshal(content, &raw); err != nil {
		return Spec{}, specError(ErrorCodeInvalidSpec, err.Error())
	}
	return Normalize(raw)
}

// Normalize validates and canonicalizes a role without mutating caller-owned
// slices or changing the instruction body beyond trimming outer whitespace.
func Normalize(raw Spec) (Spec, error) {
	id := normalizeID(raw.ID)
	name := strings.TrimSpace(raw.Name)
	schema := strings.ToLower(strings.TrimSpace(raw.SchemaVersion))
	if schema == "" {
		schema = SchemaVersionV1
	}
	if schema != SchemaVersionV1 {
		return Spec{}, specError(ErrorCodeUnsupportedSchema, schema)
	}
	description := strings.TrimSpace(raw.Description)
	instructions := strings.TrimSpace(raw.Instructions)
	if id == "" || name == "" || instructions == "" || len(id) > maxIdentityBytes || len(name) > maxNameBytes ||
		len(description) > maxDescriptionBytes || len(instructions) > maxInstructionBytes ||
		len(raw.Requirements) > maxRequirements || len(raw.Tags) > maxTags {
		return Spec{}, specError(ErrorCodeInvalidSpec, "required field or size bound failed")
	}
	requirements, err := normalizeRequirements(raw.Requirements)
	if err != nil {
		return Spec{}, err
	}
	tags, err := normalizeTags(raw.Tags)
	if err != nil {
		return Spec{}, err
	}
	return Spec{
		ID: id, SchemaVersion: schema, Name: name, Version: strings.TrimSpace(raw.Version),
		Description: description, Instructions: instructions, Requirements: requirements, Tags: tags,
	}, nil
}

// Project creates a discovery-only Expert asset. Instructions and capability
// requirements intentionally remain outside the catalog envelope.
func Project(sourceRef string, raw Spec) (capabilitycatalog.Asset, error) {
	normalized, err := Normalize(raw)
	if err != nil {
		return capabilitycatalog.Asset{}, err
	}
	return capabilitycatalog.Asset{
		Identity: capabilitycatalog.Identity{Kind: capabilitycatalog.KindExpert, ID: normalized.ID},
		Name:     normalized.Name, Description: normalized.Description, Version: normalized.Version,
		SourceRef: strings.TrimSpace(sourceRef), Tags: append([]string(nil), normalized.Tags...),
	}, nil
}

func normalizeRequirements(raw []Requirement) ([]Requirement, error) {
	seen := map[string]bool{}
	out := make([]Requirement, 0, len(raw))
	for _, item := range raw {
		kind := capabilitycatalog.Kind(strings.ToLower(strings.TrimSpace(string(item.Kind))))
		if kind != capabilitycatalog.KindTool && kind != capabilitycatalog.KindSkill &&
			kind != capabilitycatalog.KindPlugin && kind != capabilitycatalog.KindConnector {
			return nil, specError(ErrorCodeInvalidSpec, "requirement kind is unsupported")
		}
		id := normalizeID(item.ID)
		if id == "" {
			return nil, specError(ErrorCodeInvalidSpec, "requirement id is invalid")
		}
		key := string(kind) + ":" + id
		if seen[key] {
			return nil, specError(ErrorCodeInvalidSpec, "requirement is duplicated")
		}
		seen[key] = true
		out = append(out, Requirement{Kind: kind, ID: id, Optional: item.Optional})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind == out[j].Kind {
			return out[i].ID < out[j].ID
		}
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

func normalizeTags(raw []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" || len(value) > maxIdentityBytes {
			return nil, specError(ErrorCodeInvalidSpec, "tag is invalid")
		}
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out, nil
}

func normalizeID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || len(value) > maxIdentityBytes {
		return ""
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return ""
		}
	}
	return value
}

func specError(code ErrorCode, detail string) error {
	return &Error{Code: code, Cause: fmt.Errorf("%s", detail)}
}
