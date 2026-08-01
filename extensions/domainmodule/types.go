// Package domainmodule provides the portable contract and registration
// coordinator used by compiled AgentX domain modules.
package domainmodule

import (
	"context"
	"fmt"
	"strings"
)

// Manifest describes the host-visible surface of a compiled domain module.
// Credentials, cache roots and product policy remain host-owned.
type Manifest struct {
	ID             string
	Name           string
	Version        string
	Description    string
	ExtensionRoot  string
	Skills         []string
	Tools          []string
	Packs          []string
	Workflows      []string
	RequiredConfig []ConfigRequirement
	OptionalConfig []ConfigRequirement
}

// ConfigRequirement is documentation and diagnostics metadata. It does not
// grant access to a secret or authorize a capability.
type ConfigRequirement struct {
	Key         string
	Description string
	Required    bool
}

// Config carries host-provided module-scoped values. The host retains
// ownership of each value and a module may type-assert only its own entry.
type Config struct {
	Modules map[string]any
}

// Value returns the value for moduleID after canonical-ID lookup, with a raw
// key fallback for compatibility with older hosts.
func (c Config) Value(moduleID string) any {
	rawID := strings.TrimSpace(moduleID)
	normalizedID := NormalizeID(rawID)
	if normalizedID == "" || len(c.Modules) == 0 {
		return nil
	}
	if value, ok := c.Modules[normalizedID]; ok {
		return value
	}
	return c.Modules[rawID]
}

// Has reports whether Config contains an entry for moduleID.
func (c Config) Has(moduleID string) bool {
	rawID := strings.TrimSpace(moduleID)
	normalizedID := NormalizeID(rawID)
	if normalizedID == "" || len(c.Modules) == 0 {
		return false
	}
	if _, ok := c.Modules[normalizedID]; ok {
		return true
	}
	_, ok := c.Modules[rawID]
	return ok
}

// With returns a copy containing value under the canonical module ID.
func (c Config) With(moduleID string, value any) Config {
	id := NormalizeID(moduleID)
	if id == "" || value == nil {
		return c
	}
	out := Config{Modules: map[string]any{}}
	for key, existing := range c.Modules {
		out.Modules[key] = existing
	}
	out.Modules[id] = value
	return out
}

// NormalizeID returns the canonical module ID used for duplicate detection and
// diagnostics.
func NormalizeID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// NormalizeManifest validates and canonicalizes a manifest without changing
// the caller-owned value.
func NormalizeManifest(manifest Manifest) (Manifest, error) {
	manifest.ID = NormalizeID(manifest.ID)
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Version = strings.TrimSpace(manifest.Version)
	manifest.Description = strings.TrimSpace(manifest.Description)
	manifest.ExtensionRoot = strings.TrimSpace(manifest.ExtensionRoot)
	if manifest.ID == "" {
		return manifest, fmt.Errorf("domain module manifest id is required")
	}
	if manifest.Name == "" {
		manifest.Name = manifest.ID
	}
	manifest.Skills = uniqueStrings(manifest.Skills)
	manifest.Tools = uniqueStrings(manifest.Tools)
	manifest.Packs = uniqueStrings(manifest.Packs)
	manifest.Workflows = uniqueStrings(manifest.Workflows)
	return manifest, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ConfigResolver lets a host derive module-scoped typed config at registration
// time. It must not implicitly grant authorization or import ambient secrets.
type ConfigResolver func(context.Context, Manifest, Config) (any, Diagnostics, error)

// ApplyFunc performs the host-owned mutation for one normalized manifest. It
// is called sequentially and may observe mutations made by earlier modules.
type ApplyFunc func(context.Context, Manifest, Config) (Diagnostics, error)

// Registration binds one manifest to its host-owned apply operation.
type Registration struct {
	Manifest Manifest
	Apply    ApplyFunc
}

// PreflightFunc validates host-owned constraints after every manifest has been
// normalized and duplicate-checked, but before config resolution or mutation.
type PreflightFunc func([]Manifest) error

// ConfigResolvedFunc receives the resolved snapshot immediately before the
// first registration is applied.
type ConfigResolvedFunc func(Config)

// ErrorDetailsFunc projects an internal resolver error into display-safe
// diagnostic details. The host owns redaction policy; returning nil is valid.
type ErrorDetailsFunc func(error) map[string]string

// RegisterOptions controls portable registration orchestration.
type RegisterOptions struct {
	Config             Config
	ConfigResolvers    []ConfigResolver
	Preflight          PreflightFunc
	ConfigResolved     ConfigResolvedFunc
	ConfigErrorDetails ErrorDetailsFunc
}
