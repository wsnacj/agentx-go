package tools

import (
	"sort"
	"strings"
)

type browserManagedOptInTargetSurface struct {
	Target       string
	Capabilities BrowserCapabilities
}

type browserManagedOptInActSurface struct {
	Targets []string
	Kinds   []string
}

type browserManagedOptInProjection struct {
	additionalActSurface browserManagedOptInActSurface
	optInActSurface      browserManagedOptInActSurface
	optInUnifiedSurface  browserManagedOptInActSurface
	actTargetsByKind     map[string][]string
	compatTargetsByTool  map[string][]string
	compatToolNames      []string
}

func browserManagedOptInTargetSurfaces(nodeBackend, sandboxBackend BrowserBackend) []browserManagedOptInTargetSurface {
	out := make([]browserManagedOptInTargetSurface, 0, 2)
	for _, candidate := range []struct {
		target  string
		backend BrowserBackend
	}{
		{target: "node", backend: nodeBackend},
		{target: "sandbox", backend: sandboxBackend},
	} {
		caps := browserCapabilitiesForConcreteBackend(candidate.backend)
		if !caps.SupportsAnyActKind() {
			continue
		}
		out = append(out, browserManagedOptInTargetSurface{
			Target:       candidate.target,
			Capabilities: caps,
		})
	}
	return out
}

func browserManagedExplicitActOptInRequested(enabled map[string]bool) bool {
	if len(enabled) == 0 {
		return false
	}
	return enabled["browser_act"] || enabled["browser"]
}

func browserManagedOptInProjectionRequested(enabled map[string]bool) bool {
	if len(enabled) == 0 {
		return false
	}
	if browserManagedExplicitActOptInRequested(enabled) {
		return true
	}
	for rawName, isEnabled := range enabled {
		if !isEnabled {
			continue
		}
		if IsBrowserCompatToolName(rawName) {
			return true
		}
	}
	return false
}

func browserManagedOptInProjectionForCapabilities(enabled map[string]bool, visible BrowserCapabilities, nodeBackend, sandboxBackend BrowserBackend) browserManagedOptInProjection {
	projection := browserManagedOptInProjection{
		actTargetsByKind:    map[string][]string{},
		compatTargetsByTool: map[string][]string{},
	}
	if !browserManagedOptInProjectionRequested(enabled) {
		return projection
	}
	for _, surface := range browserManagedOptInTargetSurfaces(nodeBackend, sandboxBackend) {
		for _, kind := range surface.Capabilities.SupportedActKinds() {
			if visible.SupportsActKind(kind) {
				continue
			}
			projection.actTargetsByKind[kind] = mergeToolMetadataStrings(
				projection.actTargetsByKind[kind],
				[]string{surface.Target},
			)
		}
	}
	if browserManagedExplicitActOptInRequested(enabled) {
		kinds := browserManagedOptInKindsFromTargets(projection.actTargetsByKind)
		if len(kinds) != 0 {
			targets := make([]string, 0, len(kinds))
			for _, kind := range kinds {
				targets = mergeToolMetadataStrings(targets, projection.actTargetsByKind[kind])
			}
			projection.additionalActSurface = browserManagedOptInActSurface{
				Targets: targets,
				Kinds:   kinds,
			}
			if !visible.SupportsTool("browser_act") {
				projection.optInActSurface = browserManagedOptInActSurface{
					Targets: append([]string(nil), targets...),
					Kinds:   append([]string(nil), kinds...),
				}
			}
			if enabled["browser"] && !visible.SupportsAnyActKind() {
				projection.optInUnifiedSurface = browserManagedOptInActSurface{
					Targets: append([]string(nil), targets...),
					Kinds:   append([]string(nil), kinds...),
				}
			}
		}
	}
	for _, name := range browserManagedOptInEnabledToolNames(enabled) {
		if !IsBrowserCompatToolName(name) || visible.SupportsTool(name) {
			continue
		}
		kind := browserCompatManagedOptInActKind(name)
		targets := projection.actTargetsForKind(kind)
		if kind == "" || len(targets) == 0 {
			continue
		}
		projection.compatTargetsByTool[name] = targets
		projection.compatToolNames = append(projection.compatToolNames, name)
	}
	return projection
}

func browserManagedOptInKindsFromTargets(targetsByKind map[string][]string) []string {
	if len(targetsByKind) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(targetsByKind))
	for kind, targets := range targetsByKind {
		if strings.TrimSpace(kind) == "" || len(targets) == 0 {
			continue
		}
		kinds = append(kinds, kind)
	}
	if len(kinds) == 0 {
		return nil
	}
	sort.Strings(kinds)
	return kinds
}

func browserManagedOptInEnabledToolNames(enabled map[string]bool) []string {
	if len(enabled) == 0 {
		return nil
	}
	names := make([]string, 0, len(enabled))
	for rawName, isEnabled := range enabled {
		if !isEnabled {
			continue
		}
		name := NormalizeToolName(rawName)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	return names
}

func (projection browserManagedOptInProjection) AdditionalActSurface() browserManagedOptInActSurface {
	return browserManagedOptInActSurface{
		Targets: append([]string(nil), projection.additionalActSurface.Targets...),
		Kinds:   append([]string(nil), projection.additionalActSurface.Kinds...),
	}
}

func (projection browserManagedOptInProjection) OptInActSurface() browserManagedOptInActSurface {
	return browserManagedOptInActSurface{
		Targets: append([]string(nil), projection.optInActSurface.Targets...),
		Kinds:   append([]string(nil), projection.optInActSurface.Kinds...),
	}
}

func (projection browserManagedOptInProjection) OptInUnifiedSurface() browserManagedOptInActSurface {
	return browserManagedOptInActSurface{
		Targets: append([]string(nil), projection.optInUnifiedSurface.Targets...),
		Kinds:   append([]string(nil), projection.optInUnifiedSurface.Kinds...),
	}
}

func (projection browserManagedOptInProjection) actTargetsForKind(kind string) []string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" || len(projection.actTargetsByKind) == 0 {
		return nil
	}
	return append([]string(nil), projection.actTargetsByKind[kind]...)
}

func (projection browserManagedOptInProjection) CompatTargets(name string) []string {
	name = NormalizeToolName(name)
	if name == "" || len(projection.compatTargetsByTool) == 0 {
		return nil
	}
	return append([]string(nil), projection.compatTargetsByTool[name]...)
}

func (projection browserManagedOptInProjection) MetadataSurfaces() map[string]browserManagedCompatMetadataSurface {
	out := map[string]browserManagedCompatMetadataSurface{}
	if actSurface := projection.OptInActSurface(); len(actSurface.Kinds) != 0 {
		out["browser_act"] = browserManagedCompatMetadataSurface{
			Targets: actSurface.Targets,
			Kinds:   actSurface.Kinds,
		}
	}
	if unifiedSurface := projection.OptInUnifiedSurface(); len(unifiedSurface.Kinds) != 0 {
		out["browser"] = browserManagedCompatMetadataSurface{
			Targets: unifiedSurface.Targets,
			Kinds:   unifiedSurface.Kinds,
		}
	}
	for _, name := range projection.compatToolNames {
		targets := projection.CompatTargets(name)
		if len(targets) == 0 {
			continue
		}
		out[name] = browserManagedCompatMetadataSurface{Targets: targets}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (projection browserManagedOptInProjection) DiagnosticsSurface(registered map[string]bool) browserRuntimeManagedOptInDiagnosticsSurface {
	if len(registered) == 0 {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	surface := browserRuntimeManagedOptInDiagnosticsSurface{}
	if registered["browser_act"] {
		actSurface := projection.AdditionalActSurface()
		surface.Targets = mergeToolMetadataStrings(surface.Targets, actSurface.Targets)
		surface.ActKinds = mergeToolMetadataStrings(surface.ActKinds, actSurface.Kinds)
		surface.Capabilities = mergeBrowserCapabilities(surface.Capabilities, BrowserCapabilitiesForActKinds(actSurface.Kinds))
	}
	for _, name := range projection.compatToolNames {
		if !registered[name] {
			continue
		}
		kind := browserCompatManagedOptInActKind(name)
		targets := projection.CompatTargets(name)
		if kind == "" || len(targets) == 0 {
			continue
		}
		surface.ToolNames = append(surface.ToolNames, name)
		surface.Targets = mergeToolMetadataStrings(surface.Targets, targets)
		surface.Capabilities = mergeBrowserCapabilities(surface.Capabilities, BrowserCapabilitiesForActKinds([]string{kind}))
	}
	surface.ToolNames = mergeToolMetadataStrings(nil, surface.ToolNames)
	surface.ActKinds = mergeToolMetadataStrings(nil, surface.ActKinds)
	surface.Targets = mergeToolMetadataStrings(nil, surface.Targets)
	return surface
}

func browserManagedAdditionalActSurfaceForCapabilities(enabled map[string]bool, visible BrowserCapabilities, nodeBackend, sandboxBackend BrowserBackend) browserManagedOptInActSurface {
	return browserManagedOptInProjectionForCapabilities(enabled, visible, nodeBackend, sandboxBackend).AdditionalActSurface()
}

func browserManagedOptInActSurfaceForCapabilities(enabled map[string]bool, visible BrowserCapabilities, nodeBackend, sandboxBackend BrowserBackend) browserManagedOptInActSurface {
	return browserManagedOptInProjectionForCapabilities(enabled, visible, nodeBackend, sandboxBackend).OptInActSurface()
}

func browserManagedOptInCompatTargets(enabled map[string]bool, visible BrowserCapabilities, name string, nodeBackend, sandboxBackend BrowserBackend) []string {
	return browserManagedOptInProjectionForCapabilities(enabled, visible, nodeBackend, sandboxBackend).CompatTargets(name)
}
