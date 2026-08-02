package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserTopLevelCapabilitySurface struct {
	BrowserTools     []string
	ArtifactTools    []string
	ArtifactKinds    []string
	ArtifactContract string
	BrowserActKinds  []string
}

func browserTopLevelCapabilitySurfaceFromFields(
	browserTools []string,
	artifactTools []string,
	artifactKinds []string,
	artifactContract string,
	browserActKinds []string,
) browserTopLevelCapabilitySurface {
	return browserTopLevelCapabilitySurface{
		BrowserTools:     mergeToolMetadataStrings(nil, browserTools),
		ArtifactTools:    mergeToolMetadataStrings(nil, artifactTools),
		ArtifactKinds:    mergeToolMetadataStrings(nil, artifactKinds),
		ArtifactContract: strings.TrimSpace(artifactContract),
		BrowserActKinds:  mergeToolMetadataStrings(nil, browserActKinds),
	}
}

func browserTopLevelCapabilitySurfaceFromMetadata(
	metadata browserRuntimeCapabilityMetadata,
) browserTopLevelCapabilitySurface {
	return browserTopLevelCapabilitySurfaceFromFields(
		metadata.BrowserTools,
		metadata.ArtifactTools,
		metadata.ArtifactKinds,
		metadata.ArtifactContract,
		metadata.BrowserActKinds,
	)
}

func browserTopLevelCapabilitySurfaceFromShared(
	capability *agentxbrowserruntime.SharedSessionBrowserCapabilitySurface,
) browserTopLevelCapabilitySurface {
	if capability == nil {
		return browserTopLevelCapabilitySurface{}
	}
	return browserTopLevelCapabilitySurfaceFromFields(
		capability.BrowserTools,
		capability.ArtifactTools,
		capability.ArtifactKinds,
		capability.ArtifactContract,
		capability.BrowserActKinds,
	)
}

func browserTopLevelCapabilitySurfaceEmpty(surface browserTopLevelCapabilitySurface) bool {
	return len(surface.BrowserTools) == 0 &&
		len(surface.ArtifactTools) == 0 &&
		len(surface.ArtifactKinds) == 0 &&
		strings.TrimSpace(surface.ArtifactContract) == "" &&
		len(surface.BrowserActKinds) == 0
}

func browserTopLevelCapabilitySurfaceMerged(
	primary browserTopLevelCapabilitySurface,
	fallback browserTopLevelCapabilitySurface,
) browserTopLevelCapabilitySurface {
	out := browserTopLevelCapabilitySurfaceFromFields(
		primary.BrowserTools,
		primary.ArtifactTools,
		primary.ArtifactKinds,
		primary.ArtifactContract,
		primary.BrowserActKinds,
	)
	if len(out.BrowserTools) == 0 {
		out.BrowserTools = append([]string(nil), fallback.BrowserTools...)
	}
	if len(out.ArtifactTools) == 0 {
		out.ArtifactTools = append([]string(nil), fallback.ArtifactTools...)
	}
	if len(out.ArtifactKinds) == 0 {
		out.ArtifactKinds = append([]string(nil), fallback.ArtifactKinds...)
	}
	if out.ArtifactContract == "" {
		out.ArtifactContract = strings.TrimSpace(fallback.ArtifactContract)
	}
	if len(out.BrowserActKinds) == 0 {
		out.BrowserActKinds = append([]string(nil), fallback.BrowserActKinds...)
	}
	return out
}

func browserTopLevelCapabilitySurfaceFromSurfaceSummary(
	surface *browserTopLevelSurfaceSummary,
) browserTopLevelCapabilitySurface {
	if surface == nil {
		return browserTopLevelCapabilitySurface{}
	}
	return browserTopLevelCapabilitySurfaceFromFields(
		surface.BrowserTools,
		surface.ArtifactTools,
		surface.ArtifactKinds,
		surface.ArtifactContract,
		surface.BrowserActKinds,
	)
}

func browserTopLevelCapabilitySurfaceFromViewSummary(
	view *browserTopLevelViewSummary,
) browserTopLevelCapabilitySurface {
	if view == nil {
		return browserTopLevelCapabilitySurface{}
	}
	return browserTopLevelCapabilitySurfaceFromFields(
		view.BrowserTools,
		view.ArtifactTools,
		view.ArtifactKinds,
		view.ArtifactContract,
		view.BrowserActKinds,
	)
}

func browserTopLevelCapabilitySurfaceFromWorkbenchSummary(
	workbench *browserRuntimeWorkbenchSurfaceSummary,
) browserTopLevelCapabilitySurface {
	if workbench == nil {
		return browserTopLevelCapabilitySurface{}
	}
	return browserTopLevelCapabilitySurfaceFromFields(
		workbench.BrowserTools,
		workbench.ArtifactTools,
		workbench.ArtifactKinds,
		workbench.ArtifactContract,
		workbench.BrowserActKinds,
	)
}

func browserTopLevelSurfaceWithCapabilitySurfaceSummary(
	surface *browserTopLevelSurfaceSummary,
	capability browserTopLevelCapabilitySurface,
) *browserTopLevelSurfaceSummary {
	if surface == nil && browserTopLevelCapabilitySurfaceEmpty(capability) {
		return nil
	}
	out := browserCloneTopLevelSurfaceSummary(surface)
	if out == nil {
		out = &browserTopLevelSurfaceSummary{}
	}
	out.BrowserTools = append([]string(nil), capability.BrowserTools...)
	out.ArtifactTools = append([]string(nil), capability.ArtifactTools...)
	out.ArtifactKinds = append([]string(nil), capability.ArtifactKinds...)
	out.ArtifactContract = strings.TrimSpace(capability.ArtifactContract)
	out.BrowserActKinds = append([]string(nil), capability.BrowserActKinds...)
	if browserTopLevelSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelViewWithCapabilitySurfaceSummary(
	view *browserTopLevelViewSummary,
	capability browserTopLevelCapabilitySurface,
) *browserTopLevelViewSummary {
	if view == nil && browserTopLevelCapabilitySurfaceEmpty(capability) {
		return nil
	}
	out := browserTopLevelViewWithDefaultCandidateRoute(view, browserRuntimeRouteDescriptorFromTopLevelView(view))
	if out == nil {
		out = &browserTopLevelViewSummary{}
	}
	out.BrowserTools = append([]string(nil), capability.BrowserTools...)
	out.ArtifactTools = append([]string(nil), capability.ArtifactTools...)
	out.ArtifactKinds = append([]string(nil), capability.ArtifactKinds...)
	out.ArtifactContract = strings.TrimSpace(capability.ArtifactContract)
	out.BrowserActKinds = append([]string(nil), capability.BrowserActKinds...)
	if browserTopLevelViewEmpty(*out) {
		return nil
	}
	return out
}

func browserWorkbenchWithCapabilitySurfaceSummary(
	workbench *browserRuntimeWorkbenchSurfaceSummary,
	capability browserTopLevelCapabilitySurface,
) *browserRuntimeWorkbenchSurfaceSummary {
	if workbench == nil && browserTopLevelCapabilitySurfaceEmpty(capability) {
		return nil
	}
	out := browserWorkbenchWithDefaultCandidateRoute(workbench, browserRuntimeRouteDescriptorFromWorkbenchSurface(workbench))
	if out == nil {
		out = &browserRuntimeWorkbenchSurfaceSummary{}
	}
	out.BrowserTools = append([]string(nil), capability.BrowserTools...)
	out.ArtifactTools = append([]string(nil), capability.ArtifactTools...)
	out.ArtifactKinds = append([]string(nil), capability.ArtifactKinds...)
	out.ArtifactContract = strings.TrimSpace(capability.ArtifactContract)
	out.BrowserActKinds = append([]string(nil), capability.BrowserActKinds...)
	if browserUnifiedWorkbenchEmpty(*out) {
		return nil
	}
	return out
}

func browserApplyTopLevelCapabilitySurfaceSummary(
	surface **browserTopLevelSurfaceSummary,
	view **browserTopLevelViewSummary,
	capability browserTopLevelCapabilitySurface,
) {
	if surface != nil {
		*surface = browserTopLevelSurfaceWithCapabilitySurfaceSummary(*surface, capability)
	}
	if view != nil {
		*view = browserTopLevelViewWithCapabilitySurfaceSummary(*view, capability)
	}
}

func browserApplyWorkbenchCapabilitySurfaceSummary(
	workbench **browserRuntimeWorkbenchSurfaceSummary,
	capability browserTopLevelCapabilitySurface,
) {
	if workbench != nil {
		*workbench = browserWorkbenchWithCapabilitySurfaceSummary(*workbench, capability)
	}
}
