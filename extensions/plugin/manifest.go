package plugin

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
)

const maxManifestBytes = 256 << 10

var forbiddenHostFields = map[string]bool{
	"allow_tools": true, "allowed_tools": true,
	"approval": true, "approval_mode": true,
	"budget": true, "credential": true, "credentials": true,
	"execution_policy": true, "policy": true, "policy_profile": true,
	"profile": true, "runtime_policy": true, "sandbox": true,
	"secret": true, "secrets": true, "token": true,
	"timeout": true, "timeout_ms": true, "tool_policy": true,
}

// Parse validates JSON and returns a detached normalized manifest. Unknown
// descriptive fields are ignored for forward compatibility; known Host-owned
// policy and credential fields fail closed.
func Parse(content []byte) (Manifest, error) {
	if len(content) == 0 || len(content) > maxManifestBytes {
		return Manifest{}, manifestError(ErrorCodeInvalidManifest, "manifest size is outside bounds")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil {
		return Manifest{}, manifestError(ErrorCodeInvalidManifest, err.Error())
	}
	if err := validateHostOwnedFields(fields); err != nil {
		return Manifest{}, err
	}
	var raw Manifest
	if err := json.Unmarshal(content, &raw); err != nil {
		return Manifest{}, manifestError(ErrorCodeInvalidManifest, err.Error())
	}
	return Normalize(raw)
}

// Normalize validates and canonicalizes a caller-provided manifest without
// mutating the caller's slices.
func Normalize(raw Manifest) (Manifest, error) {
	name := normalizeID(raw.Name)
	if name == "" {
		return Manifest{}, manifestError(ErrorCodeInvalidManifest, "name is required")
	}
	schema := strings.ToLower(strings.TrimSpace(raw.SchemaVersion))
	if schema == "" {
		schema = SchemaVersionV1
	}
	if schema != SchemaVersionV1 {
		return Manifest{}, manifestError(ErrorCodeUnsupportedSchema, schema)
	}
	trust := normalizeTrustBoundary(raw.TrustBoundary)
	if trust == "" {
		return Manifest{}, manifestError(ErrorCodeInvalidManifest, "trust boundary is unsupported")
	}
	roots, err := normalizeRoots(raw.Roots)
	if err != nil {
		return Manifest{}, err
	}
	entrypoints, err := normalizeEntrypoints(raw.Entrypoints, roots)
	if err != nil {
		return Manifest{}, err
	}
	dependencies, err := normalizeDependencies(raw.Dependencies)
	if err != nil {
		return Manifest{}, err
	}
	permissions, err := normalizePermissionRequests(raw.RequestedPermissions)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		Name:                 name,
		SchemaVersion:        schema,
		Version:              strings.TrimSpace(raw.Version),
		Description:          strings.TrimSpace(raw.Description),
		TrustBoundary:        trust,
		Roots:                roots,
		Entrypoints:          entrypoints,
		Dependencies:         dependencies,
		RequestedPermissions: permissions,
	}, nil
}

func validateHostOwnedFields(fields map[string]json.RawMessage) error {
	names := make([]string, 0, len(fields))
	for field := range fields {
		names = append(names, field)
	}
	sort.Strings(names)
	for _, field := range names {
		value := fields[field]
		key := strings.ToLower(strings.TrimSpace(field))
		if forbiddenHostFields[key] {
			return manifestError(ErrorCodeForbiddenField, key)
		}
		if key != "metadata" {
			continue
		}
		var metadata map[string]json.RawMessage
		if err := json.Unmarshal(value, &metadata); err != nil {
			continue
		}
		nestedNames := make([]string, 0, len(metadata))
		for nested := range metadata {
			nestedNames = append(nestedNames, nested)
		}
		sort.Strings(nestedNames)
		for _, nested := range nestedNames {
			key = strings.ToLower(strings.TrimSpace(nested))
			if forbiddenHostFields[key] {
				return manifestError(ErrorCodeForbiddenField, "metadata."+key)
			}
		}
	}
	return nil
}

func normalizeID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
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

func normalizeTrustBoundary(raw TrustBoundary) TrustBoundary {
	switch TrustBoundary(strings.ToLower(strings.TrimSpace(string(raw)))) {
	case "", TrustBoundaryWorkspace:
		return TrustBoundaryWorkspace
	case TrustBoundaryReviewed:
		return TrustBoundaryReviewed
	case TrustBoundaryTrusted:
		return TrustBoundaryTrusted
	default:
		return ""
	}
}

func normalizeRoots(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return []string{"."}, nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value, err := normalizeContainedPath(item, true)
		if err != nil {
			return nil, err
		}
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	if len(out) == 0 {
		return []string{"."}, nil
	}
	return out, nil
}

func normalizeEntrypoints(raw Entrypoints, roots []string) (Entrypoints, error) {
	values := []*string{&raw.Skills, &raw.Tools, &raw.Hooks, &raw.Commands}
	for _, value := range values {
		if strings.TrimSpace(*value) == "" {
			*value = ""
			continue
		}
		normalized, err := normalizeContainedPath(*value, false)
		if err != nil {
			return Entrypoints{}, err
		}
		if !withinRoots(normalized, roots) {
			return Entrypoints{}, manifestError(ErrorCodeInvalidPath, "entrypoint is outside declared roots")
		}
		*value = normalized
	}
	return raw, nil
}

func normalizeContainedPath(raw string, allowDot bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsRune(value, '\x00') {
		return "", manifestError(ErrorCodeInvalidPath, "path is empty")
	}
	value = strings.ReplaceAll(value, "\\", "/")
	if strings.HasPrefix(value, "/") || (len(value) >= 2 && value[1] == ':') {
		return "", manifestError(ErrorCodeInvalidPath, "path is absolute")
	}
	value = path.Clean(value)
	if value == "." {
		if allowDot {
			return value, nil
		}
		return "", manifestError(ErrorCodeInvalidPath, "entrypoint cannot be the plugin root")
	}
	if value == ".." || strings.HasPrefix(value, "../") {
		return "", manifestError(ErrorCodeInvalidPath, "path escapes plugin root")
	}
	return value, nil
}

func withinRoots(entrypoint string, roots []string) bool {
	for _, root := range roots {
		if root == "." || entrypoint == root || strings.HasPrefix(entrypoint, root+"/") {
			return true
		}
	}
	return false
}

func normalizeDependencies(raw []Dependency) ([]Dependency, error) {
	seen := map[string]bool{}
	out := make([]Dependency, 0, len(raw))
	for _, item := range raw {
		kind := strings.ToLower(strings.TrimSpace(item.Kind))
		if kind != "plugin" && kind != "connector" {
			return nil, manifestError(ErrorCodeInvalidManifest, "dependency kind is unsupported")
		}
		id := normalizeID(item.ID)
		if id == "" {
			return nil, manifestError(ErrorCodeInvalidManifest, "dependency id is invalid")
		}
		key := kind + ":" + id
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Dependency{Kind: kind, ID: id, Version: strings.TrimSpace(item.Version)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind == out[j].Kind {
			return out[i].ID < out[j].ID
		}
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

func normalizePermissionRequests(raw []PermissionRequest) ([]PermissionRequest, error) {
	seen := map[string]bool{}
	out := make([]PermissionRequest, 0, len(raw))
	for _, item := range raw {
		capability := normalizeID(item.Capability)
		if capability == "" {
			return nil, manifestError(ErrorCodeInvalidManifest, "permission capability is invalid")
		}
		if seen[capability] {
			continue
		}
		seen[capability] = true
		out = append(out, PermissionRequest{Capability: capability, Reason: strings.TrimSpace(item.Reason)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Capability < out[j].Capability })
	return out, nil
}

func manifestError(code ErrorCode, detail string) error {
	return &Error{Code: code, Cause: fmt.Errorf("%s", detail)}
}
