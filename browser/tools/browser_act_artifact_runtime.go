package tools

import (
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserActArtifactOutputPath struct {
	Resolved string
	Display  string
}

type browserActActionRuntimeEventOptions struct {
	ResultBackend     string
	ResultBrowserApp  string
	PreferredTargetID string
	TrackCurrent      bool
	FinalURL          string
	Title             string
	Source            string
	SetCurrent        bool
	ReviewDecision    string
	ReviewReady       bool
	Note              string
}

func browserActWaitDownloadWaitMs(requestedWaitMs int, defaultWaitMs int) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	return maxInt(defaultWaitMs, defaultBrowserWaitDownloadMs)
}

func browserActTraceWaitMs(requestedWaitMs int, target browserToolTarget) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	if target.Explicit {
		return browserTabTargetWaitMs
	}
	return 0
}

func browserActResolveArtifactOutputPath(root string, pathValue string) (browserActArtifactOutputPath, error) {
	resolvedPath, displayPath, err := resolvePathWithinRoot(root, pathValue)
	if err != nil {
		return browserActArtifactOutputPath{}, err
	}
	return browserActArtifactOutputPath{
		Resolved: resolvedPath,
		Display:  displayPath,
	}, nil
}

func browserActResolveOptionalArtifactOutputPath(root string, pathValue string) (browserActArtifactOutputPath, error) {
	if strings.TrimSpace(pathValue) == "" {
		return browserActArtifactOutputPath{}, nil
	}
	return browserActResolveArtifactOutputPath(root, pathValue)
}

func browserActReviewBlockedResultWithPath(
	kind string,
	runtimeInfo BrowserRuntimeInfo,
	browserApp string,
	target browserToolTarget,
	force bool,
	note string,
	path string,
) BrowserActResult {
	result := browserActReviewBlockedResult(kind, runtimeInfo, browserApp, target, force, note)
	result.Path = strings.TrimSpace(path)
	return result
}

func browserActApplyActionRuntimeResultEvent(
	pageCtx browserActPageActionContext,
	options browserActActionRuntimeEventOptions,
) agentxbrowserruntime.SharedSessionBrowserActionResultEventResult {
	return agentxbrowserruntime.ApplySharedSessionBrowserActionResultWithContext(
		pageCtx.SharedMutationCtx,
		agentxbrowserruntime.SharedSessionBrowserActionResultEventRequest{
			SessionID:         pageCtx.SessionID,
			Route:             browserSessionRoute(pageCtx.RuntimeInfo, firstNonEmpty(strings.TrimSpace(options.ResultBrowserApp), pageCtx.BrowserApp), strings.TrimSpace(options.ResultBackend)),
			PreferredTargetID: strings.TrimSpace(options.PreferredTargetID),
			TabIndex:          pageCtx.TabIndex,
			TrackCurrent:      options.TrackCurrent,
			URL:               strings.TrimSpace(options.FinalURL),
			Title:             strings.TrimSpace(options.Title),
			Source:            strings.TrimSpace(options.Source),
			SetCurrent:        options.SetCurrent,
			ReviewDecision:    strings.TrimSpace(options.ReviewDecision),
			ReviewReady:       options.ReviewReady,
			Note:              strings.TrimSpace(options.Note),
		},
	)
}

func browserActExecuteDownload(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	fileBackend, ok := pageCtx.RoutedBackend.(BrowserArtifactActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "download")
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	outputPath, err := browserActResolveOptionalArtifactOutputPath(
		pageCtx.Options.Root,
		firstString(params, "path", "output", "output_path"),
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	reviewNote := agentxbrowserruntime.SharedSessionBrowserActionReviewReason("download", pageCtx.Force)
	if agentxbrowserruntime.SharedSessionBrowserActionRequiresReview("download") && !pageCtx.Force {
		return browserActReviewBlockedResultWithPath(
			"download",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			pageCtx.Force,
			reviewNote,
			outputPath.Display,
		), nil
	}
	request := BrowserDownloadRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("download", pageCtx.Force),
		ReviewReady:       pageCtx.Force,
		Note:              reviewNote,
	}
	var result BrowserDownloadResult
	backendCalled := false
	if outputPath.Resolved == "" {
		result, err = fileBackend.Download(pageCtx.CallCtx, request)
		if err != nil {
			return BrowserActResult{}, err
		}
		backendCalled = true
		outputPath, err = browserActResolveArtifactOutputPath(
			pageCtx.Options.Root,
			defaultBrowserDownloadRelPath(result.Path, pageCtx.RequestURL, result.FinalURL),
		)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
	}
	artifactBytes, _, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"download",
		outputPath.Resolved,
		func(stagePath string) (string, bool, error) {
			if !backendCalled {
				stageRequest := request
				stageRequest.OutputPath = stagePath
				var downloadErr error
				result, downloadErr = fileBackend.Download(pageCtx.CallCtx, stageRequest)
				if downloadErr != nil {
					return "", false, downloadErr
				}
				backendCalled = true
			}
			return result.Path, true, nil
		},
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	artifactDisplayPath := outputPath.Display
	finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), pageCtx.RequestURL)
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          finalURL,
			Title:             result.Title,
			Source:            "browser_act_download",
			SetCurrent:        pageCtx.TabIndex > 0,
			ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("download", pageCtx.Force),
			ReviewReady:       pageCtx.Force,
			Note:              firstNonEmpty(strings.TrimSpace(result.Note), reviewNote),
		},
	)
	return BrowserActResult{
		Kind:           "download",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       actionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       finalURL,
		Title:          result.Title,
		Path:           artifactDisplayPath,
		FilesTouched:   browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:          artifactBytes,
		ContentType:    strings.TrimSpace(result.ContentType),
		Status:         "downloaded",
		Force:          pageCtx.Force,
		ReviewDecision: actionResult.ReviewDecision,
		ReviewReady:    actionResult.ReviewReady,
		TabIndex:       pageCtx.TabIndex,
		Note:           actionResult.Note,
	}, nil
}

func browserActExecuteWaitDownload(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	fileBackend, ok := pageCtx.RoutedBackend.(BrowserArtifactActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "wait_download")
	}
	waitMs := browserActWaitDownloadWaitMs(firstInt(params, "wait_ms", "timeout_ms"), pageCtx.DefaultWaitMs)
	outputPath, err := browserActResolveOptionalArtifactOutputPath(
		pageCtx.Options.Root,
		firstString(params, "path", "output", "output_path"),
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	reviewNote := agentxbrowserruntime.SharedSessionBrowserActionReviewReason("wait_download", pageCtx.Force)
	if agentxbrowserruntime.SharedSessionBrowserActionRequiresReview("wait_download") && !pageCtx.Force {
		return browserActReviewBlockedResultWithPath(
			"wait_download",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			pageCtx.Force,
			reviewNote,
			outputPath.Display,
		), nil
	}
	request := BrowserWaitDownloadRequest{
		BrowserApp:               pageCtx.BrowserApp,
		WaitMs:                   waitMs,
		AllowRecentDownloadReuse: firstBool(params, "allow_recent_download_reuse", "AllowRecentDownloadReuse"),
		TabIndex:                 pageCtx.TabIndex,
		PreferredTargetID:        strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision:           agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("wait_download", pageCtx.Force),
		ReviewReady:              pageCtx.Force,
		Note:                     reviewNote,
	}
	var result BrowserWaitDownloadResult
	backendCalled := false
	if outputPath.Resolved == "" {
		result, err = fileBackend.WaitDownload(pageCtx.CallCtx, request)
		if err != nil {
			return BrowserActResult{}, err
		}
		backendCalled = true
		outputPath, err = browserActResolveArtifactOutputPath(
			pageCtx.Options.Root,
			defaultBrowserDownloadRelPath(result.Path, result.FinalURL),
		)
		if err != nil {
			return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
		}
	}
	artifactBytes, _, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"wait_download",
		outputPath.Resolved,
		func(stagePath string) (string, bool, error) {
			if !backendCalled {
				stageRequest := request
				stageRequest.OutputPath = stagePath
				var waitErr error
				result, waitErr = fileBackend.WaitDownload(pageCtx.CallCtx, stageRequest)
				if waitErr != nil {
					return "", false, waitErr
				}
				backendCalled = true
			}
			return result.Path, true, nil
		},
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	artifactDisplayPath := outputPath.Display
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             result.Title,
			Source:            "browser_act_wait_download",
			SetCurrent:        true,
			ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("wait_download", pageCtx.Force),
			ReviewReady:       pageCtx.Force,
			Note: firstNonEmpty(
				strings.TrimSpace(result.Note),
				firstNonEmpty(
					reviewNote,
					fmt.Sprintf("waited up to %dms for next download", waitMs),
				),
			),
		},
	)
	return BrowserActResult{
		Kind:           "wait_download",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       actionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       strings.TrimSpace(result.FinalURL),
		Title:          result.Title,
		Path:           artifactDisplayPath,
		FilesTouched:   browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:          artifactBytes,
		ContentType:    strings.TrimSpace(result.ContentType),
		Status:         "downloaded",
		Force:          pageCtx.Force,
		ReviewDecision: actionResult.ReviewDecision,
		ReviewReady:    actionResult.ReviewReady,
		TabIndex:       pageCtx.TabIndex,
		Note:           actionResult.Note,
	}, nil
}

func browserActExecuteSavePDF(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	fileBackend, ok := pageCtx.RoutedBackend.(BrowserArtifactActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "save_pdf")
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	pathValue := firstString(params, "path", "output", "output_path")
	if strings.TrimSpace(pathValue) == "" {
		pathValue = defaultBrowserPDFRelPath()
	}
	outputPath, err := browserActResolveArtifactOutputPath(pageCtx.Options.Root, pathValue)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	reviewNote := agentxbrowserruntime.SharedSessionBrowserActionReviewReason("save_pdf", pageCtx.Force)
	if agentxbrowserruntime.SharedSessionBrowserActionRequiresReview("save_pdf") && !pageCtx.Force {
		return browserActReviewBlockedResultWithPath(
			"save_pdf",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			pageCtx.Force,
			reviewNote,
			outputPath.Display,
		), nil
	}
	request := BrowserSavePDFRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("save_pdf", pageCtx.Force),
		ReviewReady:       pageCtx.Force,
		Note:              reviewNote,
		Landscape:         firstBool(params, "landscape"),
		PrintBackground:   firstBool(params, "print_background"),
	}
	var result BrowserSavePDFResult
	artifactBytes, _, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"save_pdf",
		outputPath.Resolved,
		func(stagePath string) (string, bool, error) {
			stageRequest := request
			stageRequest.OutputPath = stagePath
			var saveErr error
			result, saveErr = fileBackend.SavePDF(pageCtx.CallCtx, stageRequest)
			if saveErr != nil {
				return "", false, saveErr
			}
			return result.Path, true, nil
		},
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	artifactDisplayPath := outputPath.Display
	finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), pageCtx.RequestURL)
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          finalURL,
			Title:             result.Title,
			Source:            "browser_act_save_pdf",
			SetCurrent:        pageCtx.TabIndex > 0,
			ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("save_pdf", pageCtx.Force),
			ReviewReady:       pageCtx.Force,
			Note:              firstNonEmpty(strings.TrimSpace(result.Note), reviewNote),
		},
	)
	return BrowserActResult{
		Kind:           "save_pdf",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       actionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       finalURL,
		Title:          result.Title,
		Path:           artifactDisplayPath,
		FilesTouched:   browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:          artifactBytes,
		Status:         "saved",
		Force:          pageCtx.Force,
		ReviewDecision: actionResult.ReviewDecision,
		ReviewReady:    actionResult.ReviewReady,
		TabIndex:       pageCtx.TabIndex,
		Note:           actionResult.Note,
	}, nil
}

func browserActExecuteTraceStop(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	traceBackend, ok := pageCtx.RoutedBackend.(BrowserTraceActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "trace_stop")
	}
	waitMs := browserActTraceWaitMs(firstInt(params, "wait_ms"), pageCtx.Target)
	pathValue := firstString(params, "path", "output", "output_path")
	if strings.TrimSpace(pathValue) == "" {
		pathValue = defaultBrowserTraceRelPath()
	}
	outputPath, err := browserActResolveArtifactOutputPath(pageCtx.Options.Root, pathValue)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	request := BrowserTraceRequest{
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		Action:            "stop",
	}
	var result BrowserTraceResult
	artifactBytes, _, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"trace_stop",
		outputPath.Resolved,
		func(stagePath string) (string, bool, error) {
			stageRequest := request
			stageRequest.OutputPath = stagePath
			var traceErr error
			result, traceErr = traceBackend.Trace(pageCtx.CallCtx, stageRequest)
			if traceErr != nil {
				return "", false, traceErr
			}
			return result.Path, true, nil
		},
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	artifactDisplayPath := outputPath.Display
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             strings.TrimSpace(result.Title),
			Source:            "browser_act_trace_stop",
			SetCurrent:        pageCtx.TabIndex > 0,
			Note:              strings.TrimSpace(result.Note),
		},
	)
	return BrowserActResult{
		Kind:          "trace_stop",
		Backend:       result.Backend,
		BrowserApp:    firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:        pageCtx.Target.Value,
		TargetID:      actionResult.TargetID,
		Profile:       pageCtx.RuntimeInfo.Profile,
		RuntimeTarget: pageCtx.RuntimeInfo.Target,
		FinalURL:      strings.TrimSpace(result.FinalURL),
		Title:         strings.TrimSpace(result.Title),
		Path:          artifactDisplayPath,
		FilesTouched:  browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:         artifactBytes,
		Status:        firstNonEmpty(strings.TrimSpace(result.Status), "saved"),
		TabIndex:      pageCtx.TabIndex,
		Note:          actionResult.Note,
	}, nil
}

func browserActExecuteSaveHTML(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	fileBackend, ok := pageCtx.RoutedBackend.(BrowserArtifactActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "save_html")
	}
	waitMs := browserActInteractivePageActionWaitMs(pageCtx.RequestURL, firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs, pageCtx.Target)
	pathValue := firstString(params, "path", "output", "output_path")
	if strings.TrimSpace(pathValue) == "" {
		pathValue = defaultBrowserHTMLRelPath()
	}
	outputPath, err := browserActResolveArtifactOutputPath(pageCtx.Options.Root, pathValue)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	reviewNote := agentxbrowserruntime.SharedSessionBrowserActionReviewReason("save_html", pageCtx.Force)
	if agentxbrowserruntime.SharedSessionBrowserActionRequiresReview("save_html") && !pageCtx.Force {
		return browserActReviewBlockedResultWithPath(
			"save_html",
			pageCtx.RuntimeInfo,
			pageCtx.BrowserApp,
			pageCtx.Target,
			pageCtx.Force,
			reviewNote,
			outputPath.Display,
		), nil
	}
	request := BrowserSaveHTMLRequest{
		URL:               pageCtx.RequestURL,
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("save_html", pageCtx.Force),
		ReviewReady:       pageCtx.Force,
		Note:              reviewNote,
	}
	var result BrowserSaveHTMLResult
	artifactBytes, _, err := publishBrowserArtifactOutput(
		pageCtx.CallCtx,
		pageCtx.Options.PublishArtifact,
		pageCtx.RoutedBackend,
		pageCtx.Options.Root,
		"save_html",
		outputPath.Resolved,
		func(stagePath string) (string, bool, error) {
			stageRequest := request
			stageRequest.OutputPath = stagePath
			var saveErr error
			result, saveErr = fileBackend.SaveHTML(pageCtx.CallCtx, stageRequest)
			if saveErr != nil {
				return "", false, saveErr
			}
			return result.Path, true, nil
		},
	)
	if err != nil {
		return BrowserActResult{}, fmt.Errorf("browser_act: %w", err)
	}
	artifactDisplayPath := outputPath.Display
	finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), pageCtx.RequestURL)
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          finalURL,
			Title:             result.Title,
			Source:            "browser_act_save_html",
			SetCurrent:        pageCtx.TabIndex > 0,
			ReviewDecision:    agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("save_html", pageCtx.Force),
			ReviewReady:       pageCtx.Force,
			Note:              firstNonEmpty(strings.TrimSpace(result.Note), reviewNote),
		},
	)
	return BrowserActResult{
		Kind:           "save_html",
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       actionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       finalURL,
		Title:          result.Title,
		Path:           artifactDisplayPath,
		FilesTouched:   browserArtifactTouchedPaths(artifactDisplayPath),
		Bytes:          artifactBytes,
		Status:         "saved",
		Force:          pageCtx.Force,
		ReviewDecision: actionResult.ReviewDecision,
		ReviewReady:    actionResult.ReviewReady,
		TabIndex:       pageCtx.TabIndex,
		Note:           actionResult.Note,
	}, nil
}
