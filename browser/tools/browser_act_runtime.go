package tools

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
)

var fullBrowserActKinds = []string{
	"open",
	"navigate",
	"extract",
	"snapshot",
	"screenshot",
	"console",
	"requests",
	"response_body",
	"errors",
	"cookies",
	"cookies_set",
	"cookies_clear",
	"storage",
	"storage_set",
	"storage_clear",
	"offline",
	"headers",
	"credentials",
	"geolocation",
	"media",
	"timezone",
	"locale",
	"device",
	"highlight",
	"trace_start",
	"trace_stop",
	"press",
	"hover",
	"drag",
	"select",
	"fill",
	"resize",
	"click",
	"type",
	"evaluate",
	"wait",
	"list_tabs",
	"focus_tab",
	"close_tab",
}

const defaultBrowserWaitDownloadMs = 120_000

func browserActKindsForRegistration(opts BrowserToolOptions) []string {
	return append([]string(nil), browserCapabilitiesForRegistration(opts).SupportedActKinds()...)
}

func browserActDefinition(kinds []string) types.Tool {
	if len(kinds) == 0 {
		kinds = append([]string(nil), fullBrowserActKinds...)
	}
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "browser_act",
			Description: "Specialist companion to the unified `browser` workbench for raw page/action execution. Use this when you explicitly want the browser_act tool name or need to isolate action-level browser work from the broader unified browser entrypoint. For title/text/summary requests, prefer unified `browser action=extract`; reserve `kind=snapshot` for element refs or structural grounding before follow-up actions. Supported kinds: " + strings.Join(kinds, ", ") + ". Prefer target=current, target=tab:N, or a prior target_id via target:<id> to stay on the same tab across actions.",
			Parameters: browserDescribedInputSchema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":             map[string]any{"type": "string", "enum": append([]string(nil), kinds...)},
					"url":              map[string]any{"type": "string"},
					"target":           map[string]any{"type": "string"},
					"tab_index":        map[string]any{"type": "integer", "minimum": 1},
					"index":            map[string]any{"type": "integer", "minimum": 1},
					"path":             map[string]any{"type": "string"},
					"paths":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"files":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"file":             map[string]any{"type": "string"},
					"output":           map[string]any{"type": "string"},
					"output_path":      map[string]any{"type": "string"},
					"landscape":        map[string]any{"type": "boolean"},
					"print_background": map[string]any{"type": "boolean"},
					"action":           map[string]any{"type": "string", "enum": []string{"accept", "dismiss"}},
					"dialog_action":    map[string]any{"type": "string", "enum": []string{"accept", "dismiss"}},
					"accept":           map[string]any{"type": "boolean"},
					"dismiss":          map[string]any{"type": "boolean"},
					"prompt":           map[string]any{"type": "string"},
					"prompt_text":      map[string]any{"type": "string"},
					"ref":              map[string]any{"type": "string"},
					"input_ref":        map[string]any{"type": "string"},
					"element_ref":      map[string]any{"type": "string"},
					"selector":         map[string]any{"type": "string"},
					"frame":            map[string]any{"type": "string"},
					"element":          map[string]any{"type": "string"},
					"label":            map[string]any{"type": "string"},
					"format":           map[string]any{"type": "string", "enum": []string{"ai", "aria"}},
					"snapshot_format":  map[string]any{"type": "string", "enum": []string{"ai", "aria"}},
					"mode":             map[string]any{"type": "string", "enum": []string{"efficient"}},
					"refs":             map[string]any{"type": "string", "enum": []string{"aria", "role"}},
					"interactive":      map[string]any{"type": "boolean"},
					"compact":          map[string]any{"type": "boolean"},
					"efficient":        map[string]any{"type": "boolean"},
					"depth":            map[string]any{"type": "integer", "minimum": 1},
					"full_page":        map[string]any{"type": "boolean"},
					"level":            map[string]any{"type": "string"},
					"filter":           map[string]any{"type": "string"},
					"request_url":      map[string]any{"type": "string"},
					"response_url":     map[string]any{"type": "string"},
					"name":             map[string]any{"type": "string"},
					"key":              map[string]any{"type": "string"},
					"values":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"domain":           map[string]any{"type": "string"},
					"same_site":        map[string]any{"type": "string"},
					"http_only":        map[string]any{"type": "boolean"},
					"secure":           map[string]any{"type": "boolean"},
					"expires":          map[string]any{"type": "integer"},
					"width":            map[string]any{"type": "integer", "minimum": 1},
					"height":           map[string]any{"type": "integer", "minimum": 1},
					"storage_kind":     map[string]any{"type": "string", "enum": []string{"local", "session"}},
					"enabled":          map[string]any{"type": "boolean"},
					"clear":            map[string]any{"type": "boolean"},
					"headers_json":     map[string]any{"type": "string"},
					"username":         map[string]any{"type": "string"},
					"password":         map[string]any{"type": "string"},
					"origin":           map[string]any{"type": "string"},
					"latitude":         map[string]any{"type": "number"},
					"longitude":        map[string]any{"type": "number"},
					"accuracy":         map[string]any{"type": "number", "minimum": 0},
					"media":            map[string]any{"type": "string", "enum": []string{"dark", "light", "no-preference", "none"}},
					"timezone":         map[string]any{"type": "string"},
					"locale":           map[string]any{"type": "string"},
					"device":           map[string]any{"type": "string"},
					"fields": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"ref":         map[string]any{"type": "string"},
								"element_ref": map[string]any{"type": "string"},
								"input_ref":   map[string]any{"type": "string"},
								"selector":    map[string]any{"type": "string"},
								"type":        map[string]any{"type": "string"},
								"value":       map[string]any{"type": "string"},
								"values":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							},
						},
					},
					"cookies": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":      map[string]any{"type": "string"},
								"value":     map[string]any{"type": "string"},
								"domain":    map[string]any{"type": "string"},
								"path":      map[string]any{"type": "string"},
								"same_site": map[string]any{"type": "string"},
								"http_only": map[string]any{"type": "boolean"},
								"secure":    map[string]any{"type": "boolean"},
								"expires":   map[string]any{"type": "integer"},
							},
						},
					},
					"entries": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"key":   map[string]any{"type": "string"},
								"value": map[string]any{"type": "string"},
							},
						},
					},
					"headers": map[string]any{
						"type":                 "object",
						"additionalProperties": map[string]any{"type": "string"},
					},
					"text":           map[string]any{"type": "string"},
					"value":          map[string]any{"type": "string"},
					"delay_ms":       map[string]any{"type": "integer", "minimum": 0},
					"start_ref":      map[string]any{"type": "string"},
					"end_ref":        map[string]any{"type": "string"},
					"start_selector": map[string]any{"type": "string"},
					"end_selector":   map[string]any{"type": "string"},
					"start_element":  map[string]any{"type": "string"},
					"end_element":    map[string]any{"type": "string"},
					"start_label":    map[string]any{"type": "string"},
					"end_label":      map[string]any{"type": "string"},
					"from": map[string]any{
						"anyOf": []map[string]any{
							{"type": "string"},
							{
								"type": "object",
								"properties": map[string]any{
									"selector": map[string]any{"type": "string"},
									"ref":      map[string]any{"type": "string"},
									"element":  map[string]any{"type": "string"},
									"label":    map[string]any{"type": "string"},
								},
							},
						},
					},
					"to": map[string]any{
						"anyOf": []map[string]any{
							{"type": "string"},
							{
								"type": "object",
								"properties": map[string]any{
									"selector": map[string]any{"type": "string"},
									"ref":      map[string]any{"type": "string"},
									"element":  map[string]any{"type": "string"},
									"label":    map[string]any{"type": "string"},
								},
							},
						},
					},
					"submit":         map[string]any{"type": "boolean"},
					"script":         map[string]any{"type": "string"},
					"javascript":     map[string]any{"type": "string"},
					"js":             map[string]any{"type": "string"},
					"browser":        map[string]any{"type": "string"},
					"browser_app":    map[string]any{"type": "string"},
					"profile":        browserRuntimeProfileSchema(),
					"runtime_target": browserRuntimeTargetSchema(),
					"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
					"post_wait_ms":   map[string]any{"type": "integer", "minimum": 0},
					"max_chars":      map[string]any{"type": "integer", "minimum": 64},
					"max_elements":   map[string]any{"type": "integer", "minimum": 1},
					"force":          map[string]any{"type": "boolean"},
				},
				"required": []string{"kind"},
			}),
		},
	}
}

func defaultBrowserScreenshotRelPath() string {
	ts := browserNow().UTC().Format("20060102-150405.000")
	return filepath.ToSlash(filepath.Join(".agentx", "browser", fmt.Sprintf("screenshot-%s.png", ts)))
}

func defaultBrowserPDFRelPath() string {
	ts := browserNow().UTC().Format("20060102-150405.000")
	return filepath.ToSlash(filepath.Join(".agentx", "browser", fmt.Sprintf("page-%s.pdf", ts)))
}

func defaultBrowserHTMLRelPath() string {
	ts := browserNow().UTC().Format("20060102-150405.000")
	return filepath.ToSlash(filepath.Join(".agentx", "browser", fmt.Sprintf("page-%s.html", ts)))
}

func defaultBrowserTraceRelPath() string {
	ts := browserNow().UTC().Format("20060102-150405.000")
	return filepath.ToSlash(filepath.Join(".agentx", "browser", fmt.Sprintf("trace-%s.zip", ts)))
}

func defaultBrowserDownloadRelPath(hints ...string) string {
	ts := browserNow().UTC().Format("20060102-150405.000")
	ext := browserArtifactExtHint(hints...)
	return filepath.ToSlash(filepath.Join(".agentx", "browser", fmt.Sprintf("download-%s%s", ts, ext)))
}

func browserArtifactExtHint(hints ...string) string {
	for _, item := range hints {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if parsed, err := neturl.Parse(value); err == nil && parsed.Path != "" {
			value = parsed.Path
		}
		ext := strings.ToLower(strings.TrimSpace(path.Ext(value)))
		if ext == "" || len(ext) > 12 {
			continue
		}
		valid := true
		for _, ch := range ext[1:] {
			if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
				valid = false
				break
			}
		}
		if valid {
			return ext
		}
	}
	return ""
}

func resolveBrowserUploadPaths(root string, params map[string]any) ([]string, []string, error) {
	inputs := make([]string, 0, 4)
	inputs = append(inputs, readStringList(params, "paths")...)
	inputs = append(inputs, readStringList(params, "files")...)
	if item := strings.TrimSpace(firstString(params, "file")); item != "" {
		inputs = append(inputs, item)
	}
	if len(inputs) == 0 {
		repairable, safeAutorepair, repairs := browserUploadPathRepairAdviceFromParams(params, "path", "filepath")
		return nil, nil, browserMissingUploadPathsErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
	}
	resolved := make([]string, 0, len(inputs))
	display := make([]string, 0, len(inputs))
	seenResolved := map[string]bool{}
	for _, item := range inputs {
		pathValue := strings.TrimSpace(item)
		if pathValue == "" {
			continue
		}
		resolvedPath, displayPath, err := resolvePathWithinRoot(root, pathValue)
		if err != nil {
			return nil, nil, fmt.Errorf("browser_act: invalid upload path %q: %w", pathValue, err)
		}
		info, err := os.Stat(resolvedPath)
		if err != nil {
			return nil, nil, fmt.Errorf("browser_act: upload path %q: %w", pathValue, err)
		}
		if info.IsDir() {
			return nil, nil, fmt.Errorf("browser_act: upload path %q is a directory", pathValue)
		}
		if seenResolved[resolvedPath] {
			continue
		}
		seenResolved[resolvedPath] = true
		resolved = append(resolved, resolvedPath)
		display = append(display, displayPath)
	}
	if len(resolved) == 0 {
		return nil, nil, browserMissingUploadPathsErrorWithRepair("browser_act", false, false, nil)
	}
	return resolved, display, nil
}

func resolveBrowserDialogAction(params map[string]any) (string, error) {
	action := strings.ToLower(strings.TrimSpace(firstString(params, "dialog_action", "action")))
	accept := firstBool(params, "accept")
	dismiss := firstBool(params, "dismiss")
	switch {
	case accept && dismiss:
		return "", browserInvalidArgumentError("browser_act", []string{"accept", "dismiss"}, "browser_act: accept and dismiss cannot both be true for kind dialog")
	case accept:
		if action != "" && action != "accept" {
			return "", browserInvalidArgumentError("browser_act", []string{"action", "accept"}, fmt.Sprintf("browser_act: action %q conflicts with accept=true for kind dialog", action))
		}
		return "accept", nil
	case dismiss:
		if action != "" && action != "dismiss" {
			return "", browserInvalidArgumentError("browser_act", []string{"action", "dismiss"}, fmt.Sprintf("browser_act: action %q conflicts with dismiss=true for kind dialog", action))
		}
		return "dismiss", nil
	}
	switch action {
	case "accept", "dismiss":
		return action, nil
	case "":
		return "", browserMissingRequiredArgumentError("browser_act", []string{"action", "accept", "dismiss"}, "browser_act: action or accept/dismiss is required for kind dialog")
	default:
		return "", browserInvalidArgumentError("browser_act", []string{"action"}, "browser_act: action must be accept or dismiss for kind dialog")
	}
}

func browserActReviewBlockedResult(kind string, runtimeInfo BrowserRuntimeInfo, browserApp string, target browserToolTarget, force bool, note string) BrowserActResult {
	return BrowserActResult{
		Kind:           kind,
		Backend:        strings.TrimSpace(runtimeInfo.Backend),
		BrowserApp:     strings.TrimSpace(browserApp),
		Target:         strings.TrimSpace(target.Value),
		TargetID:       strings.TrimSpace(target.TargetID),
		Profile:        strings.TrimSpace(runtimeInfo.Profile),
		RuntimeTarget:  strings.TrimSpace(runtimeInfo.Target),
		Status:         "review_required",
		Force:          force,
		ReviewDecision: agentxbrowserruntime.SharedSessionBrowserActionReviewDecision(kind, force),
		ReviewReady:    false,
		TabIndex:       target.TabIndex,
		Note:           strings.TrimSpace(note),
	}
}

func browserActUsesPageActionManagedDefaultDispatchBase(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "extract", "snapshot", "screenshot", "click", "type", "evaluate":
		return true
	default:
		return false
	}
}

type browserActDispatch struct {
	Route                         browserResolvedExecutionRoute
	Backend                       BrowserBackend
	RuntimeInfo                   BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	RequestedBrowserApp           string
	ExplicitRuntimeTarget         bool
	Target                        browserToolTarget
	BrowserApp                    string
}

type browserActPageReviewState struct {
	Review   browserPendingTargetReviewState
	TargetID string
}

type browserActPageActionEventApplyOptions struct {
	ResultBackend     string
	ResultBrowserApp  string
	ResultFinalURL    string
	ResultTitle       string
	Source            string
	Actor             string
	Force             bool
	TabIndex          int
	PreferredTargetID string
	Review            browserPendingTargetReviewState
}

func resolveBrowserActDispatch(
	ctx context.Context,
	backend BrowserBackend,
	sessionRegistry *BrowserSessionRegistry,
	sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	baseRuntimeInfo BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	opts BrowserToolOptions,
	maxChars int,
	params map[string]any,
	kind string,
) (browserActDispatch, error) {
	dispatchBase := baseRuntimeInfo
	if browserActUsesPageActionManagedDefaultDispatchBase(kind) {
		dispatchBase = browserPageActionDispatchBaseForManagedDefaultParams(dispatchBase, hiddenImplicitHostDefaultBase, params)
	}
	explicitRuntimeTarget := browserHasExplicitRuntimeTarget(params)
	route, err := resolveBrowserManagedRouteForSession(
		ctx,
		sessionStateRegistry,
		sessionRegistry,
		watchManagerProvider,
		params,
		dispatchBase,
		hiddenImplicitHostDefaultBase,
		backend,
		maxChars,
	)
	if err != nil {
		return browserActDispatch{}, fmt.Errorf("browser_act: %w", err)
	}
	routedBackend := route.Backend
	runtimeInfo := route.RuntimeInfo
	hiddenImplicitHostDefaultBase = route.hiddenImplicitHostDefaultBase
	if err := browserImplicitLegacyHostManagedActKindFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, kind); err != nil {
		return browserActDispatch{}, err
	}
	if !route.Capabilities.SupportsActKind(kind) {
		return browserActDispatch{}, fmt.Errorf(
			"browser_act: unsupported kind %q for runtime_target=%q backend=%q",
			kind,
			strings.TrimSpace(runtimeInfo.Target),
			browserRuntimeBackendName(runtimeInfo),
		)
	}
	requestedBrowserApp := firstNonEmpty(strings.TrimSpace(firstString(params, "browser", "browser_app", "app")), strings.TrimSpace(opts.DefaultBrowserApp))
	target, err := resolveBrowserToolTarget(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, requestedBrowserApp, params)
	if err != nil {
		return browserActDispatch{}, fmt.Errorf("browser_act: %w", err)
	}
	browserApp := browserEffectiveBrowserApp(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, requestedBrowserApp, target)
	return browserActDispatch{
		Route:                         route,
		Backend:                       routedBackend,
		RuntimeInfo:                   runtimeInfo,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		RequestedBrowserApp:           requestedBrowserApp,
		ExplicitRuntimeTarget:         explicitRuntimeTarget,
		Target:                        target,
		BrowserApp:                    browserApp,
	}, nil
}

func resolveBrowserActPageReviewState(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runtimeInfo BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	browserApp string,
	target browserToolTarget,
	requestURL string,
) browserActPageReviewState {
	review := browserPendingTargetReviewStateForRuntimeTarget(ctx, sessionRegistry, runtimeInfo, browserApp, target)
	if strings.TrimSpace(requestURL) == "" {
		review = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
	}
	targetID := strings.TrimSpace(target.TargetID)
	if review.Review != nil {
		targetID = firstNonEmpty(targetID, strings.TrimSpace(review.Review.ID))
	}
	return browserActPageReviewState{
		Review:   review,
		TargetID: targetID,
	}
}

func browserActPageReviewBlockedResult(kind string, runtimeInfo BrowserRuntimeInfo, browserApp string, target browserToolTarget, reviewState browserActPageReviewState, force bool, note string) BrowserActResult {
	result := BrowserActResult{
		Kind:           kind,
		Backend:        strings.TrimSpace(runtimeInfo.Backend),
		BrowserApp:     strings.TrimSpace(browserApp),
		Target:         strings.TrimSpace(target.Value),
		TargetID:       strings.TrimSpace(reviewState.TargetID),
		Profile:        strings.TrimSpace(runtimeInfo.Profile),
		RuntimeTarget:  strings.TrimSpace(runtimeInfo.Target),
		Status:         "review_required",
		Force:          force,
		ReviewDecision: browserPendingTargetReviewDecisionWithState(reviewState.Review, force),
		ReviewReady:    false,
		TabIndex:       target.TabIndex,
		Note:           strings.TrimSpace(note),
	}
	return result
}

func applyBrowserActPageActionEventResult(
	sharedMutationCtx agentxbrowserruntime.SharedSessionBrowserMutationContext,
	sessionID string,
	runtimeInfo BrowserRuntimeInfo,
	defaultBrowserApp string,
	defaultFinalURL string,
	options browserActPageActionEventApplyOptions,
) agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult {
	return agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
		sharedMutationCtx,
		agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
			SessionID:         sessionID,
			Route:             browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(options.ResultBrowserApp), defaultBrowserApp), strings.TrimSpace(options.ResultBackend)),
			PreferredTargetID: strings.TrimSpace(options.PreferredTargetID),
			TabIndex:          options.TabIndex,
			URL:               firstNonEmpty(strings.TrimSpace(options.ResultFinalURL), defaultFinalURL),
			Title:             strings.TrimSpace(options.ResultTitle),
			Source:            strings.TrimSpace(options.Source),
			Actor:             strings.TrimSpace(options.Actor),
			Force:             options.Force,
			Review:            options.Review,
		},
	)
}

func browserActPageActionResultNote(resultNote string, pageActionNote string, review browserPendingTargetReviewState) string {
	if review.Review != nil {
		return firstNonEmpty(strings.TrimSpace(resultNote), strings.TrimSpace(pageActionNote))
	}
	return strings.TrimSpace(resultNote)
}

func executeBrowserAct(ctx context.Context, backend BrowserBackend, sessionRegistry *BrowserSessionRegistry, sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, baseRuntimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, policy outboundNetworkPolicy, opts BrowserToolOptions, defaultWaitMs int, maxChars int, params map[string]any) (BrowserActResult, error) {
	watchManagerProvider := agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(
		sessionRegistry,
		opts.SessionRunRegistry,
		sessionStateRegistry,
		browserRuntimeReconnectWatchdogWindow,
	)
	kind := browserNormalizeToolToken(firstString(params, "kind"))
	if kind == "" {
		return BrowserActResult{}, browserMissingActKindError("browser_act", "")
	}
	if current := firstString(params, "kind"); current != kind {
		params = browserUnifiedCloneParams(params)
		params["kind"] = kind
	}
	lane, dispatch, err := resolveBrowserExecutionLaneForActDispatch(
		ctx,
		backend,
		sessionRegistry,
		sessionStateRegistry,
		watchManagerProvider,
		baseRuntimeInfo,
		hiddenImplicitHostDefaultBase,
		opts,
		maxChars,
		params,
		kind,
	)
	if err != nil {
		return BrowserActResult{}, err
	}
	return executeBrowserActResolved(
		ctx,
		sessionRegistry,
		sessionStateRegistry,
		policy,
		opts,
		defaultWaitMs,
		maxChars,
		params,
		lane,
		dispatch,
	)
}

func executeBrowserActResolved(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	policy outboundNetworkPolicy,
	opts BrowserToolOptions,
	defaultWaitMs int,
	maxChars int,
	params map[string]any,
	lane BrowserExecutionLane,
	dispatch browserActDispatch,
) (BrowserActResult, error) {
	effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(ctx, policy)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: invalid runtime network policy: %w", err)
	}
	watchManagerProvider := agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(
		sessionRegistry,
		opts.SessionRunRegistry,
		sessionStateRegistry,
		browserRuntimeReconnectWatchdogWindow,
	)
	kind := browserNormalizeToolToken(firstString(params, "kind"))
	if kind == "" {
		return BrowserActResult{}, browserMissingActKindError("browser_act", "")
	}
	if current := firstString(params, "kind"); current != kind {
		params = browserUnifiedCloneParams(params)
		params["kind"] = kind
	}
	route := dispatch.Route
	routedBackend := lane.Backend
	runtimeInfo := lane.Runtime
	hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
	explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
	target := dispatch.Target
	browserApp := dispatch.BrowserApp
	force := firstBool(params, "force")
	sharedMutationCtx := browserSharedMutationContext(watchManagerProvider, sessionRegistry)
	sessionID := ToolSessionIDFromContext(ctx)
	browserApplyResolvedTargetRuntimeResult := func(
		_ context.Context,
		_ agentxbrowserruntime.SharedSessionBrowserObserverManager,
		_ *BrowserSessionRegistry,
		runtimeInfo BrowserRuntimeInfo,
		browserApp string,
		backend string,
		preferredTargetID string,
		tabIndex int,
		finalURL string,
		title string,
		source string,
	) agentxbrowserruntime.SharedSessionBrowserResolvedTargetEventResult {
		return agentxbrowserruntime.ApplySharedSessionBrowserResolvedTargetWithContext(
			sharedMutationCtx,
			agentxbrowserruntime.SharedSessionBrowserResolvedTargetEventRequest{
				SessionID:        sessionID,
				Route:            browserSessionRoute(runtimeInfo, browserApp, backend),
				ExplicitTargetID: strings.TrimSpace(preferredTargetID),
				TabIndex:         tabIndex,
				URL:              strings.TrimSpace(finalURL),
				Title:            strings.TrimSpace(title),
				Source:           source,
			},
		)
	}
	browserApplyActionRuntimeResult := func(
		_ context.Context,
		_ agentxbrowserruntime.SharedSessionBrowserObserverManager,
		_ *BrowserSessionRegistry,
		runtimeInfo BrowserRuntimeInfo,
		browserApp string,
		backend string,
		preferredTargetID string,
		tabIndex int,
		trackCurrent bool,
		finalURL string,
		title string,
		source string,
		setCurrent bool,
		reviewDecision string,
		reviewReady bool,
		note string,
	) agentxbrowserruntime.SharedSessionBrowserActionResultEventResult {
		return agentxbrowserruntime.ApplySharedSessionBrowserActionResultWithContext(
			sharedMutationCtx,
			agentxbrowserruntime.SharedSessionBrowserActionResultEventRequest{
				SessionID:         sessionID,
				Route:             browserSessionRoute(runtimeInfo, browserApp, backend),
				PreferredTargetID: strings.TrimSpace(preferredTargetID),
				TabIndex:          tabIndex,
				TrackCurrent:      trackCurrent,
				URL:               strings.TrimSpace(finalURL),
				Title:             strings.TrimSpace(title),
				Source:            source,
				SetCurrent:        setCurrent,
				ReviewDecision:    strings.TrimSpace(reviewDecision),
				ReviewReady:       reviewReady,
				Note:              strings.TrimSpace(note),
			},
		)
	}
	reqURL := strings.TrimSpace(firstString(params, "url"))
	if reqURL != "" {
		parsed, err := effectivePolicy.validateURL(ctx, reqURL)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		reqURL = parsed.String()
	} else if target.TabIndex <= 0 {
		if trackedURL := browserResolvedTargetURL(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target); trackedURL != "" {
			parsed, err := effectivePolicy.validateURL(ctx, trackedURL)
			if err != nil {
				return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
			}
			reqURL = parsed.String()
		}
	}
	waitMs := firstInt(params, "wait_ms")
	postWaitMs := firstInt(params, "post_wait_ms")
	tabIndex := target.TabIndex
	pageActionCtx := browserActPageActionContext{
		CallCtx:                       ctx,
		Route:                         route,
		RoutedBackend:                 routedBackend,
		SessionRegistry:               sessionRegistry,
		SharedMutationCtx:             sharedMutationCtx,
		SessionID:                     sessionID,
		RuntimeInfo:                   runtimeInfo,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		ExplicitRuntimeTarget:         explicitRuntimeTarget,
		Target:                        target,
		BrowserApp:                    browserApp,
		RequestURL:                    reqURL,
		Force:                         force,
		TabIndex:                      tabIndex,
		DefaultWaitMs:                 defaultWaitMs,
		MaxChars:                      maxChars,
		Options:                       opts,
	}
	tabsCtx := browserActTabsContext{
		CallCtx:                       ctx,
		RoutedBackend:                 routedBackend,
		SessionRegistry:               sessionRegistry,
		SharedMutationCtx:             sharedMutationCtx,
		SessionID:                     sessionID,
		RuntimeInfo:                   runtimeInfo,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		Target:                        target,
		BrowserApp:                    browserApp,
		Force:                         force,
		TabIndex:                      tabIndex,
	}
	switch kind {
	case "open":
		if reqURL == "" {
			return BrowserActResult{}, browserMissingURLForKindError("browser_act", "open")
		}
		if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_act kind open", hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, browserToolTarget{}, reqURL); err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 {
			waitMs = defaultWaitMs
		}
		result, err := routedBackend.Open(ctx, BrowserOpenRequest{
			URL:        reqURL,
			BrowserApp: browserApp,
			WaitMs:     waitMs,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		resolvedBrowserApp := firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp)
		targetID := agentxbrowserruntime.ApplySharedSessionBrowserOpenResultWithContext(
			agentxbrowserruntime.SharedSessionBrowserMutationContext{
				Registry:        sessionRegistry,
				RunRegistry:     watchManagerProvider.RunRegistry,
				StateRegistry:   watchManagerProvider.StateRegistry,
				ReconnectWindow: browserRuntimeReconnectWatchdogWindow,
			},
			agentxbrowserruntime.SharedSessionBrowserOpenResultEventRequest{
				SessionID: ToolSessionIDFromContext(ctx),
				Route:     browserSessionRoute(runtimeInfo, resolvedBrowserApp, strings.TrimSpace(result.Backend)),
				URL:       reqURL,
				Source:    "browser_act_open",
			},
		).TargetID
		return BrowserActResult{
			Kind:       kind,
			Backend:    result.Backend,
			BrowserApp: resolvedBrowserApp,
			Target: func() string {
				if strings.TrimSpace(targetID) == "" {
					return ""
				}
				return "current"
			}(),
			TargetID:      targetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			Status:        firstNonEmpty(result.Status, "opened"),
			Note:          result.Note,
		}, nil
	case "navigate":
		if reqURL == "" {
			return BrowserActResult{}, browserMissingURLForKindError("browser_act", "navigate")
		}
		if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_act kind navigate", hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, reqURL); err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 {
			waitMs = defaultWaitMs
		}
		result, err := routedBackend.Navigate(ctx, BrowserNavigateRequest{
			URL:              reqURL,
			BrowserApp:       browserApp,
			WaitMs:           waitMs,
			TabIndex:         tabIndex,
			Force:            force,
			ExplicitTargetID: strings.TrimSpace(target.TargetID),
			PriorSelection:   browserCurrentTargetSelectionSnapshotForRoute(ctx, sessionRegistry, runtimeInfo, browserApp, strings.TrimSpace(runtimeInfo.Backend)),
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), reqURL)
		finalURL, err = browserValidateFinalURL(ctx, effectivePolicy, finalURL)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		resolvedBrowserApp := firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp)
		navigationResult := agentxbrowserruntime.ApplySharedSessionBrowserNavigationResultWithContext(
			agentxbrowserruntime.SharedSessionBrowserMutationContext{
				Registry:        sessionRegistry,
				RunRegistry:     watchManagerProvider.RunRegistry,
				StateRegistry:   watchManagerProvider.StateRegistry,
				ReconnectWindow: browserRuntimeReconnectWatchdogWindow,
			},
			agentxbrowserruntime.SharedSessionBrowserNavigationResultEventRequest{
				SessionID:        ToolSessionIDFromContext(ctx),
				Route:            browserSessionRoute(runtimeInfo, resolvedBrowserApp, strings.TrimSpace(result.Backend)),
				ExplicitTargetID: strings.TrimSpace(target.TargetID),
				TabIndex:         tabIndex,
				RequestedURL:     reqURL,
				FinalURL:         finalURL,
				Title:            strings.TrimSpace(result.Title),
				Source:           "browser_act_navigate",
				Force:            force,
				PriorSelection:   browserCurrentTargetSelectionSnapshotForRoute(ctx, sessionRegistry, runtimeInfo, resolvedBrowserApp, strings.TrimSpace(result.Backend)),
				Note:             strings.TrimSpace(result.Note),
			},
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      navigationResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      finalURL,
			Title:         result.Title,
			Status: func() string {
				if navigationResult.ReviewRequired {
					return "review_required"
				}
				return result.Status
			}(),
			Force:          force,
			ReviewDecision: navigationResult.ReviewDecision,
			ReviewReady:    navigationResult.ReviewReady,
			TabIndex:       tabIndex,
			Note:           navigationResult.Note,
		}, nil
	case "extract":
		return browserActExecuteExtract(pageActionCtx, params)
	case "snapshot":
		return browserActExecuteSnapshot(pageActionCtx, params)
	case "screenshot":
		return browserActExecuteScreenshot(pageActionCtx, params)
	case "download":
		return browserActExecuteDownload(pageActionCtx, params)
	case "wait_download":
		result, err := browserActExecuteWaitDownload(pageActionCtx, params)
		if err != nil {
			return BrowserActResult{}, err
		}
		return browserLocalPlannerExecuteForActResult(pageActionCtx, params, result), nil
	case "console":
		consoleBackend, ok := routedBackend.(BrowserConsoleActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		level := strings.ToLower(strings.TrimSpace(firstString(params, "level")))
		result, err := consoleBackend.Console(ctx, BrowserConsoleRequest{
			BrowserApp: browserApp,
			WaitMs:     waitMs,
			TabIndex:   tabIndex,
			Level:      level,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		resolvedResult := browserApplyResolvedTargetRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_console",
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      resolvedResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Messages:      append([]BrowserConsoleMessage(nil), result.Messages...),
			Status:        "ok",
			TabIndex:      tabIndex,
			Note:          strings.TrimSpace(result.Note),
		}, nil
	case "requests":
		requestsBackend, ok := routedBackend.(BrowserRequestsActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		filter := strings.TrimSpace(firstString(params, "filter"))
		clear := firstBool(params, "clear")
		result, err := requestsBackend.Requests(ctx, BrowserRequestsRequest{
			BrowserApp: browserApp,
			WaitMs:     waitMs,
			TabIndex:   tabIndex,
			Filter:     filter,
			Clear:      clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		resolvedResult := browserApplyResolvedTargetRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_requests",
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      resolvedResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Requests:      append([]BrowserRequestEntry(nil), result.Requests...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "ok")),
			TabIndex:      tabIndex,
			Note:          strings.TrimSpace(result.Note),
		}, nil
	case "response_body":
		responseBackend, ok := routedBackend.(BrowserResponseBodyActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		filter := strings.TrimSpace(firstString(params, "filter"))
		requestURL := strings.TrimSpace(firstString(params, "response_url", "request_url", "url"))
		if filter == "" && requestURL == "" {
			return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"filter", "request_url"}, "browser_act: kind response_body requires filter or request_url")
		}
		maxChars := firstInt(params, "max_chars")
		if maxChars < 0 {
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"max_chars"}, "browser_act: max_chars must be >= 0 for kind response_body")
		}
		result, err := responseBackend.ResponseBody(ctx, BrowserResponseBodyRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Filter:            filter,
			URL:               requestURL,
			MaxChars:          maxChars,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_response_body",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:               kind,
			Backend:            result.Backend,
			BrowserApp:         firstNonEmpty(result.BrowserApp, browserApp),
			Target:             target.Value,
			TargetID:           actionResult.TargetID,
			Profile:            runtimeInfo.Profile,
			RuntimeTarget:      runtimeInfo.Target,
			FinalURL:           strings.TrimSpace(result.FinalURL),
			Title:              strings.TrimSpace(result.Title),
			RequestURL:         firstNonEmpty(strings.TrimSpace(result.URL), requestURL),
			RequestMethod:      strings.TrimSpace(result.Method),
			ResponseStatusCode: result.StatusCode,
			Content:            result.Body,
			ContentType:        strings.TrimSpace(result.ContentType),
			Truncated:          result.Truncated,
			Status:             "ok",
			TabIndex:           tabIndex,
			Note:               actionResult.Note,
		}, nil
	case "errors":
		errorsBackend, ok := routedBackend.(BrowserErrorsActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		clear := firstBool(params, "clear")
		result, err := errorsBackend.Errors(ctx, BrowserErrorsRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_errors",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Errors:        append([]BrowserErrorEntry(nil), result.Errors...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "ok")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "cookies":
		cookiesBackend, ok := routedBackend.(BrowserCookiesActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		filter := strings.TrimSpace(firstString(params, "filter"))
		result, err := cookiesBackend.Cookies(ctx, BrowserCookiesRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Filter:            filter,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_cookies",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Cookies:       append([]BrowserCookieEntry(nil), result.Cookies...),
			Status:        "ok",
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "cookies_set":
		cookiesBackend, ok := routedBackend.(BrowserCookiesMutatingActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		cookies, err := browserActCookieEntries(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if len(cookies) == 0 {
			repairable, safeAutorepair, repairs := browserCookieEntryAliasRepairAdvice(params["cookie"])
			return BrowserActResult{}, browserMissingCookieEntriesErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
		}
		result, err := cookiesBackend.SetCookies(ctx, BrowserCookiesSetRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			URL:               strings.TrimSpace(firstString(params, "url")),
			Cookies:           cookies,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_cookies_set",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Cookies:       append([]BrowserCookieEntry(nil), result.Cookies...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "updated"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "cookies_clear":
		cookiesBackend, ok := routedBackend.(BrowserCookiesMutatingActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := cookiesBackend.ClearCookies(ctx, BrowserCookiesClearRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			URL:               strings.TrimSpace(firstString(params, "url")),
			Filter:            strings.TrimSpace(firstString(params, "filter")),
			Name:              strings.TrimSpace(firstString(params, "name")),
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_cookies_clear",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Cookies:       append([]BrowserCookieEntry(nil), result.Cookies...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "cleared"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "storage":
		storageBackend, ok := routedBackend.(BrowserStorageActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		storageKind := strings.ToLower(strings.TrimSpace(firstString(params, "storage_kind")))
		switch storageKind {
		case "", "local":
			storageKind = "local"
		case "session":
		default:
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"storage_kind"}, "browser_act: storage_kind must be local or session for kind storage")
		}
		filter := strings.TrimSpace(firstString(params, "filter"))
		result, err := storageBackend.Storage(ctx, BrowserStorageRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Kind:              storageKind,
			Filter:            filter,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_storage",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			StorageKind:   firstNonEmpty(strings.TrimSpace(result.Kind), storageKind),
			Storage:       append([]BrowserStorageEntry(nil), result.Entries...),
			Status:        "ok",
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "storage_set":
		storageBackend, ok := routedBackend.(BrowserStorageMutatingActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		storageKind := strings.ToLower(strings.TrimSpace(firstString(params, "storage_kind")))
		switch storageKind {
		case "", "local":
			storageKind = "local"
		case "session":
		default:
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"storage_kind"}, "browser_act: storage_kind must be local or session for kind storage_set")
		}
		entries, err := browserActStorageEntries(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if len(entries) == 0 {
			repairable, safeAutorepair, repairs := browserStorageEntryAliasRepairAdvice(params["entry"])
			return BrowserActResult{}, browserMissingStorageEntriesErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
		}
		result, err := storageBackend.SetStorage(ctx, BrowserStorageSetRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Kind:              storageKind,
			Entries:           entries,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_storage_set",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			StorageKind:   firstNonEmpty(strings.TrimSpace(result.Kind), storageKind),
			Storage:       append([]BrowserStorageEntry(nil), result.Entries...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "updated"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "storage_clear":
		storageBackend, ok := routedBackend.(BrowserStorageMutatingActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		storageKind := strings.ToLower(strings.TrimSpace(firstString(params, "storage_kind")))
		switch storageKind {
		case "", "local":
			storageKind = "local"
		case "session":
		default:
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"storage_kind"}, "browser_act: storage_kind must be local or session for kind storage_clear")
		}
		result, err := storageBackend.ClearStorage(ctx, BrowserStorageClearRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Kind:              storageKind,
			Filter:            strings.TrimSpace(firstString(params, "filter")),
			Key:               strings.TrimSpace(firstString(params, "key")),
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_storage_clear",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			StorageKind:   firstNonEmpty(strings.TrimSpace(result.Kind), storageKind),
			Storage:       append([]BrowserStorageEntry(nil), result.Entries...),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "cleared"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "offline":
		offlineBackend, ok := routedBackend.(BrowserOfflineActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		enabled, ok := browserActOfflineEnabled(params)
		if !ok {
			return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"enabled"}, "browser_act: kind offline requires enabled=true|false")
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := offlineBackend.SetOffline(ctx, BrowserOfflineRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Enabled:           enabled,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_offline",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Offline:       result.Enabled,
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "updated"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "headers":
		headersBackend, ok := routedBackend.(BrowserHeadersActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		headers, clear, err := browserActHeadersRequest(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if !clear && len(headers) == 0 {
			repairable, safeAutorepair, repairs := browserHeadersRepairAdviceFromParams(params)
			return BrowserActResult{}, browserMissingHeadersErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := headersBackend.SetHeaders(ctx, BrowserHeadersRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Headers:           headers,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_headers",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		headerNames := append([]string(nil), result.HeaderNames...)
		if len(headerNames) == 0 {
			for key := range headers {
				headerNames = append(headerNames, strings.TrimSpace(key))
			}
		}
		headerNames = mergeToolMetadataStrings(nil, headerNames)
		headerCount := result.HeaderCount
		if headerCount <= 0 {
			headerCount = len(headerNames)
		}
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			HeaderNames:   headerNames,
			HeaderCount:   headerCount,
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "credentials":
		credentialsBackend, ok := routedBackend.(BrowserCredentialsActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		origin := strings.TrimSpace(firstString(params, "origin"))
		username := strings.TrimSpace(firstString(params, "username"))
		password := firstString(params, "password")
		clear := firstBool(params, "clear")
		if !clear {
			if origin == "" {
				return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"origin"}, "browser_act: origin is required for kind credentials")
			}
			if username == "" {
				return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"username"}, "browser_act: username is required for kind credentials")
			}
			if password == "" {
				return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"password"}, "browser_act: password is required for kind credentials")
			}
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := credentialsBackend.SetCredentials(ctx, BrowserCredentialsRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Origin:            origin,
			Username:          username,
			Password:          password,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_credentials",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            actionResult.TargetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			CredentialsOrigin:   firstNonEmpty(strings.TrimSpace(result.Origin), origin),
			CredentialsUsername: firstNonEmpty(strings.TrimSpace(result.Username), username),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "geolocation":
		geolocationBackend, ok := routedBackend.(BrowserGeolocationActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		latitude, longitude, accuracy, origin, clear, err := browserActGeolocationRequest(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := geolocationBackend.SetGeolocation(ctx, BrowserGeolocationRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Latitude:          latitude,
			Longitude:         longitude,
			Accuracy:          accuracy,
			Origin:            origin,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_geolocation",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:              kind,
			Backend:           result.Backend,
			BrowserApp:        firstNonEmpty(result.BrowserApp, browserApp),
			Target:            target.Value,
			TargetID:          actionResult.TargetID,
			Profile:           runtimeInfo.Profile,
			RuntimeTarget:     runtimeInfo.Target,
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             strings.TrimSpace(result.Title),
			Latitude:          result.Latitude,
			Longitude:         result.Longitude,
			Accuracy:          result.Accuracy,
			GeolocationOrigin: firstNonEmpty(strings.TrimSpace(result.Origin), origin),
			Status:            firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:          tabIndex,
			Note:              actionResult.Note,
		}, nil
	case "media":
		mediaBackend, ok := routedBackend.(BrowserMediaActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		mediaValue, clear, err := browserActMediaRequest(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := mediaBackend.SetMedia(ctx, BrowserMediaRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Media:             mediaValue,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_media",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Media:         firstNonEmpty(strings.TrimSpace(result.Media), mediaValue),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "timezone":
		timezoneBackend, ok := routedBackend.(BrowserTimezoneActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		timezone, clear, err := browserActStringSetting(params, "timezone", "timezone", "value")
		if err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := timezoneBackend.SetTimezone(ctx, BrowserTimezoneRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Timezone:          timezone,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_timezone",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Timezone:      firstNonEmpty(strings.TrimSpace(result.Timezone), timezone),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "locale":
		localeBackend, ok := routedBackend.(BrowserLocaleActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		locale, clear, err := browserActStringSetting(params, "locale", "locale", "value")
		if err != nil {
			return BrowserActResult{}, err
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := localeBackend.SetLocale(ctx, BrowserLocaleRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Locale:            locale,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_locale",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Locale:        firstNonEmpty(strings.TrimSpace(result.Locale), locale),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "device":
		deviceBackend, ok := routedBackend.(BrowserDeviceActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		device, clear, err := browserActStringSetting(params, "device", "device", "value")
		if err != nil {
			return BrowserActResult{}, err
		}
		width := firstInt(params, "width")
		height := firstInt(params, "height")
		if (width < 0) || (height < 0) {
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"width", "height"}, "browser_act: width and height must be >= 0 for kind device")
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		result, err := deviceBackend.SetDevice(ctx, BrowserDeviceRequest{
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
			Device:            device,
			Width:             width,
			Height:            height,
			Clear:             clear,
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_device",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		resultWidth := result.Width
		if resultWidth <= 0 {
			resultWidth = width
		}
		resultHeight := result.Height
		if resultHeight <= 0 {
			resultHeight = height
		}
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Device:        firstNonEmpty(strings.TrimSpace(result.Device), device),
			Width:         resultWidth,
			Height:        resultHeight,
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), browserActClearStatus(clear, "updated")),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "highlight":
		highlightBackend, ok := routedBackend.(BrowserHighlightActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if waitMs <= 0 && target.Explicit {
			waitMs = browserTabTargetWaitMs
		}
		elementTarget, err := resolveBrowserActionElementTarget(
			firstString(params, "selector"),
			firstString(params, "ref", "element_ref", "input_ref"),
			firstString(params, "element"),
		)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if !browserElementTargetHasActionableLocator(elementTarget) {
			repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "text", "label")
			return BrowserActResult{}, browserMissingLocatorError(
				"browser_act",
				"browser_act: ref or selector is required for kind highlight",
				repairable,
				safeAutorepair,
				repairs,
			)
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			ctx,
			route,
			func(_ BrowserBackend) (BrowserHighlightResult, error) {
				return highlightBackend.Highlight(ctx, BrowserHighlightRequest{
					BrowserApp:  browserApp,
					WaitMs:      waitMs,
					TabIndex:    tabIndex,
					Ref:         elementTarget.Ref,
					ElementHint: browserElementHintForTarget(elementTarget),
					Selector:    elementTarget.Selector,
				})
			},
			browserManagedRouteExecutionArgs{
				URL:        reqURL,
				BrowserApp: browserApp,
				WaitMs:     waitMs,
				TabIndex:   tabIndex,
				Force:      force,
				FinalURL:   route.managedFinalURL(ctx, browserApp, target, reqURL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserHighlightResult {
				return BrowserHighlightResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Ref:             elementTarget.Ref,
					Selector:        elementTarget.Selector,
					Status:          policy.Status,
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return BrowserActResult{}, err
		}
		result := managed.Result
		recovery := managed.Recovery
		snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
		actionResult := agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
			TargetID: strings.TrimSpace(target.TargetID),
			Note:     strings.TrimSpace(result.Note),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			actionResult = browserApplyActionRuntimeResult(
				ctx,
				watchManagerProvider,
				sessionRegistry,
				runtimeInfo,
				firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
				strings.TrimSpace(result.Backend),
				strings.TrimSpace(target.TargetID),
				tabIndex,
				false,
				strings.TrimSpace(result.FinalURL),
				strings.TrimSpace(result.Title),
				"browser_act_highlight",
				tabIndex > 0,
				"",
				false,
				strings.TrimSpace(result.Note),
			)
		}
		targetID := firstNonEmpty(strings.TrimSpace(actionResult.TargetID), strings.TrimSpace(target.TargetID))
		targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            targetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            snapshotRecovery.Text,
			SnapshotFormat:      snapshotRecovery.Format,
			SnapshotMode:        snapshotRecovery.Mode,
			SnapshotRefs:        snapshotRecovery.Refs,
			SnapshotInteractive: snapshotRecovery.Interactive,
			SnapshotCompact:     snapshotRecovery.Compact,
			SnapshotDepth:       snapshotRecovery.Depth,
			SnapshotFrame:       snapshotRecovery.Frame,
			Elements:            snapshotRecovery.Elements,
			Ref:                 browserPreferredResultRef(elementTarget.Ref, result.Ref),
			Selector:            firstNonEmpty(strings.TrimSpace(result.Selector), elementTarget.Selector),
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "highlighted"),
			Truncated:           snapshotRecovery.Truncated,
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "trace_start":
		return browserActExecuteTraceStart(pageActionCtx, params)
	case "trace_stop":
		return browserActExecuteTraceStop(pageActionCtx, params)
	case "save_pdf":
		return browserActExecuteSavePDF(pageActionCtx, params)
	case "save_html":
		return browserActExecuteSaveHTML(pageActionCtx, params)
	case "dialog":
		return browserActExecuteDialog(pageActionCtx, params)
	case "upload":
		return browserActExecuteUpload(pageActionCtx, params)
	case "press":
		pressBackend, ok := routedBackend.(BrowserPressActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		key := strings.TrimSpace(firstString(params, "key"))
		if key == "" {
			repairable, safeAutorepair, repairs := browserPressKeyRepairAdviceFromParams(params)
			return BrowserActResult{}, browserMissingPressKeyErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 250
		}
		delayMs := firstInt(params, "delay_ms", "delayMs")
		if delayMs < 0 {
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"delay_ms"}, "browser_act: delay_ms must be >= 0 for kind press")
		}
		result, err := pressBackend.Press(ctx, BrowserPressRequest{
			URL:               reqURL,
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			PostWaitMs:        postWaitMs,
			Key:               key,
			DelayMs:           delayMs,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_press",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Key:           firstNonEmpty(strings.TrimSpace(result.Key), key),
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "pressed"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "hover":
		hoverBackend, ok := routedBackend.(BrowserHoverActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		elementTarget, err := resolveBrowserActionElementTarget(
			firstString(params, "selector"),
			firstString(params, "ref", "element_ref"),
			firstString(params, "element"),
		)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if err := browserValidateElementTargetPageBinding(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, reqURL, elementTarget); err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if !browserElementTargetHasActionableLocator(elementTarget) {
			repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "text", "label")
			return BrowserActResult{}, browserMissingLocatorError(
				"browser_act",
				"browser_act: selector or ref is required for kind hover",
				repairable,
				safeAutorepair,
				repairs,
			)
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 250
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			ctx,
			route,
			func(_ BrowserBackend) (BrowserHoverResult, error) {
				return hoverBackend.Hover(ctx, BrowserHoverRequest{
					URL:         reqURL,
					BrowserApp:  browserApp,
					WaitMs:      waitMs,
					PostWaitMs:  postWaitMs,
					ElementRef:  elementTarget.Ref,
					ElementHint: browserElementHintForTarget(elementTarget),
					Selector:    elementTarget.Selector,
					TabIndex:    tabIndex,
				})
			},
			browserManagedRouteExecutionArgs{
				URL:        reqURL,
				BrowserApp: browserApp,
				WaitMs:     waitMs,
				TabIndex:   tabIndex,
				Force:      force,
				FinalURL:   route.managedFinalURL(ctx, browserApp, target, reqURL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserHoverResult {
				return BrowserHoverResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Status:          policy.Status,
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return BrowserActResult{}, err
		}
		result := managed.Result
		recovery := managed.Recovery
		snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
		actionResult := agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
			TargetID: strings.TrimSpace(target.TargetID),
			Note:     strings.TrimSpace(result.Note),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			actionResult = browserApplyActionRuntimeResult(
				ctx,
				watchManagerProvider,
				sessionRegistry,
				runtimeInfo,
				firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
				strings.TrimSpace(result.Backend),
				strings.TrimSpace(target.TargetID),
				tabIndex,
				false,
				strings.TrimSpace(result.FinalURL),
				strings.TrimSpace(result.Title),
				"browser_act_hover",
				tabIndex > 0,
				"",
				false,
				strings.TrimSpace(result.Note),
			)
		}
		targetID := firstNonEmpty(strings.TrimSpace(actionResult.TargetID), strings.TrimSpace(target.TargetID))
		targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            targetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            snapshotRecovery.Text,
			SnapshotFormat:      snapshotRecovery.Format,
			SnapshotMode:        snapshotRecovery.Mode,
			SnapshotRefs:        snapshotRecovery.Refs,
			SnapshotInteractive: snapshotRecovery.Interactive,
			SnapshotCompact:     snapshotRecovery.Compact,
			SnapshotDepth:       snapshotRecovery.Depth,
			SnapshotFrame:       snapshotRecovery.Frame,
			Elements:            snapshotRecovery.Elements,
			Ref:                 elementTarget.Ref,
			Selector:            elementTarget.Selector,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "hovered"),
			Truncated:           snapshotRecovery.Truncated,
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "drag":
		dragBackend, ok := routedBackend.(BrowserDragActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		startTarget, endTarget, err := browserActDragTargets(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if err := browserValidateElementTargetPageBinding(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, reqURL, startTarget); err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if err := browserValidateElementTargetPageBinding(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, reqURL, endTarget); err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 500
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			ctx,
			route,
			func(_ BrowserBackend) (BrowserDragResult, error) {
				return dragBackend.Drag(ctx, BrowserDragRequest{
					URL:           reqURL,
					BrowserApp:    browserApp,
					WaitMs:        waitMs,
					PostWaitMs:    postWaitMs,
					StartRef:      startTarget.Ref,
					StartHint:     browserElementHintForTarget(startTarget),
					StartSelector: startTarget.Selector,
					EndRef:        endTarget.Ref,
					EndHint:       browserElementHintForTarget(endTarget),
					EndSelector:   endTarget.Selector,
					TabIndex:      tabIndex,
				})
			},
			browserManagedRouteExecutionArgs{
				URL:        reqURL,
				BrowserApp: browserApp,
				WaitMs:     waitMs,
				TabIndex:   tabIndex,
				Force:      force,
				FinalURL:   route.managedFinalURL(ctx, browserApp, target, reqURL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserDragResult {
				return BrowserDragResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Status:          policy.Status,
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return BrowserActResult{}, err
		}
		result := managed.Result
		recovery := managed.Recovery
		snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
		actionResult := agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
			TargetID: strings.TrimSpace(target.TargetID),
			Note:     strings.TrimSpace(result.Note),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			actionResult = browserApplyActionRuntimeResult(
				ctx,
				watchManagerProvider,
				sessionRegistry,
				runtimeInfo,
				firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
				strings.TrimSpace(result.Backend),
				strings.TrimSpace(target.TargetID),
				tabIndex,
				false,
				strings.TrimSpace(result.FinalURL),
				strings.TrimSpace(result.Title),
				"browser_act_drag",
				tabIndex > 0,
				"",
				false,
				strings.TrimSpace(result.Note),
			)
		}
		targetID := firstNonEmpty(strings.TrimSpace(actionResult.TargetID), strings.TrimSpace(target.TargetID))
		targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            targetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            snapshotRecovery.Text,
			SnapshotFormat:      snapshotRecovery.Format,
			SnapshotMode:        snapshotRecovery.Mode,
			SnapshotRefs:        snapshotRecovery.Refs,
			SnapshotInteractive: snapshotRecovery.Interactive,
			SnapshotCompact:     snapshotRecovery.Compact,
			SnapshotDepth:       snapshotRecovery.Depth,
			SnapshotFrame:       snapshotRecovery.Frame,
			Elements:            snapshotRecovery.Elements,
			Ref:                 startTarget.Ref,
			Selector:            startTarget.Selector,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "dragged"),
			Truncated:           snapshotRecovery.Truncated,
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "select":
		selectBackend, ok := routedBackend.(BrowserSelectActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		elementTarget, err := resolveBrowserActionElementTarget(
			firstString(params, "selector"),
			firstString(params, "ref", "element_ref", "input_ref"),
			firstString(params, "element"),
		)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if err := browserValidateElementTargetPageBinding(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, reqURL, elementTarget); err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		if !browserElementTargetHasActionableLocator(elementTarget) {
			repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "text", "label")
			return BrowserActResult{}, browserMissingLocatorError(
				"browser_act",
				"browser_act: selector or ref is required for kind select",
				repairable,
				safeAutorepair,
				repairs,
			)
		}
		values := browserActStringValues(params, "values")
		if len(values) == 0 {
			value := strings.TrimSpace(firstString(params, "value"))
			if value != "" {
				values = []string{value}
			}
		}
		if len(values) == 0 {
			repairable, safeAutorepair, repairs := browserSelectValueRepairAdviceFromParams(params)
			return BrowserActResult{}, browserMissingValueErrorWithRepair("browser_act", "select", repairable, safeAutorepair, repairs)
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 250
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			ctx,
			route,
			func(_ BrowserBackend) (BrowserSelectResult, error) {
				return selectBackend.Select(ctx, BrowserSelectRequest{
					URL:         reqURL,
					BrowserApp:  browserApp,
					WaitMs:      waitMs,
					PostWaitMs:  postWaitMs,
					ElementRef:  elementTarget.Ref,
					ElementHint: browserElementHintForTarget(elementTarget),
					Selector:    elementTarget.Selector,
					Values:      append([]string(nil), values...),
					TabIndex:    tabIndex,
				})
			},
			browserManagedRouteExecutionArgs{
				URL:        reqURL,
				BrowserApp: browserApp,
				WaitMs:     waitMs,
				TabIndex:   tabIndex,
				Force:      force,
				FinalURL:   route.managedFinalURL(ctx, browserApp, target, reqURL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserSelectResult {
				return BrowserSelectResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Values:          append([]string(nil), values...),
					Status:          policy.Status,
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return BrowserActResult{}, err
		}
		result := managed.Result
		recovery := managed.Recovery
		snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
		actionResult := agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
			TargetID: strings.TrimSpace(target.TargetID),
			Note:     strings.TrimSpace(result.Note),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			actionResult = browserApplyActionRuntimeResult(
				ctx,
				watchManagerProvider,
				sessionRegistry,
				runtimeInfo,
				firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
				strings.TrimSpace(result.Backend),
				strings.TrimSpace(target.TargetID),
				tabIndex,
				false,
				strings.TrimSpace(result.FinalURL),
				strings.TrimSpace(result.Title),
				"browser_act_select",
				tabIndex > 0,
				"",
				false,
				strings.TrimSpace(result.Note),
			)
		}
		targetID := firstNonEmpty(strings.TrimSpace(actionResult.TargetID), strings.TrimSpace(target.TargetID))
		targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
		outputValues := append([]string(nil), result.Values...)
		if len(outputValues) == 0 {
			outputValues = append([]string(nil), values...)
		}
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            targetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            snapshotRecovery.Text,
			SnapshotFormat:      snapshotRecovery.Format,
			SnapshotMode:        snapshotRecovery.Mode,
			SnapshotRefs:        snapshotRecovery.Refs,
			SnapshotInteractive: snapshotRecovery.Interactive,
			SnapshotCompact:     snapshotRecovery.Compact,
			SnapshotDepth:       snapshotRecovery.Depth,
			SnapshotFrame:       snapshotRecovery.Frame,
			Elements:            snapshotRecovery.Elements,
			Ref:                 elementTarget.Ref,
			Selector:            elementTarget.Selector,
			Values:              outputValues,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "selected"),
			Truncated:           snapshotRecovery.Truncated,
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "fill":
		fillBackend, ok := routedBackend.(BrowserFillActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		if _, hasFields := params["fields"]; !hasFields {
			values := browserActStringValues(params, "values")
			if len(values) == 0 && strings.TrimSpace(firstString(params, "value")) == "" {
				target, err := resolveBrowserElementTarget(
					firstString(params, "selector", "element"),
					firstString(params, "ref", "element_ref", "input_ref"),
				)
				if err != nil {
					return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
				}
				if browserElementTargetHasActionableLocator(target) {
					repairable, safeAutorepair, repairs := browserFillValueRepairAdviceFromParams(params)
					if repairable {
						return BrowserActResult{}, browserMissingValueErrorWithRepair("browser_act", "fill", repairable, safeAutorepair, repairs)
					}
				}
			}
		}
		fields, err := browserActFillFields(params)
		if err != nil {
			return BrowserActResult{}, err
		}
		if len(fields) == 0 {
			return BrowserActResult{}, browserMissingFillInputError("browser_act")
		}
		for _, field := range fields {
			if err := browserValidateElementTargetPageBinding(ctx, sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, reqURL, browserElementTarget{
				Selector: strings.TrimSpace(field.Selector),
				Ref:      strings.TrimSpace(field.Ref),
			}); err != nil {
				return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
			}
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 250
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			ctx,
			route,
			func(_ BrowserBackend) (BrowserFillResult, error) {
				return fillBackend.Fill(ctx, BrowserFillRequest{
					URL:        reqURL,
					BrowserApp: browserApp,
					WaitMs:     waitMs,
					PostWaitMs: postWaitMs,
					Fields:     append([]BrowserFillField(nil), fields...),
					Submit:     firstBool(params, "submit"),
					TabIndex:   tabIndex,
				})
			},
			browserManagedRouteExecutionArgs{
				URL:        reqURL,
				BrowserApp: browserApp,
				WaitMs:     waitMs,
				TabIndex:   tabIndex,
				Force:      force,
				FinalURL:   route.managedFinalURL(ctx, browserApp, target, reqURL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserFillResult {
				return BrowserFillResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					FieldCount:      len(fields),
					Status:          policy.Status,
					Submitted:       firstBool(params, "submit"),
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return BrowserActResult{}, err
		}
		result := managed.Result
		recovery := managed.Recovery
		snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
		actionResult := agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
			TargetID: strings.TrimSpace(target.TargetID),
			Note:     strings.TrimSpace(result.Note),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			actionResult = browserApplyActionRuntimeResult(
				ctx,
				watchManagerProvider,
				sessionRegistry,
				runtimeInfo,
				firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
				strings.TrimSpace(result.Backend),
				strings.TrimSpace(target.TargetID),
				tabIndex,
				false,
				strings.TrimSpace(result.FinalURL),
				strings.TrimSpace(result.Title),
				"browser_act_fill",
				tabIndex > 0,
				"",
				false,
				strings.TrimSpace(result.Note),
			)
		}
		targetID := firstNonEmpty(strings.TrimSpace(actionResult.TargetID), strings.TrimSpace(target.TargetID))
		targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
		fieldCount := result.FieldCount
		if fieldCount <= 0 {
			fieldCount = len(fields)
		}
		return browserActResultWithResolverFallbackSummary(BrowserActResult{
			Kind:                kind,
			Backend:             result.Backend,
			BrowserApp:          firstNonEmpty(result.BrowserApp, browserApp),
			Target:              target.Value,
			TargetID:            targetID,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            snapshotRecovery.Text,
			SnapshotFormat:      snapshotRecovery.Format,
			SnapshotMode:        snapshotRecovery.Mode,
			SnapshotRefs:        snapshotRecovery.Refs,
			SnapshotInteractive: snapshotRecovery.Interactive,
			SnapshotCompact:     snapshotRecovery.Compact,
			SnapshotDepth:       snapshotRecovery.Depth,
			SnapshotFrame:       snapshotRecovery.Frame,
			Elements:            snapshotRecovery.Elements,
			FieldCount:          fieldCount,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "filled"),
			Submitted:           result.Submitted,
			Truncated:           snapshotRecovery.Truncated,
			TabIndex:            tabIndex,
			Note:                actionResult.Note,
		}), nil
	case "resize":
		resizeBackend, ok := routedBackend.(BrowserResizeActionBackend)
		if !ok {
			return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(runtimeInfo.Backend), kind)
		}
		width := firstInt(params, "width")
		height := firstInt(params, "height")
		if width <= 0 || height <= 0 {
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"width", "height"}, "browser_act: width and height must both be > 0 for kind resize")
		}
		if waitMs <= 0 {
			if reqURL != "" {
				waitMs = defaultWaitMs
			} else if target.Explicit {
				waitMs = browserTabTargetWaitMs
			}
		}
		if postWaitMs <= 0 {
			postWaitMs = 250
		}
		result, err := resizeBackend.Resize(ctx, BrowserResizeRequest{
			URL:               reqURL,
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			PostWaitMs:        postWaitMs,
			Width:             width,
			Height:            height,
			TabIndex:          tabIndex,
			PreferredTargetID: strings.TrimSpace(target.TargetID),
		})
		if err != nil {
			return BrowserActResult{}, err
		}
		actionResult := browserApplyActionRuntimeResult(
			ctx,
			watchManagerProvider,
			sessionRegistry,
			runtimeInfo,
			firstNonEmpty(strings.TrimSpace(result.BrowserApp), browserApp),
			strings.TrimSpace(result.Backend),
			strings.TrimSpace(target.TargetID),
			tabIndex,
			false,
			strings.TrimSpace(result.FinalURL),
			strings.TrimSpace(result.Title),
			"browser_act_resize",
			tabIndex > 0,
			"",
			false,
			strings.TrimSpace(result.Note),
		)
		resultWidth := result.Width
		if resultWidth <= 0 {
			resultWidth = width
		}
		resultHeight := result.Height
		if resultHeight <= 0 {
			resultHeight = height
		}
		return BrowserActResult{
			Kind:          kind,
			Backend:       result.Backend,
			BrowserApp:    firstNonEmpty(result.BrowserApp, browserApp),
			Target:        target.Value,
			TargetID:      actionResult.TargetID,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			FinalURL:      strings.TrimSpace(result.FinalURL),
			Title:         strings.TrimSpace(result.Title),
			Width:         resultWidth,
			Height:        resultHeight,
			Status:        firstNonEmpty(strings.TrimSpace(result.Status), "resized"),
			TabIndex:      tabIndex,
			Note:          actionResult.Note,
		}, nil
	case "click":
		return browserActExecuteClick(pageActionCtx, params, postWaitMs)
	case "type":
		return browserActExecuteType(pageActionCtx, params, postWaitMs)
	case "evaluate":
		return browserActExecuteEvaluate(pageActionCtx, params)
	case "wait":
		if waitMs <= 0 {
			waitMs = defaultWaitMs
		}
		waitCtx, cancel := context.WithTimeout(ctx, time.Duration(maxInt(waitMs+1000, waitMs))*time.Millisecond)
		defer cancel()
		if err := sleepWithContext(waitCtx, time.Duration(waitMs)*time.Millisecond); err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
		return BrowserActResult{
			Kind:          kind,
			BrowserApp:    browserApp,
			Profile:       runtimeInfo.Profile,
			RuntimeTarget: runtimeInfo.Target,
			Status:        "waited",
			Note:          fmt.Sprintf("waited %dms", waitMs),
		}, nil
	case "list_tabs", "focus_tab", "close_tab":
		return browserActExecuteTabs(tabsCtx, params, kind)
	default:
		return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"kind"}, fmt.Sprintf("browser_act: unsupported kind %q", kind))
	}
}

func browserActCookieEntries(params map[string]any) ([]BrowserCookieEntry, error) {
	entries := make([]BrowserCookieEntry, 0, 2)
	if raw, ok := params["cookies"]; ok {
		items, ok := raw.([]any)
		if !ok {
			repairable, safeAutorepair, repairs := browserCookieEntriesShapeRepairAdvice(raw)
			return nil, browserInvalidCookieEntriesShapeError("browser_act", repairable, safeAutorepair, repairs)
		}
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, browserInvalidArgumentError("browser_act", []string{"cookies"}, "browser_act: cookies entries must be objects")
			}
			entry := BrowserCookieEntry{
				Name:     strings.TrimSpace(firstString(obj, "name")),
				Value:    firstString(obj, "value"),
				Domain:   strings.TrimSpace(firstString(obj, "domain")),
				Path:     strings.TrimSpace(firstString(obj, "path")),
				SameSite: strings.TrimSpace(firstString(obj, "same_site", "sameSite")),
				HTTPOnly: firstBool(obj, "http_only", "httpOnly"),
				Secure:   firstBool(obj, "secure"),
			}
			if expires := firstInt(obj, "expires"); expires > 0 {
				entry.Expires = int64(expires)
			}
			if entry.Name == "" {
				return nil, browserMissingRequiredArgumentError("browser_act", []string{"cookies[].name"}, "browser_act: cookies entries require name")
			}
			entries = append(entries, entry)
		}
	}
	if len(entries) > 0 {
		return entries, nil
	}
	name := strings.TrimSpace(firstString(params, "name"))
	if name == "" {
		return nil, nil
	}
	entry := BrowserCookieEntry{
		Name:     name,
		Value:    firstString(params, "value"),
		Domain:   strings.TrimSpace(firstString(params, "domain")),
		Path:     strings.TrimSpace(firstString(params, "path")),
		SameSite: strings.TrimSpace(firstString(params, "same_site", "sameSite")),
		HTTPOnly: firstBool(params, "http_only", "httpOnly"),
		Secure:   firstBool(params, "secure"),
	}
	if expires := firstInt(params, "expires"); expires > 0 {
		entry.Expires = int64(expires)
	}
	return []BrowserCookieEntry{entry}, nil
}

func browserActStorageEntries(params map[string]any) ([]BrowserStorageEntry, error) {
	entries := make([]BrowserStorageEntry, 0, 2)
	if raw, ok := params["entries"]; ok {
		items, ok := raw.([]any)
		if !ok {
			repairable, safeAutorepair, repairs := browserStorageEntriesShapeRepairAdvice(raw)
			return nil, browserInvalidStorageEntriesShapeError("browser_act", repairable, safeAutorepair, repairs)
		}
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, browserInvalidArgumentError("browser_act", []string{"entries"}, "browser_act: entries must contain objects")
			}
			entry := BrowserStorageEntry{
				Key:   strings.TrimSpace(firstString(obj, "key")),
				Value: firstString(obj, "value"),
			}
			if entry.Key == "" {
				return nil, browserMissingRequiredArgumentError("browser_act", []string{"entries[].key"}, "browser_act: storage entries require key")
			}
			entries = append(entries, entry)
		}
	}
	if len(entries) > 0 {
		return entries, nil
	}
	key := strings.TrimSpace(firstString(params, "key"))
	if key == "" {
		return nil, nil
	}
	return []BrowserStorageEntry{{Key: key, Value: firstString(params, "value")}}, nil
}

func browserActStringValues(params map[string]any, key string) []string {
	return readStringList(params, key)
}

func browserActOfflineEnabled(params map[string]any) (bool, bool) {
	for _, key := range []string{"enabled", "offline"} {
		if raw, ok := params[key]; ok {
			switch value := raw.(type) {
			case bool:
				return value, true
			case string:
				switch strings.ToLower(strings.TrimSpace(value)) {
				case "on", "true", "1", "enabled":
					return true, true
				case "off", "false", "0", "disabled":
					return false, true
				}
			}
		}
	}
	if value := strings.ToLower(strings.TrimSpace(firstString(params, "value", "state", "mode"))); value != "" {
		switch value {
		case "on", "true", "1", "enabled", "offline":
			return true, true
		case "off", "false", "0", "disabled", "online":
			return false, true
		}
	}
	return false, false
}

func browserActHeadersRequest(params map[string]any) (map[string]string, bool, error) {
	clear := firstBool(params, "clear")
	headers := readStringMap(params["headers"])
	if raw := strings.TrimSpace(firstString(params, "headers_json")); raw != "" {
		parsed := map[string]string{}
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, false, browserInvalidArgumentError("browser_act", []string{"headers_json"}, fmt.Sprintf("browser_act: headers_json must be valid JSON object: %v", err))
		}
		if headers == nil {
			headers = map[string]string{}
		}
		for key, value := range parsed {
			if strings.TrimSpace(key) == "" {
				continue
			}
			headers[strings.TrimSpace(key)] = value
		}
	}
	if clear && len(headers) > 0 {
		return nil, false, browserInvalidArgumentError("browser_act", []string{"headers", "clear"}, "browser_act: headers and clear=true cannot be used together")
	}
	return headers, clear, nil
}

func browserActStringSetting(params map[string]any, kind string, keys ...string) (string, bool, error) {
	clear := firstBool(params, "clear")
	value := strings.TrimSpace(firstString(params, keys...))
	if clear && value != "" {
		return "", false, browserInvalidArgumentError("browser_act", []string{keys[0], "clear"}, fmt.Sprintf("browser_act: %s and clear=true cannot be used together for kind %s", keys[0], kind))
	}
	if !clear && value == "" {
		return "", false, browserMissingRequiredArgumentError("browser_act", []string{keys[0]}, fmt.Sprintf("browser_act: %s is required for kind %s", keys[0], kind))
	}
	return value, clear, nil
}

func browserActFirstFloat(params map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		raw, ok := params[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case float64:
			return value, true
		case float32:
			return float64(value), true
		case int:
			return float64(value), true
		case int64:
			return float64(value), true
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func browserActRememberTargetCandidate(result BrowserActResult) (string, int) {
	switch strings.ToLower(strings.TrimSpace(result.Kind)) {
	case "list_tabs", "close_tab":
		if result.ActiveIndex > 0 {
			return "", result.ActiveIndex
		}
	case "focus_tab":
		if result.ActiveIndex > 0 {
			return "", result.ActiveIndex
		}
		if result.TabIndex > 0 {
			return "", result.TabIndex
		}
	}
	if strings.TrimSpace(result.TargetID) != "" {
		return strings.TrimSpace(result.TargetID), 0
	}
	if result.TabIndex > 0 {
		return "", result.TabIndex
	}
	return "", 0
}

func browserActGeolocationRequest(params map[string]any) (float64, float64, float64, string, bool, error) {
	clear := firstBool(params, "clear")
	origin := strings.TrimSpace(firstString(params, "origin"))
	lat, hasLat := browserActFirstFloat(params, "latitude", "lat")
	lon, hasLon := browserActFirstFloat(params, "longitude", "lon", "lng")
	acc, hasAcc := browserActFirstFloat(params, "accuracy")
	if clear && (hasLat || hasLon || hasAcc || origin != "") {
		return 0, 0, 0, "", false, browserInvalidArgumentError("browser_act", []string{"latitude", "longitude", "accuracy", "origin", "clear"}, "browser_act: latitude/longitude/accuracy/origin and clear=true cannot be used together for kind geolocation")
	}
	if !clear && (!hasLat || !hasLon) {
		return 0, 0, 0, "", false, browserMissingRequiredArgumentError("browser_act", []string{"latitude", "longitude"}, "browser_act: latitude and longitude are required for kind geolocation")
	}
	if hasAcc && acc < 0 {
		return 0, 0, 0, "", false, browserInvalidArgumentError("browser_act", []string{"accuracy"}, "browser_act: accuracy must be >= 0 for kind geolocation")
	}
	return lat, lon, acc, origin, clear, nil
}

func browserActMediaRequest(params map[string]any) (string, bool, error) {
	value, clear, err := browserActStringSetting(params, "media", "media", "value", "scheme")
	if err != nil {
		return "", false, err
	}
	if clear {
		return "", true, nil
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "dark", "light", "no-preference", "none":
		return strings.ToLower(strings.TrimSpace(value)), false, nil
	default:
		return "", false, browserInvalidArgumentError("browser_act", []string{"media"}, "browser_act: media must be one of dark, light, no-preference, none for kind media")
	}
}

func browserActClearStatus(clear bool, updated string) string {
	if clear {
		return "cleared"
	}
	return updated
}

func browserActDragEndpointParams(params map[string]any, selectorKeys []string, refKeys []string, elementKeys []string, aliasKey string) (string, string, string) {
	selector := firstString(params, selectorKeys...)
	ref := firstString(params, refKeys...)
	element := firstString(params, elementKeys...)
	if selector != "" || ref != "" || element != "" {
		return selector, ref, element
	}
	raw, ok := params[aliasKey]
	if !ok || raw == nil {
		return "", "", ""
	}
	switch value := raw.(type) {
	case string:
		return "", "", browserSanitizeLooseArgumentString(value)
	case map[string]any:
		return firstString(value, "selector"),
			firstString(value, "ref", "element_ref", "input_ref"),
			firstString(value, "element", "label")
	default:
		return "", "", ""
	}
}

func browserActDragTargets(params map[string]any) (browserElementTarget, browserElementTarget, error) {
	startSelector, startRef, startElement := browserActDragEndpointParams(
		params,
		[]string{"start_selector", "startSelector"},
		[]string{"start_ref", "startRef"},
		[]string{"start_element", "startElement", "start_label", "startLabel"},
		"from",
	)
	startTarget, err := resolveBrowserActionElementTarget(
		startSelector,
		startRef,
		startElement,
	)
	if err != nil {
		return browserElementTarget{}, browserElementTarget{}, fmt.Errorf("browser_act: %w", err)
	}
	if !browserElementTargetHasActionableLocator(startTarget) {
		return browserElementTarget{}, browserElementTarget{}, browserMissingRequiredArgumentError("browser_act", []string{"start_selector", "start_ref", "start_element", "from"}, "browser_act: start_selector or start_ref or start_element or from is required for kind drag")
	}
	endSelector, endRef, endElement := browserActDragEndpointParams(
		params,
		[]string{"end_selector", "endSelector"},
		[]string{"end_ref", "endRef"},
		[]string{"end_element", "endElement", "end_label", "endLabel"},
		"to",
	)
	endTarget, err := resolveBrowserActionElementTarget(
		endSelector,
		endRef,
		endElement,
	)
	if err != nil {
		return browserElementTarget{}, browserElementTarget{}, fmt.Errorf("browser_act: %w", err)
	}
	if !browserElementTargetHasActionableLocator(endTarget) {
		return browserElementTarget{}, browserElementTarget{}, browserMissingRequiredArgumentError("browser_act", []string{"end_selector", "end_ref", "end_element", "to"}, "browser_act: end_selector or end_ref or end_element or to is required for kind drag")
	}
	return startTarget, endTarget, nil
}

func browserActFillFields(params map[string]any) ([]BrowserFillField, error) {
	fields := make([]BrowserFillField, 0, 4)
	if _, hasFields := params["fields"]; !hasFields {
		if raw, ok := params["field"]; ok && raw != nil {
			repairable, safeAutorepair, repairs := browserFillFieldAliasRepairAdvice(raw)
			return nil, browserMissingFillInputErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
		}
	}
	if raw, ok := params["fields"]; ok && raw != nil {
		items, ok := raw.([]any)
		if !ok {
			repairable, safeAutorepair, repairs := browserFillFieldsShapeRepairAdvice(raw)
			return nil, browserInvalidFillFieldsShapeError("browser_act", repairable, safeAutorepair, repairs)
		}
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, browserInvalidFillFieldsShapeError("browser_act", false, false, nil)
			}
			field, err := browserActFillField(obj)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
	}
	if len(fields) > 0 {
		return fields, nil
	}
	field, err := browserActFillField(params)
	if err != nil {
		if argErr, ok := AsToolArgumentError(err); ok {
			switch strings.TrimSpace(argErr.Code) {
			case "missing_locator", "missing_value":
				if browserActShouldPreserveTopLevelFillArgError(params, strings.TrimSpace(argErr.Code)) {
					return nil, err
				}
				return nil, nil
			}
		}
		return nil, err
	}
	return []BrowserFillField{field}, nil
}

func browserActShouldPreserveTopLevelFillArgError(params map[string]any, code string) bool {
	switch strings.TrimSpace(code) {
	case "missing_locator":
		return browserActHasTopLevelFillValueSignal(params)
	case "missing_value":
		return browserActHasTopLevelFillExplicitValueAlias(params)
	default:
		return false
	}
}

func browserActHasTopLevelFillValueSignal(params map[string]any) bool {
	if params == nil {
		return false
	}
	for _, key := range []string{"text", "value"} {
		if strings.TrimSpace(firstString(params, key)) != "" {
			return true
		}
	}
	return len(readStringList(params, "values")) > 0
}

func browserActHasTopLevelFillExplicitValueAlias(params map[string]any) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "text")) != "" {
		return true
	}
	return len(readStringList(params, "values")) > 0 || strings.TrimSpace(firstString(params, "value")) != ""
}

func browserActFillField(params map[string]any) (BrowserFillField, error) {
	target, err := resolveBrowserActionElementTarget(
		firstString(params, "selector"),
		firstString(params, "ref", "element_ref", "input_ref"),
		firstString(params, "element"),
	)
	if err != nil {
		return BrowserFillField{}, fmt.Errorf("browser_act: %w", err)
	}
	if !browserElementTargetHasActionableLocator(target) {
		repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "element", "label")
		if repairable {
			if valueRepairable, valueSafeAutorepair, valueRepairs := browserFillValueRepairAdviceFromParams(params); valueRepairable {
				safeAutorepair = safeAutorepair || valueSafeAutorepair
				repairs = append(repairs, valueRepairs...)
			}
		}
		return BrowserFillField{}, browserMissingLocatorError(
			"browser_act",
			"browser_act: selector or ref is required for kind fill",
			repairable,
			safeAutorepair,
			repairs,
		)
	}
	values := browserActStringValues(params, "values")
	value := firstString(params, "value")
	if len(values) == 0 && strings.TrimSpace(value) != "" {
		values = []string{value}
	}
	if len(values) == 0 {
		repairable, safeAutorepair, repairs := browserFillValueRepairAdviceFromParams(params)
		return BrowserFillField{}, browserMissingValueErrorWithRepair("browser_act", "fill", repairable, safeAutorepair, repairs)
	}
	field := BrowserFillField{
		Ref:      target.Ref,
		Hint:     browserElementHintForTarget(target),
		Selector: target.Selector,
		Type:     strings.TrimSpace(firstString(params, "type")),
		Values:   append([]string(nil), values...),
	}
	field.Value = values[0]
	return field, nil
}
