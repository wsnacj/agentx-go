package tools

import (
	"context"
	"io"
	"os"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
)

// BrowserResolvedExecutionRoute is the Experimental route value returned by a
// concrete Host backend that resolves an executable browser route.
type BrowserResolvedExecutionRoute = browserResolvedExecutionRoute

// BrowserArtifactResolveRequest is the Experimental request presented to a
// Host backend when an artifact path must be materialized locally.
type BrowserArtifactResolveRequest = browserArtifactResolveRequest

// BrowserElementTarget is the normalized locator value shared with concrete
// Host backends. Callers should obtain it from ResolveBrowserElementTargetWithHint.
type BrowserElementTarget = browserElementTarget

// BrowserDoctorRouteMetadata is host-observed route provenance for diagnostics.
type BrowserDoctorRouteMetadata struct {
	Source   string
	Endpoint string
}

// BrowserDoctorRouteMetadataProvider may be implemented by a concrete Host
// backend without importing product runtime types.
type BrowserDoctorRouteMetadataProvider interface {
	BrowserDoctorRouteMetadata() BrowserDoctorRouteMetadata
}

// DefaultBrowserCapabilities returns the legacy system-host capability set.
// It is exposed only for compatibility adapters; new backends should declare
// their actual, narrower capabilities.
func DefaultBrowserCapabilities() BrowserCapabilities { return defaultBrowserCapabilities() }

// NormalizeBrowserRuntimeInfo canonicalizes backend, profile and target names.
func NormalizeBrowserRuntimeInfo(info BrowserRuntimeInfo) BrowserRuntimeInfo {
	return normalizeBrowserRuntimeInfo(info)
}

// MergeBrowserRuntimeInfo overlays non-empty canonical route fields.
func MergeBrowserRuntimeInfo(base BrowserRuntimeInfo, overlay BrowserRuntimeInfo) BrowserRuntimeInfo {
	return mergeBrowserRuntimeInfo(base, overlay)
}

// BrowserRuntimeInfoForConcreteBackend resolves a backend-provided route with
// a host-selected fallback.
func BrowserRuntimeInfoForConcreteBackend(backend BrowserBackend, fallback BrowserRuntimeInfo) BrowserRuntimeInfo {
	return browserRuntimeInfoForConcreteBackend(backend, fallback)
}

// NewBrowserManagedRouteUnavailableError preserves the typed managed-route
// failure identity used by canonical diagnostics.
func NewBrowserManagedRouteUnavailableError(target string, endpoint string, cause error) error {
	return &browserManagedRouteUnavailableError{target: target, endpoint: endpoint, cause: cause}
}

// ResolveBrowserElementTargetWithHint normalizes the selector/ref/hint tuple.
func ResolveBrowserElementTargetWithHint(selector string, ref string, hint *BrowserElementHint) (BrowserElementTarget, error) {
	return resolveBrowserElementTargetWithHint(selector, ref, hint)
}

// BrowserElementResolverJS returns the bounded selector/ref resolver used by
// the legacy system-host adapter.
func BrowserElementResolverJS(target BrowserElementTarget) string {
	return browserElementResolverJS(target)
}

// BrowserPrefersSafari reports whether a host browser label selects Safari.
func BrowserPrefersSafari(browserApp string) bool { return prefersSafari(browserApp) }

// ParseSafariTabsPayload parses the legacy system-host tab observation payload.
func ParseSafariTabsPayload(raw string, activeIndex int) []BrowserTab {
	return parseSafariTabsPayload(raw, activeIndex)
}

// BrowserRemoteLocatorProjectionForTarget projects a normalized target into
// the transport-neutral remote locator contract.
func BrowserRemoteLocatorProjectionForTarget(target BrowserElementTarget) (string, string, *BrowserElementHint, *agentxbrowserruntime.BrowserElementResolverRequest, agentxbrowserruntime.BrowserElementRemoteProjection) {
	return browserRemoteLocatorProjectionForTarget(target)
}

// BrowserElementRefHasKnownPrefix reports whether a ref uses a canonical encoding.
func BrowserElementRefHasKnownPrefix(ref string) bool { return browserElementRefHasKnownPrefix(ref) }

// NormalizeBrowserSnapshotElements applies canonical element/ref projection.
func NormalizeBrowserSnapshotElements(elements []BrowserSnapshotElement, pageURL string, pageTitle string) []BrowserSnapshotElement {
	return browserNormalizeSnapshotElements(elements, pageURL, pageTitle)
}

// BrowserResolverOutcomeAllowsTargetTracking reports whether session tracking is safe.
func BrowserResolverOutcomeAllowsTargetTracking(outcome *BrowserElementResolverOutcome) bool {
	return browserResolverOutcomeAllowsTargetTracking(outcome)
}

// BrowserArtifactPublicationStageFile returns the active private publication stage.
func BrowserArtifactPublicationStageFile(ctx context.Context, targetPath string) (*os.File, bool) {
	return browserArtifactPublicationStageFile(ctx, targetPath)
}

// BrowserArtifactPublicationRequestedPath returns the caller-visible target.
func BrowserArtifactPublicationRequestedPath(ctx context.Context, targetPath string) (string, bool) {
	return browserArtifactPublicationRequestedPath(ctx, targetPath)
}

// BrowserArtifactSamePath compares canonicalized artifact paths.
func BrowserArtifactSamePath(left string, right string) bool {
	return browserArtifactSamePath(left, right)
}

// CopyBrowserArtifactWithContext copies an artifact while honoring cancellation.
func CopyBrowserArtifactWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	return copyBrowserArtifactWithContext(ctx, dst, src)
}

// BrowserElementRefForSnapshotElement encodes a stable ref for a host snapshot element.
func BrowserElementRefForSnapshotElement(element BrowserSnapshotElement, pageURL string, pageTitle string) string {
	return browserElementRefForSnapshotElement(element, pageURL, pageTitle)
}

// BrowserInventoryDefinition returns the canonical schema used for catalog
// inventory without registering a backend or executing a side effect.
func BrowserInventoryDefinition(name string) (types.Tool, bool) {
	switch NormalizeToolName(name) {
	case "browser":
		return browserUnifiedInventoryDefinition(), true
	case "browser_runtime":
		return browserRuntimeDefinition(browserUnifiedInventoryRuntimeActions()), true
	case "browser_act":
		return browserActDefinition(browserUnifiedInventoryActKinds()), true
	case "browser_open":
		return browserOpenDefinition(), true
	case "browser_navigate":
		return browserNavigateDefinition(), true
	case "browser_tabs":
		return browserTabsDefinition(), true
	case "browser_extract":
		return browserExtractDefinition(), true
	case "browser_screenshot":
		return browserScreenshotDefinition(), true
	case "browser_click":
		return browserClickDefinition(), true
	case "browser_type":
		return browserTypeDefinition(), true
	case "browser_eval":
		return browserEvalDefinition(), true
	default:
		return types.Tool{}, false
	}
}

// NormalizeBrowserToolToken applies canonical action/kind token normalization.
func NormalizeBrowserToolToken(raw string) string { return browserNormalizeToolToken(raw) }

// BrowserUnifiedActKind resolves a unified Browser action into its canonical act kind.
func BrowserUnifiedActKind(params map[string]any, action string) string {
	return browserUnifiedActKind(params, action)
}

// BrowserUnifiedRuntimeAction resolves a unified runtime alias.
func BrowserUnifiedRuntimeAction(action string) (string, bool) {
	alias, ok := browserUnifiedRuntimeActionAliases[browserNormalizeToolToken(action)]
	return alias.Action, ok
}

// BrowserCompatForceConfirmationNeedsGuardianReview reports the canonical
// force-confirmation posture for a compatibility tool. The Host still owns the
// approval decision and hook.
func BrowserCompatForceConfirmationNeedsGuardianReview(name string, params map[string]any) bool {
	return browserCompatForceConfirmationNeedsGuardianReview(name, params)
}

// InferBrowserToolMetadataMissing returns canonical metadata only for Browser
// definitions not already covered by complete Host metadata.
func InferBrowserToolMetadataMissing(defs []types.Tool, provided map[string]ToolMetadata) map[string]ToolMetadata {
	return inferBrowserToolMetadataMissing(defs, provided)
}

// BrowserArtifactSourceForTool returns the canonical media artifact source label.
func BrowserArtifactSourceForTool(name string) string { return browserArtifactSourceForTool(name) }

// BrowserCompatToolErrorf preserves the specialist-tool error prefix.
func BrowserCompatToolErrorf(kind string, format string, args ...any) error {
	return browserCompatToolErrorf(kind, format, args...)
}
