package tools

import "strings"

// BrowserCapabilitiesForActKinds converts the configured browser act kind list
// into the capability surface exposed by a routed browser backend.
func BrowserCapabilitiesForActKinds(kinds []string) BrowserCapabilities {
	var caps BrowserCapabilities
	for _, item := range kinds {
		switch strings.ToLower(strings.TrimSpace(item)) {
		case "open":
			caps.Open = true
		case "navigate":
			caps.Navigate = true
		case "extract":
			caps.Extract = true
		case "snapshot":
			caps.Snapshot = true
		case "screenshot":
			caps.Screenshot = true
		case "console":
			caps.Console = true
		case "requests":
			caps.Requests = true
		case "response_body":
			caps.ResponseBody = true
		case "errors":
			caps.Errors = true
		case "cookies":
			caps.Cookies = true
		case "cookies_set":
			caps.CookiesSet = true
		case "cookies_clear":
			caps.CookiesClear = true
		case "storage":
			caps.Storage = true
		case "storage_set":
			caps.StorageSet = true
		case "storage_clear":
			caps.StorageClear = true
		case "offline":
			caps.Offline = true
		case "headers":
			caps.Headers = true
		case "credentials":
			caps.Credentials = true
		case "geolocation":
			caps.Geolocation = true
		case "media":
			caps.Media = true
		case "timezone":
			caps.Timezone = true
		case "locale":
			caps.Locale = true
		case "device":
			caps.Device = true
		case "highlight":
			caps.Highlight = true
		case "trace_start":
			caps.TraceStart = true
		case "trace_stop":
			caps.TraceStop = true
		case "download":
			caps.Download = true
		case "wait_download":
			caps.WaitDownload = true
		case "save_pdf":
			caps.SavePDF = true
		case "save_html":
			caps.SaveHTML = true
		case "dialog":
			caps.Dialog = true
		case "upload":
			caps.Upload = true
		case "press":
			caps.Press = true
		case "hover":
			caps.Hover = true
		case "drag":
			caps.Drag = true
		case "select":
			caps.Select = true
		case "fill":
			caps.Fill = true
		case "resize":
			caps.Resize = true
		case "click":
			caps.Click = true
		case "type":
			caps.TypeText = true
		case "evaluate":
			caps.Evaluate = true
		case "wait":
			caps.Wait = true
		case "list_tabs", "focus_tab", "close_tab":
			caps.Tabs = true
		}
	}
	return caps
}
