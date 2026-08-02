package hostkit

const (
	ToolBrowserOpenTarget                = "browser_open_target"
	ToolBrowserFillFields                = "browser_fill_fields"
	ToolBrowserCapturePageSnapshot       = "browser_capture_page_snapshot"
	ToolBrowserCaptureSubmissionEvidence = "browser_capture_submission_evidence"
	ToolBrowserDownloadFile              = "browser_download_file"

	RuntimeToolBrowserAct        = "browser_act"
	RuntimeToolBrowserScreenshot = "browser_screenshot"
)

func ToolNames() []string {
	return []string{
		ToolBrowserOpenTarget,
		ToolBrowserFillFields,
		ToolBrowserCapturePageSnapshot,
		ToolBrowserCaptureSubmissionEvidence,
		ToolBrowserDownloadFile,
	}
}
