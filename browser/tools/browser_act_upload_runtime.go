package tools

import (
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserActUploadLocator struct {
	Ref         string
	InputRef    string
	Selector    string
	ElementHint *BrowserElementHint
}

func browserActReviewBlockedResultWithPaths(
	kind string,
	runtimeInfo BrowserRuntimeInfo,
	browserApp string,
	target browserToolTarget,
	force bool,
	note string,
	paths []string,
) BrowserActResult {
	result := browserActReviewBlockedResult(kind, runtimeInfo, browserApp, target, force, note)
	result.Paths = append([]string(nil), paths...)
	return result
}

func browserActResolveUploadLocator(params map[string]any) (browserActUploadLocator, error) {
	ref := strings.TrimSpace(firstString(params, "ref", "element_ref"))
	inputRef := strings.TrimSpace(firstString(params, "input_ref"))
	selector := strings.TrimSpace(firstString(params, "selector", "element"))
	elementRef := firstNonEmpty(inputRef, ref)
	locator := browserActUploadLocator{
		Ref:      ref,
		InputRef: inputRef,
		Selector: selector,
	}
	if selector == "" && !browserElementRefHasKnownPrefix(elementRef) {
		return locator, nil
	}
	elementTarget, err := resolveBrowserElementTarget(selector, elementRef)
	if err != nil {
		return browserActUploadLocator{}, fmt.Errorf("browser_act: %w", err)
	}
	locator.Selector = elementTarget.Selector
	locator.ElementHint = browserElementHintForTarget(elementTarget)
	if inputRef != "" {
		locator.InputRef = elementTarget.Ref
		locator.Ref = ""
	} else if ref != "" {
		locator.Ref = elementTarget.Ref
	}
	return locator, nil
}

func browserActUploadDefaultStatus(locator browserActUploadLocator) string {
	if strings.TrimSpace(locator.InputRef) != "" || strings.TrimSpace(locator.Ref) != "" || strings.TrimSpace(locator.Selector) != "" {
		return "uploaded"
	}
	return "armed"
}

func browserActExecuteUpload(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	uploadBackend, ok := pageCtx.RoutedBackend.(BrowserUploadActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "upload")
	}
	waitMs := firstInt(params, "wait_ms")
	if waitMs <= 0 {
		waitMs = pageCtx.DefaultWaitMs
	}
	resolvedPaths, displayPaths, err := resolveBrowserUploadPaths(pageCtx.Options.Root, params)
	if err != nil {
		return BrowserActResult{}, err
	}
	reviewNote := agentxbrowserruntime.SharedSessionBrowserActionReviewReason("upload", pageCtx.Force)
	if agentxbrowserruntime.SharedSessionBrowserActionRequiresReview("upload") && !pageCtx.Force {
		return browserActReviewBlockedResultWithPaths(
			"upload",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			pageCtx.Force,
			reviewNote,
			displayPaths,
		), nil
	}
	locator, err := browserActResolveUploadLocator(params)
	if err != nil {
		return BrowserActResult{}, err
	}
	managed, err := browserResolvedExecutionRouteExecuteManaged(
		pageCtx.CallCtx,
		pageCtx.Route,
		func(_ BrowserBackend) (BrowserUploadResult, error) {
			return uploadBackend.Upload(pageCtx.CallCtx, BrowserUploadRequest{
				BrowserApp:        pageCtx.BrowserApp,
				WaitMs:            waitMs,
				TabIndex:          pageCtx.TabIndex,
				PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
				ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("upload", pageCtx.Force),
				ReviewReady:       pageCtx.Force,
				Note:              reviewNote,
				Paths:             append([]string(nil), resolvedPaths...),
				Ref:               locator.Ref,
				InputRef:          locator.InputRef,
				ElementHint:       locator.ElementHint,
				Selector:          locator.Selector,
			})
		},
		browserManagedRouteExecutionArgs{
			URL:        "",
			BrowserApp: pageCtx.BrowserApp,
			WaitMs:     waitMs,
			TabIndex:   pageCtx.TabIndex,
			Force:      pageCtx.Force,
			FinalURL:   pageCtx.Route.managedFinalURL(pageCtx.CallCtx, pageCtx.BrowserApp, pageCtx.Target, ""),
		},
		func(policy browserManagedResolverFailurePolicyResult) BrowserUploadResult {
			return BrowserUploadResult{
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
	actionResult := browserActActionRuntimeResultEventResultForOutcome(
		pageCtx,
		result.ResolverOutcome,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             strings.TrimSpace(result.Title),
			Source:            "browser_act_upload",
			SetCurrent:        pageCtx.TabIndex > 0,
			ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("upload", pageCtx.Force),
			ReviewReady:       pageCtx.Force,
			Note:              firstNonEmpty(strings.TrimSpace(result.Note), reviewNote),
		},
	)
	targetID := strings.TrimSpace(pageCtx.Target.TargetID)
	if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
		targetID = firstNonEmpty(strings.TrimSpace(actionResult.TargetID), targetID)
	}
	targetID = browserManagedResolverApplyTargetInvalidation(targetID, recovery)
	return browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:           "upload",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       targetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       strings.TrimSpace(result.FinalURL),
		Title:          strings.TrimSpace(result.Title),
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
		Paths:               append([]string(nil), displayPaths...),
		Ref:                 firstNonEmpty(locator.InputRef, locator.Ref),
		Selector:            locator.Selector,
		ResolverOutcome:     result.ResolverOutcome,
		Actionability:       result.Actionability,
		FailureEvidence:     result.FailureEvidence,
		RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
		Status:              firstNonEmpty(strings.TrimSpace(result.Status), browserActUploadDefaultStatus(locator)),
		Force:               pageCtx.Force,
		ReviewDecision:      actionResult.ReviewDecision,
		ReviewReady:         actionResult.ReviewReady,
		Truncated:           recovery.SnapshotTruncated,
		TabIndex:            pageCtx.TabIndex,
		Note:                actionResult.Note,
	}), nil
}
