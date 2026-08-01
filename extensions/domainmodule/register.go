package domainmodule

import (
	"context"
	"fmt"
	"strconv"
)

// RegisterAll normalizes and validates every manifest before mutation,
// resolves host-provided config, and then applies registrations sequentially.
// It deliberately preserves successful earlier mutations when a later apply
// fails; callers must not treat the operation as atomic.
func RegisterAll(ctx context.Context, registrations []Registration, opts RegisterOptions) (Report, error) {
	var report Report
	normalized := make([]Registration, 0, len(registrations))
	manifests := make([]Manifest, 0, len(registrations))
	seen := map[string]Manifest{}
	for _, registration := range registrations {
		manifest, err := NormalizeManifest(registration.Manifest)
		if err != nil {
			return report, err
		}
		if previous, exists := seen[manifest.ID]; exists {
			report.appendDiagnostics(manifest, Diagnostics{
				NewDiagnostic(manifest.ID, SeverityError, DiagnosticModuleDuplicateID, "duplicate domain module id", map[string]string{
					"previous_name": previous.Name,
					"next_name":     manifest.Name,
				}),
			})
			return report, fmt.Errorf("duplicate domain module id %q", manifest.ID)
		}
		seen[manifest.ID] = manifest
		registration.Manifest = manifest
		normalized = append(normalized, registration)
		manifests = append(manifests, manifest)
	}
	if opts.Preflight != nil {
		if err := opts.Preflight(append([]Manifest(nil), manifests...)); err != nil {
			return report, err
		}
	}
	resolved, resolverDiagnostics, err := resolveModuleConfig(ctx, manifests, opts.Config, opts.ConfigResolvers, opts.ConfigErrorDetails)
	for _, manifest := range manifests {
		report.appendDiagnostics(manifest, resolverDiagnostics[manifest.ID])
	}
	if err != nil {
		return report, err
	}
	if opts.ConfigResolved != nil {
		opts.ConfigResolved(resolved)
	}
	for _, registration := range normalized {
		manifest := registration.Manifest
		report.appendDiagnostics(manifest, configRequirementDiagnostics(manifest, resolved))
		if registration.Apply != nil {
			diagnostics, err := registration.Apply(ctx, manifest, resolved)
			report.appendDiagnostics(manifest, diagnostics)
			if err != nil {
				return report, err
			}
		}
		report.appendDiagnostics(manifest, Diagnostics{
			NewDiagnostic(manifest.ID, SeverityInfo, DiagnosticModuleRegistered, "domain module registered", map[string]string{
				"skills": strconv.Itoa(len(manifest.Skills)), "tools": strconv.Itoa(len(manifest.Tools)),
				"packs": strconv.Itoa(len(manifest.Packs)), "workflows": strconv.Itoa(len(manifest.Workflows)),
			}),
		})
	}
	return report, nil
}
