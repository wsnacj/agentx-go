package tools

import (
	"context"
	"os"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
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

// BrowserCompatToolErrorf preserves the specialist-tool error prefix.
func BrowserCompatToolErrorf(kind string, format string, args ...any) error {
	return browserCompatToolErrorf(kind, format, args...)
}
