package domainmodule

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func resolveModuleConfig(ctx context.Context, manifests []Manifest, base Config, resolvers []ConfigResolver, errorDetails ErrorDetailsFunc) (Config, map[string]Diagnostics, error) {
	resolved := base
	diagnostics := map[string]Diagnostics{}
	for _, manifest := range manifests {
		moduleID := NormalizeID(manifest.ID)
		if moduleID == "" || resolved.Has(moduleID) {
			continue
		}
		for _, resolver := range resolvers {
			if resolver == nil {
				continue
			}
			value, resolverDiagnostics, err := resolver(ctx, manifest, resolved)
			if len(resolverDiagnostics) > 0 {
				diagnostics[moduleID] = append(diagnostics[moduleID], resolverDiagnostics...)
			}
			if err != nil {
				var details map[string]string
				if errorDetails != nil {
					details = errorDetails(err)
				}
				diagnostics[moduleID] = append(diagnostics[moduleID], NewDiagnostic(moduleID, SeverityError, DiagnosticConfigResolveError, "domain module config resolver failed", details))
				return resolved, diagnostics, fmt.Errorf("resolve domain module %q config: %w", moduleID, err)
			}
			if value == nil {
				continue
			}
			resolved = resolved.With(moduleID, value)
			diagnostics[moduleID] = append(diagnostics[moduleID], NewDiagnostic(moduleID, SeverityInfo, DiagnosticConfigResolved, "domain module config resolved by host resolver", nil))
			break
		}
	}
	return resolved, diagnostics, nil
}

func configRequirementDiagnostics(manifest Manifest, cfg Config) Diagnostics {
	if len(manifest.RequiredConfig) == 0 {
		return nil
	}
	value := cfg.Value(manifest.ID)
	out := Diagnostics{}
	for _, requirement := range manifest.RequiredConfig {
		key := strings.TrimSpace(requirement.Key)
		if key == "" || !requirement.Required || configValueHasKey(value, key) {
			continue
		}
		out = append(out, NewDiagnostic(manifest.ID, SeverityWarning, DiagnosticMissingConfig, "required module config is not present", map[string]string{
			"key": key, "description": requirement.Description,
		}))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func configValueHasKey(value any, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" || value == nil {
		return false
	}
	switch typed := value.(type) {
	case map[string]any:
		_, ok := typed[key]
		return ok
	case map[string]string:
		_, ok := typed[key]
		return ok
	case map[string]bool:
		_, ok := typed[key]
		return ok
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if strings.EqualFold(field.Name, key) {
			return true
		}
		jsonTag := strings.TrimSpace(strings.Split(field.Tag.Get("json"), ",")[0])
		if jsonTag != "" && jsonTag != "-" && jsonTag == key {
			return true
		}
	}
	return false
}
