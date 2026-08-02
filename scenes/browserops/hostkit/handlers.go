package hostkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	DefaultSource              = "agentx_browser_ops_livekit"
	defaultOpenWaitMS          = 800
	defaultSnapshotMaxChars    = 6000
	defaultSnapshotMaxElements = 80
)

type Config struct {
	Executor              agentxtools.Executor
	Source                string
	RuntimeActTool        string
	RuntimeScreenshotTool string
	TaskObserver          TaskObserver
	DefaultOpenWaitMS     int
	SnapshotMaxChars      int
	SnapshotMaxElements   int
}

func DefaultConfig() Config {
	return Config{
		Source:                DefaultSource,
		RuntimeActTool:        RuntimeToolBrowserAct,
		RuntimeScreenshotTool: RuntimeToolBrowserScreenshot,
		DefaultOpenWaitMS:     defaultOpenWaitMS,
		SnapshotMaxChars:      defaultSnapshotMaxChars,
		SnapshotMaxElements:   defaultSnapshotMaxElements,
	}
}

type Kit struct {
	cfg Config
}

func New(cfg Config) *Kit {
	defaults := DefaultConfig()
	if strings.TrimSpace(cfg.Source) == "" {
		cfg.Source = defaults.Source
	}
	if strings.TrimSpace(cfg.RuntimeActTool) == "" {
		cfg.RuntimeActTool = defaults.RuntimeActTool
	}
	if strings.TrimSpace(cfg.RuntimeScreenshotTool) == "" {
		cfg.RuntimeScreenshotTool = defaults.RuntimeScreenshotTool
	}
	if cfg.DefaultOpenWaitMS <= 0 {
		cfg.DefaultOpenWaitMS = defaults.DefaultOpenWaitMS
	}
	if cfg.SnapshotMaxChars <= 0 {
		cfg.SnapshotMaxChars = defaults.SnapshotMaxChars
	}
	if cfg.SnapshotMaxElements <= 0 {
		cfg.SnapshotMaxElements = defaults.SnapshotMaxElements
	}
	return &Kit{cfg: cfg}
}

func BuildStandardToolHandlers(cfg Config) ToolHandlers {
	kit := New(cfg)
	return ToolHandlers{
		OpenTarget: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.OpenTarget(ctx, params)
		},
		FillFields: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.FillFields(ctx, params)
		},
		CapturePageSnapshot: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.CapturePageSnapshot(ctx, params)
		},
		CaptureSubmissionEvidence: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.CaptureSubmissionEvidence(ctx, params)
		},
		DownloadFile: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.DownloadFile(ctx, params)
		},
	}
}

func (k *Kit) OpenTarget(ctx context.Context, params map[string]any) (any, error) {
	args := map[string]any{
		"kind":    "open",
		"wait_ms": k.cfg.DefaultOpenWaitMS,
	}
	copyPresent(args, params, browserCommonRuntimeArgKeys...)
	if url := firstString(params, "url", "target_url"); url != "" {
		args["url"] = url
	}
	return k.executeRuntime(ctx, ToolBrowserOpenTarget, k.cfg.RuntimeActTool, args)
}

func (k *Kit) FillFields(ctx context.Context, params map[string]any) (any, error) {
	args := map[string]any{
		"kind":   "fill",
		"target": "current",
	}
	copyPresent(args, params, browserCommonRuntimeArgKeys...)
	copyPresent(args, params, "fields", "submit")
	return k.executeRuntime(ctx, ToolBrowserFillFields, k.cfg.RuntimeActTool, args)
}

func (k *Kit) CapturePageSnapshot(ctx context.Context, params map[string]any) (any, error) {
	args := map[string]any{
		"kind":         "snapshot",
		"target":       "current",
		"format":       "aria",
		"mode":         "efficient",
		"interactive":  true,
		"compact":      true,
		"max_chars":    k.cfg.SnapshotMaxChars,
		"max_elements": k.cfg.SnapshotMaxElements,
	}
	copyPresent(args, params, browserCommonRuntimeArgKeys...)
	copyPresent(args, params, "max_chars", "max_elements")
	return k.executeRuntime(ctx, ToolBrowserCapturePageSnapshot, k.cfg.RuntimeActTool, args)
}

func (k *Kit) CaptureSubmissionEvidence(ctx context.Context, params map[string]any) (any, error) {
	runtimeTool := strings.TrimSpace(k.cfg.RuntimeScreenshotTool)
	args := map[string]any{
		"target":    "current",
		"full_page": true,
	}
	if runtimeTool == RuntimeToolBrowserAct {
		args["kind"] = "screenshot"
	}
	copyPresent(args, params, browserScreenshotRuntimeArgKeys...)
	return k.executeRuntime(ctx, ToolBrowserCaptureSubmissionEvidence, runtimeTool, args)
}

func (k *Kit) DownloadFile(ctx context.Context, params map[string]any) (any, error) {
	args := map[string]any{
		"kind":   normalizeDownloadMode(firstString(params, "mode", "kind", "action"), firstString(params, "url", "href", "download_url")),
		"target": "current",
	}
	copyPresent(args, params, browserCommonRuntimeArgKeys...)
	copyPresent(args, params, browserDownloadRuntimeArgKeys...)
	if url := firstString(params, "url", "href", "download_url"); url != "" {
		args["url"] = url
	}
	return k.executeRuntime(ctx, ToolBrowserDownloadFile, k.cfg.RuntimeActTool, args)
}

func (k *Kit) executeRuntime(ctx context.Context, semanticTool string, runtimeTool string, args map[string]any) (any, error) {
	runtimeTool = strings.TrimSpace(runtimeTool)
	if runtimeTool == "" {
		runtimeTool = RuntimeToolBrowserAct
	}
	if k.cfg.Executor == nil {
		payload := k.unsupportedPayload(semanticTool, runtimeTool, args)
		k.observeTask(ctx, TaskObservation{
			SemanticTool:  semanticTool,
			RuntimeTool:   runtimeTool,
			RuntimeKind:   stringArg(args["kind"]),
			Status:        "unsupported",
			AdapterStatus: "unsupported",
			FailureCode:   "browser_runtime_executor_not_configured",
			Source:        k.cfg.Source,
		})
		return payload, nil
	}
	blob, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%s: encode runtime arguments: %w", semanticTool, err)
	}
	out, err := k.cfg.Executor.Execute(ctx, llm.FunctionCall{
		Name:      runtimeTool,
		Arguments: string(blob),
	})
	k.observeTask(ctx, browserTaskObservationFromRuntimeOutput(semanticTool, runtimeTool, stringArg(args["kind"]), k.cfg.Source, out, err))
	return out, err
}

func (k *Kit) unsupportedPayload(semanticTool string, runtimeTool string, args map[string]any) map[string]any {
	payload := map[string]any{
		"tool":           semanticTool,
		"source":         k.cfg.Source,
		"status":         "unsupported",
		"adapter_status": "unsupported",
		"failure_code":   "browser_runtime_executor_not_configured",
		"missing":        "livekit.executor",
		"runtime_tool":   runtimeTool,
	}
	switch semanticTool {
	case ToolBrowserFillFields:
		payload["field_count"] = 0
		payload["submitted"] = boolArg(args["submit"])
	case ToolBrowserCapturePageSnapshot:
		payload["snapshot"] = ""
	case ToolBrowserCaptureSubmissionEvidence:
		payload["path"] = ""
	case ToolBrowserDownloadFile:
		payload["path"] = ""
		payload["bytes"] = 0
		payload["content_type"] = ""
	}
	return payload
}

func (k *Kit) observeTask(ctx context.Context, observation TaskObservation) {
	if k.cfg.TaskObserver == nil {
		return
	}
	if strings.TrimSpace(observation.Source) == "" {
		observation.Source = k.cfg.Source
	}
	k.cfg.TaskObserver.ObserveBrowserTask(ctx, observation)
}

func browserTaskObservationFromRuntimeOutput(semanticTool, runtimeTool, runtimeKind, source, output string, err error) TaskObservation {
	observation := TaskObservation{
		SemanticTool: semanticTool,
		RuntimeTool:  runtimeTool,
		RuntimeKind:  runtimeKind,
		Status:       "observed",
		Source:       source,
	}
	if err != nil {
		observation.Status = "failed"
		observation.FailureCode = err.Error()
		return observation
	}
	payload := map[string]any{}
	if decodeErr := json.Unmarshal([]byte(output), &payload); decodeErr != nil {
		return observation
	}
	adapterStatus := strings.ToLower(strings.TrimSpace(firstNonEmptyString(stringArg(payload["status"]), stringArg(payload["adapter_status"]))))
	observation.AdapterStatus = adapterStatus
	failureCode := firstNonEmptyString(stringArg(payload["failure_code"]), stringArg(payload["error"]), stringArg(payload["reason"]))
	observation.FailureCode = failureCode
	switch adapterStatus {
	case "failed", "degraded", "unsupported", "error":
		observation.Status = adapterStatus
	default:
		if failureCode != "" {
			observation.Status = "failed"
		}
	}
	return observation
}

var browserCommonRuntimeArgKeys = []string{
	"target",
	"tab_index",
	"index",
	"browser",
	"browser_app",
	"profile",
	"runtime_target",
	"wait_ms",
	"post_wait_ms",
	"force",
}

var browserScreenshotRuntimeArgKeys = []string{
	"target",
	"tab_index",
	"index",
	"browser",
	"browser_app",
	"profile",
	"runtime_target",
	"wait_ms",
	"force",
	"remember_target",
	"remember",
	"path",
	"output",
	"output_path",
	"full_page",
}

var browserDownloadRuntimeArgKeys = []string{
	"path",
	"output",
	"output_path",
	"wait_ms",
	"timeout_ms",
	"force",
	"allow_recent_download_reuse",
}

func normalizeDownloadMode(mode string, url string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "wait", "wait-download", "wait_download":
		return "wait_download"
	case "direct", "download":
		return "download"
	case "":
		if strings.TrimSpace(url) == "" {
			return "wait_download"
		}
		return "download"
	default:
		return strings.ToLower(strings.TrimSpace(mode))
	}
}

func copyPresent(dst map[string]any, src map[string]any, keys ...string) {
	if dst == nil || len(src) == 0 {
		return
	}
	for _, key := range keys {
		if value, ok := src[key]; ok && value != nil {
			dst[key] = value
		}
	}
}

func firstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringArg(params[key])); value != "" {
			return value
		}
	}
	return ""
}

func stringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func boolArg(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}
