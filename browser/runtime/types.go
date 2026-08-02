package browserruntime

import (
	"context"
	"encoding/json"
	"time"
)

// BrowserToolOptions defines the shared browser runtime configuration surface.
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
	SandboxBackend               BrowserBackend
	NodeBackend                  BrowserBackend
	SessionRegistry              *BrowserSessionRegistry
	SessionRunRegistry           SharedSessionRunRegistry
	SessionStateRegistry         *BrowserSessionStateRegistry
}

// BrowserLocalPlannerResultSummary is the compact browser-local planner
// telemetry surface emitted in browser_act payloads. It supports both dry-run
// diagnostics and behind-flag single-step constrained execution summaries.
type BrowserLocalPlannerResultSummary struct {
	Mode                   string `json:"mode,omitempty"`
	Eligible               bool   `json:"eligible,omitempty"`
	ReasonCode             string `json:"reason_code,omitempty"`
	FollowupKind           string `json:"followup_kind,omitempty"`
	FailedCheck            string `json:"failed_check,omitempty"`
	FailureReason          string `json:"failure_reason,omitempty"`
	ManualRetryHint        string `json:"manual_retry_hint,omitempty"`
	RecoveryAction         string `json:"recovery_action,omitempty"`
	RequiresForce          bool   `json:"requires_force,omitempty"`
	BlockedReason          string `json:"blocked_reason,omitempty"`
	Decision               string `json:"decision,omitempty"`
	Model                  string `json:"model,omitempty"`
	LatencyMs              int64  `json:"latency_ms,omitempty"`
	FollowupOK             bool   `json:"followup_ok,omitempty"`
	FollowupRecovered      bool   `json:"followup_recovered,omitempty"`
	DiscardedInvalidOutput bool   `json:"discarded_invalid_output,omitempty"`
}

// SharedSessionRunInfo is a compact cross-tool session run snapshot that can be
// surfaced by browser/session inspection tools without depending on node internals.
type SharedSessionRunInfo struct {
	RunID    string
	NodeID   string
	Status   string
	Provider string
	Action   string
}

// SharedSessionRunRegistry exposes session-scoped remote run snapshots.
type SharedSessionRunRegistry interface {
	SnapshotSessionRuns(sessionID string) []SharedSessionRunInfo
}

// SharedSessionBrowserProfileState is a session-scoped browser lifecycle snapshot
// that can be shared across browser/node tools.
type SharedSessionBrowserProfileState struct {
	Backend       string
	Profile       string
	RuntimeTarget string
	BrowserApp    string
	Status        string
	Running       bool
	Connected     bool
	Note          string
	ObservedAt    time.Time
	StatusSince   time.Time
}

// SharedSessionBrowserProfileSelection is a session-scoped preferred managed
// browser profile selection that browser tools can reuse when the caller omits
// an explicit profile/runtime target.
type SharedSessionBrowserProfileSelection struct {
	Backend       string
	Profile       string
	RuntimeTarget string
	BrowserApp    string
	Source        string
}

// SharedSessionBrowserStateRegistry exposes session-scoped browser lifecycle
// snapshots that can be surfaced across browser/node tools.
type SharedSessionBrowserStateRegistry interface {
	SnapshotSessionBrowserProfiles(sessionID string) []SharedSessionBrowserProfileState
	SnapshotSessionBrowserProfilesForScope(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) []SharedSessionBrowserProfileState
	ResolveSessionBrowserProfileState(sessionID string, desired SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool)
	ResolveSessionBrowserProfileStatus(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, fallback BrowserProfileStatusResult) (BrowserProfileStatusResult, bool)
	SyncSessionBrowserProfileObservation(sessionID string, state SharedSessionBrowserProfileState, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool)
	SyncSessionBrowserProfileStatusObservation(sessionID string, selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult, observedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool)
	SyncSessionBrowserProfileLifecycleObservation(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string, observedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool)
	SyncSessionBrowserProfileStatusResolution(sessionID string, selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult, observedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool)
	SyncSessionBrowserProfileLifecycleResolution(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string, observedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool)
	SyncSessionBrowserProfilesObservation(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result BrowserProfilesResult, observedAt time.Time, reconnectWindow time.Duration) []SharedSessionBrowserProfileState
	SyncSessionBrowserProfilesResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result BrowserProfilesResult, observedAt time.Time, reconnectWindow time.Duration) []SharedSessionBrowserProfileState
	SyncSessionBrowserStatusAndProfilesObservations(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, status *BrowserProfileStatusResult, statusObservedAt time.Time, profiles *BrowserProfilesResult, profilesObservedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState)
	SyncSessionBrowserExecutionObservations(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, profile string, profiles *BrowserProfilesResult, profilesObservedAt time.Time, result BrowserProfileStatusResult, resultObservedAt time.Time, decision string, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState)
	SyncSessionBrowserStatusAndProfilesResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, status *BrowserProfileStatusResult, statusObservedAt time.Time, profiles *BrowserProfilesResult, profilesObservedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState)
	SyncSessionBrowserExecutionResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, profile string, profiles *BrowserProfilesResult, profilesObservedAt time.Time, result BrowserProfileStatusResult, resultObservedAt time.Time, decision string, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState)
	SelectBrowserProfile(sessionID string, selection SharedSessionBrowserProfileSelection)
	ClearSelectedBrowserProfile(sessionID string, runtimeTarget string)
	SelectedBrowserProfile(sessionID string, runtimeTarget string) (SharedSessionBrowserProfileSelection, bool)
	ClearSessionBrowserProfiles(sessionID string, filter SharedSessionBrowserProfileState) int
}

// BrowserBackend is the shared backend contract used by browser tools.
type BrowserBackend interface {
	Open(context.Context, BrowserOpenRequest) (BrowserOpenResult, error)
	Navigate(context.Context, BrowserNavigateRequest) (BrowserNavigateResult, error)
	Tabs(context.Context, BrowserTabsRequest) (BrowserTabsResult, error)
	Extract(context.Context, BrowserExtractRequest) (BrowserExtractResult, error)
	Snapshot(context.Context, BrowserSnapshotRequest) (BrowserSnapshotResult, error)
	Screenshot(context.Context, BrowserScreenshotRequest) (BrowserScreenshotResult, error)
	Click(context.Context, BrowserClickRequest) (BrowserClickResult, error)
	Type(context.Context, BrowserTypeRequest) (BrowserTypeResult, error)
	Eval(context.Context, BrowserEvalRequest) (BrowserEvalResult, error)
}

// BrowserArtifactActionBackend allows a backend to expose file-producing browser actions
// such as downloads or page-to-PDF rendering without requiring separate legacy tools.
type BrowserArtifactActionBackend interface {
	Download(context.Context, BrowserDownloadRequest) (BrowserDownloadResult, error)
	WaitDownload(context.Context, BrowserWaitDownloadRequest) (BrowserWaitDownloadResult, error)
	SavePDF(context.Context, BrowserSavePDFRequest) (BrowserSavePDFResult, error)
	SaveHTML(context.Context, BrowserSaveHTMLRequest) (BrowserSaveHTMLResult, error)
}

// BrowserConsoleActionBackend allows a backend to read console messages from the
// current page/runtime target without mutating browser state.
type BrowserConsoleActionBackend interface {
	Console(context.Context, BrowserConsoleRequest) (BrowserConsoleResult, error)
}

// BrowserRequestsActionBackend allows a backend to read recent network requests
// for the current page/runtime target without mutating browser state.
type BrowserRequestsActionBackend interface {
	Requests(context.Context, BrowserRequestsRequest) (BrowserRequestsResult, error)
}

// BrowserResponseBodyActionBackend allows a backend to read a matching response
// body for the current page/runtime target without mutating browser state.
type BrowserResponseBodyActionBackend interface {
	ResponseBody(context.Context, BrowserResponseBodyRequest) (BrowserResponseBodyResult, error)
}

// BrowserErrorsActionBackend allows a backend to read recent page/runtime errors
// without mutating browser state.
type BrowserErrorsActionBackend interface {
	Errors(context.Context, BrowserErrorsRequest) (BrowserErrorsResult, error)
}

// BrowserCookiesActionBackend allows a backend to read current page/browser cookies
// without mutating browser state.
type BrowserCookiesActionBackend interface {
	Cookies(context.Context, BrowserCookiesRequest) (BrowserCookiesResult, error)
}

// BrowserCookiesMutatingActionBackend allows a backend to update or clear cookies
// for the current page/runtime target.
type BrowserCookiesMutatingActionBackend interface {
	SetCookies(context.Context, BrowserCookiesSetRequest) (BrowserCookiesResult, error)
	ClearCookies(context.Context, BrowserCookiesClearRequest) (BrowserCookiesResult, error)
}

// BrowserTraceActionBackend allows a backend to arm or stop a page trace for the
// current runtime target and optionally return a trace artifact.
type BrowserTraceActionBackend interface {
	Trace(context.Context, BrowserTraceRequest) (BrowserTraceResult, error)
}

// BrowserStorageActionBackend allows a backend to inspect browser storage state
// without mutating browser state.
type BrowserStorageActionBackend interface {
	Storage(context.Context, BrowserStorageRequest) (BrowserStorageResult, error)
}

// BrowserStorageMutatingActionBackend allows a backend to update or clear storage
// entries for the current page/runtime target.
type BrowserStorageMutatingActionBackend interface {
	SetStorage(context.Context, BrowserStorageSetRequest) (BrowserStorageResult, error)
	ClearStorage(context.Context, BrowserStorageClearRequest) (BrowserStorageResult, error)
}

// BrowserHighlightActionBackend allows a backend to visually highlight a ref or selector
// so the caller can confirm which DOM target a follow-up action would hit.
type BrowserHighlightActionBackend interface {
	Highlight(context.Context, BrowserHighlightRequest) (BrowserHighlightResult, error)
}

// BrowserRuntimeControlBackend allows a backend to expose managed-browser
// profile lifecycle controls such as status/start/stop/list profiles.
type BrowserRuntimeControlBackend interface {
	RuntimeStatus(context.Context, BrowserProfileStatusRequest) (BrowserProfileStatusResult, error)
	RuntimeStart(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error)
	RuntimeStop(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error)
	RuntimeProfiles(context.Context, BrowserProfilesRequest) (BrowserProfilesResult, error)
}

// BrowserRuntimeProfileManagementBackend allows a backend to create or delete
// managed browser profiles.
type BrowserRuntimeProfileManagementBackend interface {
	RuntimeCreateProfile(context.Context, BrowserProfileCreateRequest) (BrowserProfileStatusResult, error)
	RuntimeDeleteProfile(context.Context, BrowserProfileDeleteRequest) (BrowserProfileStatusResult, error)
}

// BrowserDialogActionBackend allows a backend to arm the next modal dialog
// (alert/confirm/prompt) before a click or keypress triggers it.
type BrowserDialogActionBackend interface {
	Dialog(context.Context, BrowserDialogRequest) (BrowserDialogResult, error)
}

// BrowserUploadActionBackend allows a backend to arm a file chooser or set files on an input.
type BrowserUploadActionBackend interface {
	Upload(context.Context, BrowserUploadRequest) (BrowserUploadResult, error)
}

// BrowserPressActionBackend allows a backend to synthesize a key press on the
// current runtime target.
type BrowserPressActionBackend interface {
	Press(context.Context, BrowserPressRequest) (BrowserPressResult, error)
}

// BrowserHoverActionBackend allows a backend to hover a DOM target before a
// follow-up interaction.
type BrowserHoverActionBackend interface {
	Hover(context.Context, BrowserHoverRequest) (BrowserHoverResult, error)
}

// BrowserDragActionBackend allows a backend to drag from one DOM target to
// another.
type BrowserDragActionBackend interface {
	Drag(context.Context, BrowserDragRequest) (BrowserDragResult, error)
}

// BrowserSelectActionBackend allows a backend to select values in a form field.
type BrowserSelectActionBackend interface {
	Select(context.Context, BrowserSelectRequest) (BrowserSelectResult, error)
}

// BrowserFillActionBackend allows a backend to fill multiple form fields in one
// action.
type BrowserFillActionBackend interface {
	Fill(context.Context, BrowserFillRequest) (BrowserFillResult, error)
}

// BrowserResizeActionBackend allows a backend to resize the active page or
// viewport.
type BrowserResizeActionBackend interface {
	Resize(context.Context, BrowserResizeRequest) (BrowserResizeResult, error)
}

// BrowserOfflineActionBackend allows a backend to toggle offline mode for the
// current page/runtime target.
type BrowserOfflineActionBackend interface {
	SetOffline(context.Context, BrowserOfflineRequest) (BrowserOfflineResult, error)
}

// BrowserHeadersActionBackend allows a backend to replace or clear extra HTTP
// headers for the current page/runtime target.
type BrowserHeadersActionBackend interface {
	SetHeaders(context.Context, BrowserHeadersRequest) (BrowserHeadersResult, error)
}

// BrowserCredentialsActionBackend allows a backend to set or clear HTTP auth
// credentials for the current page/runtime target.
type BrowserCredentialsActionBackend interface {
	SetCredentials(context.Context, BrowserCredentialsRequest) (BrowserCredentialsResult, error)
}

// BrowserGeolocationActionBackend allows a backend to set or clear geolocation
// overrides for the current page/runtime target.
type BrowserGeolocationActionBackend interface {
	SetGeolocation(context.Context, BrowserGeolocationRequest) (BrowserGeolocationResult, error)
}

// BrowserMediaActionBackend allows a backend to update media emulation such as
// prefers-color-scheme for the current page/runtime target.
type BrowserMediaActionBackend interface {
	SetMedia(context.Context, BrowserMediaRequest) (BrowserMediaResult, error)
}

// BrowserTimezoneActionBackend allows a backend to set or clear timezone
// emulation for the current page/runtime target.
type BrowserTimezoneActionBackend interface {
	SetTimezone(context.Context, BrowserTimezoneRequest) (BrowserTimezoneResult, error)
}

// BrowserLocaleActionBackend allows a backend to set or clear locale/language
// emulation for the current page/runtime target.
type BrowserLocaleActionBackend interface {
	SetLocale(context.Context, BrowserLocaleRequest) (BrowserLocaleResult, error)
}

// BrowserDeviceActionBackend allows a backend to apply or clear device
// emulation for the current page/runtime target.
type BrowserDeviceActionBackend interface {
	SetDevice(context.Context, BrowserDeviceRequest) (BrowserDeviceResult, error)
}

// BrowserCapabilities describes which browser actions and tool surfaces a backend supports.
type BrowserCapabilities struct {
	RuntimeStatus       bool
	RuntimeWorkbench    bool
	RuntimePrepare      bool
	RuntimeCoordinate   bool
	RuntimeStart        bool
	RuntimeRestart      bool
	RuntimeStop         bool
	RuntimeCreate       bool
	RuntimeDelete       bool
	RuntimeSelect       bool
	RuntimeClear        bool
	RuntimeClearSession bool
	RuntimeSyncSession  bool
	RuntimeSelectTarget bool
	RuntimeClearTarget  bool
	RuntimeList         bool
	RuntimeSessions     bool
	Open                bool
	Navigate            bool
	Tabs                bool
	Extract             bool
	Snapshot            bool
	Screenshot          bool
	Console             bool
	Requests            bool
	ResponseBody        bool
	Errors              bool
	Cookies             bool
	CookiesSet          bool
	CookiesClear        bool
	Storage             bool
	StorageSet          bool
	StorageClear        bool
	Offline             bool
	Headers             bool
	Credentials         bool
	Geolocation         bool
	Media               bool
	Timezone            bool
	Locale              bool
	Device              bool
	Highlight           bool
	TraceStart          bool
	TraceStop           bool
	Download            bool
	WaitDownload        bool
	SavePDF             bool
	SaveHTML            bool
	Dialog              bool
	Upload              bool
	Press               bool
	Hover               bool
	Drag                bool
	Select              bool
	Fill                bool
	Resize              bool
	Click               bool
	TypeText            bool
	Evaluate            bool
	Wait                bool
}

// BrowserCapabilityProvider allows a backend to advertise a narrower capability surface.
type BrowserCapabilityProvider interface {
	BrowserCapabilities() BrowserCapabilities
}

// BrowserRuntimeInfo describes the runtime location and profile for browser tools.
type BrowserRuntimeInfo struct {
	Backend string
	Profile string
	Target  string
}

// BrowserRuntimeInfoProvider allows a backend to advertise runtime placement details.
type BrowserRuntimeInfoProvider interface {
	BrowserRuntimeInfo() BrowserRuntimeInfo
}

// BrowserRuntimeRouteResolver allows a backend to validate or rewrite a requested runtime route.
type BrowserRuntimeRouteResolver interface {
	ResolveBrowserRuntimeRoute(BrowserRuntimeInfo) (BrowserRuntimeInfo, error)
}

// BrowserRuntimeBackendRouter allows browser tools to select a concrete backend for a runtime route.
type BrowserRuntimeBackendRouter interface {
	ResolveBrowserBackend(BrowserRuntimeInfo) (BrowserBackend, BrowserRuntimeInfo, error)
}

func (c BrowserCapabilities) SupportsTool(name string) bool {
	switch normalizeBrowserToolSurfaceToken(name) {
	case "browser":
		return true
	case "browser_runtime":
		return true
	case "browser_act":
		return c.SupportsAnyActKind()
	default:
		actKind := BrowserCompatActKindForToolName(name)
		return actKind != "" && c.SupportsActKind(actKind)
	}
}

func (c BrowserCapabilities) SupportsActKind(kind string) bool {
	switch kind {
	case "open":
		return c.Open
	case "navigate":
		return c.Navigate
	case "extract":
		return c.Extract
	case "snapshot":
		return c.Snapshot
	case "screenshot":
		return c.Screenshot
	case "console":
		return c.Console
	case "requests":
		return c.Requests
	case "response_body":
		return c.ResponseBody
	case "errors":
		return c.Errors
	case "cookies":
		return c.Cookies
	case "cookies_set":
		return c.CookiesSet
	case "cookies_clear":
		return c.CookiesClear
	case "storage":
		return c.Storage
	case "storage_set":
		return c.StorageSet
	case "storage_clear":
		return c.StorageClear
	case "offline":
		return c.Offline
	case "headers":
		return c.Headers
	case "credentials":
		return c.Credentials
	case "geolocation":
		return c.Geolocation
	case "media":
		return c.Media
	case "timezone":
		return c.Timezone
	case "locale":
		return c.Locale
	case "device":
		return c.Device
	case "highlight":
		return c.Highlight
	case "trace_start":
		return c.TraceStart
	case "trace_stop":
		return c.TraceStop
	case "download":
		return c.Download
	case "wait_download":
		return c.WaitDownload
	case "save_pdf":
		return c.SavePDF
	case "save_html":
		return c.SaveHTML
	case "dialog":
		return c.Dialog
	case "upload":
		return c.Upload
	case "press":
		return c.Press
	case "hover":
		return c.Hover
	case "drag":
		return c.Drag
	case "select":
		return c.Select
	case "fill":
		return c.Fill
	case "resize":
		return c.Resize
	case "click":
		return c.Click
	case "type":
		return c.TypeText
	case "evaluate":
		return c.Evaluate
	case "wait":
		return c.Wait
	case "list_tabs", "focus_tab", "close_tab":
		return c.Tabs
	default:
		return false
	}
}

func (c BrowserCapabilities) SupportsAnyActKind() bool {
	return c.Open || c.Navigate || c.Extract || c.Snapshot || c.Screenshot || c.Console || c.Requests || c.ResponseBody || c.Errors || c.Cookies || c.CookiesSet || c.CookiesClear || c.Storage || c.StorageSet || c.StorageClear || c.Offline || c.Headers || c.Credentials || c.Geolocation || c.Media || c.Timezone || c.Locale || c.Device || c.Highlight || c.TraceStart || c.TraceStop || c.Download || c.WaitDownload || c.SavePDF || c.SaveHTML || c.Dialog || c.Upload || c.Press || c.Hover || c.Drag || c.Select || c.Fill || c.Resize || c.Click || c.TypeText || c.Evaluate || c.Wait || c.Tabs
}

func (c BrowserCapabilities) SupportedToolNames() []string {
	names := []string{"browser_runtime"}
	for _, name := range BrowserCompatToolNames() {
		if c.SupportsActKind(BrowserCompatActKindForToolName(name)) {
			names = append(names, name)
		}
	}
	if c.SupportsAnyActKind() {
		names = append(names, "browser_act")
	}
	return names
}

func (c BrowserCapabilities) SupportedRuntimeActions() []string {
	actions := []string{"status"}
	if c.RuntimeWorkbench {
		actions = append(actions, "workbench")
	}
	if c.RuntimePrepare {
		actions = append(actions, "prepare")
	}
	if c.RuntimeCoordinate {
		actions = append(actions, "coordinate")
	}
	if c.RuntimeStart {
		actions = append(actions, "start")
	}
	if c.RuntimeRestart {
		actions = append(actions, "restart")
		actions = append(actions, "refresh")
	}
	if c.RuntimeStop {
		actions = append(actions, "stop")
	}
	if c.RuntimeCreate {
		actions = append(actions, "create_profile")
	}
	if c.RuntimeDelete {
		actions = append(actions, "delete_profile")
	}
	if c.RuntimeSelect {
		actions = append(actions, "select_profile")
	}
	if c.RuntimeClear {
		actions = append(actions, "clear_profile")
	}
	if c.RuntimeClearSession {
		actions = append(actions, "clear_session")
	}
	if c.RuntimeSyncSession {
		actions = append(actions, "sync_session")
	}
	if c.RuntimeSelectTarget {
		actions = append(actions, "select_target")
	}
	if c.RuntimeClearTarget {
		actions = append(actions, "clear_target")
	}
	if c.RuntimeList {
		actions = append(actions, "profiles")
	}
	if c.RuntimeSessions {
		actions = append(actions, "sessions")
	}
	return actions
}

func (c BrowserCapabilities) SupportedActKinds() []string {
	kinds := make([]string, 0, 12)
	if c.Open {
		kinds = append(kinds, "open")
	}
	if c.Navigate {
		kinds = append(kinds, "navigate")
	}
	if c.Extract {
		kinds = append(kinds, "extract")
	}
	if c.Snapshot {
		kinds = append(kinds, "snapshot")
	}
	if c.Screenshot {
		kinds = append(kinds, "screenshot")
	}
	if c.Console {
		kinds = append(kinds, "console")
	}
	if c.Requests {
		kinds = append(kinds, "requests")
	}
	if c.ResponseBody {
		kinds = append(kinds, "response_body")
	}
	if c.Errors {
		kinds = append(kinds, "errors")
	}
	if c.Cookies {
		kinds = append(kinds, "cookies")
	}
	if c.CookiesSet {
		kinds = append(kinds, "cookies_set")
	}
	if c.CookiesClear {
		kinds = append(kinds, "cookies_clear")
	}
	if c.Storage {
		kinds = append(kinds, "storage")
	}
	if c.StorageSet {
		kinds = append(kinds, "storage_set")
	}
	if c.StorageClear {
		kinds = append(kinds, "storage_clear")
	}
	if c.Offline {
		kinds = append(kinds, "offline")
	}
	if c.Headers {
		kinds = append(kinds, "headers")
	}
	if c.Credentials {
		kinds = append(kinds, "credentials")
	}
	if c.Geolocation {
		kinds = append(kinds, "geolocation")
	}
	if c.Media {
		kinds = append(kinds, "media")
	}
	if c.Timezone {
		kinds = append(kinds, "timezone")
	}
	if c.Locale {
		kinds = append(kinds, "locale")
	}
	if c.Device {
		kinds = append(kinds, "device")
	}
	if c.Highlight {
		kinds = append(kinds, "highlight")
	}
	if c.TraceStart {
		kinds = append(kinds, "trace_start")
	}
	if c.TraceStop {
		kinds = append(kinds, "trace_stop")
	}
	if c.Download {
		kinds = append(kinds, "download")
	}
	if c.WaitDownload {
		kinds = append(kinds, "wait_download")
	}
	if c.SavePDF {
		kinds = append(kinds, "save_pdf")
	}
	if c.SaveHTML {
		kinds = append(kinds, "save_html")
	}
	if c.Dialog {
		kinds = append(kinds, "dialog")
	}
	if c.Upload {
		kinds = append(kinds, "upload")
	}
	if c.Press {
		kinds = append(kinds, "press")
	}
	if c.Hover {
		kinds = append(kinds, "hover")
	}
	if c.Drag {
		kinds = append(kinds, "drag")
	}
	if c.Select {
		kinds = append(kinds, "select")
	}
	if c.Fill {
		kinds = append(kinds, "fill")
	}
	if c.Resize {
		kinds = append(kinds, "resize")
	}
	if c.Click {
		kinds = append(kinds, "click")
	}
	if c.TypeText {
		kinds = append(kinds, "type")
	}
	if c.Evaluate {
		kinds = append(kinds, "evaluate")
	}
	if c.Wait {
		kinds = append(kinds, "wait")
	}
	if c.Tabs {
		kinds = append(kinds, "list_tabs", "focus_tab", "close_tab")
	}
	return kinds
}

type BrowserOpenRequest struct {
	URL        string
	BrowserApp string
	WaitMs     int
}

type BrowserOpenResult struct {
	Backend    string
	BrowserApp string
	Status     string
	Note       string
}

type BrowserNavigateRequest struct {
	URL              string
	BrowserApp       string
	WaitMs           int
	TabIndex         int
	Force            bool
	ExplicitTargetID string
	PriorSelection   *BrowserSessionTargetSelection
}

type BrowserNavigateResult struct {
	Backend    string
	BrowserApp string
	FinalURL   string
	Title      string
	Status     string
	Note       string
}

type BrowserExtractRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	MaxChars          int
	TabIndex          int
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

type BrowserExtractResult struct {
	Backend     string
	BrowserApp  string
	Title       string
	Content     string
	FinalURL    string
	ContentType string
}

type BrowserSnapshotRequest struct {
	URL               string                                       `json:"url,omitempty"`
	BrowserApp        string                                       `json:"browser_app,omitempty"`
	WaitMs            int                                          `json:"wait_ms,omitempty"`
	MaxChars          int                                          `json:"max_chars,omitempty"`
	MaxElements       int                                          `json:"max_elements,omitempty"`
	TabIndex          int                                          `json:"tab_index,omitempty"`
	PreferredTargetID string                                       `json:"preferred_target_id,omitempty"`
	Actor             string                                       `json:"actor,omitempty"`
	Force             bool                                         `json:"force,omitempty"`
	Review            SharedSessionBrowserPendingTargetReviewState `json:"review,omitempty"`
	Format            string                                       `json:"format,omitempty"`
	Mode              string                                       `json:"mode,omitempty"`
	Refs              string                                       `json:"refs,omitempty"`
	Interactive       bool                                         `json:"interactive,omitempty"`
	Compact           bool                                         `json:"compact,omitempty"`
	Depth             int                                          `json:"depth,omitempty"`
	Selector          string                                       `json:"selector,omitempty"`
	Frame             string                                       `json:"frame,omitempty"`
}

type BrowserSnapshotResult struct {
	Backend     string                   `json:"backend,omitempty"`
	BrowserApp  string                   `json:"browser_app,omitempty"`
	FinalURL    string                   `json:"final_url,omitempty"`
	Title       string                   `json:"title,omitempty"`
	Snapshot    string                   `json:"snapshot,omitempty"`
	Elements    []BrowserSnapshotElement `json:"elements,omitempty"`
	Truncated   bool                     `json:"truncated,omitempty"`
	Format      string                   `json:"format,omitempty"`
	Mode        string                   `json:"mode,omitempty"`
	Refs        string                   `json:"refs,omitempty"`
	Interactive bool                     `json:"interactive,omitempty"`
	Compact     bool                     `json:"compact,omitempty"`
	Depth       int                      `json:"depth,omitempty"`
	Selector    string                   `json:"selector,omitempty"`
	Frame       string                   `json:"frame,omitempty"`
	Note        string                   `json:"note,omitempty"`
}

type BrowserSnapshotElement struct {
	Index         int    `json:"index"`
	Role          string `json:"role,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Label         string `json:"label,omitempty"`
	Ref           string `json:"ref,omitempty"`
	Selector      string `json:"selector,omitempty"`
	SelectorIndex int    `json:"selector_index,omitempty"`
	FramePath     string `json:"frame_path,omitempty"`
	Type          string `json:"type,omitempty"`
	Href          string `json:"href,omitempty"`
	Placeholder   string `json:"placeholder,omitempty"`
}

type BrowserElementHint struct {
	Selector       string                    `json:"selector,omitempty"`
	SelectorIndex  int                       `json:"selector_index,omitempty"`
	FramePath      string                    `json:"frame_path,omitempty"`
	NativeRef      string                    `json:"native_ref,omitempty"`
	Role           string                    `json:"role,omitempty"`
	Tag            string                    `json:"tag,omitempty"`
	Label          string                    `json:"label,omitempty"`
	Type           string                    `json:"type,omitempty"`
	Href           string                    `json:"href,omitempty"`
	Placeholder    string                    `json:"placeholder,omitempty"`
	PageURL        string                    `json:"page_url,omitempty"`
	PageOrigin     string                    `json:"page_origin,omitempty"`
	PagePath       string                    `json:"page_path,omitempty"`
	PageTitle      string                    `json:"page_title,omitempty"`
	TabIndex       int                       `json:"tab_index,omitempty"`
	LocatorOrder   []string                  `json:"locator_order,omitempty"`
	LocatorPlan    []BrowserLocatorCandidate `json:"locator_plan,omitempty"`
	ResolutionMode string                    `json:"resolution_mode,omitempty"`
}

type BrowserLocatorCandidate struct {
	Kind          string `json:"kind,omitempty"`
	Selector      string `json:"selector,omitempty"`
	SelectorIndex int    `json:"selector_index,omitempty"`
	FramePath     string `json:"frame_path,omitempty"`
	NativeRef     string `json:"native_ref,omitempty"`
	Role          string `json:"role,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Label         string `json:"label,omitempty"`
	Type          string `json:"type,omitempty"`
	Href          string `json:"href,omitempty"`
	Placeholder   string `json:"placeholder,omitempty"`
	PageURL       string `json:"page_url,omitempty"`
	PageOrigin    string `json:"page_origin,omitempty"`
	PagePath      string `json:"page_path,omitempty"`
	PageTitle     string `json:"page_title,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
}

type BrowserElementResolverRequest struct {
	ResolutionMode string                    `json:"resolution_mode,omitempty"`
	PrimaryKind    string                    `json:"primary_kind,omitempty"`
	ElementRef     string                    `json:"element_ref,omitempty"`
	Selector       string                    `json:"selector,omitempty"`
	SelectorIndex  int                       `json:"selector_index,omitempty"`
	FramePath      string                    `json:"frame_path,omitempty"`
	LocatorOrder   []string                  `json:"locator_order,omitempty"`
	LocatorPlan    []BrowserLocatorCandidate `json:"locator_plan,omitempty"`
	MatchPlan      []BrowserLocatorCandidate `json:"match_plan,omitempty"`
	PageBinding    *BrowserLocatorCandidate  `json:"page_binding,omitempty"`
}

// BrowserElementResolverOutcome captures the machine-readable result of
// resolving a BrowserElementResolverRequest against a backend-specific lookup
// adapter.
type BrowserElementResolverOutcome struct {
	Status                        string   `json:"status,omitempty"`
	ResolutionMode                string   `json:"resolution_mode,omitempty"`
	PrimaryKind                   string   `json:"primary_kind,omitempty"`
	AttemptCount                  int      `json:"attempt_count,omitempty"`
	MatchedKind                   string   `json:"matched_kind,omitempty"`
	MatchedIndex                  int      `json:"matched_index,omitempty"`
	MatchedCandidateKind          string   `json:"matched_candidate_kind,omitempty"`
	ResolvedFramePath             string   `json:"resolved_frame_path,omitempty"`
	ResolvedSelectorIndex         int      `json:"resolved_selector_index,omitempty"`
	NativeRefRebound              bool     `json:"native_ref_rebound,omitempty"`
	FallbackFromKind              string   `json:"fallback_from_kind,omitempty"`
	FallbackFromIndex             int      `json:"fallback_from_index,omitempty"`
	FallbackFromBlockedBy         string   `json:"fallback_from_blocked_by,omitempty"`
	FallbackFromAmbiguityClass    string   `json:"fallback_from_ambiguity_class,omitempty"`
	FallbackFromCandidateStrength string   `json:"fallback_from_candidate_strength,omitempty"`
	FallbackFromManualRetryHint   string   `json:"fallback_from_manual_retry_hint,omitempty"`
	FallbackFromSpecificityFields []string `json:"fallback_from_specificity_fields,omitempty"`
	CandidateKind                 string   `json:"candidate_kind,omitempty"`
	CandidateStrength             string   `json:"candidate_strength,omitempty"`
	AmbiguityClass                string   `json:"ambiguity_class,omitempty"`
	RetryDisposition              string   `json:"retry_disposition,omitempty"`
	ManualRetryHint               string   `json:"manual_retry_hint,omitempty"`
	NextStepAlias                 string   `json:"next_step_alias,omitempty"`
	BlockedBy                     string   `json:"blocked_by,omitempty"`
	LocatorCount                  int      `json:"locator_count,omitempty"`
	CandidateCount                int      `json:"candidate_count,omitempty"`
	PreferredOrdinal              int      `json:"preferred_ordinal,omitempty"`
	SpecificityFields             []string `json:"specificity_fields,omitempty"`
	RecoveryAction                string   `json:"recovery_action,omitempty"`
	Note                          string   `json:"note,omitempty"`
}

type BrowserScreenshotRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	OutputPath        string
	ElementRef        string
	ElementHint       *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver   *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector          string
	FullPage          bool
	TabIndex          int
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

type BrowserScreenshotResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	Path            string                         `json:"path,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	CaptureScope    string                         `json:"capture_scope,omitempty"`
	CaptureWidth    int                            `json:"capture_width,omitempty"`
	CaptureHeight   int                            `json:"capture_height,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

// UnmarshalJSON accepts the stable snake_case proxy contract and the legacy
// PascalCase shape emitted by older browserd screenshot responses.
func (r *BrowserScreenshotResult) UnmarshalJSON(data []byte) error {
	type screenshotResult BrowserScreenshotResult
	var current screenshotResult
	if err := json.Unmarshal(data, &current); err != nil {
		return err
	}
	var legacy struct {
		Backend         string                         `json:"Backend,omitempty"`
		BrowserApp      string                         `json:"BrowserApp,omitempty"`
		Path            string                         `json:"Path,omitempty"`
		FinalURL        string                         `json:"FinalURL,omitempty"`
		Title           string                         `json:"Title,omitempty"`
		CaptureScope    string                         `json:"CaptureScope,omitempty"`
		CaptureWidth    int                            `json:"CaptureWidth,omitempty"`
		CaptureHeight   int                            `json:"CaptureHeight,omitempty"`
		Status          string                         `json:"Status,omitempty"`
		Note            string                         `json:"Note,omitempty"`
		ResolverOutcome *BrowserElementResolverOutcome `json:"ResolverOutcome,omitempty"`
		Actionability   *BrowserActionabilityReport    `json:"Actionability,omitempty"`
		FailureEvidence *BrowserActionFailureEvidence  `json:"FailureEvidence,omitempty"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if current.Backend == "" {
		current.Backend = legacy.Backend
	}
	if current.BrowserApp == "" {
		current.BrowserApp = legacy.BrowserApp
	}
	if current.Path == "" {
		current.Path = legacy.Path
	}
	if current.FinalURL == "" {
		current.FinalURL = legacy.FinalURL
	}
	if current.Title == "" {
		current.Title = legacy.Title
	}
	if current.CaptureScope == "" {
		current.CaptureScope = legacy.CaptureScope
	}
	if current.CaptureWidth == 0 {
		current.CaptureWidth = legacy.CaptureWidth
	}
	if current.CaptureHeight == 0 {
		current.CaptureHeight = legacy.CaptureHeight
	}
	if current.Status == "" {
		current.Status = legacy.Status
	}
	if current.Note == "" {
		current.Note = legacy.Note
	}
	if current.ResolverOutcome == nil {
		current.ResolverOutcome = legacy.ResolverOutcome
	}
	if current.Actionability == nil {
		current.Actionability = legacy.Actionability
	}
	if current.FailureEvidence == nil {
		current.FailureEvidence = legacy.FailureEvidence
	}
	*r = BrowserScreenshotResult(current)
	return nil
}

type BrowserConsoleRequest struct {
	BrowserApp string `json:"browser_app,omitempty"`
	WaitMs     int    `json:"wait_ms,omitempty"`
	TabIndex   int    `json:"tab_index,omitempty"`
	Level      string `json:"level,omitempty"`
}

type BrowserConsoleMessage struct {
	Level  string `json:"level,omitempty"`
	Text   string `json:"text,omitempty"`
	Source string `json:"source,omitempty"`
}

type BrowserConsoleResult struct {
	Backend    string                  `json:"backend,omitempty"`
	BrowserApp string                  `json:"browser_app,omitempty"`
	FinalURL   string                  `json:"final_url,omitempty"`
	Title      string                  `json:"title,omitempty"`
	Messages   []BrowserConsoleMessage `json:"messages,omitempty"`
	Note       string                  `json:"note,omitempty"`
}

type BrowserRequestsRequest struct {
	BrowserApp string `json:"browser_app,omitempty"`
	WaitMs     int    `json:"wait_ms,omitempty"`
	TabIndex   int    `json:"tab_index,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Clear      bool   `json:"clear,omitempty"`
}

type BrowserRequestEntry struct {
	Method       string `json:"method,omitempty"`
	URL          string `json:"url,omitempty"`
	Status       int    `json:"status,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	Failure      string `json:"failure,omitempty"`
}

type BrowserResponseBodyRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Filter            string `json:"filter,omitempty"`
	URL               string `json:"url,omitempty"`
	MaxChars          int    `json:"max_chars,omitempty"`
}

type BrowserResponseBodyResult struct {
	Backend     string `json:"backend,omitempty"`
	BrowserApp  string `json:"browser_app,omitempty"`
	FinalURL    string `json:"final_url,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Method      string `json:"method,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Body        string `json:"body,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
	Note        string `json:"note,omitempty"`
}

type BrowserRequestsResult struct {
	Backend    string                `json:"backend,omitempty"`
	BrowserApp string                `json:"browser_app,omitempty"`
	FinalURL   string                `json:"final_url,omitempty"`
	Title      string                `json:"title,omitempty"`
	Requests   []BrowserRequestEntry `json:"requests,omitempty"`
	Status     string                `json:"status,omitempty"`
	Note       string                `json:"note,omitempty"`
}

type BrowserErrorsRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserErrorEntry struct {
	Message           string   `json:"message,omitempty"`
	Source            string   `json:"source,omitempty"`
	Event             string   `json:"event,omitempty"`
	Category          string   `json:"category,omitempty"`
	Severity          string   `json:"severity,omitempty"`
	ResolverStatus    string   `json:"resolver_status,omitempty"`
	CandidateKind     string   `json:"candidate_kind,omitempty"`
	CandidateStrength string   `json:"candidate_strength,omitempty"`
	AmbiguityClass    string   `json:"ambiguity_class,omitempty"`
	RetryDisposition  string   `json:"retry_disposition,omitempty"`
	ManualRetryHint   string   `json:"manual_retry_hint,omitempty"`
	NextStepAlias     string   `json:"next_step_alias,omitempty"`
	BlockedBy         string   `json:"blocked_by,omitempty"`
	LocatorCount      int      `json:"locator_count,omitempty"`
	CandidateCount    int      `json:"candidate_count,omitempty"`
	PreferredOrdinal  int      `json:"preferred_ordinal,omitempty"`
	SpecificityFields []string `json:"specificity_fields,omitempty"`
	RecoveryAction    string   `json:"recovery_action,omitempty"`
	TargetID          string   `json:"target_id,omitempty"`
	TabIndex          int      `json:"tab_index,omitempty"`
	URL               string   `json:"url,omitempty"`
	OccurredAt        int64    `json:"occurred_at,omitempty"`
}

type BrowserSessionHealthSummary struct {
	State                       string `json:"state,omitempty"`
	Reason                      string `json:"reason,omitempty"`
	RecoveryAction              string `json:"recovery_action,omitempty"`
	ReconnectHint               string `json:"reconnect_hint,omitempty"`
	DisconnectCount             int    `json:"disconnect_count,omitempty"`
	DisconnectBurstCount        int    `json:"disconnect_burst_count,omitempty"`
	DisconnectBurstWindowMs     int    `json:"disconnect_burst_window_ms,omitempty"`
	CooldownRemainingMs         int    `json:"cooldown_remaining_ms,omitempty"`
	RetryBackoffRemainingMs     int    `json:"retry_backoff_remaining_ms,omitempty"`
	RestartAttemptCount         int    `json:"restart_attempt_count,omitempty"`
	RestartFailureCount         int    `json:"restart_failure_count,omitempty"`
	LastDisconnectUnixMilli     int64  `json:"last_disconnect_unix_milli,omitempty"`
	LastReconnectUnixMilli      int64  `json:"last_reconnect_unix_milli,omitempty"`
	LastRestartAttemptUnixMilli int64  `json:"last_restart_attempt_unix_milli,omitempty"`
	LastRestartResult           string `json:"last_restart_result,omitempty"`
	LastRestartError            string `json:"last_restart_error,omitempty"`
	RecommendedBackoffMs        int    `json:"recommended_backoff_ms,omitempty"`
	EventsBuffered              int    `json:"events_buffered,omitempty"`
	PopupCount                  int    `json:"popup_count,omitempty"`
	LastEvent                   string `json:"last_event,omitempty"`
	LastDialogAction            string `json:"last_dialog_action,omitempty"`
	LastDialogType              string `json:"last_dialog_type,omitempty"`
	LastDialogSource            string `json:"last_dialog_source,omitempty"`
	LastDownloadFile            string `json:"last_download_filename,omitempty"`
	LastDownloadOutput          string `json:"last_download_output,omitempty"`
	CurrentTargetID             string `json:"current_target_id,omitempty"`
	CurrentTabIndex             int    `json:"current_tab_index,omitempty"`
	PendingDownloads            int    `json:"pending_downloads,omitempty"`
	PendingDialogAction         string `json:"pending_dialog_action,omitempty"`
	StaleTargetResolverStatus   string `json:"stale_target_resolver_status,omitempty"`
	StaleTargetBlockedBy        string `json:"stale_target_blocked_by,omitempty"`
	LastUpdatedUnixMilli        int64  `json:"last_updated_unix_milli,omitempty"`
}

type BrowserErrorsResult struct {
	Backend       string                       `json:"backend,omitempty"`
	BrowserApp    string                       `json:"browser_app,omitempty"`
	FinalURL      string                       `json:"final_url,omitempty"`
	Title         string                       `json:"title,omitempty"`
	Errors        []BrowserErrorEntry          `json:"errors,omitempty"`
	SessionHealth *BrowserSessionHealthSummary `json:"session_health,omitempty"`
	Status        string                       `json:"status,omitempty"`
	Note          string                       `json:"note,omitempty"`
}

type BrowserCookiesRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Filter            string `json:"filter,omitempty"`
}

type BrowserCookiesSetRequest struct {
	BrowserApp        string               `json:"browser_app,omitempty"`
	WaitMs            int                  `json:"wait_ms,omitempty"`
	TabIndex          int                  `json:"tab_index,omitempty"`
	PreferredTargetID string               `json:"preferred_target_id,omitempty"`
	URL               string               `json:"url,omitempty"`
	Cookies           []BrowserCookieEntry `json:"cookies,omitempty"`
}

type BrowserCookiesClearRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	URL               string `json:"url,omitempty"`
	Filter            string `json:"filter,omitempty"`
	Name              string `json:"name,omitempty"`
}

type BrowserCookieEntry struct {
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
	SameSite string `json:"same_site,omitempty"`
	HTTPOnly bool   `json:"http_only,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
}

type BrowserCookiesResult struct {
	Backend    string               `json:"backend,omitempty"`
	BrowserApp string               `json:"browser_app,omitempty"`
	FinalURL   string               `json:"final_url,omitempty"`
	Title      string               `json:"title,omitempty"`
	Status     string               `json:"status,omitempty"`
	Cookies    []BrowserCookieEntry `json:"cookies,omitempty"`
	Note       string               `json:"note,omitempty"`
}

type BrowserTraceRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Action            string `json:"action,omitempty"`
	OutputPath        string `json:"output_path,omitempty"`
}

type BrowserTraceResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	Path       string `json:"path,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserStorageRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Kind              string `json:"kind,omitempty"`
	Filter            string `json:"filter,omitempty"`
}

type BrowserStorageSetRequest struct {
	BrowserApp        string                `json:"browser_app,omitempty"`
	WaitMs            int                   `json:"wait_ms,omitempty"`
	TabIndex          int                   `json:"tab_index,omitempty"`
	PreferredTargetID string                `json:"preferred_target_id,omitempty"`
	Kind              string                `json:"kind,omitempty"`
	Entries           []BrowserStorageEntry `json:"entries,omitempty"`
}

type BrowserStorageClearRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Kind              string `json:"kind,omitempty"`
	Filter            string `json:"filter,omitempty"`
	Key               string `json:"key,omitempty"`
}

type BrowserStorageEntry struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type BrowserStorageResult struct {
	Backend    string                `json:"backend,omitempty"`
	BrowserApp string                `json:"browser_app,omitempty"`
	FinalURL   string                `json:"final_url,omitempty"`
	Title      string                `json:"title,omitempty"`
	Kind       string                `json:"kind,omitempty"`
	Status     string                `json:"status,omitempty"`
	Entries    []BrowserStorageEntry `json:"entries,omitempty"`
	Note       string                `json:"note,omitempty"`
}

type BrowserOfflineRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Enabled           bool   `json:"enabled,omitempty"`
}

type BrowserOfflineResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Enabled    bool   `json:"enabled,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserHeadersRequest struct {
	BrowserApp        string            `json:"browser_app,omitempty"`
	WaitMs            int               `json:"wait_ms,omitempty"`
	TabIndex          int               `json:"tab_index,omitempty"`
	PreferredTargetID string            `json:"preferred_target_id,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	Clear             bool              `json:"clear,omitempty"`
}

type BrowserHeadersResult struct {
	Backend     string   `json:"backend,omitempty"`
	BrowserApp  string   `json:"browser_app,omitempty"`
	FinalURL    string   `json:"final_url,omitempty"`
	Title       string   `json:"title,omitempty"`
	Status      string   `json:"status,omitempty"`
	HeaderNames []string `json:"header_names,omitempty"`
	HeaderCount int      `json:"header_count,omitempty"`
	Note        string   `json:"note,omitempty"`
}

type BrowserCredentialsRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserCredentialsResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Origin     string `json:"origin,omitempty"`
	Username   string `json:"username,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserGeolocationRequest struct {
	BrowserApp        string  `json:"browser_app,omitempty"`
	WaitMs            int     `json:"wait_ms,omitempty"`
	TabIndex          int     `json:"tab_index,omitempty"`
	PreferredTargetID string  `json:"preferred_target_id,omitempty"`
	Latitude          float64 `json:"latitude,omitempty"`
	Longitude         float64 `json:"longitude,omitempty"`
	Accuracy          float64 `json:"accuracy,omitempty"`
	Origin            string  `json:"origin,omitempty"`
	Clear             bool    `json:"clear,omitempty"`
}

type BrowserGeolocationResult struct {
	Backend    string  `json:"backend,omitempty"`
	BrowserApp string  `json:"browser_app,omitempty"`
	FinalURL   string  `json:"final_url,omitempty"`
	Title      string  `json:"title,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Accuracy   float64 `json:"accuracy,omitempty"`
	Origin     string  `json:"origin,omitempty"`
	Status     string  `json:"status,omitempty"`
	Note       string  `json:"note,omitempty"`
}

type BrowserMediaRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Media             string `json:"media,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserMediaResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Media      string `json:"media,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserTimezoneRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Timezone          string `json:"timezone,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserTimezoneResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserLocaleRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Locale            string `json:"locale,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserLocaleResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Locale     string `json:"locale,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserDeviceRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	Device            string `json:"device,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	Clear             bool   `json:"clear,omitempty"`
}

type BrowserDeviceResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Device     string `json:"device,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserHighlightRequest struct {
	BrowserApp      string                         `json:"browser_app,omitempty"`
	WaitMs          int                            `json:"wait_ms,omitempty"`
	TabIndex        int                            `json:"tab_index,omitempty"`
	Ref             string                         `json:"ref,omitempty"`
	ElementHint     *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector        string                         `json:"selector,omitempty"`
}

type BrowserHighlightResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Ref             string                         `json:"ref,omitempty"`
	Selector        string                         `json:"selector,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserProfileStatusRequest struct {
	Profile string `json:"profile,omitempty"`
}

type BrowserProfileLifecycleRequest struct {
	Profile string `json:"profile,omitempty"`
}

type BrowserProfileCreateRequest struct {
	Profile    string `json:"profile,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	Color      string `json:"color,omitempty"`
	CopyFrom   string `json:"copy_from,omitempty"`
}

type BrowserProfileDeleteRequest struct {
	Profile string `json:"profile,omitempty"`
	Force   bool   `json:"force,omitempty"`
}

type BrowserProfileStatusResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	Profile         string                         `json:"profile,omitempty"`
	StateRoot       string                         `json:"state_root,omitempty"`
	ProfilesRoot    string                         `json:"profiles_root,omitempty"`
	ArtifactsRoot   string                         `json:"artifacts_root,omitempty"`
	LogsRoot        string                         `json:"logs_root,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Running         bool                           `json:"running,omitempty"`
	Connected       bool                           `json:"connected,omitempty"`
	PlaywrightCache *BrowserPlaywrightCacheSummary `json:"playwright_cache,omitempty"`
	SessionHealth   *BrowserSessionHealthSummary   `json:"session_health,omitempty"`
	Note            string                         `json:"note,omitempty"`
}

type BrowserPlaywrightCacheSummary struct {
	HostOS                      string   `json:"host_os,omitempty"`
	HostArch                    string   `json:"host_arch,omitempty"`
	NodeVersion                 string   `json:"node_version,omitempty"`
	PlaywrightPackage           string   `json:"playwright_package,omitempty"`
	PlaywrightPackageVersion    string   `json:"playwright_package_version,omitempty"`
	RuntimeSummaryGeneration    string   `json:"runtime_summary_generation,omitempty"`
	RuntimeBaselineReady        bool     `json:"runtime_baseline_ready,omitempty"`
	RuntimeBaselineBlockReason  string   `json:"runtime_baseline_block_reason,omitempty"`
	Path                        string   `json:"path,omitempty"`
	Source                      string   `json:"source,omitempty"`
	Pinned                      bool     `json:"pinned,omitempty"`
	BundleGeneration            string   `json:"bundle_generation,omitempty"`
	DependencyGeneration        string   `json:"dependency_generation,omitempty"`
	BrowserRevision             string   `json:"browser_revision,omitempty"`
	DeliveryGeneration          string   `json:"delivery_generation,omitempty"`
	TargetDeliveryGeneration    string   `json:"target_delivery_generation,omitempty"`
	LastReadyDelivery           string   `json:"last_ready_delivery_generation,omitempty"`
	RetainedDeliveries          []string `json:"retained_delivery_generations,omitempty"`
	LastEvictedDelivery         string   `json:"last_evicted_delivery_generation,omitempty"`
	LastDeliverySwitchUnix      int64    `json:"last_delivery_generation_switch_unix_milli,omitempty"`
	RetainedDeliveryRevision    string   `json:"retained_delivery_browser_revision,omitempty"`
	RetainedDeliveryReady       bool     `json:"retained_delivery_cache_ready,omitempty"`
	RetainedFallbackDelivery    string   `json:"retained_fallback_delivery_generation,omitempty"`
	RetainedFallbackPayload     bool     `json:"retained_fallback_payload_ready,omitempty"`
	RetainedFallbackPayloadBR   string   `json:"retained_fallback_payload_block_reason,omitempty"`
	RetainedFallbackPayloadSrc  string   `json:"retained_fallback_payload_source,omitempty"`
	RetainedFallbackPayloadDirs []string `json:"retained_fallback_payload_dirs,omitempty"`
	RetainedFallbackLaunch      bool     `json:"retained_fallback_launch_ready,omitempty"`
	RetainedFallbackBlock       string   `json:"retained_fallback_launch_block_reason,omitempty"`
	SelectedLaunchDelivery      string   `json:"selected_launch_delivery_generation,omitempty"`
	SelectedLaunchSource        string   `json:"selected_launch_source,omitempty"`
	SelectedLaunchReady         bool     `json:"selected_launch_ready,omitempty"`
	SelectedLaunchBlockReason   string   `json:"selected_launch_block_reason,omitempty"`
	SelectedLaunchRevision      string   `json:"selected_launch_browser_revision,omitempty"`
	SelectedLaunchPayloadSrc    string   `json:"selected_launch_payload_source,omitempty"`
	SelectedLaunchPayloadDirs   []string `json:"selected_launch_payload_dirs,omitempty"`
	SelectedLaunchPayloadReady  bool     `json:"selected_launch_payload_ready,omitempty"`
	SelectedLaunchPayloadBR     string   `json:"selected_launch_payload_block_reason,omitempty"`
	SelectedLaunchExecutable    string   `json:"selected_launch_executable_path,omitempty"`
	SelectedLaunchExecutableOK  bool     `json:"selected_launch_executable_ready,omitempty"`
	SelectedLaunchExecutableBR  string   `json:"selected_launch_executable_block_reason,omitempty"`
	DeliveryTransitionPending   bool     `json:"delivery_transition_pending,omitempty"`
	DeliveryTransitionStage     string   `json:"delivery_transition_stage,omitempty"`
	LaunchReady                 bool     `json:"launch_ready,omitempty"`
	LaunchBlockReason           string   `json:"launch_block_reason,omitempty"`
	BundleReady                 bool     `json:"bundle_ready,omitempty"`
	DeliveryReady               bool     `json:"delivery_ready,omitempty"`
	PolicyVersion               string   `json:"policy_version,omitempty"`
	RetentionMode               string   `json:"retention_mode,omitempty"`
	RetainedDirs                []string `json:"retained_dirs,omitempty"`
	LastGCPrunedDirCount        int      `json:"last_gc_pruned_dir_count,omitempty"`
	BootstrapState              string   `json:"bootstrap_state,omitempty"`
	BootstrapErrorCode          string   `json:"bootstrap_error_code,omitempty"`
	NodeModulesReady            bool     `json:"node_modules_ready,omitempty"`
	BrowserReady                bool     `json:"browser_ready,omitempty"`
}

type BrowserProfileInfo struct {
	Profile    string `json:"profile,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	Status     string `json:"status,omitempty"`
	Running    bool   `json:"running,omitempty"`
	Connected  bool   `json:"connected,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserProfilesRequest struct {
	Profile string `json:"profile,omitempty"`
}

type BrowserProfilesResult struct {
	Backend        string               `json:"backend,omitempty"`
	DefaultProfile string               `json:"default_profile,omitempty"`
	Profiles       []BrowserProfileInfo `json:"profiles,omitempty"`
	Note           string               `json:"note,omitempty"`
}

type BrowserDownloadRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	OutputPath        string
	TabIndex          int
	PreferredTargetID string
	ReviewDecision    string
	ReviewReady       bool
	Note              string
}

type BrowserDownloadResult struct {
	Backend     string
	BrowserApp  string
	Path        string
	FinalURL    string
	Title       string
	ContentType string
	Download    *BrowserDownloadMetadata `json:"download,omitempty"`
	Note        string
}

type BrowserDownloadMetadata struct {
	Mode              string `json:"mode,omitempty"`
	SuggestedFilename string `json:"suggested_filename,omitempty"`
	OutputMode        string `json:"output_mode,omitempty"`
	BackendPath       string `json:"backend_path,omitempty"`
	ByteSize          int64  `json:"byte_size,omitempty"`
}

type BrowserWaitDownloadRequest struct {
	BrowserApp               string
	WaitMs                   int
	OutputPath               string
	AllowRecentDownloadReuse bool
	TabIndex                 int
	PreferredTargetID        string
	ReviewDecision           string
	ReviewReady              bool
	Note                     string
}

type BrowserWaitDownloadResult struct {
	Backend     string
	BrowserApp  string
	Path        string
	FinalURL    string
	Title       string
	ContentType string
	Download    *BrowserDownloadMetadata `json:"download,omitempty"`
	Note        string
}

type BrowserSavePDFRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	OutputPath        string
	TabIndex          int
	PreferredTargetID string
	ReviewDecision    string
	ReviewReady       bool
	Note              string
	Landscape         bool
	PrintBackground   bool
}

type BrowserSavePDFResult struct {
	Backend    string
	BrowserApp string
	Path       string
	FinalURL   string
	Title      string
	PageCount  int
	Note       string
}

type BrowserSaveHTMLRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	OutputPath        string
	TabIndex          int
	PreferredTargetID string
	ReviewDecision    string
	ReviewReady       bool
	Note              string
}

type BrowserSaveHTMLResult struct {
	Backend    string
	BrowserApp string
	Path       string
	FinalURL   string
	Title      string
	Note       string
}

type BrowserDialogRequest struct {
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
	ReviewDecision    string `json:"review_decision,omitempty"`
	ReviewReady       bool   `json:"review_ready,omitempty"`
	Note              string `json:"note,omitempty"`
	Action            string `json:"action,omitempty"`
	PromptText        string `json:"prompt_text,omitempty"`
}

type BrowserDialogResult struct {
	Backend    string
	BrowserApp string
	FinalURL   string
	Title      string
	Status     string
	Dialog     *BrowserDialogMetadata `json:"dialog,omitempty"`
	Note       string
}

type BrowserDialogMetadata struct {
	Action      string `json:"action,omitempty"`
	PromptState string `json:"prompt_state,omitempty"`
}

type BrowserUploadRequest struct {
	BrowserApp        string                         `json:"browser_app,omitempty"`
	WaitMs            int                            `json:"wait_ms,omitempty"`
	TabIndex          int                            `json:"tab_index,omitempty"`
	PreferredTargetID string                         `json:"preferred_target_id,omitempty"`
	ReviewDecision    string                         `json:"review_decision,omitempty"`
	ReviewReady       bool                           `json:"review_ready,omitempty"`
	Note              string                         `json:"note,omitempty"`
	Paths             []string                       `json:"paths,omitempty"`
	Ref               string                         `json:"ref,omitempty"`
	InputRef          string                         `json:"input_ref,omitempty"`
	ElementHint       *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver   *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector          string                         `json:"selector,omitempty"`
}

type BrowserUploadResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserPressRequest struct {
	URL               string `json:"url,omitempty"`
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	PostWaitMs        int    `json:"post_wait_ms,omitempty"`
	Key               string `json:"key,omitempty"`
	DelayMs           int    `json:"delay_ms,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
}

type BrowserPressResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Key        string `json:"key,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserHoverRequest struct {
	URL             string                         `json:"url,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	WaitMs          int                            `json:"wait_ms,omitempty"`
	PostWaitMs      int                            `json:"post_wait_ms,omitempty"`
	ElementRef      string                         `json:"element_ref,omitempty"`
	ElementHint     *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector        string                         `json:"selector,omitempty"`
	TabIndex        int                            `json:"tab_index,omitempty"`
}

type BrowserHoverResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserDragRequest struct {
	URL           string                         `json:"url,omitempty"`
	BrowserApp    string                         `json:"browser_app,omitempty"`
	WaitMs        int                            `json:"wait_ms,omitempty"`
	PostWaitMs    int                            `json:"post_wait_ms,omitempty"`
	StartRef      string                         `json:"start_ref,omitempty"`
	StartHint     *BrowserElementHint            `json:"start_hint,omitempty"`
	StartResolver *BrowserElementResolverRequest `json:"start_resolver,omitempty"`
	StartSelector string                         `json:"start_selector,omitempty"`
	EndRef        string                         `json:"end_ref,omitempty"`
	EndHint       *BrowserElementHint            `json:"end_hint,omitempty"`
	EndResolver   *BrowserElementResolverRequest `json:"end_resolver,omitempty"`
	EndSelector   string                         `json:"end_selector,omitempty"`
	TabIndex      int                            `json:"tab_index,omitempty"`
}

type BrowserDragResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserSelectRequest struct {
	URL             string                         `json:"url,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	WaitMs          int                            `json:"wait_ms,omitempty"`
	PostWaitMs      int                            `json:"post_wait_ms,omitempty"`
	ElementRef      string                         `json:"element_ref,omitempty"`
	ElementHint     *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector        string                         `json:"selector,omitempty"`
	Values          []string                       `json:"values,omitempty"`
	TabIndex        int                            `json:"tab_index,omitempty"`
}

type BrowserSelectResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Values          []string                       `json:"values,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserFillField struct {
	Ref      string                         `json:"ref,omitempty"`
	Hint     *BrowserElementHint            `json:"hint,omitempty"`
	Resolver *BrowserElementResolverRequest `json:"resolver,omitempty"`
	Selector string                         `json:"selector,omitempty"`
	Type     string                         `json:"type,omitempty"`
	Value    string                         `json:"value,omitempty"`
	Values   []string                       `json:"values,omitempty"`
}

type BrowserFillRequest struct {
	URL        string             `json:"url,omitempty"`
	BrowserApp string             `json:"browser_app,omitempty"`
	WaitMs     int                `json:"wait_ms,omitempty"`
	PostWaitMs int                `json:"post_wait_ms,omitempty"`
	Fields     []BrowserFillField `json:"fields,omitempty"`
	Submit     bool               `json:"submit,omitempty"`
	TabIndex   int                `json:"tab_index,omitempty"`
}

type BrowserFillResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	FieldCount      int                            `json:"field_count,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Submitted       bool                           `json:"submitted,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserResizeRequest struct {
	URL               string `json:"url,omitempty"`
	BrowserApp        string `json:"browser_app,omitempty"`
	WaitMs            int    `json:"wait_ms,omitempty"`
	PostWaitMs        int    `json:"post_wait_ms,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	TabIndex          int    `json:"tab_index,omitempty"`
	PreferredTargetID string `json:"preferred_target_id,omitempty"`
}

type BrowserResizeResult struct {
	Backend    string `json:"backend,omitempty"`
	BrowserApp string `json:"browser_app,omitempty"`
	FinalURL   string `json:"final_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Status     string `json:"status,omitempty"`
	Note       string `json:"note,omitempty"`
}

type BrowserClickRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	PostWaitMs        int
	ElementRef        string
	ElementHint       *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver   *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector          string
	TabIndex          int
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

type BrowserClickResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserTypeRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	PostWaitMs        int
	ElementRef        string
	ElementHint       *BrowserElementHint            `json:"element_hint,omitempty"`
	ElementResolver   *BrowserElementResolverRequest `json:"element_resolver,omitempty"`
	Selector          string
	Text              string
	Submit            bool
	TabIndex          int
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

type BrowserTypeResult struct {
	Backend         string                         `json:"backend,omitempty"`
	BrowserApp      string                         `json:"browser_app,omitempty"`
	FinalURL        string                         `json:"final_url,omitempty"`
	Title           string                         `json:"title,omitempty"`
	Value           string                         `json:"value,omitempty"`
	Status          string                         `json:"status,omitempty"`
	Submitted       bool                           `json:"submitted,omitempty"`
	Note            string                         `json:"note,omitempty"`
	ResolverOutcome *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability   *BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence *BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
}

type BrowserEvalRequest struct {
	URL               string
	BrowserApp        string
	WaitMs            int
	Script            string
	MaxChars          int
	TabIndex          int
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

type BrowserEvalResult struct {
	Backend    string
	BrowserApp string
	FinalURL   string
	Title      string
	Result     string
	Status     string
	Note       string
}

type BrowserActRequest struct {
	Kind       string
	Action     string
	URL        string
	Target     string
	BrowserApp string
	WaitMs     int
	PostWaitMs int
	TabIndex   int
	OutputPath string
	Ref        string
	Selector   string
	Text       string
	Submit     bool
	Script     string
	MaxChars   int
}

type BrowserActResult struct {
	Kind                              string
	Action                            string
	Backend                           string
	BrowserApp                        string
	Target                            string
	TargetID                          string
	Profile                           string
	RuntimeTarget                     string
	FinalURL                          string
	Title                             string
	Content                           string
	ContentType                       string
	Snapshot                          string
	SnapshotFormat                    string
	SnapshotMode                      string
	SnapshotRefs                      string
	SnapshotInteractive               bool
	SnapshotCompact                   bool
	SnapshotDepth                     int
	SnapshotFrame                     string
	Elements                          []BrowserSnapshotElement
	Messages                          []BrowserConsoleMessage
	Requests                          []BrowserRequestEntry
	RequestURL                        string
	RequestMethod                     string
	ResponseStatusCode                int
	Errors                            []BrowserErrorEntry
	Cookies                           []BrowserCookieEntry
	StorageKind                       string
	Storage                           []BrowserStorageEntry
	Offline                           bool
	HeaderNames                       []string
	HeaderCount                       int
	CredentialsOrigin                 string
	CredentialsUsername               string
	GeolocationOrigin                 string
	Latitude                          float64
	Longitude                         float64
	Accuracy                          float64
	Media                             string
	Timezone                          string
	Locale                            string
	Device                            string
	HighlightRef                      string
	Path                              string
	Paths                             []string
	FilesTouched                      []string
	Bytes                             int64
	CaptureScope                      string
	CaptureWidth                      int
	CaptureHeight                     int
	Key                               string
	Ref                               string
	Selector                          string
	ResolverOutcome                   *BrowserElementResolverOutcome
	Actionability                     *BrowserActionabilityReport
	FailureEvidence                   *BrowserActionFailureEvidence
	ResolvedViaFallback               bool
	ResolverFallbackKind              string
	ResolverFallbackIndex             *int
	ResolverFallbackCandidateStrength string
	ResolverFallbackBlockedBy         string
	ResolverFallbackAmbiguityClass    string
	ResolverFallbackManualRetryHint   string
	ResolverFallbackSpecificityFields []string
	ResolverBlockedBy                 string
	ResolverAmbiguityClass            string
	ResolverCandidateKind             string
	ResolverCandidateStrength         string
	ResolverRetryDisposition          string
	ResolverManualRetryHint           string
	ResolverNextStepAlias             string
	BrowserLocalPlanner               *BrowserLocalPlannerResultSummary
	RecoveryAction                    string
	Result                            string
	Value                             string
	Values                            []string
	FieldCount                        int
	Width                             int
	Height                            int
	Status                            string
	Force                             bool
	ReviewDecision                    string
	ReviewReady                       bool
	Submitted                         bool
	Truncated                         bool
	TabIndex                          int
	Tabs                              []BrowserTab
	ActiveIndex                       int
	RememberDecision                  string
	RememberReady                     bool
	SessionTargetSelection            *BrowserSessionTargetSelection
	Note                              string
}

type BrowserSessionTargetSelection struct {
	ID            string `json:"id,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Source        string `json:"source,omitempty"`
}

type BrowserTab struct {
	Index    int    `json:"index"`
	Title    string `json:"title,omitempty"`
	URL      string `json:"url,omitempty"`
	TargetID string `json:"target_id,omitempty"`
	Active   bool   `json:"active"`
}

type BrowserTabsRequest struct {
	BrowserApp             string
	Action                 string
	TabIndex               int
	WaitMs                 int
	Force                  bool
	RememberTarget         bool
	Review                 SharedSessionBrowserPendingTargetReviewState
	Actor                  string
	ExplicitTargetID       string
	PriorSelection         *BrowserSessionTargetSelection
	PriorActiveTargetID    string
	PriorRequestedTargetID string
}

type BrowserTabsResult struct {
	Backend     string
	BrowserApp  string
	Action      string
	Status      string
	Tabs        []BrowserTab
	ActiveIndex int
	Note        string
}
