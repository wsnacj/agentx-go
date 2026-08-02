package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

type fakeBrowserBackend struct {
	openReqs            []BrowserOpenRequest
	navigateReqs        []BrowserNavigateRequest
	tabsReqs            []BrowserTabsRequest
	extractReqs         []BrowserExtractRequest
	snapshotReqs        []BrowserSnapshotRequest
	screenshotReqs      []BrowserScreenshotRequest
	downloadReqs        []BrowserDownloadRequest
	waitDownloadReqs    []BrowserWaitDownloadRequest
	savePDFReqs         []BrowserSavePDFRequest
	saveHTMLReqs        []BrowserSaveHTMLRequest
	consoleReqs         []BrowserConsoleRequest
	requestsReqs        []BrowserRequestsRequest
	responseBodyReqs    []BrowserResponseBodyRequest
	errorsReqs          []BrowserErrorsRequest
	headersReqs         []BrowserHeadersRequest
	cookiesReqs         []BrowserCookiesRequest
	cookiesSetReqs      []BrowserCookiesSetRequest
	cookiesClearReqs    []BrowserCookiesClearRequest
	storageReqs         []BrowserStorageRequest
	storageSetReqs      []BrowserStorageSetRequest
	storageClearReqs    []BrowserStorageClearRequest
	offlineReqs         []BrowserOfflineRequest
	credentialsReqs     []BrowserCredentialsRequest
	geolocationReqs     []BrowserGeolocationRequest
	mediaReqs           []BrowserMediaRequest
	deviceReqs          []BrowserDeviceRequest
	highlightReqs       []BrowserHighlightRequest
	runtimeStatusReqs   []BrowserProfileStatusRequest
	runtimeStartReqs    []BrowserProfileLifecycleRequest
	runtimeCreateReqs   []BrowserProfileCreateRequest
	runtimeDeleteReqs   []BrowserProfileDeleteRequest
	runtimeStopReqs     []BrowserProfileLifecycleRequest
	runtimeProfilesReqs []BrowserProfilesRequest
	traceReqs           []BrowserTraceRequest
	dialogReqs          []BrowserDialogRequest
	uploadReqs          []BrowserUploadRequest
	pressReqs           []BrowserPressRequest
	hoverReqs           []BrowserHoverRequest
	dragReqs            []BrowserDragRequest
	selectReqs          []BrowserSelectRequest
	fillReqs            []BrowserFillRequest
	resizeReqs          []BrowserResizeRequest
	clickReqs           []BrowserClickRequest
	typeReqs            []BrowserTypeRequest
	evalReqs            []BrowserEvalRequest

	openResult            BrowserOpenResult
	navigateResult        BrowserNavigateResult
	tabsResult            BrowserTabsResult
	extractResult         BrowserExtractResult
	snapshotResult        BrowserSnapshotResult
	screenshotResult      BrowserScreenshotResult
	downloadResult        BrowserDownloadResult
	waitDownloadResult    BrowserWaitDownloadResult
	savePDFResult         BrowserSavePDFResult
	saveHTMLResult        BrowserSaveHTMLResult
	consoleResult         BrowserConsoleResult
	requestsResult        BrowserRequestsResult
	responseBodyResult    BrowserResponseBodyResult
	errorsResult          BrowserErrorsResult
	headersResult         BrowserHeadersResult
	cookiesResult         BrowserCookiesResult
	cookiesSetResult      BrowserCookiesResult
	cookiesClearResult    BrowserCookiesResult
	storageResult         BrowserStorageResult
	storageSetResult      BrowserStorageResult
	storageClearResult    BrowserStorageResult
	offlineResult         BrowserOfflineResult
	credentialsResult     BrowserCredentialsResult
	geolocationResult     BrowserGeolocationResult
	mediaResult           BrowserMediaResult
	deviceResult          BrowserDeviceResult
	highlightResult       BrowserHighlightResult
	runtimeStatusResult   BrowserProfileStatusResult
	runtimeStartResult    BrowserProfileStatusResult
	runtimeCreateResult   BrowserProfileStatusResult
	runtimeDeleteResult   BrowserProfileStatusResult
	runtimeStopResult     BrowserProfileStatusResult
	runtimeProfilesResult BrowserProfilesResult
	traceResult           BrowserTraceResult
	dialogResult          BrowserDialogResult
	uploadResult          BrowserUploadResult
	pressResult           BrowserPressResult
	hoverResult           BrowserHoverResult
	dragResult            BrowserDragResult
	selectResult          BrowserSelectResult
	fillResult            BrowserFillResult
	resizeResult          BrowserResizeResult
	clickResult           BrowserClickResult
	typeResult            BrowserTypeResult
	evalResult            BrowserEvalResult

	openErr            error
	navigateErr        error
	tabsErr            error
	extractErr         error
	snapshotErr        error
	screenshotErr      error
	downloadErr        error
	waitDownloadErr    error
	savePDFErr         error
	saveHTMLErr        error
	consoleErr         error
	requestsErr        error
	responseBodyErr    error
	errorsErr          error
	headersErr         error
	cookiesErr         error
	cookiesSetErr      error
	cookiesClearErr    error
	storageErr         error
	storageSetErr      error
	storageClearErr    error
	offlineErr         error
	credentialsErr     error
	geolocationErr     error
	mediaErr           error
	deviceErr          error
	highlightErr       error
	runtimeStatusErr   error
	runtimeStartErr    error
	runtimeCreateErr   error
	runtimeDeleteErr   error
	runtimeStopErr     error
	runtimeProfilesErr error
	traceErr           error
	dialogErr          error
	uploadErr          error
	pressErr           error
	hoverErr           error
	dragErr            error
	selectErr          error
	fillErr            error
	resizeErr          error
	clickErr           error
	typeErr            error
	evalErr            error
}

func (f *fakeBrowserBackend) Open(_ context.Context, req BrowserOpenRequest) (BrowserOpenResult, error) {
	f.openReqs = append(f.openReqs, req)
	if f.openErr != nil {
		return BrowserOpenResult{}, f.openErr
	}
	return f.openResult, nil
}

func (f *fakeBrowserBackend) Navigate(_ context.Context, req BrowserNavigateRequest) (BrowserNavigateResult, error) {
	f.navigateReqs = append(f.navigateReqs, req)
	if f.navigateErr != nil {
		return BrowserNavigateResult{}, f.navigateErr
	}
	return f.navigateResult, nil
}

func (f *fakeBrowserBackend) Tabs(_ context.Context, req BrowserTabsRequest) (BrowserTabsResult, error) {
	f.tabsReqs = append(f.tabsReqs, req)
	if f.tabsErr != nil {
		return BrowserTabsResult{}, f.tabsErr
	}
	return f.tabsResult, nil
}

func (f *fakeBrowserBackend) Extract(_ context.Context, req BrowserExtractRequest) (BrowserExtractResult, error) {
	f.extractReqs = append(f.extractReqs, req)
	if f.extractErr != nil {
		return BrowserExtractResult{}, f.extractErr
	}
	return f.extractResult, nil
}

func (f *fakeBrowserBackend) Snapshot(_ context.Context, req BrowserSnapshotRequest) (BrowserSnapshotResult, error) {
	f.snapshotReqs = append(f.snapshotReqs, req)
	if f.snapshotErr != nil {
		return BrowserSnapshotResult{}, f.snapshotErr
	}
	return f.snapshotResult, nil
}

func (f *fakeBrowserBackend) Screenshot(_ context.Context, req BrowserScreenshotRequest) (BrowserScreenshotResult, error) {
	f.screenshotReqs = append(f.screenshotReqs, req)
	if f.screenshotErr != nil {
		return BrowserScreenshotResult{}, f.screenshotErr
	}
	if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
		return BrowserScreenshotResult{}, err
	}
	if err := os.WriteFile(req.OutputPath, []byte("fake-png"), 0o644); err != nil {
		return BrowserScreenshotResult{}, err
	}
	result := f.screenshotResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) Download(_ context.Context, req BrowserDownloadRequest) (BrowserDownloadResult, error) {
	f.downloadReqs = append(f.downloadReqs, req)
	if f.downloadErr != nil {
		return BrowserDownloadResult{}, f.downloadErr
	}
	if strings.TrimSpace(req.OutputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
			return BrowserDownloadResult{}, err
		}
		if err := os.WriteFile(req.OutputPath, []byte("fake-download"), 0o644); err != nil {
			return BrowserDownloadResult{}, err
		}
	}
	result := f.downloadResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) WaitDownload(_ context.Context, req BrowserWaitDownloadRequest) (BrowserWaitDownloadResult, error) {
	f.waitDownloadReqs = append(f.waitDownloadReqs, req)
	if f.waitDownloadErr != nil {
		return BrowserWaitDownloadResult{}, f.waitDownloadErr
	}
	if strings.TrimSpace(req.OutputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
			return BrowserWaitDownloadResult{}, err
		}
		if err := os.WriteFile(req.OutputPath, []byte("fake-wait-download"), 0o644); err != nil {
			return BrowserWaitDownloadResult{}, err
		}
	}
	result := f.waitDownloadResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) SavePDF(_ context.Context, req BrowserSavePDFRequest) (BrowserSavePDFResult, error) {
	f.savePDFReqs = append(f.savePDFReqs, req)
	if f.savePDFErr != nil {
		return BrowserSavePDFResult{}, f.savePDFErr
	}
	if strings.TrimSpace(req.OutputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
			return BrowserSavePDFResult{}, err
		}
		if err := os.WriteFile(req.OutputPath, []byte("%PDF-fake"), 0o644); err != nil {
			return BrowserSavePDFResult{}, err
		}
	}
	result := f.savePDFResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) SaveHTML(_ context.Context, req BrowserSaveHTMLRequest) (BrowserSaveHTMLResult, error) {
	f.saveHTMLReqs = append(f.saveHTMLReqs, req)
	if f.saveHTMLErr != nil {
		return BrowserSaveHTMLResult{}, f.saveHTMLErr
	}
	if strings.TrimSpace(req.OutputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
			return BrowserSaveHTMLResult{}, err
		}
		if err := os.WriteFile(req.OutputPath, []byte("<html><body>fake-html</body></html>"), 0o644); err != nil {
			return BrowserSaveHTMLResult{}, err
		}
	}
	result := f.saveHTMLResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) Console(_ context.Context, req BrowserConsoleRequest) (BrowserConsoleResult, error) {
	f.consoleReqs = append(f.consoleReqs, req)
	if f.consoleErr != nil {
		return BrowserConsoleResult{}, f.consoleErr
	}
	return f.consoleResult, nil
}

func (f *fakeBrowserBackend) Requests(_ context.Context, req BrowserRequestsRequest) (BrowserRequestsResult, error) {
	f.requestsReqs = append(f.requestsReqs, req)
	if f.requestsErr != nil {
		return BrowserRequestsResult{}, f.requestsErr
	}
	return f.requestsResult, nil
}

func (f *fakeBrowserBackend) ResponseBody(_ context.Context, req BrowserResponseBodyRequest) (BrowserResponseBodyResult, error) {
	f.responseBodyReqs = append(f.responseBodyReqs, req)
	if f.responseBodyErr != nil {
		return BrowserResponseBodyResult{}, f.responseBodyErr
	}
	return f.responseBodyResult, nil
}

func (f *fakeBrowserBackend) Errors(_ context.Context, req BrowserErrorsRequest) (BrowserErrorsResult, error) {
	f.errorsReqs = append(f.errorsReqs, req)
	if f.errorsErr != nil {
		return BrowserErrorsResult{}, f.errorsErr
	}
	return f.errorsResult, nil
}

func (f *fakeBrowserBackend) SetHeaders(_ context.Context, req BrowserHeadersRequest) (BrowserHeadersResult, error) {
	f.headersReqs = append(f.headersReqs, req)
	if f.headersErr != nil {
		return BrowserHeadersResult{}, f.headersErr
	}
	return f.headersResult, nil
}

func (f *fakeBrowserBackend) Cookies(_ context.Context, req BrowserCookiesRequest) (BrowserCookiesResult, error) {
	f.cookiesReqs = append(f.cookiesReqs, req)
	if f.cookiesErr != nil {
		return BrowserCookiesResult{}, f.cookiesErr
	}
	return f.cookiesResult, nil
}

func (f *fakeBrowserBackend) SetCookies(_ context.Context, req BrowserCookiesSetRequest) (BrowserCookiesResult, error) {
	f.cookiesSetReqs = append(f.cookiesSetReqs, req)
	if f.cookiesSetErr != nil {
		return BrowserCookiesResult{}, f.cookiesSetErr
	}
	return f.cookiesSetResult, nil
}

func (f *fakeBrowserBackend) ClearCookies(_ context.Context, req BrowserCookiesClearRequest) (BrowserCookiesResult, error) {
	f.cookiesClearReqs = append(f.cookiesClearReqs, req)
	if f.cookiesClearErr != nil {
		return BrowserCookiesResult{}, f.cookiesClearErr
	}
	return f.cookiesClearResult, nil
}

func (f *fakeBrowserBackend) Trace(_ context.Context, req BrowserTraceRequest) (BrowserTraceResult, error) {
	f.traceReqs = append(f.traceReqs, req)
	if f.traceErr != nil {
		return BrowserTraceResult{}, f.traceErr
	}
	if strings.TrimSpace(req.OutputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(req.OutputPath), 0o755); err != nil {
			return BrowserTraceResult{}, err
		}
		if err := os.WriteFile(req.OutputPath, []byte("fake-trace"), 0o644); err != nil {
			return BrowserTraceResult{}, err
		}
	}
	result := f.traceResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = req.OutputPath
	}
	return result, nil
}

func (f *fakeBrowserBackend) Storage(_ context.Context, req BrowserStorageRequest) (BrowserStorageResult, error) {
	f.storageReqs = append(f.storageReqs, req)
	if f.storageErr != nil {
		return BrowserStorageResult{}, f.storageErr
	}
	return f.storageResult, nil
}

func (f *fakeBrowserBackend) SetStorage(_ context.Context, req BrowserStorageSetRequest) (BrowserStorageResult, error) {
	f.storageSetReqs = append(f.storageSetReqs, req)
	if f.storageSetErr != nil {
		return BrowserStorageResult{}, f.storageSetErr
	}
	return f.storageSetResult, nil
}

func (f *fakeBrowserBackend) ClearStorage(_ context.Context, req BrowserStorageClearRequest) (BrowserStorageResult, error) {
	f.storageClearReqs = append(f.storageClearReqs, req)
	if f.storageClearErr != nil {
		return BrowserStorageResult{}, f.storageClearErr
	}
	return f.storageClearResult, nil
}

func (f *fakeBrowserBackend) SetOffline(_ context.Context, req BrowserOfflineRequest) (BrowserOfflineResult, error) {
	f.offlineReqs = append(f.offlineReqs, req)
	if f.offlineErr != nil {
		return BrowserOfflineResult{}, f.offlineErr
	}
	return f.offlineResult, nil
}

func (f *fakeBrowserBackend) SetCredentials(_ context.Context, req BrowserCredentialsRequest) (BrowserCredentialsResult, error) {
	f.credentialsReqs = append(f.credentialsReqs, req)
	if f.credentialsErr != nil {
		return BrowserCredentialsResult{}, f.credentialsErr
	}
	return f.credentialsResult, nil
}

func (f *fakeBrowserBackend) SetGeolocation(_ context.Context, req BrowserGeolocationRequest) (BrowserGeolocationResult, error) {
	f.geolocationReqs = append(f.geolocationReqs, req)
	if f.geolocationErr != nil {
		return BrowserGeolocationResult{}, f.geolocationErr
	}
	return f.geolocationResult, nil
}

func (f *fakeBrowserBackend) SetMedia(_ context.Context, req BrowserMediaRequest) (BrowserMediaResult, error) {
	f.mediaReqs = append(f.mediaReqs, req)
	if f.mediaErr != nil {
		return BrowserMediaResult{}, f.mediaErr
	}
	return f.mediaResult, nil
}

func (f *fakeBrowserBackend) SetDevice(_ context.Context, req BrowserDeviceRequest) (BrowserDeviceResult, error) {
	f.deviceReqs = append(f.deviceReqs, req)
	if f.deviceErr != nil {
		return BrowserDeviceResult{}, f.deviceErr
	}
	return f.deviceResult, nil
}

func (f *fakeBrowserBackend) Highlight(_ context.Context, req BrowserHighlightRequest) (BrowserHighlightResult, error) {
	f.highlightReqs = append(f.highlightReqs, req)
	if f.highlightErr != nil {
		return BrowserHighlightResult{}, f.highlightErr
	}
	return f.highlightResult, nil
}

func (f *fakeBrowserBackend) Upload(_ context.Context, req BrowserUploadRequest) (BrowserUploadResult, error) {
	f.uploadReqs = append(f.uploadReqs, req)
	if f.uploadErr != nil {
		return BrowserUploadResult{}, f.uploadErr
	}
	return f.uploadResult, nil
}

func (f *fakeBrowserBackend) Press(_ context.Context, req BrowserPressRequest) (BrowserPressResult, error) {
	f.pressReqs = append(f.pressReqs, req)
	if f.pressErr != nil {
		return BrowserPressResult{}, f.pressErr
	}
	return f.pressResult, nil
}

func (f *fakeBrowserBackend) Hover(_ context.Context, req BrowserHoverRequest) (BrowserHoverResult, error) {
	f.hoverReqs = append(f.hoverReqs, req)
	if f.hoverErr != nil {
		return BrowserHoverResult{}, f.hoverErr
	}
	return f.hoverResult, nil
}

func (f *fakeBrowserBackend) Drag(_ context.Context, req BrowserDragRequest) (BrowserDragResult, error) {
	f.dragReqs = append(f.dragReqs, req)
	if f.dragErr != nil {
		return BrowserDragResult{}, f.dragErr
	}
	return f.dragResult, nil
}

func (f *fakeBrowserBackend) Select(_ context.Context, req BrowserSelectRequest) (BrowserSelectResult, error) {
	f.selectReqs = append(f.selectReqs, req)
	if f.selectErr != nil {
		return BrowserSelectResult{}, f.selectErr
	}
	return f.selectResult, nil
}

func (f *fakeBrowserBackend) Fill(_ context.Context, req BrowserFillRequest) (BrowserFillResult, error) {
	f.fillReqs = append(f.fillReqs, req)
	if f.fillErr != nil {
		return BrowserFillResult{}, f.fillErr
	}
	return f.fillResult, nil
}

func (f *fakeBrowserBackend) Resize(_ context.Context, req BrowserResizeRequest) (BrowserResizeResult, error) {
	f.resizeReqs = append(f.resizeReqs, req)
	if f.resizeErr != nil {
		return BrowserResizeResult{}, f.resizeErr
	}
	return f.resizeResult, nil
}

func (f *fakeBrowserBackend) Click(_ context.Context, req BrowserClickRequest) (BrowserClickResult, error) {
	f.clickReqs = append(f.clickReqs, req)
	if f.clickErr != nil {
		return BrowserClickResult{}, f.clickErr
	}
	return f.clickResult, nil
}

func (f *fakeBrowserBackend) Type(_ context.Context, req BrowserTypeRequest) (BrowserTypeResult, error) {
	f.typeReqs = append(f.typeReqs, req)
	if f.typeErr != nil {
		return BrowserTypeResult{}, f.typeErr
	}
	return f.typeResult, nil
}

func (f *fakeBrowserBackend) Eval(_ context.Context, req BrowserEvalRequest) (BrowserEvalResult, error) {
	f.evalReqs = append(f.evalReqs, req)
	if f.evalErr != nil {
		return BrowserEvalResult{}, f.evalErr
	}
	return f.evalResult, nil
}

func (f *fakeBrowserBackend) Dialog(_ context.Context, req BrowserDialogRequest) (BrowserDialogResult, error) {
	f.dialogReqs = append(f.dialogReqs, req)
	if f.dialogErr != nil {
		return BrowserDialogResult{}, f.dialogErr
	}
	return f.dialogResult, nil
}

type capabilityBrowserBackend struct {
	*fakeBrowserBackend
	capabilities BrowserCapabilities
}

func (b *capabilityBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	return b.capabilities
}

type capabilityRemoteArtifactBrowserBackend struct {
	*remoteArtifactBrowserBackend
	capabilities BrowserCapabilities
}

func (b *capabilityRemoteArtifactBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	return b.capabilities
}

type runtimeInfoBrowserBackend struct {
	*fakeBrowserBackend
	runtimeInfo   BrowserRuntimeInfo
	routeSource   string
	routeEndpoint string
}

func (b *runtimeInfoBrowserBackend) BrowserRuntimeInfo() BrowserRuntimeInfo {
	return b.runtimeInfo
}

func (b *runtimeInfoBrowserBackend) browserDoctorRouteMetadata() browserDoctorRouteMetadata {
	if b == nil {
		return browserDoctorRouteMetadata{}
	}
	return browserDoctorRouteMetadata{
		Source:   strings.TrimSpace(b.routeSource),
		Endpoint: strings.TrimSpace(b.routeEndpoint),
	}
}

type remoteTargetGuardRuntimeInfoBrowserBackend struct {
	*runtimeInfoBrowserBackend
}

func (b *remoteTargetGuardRuntimeInfoBrowserBackend) BrowserRemoteTargetURLGuardEnabled() bool {
	return true
}

type runtimeControlBrowserBackend struct {
	*runtimeInfoBrowserBackend
}

type capabilityRuntimeControlBrowserBackend struct {
	*runtimeControlBrowserBackend
	capabilities BrowserCapabilities
}

func (b *capabilityRuntimeControlBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	return b.capabilities
}

type countingCapabilityRuntimeControlBrowserBackend struct {
	*runtimeControlBrowserBackend
	capabilities    BrowserCapabilities
	capabilityCalls int
}

func (b *countingCapabilityRuntimeControlBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	b.capabilityCalls++
	return b.capabilities
}

type routeResolverCapabilityRuntimeControlBrowserBackend struct {
	*capabilityRuntimeControlBrowserBackend
	resolve      func(BrowserRuntimeInfo) (BrowserRuntimeInfo, error)
	resolveCalls int
}

func (b *routeResolverCapabilityRuntimeControlBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	b.resolveCalls++
	if b.resolve != nil {
		return b.resolve(requested)
	}
	return requested, nil
}

func (b *routeResolverCapabilityRuntimeControlBrowserBackend) ResolveBrowserExecutionRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	info, err := b.ResolveBrowserRuntimeRoute(requested)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	info = firstBrowserRuntimeInfo(info, requested, b.runtimeInfo)
	return browserResolvedExecutionRoute{
		Backend:      b,
		RuntimeInfo:  info,
		Capabilities: b.capabilities,
	}, nil
}

func (b *routeResolverCapabilityRuntimeControlBrowserBackend) ResolveBrowserExecutionRouteForSession(_ context.Context, _ map[string]any, base BrowserRuntimeInfo, _ bool) (browserResolvedExecutionRoute, error) {
	return b.ResolveBrowserExecutionRoute(base)
}

func (b *runtimeControlBrowserBackend) RuntimeStatus(_ context.Context, req BrowserProfileStatusRequest) (BrowserProfileStatusResult, error) {
	b.runtimeStatusReqs = append(b.runtimeStatusReqs, req)
	if b.runtimeStatusErr != nil {
		return BrowserProfileStatusResult{}, b.runtimeStatusErr
	}
	return b.runtimeStatusResult, nil
}

func (b *runtimeControlBrowserBackend) RuntimeStart(_ context.Context, req BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	b.runtimeStartReqs = append(b.runtimeStartReqs, req)
	if b.runtimeStartErr != nil {
		return BrowserProfileStatusResult{}, b.runtimeStartErr
	}
	return b.runtimeStartResult, nil
}

func (b *runtimeControlBrowserBackend) RuntimeCreateProfile(_ context.Context, req BrowserProfileCreateRequest) (BrowserProfileStatusResult, error) {
	b.runtimeCreateReqs = append(b.runtimeCreateReqs, req)
	if b.runtimeCreateErr != nil {
		return BrowserProfileStatusResult{}, b.runtimeCreateErr
	}
	return b.runtimeCreateResult, nil
}

func (b *runtimeControlBrowserBackend) RuntimeDeleteProfile(_ context.Context, req BrowserProfileDeleteRequest) (BrowserProfileStatusResult, error) {
	b.runtimeDeleteReqs = append(b.runtimeDeleteReqs, req)
	if b.runtimeDeleteErr != nil {
		return BrowserProfileStatusResult{}, b.runtimeDeleteErr
	}
	return b.runtimeDeleteResult, nil
}

func (b *runtimeControlBrowserBackend) RuntimeStop(_ context.Context, req BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	b.runtimeStopReqs = append(b.runtimeStopReqs, req)
	if b.runtimeStopErr != nil {
		return BrowserProfileStatusResult{}, b.runtimeStopErr
	}
	return b.runtimeStopResult, nil
}

func (b *runtimeControlBrowserBackend) RuntimeProfiles(_ context.Context, req BrowserProfilesRequest) (BrowserProfilesResult, error) {
	b.runtimeProfilesReqs = append(b.runtimeProfilesReqs, req)
	if b.runtimeProfilesErr != nil {
		return BrowserProfilesResult{}, b.runtimeProfilesErr
	}
	return b.runtimeProfilesResult, nil
}

type remoteArtifactBrowserBackend struct {
	*fakeBrowserBackend
	remotePath    string
	remoteContent []byte
	resolveReqs   []browserArtifactResolveRequest
	resolveErr    error
}

func (b *remoteArtifactBrowserBackend) Screenshot(_ context.Context, req BrowserScreenshotRequest) (BrowserScreenshotResult, error) {
	b.screenshotReqs = append(b.screenshotReqs, req)
	if b.screenshotErr != nil {
		return BrowserScreenshotResult{}, b.screenshotErr
	}
	result := b.screenshotResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/screenshot.png")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) Download(_ context.Context, req BrowserDownloadRequest) (BrowserDownloadResult, error) {
	b.downloadReqs = append(b.downloadReqs, req)
	if b.downloadErr != nil {
		return BrowserDownloadResult{}, b.downloadErr
	}
	result := b.downloadResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/download.bin")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) WaitDownload(_ context.Context, req BrowserWaitDownloadRequest) (BrowserWaitDownloadResult, error) {
	b.waitDownloadReqs = append(b.waitDownloadReqs, req)
	if b.waitDownloadErr != nil {
		return BrowserWaitDownloadResult{}, b.waitDownloadErr
	}
	result := b.waitDownloadResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/download.bin")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) SavePDF(_ context.Context, req BrowserSavePDFRequest) (BrowserSavePDFResult, error) {
	b.savePDFReqs = append(b.savePDFReqs, req)
	if b.savePDFErr != nil {
		return BrowserSavePDFResult{}, b.savePDFErr
	}
	result := b.savePDFResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/page.pdf")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) SaveHTML(_ context.Context, req BrowserSaveHTMLRequest) (BrowserSaveHTMLResult, error) {
	b.saveHTMLReqs = append(b.saveHTMLReqs, req)
	if b.saveHTMLErr != nil {
		return BrowserSaveHTMLResult{}, b.saveHTMLErr
	}
	result := b.saveHTMLResult
	if strings.TrimSpace(result.Path) == "" {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/page.html")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) Trace(_ context.Context, req BrowserTraceRequest) (BrowserTraceResult, error) {
	b.traceReqs = append(b.traceReqs, req)
	if b.traceErr != nil {
		return BrowserTraceResult{}, b.traceErr
	}
	result := b.traceResult
	if strings.TrimSpace(result.Path) == "" && strings.EqualFold(strings.TrimSpace(req.Action), "stop") {
		result.Path = firstNonEmpty(strings.TrimSpace(b.remotePath), "/remote/browser/trace.zip")
	}
	return result, nil
}

func (b *remoteArtifactBrowserBackend) ResolveBrowserArtifact(_ context.Context, req browserArtifactResolveRequest) (string, error) {
	b.resolveReqs = append(b.resolveReqs, req)
	if b.resolveErr != nil {
		return "", b.resolveErr
	}
	if err := os.MkdirAll(filepath.Dir(req.RequestedPath), 0o755); err != nil {
		return "", err
	}
	content := b.remoteContent
	if len(content) == 0 {
		content = []byte("remote-browser-artifact")
	}
	if err := os.WriteFile(req.RequestedPath, content, 0o644); err != nil {
		return "", err
	}
	return req.RequestedPath, nil
}

func browserDefinitionNames(reg *llmxtools.Registry) []string {
	defs := reg.Definitions()
	names := make([]string, 0, len(defs))
	for _, def := range defs {
		names = append(names, def.Function.Name)
	}
	return names
}

func browserStringSliceContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func browserActDefinitionKinds(reg *llmxtools.Registry) []string {
	for _, def := range reg.Definitions() {
		if def.Function.Name != "browser_act" {
			continue
		}
		properties, _ := def.Function.Parameters["properties"].(map[string]any)
		kindDef, _ := properties["kind"].(map[string]any)
		rawKinds, _ := kindDef["enum"].([]string)
		if len(rawKinds) > 0 {
			return append([]string(nil), rawKinds...)
		}
		rawAny, _ := kindDef["enum"].([]any)
		kinds := make([]string, 0, len(rawAny))
		for _, item := range rawAny {
			if text, ok := item.(string); ok {
				kinds = append(kinds, text)
			}
		}
		return kinds
	}
	return nil
}

func browserUnifiedDefinitionActions(reg *llmxtools.Registry) []string {
	for _, def := range reg.Definitions() {
		if def.Function.Name != "browser" {
			continue
		}
		properties, _ := def.Function.Parameters["properties"].(map[string]any)
		actionDef, _ := properties["action"].(map[string]any)
		rawActions, _ := actionDef["enum"].([]string)
		if len(rawActions) > 0 {
			return append([]string(nil), rawActions...)
		}
		rawAny, _ := actionDef["enum"].([]any)
		actions := make([]string, 0, len(rawAny))
		for _, item := range rawAny {
			if text, ok := item.(string); ok {
				actions = append(actions, text)
			}
		}
		return actions
	}
	return nil
}

func expectBrowserCompatToolExplicitFallbackOrNotRegistered(t *testing.T, reg *llmxtools.Registry, call types.FunctionCall, toolName string) {
	t.Helper()
	_, err := reg.Execute(context.Background(), call)
	if browserStringSliceContains(browserDefinitionNames(reg), toolName) {
		if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
			t.Fatalf("expected %s to require explicit runtime_target when registered, got %v", toolName, err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("expected %s to stay unregistered for the current capability surface, got %v", toolName, err)
	}
}

func expectBrowserActKindExplicitFallbackOrUnsupported(t *testing.T, reg *llmxtools.Registry, call types.FunctionCall, kind string) {
	t.Helper()
	_, err := reg.Execute(context.Background(), call)
	if browserStringSliceContains(browserActDefinitionKinds(reg), kind) {
		if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
			t.Fatalf("expected browser_act kind %s to require explicit runtime_target when registered, got %v", kind, err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected browser_act kind %s to stay unsupported for the current capability surface, got nil error", kind)
	}
	if !strings.Contains(err.Error(), "unsupported kind") &&
		!strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected browser_act kind %s to stay unsupported for the current capability surface, got %v", kind, err)
	}
}

func TestRegisterBrowserTools_RuntimeStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-session")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					HostOS:                     "darwin",
					HostArch:                   "arm64",
					NodeVersion:                "24.2.0",
					PlaywrightPackageVersion:   "1.55.0",
					RuntimeSummaryGeneration:   "runtime-123",
					RuntimeBaselineReady:       true,
					SelectedLaunchSource:       "runtime_observed",
					SelectedLaunchDelivery:     "delivery-123",
					SelectedLaunchReady:        true,
					SelectedLaunchExecutableOK: true,
					LaunchReady:                true,
					BundleReady:                true,
					DeliveryReady:              true,
					BootstrapState:             "ready",
				},
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "work", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"profile":"isolated","runtime_target":"node","include_routes":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime: %v", err)
	}
	var payload struct {
		Status         string `json:"status"`
		SessionID      string `json:"session_id"`
		SessionBinding struct {
			SessionKey          string `json:"session_key"`
			CurrentTargetID     string `json:"current_target_id"`
			RouteTargetCount    int    `json:"route_target_count"`
			BrowserProfileCount int    `json:"browser_profile_count"`
			PropagatedToProxy   bool   `json:"propagated_to_proxy"`
		} `json:"session_binding"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		RuntimeActions     []string `json:"runtime_actions"`
		BrowserTools       []string `json:"browser_tools"`
		ArtifactTools      []string `json:"artifact_tools"`
		ArtifactKinds      []string `json:"artifact_kinds"`
		ArtifactContract   string   `json:"artifact_contract"`
		BrowserActKinds    []string `json:"browser_act_kinds"`
		ProfileStatus      struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
		LaunchDiagnostics struct {
			Source                           string `json:"source"`
			Backend                          string `json:"backend"`
			Profile                          string `json:"profile"`
			RuntimeTarget                    string `json:"runtime_target"`
			BrowserApp                       string `json:"browser_app"`
			Status                           string `json:"status"`
			HostOS                           string `json:"host_os"`
			HostArch                         string `json:"host_arch"`
			NodeVersion                      string `json:"node_version"`
			PlaywrightPackageVersion         string `json:"playwright_package_version"`
			RuntimeSummaryGeneration         string `json:"runtime_summary_generation"`
			RuntimeBaselineReady             *bool  `json:"runtime_baseline_ready"`
			SelectedLaunchSource             string `json:"selected_launch_source"`
			SelectedLaunchDeliveryGeneration string `json:"selected_launch_delivery_generation"`
			SelectedLaunchReady              *bool  `json:"selected_launch_ready"`
			SelectedLaunchExecutableReady    *bool  `json:"selected_launch_executable_ready"`
			LaunchReady                      *bool  `json:"launch_ready"`
			BundleReady                      *bool  `json:"bundle_ready"`
			DeliveryReady                    *bool  `json:"delivery_ready"`
			BootstrapState                   string `json:"bootstrap_state"`
		} `json:"launch_diagnostics"`
		Doctor struct {
			Status            string `json:"status"`
			Ready             bool   `json:"ready"`
			RepairCommand     string `json:"repair_command"`
			AcceptanceCommand string `json:"acceptance_command"`
			Route             struct {
				Status        string `json:"status"`
				Code          string `json:"code"`
				RuntimeTarget string `json:"runtime_target"`
			} `json:"route"`
			Launch struct {
				Status            string `json:"status"`
				Code              string `json:"code"`
				BootstrapState    string `json:"bootstrap_state"`
				LaunchBlockReason string `json:"launch_block_reason"`
			} `json:"launch"`
		} `json:"doctor"`
		ConfiguredTargets          []string `json:"configured_targets"`
		SubstratePosture           string   `json:"substrate_posture"`
		SubstrateStatus            string   `json:"substrate_status"`
		SubstrateReason            string   `json:"substrate_reason"`
		SubstrateSelectionStrategy string   `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string   `json:"substrate_selection_reason"`
		SelectedRoute              struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SubstrateMatrix []struct {
			Role             string `json:"role"`
			SelectionState   string `json:"selection_state"`
			SelectionReason  string `json:"selection_reason"`
			Status           string `json:"status"`
			Backend          string `json:"backend"`
			Profile          string `json:"profile"`
			RuntimeTarget    string `json:"runtime_target"`
			SubstratePosture string `json:"substrate_posture"`
		} `json:"substrate_matrix"`
		Routes []struct {
			Status           string   `json:"status"`
			Backend          string   `json:"backend"`
			Profile          string   `json:"profile"`
			RuntimeTarget    string   `json:"runtime_target"`
			RuntimeActions   []string `json:"runtime_actions"`
			ArtifactTools    []string `json:"artifact_tools"`
			ArtifactKinds    []string `json:"artifact_kinds"`
			ArtifactContract string   `json:"artifact_contract"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.SessionID != "browser-runtime-status-session" {
		t.Fatalf("expected runtime status payload session id, got %#v", payload)
	}
	if payload.SessionBinding.SessionKey != "browser-runtime-status-session" || !payload.SessionBinding.PropagatedToProxy || payload.SessionBinding.RouteTargetCount != 0 || payload.SessionBinding.BrowserProfileCount != 1 {
		t.Fatalf("unexpected runtime status session binding: %#v", payload.SessionBinding)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime status requests: %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime profiles requests: %#v", nodeBackend.runtimeProfilesReqs)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "isolated" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("unexpected selected route: %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "isolated" || !browserStringSliceContains(payload.ConfiguredProfiles, "isolated") {
		t.Fatalf("unexpected configured profiles payload: default=%q profiles=%#v", payload.DefaultProfile, payload.ConfiguredProfiles)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") || !browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected runtime and act tools in runtime payload, got %#v", payload.BrowserTools)
	}
	if !browserStringSliceContains(payload.ArtifactTools, "browser_act") || browserStringSliceContains(payload.ArtifactTools, "browser_screenshot") {
		t.Fatalf("expected runtime payload artifact tools to reflect enabled screenshot-capable tools, got %#v", payload.ArtifactTools)
	}
	if payload.ArtifactContract != browserArtifactContract || !browserStringSliceContains(payload.ArtifactKinds, "screenshot") {
		t.Fatalf("expected runtime payload artifact contract/kinds, got contract=%q kinds=%#v", payload.ArtifactContract, payload.ArtifactKinds)
	}
	if !browserStringSliceContains(payload.BrowserActKinds, "open") || !browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected browser_act kinds in runtime payload, got %#v", payload.BrowserActKinds)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "prepare") || !browserStringSliceContains(payload.RuntimeActions, "start") || !browserStringSliceContains(payload.RuntimeActions, "profiles") {
		t.Fatalf("expected runtime control actions in payload, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected runtime profile status payload: %#v", payload.ProfileStatus)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Backend != "proxy" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.BrowserApp != "Chromium" ||
		payload.LaunchDiagnostics.Status != "running" ||
		payload.LaunchDiagnostics.HostOS != "darwin" ||
		payload.LaunchDiagnostics.HostArch != "arm64" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.PlaywrightPackageVersion != "1.55.0" ||
		payload.LaunchDiagnostics.RuntimeSummaryGeneration != "runtime-123" ||
		payload.LaunchDiagnostics.RuntimeBaselineReady == nil || !*payload.LaunchDiagnostics.RuntimeBaselineReady ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchDeliveryGeneration != "delivery-123" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.SelectedLaunchExecutableReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchExecutableReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady ||
		payload.LaunchDiagnostics.BundleReady == nil || !*payload.LaunchDiagnostics.BundleReady ||
		payload.LaunchDiagnostics.DeliveryReady == nil || !*payload.LaunchDiagnostics.DeliveryReady ||
		payload.LaunchDiagnostics.BootstrapState != "ready" {
		t.Fatalf("unexpected runtime launch diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
	if payload.Doctor.Status != "ok" || !payload.Doctor.Ready || payload.Doctor.Route.Status != "ok" || payload.Doctor.Route.Code != "managed_default_route" || payload.Doctor.Route.RuntimeTarget != "node" || payload.Doctor.Launch.Status != "ok" || payload.Doctor.Launch.Code != "launch_ready" || payload.Doctor.Launch.BootstrapState != "ready" {
		t.Fatalf("unexpected runtime doctor payload: %#v", payload.Doctor)
	}
	if payload.Doctor.RepairCommand != "" || payload.Doctor.AcceptanceCommand != "" {
		t.Fatalf("expected portable runtime doctor to omit host-owned commands, got %#v", payload.Doctor)
	}
	if !browserStringSliceContains(payload.ConfiguredTargets, "host") || !browserStringSliceContains(payload.ConfiguredTargets, "node") {
		t.Fatalf("expected configured targets in runtime payload, got %#v", payload.ConfiguredTargets)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy || payload.SubstrateSelectionReason == "" {
		t.Fatalf("unexpected substrate selection payload: strategy=%q reason=%q", payload.SubstrateSelectionStrategy, payload.SubstrateSelectionReason)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		!strings.Contains(payload.SubstrateReason, "node runtime backend `proxy`") {
		t.Fatalf("unexpected top-level substrate projection: posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	foundDefaultSubstrate := false
	foundHostFallback := false
	for _, substrate := range payload.SubstrateMatrix {
		if substrate.Role == "default" && substrate.Status == "default" && substrate.Backend == "proxy" && substrate.RuntimeTarget == "node" && substrate.SubstratePosture == BrowserSubstrateNodeRuntime {
			foundDefaultSubstrate = true
		}
		if substrate.Role == "host" && substrate.SelectionState == "explicit_fallback" && substrate.Backend == "system" && substrate.RuntimeTarget == "host" && substrate.SubstratePosture == BrowserSubstrateLegacySystemHost {
			foundHostFallback = true
		}
	}
	if !foundDefaultSubstrate || !foundHostFallback {
		t.Fatalf("expected substrate matrix to include promoted default plus explicit host fallback, got %#v", payload.SubstrateMatrix)
	}
	foundNodeRoute := false
	foundWorkRoute := false
	for _, route := range payload.Routes {
		if (route.Status == "available" || route.Status == "default") && route.RuntimeTarget == "node" && route.Profile == "isolated" && route.Backend == "proxy" {
			if !browserStringSliceContains(route.RuntimeActions, "prepare") || !browserStringSliceContains(route.RuntimeActions, "start") || !browserStringSliceContains(route.RuntimeActions, "profiles") {
				t.Fatalf("expected node route runtime actions, got %#v", route.RuntimeActions)
			}
			if !browserStringSliceContains(route.ArtifactTools, "browser_act") || browserStringSliceContains(route.ArtifactTools, "browser_screenshot") {
				t.Fatalf("expected node route artifact tools to reflect enabled screenshot-capable tools, got %#v", route.ArtifactTools)
			}
			if route.ArtifactContract != browserArtifactContract || !browserStringSliceContains(route.ArtifactKinds, "screenshot") {
				t.Fatalf("expected node route artifact contract/kinds, got contract=%q kinds=%#v", route.ArtifactContract, route.ArtifactKinds)
			}
			foundNodeRoute = true
		}
		if route.Status == "available" && route.RuntimeTarget == "node" && route.Profile == "work" && route.Backend == "proxy" {
			foundWorkRoute = true
		}
	}
	if !foundNodeRoute {
		t.Fatalf("expected available node route in matrix, got %#v", payload.Routes)
	}
	if !foundWorkRoute {
		t.Fatalf("expected discovered profile route in matrix, got %#v", payload.Routes)
	}
}

func TestRegisterBrowserTools_RuntimeStatusIncludesImplicitSpecialistSurfaceWhenOnlyBrowserEnabled(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-unified-enabled")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"profile":"workbench","runtime_target":"node","include_routes":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime implicit specialist surface from browser enablement: %v", err)
	}
	var payload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		BrowserTools    []string `json:"browser_tools"`
		ArtifactTools   []string `json:"artifact_tools"`
		BrowserActKinds []string `json:"browser_act_kinds"`
		Routes          []struct {
			Backend         string   `json:"backend"`
			Profile         string   `json:"profile"`
			RuntimeTarget   string   `json:"runtime_target"`
			Status          string   `json:"status"`
			BrowserTools    []string `json:"browser_tools"`
			BrowserActKinds []string `json:"browser_act_kinds"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode implicit specialist surface runtime payload: %v", err)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected implicit specialist surface runtime payload to resolve explicit node route, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") || !browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected implicit browser enablement to surface runtime and act tools, got %#v", payload.BrowserTools)
	}
	if !browserStringSliceContains(payload.ArtifactTools, "browser_act") {
		t.Fatalf("expected implicit browser enablement to surface browser_act artifact tool, got %#v", payload.ArtifactTools)
	}
	if !browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected implicit browser enablement to surface browser_act kinds for selected route, got %#v", payload.BrowserActKinds)
	}
	foundSelectedNode := false
	for _, route := range payload.Routes {
		if route.Backend != "proxy" || route.Profile != "workbench" || route.RuntimeTarget != "node" || (route.Status != "default" && route.Status != "available") {
			continue
		}
		foundSelectedNode = true
		if !browserStringSliceContains(route.BrowserTools, "browser_act") || !browserStringSliceContains(route.BrowserActKinds, "click") {
			t.Fatalf("expected node route row to reuse implicit specialist surface, got %#v", route)
		}
	}
	if !foundSelectedNode {
		t.Fatalf("expected route matrix to include selected node row, got %#v", payload.Routes)
	}
}

func TestRegisterBrowserTools_RuntimeStatusPreservesManagedTransitionState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-preserve-transition")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-preserve-transition", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  false,
				Note:       "cdp reconnect in progress",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false, Note: "cdp reconnect in progress"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status preserve transition: %v", err)
	}
	var payload struct {
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
			Note      string `json:"note"`
		} `json:"profile_status"`
		Profiles []struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
			Note      string `json:"note"`
		} `json:"profiles"`
		SessionBinding struct {
			SessionHealthState string `json:"session_health_state"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected runtime status payload to reuse synced reconnecting state, got %#v", payload.ProfileStatus)
	}
	if payload.ProfileStatus.Note != "cdp reconnect in progress" {
		t.Fatalf("expected runtime status payload note to preserve managed transition context, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "isolated" || payload.Profiles[0].Status != "reconnecting" || !payload.Profiles[0].Running || payload.Profiles[0].Connected {
		t.Fatalf("expected runtime status profiles payload to reuse synced reconnecting state, got %#v", payload.Profiles)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" {
		t.Fatalf("expected runtime status session health to observe reconnecting state, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeStatusDefaultsToPromotedNodeRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime: %v", err)
	}
	if len(hostBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected promoted default route to avoid host runtime status, got %#v", hostBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("expected promoted default node runtime status request, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var payload struct {
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected default route to reflect promoted node runtime, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "isolated" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selected route to follow promoted node runtime, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "running" {
		t.Fatalf("unexpected profile status payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStatusKeepsSandboxExplicitByDefault(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sandboxBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{
					runtimeStatusResult: BrowserProfileStatusResult{
						Backend:    "sandbox",
						BrowserApp: "Chromium",
						Profile:    "default",
						Status:     "running",
						Running:    true,
						Connected:  true,
					},
				},
				runtimeInfo: BrowserRuntimeInfo{Backend: "sandbox", Profile: "default", Target: "sandbox"},
			}},
			capabilities: BrowserCapabilitiesForActKinds([]string{
				"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate",
			}),
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:           t.TempDir(),
		SandboxBackend: sandboxBackend,
		EnabledTools:   []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime: %v", err)
	}
	if len(sandboxBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected sandbox lane to stay explicit-only for default runtime status, got %#v", sandboxBackend.runtimeStatusReqs)
	}
	var payload struct {
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SubstrateSelectionStrategy string `json:"substrate_selection_strategy"`
		ProfileStatus              struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected default route to stay hidden when only sandbox is configured, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected selected route to stay unset when sandbox remains explicit-only, got %#v", payload.SelectedRoute)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionLegacyHostDefault {
		t.Fatalf("expected sandbox-only runtime status to keep legacy-host default strategy, got %#v", payload)
	}
	if payload.ProfileStatus.Profile != "" || payload.ProfileStatus.Status != "" {
		t.Fatalf("unexpected profile status payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeAvailableActionsSkipUnresolvedNodeRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{
				"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate",
			}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		NodeBackend: nodeBackend,
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	actions := browserRuntimeAvailableActions(ctx)
	for _, forbidden := range []string{"prepare", "coordinate", "start", "restart", "refresh", "stop", "profiles", "create_profile", "delete_profile"} {
		if browserStringSliceContains(actions, forbidden) {
			t.Fatalf("expected unresolved node route to suppress %q, got %#v", forbidden, actions)
		}
	}
}

func TestRegisterBrowserTools_RuntimeAvailableActionsSkipUnresolvedSandboxRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sandboxBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "sandbox", Target: "sandbox"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{
				"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate",
			}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root:           t.TempDir(),
		SandboxBackend: sandboxBackend,
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	actions := browserRuntimeAvailableActions(ctx)
	for _, forbidden := range []string{"prepare", "coordinate", "start", "restart", "refresh", "stop", "profiles", "create_profile", "delete_profile"} {
		if browserStringSliceContains(actions, forbidden) {
			t.Fatalf("expected unresolved sandbox route to suppress %q, got %#v", forbidden, actions)
		}
	}
}

func TestRegisterBrowserTools_RuntimeAvailableActionsKeepDiagnosticsForBrokenDefaultHostRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{
				"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate",
			}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	actions := browserRuntimeAvailableActions(ctx)
	for _, required := range []string{"status", "profiles", "sessions", "workbench"} {
		if !browserStringSliceContains(actions, required) {
			t.Fatalf("expected broken default host route to keep diagnostics action %q, got %#v", required, actions)
		}
	}
	for _, forbidden := range []string{"prepare", "coordinate", "start", "restart", "refresh", "stop", "create_profile", "delete_profile"} {
		if browserStringSliceContains(actions, forbidden) {
			t.Fatalf("expected broken default host route without managed lane to suppress %q, got %#v", forbidden, actions)
		}
	}
}

func TestRegisterBrowserTools_RuntimeSessionsPreservesExplicitCustomHostSubstrate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Status                     string   `json:"status"`
		ConfiguredTargets          []string `json:"configured_targets"`
		SubstrateSelectionStrategy string   `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string   `json:"substrate_selection_reason"`
		SelectedRoute              struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SubstrateMatrix []struct {
			Role             string `json:"role"`
			Status           string `json:"status"`
			SelectionState   string `json:"selection_state"`
			Backend          string `json:"backend"`
			Profile          string `json:"profile"`
			RuntimeTarget    string `json:"runtime_target"`
			SubstratePosture string `json:"substrate_posture"`
		} `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "custom-playwright" || payload.SelectedRoute.Profile != "default" || payload.SelectedRoute.RuntimeTarget != "host" {
		t.Fatalf("expected explicit custom host to remain selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferCustomBackend || payload.SubstrateSelectionReason == "" {
		t.Fatalf("unexpected substrate selection payload: strategy=%q reason=%q", payload.SubstrateSelectionStrategy, payload.SubstrateSelectionReason)
	}
	if !browserStringSliceContains(payload.ConfiguredTargets, "host") || !browserStringSliceContains(payload.ConfiguredTargets, "node") {
		t.Fatalf("expected configured targets to retain host and node routes, got %#v", payload.ConfiguredTargets)
	}
	foundDefaultCustom := false
	foundNodeRoute := false
	for _, substrate := range payload.SubstrateMatrix {
		if substrate.Role == "default" && substrate.Status == "default" && substrate.Backend == "custom-playwright" && substrate.Profile == "default" && substrate.RuntimeTarget == "host" && substrate.SubstratePosture == BrowserSubstrateCustomBackend {
			foundDefaultCustom = true
		}
		if substrate.Role == "node" && substrate.SelectionState == "available" && substrate.Backend == "proxy" && substrate.RuntimeTarget == "node" && substrate.SubstratePosture == BrowserSubstrateNodeRuntime {
			foundNodeRoute = true
		}
	}
	if !foundDefaultCustom || !foundNodeRoute {
		t.Fatalf("expected substrate matrix to keep custom host default and node as explicit route, got %#v", payload.SubstrateMatrix)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsNodeRouteResolutionFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Status                     string `json:"status"`
		SubstratePosture           string `json:"substrate_posture"`
		SubstrateStatus            string `json:"substrate_status"`
		SubstrateReason            string `json:"substrate_reason"`
		SubstrateSelectionStrategy string `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string `json:"substrate_selection_reason"`
		DefaultRoute               struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SubstrateMatrix []struct {
			Role            string   `json:"role"`
			Status          string   `json:"status"`
			SelectionState  string   `json:"selection_state"`
			SelectionReason string   `json:"selection_reason"`
			Backend         string   `json:"backend"`
			Profile         string   `json:"profile"`
			RuntimeTarget   string   `json:"runtime_target"`
			RuntimeActions  []string `json:"runtime_actions"`
			Note            string   `json:"note"`
		} `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy || !strings.Contains(payload.SubstrateSelectionReason, "could not be resolved") {
		t.Fatalf("unexpected substrate selection payload: %#v", payload)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		!strings.Contains(payload.SubstrateReason, "could not be resolved") {
		t.Fatalf("expected top-level substrate projection to reuse hidden node route failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	foundUnsupportedDefaultHost := false
	foundExplicitHostFallback := false
	foundUnsupportedNode := false
	for _, substrate := range payload.SubstrateMatrix {
		if substrate.Role == "default" && substrate.Status == "unsupported" && substrate.SelectionState == "default" && substrate.Backend == "" && substrate.Profile == "" && substrate.RuntimeTarget == "" && strings.Contains(substrate.SelectionReason, "could not be resolved") {
			foundUnsupportedDefaultHost = true
		}
		if substrate.Role == "host" && substrate.Status == "available" && substrate.SelectionState == "explicit_fallback" && substrate.Backend == "system" && substrate.Profile == "default" && substrate.RuntimeTarget == "host" {
			foundExplicitHostFallback = true
		}
		if substrate.Role == "node" && substrate.Status == "unsupported" && substrate.SelectionState == "unsupported" && substrate.Backend == "proxy" && substrate.Profile == "isolated" && substrate.RuntimeTarget == "node" && strings.Contains(substrate.SelectionReason, "could not be resolved") && strings.Contains(substrate.Note, "context deadline exceeded") {
			foundUnsupportedNode = true
		}
	}
	if !foundUnsupportedDefaultHost || !foundExplicitHostFallback || !foundUnsupportedNode {
		t.Fatalf("expected substrate matrix to expose unsupported default plus explicit host fallback, got %#v", payload.SubstrateMatrix)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsHideImplicitHostCurrentTargetFallbackSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-implicit-host-current-target")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-sessions-implicit-host-current-target", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions implicit host current target: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution        any `json:"route_resolution"`
		SessionTargetSelection *struct {
			ID            string `json:"id"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			CurrentTargetID             string `json:"current_target_id"`
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTarget       string `json:"selected_browser_target"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
			Coordination                *struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
			Targets         []struct {
				ID      string `json:"id"`
				Current bool   `json:"current"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sessions output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected top-level session target selection to hide implicit host fallback selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.RouteResolution != nil {
		t.Fatalf("expected top-level route resolution to hide implicit host fallback source, got %#v", payload.RouteResolution)
	}
	if payload.SessionBinding.CurrentTargetID != "" || payload.SessionBinding.SelectedBrowserTargetID != "" || payload.SessionBinding.SelectedBrowserTarget != "" || payload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("expected top-level session binding to hide implicit host fallback target selection, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination != nil {
		t.Fatalf("expected top-level session binding to hide implicit host fallback coordination, got %#v", payload.SessionBinding.Coordination)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "system" || payload.SessionRoutes[0].Profile != "default" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("expected explicit host fallback route snapshot to remain available, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected explicit host fallback route snapshot to retain current target, got %#v", payload.SessionRoutes[0])
	}
	if len(payload.SessionRoutes[0].Targets) != 1 || payload.SessionRoutes[0].Targets[0].ID != tracked.ID || !payload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("expected explicit host fallback route targets to remain available, got %#v", payload.SessionRoutes[0].Targets)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsHideImplicitHostProfileFallbackSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-implicit-host-profile")
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-sessions-implicit-host-profile", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions implicit host profile: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution         any `json:"route_resolution"`
		SessionProfileSelection *struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_profile_selection"`
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			Coordination                 *struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sessions output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected top-level session profile selection to hide implicit host fallback selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.RouteResolution != nil {
		t.Fatalf("expected top-level route resolution to hide implicit host fallback source, got %#v", payload.RouteResolution)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "" || payload.SessionBinding.SelectedBrowserProfileSource != "" {
		t.Fatalf("expected top-level session binding to hide implicit host fallback profile selection, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination != nil {
		t.Fatalf("expected top-level session binding to hide implicit host fallback coordination, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchHideImplicitHostFallbackCoordination(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-implicit-host-coordination")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-workbench-implicit-host-coordination", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-implicit-host-coordination", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-workbench-implicit-host-coordination", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench implicit host coordination: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		Profiles        []struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"profiles"`
		ProfileStatus *struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"profile_status"`
		WorkbenchReady                     bool     `json:"workbench_ready"`
		WorkbenchSections                  []string `json:"workbench_sections"`
		WorkbenchPrimaryBrowserAction      string   `json:"workbench_primary_browser_action"`
		WorkbenchPrimaryNodeAction         string   `json:"workbench_primary_node_action"`
		WorkbenchNextStep                  string   `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string `json:"workbench_recommended_browser_actions"`
		WorkbenchRecommendedNodeActions    []string `json:"workbench_recommended_node_actions"`
		SessionBinding                     struct {
			BrowserProfileCount  int    `json:"browser_profile_count"`
			ActiveBrowserProfile string `json:"active_browser_profile"`
			Coordination         *struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode workbench output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected workbench hidden-host coordination to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.RouteResolution != nil {
		t.Fatalf("expected top-level route resolution to hide implicit host fallback source, got %#v", payload.RouteResolution)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil || len(payload.Profiles) != 0 {
		t.Fatalf("expected top-level workbench status/profile summary to hide implicit host fallback, got %#v", payload)
	}
	if !payload.WorkbenchReady {
		t.Fatalf("expected workbench payload to stay ready with remaining sections, got %#v", payload)
	}
	if browserStringSliceContains(payload.WorkbenchSections, "coordination") {
		t.Fatalf("expected workbench sections to hide implicit host fallback coordination, got %#v", payload.WorkbenchSections)
	}
	if browserStringSliceContains(payload.WorkbenchSections, "status") || browserStringSliceContains(payload.WorkbenchSections, "profiles") {
		t.Fatalf("expected workbench sections to hide implicit host fallback status/profile summary, got %#v", payload.WorkbenchSections)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected workbench note to surface hidden managed route guidance, got %#v", payload)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=prepare" || payload.WorkbenchPrimaryNodeAction != "" || payload.WorkbenchNextStep != "browser_runtime action=prepare" {
		t.Fatalf("expected top-level workbench plan to promote hidden managed route guidance, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=prepare") || len(payload.WorkbenchRecommendedNodeActions) != 0 {
		t.Fatalf("expected workbench recommendations to surface runtime prepare guidance without node actions, got %#v", payload)
	}
	if payload.SessionBinding.Coordination != nil {
		t.Fatalf("expected top-level session binding to hide implicit host fallback coordination, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.BrowserProfileCount != 0 || payload.SessionBinding.ActiveBrowserProfile != "" {
		t.Fatalf("expected top-level session binding to hide implicit host fallback profile summary, got %#v", payload.SessionBinding)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "system" || payload.SessionRoutes[0].Profile != "default" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("expected explicit host fallback route snapshot to remain available, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected explicit host fallback route snapshot to retain current target, got %#v", payload.SessionRoutes[0])
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchTreatsImplicitHostDefaultProfileAsDiagnosticsRequest(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-implicit-host-default-profile")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-workbench-implicit-host-default-profile", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","profile":"default"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench implicit host default profile: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		Note            string `json:"note"`
		SelectedRoute   any    `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		SessionRoutes   []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode workbench output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected default-profile workbench diagnostics to stay ok, got %#v", payload)
	}
	if payload.Note != "" {
		t.Fatalf("expected default-profile workbench diagnostics to short-circuit without route note, got %#v", payload)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected default-profile workbench diagnostics to hide implicit host route state, got %#v", payload)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "system" || payload.SessionRoutes[0].Profile != "default" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("expected default-profile workbench diagnostics to keep explicit host route snapshot, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected default-profile workbench diagnostics to retain explicit host current target, got %#v", payload.SessionRoutes[0])
	}
}

func TestRegisterBrowserTools_RuntimeStatusHideImplicitHostProfileSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-implicit-host-profile-summary")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-status-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status implicit host profile summary: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		ProfileStatus   any    `json:"profile_status"`
		SessionBinding  struct {
			BrowserProfileCount         int    `json:"browser_profile_count"`
			ActiveBrowserProfile        string `json:"active_browser_profile"`
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthReason         string `json:"session_health_reason"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.RouteResolution != nil {
		t.Fatalf("expected top-level route resolution to hide implicit host fallback source, got %#v", payload.RouteResolution)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected top-level status summary to hide implicit host fallback, got %#v", payload)
	}
	if payload.SessionBinding.BrowserProfileCount != 0 || payload.SessionBinding.ActiveBrowserProfile != "" || payload.SessionBinding.SessionHealthState != "" || payload.SessionBinding.SessionHealthReason != "" || payload.SessionBinding.SessionHealthRecoveryAction != "" || payload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected session binding to hide implicit host profile summary, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesHideImplicitHostProfileSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-implicit-host-profile-summary")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-profiles-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles implicit host profile summary: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		Profiles        []any  `json:"profiles"`
		SessionBinding  struct {
			BrowserProfileCount         int    `json:"browser_profile_count"`
			ActiveBrowserProfile        string `json:"active_browser_profile"`
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthReason         string `json:"session_health_reason"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode profiles output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected unsupported status for implicit host fallback profiles, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host fallback to omit top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.RouteResolution != nil {
		t.Fatalf("expected top-level route resolution to hide implicit host fallback source, got %#v", payload.RouteResolution)
	}
	if payload.Note != "" {
		t.Fatalf("expected implicit host fallback profiles to skip route error note, got %#v", payload)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected top-level profiles summary to hide implicit host fallback, got %#v", payload)
	}
	if payload.SessionBinding.BrowserProfileCount != 0 || payload.SessionBinding.ActiveBrowserProfile != "" || payload.SessionBinding.SessionHealthState != "" || payload.SessionBinding.SessionHealthReason != "" || payload.SessionBinding.SessionHealthRecoveryAction != "" || payload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected session binding to hide implicit host profile summary, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeStatusTreatsImplicitHostDefaultProfileAsDiagnosticsRequest(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-implicit-host-default-profile")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-implicit-host-default-profile", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"default"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status implicit host default profile: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		Note            string `json:"note"`
		SelectedRoute   any    `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		ProfileStatus   any    `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected default-profile status diagnostics to stay ok, got %#v", payload)
	}
	if payload.Note != "" {
		t.Fatalf("expected default-profile status diagnostics to short-circuit without route note, got %#v", payload)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected default-profile status diagnostics to hide implicit host route state, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected default-profile status diagnostics to keep top-level implicit host summary hidden, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesTreatsImplicitHostDefaultProfileAsDiagnosticsRequest(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-implicit-host-default-profile")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-implicit-host-default-profile", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","profile":"default"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles implicit host default profile: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		Note            string `json:"note"`
		SelectedRoute   any    `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		Profiles        []any  `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode profiles output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected default-profile profiles diagnostics to stay unsupported, got %#v", payload)
	}
	if payload.Note != "" {
		t.Fatalf("expected default-profile profiles diagnostics to short-circuit without route note, got %#v", payload)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected default-profile profiles diagnostics to hide implicit host route state, got %#v", payload)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected default-profile profiles diagnostics to keep implicit host summary hidden, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsTreatsImplicitHostDefaultProfileAsDiagnosticsRequest(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-implicit-host-default-profile")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-sessions-implicit-host-default-profile", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","profile":"default"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions implicit host default profile: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		Note            string `json:"note"`
		SelectedRoute   any    `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		SessionRoutes   []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sessions output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected default-profile sessions diagnostics to stay ok, got %#v", payload)
	}
	if payload.Note != "" {
		t.Fatalf("expected default-profile sessions diagnostics to short-circuit without route note, got %#v", payload)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected default-profile sessions diagnostics to hide implicit host route state, got %#v", payload)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "system" || payload.SessionRoutes[0].Profile != "default" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("expected default-profile sessions diagnostics to keep explicit host route snapshot, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected default-profile sessions diagnostics to retain explicit host current target, got %#v", payload.SessionRoutes[0])
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsSandboxRouteResolutionFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sandboxBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "sandbox", Target: "sandbox"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:           t.TempDir(),
		SandboxBackend: sandboxBackend,
		EnabledTools:   []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		SubstrateMatrix []struct {
			Role            string `json:"role"`
			Status          string `json:"status"`
			SelectionState  string `json:"selection_state"`
			SelectionReason string `json:"selection_reason"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			Backend         string `json:"backend"`
			Note            string `json:"note"`
		} `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", payload)
	}
	foundUnsupportedSandbox := false
	for _, substrate := range payload.SubstrateMatrix {
		if substrate.Role == "sandbox" && substrate.Status == "unsupported" && substrate.SelectionState == "unsupported" && substrate.Backend == "sandbox" && substrate.Profile == "default" && substrate.RuntimeTarget == "sandbox" && strings.Contains(substrate.SelectionReason, "could not be resolved") && strings.Contains(substrate.Note, "context deadline exceeded") {
			foundUnsupportedSandbox = true
		}
	}
	if !foundUnsupportedSandbox {
		t.Fatalf("expected substrate matrix to expose sandbox route-resolution failure, got %#v", payload.SubstrateMatrix)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesRejectsImplicitHostAlternateProfileDiagnosticsDegrade(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-implicit-host-alternate-profile")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-implicit-host-alternate-profile", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "relay",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host relay profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","profile":"relay"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles implicit-host alternate profile: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Profiles []struct {
			Profile string `json:"profile"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected implicit legacy host alternate profile to stay unsupported, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "requires an explicit runtime_target") &&
		!strings.Contains(payload.Note, `profile "relay" is unsupported for backend "system"`) {
		t.Fatalf("expected unsupported payload note to surface the targetless implicit-host route error, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host alternate profile to omit selected route, got %#v", payload.SelectedRoute)
	}
	if len(payload.Profiles) != 0 {
		t.Fatalf("expected implicit legacy host alternate profile to skip degraded cached profiles payload, got %#v", payload.Profiles)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsRejectsImplicitHostAlternateProfileDiagnosticsDegrade(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-implicit-host-alternate-profile")
	sessionRegistry.TrackTabs("browser-runtime-sessions-implicit-host-alternate-profile", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/relay",
			Title:      "Relay",
			BrowserApp: "Safari",
			Backend:    "system",
			Profile:    "relay",
			Target:     "host",
		},
	}, 1)
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","profile":"relay"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions implicit-host alternate profile: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetCount int `json:"session_target_count"`
		SessionRoutes      []struct {
			Profile string `json:"profile"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected implicit legacy host alternate session profile to stay unsupported, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "requires an explicit runtime_target") &&
		!strings.Contains(payload.Note, `profile "relay" is unsupported for backend "system"`) {
		t.Fatalf("expected unsupported payload note to surface the targetless implicit-host route error, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected implicit legacy host to stay out of payload default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected implicit legacy host alternate session profile to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetCount != 0 || len(payload.SessionRoutes) != 0 {
		t.Fatalf("expected implicit legacy host alternate session profile to skip degraded cached session payload, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsHostRouteResolutionFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Status                     string `json:"status"`
		SubstrateSelectionStrategy string `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string `json:"substrate_selection_reason"`
		DefaultRoute               struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Note            string `json:"note"`
		SubstrateMatrix []struct {
			Role            string   `json:"role"`
			Status          string   `json:"status"`
			SelectionState  string   `json:"selection_state"`
			SelectionReason string   `json:"selection_reason"`
			Profile         string   `json:"profile"`
			Backend         string   `json:"backend"`
			RuntimeTarget   string   `json:"runtime_target"`
			RuntimeActions  []string `json:"runtime_actions"`
			Note            string   `json:"note"`
		} `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected diagnostics action to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferCustomBackend || !strings.Contains(payload.SubstrateSelectionReason, "could not be resolved") {
		t.Fatalf("expected substrate selection payload to surface host route-resolution failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	foundDefaultHostFailure := false
	foundAvailableNode := false
	for _, substrate := range payload.SubstrateMatrix {
		if substrate.Role == "default" && substrate.Status == "unsupported" && substrate.SelectionState == "default" && substrate.Backend == "custom-playwright" && substrate.Profile == "workbench" && substrate.RuntimeTarget == "host" && strings.Contains(substrate.SelectionReason, "could not be resolved") && strings.Contains(substrate.Note, "context deadline exceeded") {
			if !browserStringSliceContains(substrate.RuntimeActions, "status") {
				t.Fatalf("expected failing default substrate row to keep diagnostics status action, got %#v", substrate)
			}
			foundDefaultHostFailure = true
		}
		if substrate.Role == "node" && substrate.Status == "available" && substrate.SelectionState == "available" && substrate.Backend == "proxy" && substrate.RuntimeTarget == "node" {
			foundAvailableNode = true
		}
	}
	if !foundDefaultHostFailure || !foundAvailableNode {
		t.Fatalf("expected substrate matrix to expose failing host default and still list node route, got %#v", payload.SubstrateMatrix)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsHostRouteResolutionFailureUsesRouteScopedSnapshotWithoutRequestedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-host-route-failure-route-scope")
	sessionRegistry.TrackTabs("browser-runtime-sessions-host-route-failure-route-scope", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/host",
			Title:      "Host Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
		{
			TabIndex:   2,
			URL:        "https://93.184.216.34/node",
			Title:      "Node Dashboard",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "isolated",
			Target:     "node",
		},
	}, 1)
	sessionRunRegistry.Record("browser-runtime-sessions-host-route-failure-route-scope", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-shared-1",
		NodeID:   "node-shared",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-sessions-host-route-failure-route-scope", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route session profile",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-sessions-host-route-failure-route-scope", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached node route session profile",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions route-scoped diagnostics: %v", err)
	}
	var payload struct {
		Status             string `json:"status"`
		Note               string `json:"note"`
		SessionTargetCount int    `json:"session_target_count"`
		SessionRoutes      []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_routes"`
		SessionProfiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected sessions diagnostics to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected degraded sessions diagnostics without requested profile to stay route-scoped, got %#v", payload)
	}
	if payload.SessionRoutes[0].Backend != "custom-playwright" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("unexpected route-scoped degraded session route payload: %#v", payload.SessionRoutes[0])
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Backend != "custom-playwright" || payload.SessionProfiles[0].Profile != "workbench" || payload.SessionProfiles[0].RuntimeTarget != "host" {
		t.Fatalf("expected degraded sessions diagnostics without requested profile to exclude unrelated node route state, got %#v", payload.SessionProfiles)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesReportsHostRouteResolutionFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-host-route-failure")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route profile",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "archive",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "stopped",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles: %v", err)
	}
	var payload struct {
		Status         string   `json:"status"`
		Note           string   `json:"note"`
		RuntimeActions []string `json:"runtime_actions"`
		DefaultRoute   struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		Profiles           []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected profiles diagnostics to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	for _, want := range []string{"status", "profiles", "workbench"} {
		if !browserStringSliceContains(payload.RuntimeActions, want) {
			t.Fatalf("expected degraded profiles payload to keep diagnostics action %q, got %#v", want, payload.RuntimeActions)
		}
	}
	if len(payload.Profiles) != 1 {
		t.Fatalf("expected requested profile filter to reuse cached route-scoped profiles, got %#v", payload.Profiles)
	}
	if payload.Profiles[0].Backend != "custom-playwright" || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].RuntimeTarget != "host" || payload.Profiles[0].Status != "running" || !payload.Profiles[0].Running || !payload.Profiles[0].Connected || payload.Profiles[0].Note != "cached host route profile" {
		t.Fatalf("unexpected degraded profiles payload: %#v", payload.Profiles[0])
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 {
		t.Fatalf("expected degraded profiles diagnostics to avoid unrelated managed runtime profile calls, got %#v", nodeBackend.runtimeProfilesReqs)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesDoesNotDegradeMissingRequestedProfileOnHostRouteFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-host-route-missing-profile")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-host-route-missing-profile", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route profile",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","profile":"archive"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles missing requested profile: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Profiles []struct {
			Profile string `json:"profile"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected missing requested profile to stay unsupported on host route failure, got %#v", payload)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected unsupported payload note to surface host route failure, got %#v", payload)
	}
	if len(payload.Profiles) != 0 {
		t.Fatalf("expected missing requested profile to avoid degraded cached profiles payload, got %#v", payload.Profiles)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 {
		t.Fatalf("expected unsupported diagnostics to avoid unrelated managed runtime profile calls, got %#v", nodeBackend.runtimeProfilesReqs)
	}
}

func TestRegisterBrowserTools_RuntimeStatusDoesNotDegradeWithoutCachedDefaultRouteSnapshotOnHostRouteFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status without cached snapshot: %v", err)
	}
	var payload struct {
		Status        string `json:"status"`
		Note          string `json:"note"`
		ProfileStatus any    `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected status diagnostics to stay unsupported without cached default-route snapshot, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected unsupported payload note to surface host route failure, got %#v", payload)
	}
	if payload.ProfileStatus != nil {
		t.Fatalf("expected unsupported status payload to skip degraded profile_status without cached snapshot, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesDoesNotDegradeWithoutCachedDefaultRouteSnapshotOnHostRouteFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles without cached snapshot: %v", err)
	}
	var payload struct {
		Status   string `json:"status"`
		Note     string `json:"note"`
		Profiles []any  `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" {
		t.Fatalf("expected profiles diagnostics to stay unsupported without cached default-route snapshot, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected unsupported payload note to surface host route failure, got %#v", payload)
	}
	if len(payload.Profiles) != 0 {
		t.Fatalf("expected unsupported profiles payload to skip degraded cached profiles without cached owner, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesReportsHostRouteResolutionFailureForRequestedProfileWithoutStateRegistry(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-host-route-failure-no-state-registry")
	sessionRegistry.TrackTabs("browser-runtime-profiles-host-route-failure-no-state-registry", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles without state registry: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend string `json:"backend"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		Profiles           []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected profiles diagnostics without state registry to stay ok, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Backend != "custom-playwright" || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].RuntimeTarget != "host" || payload.Profiles[0].BrowserApp != "chrome" || payload.Profiles[0].Status != "" || payload.Profiles[0].Running || payload.Profiles[0].Connected || payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("unexpected degraded profiles payload without state registry: %#v", payload.Profiles)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 {
		t.Fatalf("expected degraded profiles diagnostics without state registry to avoid unrelated managed runtime profile calls, got %#v", nodeBackend.runtimeProfilesReqs)
	}
}

func TestRegisterBrowserTools_RuntimeStatusReportsHostRouteResolutionFailure(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-host-route-failure")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route status",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status: %v", err)
	}
	var payload struct {
		Status         string   `json:"status"`
		Note           string   `json:"note"`
		RuntimeActions []string `json:"runtime_actions"`
		DefaultRoute   struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		ProfileStatus      *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profile_status"`
		Routes []struct {
			Backend        string   `json:"backend"`
			Profile        string   `json:"profile"`
			RuntimeTarget  string   `json:"runtime_target"`
			Status         string   `json:"status"`
			RuntimeActions []string `json:"runtime_actions"`
			Note           string   `json:"note"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected status diagnostics to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	for _, want := range []string{"status", "profiles", "workbench"} {
		if !browserStringSliceContains(payload.RuntimeActions, want) {
			t.Fatalf("expected degraded status payload to keep diagnostics action %q, got %#v", want, payload.RuntimeActions)
		}
	}
	if payload.ProfileStatus == nil {
		t.Fatalf("expected degraded status diagnostics to reuse cached route-scoped profile status, got %#v", payload)
	}
	if payload.ProfileStatus.Backend != "custom-playwright" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "host" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected || payload.ProfileStatus.Note != "cached host route status" {
		t.Fatalf("unexpected degraded profile status payload: %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
	foundUnsupportedDefault := false
	foundAvailableNode := false
	for _, route := range payload.Routes {
		if route.Status == "unsupported" &&
			route.Backend == "custom-playwright" &&
			route.Profile == "workbench" &&
			route.RuntimeTarget == "host" &&
			strings.Contains(route.Note, "context deadline exceeded") {
			for _, want := range []string{"status", "profiles", "workbench"} {
				if !browserStringSliceContains(route.RuntimeActions, want) {
					t.Fatalf("expected failing default route row to keep diagnostics action %q, got %#v", want, route)
				}
			}
			foundUnsupportedDefault = true
		}
		if route.Status == "available" &&
			route.Backend == "proxy" &&
			route.RuntimeTarget == "node" {
			foundAvailableNode = true
		}
	}
	if !foundUnsupportedDefault || !foundAvailableNode {
		t.Fatalf("expected route matrix to include cached default failure and available node lane, got %#v", payload.Routes)
	}
	if len(nodeBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected degraded status diagnostics to avoid unrelated managed runtime status calls, got %#v", nodeBackend.runtimeStatusReqs)
	}
}

func TestRegisterBrowserTools_RuntimeStatusReportsHostRouteResolutionFailureIncludesExplicitManagedOptInSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-host-route-failure-explicit-opt-in")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-host-route-failure-explicit-opt-in", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route status",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act", "browser_click"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status explicit opt-in diagnostics: %v", err)
	}
	var payload struct {
		Status              string          `json:"status"`
		SelectedRoute       *struct{}       `json:"selected_route"`
		BrowserTools        []string        `json:"browser_tools"`
		BrowserActKinds     []string        `json:"browser_act_kinds"`
		BrowserSurface      string          `json:"browser_surface"`
		BrowserOptInTargets []string        `json:"browser_opt_in_targets"`
		Capabilities        map[string]bool `json:"capabilities"`
		Routes              []struct {
			Backend             string          `json:"backend"`
			Profile             string          `json:"profile"`
			RuntimeTarget       string          `json:"runtime_target"`
			Status              string          `json:"status"`
			BrowserTools        []string        `json:"browser_tools"`
			BrowserActKinds     []string        `json:"browser_act_kinds"`
			BrowserSurface      string          `json:"browser_surface"`
			BrowserOptInTargets []string        `json:"browser_opt_in_targets"`
			Capabilities        map[string]bool `json:"capabilities"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode explicit opt-in diagnostics output: %v", err)
	}
	if payload.Status != "ok" || payload.SelectedRoute != nil {
		t.Fatalf("expected degraded status diagnostics without selected route, got %#v", payload)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_click") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected degraded status diagnostics to expose explicit managed opt-in tools, got %#v", payload.BrowserTools)
	}
	if !browserStringSliceContains(payload.BrowserActKinds, "click") || payload.Capabilities["click"] != true {
		t.Fatalf("expected degraded status diagnostics to expose explicit managed browser_act click surface, got tools=%#v kinds=%#v capabilities=%#v", payload.BrowserTools, payload.BrowserActKinds, payload.Capabilities)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected degraded status diagnostics to expose explicit managed opt-in surface contract, got %#v", payload)
	}
	foundUnsupportedDefault := false
	for _, route := range payload.Routes {
		if route.Status != "unsupported" ||
			route.Backend != "custom-playwright" ||
			route.Profile != "workbench" ||
			route.RuntimeTarget != "host" {
			continue
		}
		foundUnsupportedDefault = true
		if !browserStringSliceContains(route.BrowserTools, "browser_click") ||
			!browserStringSliceContains(route.BrowserTools, "browser_act") ||
			!browserStringSliceContains(route.BrowserActKinds, "click") ||
			route.Capabilities["click"] != true {
			t.Fatalf("expected unsupported default route row to reuse explicit managed opt-in diagnostics surface, got %#v", route)
		}
		if route.BrowserSurface != "explicit_managed_opt_in" ||
			len(route.BrowserOptInTargets) != 1 ||
			route.BrowserOptInTargets[0] != "node" {
			t.Fatalf("expected unsupported default route row to expose explicit managed opt-in surface contract, got %#v", route)
		}
	}
	if !foundUnsupportedDefault {
		t.Fatalf("expected route matrix to include unsupported default host row, got %#v", payload.Routes)
	}
}

func TestRegisterBrowserTools_RuntimeStatusReportsHostRouteResolutionFailureForRequestedProfileWithoutStateRegistry(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-host-route-failure-no-state-registry")
	sessionRegistry.TrackTabs("browser-runtime-status-host-route-failure-no-state-registry", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status without state registry: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend string `json:"backend"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		ProfileStatus      *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected status diagnostics without state registry to stay ok, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Backend != "custom-playwright" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "host" || payload.ProfileStatus.BrowserApp != "chrome" || payload.ProfileStatus.Status != "" || payload.ProfileStatus.Running || payload.ProfileStatus.Connected || payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("unexpected degraded status payload without state registry: %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(nodeBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected degraded status diagnostics without state registry to avoid unrelated managed runtime status calls, got %#v", nodeBackend.runtimeStatusReqs)
	}
}

func TestRegisterBrowserTools_RuntimeStatusDiagnosticsUseConcreteManagedOptInCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-concrete-opt-in-capabilities")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-concrete-opt-in-capabilities", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	var nodeBackend *countingExecutionRouteResolverBrowserBackend
	nodeBackend = &countingExecutionRouteResolverBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
		resolveExecution: func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if requested != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
				return browserResolvedExecutionRoute{}, context.DeadlineExceeded
			}
			return browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  requested,
				Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
			}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act", "browser_click", "browser_screenshot"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status concrete opt-in diagnostics: %v", err)
	}
	var payload struct {
		BrowserTools        []string `json:"browser_tools"`
		ArtifactTools       []string `json:"artifact_tools"`
		ArtifactKinds       []string `json:"artifact_kinds"`
		ArtifactContract    string   `json:"artifact_contract"`
		BrowserActKinds     []string `json:"browser_act_kinds"`
		BrowserSurface      string   `json:"browser_surface"`
		BrowserOptInTargets []string `json:"browser_opt_in_targets"`
		Routes              []struct {
			Backend             string   `json:"backend"`
			Profile             string   `json:"profile"`
			RuntimeTarget       string   `json:"runtime_target"`
			Status              string   `json:"status"`
			BrowserTools        []string `json:"browser_tools"`
			ArtifactTools       []string `json:"artifact_tools"`
			ArtifactKinds       []string `json:"artifact_kinds"`
			ArtifactContract    string   `json:"artifact_contract"`
			BrowserActKinds     []string `json:"browser_act_kinds"`
			BrowserSurface      string   `json:"browser_surface"`
			BrowserOptInTargets []string `json:"browser_opt_in_targets"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode concrete opt-in diagnostics output: %v", err)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_click") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		browserStringSliceContains(payload.BrowserTools, "browser_screenshot") ||
		len(payload.ArtifactTools) != 0 ||
		len(payload.ArtifactKinds) != 0 ||
		payload.ArtifactContract != "" ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") ||
		browserStringSliceContains(payload.BrowserActKinds, "screenshot") {
		t.Fatalf("expected degraded status diagnostics to use concrete managed opt-in capabilities, got %#v", payload)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected degraded status diagnostics to expose explicit managed opt-in route surface, got %#v", payload)
	}
	foundUnsupportedDefault := false
	for _, route := range payload.Routes {
		if route.Status != "unsupported" ||
			route.Backend != "custom-playwright" ||
			route.Profile != "workbench" ||
			route.RuntimeTarget != "host" {
			continue
		}
		foundUnsupportedDefault = true
		if !browserStringSliceContains(route.BrowserTools, "browser_click") ||
			!browserStringSliceContains(route.BrowserTools, "browser_act") ||
			browserStringSliceContains(route.BrowserTools, "browser_screenshot") ||
			len(route.ArtifactTools) != 0 ||
			len(route.ArtifactKinds) != 0 ||
			route.ArtifactContract != "" ||
			!browserStringSliceContains(route.BrowserActKinds, "click") ||
			browserStringSliceContains(route.BrowserActKinds, "screenshot") {
			t.Fatalf("expected unsupported default route row to use concrete managed opt-in capabilities, got %#v", route)
		}
		if route.BrowserSurface != "explicit_managed_opt_in" ||
			len(route.BrowserOptInTargets) != 1 ||
			route.BrowserOptInTargets[0] != "node" {
			t.Fatalf("expected unsupported default route row to expose explicit managed opt-in route surface, got %#v", route)
		}
	}
	if !foundUnsupportedDefault {
		t.Fatalf("expected route matrix to include unsupported default host row, got %#v", payload.Routes)
	}
	if nodeBackend.executionCalls == 0 {
		t.Fatalf("expected concrete opt-in diagnostics to resolve a node route")
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsHostRouteResolutionFailureForRequestedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-host-route-failure")
	sessionRegistry.TrackTabs("browser-runtime-sessions-host-route-failure", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	sessionRunRegistry.Record("browser-runtime-sessions-host-route-failure", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-host-1",
		NodeID:   "node-host",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-sessions-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route session profile",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		SessionTargetCount int      `json:"session_target_count"`
		SessionRoutes      []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Targets       []struct {
				ID string `json:"id"`
			} `json:"targets"`
		} `json:"session_routes"`
		SessionRuns []struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
		} `json:"session_runs"`
		SessionProfiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"session_profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected sessions diagnostics to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected degraded sessions diagnostics to reuse cached session routes, got %#v", payload)
	}
	if payload.SessionRoutes[0].Backend != "custom-playwright" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "host" || len(payload.SessionRoutes[0].Targets) != 1 {
		t.Fatalf("unexpected degraded session route payload: %#v", payload.SessionRoutes[0])
	}
	if len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-host-1" || payload.SessionRuns[0].Status != "running" {
		t.Fatalf("unexpected degraded session runs payload: %#v", payload.SessionRuns)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Backend != "custom-playwright" || payload.SessionProfiles[0].Profile != "workbench" || payload.SessionProfiles[0].RuntimeTarget != "host" || payload.SessionProfiles[0].Status != "running" || !payload.SessionProfiles[0].Running || !payload.SessionProfiles[0].Connected || payload.SessionProfiles[0].Note != "cached host route session profile" {
		t.Fatalf("unexpected degraded session profiles payload: %#v", payload.SessionProfiles)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReportsHostRouteResolutionFailureForRequestedProfileWithoutStateRegistry(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-host-route-failure-no-state-registry")
	sessionRegistry.TrackTabs("browser-runtime-sessions-host-route-failure-no-state-registry", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	sessionRunRegistry.Record("browser-runtime-sessions-host-route-failure-no-state-registry", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-host-no-state",
		NodeID:   "node-host",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            hostBackend,
		NodeBackend:        nodeBackend,
		SessionRegistry:    sessionRegistry,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions without state registry: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetCount int `json:"session_target_count"`
		SessionRoutes      []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_routes"`
		SessionRuns []struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
		} `json:"session_runs"`
		SessionProfiles []struct {
			Profile string `json:"profile"`
		} `json:"session_profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected requested-profile sessions diagnostics to degrade from session route snapshot without state registry, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "custom-playwright" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "host" {
		t.Fatalf("expected degraded sessions diagnostics to reuse session route snapshot without state registry, got %#v", payload)
	}
	if len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-host-no-state" || payload.SessionRuns[0].Status != "running" {
		t.Fatalf("unexpected degraded session runs payload without state registry: %#v", payload.SessionRuns)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Profile != "workbench" {
		t.Fatalf("expected degraded sessions diagnostics without state registry to synthesize route-scoped session profiles, got %#v", payload.SessionProfiles)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchReportsHostRouteResolutionFailureForRequestedProfileWithoutStateRegistry(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-host-route-failure-no-state-registry")
	sessionRegistry.TrackTabs("browser-runtime-workbench-host-route-failure-no-state-registry", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	sessionRunRegistry.Record("browser-runtime-workbench-host-route-failure-no-state-registry", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-host-workbench-no-state",
		NodeID:   "node-host",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            hostBackend,
		NodeBackend:        nodeBackend,
		SessionRegistry:    sessionRegistry,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench without state registry: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		WorkbenchReady     bool     `json:"workbench_ready"`
		WorkbenchSections  []string `json:"workbench_sections"`
		ProfileStatus      *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profile_status"`
		Profiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profiles"`
		SessionTargetCount int `json:"session_target_count"`
		SessionRuns        []struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
		} `json:"session_runs"`
		SessionProfiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Note          string `json:"note"`
		} `json:"session_profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" || !payload.WorkbenchReady {
		t.Fatalf("expected workbench diagnostics without state registry to stay ok, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench sections to include %q, got %#v", want, payload.WorkbenchSections)
		}
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Backend != "custom-playwright" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "host" || payload.ProfileStatus.BrowserApp != "chrome" || payload.ProfileStatus.Status != "" || payload.ProfileStatus.Running || payload.ProfileStatus.Connected || payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("unexpected degraded workbench profile status payload without state registry: %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Backend != "custom-playwright" || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].RuntimeTarget != "host" || payload.Profiles[0].BrowserApp != "chrome" || payload.Profiles[0].Status != "" || payload.Profiles[0].Running || payload.Profiles[0].Connected || payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("unexpected degraded workbench profiles payload without state registry: %#v", payload.Profiles)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-host-workbench-no-state" || payload.SessionRuns[0].Status != "running" {
		t.Fatalf("expected degraded workbench diagnostics without state registry to reuse route-scoped session snapshot, got %#v", payload)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Backend != "custom-playwright" || payload.SessionProfiles[0].Profile != "workbench" || payload.SessionProfiles[0].RuntimeTarget != "host" || payload.SessionProfiles[0].BrowserApp != "chrome" || payload.SessionProfiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("unexpected degraded workbench session profiles payload without state registry: %#v", payload.SessionProfiles)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchReportsHostRouteResolutionFailureForRequestedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-host-route-failure")
	sessionRegistry.TrackTabs("browser-runtime-workbench-host-route-failure", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chrome",
			Backend:    "custom-playwright",
			Profile:    "workbench",
			Target:     "host",
		},
	}, 1)
	sessionRunRegistry.Record("browser-runtime-workbench-host-route-failure", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-host-2",
		NodeID:   "node-host",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached host route workbench profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-workbench-host-route-failure", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "custom-playwright",
		Profile:       "workbench",
		RuntimeTarget: "host",
		BrowserApp:    "Chrome",
		Source:        "select_profile",
	})
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench: %v", err)
	}
	var payload struct {
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		WorkbenchReady     bool     `json:"workbench_ready"`
		WorkbenchSections  []string `json:"workbench_sections"`
		ProfileStatus      *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profile_status"`
		Profiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
		} `json:"profiles"`
		SessionTargetCount int `json:"session_target_count"`
		SessionRuns        []struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
		} `json:"session_runs"`
		SessionBinding *struct {
			ActiveNodeRunID              string `json:"active_node_run_id"`
			ActiveBrowserProfile         string `json:"active_browser_profile"`
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "ok" || !payload.WorkbenchReady {
		t.Fatalf("expected workbench diagnostics to stay ok when default host route cannot resolve, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "custom-playwright" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "host" {
		t.Fatalf("unexpected default route payload: %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected unresolved host default route to omit selected route, got %#v", payload.SelectedRoute)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected fallback default profile to stay on default host route, got %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected diagnostics payload note to surface route resolution failure, got %#v", payload)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench sections to include %q, got %#v", want, payload.WorkbenchSections)
		}
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Backend != "custom-playwright" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "host" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected || payload.ProfileStatus.Note != "cached host route workbench profile" {
		t.Fatalf("unexpected degraded workbench profile status payload: %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Backend != "custom-playwright" || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].RuntimeTarget != "host" || payload.Profiles[0].Status != "running" {
		t.Fatalf("unexpected degraded workbench profiles payload: %#v", payload.Profiles)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-host-2" || payload.SessionRuns[0].Status != "running" {
		t.Fatalf("expected degraded workbench diagnostics to reuse cached session snapshot, got %#v", payload)
	}
	if payload.SessionBinding == nil {
		t.Fatalf("expected degraded workbench diagnostics to reuse cached session binding, got %#v", payload)
	}
	if payload.SessionBinding.ActiveNodeRunID != "run-host-2" || payload.SessionBinding.ActiveBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserProfileSource != "select_profile" || payload.SessionBinding.SelectedBrowserTargetSource != "tracked_active_tab" {
		t.Fatalf("unexpected degraded workbench session binding payload: %#v", payload.SessionBinding)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected configured profiles to retain requested/default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(nodeBackend.runtimeStatusReqs) != 0 || len(nodeBackend.runtimeProfilesReqs) != 0 {
		t.Fatalf("expected degraded workbench diagnostics to avoid unrelated managed runtime calls, got status=%#v profiles=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeProfilesReqs)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserDelegatesActAction(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"` + publicExampleIPURL + `","wait_ms":700}`,
	})
	if err != nil {
		t.Fatalf("browser open: %v", err)
	}
	if len(backend.openReqs) != 1 || backend.openReqs[0].URL != publicExampleIPURL || backend.openReqs[0].WaitMs != 700 {
		t.Fatalf("unexpected browser open dispatch: %#v", backend.openReqs)
	}
	var payload struct {
		Kind    string `json:"kind"`
		Backend string `json:"backend"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "open" || payload.Backend != "fake-open" || payload.Status != "opened" {
		t.Fatalf("unexpected browser open output: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserOpenTracksCurrentTargetForLaterExtract(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "system_open", Status: "opened"},
			extractResult: BrowserExtractResult{
				Backend:     "fake-extract",
				Title:       "Opened Page",
				Content:     "tracked current content",
				FinalURL:    publicExampleIPURL,
				ContentType: "text/plain",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-unified-open-track")
	openOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser unified open tracked current target: %v", err)
	}
	var openPayload struct {
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(openOut), &openPayload); err != nil {
		t.Fatalf("decode unified open output: %v", err)
	}
	if openPayload.Target != "current" || strings.TrimSpace(openPayload.TargetID) == "" {
		t.Fatalf("expected tracked current target from unified browser open, got %#v", openPayload)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"extract","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser unified extract current after open: %v", err)
	}
	assertBrowserUnifiedOutputKeysCoveredBySchema(t, extractOut)
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].URL != publicExampleIPURL || backend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser extract to reuse tracked open url, got %#v", backend.extractReqs)
	}
	var extractPayload struct {
		FinalURL string `json:"final_url"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(extractOut), &extractPayload); err != nil {
		t.Fatalf("decode unified extract output: %v", err)
	}
	if extractPayload.FinalURL != publicExampleIPURL || strings.TrimSpace(extractPayload.TargetID) == "" {
		t.Fatalf("unexpected unified extract payload after open tracking: %#v", extractPayload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserCurrentTargetSurvivesFallbackReads(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "system_open", Status: "opened"},
			extractResult: BrowserExtractResult{
				Backend:     "http_extract_fallback",
				Title:       "Opened Page",
				Content:     "tracked current content",
				FinalURL:    publicExampleIPURL,
				ContentType: "text/plain",
			},
			snapshotResult: BrowserSnapshotResult{
				Backend:  "http_snapshot_fallback",
				Title:    "Opened Page",
				Snapshot: "tracked current snapshot",
				FinalURL: publicExampleIPURL,
				Format:   "ai",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-unified-current-fallback")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"` + publicExampleIPURL + `"}`,
	}); err != nil {
		t.Fatalf("browser unified open tracked current target: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"extract","max_chars":32}`,
	}); err != nil {
		t.Fatalf("browser unified extract current after open: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"extract","max_chars":48}`,
	}); err != nil {
		t.Fatalf("browser unified repeated extract after fallback: %v", err)
	}
	if len(backend.extractReqs) != 2 || backend.extractReqs[0].URL != publicExampleIPURL || backend.extractReqs[1].URL != publicExampleIPURL {
		t.Fatalf("expected repeated unified browser extracts to reuse tracked current url, got %#v", backend.extractReqs)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"snapshot","format":"ai","mode":"efficient"}`,
	}); err != nil {
		t.Fatalf("browser unified snapshot after fallback extract: %v", err)
	}
	if len(backend.snapshotReqs) != 1 || backend.snapshotReqs[0].URL != publicExampleIPURL || backend.snapshotReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser snapshot to reuse tracked current url after fallback extract, got %#v", backend.snapshotReqs)
	}
}

func TestRegisterBrowserTools_RuntimeStatusIncludesTrackedSessionBinding(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-bound-session")
	tracked := sessionRegistry.TrackTab("browser-runtime-bound-session", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/console",
		Title:      "Node Console",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime tracked session binding: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			SessionKey        string `json:"session_key"`
			CurrentTargetID   string `json:"current_target_id"`
			RouteTargetCount  int    `json:"route_target_count"`
			PropagatedToProxy bool   `json:"propagated_to_proxy"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.SessionKey != "browser-runtime-bound-session" || payload.SessionBinding.CurrentTargetID != tracked.ID || payload.SessionBinding.RouteTargetCount != 1 || !payload.SessionBinding.PropagatedToProxy {
		t.Fatalf("unexpected tracked session binding payload: %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeStatusInvalidatesManagedCurrentTargetSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-invalidate-managed-target")
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
	sessionRegistry.TrackTab("browser-runtime-status-invalidate-managed-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/console",
		Title:      "Node Console",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "disconnected",
				Running:    true,
				Connected:  false,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status invalidates managed target: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			CurrentTargetID  string `json:"current_target_id"`
			RouteTargetCount int    `json:"route_target_count"`
		} `json:"session_binding"`
		ProfileStatus struct {
			Status string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.ProfileStatus.Status != "disconnected" {
		t.Fatalf("expected disconnected profile status, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.CurrentTargetID != "" || payload.SessionBinding.RouteTargetCount != 1 {
		t.Fatalf("expected tracked target to remain but current selection to be cleared, got %#v", payload.SessionBinding)
	}
	if _, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-status-invalidate-managed-target", route); ok {
		t.Fatalf("expected managed current target selection to be invalidated after status watch")
	}
	if _, ok := sessionRegistry.ResolveTabForRoute("browser-runtime-status-invalidate-managed-target", route, 2); !ok {
		t.Fatalf("expected tracked tab to remain after current target invalidation")
	}
}

func TestRegisterBrowserTools_RuntimeStatusReportsUnsupportedRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime: %v", err)
	}
	var payload struct {
		Status string `json:"status"`
		Note   string `json:"note"`
		Routes []struct {
			Status string `json:"status"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "unsupported" || !strings.Contains(payload.Note, `runtime_target "node" is unsupported`) {
		t.Fatalf("expected structured unsupported runtime response, got %#v", payload)
	}
	if len(payload.Routes) == 0 {
		t.Fatalf("expected route matrix when runtime route is unsupported, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeRestartFromStopped(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart from stopped: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("unexpected restart-from-stopped lifecycle requests: status=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStartReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 0 {
		t.Fatalf("expected no stop call for already stopped profile, got %#v", nodeBackend.runtimeStopReqs)
	}
	if !strings.Contains(out, `"restart_decision":"restart_started"`) || !strings.Contains(out, `"action":"restart"`) {
		t.Fatalf("unexpected runtime restart-from-stopped output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeRestartBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-blocked")
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-restart-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-91",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:        nodeBackend,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart blocked: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 0 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected blocked restart to avoid runtime lifecycle calls, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	if !strings.Contains(out, `"restart_decision":"restart_blocked_active_node_run"`) {
		t.Fatalf("unexpected runtime restart-blocked output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeRestartForceBypassesActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-force")
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-restart-force", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-92",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:        nodeBackend,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart force: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected forced restart to execute lifecycle calls, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	if !strings.Contains(out, `"force":true`) || !strings.Contains(out, `"restart_decision":"restarted"`) {
		t.Fatalf("unexpected runtime restart-force output: %s", out)
	}
}
