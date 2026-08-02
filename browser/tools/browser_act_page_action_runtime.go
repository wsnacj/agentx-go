package tools

import (
	"context"
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

const defaultBrowserInteractiveActionWaitMs = 15_000

type browserActPageActionContext struct {
	CallCtx                       context.Context
	Route                         browserResolvedExecutionRoute
	RoutedBackend                 BrowserBackend
	SessionRegistry               *BrowserSessionRegistry
	SharedMutationCtx             agentxbrowserruntime.SharedSessionBrowserMutationContext
	SessionID                     string
	RuntimeInfo                   BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	ExplicitRuntimeTarget         bool
	Target                        browserToolTarget
	BrowserApp                    string
	RequestURL                    string
	Force                         bool
	TabIndex                      int
	DefaultWaitMs                 int
	MaxChars                      int
	Options                       BrowserToolOptions
}

func browserActPageActionWaitMs(requestURL string, requestedWaitMs int, urlWaitMs int, targetWaitMs int) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	if strings.TrimSpace(requestURL) != "" {
		return urlWaitMs
	}
	return targetWaitMs
}

func browserActInteractivePageActionWaitMs(
	requestURL string,
	requestedWaitMs int,
	defaultWaitMs int,
	target browserToolTarget,
) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	waitMs := maxInt(defaultWaitMs, defaultBrowserInteractiveActionWaitMs)
	if strings.TrimSpace(requestURL) != "" {
		return waitMs
	}
	if target.Explicit {
		return maxInt(waitMs, browserTabTargetWaitMs)
	}
	return waitMs
}

func browserActPageReviewBlockedResultWithPath(
	kind string,
	runtimeInfo BrowserRuntimeInfo,
	browserApp string,
	target browserToolTarget,
	reviewState browserActPageReviewState,
	force bool,
	note string,
	path string,
) BrowserActResult {
	result := browserActPageReviewBlockedResult(kind, runtimeInfo, browserApp, target, reviewState, force, note)
	result.Path = strings.TrimSpace(path)
	return result
}

func browserActPageReviewBlockedResultWithFields(
	kind string,
	runtimeInfo BrowserRuntimeInfo,
	browserApp string,
	target browserToolTarget,
	reviewState browserActPageReviewState,
	force bool,
	note string,
	ref string,
	selector string,
	value string,
) BrowserActResult {
	result := browserActPageReviewBlockedResult(kind, runtimeInfo, browserApp, target, reviewState, force, note)
	result.Ref = strings.TrimSpace(ref)
	result.Selector = strings.TrimSpace(selector)
	result.Value = value
	return result
}

func browserActExecuteExtract(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind extract", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResult(
			"extract",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act extract", review, pageCtx.Force),
		), nil
	}
	waitMs := browserActPageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, browserTabTargetWaitMs)
	requestMaxChars := firstInt(params, "max_chars")
	if requestMaxChars <= 0 || requestMaxChars > pageCtx.MaxChars {
		requestMaxChars = pageCtx.MaxChars
	}
	result, err := pageCtx.RoutedBackend.Extract(pageCtx.CallCtx, BrowserExtractRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		MaxChars:          requestMaxChars,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: reviewState.TargetID,
		Actor:             "browser_act extract",
		Force:             pageCtx.Force,
		Review:            review,
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	content := strings.TrimSpace(result.Content)
	truncated := false
	if trimmed, changed := trimToMaxChars(content, requestMaxChars); changed {
		content = trimmed
		truncated = true
	}
	pageActionResult := applyBrowserActPageActionEventResult(
		pageCtx.SharedMutationCtx,
		pageCtx.SessionID,
		pageCtx.RuntimeInfo,
		pageCtx.BrowserApp,
		pageCtx.RequestURL,
		browserActPageActionEventApplyOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			ResultFinalURL:    result.FinalURL,
			ResultTitle:       result.Title,
			Source:            "browser_act_extract",
			Actor:             "browser_act extract",
			Force:             pageCtx.Force,
			TabIndex:          pageCtx.TabIndex,
			PreferredTargetID: reviewState.TargetID,
			Review:            review,
		},
	)
	return BrowserActResult{
		Kind:           "extract",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       pageActionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       result.FinalURL,
		Title:          result.Title,
		Content:        content,
		ContentType:    result.ContentType,
		Status:         "extracted",
		Force:          pageCtx.Force,
		ReviewDecision: pageActionResult.ReviewDecision,
		ReviewReady:    pageActionResult.ReviewReady,
		Truncated:      truncated,
		TabIndex:       pageCtx.TabIndex,
		Note:           pageActionResult.Note,
	}, nil
}

func browserActExecuteSnapshot(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind snapshot", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResult(
			"snapshot",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act snapshot", review, pageCtx.Force),
		), nil
	}
	requestMaxChars := firstInt(params, "max_chars")
	if requestMaxChars <= 0 || requestMaxChars > pageCtx.MaxChars {
		requestMaxChars = pageCtx.MaxChars
	}
	requestMaxElements := firstInt(params, "max_elements")
	if requestMaxElements <= 0 {
		requestMaxElements = 16
	}
	if requestMaxElements > 24 {
		requestMaxElements = 24
	}
	waitMs := browserActPageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, 250)
	snapshotFormat := browserCanonicalSnapshotFormat(firstString(params, "snapshot_format", "format"))
	snapshotMode := browserNormalizeToolToken(firstString(params, "mode"))
	if firstBool(params, "efficient") {
		if snapshotMode != "" && snapshotMode != "efficient" {
			return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"mode", "efficient"}, fmt.Sprintf("browser_act: mode %q conflicts with efficient=true for kind snapshot", snapshotMode))
		}
		snapshotMode = "efficient"
	}
	if snapshotMode != "" && snapshotMode != "efficient" {
		return BrowserActResult{}, browserInvalidSnapshotModeError("browser_act")
	}
	snapshotRefs := browserNormalizeToolToken(firstString(params, "refs"))
	if snapshotRefs != "" && snapshotRefs != "aria" && snapshotRefs != "role" {
		return BrowserActResult{}, browserInvalidSnapshotRefsError("browser_act")
	}
	snapshotInteractive := firstBool(params, "interactive")
	snapshotCompact := firstBool(params, "compact")
	snapshotDepth := firstInt(params, "depth")
	if rawDepth, ok := params["depth"]; ok && rawDepth != nil && snapshotDepth < 1 {
		return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"depth"}, "browser_act: depth must be >= 1 for kind snapshot")
	}
	snapshotSelector := strings.TrimSpace(firstString(params, "selector"))
	snapshotFrame := strings.TrimSpace(firstString(params, "frame"))
	if snapshotRefs == "" && (snapshotInteractive || snapshotCompact || snapshotDepth > 0 || snapshotSelector != "" || snapshotFrame != "") {
		snapshotRefs = "role"
	}
	result, err := pageCtx.RoutedBackend.Snapshot(pageCtx.CallCtx, BrowserSnapshotRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		MaxChars:          requestMaxChars,
		MaxElements:       requestMaxElements,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: reviewState.TargetID,
		Actor:             "browser_act snapshot",
		Force:             pageCtx.Force,
		Review:            review,
		Format:            snapshotFormat,
		Mode:              snapshotMode,
		Refs:              snapshotRefs,
		Interactive:       snapshotInteractive,
		Compact:           snapshotCompact,
		Depth:             snapshotDepth,
		Selector:          snapshotSelector,
		Frame:             snapshotFrame,
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	result.Elements = browserNormalizeSnapshotElements(
		result.Elements,
		firstNonEmpty(strings.TrimSpace(result.FinalURL), pageCtx.RequestURL),
		strings.TrimSpace(result.Title),
	)
	resolvedSnapshotDepth := result.Depth
	if resolvedSnapshotDepth <= 0 {
		resolvedSnapshotDepth = snapshotDepth
	}
	snapshot := strings.TrimSpace(result.Snapshot)
	truncated := result.Truncated
	if trimmed, changed := trimToMaxChars(snapshot, requestMaxChars); changed {
		snapshot = trimmed
		truncated = true
	}
	pageActionResult := applyBrowserActPageActionEventResult(
		pageCtx.SharedMutationCtx,
		pageCtx.SessionID,
		pageCtx.RuntimeInfo,
		pageCtx.BrowserApp,
		pageCtx.RequestURL,
		browserActPageActionEventApplyOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			ResultFinalURL:    result.FinalURL,
			ResultTitle:       result.Title,
			Source:            "browser_act_snapshot",
			Actor:             "browser_act snapshot",
			Force:             pageCtx.Force,
			TabIndex:          pageCtx.TabIndex,
			PreferredTargetID: reviewState.TargetID,
			Review:            review,
		},
	)
	return BrowserActResult{
		Kind:                "snapshot",
		Backend:             result.Backend,
		BrowserApp:          firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:              pageCtx.Target.Value,
		TargetID:            pageActionResult.TargetID,
		Profile:             pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:       pageCtx.RuntimeInfo.Target,
		FinalURL:            result.FinalURL,
		Title:               result.Title,
		Snapshot:            snapshot,
		SnapshotFormat:      firstNonEmpty(strings.TrimSpace(result.Format), snapshotFormat),
		SnapshotMode:        firstNonEmpty(strings.TrimSpace(result.Mode), snapshotMode),
		SnapshotRefs:        firstNonEmpty(strings.TrimSpace(result.Refs), snapshotRefs),
		SnapshotInteractive: result.Interactive || snapshotInteractive,
		SnapshotCompact:     result.Compact || snapshotCompact,
		SnapshotDepth:       resolvedSnapshotDepth,
		SnapshotFrame:       firstNonEmpty(strings.TrimSpace(result.Frame), snapshotFrame),
		Elements:            append([]BrowserSnapshotElement(nil), result.Elements...),
		Status:              "snapshotted",
		Force:               pageCtx.Force,
		ReviewDecision:      pageActionResult.ReviewDecision,
		ReviewReady:         pageActionResult.ReviewReady,
		Selector:            firstNonEmpty(strings.TrimSpace(result.Selector), snapshotSelector),
		Truncated:           truncated,
		TabIndex:            pageCtx.TabIndex,
		Note:                browserActPageActionResultNote(result.Note, pageActionResult.Note, review),
	}, nil
}

func browserActExecuteScreenshot(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind screenshot", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	urlWaitMs := pageCtx.Options.ScreenshotWaitMs
	if urlWaitMs <= 0 {
		urlWaitMs = 2_500
	}
	waitMs := browserActPageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), urlWaitMs, browserTabTargetWaitMs)
	elementTarget, err := resolveBrowserElementTarget(firstString(params, "selector"), firstString(params, "ref", "element_ref"))
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	if err := browserValidateElementTargetPageBinding(pageCtx.CallCtx, pageCtx.SessionRegistry, pageCtx.RuntimeInfo, pageCtx.HiddenImplicitHostDefaultBase, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL, elementTarget); err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	selector := elementTarget.Selector
	fullPage := firstBool(params, "full_page")
	if selector != "" && fullPage {
		return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"ref", "selector", "full_page"}, "browser_act: ref/selector and full_page cannot be used together for kind screenshot")
	}
	pathValue := firstString(params, "path", "output", "output_path")
	if strings.TrimSpace(pathValue) == "" {
		pathValue = defaultBrowserScreenshotRelPath()
	}
	resolvedPath, displayPath, err := resolvePathWithinRoot(pageCtx.Options.Root, pathValue)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResultWithPath(
			"screenshot",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act screenshot", review, pageCtx.Force),
			displayPath,
		), nil
	}
	var result BrowserScreenshotResult
	var recovery browserManagedResolverRecoveryResult
	artifactBytes, artifactPublished, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"screenshot",
		resolvedPath,
		func(stagePath string) (string, bool, error) {
			managed, executeErr := browserResolvedExecutionRouteExecuteManaged(
				pageCtx.CallCtx,
				pageCtx.Route,
				func(backend BrowserBackend) (BrowserScreenshotResult, error) {
					return backend.Screenshot(pageCtx.CallCtx, BrowserScreenshotRequest{
						URL:               pageCtx.RequestURL,
						BrowserApp:        pageCtx.BrowserApp,
						WaitMs:            waitMs,
						OutputPath:        stagePath,
						ElementRef:        elementTarget.Ref,
						ElementHint:       browserElementHintForTarget(elementTarget),
						Selector:          selector,
						FullPage:          fullPage,
						TabIndex:          pageCtx.TabIndex,
						PreferredTargetID: reviewState.TargetID,
						Actor:             "browser_act screenshot",
						Force:             pageCtx.Force,
						Review:            review,
					})
				},
				browserManagedRouteExecutionArgs{
					URL:        pageCtx.RequestURL,
					BrowserApp: pageCtx.BrowserApp,
					WaitMs:     waitMs,
					TabIndex:   pageCtx.TabIndex,
					Force:      pageCtx.Force,
					FinalURL:   pageCtx.Route.managedFinalURL(pageCtx.CallCtx, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL),
				},
				func(policy browserManagedResolverFailurePolicyResult) BrowserScreenshotResult {
					return BrowserScreenshotResult{
						Backend:         policy.Backend,
						BrowserApp:      policy.BrowserApp,
						Path:            displayPath,
						FinalURL:        policy.FinalURL,
						Title:           policy.Title,
						Status:          policy.Status,
						Note:            policy.Note,
						ResolverOutcome: policy.Outcome,
					}
				},
			)
			if executeErr != nil {
				return "", false, executeErr
			}
			result = managed.Result
			recovery = managed.Recovery
			return result.Path, browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome), nil
		},
	)
	if err != nil {
		return BrowserActResult{}, err
	}
	snapshotRecovery := browserManagedResolverSnapshotPayloadForRecovery(recovery)
	artifactDisplayPath := displayPath
	if !artifactPublished {
		artifactDisplayPath = displayPath
	}
	pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
		TargetID: strings.TrimSpace(reviewState.TargetID),
	}
	if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
		pageActionResult = applyBrowserActPageActionEventResult(
			pageCtx.SharedMutationCtx,
			pageCtx.SessionID,
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.RequestURL,
			browserActPageActionEventApplyOptions{
				ResultBackend:     result.Backend,
				ResultBrowserApp:  result.BrowserApp,
				ResultFinalURL:    result.FinalURL,
				ResultTitle:       result.Title,
				Source:            "browser_act_screenshot",
				Actor:             "browser_act screenshot",
				Force:             pageCtx.Force,
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: reviewState.TargetID,
				Review:            review,
			},
		)
	}
	targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
	return browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:                "screenshot",
		Backend:             result.Backend,
		BrowserApp:          firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:              pageCtx.Target.Value,
		TargetID:            targetID,
		Profile:             pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:       pageCtx.RuntimeInfo.Target,
		FinalURL:            firstNonEmpty(result.FinalURL, pageCtx.RequestURL),
		Title:               result.Title,
		Snapshot:            snapshotRecovery.Text,
		SnapshotFormat:      snapshotRecovery.Format,
		SnapshotMode:        snapshotRecovery.Mode,
		SnapshotRefs:        snapshotRecovery.Refs,
		SnapshotInteractive: snapshotRecovery.Interactive,
		SnapshotCompact:     snapshotRecovery.Compact,
		SnapshotDepth:       snapshotRecovery.Depth,
		SnapshotFrame:       snapshotRecovery.Frame,
		Elements:            snapshotRecovery.Elements,
		Path:                artifactDisplayPath,
		FilesTouched:        browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:               artifactBytes,
		CaptureScope:        result.CaptureScope,
		CaptureWidth:        result.CaptureWidth,
		CaptureHeight:       result.CaptureHeight,
		Ref:                 elementTarget.Ref,
		Selector:            selector,
		ResolverOutcome:     result.ResolverOutcome,
		Actionability:       result.Actionability,
		FailureEvidence:     result.FailureEvidence,
		RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
		Status:              firstNonEmpty(strings.TrimSpace(result.Status), "captured"),
		Force:               pageCtx.Force,
		ReviewDecision:      pageActionResult.ReviewDecision,
		ReviewReady:         pageActionResult.ReviewReady,
		Truncated:           snapshotRecovery.Truncated,
		TabIndex:            pageCtx.TabIndex,
		Note:                browserActPageActionResultNote(result.Note, pageActionResult.Note, review),
	}), nil
}

func browserActExecuteClick(pageCtx browserActPageActionContext, params map[string]any, postWaitMs int) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind click", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	elementTarget, err := resolveBrowserActionElementTarget(
		firstString(params, "selector"),
		firstString(params, "ref", "element_ref"),
		browserClickElementHintValue(params),
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	if err := browserValidateElementTargetPageBinding(pageCtx.CallCtx, pageCtx.SessionRegistry, pageCtx.RuntimeInfo, pageCtx.HiddenImplicitHostDefaultBase, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL, elementTarget); err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	selector := elementTarget.Selector
	if !browserElementTargetHasActionableLocator(elementTarget) {
		repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "element", "text", "label")
		return BrowserActResult{}, browserMissingLocatorError(
			"browser_act",
			"browser_act: selector or ref is required for kind click",
			repairable,
			safeAutorepair,
			repairs,
		)
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	if postWaitMs <= 0 {
		postWaitMs = 750
	}
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResultWithFields(
			"click",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act click", review, pageCtx.Force),
			elementTarget.Ref,
			selector,
			"",
		), nil
	}
	managed, err := browserResolvedExecutionRouteExecuteManaged(
		pageCtx.CallCtx,
		pageCtx.Route,
		func(backend BrowserBackend) (BrowserClickResult, error) {
			return backend.Click(pageCtx.CallCtx, BrowserClickRequest{
				URL:               pageCtx.RequestURL,
				BrowserApp:        pageCtx.BrowserApp,
				WaitMs:            waitMs,
				PostWaitMs:        postWaitMs,
				ElementRef:        elementTarget.Ref,
				ElementHint:       browserClickElementHintForTarget(elementTarget),
				Selector:          selector,
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: reviewState.TargetID,
				Actor:             "browser_act click",
				Force:             pageCtx.Force,
				Review:            review,
			})
		},
		browserManagedRouteExecutionArgs{
			URL:        pageCtx.RequestURL,
			BrowserApp: pageCtx.BrowserApp,
			WaitMs:     waitMs,
			TabIndex:   pageCtx.TabIndex,
			Force:      pageCtx.Force,
			FinalURL:   pageCtx.Route.managedFinalURL(pageCtx.CallCtx, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL),
		},
		func(policy browserManagedResolverFailurePolicyResult) BrowserClickResult {
			return BrowserClickResult{
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
	pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
		TargetID: reviewState.TargetID,
	}
	if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
		pageActionResult = applyBrowserActPageActionEventResult(
			pageCtx.SharedMutationCtx,
			pageCtx.SessionID,
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			"",
			browserActPageActionEventApplyOptions{
				ResultBackend:     result.Backend,
				ResultBrowserApp:  result.BrowserApp,
				ResultFinalURL:    result.FinalURL,
				ResultTitle:       result.Title,
				Source:            "browser_act_click",
				Actor:             "browser_act click",
				Force:             pageCtx.Force,
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: reviewState.TargetID,
				Review:            review,
			},
		)
	}
	targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
	return browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:           "click",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       targetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       result.FinalURL,
		Title:          result.Title,
		Snapshot:       recovery.SnapshotText,
		SnapshotFormat: strings.TrimSpace(recovery.Snapshot.Format),
		SnapshotMode:   strings.TrimSpace(recovery.Snapshot.Mode),
		SnapshotRefs: func() string {
			if recovery.SnapshotRecovered {
				return firstNonEmpty(strings.TrimSpace(recovery.Snapshot.Refs), "role")
			}
			return ""
		}(),
		SnapshotInteractive: recovery.Snapshot.Interactive || recovery.SnapshotRecovered,
		SnapshotCompact:     recovery.Snapshot.Compact,
		SnapshotDepth:       recovery.Snapshot.Depth,
		SnapshotFrame:       strings.TrimSpace(recovery.Snapshot.Frame),
		Elements:            append([]BrowserSnapshotElement(nil), recovery.Snapshot.Elements...),
		Ref:                 elementTarget.Ref,
		Selector:            selector,
		ResolverOutcome:     result.ResolverOutcome,
		Actionability:       result.Actionability,
		FailureEvidence:     result.FailureEvidence,
		RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
		Status:              result.Status,
		Force:               pageCtx.Force,
		ReviewDecision:      pageActionResult.ReviewDecision,
		ReviewReady:         pageActionResult.ReviewReady,
		Truncated:           recovery.SnapshotTruncated,
		TabIndex:            pageCtx.TabIndex,
		Note:                browserActPageActionResultNote(result.Note, pageActionResult.Note, review),
	}), nil
}

func browserActExecuteType(pageCtx browserActPageActionContext, params map[string]any, postWaitMs int) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind type", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	elementTarget, err := resolveBrowserActionElementTarget(
		firstString(params, "selector"),
		firstString(params, "ref", "element_ref"),
		firstString(params, "element"),
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	if err := browserValidateElementTargetPageBinding(pageCtx.CallCtx, pageCtx.SessionRegistry, pageCtx.RuntimeInfo, pageCtx.HiddenImplicitHostDefaultBase, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL, elementTarget); err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	selector := elementTarget.Selector
	if !browserElementTargetHasActionableLocator(elementTarget) {
		repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "element")
		return BrowserActResult{}, browserMissingLocatorError(
			"browser_act",
			"browser_act: selector or ref is required for kind type",
			repairable,
			safeAutorepair,
			repairs,
		)
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	if postWaitMs <= 0 {
		postWaitMs = 250
	}
	value := firstString(params, "text", "value")
	if strings.TrimSpace(value) == "" {
		repairable, safeAutorepair, repairs := browserTypeValueRepairAdviceFromParams(params)
		return BrowserActResult{}, browserMissingTypeTextErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
	}
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResultWithFields(
			"type",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act type", review, pageCtx.Force),
			elementTarget.Ref,
			selector,
			value,
		), nil
	}
	managed, err := browserResolvedExecutionRouteExecuteManaged(
		pageCtx.CallCtx,
		pageCtx.Route,
		func(backend BrowserBackend) (BrowserTypeResult, error) {
			return backend.Type(pageCtx.CallCtx, BrowserTypeRequest{
				URL:               pageCtx.RequestURL,
				BrowserApp:        pageCtx.BrowserApp,
				WaitMs:            waitMs,
				PostWaitMs:        postWaitMs,
				ElementRef:        elementTarget.Ref,
				ElementHint:       browserElementHintForTarget(elementTarget),
				Selector:          selector,
				Text:              value,
				Submit:            firstBool(params, "submit"),
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: reviewState.TargetID,
				Actor:             "browser_act type",
				Force:             pageCtx.Force,
				Review:            review,
			})
		},
		browserManagedRouteExecutionArgs{
			URL:        pageCtx.RequestURL,
			BrowserApp: pageCtx.BrowserApp,
			WaitMs:     waitMs,
			TabIndex:   pageCtx.TabIndex,
			Force:      pageCtx.Force,
			FinalURL:   pageCtx.Route.managedFinalURL(pageCtx.CallCtx, pageCtx.BrowserApp, pageCtx.Target, pageCtx.RequestURL),
		},
		func(policy browserManagedResolverFailurePolicyResult) BrowserTypeResult {
			return BrowserTypeResult{
				Backend:         policy.Backend,
				BrowserApp:      policy.BrowserApp,
				FinalURL:        policy.FinalURL,
				Title:           policy.Title,
				Value:           value,
				Status:          policy.Status,
				ResolverOutcome: policy.Outcome,
				Note:            policy.Note,
			}
		},
	)
	if err != nil {
		return BrowserActResult{}, err
	}
	result := managed.Result
	recovery := managed.Recovery
	pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
		TargetID: reviewState.TargetID,
	}
	if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
		pageActionResult = applyBrowserActPageActionEventResult(
			pageCtx.SharedMutationCtx,
			pageCtx.SessionID,
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			"",
			browserActPageActionEventApplyOptions{
				ResultBackend:     result.Backend,
				ResultBrowserApp:  result.BrowserApp,
				ResultFinalURL:    result.FinalURL,
				ResultTitle:       result.Title,
				Source:            "browser_act_type",
				Actor:             "browser_act type",
				Force:             pageCtx.Force,
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: reviewState.TargetID,
				Review:            review,
			},
		)
	}
	targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
	return browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:           "type",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       targetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       result.FinalURL,
		Title:          result.Title,
		Snapshot:       recovery.SnapshotText,
		SnapshotFormat: strings.TrimSpace(recovery.Snapshot.Format),
		SnapshotMode:   strings.TrimSpace(recovery.Snapshot.Mode),
		SnapshotRefs: func() string {
			if recovery.SnapshotRecovered {
				return firstNonEmpty(strings.TrimSpace(recovery.Snapshot.Refs), "role")
			}
			return ""
		}(),
		SnapshotInteractive: recovery.Snapshot.Interactive || recovery.SnapshotRecovered,
		SnapshotCompact:     recovery.Snapshot.Compact,
		SnapshotDepth:       recovery.Snapshot.Depth,
		SnapshotFrame:       strings.TrimSpace(recovery.Snapshot.Frame),
		Elements:            append([]BrowserSnapshotElement(nil), recovery.Snapshot.Elements...),
		Ref:                 elementTarget.Ref,
		Selector:            selector,
		ResolverOutcome:     result.ResolverOutcome,
		Actionability:       result.Actionability,
		FailureEvidence:     result.FailureEvidence,
		RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
		Value:               result.Value,
		Status:              result.Status,
		Submitted:           result.Submitted,
		Force:               pageCtx.Force,
		ReviewDecision:      pageActionResult.ReviewDecision,
		ReviewReady:         pageActionResult.ReviewReady,
		Truncated:           recovery.SnapshotTruncated,
		TabIndex:            pageCtx.TabIndex,
		Note:                browserActPageActionResultNote(result.Note, pageActionResult.Note, review),
	}), nil
}

func browserActExecuteEvaluate(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_act kind evaluate", pageCtx.HiddenImplicitHostDefaultBase, pageCtx.ExplicitRuntimeTarget, pageCtx.RuntimeInfo, pageCtx.Target, pageCtx.RequestURL); err != nil {
		return BrowserActResult{}, err
	}
	reviewState := resolveBrowserActPageReviewState(
		pageCtx.CallCtx,
		pageCtx.SessionRegistry,
		pageCtx.RuntimeInfo,
		pageCtx.HiddenImplicitHostDefaultBase,
		pageCtx.BrowserApp,
		pageCtx.Target,
		pageCtx.RequestURL,
	)
	review := reviewState.Review
	script := strings.TrimSpace(firstString(params, "script", "javascript", "js"))
	if script == "" {
		repairable, safeAutorepair, repairs := browserEvaluateScriptRepairAdviceFromParams(params)
		return BrowserActResult{}, browserMissingScriptErrorWithRepair("browser_act", repairable, safeAutorepair, repairs)
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	requestMaxChars := firstInt(params, "max_chars")
	if requestMaxChars <= 0 || requestMaxChars > pageCtx.MaxChars {
		requestMaxChars = pageCtx.MaxChars
	}
	if review.Review != nil && !pageCtx.Force {
		return browserActPageReviewBlockedResult(
			"evaluate",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			reviewState,
			pageCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act evaluate", review, pageCtx.Force),
		), nil
	}
	if _, err := browserCheckCDPEscapeHatchPolicy(pageCtx.Options); err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	result, err := pageCtx.RoutedBackend.Eval(pageCtx.CallCtx, BrowserEvalRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		Script:            script,
		MaxChars:          requestMaxChars,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: reviewState.TargetID,
		Actor:             "browser_act evaluate",
		Force:             pageCtx.Force,
		Review:            review,
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	rendered := strings.TrimSpace(result.Result)
	truncated := false
	if trimmed, changed := trimToMaxChars(rendered, requestMaxChars); changed {
		rendered = trimmed
		truncated = true
	}
	pageActionResult := applyBrowserActPageActionEventResult(
		pageCtx.SharedMutationCtx,
		pageCtx.SessionID,
		pageCtx.RuntimeInfo,
		pageCtx.BrowserApp,
		"",
		browserActPageActionEventApplyOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			ResultFinalURL:    result.FinalURL,
			ResultTitle:       result.Title,
			Source:            "browser_act_evaluate",
			Actor:             "browser_act evaluate",
			Force:             pageCtx.Force,
			TabIndex:          pageCtx.TabIndex,
			PreferredTargetID: reviewState.TargetID,
			Review:            review,
		},
	)
	return BrowserActResult{
		Kind:           "evaluate",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       pageActionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       result.FinalURL,
		Title:          result.Title,
		Result:         rendered,
		Status:         result.Status,
		Force:          pageCtx.Force,
		ReviewDecision: pageActionResult.ReviewDecision,
		ReviewReady:    pageActionResult.ReviewReady,
		Truncated:      truncated,
		TabIndex:       pageCtx.TabIndex,
		Note:           browserActPageActionResultNote(result.Note, pageActionResult.Note, review),
	}, nil
}
