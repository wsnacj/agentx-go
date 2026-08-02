package tools

import (
	"context"
	"os"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

var browserNow = time.Now

const (
	browserTabTargetWaitMs      = 250
	browserSafariScrollSettleMs = 150
	browserMaxScreenshotTiles   = 48
	browserMaxStitchedPixels    = 80_000_000
	browserElementRefPrefix     = "css1:"
	browserElementMetaRefPrefix = "elem1:"
)

// BrowserToolOptions combines portable browser runtime configuration with the
// host-owned side-effect ports required by optional features.
type BrowserToolOptions struct {
	Root                         string
	TimeoutMs                    int
	MaxChars                     int
	DefaultBrowserApp            string
	OpenWaitMs                   int
	ScreenshotWaitMs             int
	BrowserCDPEscapeHatchAllowed *bool
	BrowserLocalPlannerDryRun    bool
	BrowserLocalPlannerExecute   bool
	BrowserLocalPlannerModel     string
	BrowserLocalPlannerTimeoutMs int
	AllowPrivateHosts            bool
	AllowCIDRs                   []string
	DenyCIDRs                    []string
	AllowPorts                   []int
	DenyPorts                    []int
	EnabledTools                 []string
	Backend                      BrowserBackend
	// ImplicitHostBackend supplies a host fallback without promoting it to an
	// explicitly configured default route. Host adapters use this seam to keep
	// legacy host execution available while preserving the explicit-selection
	// and diagnostics semantics of a nil Backend.
	ImplicitHostBackend  BrowserBackend
	SandboxBackend       BrowserBackend
	NodeBackend          BrowserBackend
	SessionRegistry      *BrowserSessionRegistry
	SessionRunRegistry   BrowserSessionRunRegistry
	SessionStateRegistry *BrowserSessionStateRegistry
	PublishArtifact      BrowserArtifactPublisher
	LocalPlannerChat     BrowserLocalPlannerChat
	RunCommand           BrowserCommandRunner
	RepairScript         string
	AcceptanceScript     string
}

// BrowserArtifactPublisher atomically publishes one browser-produced file.
type BrowserArtifactPublisher func(context.Context, string, string, func(context.Context, *os.File, string) error) (int64, error)

// BrowserLocalPlannerChat invokes an explicitly configured model adapter and
// returns only its textual decision payload.
type BrowserLocalPlannerChat func(context.Context, string, string, string) (string, error)

// BrowserCommandRunner executes an explicitly authorized host command.
type BrowserCommandRunner func(context.Context, string, []string) ([]byte, error)
type BrowserBackend = agentxbrowserruntime.BrowserBackend
type BrowserArtifactActionBackend = agentxbrowserruntime.BrowserArtifactActionBackend
type BrowserConsoleActionBackend = agentxbrowserruntime.BrowserConsoleActionBackend
type BrowserRequestsActionBackend = agentxbrowserruntime.BrowserRequestsActionBackend
type BrowserResponseBodyActionBackend = agentxbrowserruntime.BrowserResponseBodyActionBackend
type BrowserErrorsActionBackend = agentxbrowserruntime.BrowserErrorsActionBackend
type BrowserCookiesActionBackend = agentxbrowserruntime.BrowserCookiesActionBackend
type BrowserCookiesMutatingActionBackend = agentxbrowserruntime.BrowserCookiesMutatingActionBackend
type BrowserTraceActionBackend = agentxbrowserruntime.BrowserTraceActionBackend
type BrowserStorageActionBackend = agentxbrowserruntime.BrowserStorageActionBackend
type BrowserStorageMutatingActionBackend = agentxbrowserruntime.BrowserStorageMutatingActionBackend
type BrowserHighlightActionBackend = agentxbrowserruntime.BrowserHighlightActionBackend
type BrowserRuntimeControlBackend = agentxbrowserruntime.BrowserRuntimeControlBackend
type BrowserRuntimeProfileManagementBackend = agentxbrowserruntime.BrowserRuntimeProfileManagementBackend
type BrowserDialogActionBackend = agentxbrowserruntime.BrowserDialogActionBackend
type BrowserUploadActionBackend = agentxbrowserruntime.BrowserUploadActionBackend
type BrowserPressActionBackend = agentxbrowserruntime.BrowserPressActionBackend
type BrowserHoverActionBackend = agentxbrowserruntime.BrowserHoverActionBackend
type BrowserDragActionBackend = agentxbrowserruntime.BrowserDragActionBackend
type BrowserSelectActionBackend = agentxbrowserruntime.BrowserSelectActionBackend
type BrowserFillActionBackend = agentxbrowserruntime.BrowserFillActionBackend
type BrowserResizeActionBackend = agentxbrowserruntime.BrowserResizeActionBackend
type BrowserCapabilities = agentxbrowserruntime.BrowserCapabilities
type BrowserCapabilityProvider = agentxbrowserruntime.BrowserCapabilityProvider
type BrowserRuntimeInfo = agentxbrowserruntime.BrowserRuntimeInfo
type BrowserRuntimeInfoProvider = agentxbrowserruntime.BrowserRuntimeInfoProvider
type BrowserRuntimeRouteResolver = agentxbrowserruntime.BrowserRuntimeRouteResolver
type BrowserRuntimeBackendRouter = agentxbrowserruntime.BrowserRuntimeBackendRouter
type BrowserDoctorCheckSummary = agentxbrowserruntime.BrowserDoctorCheckSummary
type BrowserDoctorRouteSummary = agentxbrowserruntime.BrowserDoctorRouteSummary
type BrowserDoctorLaunchSummary = agentxbrowserruntime.BrowserDoctorLaunchSummary
type BrowserDoctorBringupStep = agentxbrowserruntime.BrowserDoctorBringupStep
type BrowserDoctorBringupSummary = agentxbrowserruntime.BrowserDoctorBringupSummary
type BrowserDoctorSummary = agentxbrowserruntime.BrowserDoctorSummary
type BrowserSessionRegistry = agentxbrowserruntime.BrowserSessionRegistry
type BrowserSessionRoute = agentxbrowserruntime.BrowserSessionRoute
type BrowserSessionStateRegistry = agentxbrowserruntime.BrowserSessionStateRegistry
type BrowserSessionRunInfo = agentxbrowserruntime.SharedSessionRunInfo
type BrowserSessionRunRegistry = agentxbrowserruntime.SharedSessionRunRegistry
type BrowserSessionTarget = agentxbrowserruntime.BrowserSessionTarget
type BrowserOpenRequest = agentxbrowserruntime.BrowserOpenRequest
type BrowserOpenResult = agentxbrowserruntime.BrowserOpenResult
type BrowserNavigateRequest = agentxbrowserruntime.BrowserNavigateRequest
type BrowserNavigateResult = agentxbrowserruntime.BrowserNavigateResult
type BrowserExtractRequest = agentxbrowserruntime.BrowserExtractRequest
type BrowserExtractResult = agentxbrowserruntime.BrowserExtractResult
type BrowserSnapshotRequest = agentxbrowserruntime.BrowserSnapshotRequest
type BrowserSnapshotResult = agentxbrowserruntime.BrowserSnapshotResult
type BrowserSnapshotElement = agentxbrowserruntime.BrowserSnapshotElement
type BrowserElementResolverOutcome = agentxbrowserruntime.BrowserElementResolverOutcome
type BrowserElementHint = agentxbrowserruntime.BrowserElementHint
type BrowserScreenshotRequest = agentxbrowserruntime.BrowserScreenshotRequest
type BrowserScreenshotResult = agentxbrowserruntime.BrowserScreenshotResult
type BrowserConsoleRequest = agentxbrowserruntime.BrowserConsoleRequest
type BrowserConsoleMessage = agentxbrowserruntime.BrowserConsoleMessage
type BrowserConsoleResult = agentxbrowserruntime.BrowserConsoleResult
type BrowserRequestsRequest = agentxbrowserruntime.BrowserRequestsRequest
type BrowserRequestEntry = agentxbrowserruntime.BrowserRequestEntry
type BrowserRequestsResult = agentxbrowserruntime.BrowserRequestsResult
type BrowserResponseBodyRequest = agentxbrowserruntime.BrowserResponseBodyRequest
type BrowserResponseBodyResult = agentxbrowserruntime.BrowserResponseBodyResult
type BrowserErrorsRequest = agentxbrowserruntime.BrowserErrorsRequest
type BrowserErrorEntry = agentxbrowserruntime.BrowserErrorEntry
type BrowserErrorsResult = agentxbrowserruntime.BrowserErrorsResult
type BrowserCookiesRequest = agentxbrowserruntime.BrowserCookiesRequest
type BrowserCookiesSetRequest = agentxbrowserruntime.BrowserCookiesSetRequest
type BrowserCookiesClearRequest = agentxbrowserruntime.BrowserCookiesClearRequest
type BrowserCookieEntry = agentxbrowserruntime.BrowserCookieEntry
type BrowserCookiesResult = agentxbrowserruntime.BrowserCookiesResult
type BrowserTraceRequest = agentxbrowserruntime.BrowserTraceRequest
type BrowserTraceResult = agentxbrowserruntime.BrowserTraceResult
type BrowserStorageRequest = agentxbrowserruntime.BrowserStorageRequest
type BrowserStorageSetRequest = agentxbrowserruntime.BrowserStorageSetRequest
type BrowserStorageClearRequest = agentxbrowserruntime.BrowserStorageClearRequest
type BrowserStorageEntry = agentxbrowserruntime.BrowserStorageEntry
type BrowserStorageResult = agentxbrowserruntime.BrowserStorageResult
type BrowserOfflineActionBackend = agentxbrowserruntime.BrowserOfflineActionBackend
type BrowserOfflineRequest = agentxbrowserruntime.BrowserOfflineRequest
type BrowserOfflineResult = agentxbrowserruntime.BrowserOfflineResult
type BrowserHeadersActionBackend = agentxbrowserruntime.BrowserHeadersActionBackend
type BrowserHeadersRequest = agentxbrowserruntime.BrowserHeadersRequest
type BrowserHeadersResult = agentxbrowserruntime.BrowserHeadersResult
type BrowserCredentialsActionBackend = agentxbrowserruntime.BrowserCredentialsActionBackend
type BrowserCredentialsRequest = agentxbrowserruntime.BrowserCredentialsRequest
type BrowserCredentialsResult = agentxbrowserruntime.BrowserCredentialsResult
type BrowserGeolocationActionBackend = agentxbrowserruntime.BrowserGeolocationActionBackend
type BrowserGeolocationRequest = agentxbrowserruntime.BrowserGeolocationRequest
type BrowserGeolocationResult = agentxbrowserruntime.BrowserGeolocationResult
type BrowserMediaActionBackend = agentxbrowserruntime.BrowserMediaActionBackend
type BrowserMediaRequest = agentxbrowserruntime.BrowserMediaRequest
type BrowserMediaResult = agentxbrowserruntime.BrowserMediaResult
type BrowserTimezoneActionBackend = agentxbrowserruntime.BrowserTimezoneActionBackend
type BrowserTimezoneRequest = agentxbrowserruntime.BrowserTimezoneRequest
type BrowserTimezoneResult = agentxbrowserruntime.BrowserTimezoneResult
type BrowserLocaleActionBackend = agentxbrowserruntime.BrowserLocaleActionBackend
type BrowserLocaleRequest = agentxbrowserruntime.BrowserLocaleRequest
type BrowserLocaleResult = agentxbrowserruntime.BrowserLocaleResult
type BrowserDeviceActionBackend = agentxbrowserruntime.BrowserDeviceActionBackend
type BrowserDeviceRequest = agentxbrowserruntime.BrowserDeviceRequest
type BrowserDeviceResult = agentxbrowserruntime.BrowserDeviceResult
type BrowserHighlightRequest = agentxbrowserruntime.BrowserHighlightRequest
type BrowserHighlightResult = agentxbrowserruntime.BrowserHighlightResult

func BrowserDoctorBringupSummaryText(bringup *BrowserDoctorBringupSummary) string {
	return agentxbrowserruntime.BrowserDoctorBringupSummaryText(bringup)
}

func BrowserDoctorBringupDetailText(bringup *BrowserDoctorBringupSummary) string {
	return agentxbrowserruntime.BrowserDoctorBringupDetailText(bringup)
}

type BrowserProfileStatusRequest = agentxbrowserruntime.BrowserProfileStatusRequest
type BrowserProfileLifecycleRequest = agentxbrowserruntime.BrowserProfileLifecycleRequest
type BrowserProfileCreateRequest = agentxbrowserruntime.BrowserProfileCreateRequest
type BrowserProfileDeleteRequest = agentxbrowserruntime.BrowserProfileDeleteRequest
type BrowserProfileStatusResult = agentxbrowserruntime.BrowserProfileStatusResult
type BrowserProfileInfo = agentxbrowserruntime.BrowserProfileInfo
type BrowserProfilesRequest = agentxbrowserruntime.BrowserProfilesRequest
type BrowserProfilesResult = agentxbrowserruntime.BrowserProfilesResult
type BrowserClickRequest = agentxbrowserruntime.BrowserClickRequest
type BrowserClickResult = agentxbrowserruntime.BrowserClickResult
type BrowserDownloadRequest = agentxbrowserruntime.BrowserDownloadRequest
type BrowserDownloadResult = agentxbrowserruntime.BrowserDownloadResult
type BrowserWaitDownloadRequest = agentxbrowserruntime.BrowserWaitDownloadRequest
type BrowserWaitDownloadResult = agentxbrowserruntime.BrowserWaitDownloadResult
type BrowserSavePDFRequest = agentxbrowserruntime.BrowserSavePDFRequest
type BrowserSavePDFResult = agentxbrowserruntime.BrowserSavePDFResult
type BrowserSaveHTMLRequest = agentxbrowserruntime.BrowserSaveHTMLRequest
type BrowserSaveHTMLResult = agentxbrowserruntime.BrowserSaveHTMLResult
type BrowserDialogRequest = agentxbrowserruntime.BrowserDialogRequest
type BrowserDialogResult = agentxbrowserruntime.BrowserDialogResult
type BrowserUploadRequest = agentxbrowserruntime.BrowserUploadRequest
type BrowserUploadResult = agentxbrowserruntime.BrowserUploadResult
type BrowserPressRequest = agentxbrowserruntime.BrowserPressRequest
type BrowserPressResult = agentxbrowserruntime.BrowserPressResult
type BrowserHoverRequest = agentxbrowserruntime.BrowserHoverRequest
type BrowserHoverResult = agentxbrowserruntime.BrowserHoverResult
type BrowserDragRequest = agentxbrowserruntime.BrowserDragRequest
type BrowserDragResult = agentxbrowserruntime.BrowserDragResult
type BrowserSelectRequest = agentxbrowserruntime.BrowserSelectRequest
type BrowserSelectResult = agentxbrowserruntime.BrowserSelectResult
type BrowserFillField = agentxbrowserruntime.BrowserFillField
type BrowserFillRequest = agentxbrowserruntime.BrowserFillRequest
type BrowserFillResult = agentxbrowserruntime.BrowserFillResult
type BrowserResizeRequest = agentxbrowserruntime.BrowserResizeRequest
type BrowserResizeResult = agentxbrowserruntime.BrowserResizeResult
type BrowserTypeRequest = agentxbrowserruntime.BrowserTypeRequest
type BrowserTypeResult = agentxbrowserruntime.BrowserTypeResult
type BrowserEvalRequest = agentxbrowserruntime.BrowserEvalRequest
type BrowserEvalResult = agentxbrowserruntime.BrowserEvalResult
type BrowserActRequest = agentxbrowserruntime.BrowserActRequest
type BrowserActResult = agentxbrowserruntime.BrowserActResult
type BrowserTab = agentxbrowserruntime.BrowserTab
type BrowserTabsRequest = agentxbrowserruntime.BrowserTabsRequest
type BrowserTabsResult = agentxbrowserruntime.BrowserTabsResult

func NewBrowserSessionRegistry() *BrowserSessionRegistry {
	return agentxbrowserruntime.NewBrowserSessionRegistry()
}

func NewBrowserSessionStateRegistry() *BrowserSessionStateRegistry {
	return agentxbrowserruntime.NewBrowserSessionStateRegistry()
}

func RegisterBrowserTools(reg *llmxtools.Registry, opts BrowserToolOptions) {
	regCtx, enabled, ok := newBrowserRegistrationContext(reg, opts)
	if !ok {
		return
	}
	registerEnabledBrowserTools(regCtx, enabled)
}
