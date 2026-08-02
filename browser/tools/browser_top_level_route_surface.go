package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserTopLevelRouteSurface struct {
	BrowserSurface      string
	BrowserOptInTargets []string
}

func browserTopLevelRouteSurfaceLabel(browserSurface string, browserOptInTargets []string) (string, []string) {
	label := strings.TrimSpace(browserSurface)
	targets := mergeToolMetadataStrings(nil, browserOptInTargets)
	if label == "" && len(targets) == 0 {
		return "", nil
	}
	return label, targets
}

func browserTopLevelRouteSurfaceFromFields(browserSurface string, browserOptInTargets []string) browserTopLevelRouteSurface {
	label, targets := browserTopLevelRouteSurfaceLabel(browserSurface, browserOptInTargets)
	return browserTopLevelRouteSurface{
		BrowserSurface:      label,
		BrowserOptInTargets: targets,
	}
}

func browserTopLevelRouteSurfaceEmpty(route browserTopLevelRouteSurface) bool {
	return strings.TrimSpace(route.BrowserSurface) == "" && len(route.BrowserOptInTargets) == 0
}

func browserTopLevelRouteSurfaceFromShared(
	route *agentxbrowserruntime.SharedSessionBrowserRouteSurface,
) browserTopLevelRouteSurface {
	if route == nil {
		return browserTopLevelRouteSurface{}
	}
	return browserTopLevelRouteSurfaceFromFields(route.BrowserSurface, route.BrowserOptInTargets)
}

func browserTopLevelRouteSurfaceFromSurfaceSummary(surface *browserTopLevelSurfaceSummary) browserTopLevelRouteSurface {
	if surface == nil {
		return browserTopLevelRouteSurface{}
	}
	return browserTopLevelRouteSurfaceFromFields(surface.BrowserSurface, surface.BrowserOptInTargets)
}

func browserTopLevelRouteSurfaceFromViewSummary(view *browserTopLevelViewSummary) browserTopLevelRouteSurface {
	if view == nil {
		return browserTopLevelRouteSurface{}
	}
	return browserTopLevelRouteSurfaceFromFields(view.BrowserSurface, view.BrowserOptInTargets)
}

func browserTopLevelRouteSurfaceFromWorkbenchSummary(workbench *browserRuntimeWorkbenchSurfaceSummary) browserTopLevelRouteSurface {
	if workbench == nil {
		return browserTopLevelRouteSurface{}
	}
	return browserTopLevelRouteSurfaceFromFields(workbench.BrowserSurface, workbench.BrowserOptInTargets)
}

func browserTopLevelSurfaceWithRouteSurface(
	surface *browserTopLevelSurfaceSummary,
	browserSurface string,
	browserOptInTargets []string,
) *browserTopLevelSurfaceSummary {
	label, targets := browserTopLevelRouteSurfaceLabel(browserSurface, browserOptInTargets)
	if surface == nil && label == "" && len(targets) == 0 {
		return nil
	}
	out := browserCloneTopLevelSurfaceSummary(surface)
	if out == nil {
		out = &browserTopLevelSurfaceSummary{}
	}
	out.BrowserSurface = label
	out.BrowserOptInTargets = append([]string(nil), targets...)
	if browserTopLevelSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSurfaceWithRouteSurfaceSummary(
	surface *browserTopLevelSurfaceSummary,
	route browserTopLevelRouteSurface,
) *browserTopLevelSurfaceSummary {
	return browserTopLevelSurfaceWithRouteSurface(surface, route.BrowserSurface, route.BrowserOptInTargets)
}

func browserTopLevelViewWithRouteSurface(
	view *browserTopLevelViewSummary,
	browserSurface string,
	browserOptInTargets []string,
) *browserTopLevelViewSummary {
	label, targets := browserTopLevelRouteSurfaceLabel(browserSurface, browserOptInTargets)
	if view == nil && label == "" && len(targets) == 0 {
		return nil
	}
	out := browserCloneTopLevelViewSummary(view)
	if out == nil {
		out = &browserTopLevelViewSummary{}
	}
	out.BrowserSurface = label
	out.BrowserOptInTargets = append([]string(nil), targets...)
	if browserTopLevelViewEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelViewWithRouteSurfaceSummary(
	view *browserTopLevelViewSummary,
	route browserTopLevelRouteSurface,
) *browserTopLevelViewSummary {
	return browserTopLevelViewWithRouteSurface(view, route.BrowserSurface, route.BrowserOptInTargets)
}

func browserApplyTopLevelRouteSurface(
	surface **browserTopLevelSurfaceSummary,
	view **browserTopLevelViewSummary,
	browserSurface string,
	browserOptInTargets []string,
) {
	if surface != nil {
		*surface = browserTopLevelSurfaceWithRouteSurface(*surface, browserSurface, browserOptInTargets)
	}
	if view != nil {
		*view = browserTopLevelViewWithRouteSurface(*view, browserSurface, browserOptInTargets)
	}
}

func browserApplyTopLevelRouteSurfaceSummary(
	surface **browserTopLevelSurfaceSummary,
	view **browserTopLevelViewSummary,
	route browserTopLevelRouteSurface,
) {
	browserApplyTopLevelRouteSurface(surface, view, route.BrowserSurface, route.BrowserOptInTargets)
}
