package tools

import (
	"reflect"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func browserUnifiedInventoryDefinition() types.Tool {
	return browserDefinition(browserUnifiedInventoryRuntimeActions(), browserUnifiedInventoryActKinds())
}

func browserUnifiedInventoryRuntimeActions() []string {
	return browserUnifiedInventoryCapabilities().SupportedRuntimeActions()
}

func browserUnifiedInventoryActKinds() []string {
	return browserUnifiedInventoryCapabilities().SupportedActKinds()
}

func browserUnifiedInventoryCapabilities() BrowserCapabilities {
	return BrowserCapabilities{
		RuntimeWorkbench:    true,
		RuntimePrepare:      true,
		RuntimeCoordinate:   true,
		RuntimeStart:        true,
		RuntimeRestart:      true,
		RuntimeStop:         true,
		RuntimeCreate:       true,
		RuntimeDelete:       true,
		RuntimeSelect:       true,
		RuntimeClear:        true,
		RuntimeClearSession: true,
		RuntimeSyncSession:  true,
		RuntimeSelectTarget: true,
		RuntimeClearTarget:  true,
		RuntimeList:         true,
		RuntimeSessions:     true,
		Open:                true,
		Navigate:            true,
		Tabs:                true,
		Extract:             true,
		Snapshot:            true,
		Screenshot:          true,
		Console:             true,
		Requests:            true,
		ResponseBody:        true,
		Errors:              true,
		Cookies:             true,
		CookiesSet:          true,
		CookiesClear:        true,
		Storage:             true,
		StorageSet:          true,
		StorageClear:        true,
		Offline:             true,
		Headers:             true,
		Credentials:         true,
		Geolocation:         true,
		Media:               true,
		Timezone:            true,
		Locale:              true,
		Device:              true,
		Highlight:           true,
		TraceStart:          true,
		TraceStop:           true,
		Download:            true,
		WaitDownload:        true,
		SavePDF:             true,
		SaveHTML:            true,
		Dialog:              true,
		Upload:              true,
		Press:               true,
		Hover:               true,
		Drag:                true,
		Select:              true,
		Fill:                true,
		Resize:              true,
		Click:               true,
		TypeText:            true,
		Evaluate:            true,
		Wait:                true,
	}
}

func browserUnifiedParametersSchema(properties map[string]any) map[string]any {
	out := make(map[string]any, len(properties))
	for name, raw := range properties {
		out[name] = browserSchemaWithInputDescriptions(raw, name)
	}
	return closedInputSchema(out, nil)
}

func browserDescribedInputSchema(parameters map[string]any) map[string]any {
	described, ok := browserSchemaWithInputDescriptions(parameters, "").(map[string]any)
	if !ok {
		return parameters
	}
	return described
}

func browserUnifiedOutputSchema() map[string]any {
	return closedOutputSchema(browserUnifiedOutputSchemaProperties(), []string{"status"})
}

func browserUnifiedOutputSchemaProperties() map[string]any {
	props := map[string]any{}
	for _, sample := range []any{
		browserRuntimePayload{},
		browserActToolPayload{},
		browserTabsToolPayload{},
		browserPageActionReviewBlockedPayload{},
		browserOpenToolPayload{},
		browserNavigateToolPayload{},
		browserExtractToolPayload{},
		browserScreenshotToolPayload{},
		browserClickToolPayload{},
		browserTypeToolPayload{},
		browserEvalToolPayload{},
	} {
		browserCollectSchemaProperties(props, reflect.TypeOf(sample))
	}
	return props
}

func browserCollectSchemaProperties(props map[string]any, typ reflect.Type) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			if field.Anonymous {
				browserCollectSchemaProperties(props, field.Type)
			}
			continue
		}
		if _, ok := props[name]; ok {
			continue
		}
		props[name] = browserSchemaForReflectType(name, field.Type)
	}
}

func browserSchemaForReflectType(name string, typ reflect.Type) map[string]any {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	description := browserUnifiedOutputPropertyDescription(name)
	switch typ.Kind() {
	case reflect.String:
		return stringSchema(description)
	case reflect.Bool:
		return boolSchema(description)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return integerOutputSchema(description)
	case reflect.Float32, reflect.Float64:
		return numberSchema(description)
	case reflect.Slice, reflect.Array:
		return browserArraySchemaForReflectType(description, typ.Elem())
	case reflect.Map, reflect.Struct:
		return looseObjectSchema(description)
	default:
		return looseObjectSchema(description)
	}
}

func browserArraySchemaForReflectType(description string, elem reflect.Type) map[string]any {
	for elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}
	switch elem.Kind() {
	case reflect.String:
		return stringArraySchema(description)
	case reflect.Bool:
		return map[string]any{
			"type":        "array",
			"description": description,
			"items":       map[string]any{"type": "boolean"},
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{
			"type":        "array",
			"description": description,
			"items":       map[string]any{"type": "integer"},
		}
	case reflect.Float32, reflect.Float64:
		return map[string]any{
			"type":        "array",
			"description": description,
			"items":       map[string]any{"type": "number"},
		}
	default:
		return looseObjectArraySchema(description)
	}
}

func browserSchemaWithDescription(raw any, description string) any {
	schema, ok := raw.(map[string]any)
	if !ok {
		return raw
	}
	out := make(map[string]any, len(schema)+1)
	for key, value := range schema {
		out[key] = value
	}
	if strings.TrimSpace(browserSchemaString(out["description"])) == "" {
		out["description"] = description
	}
	return out
}

func browserSchemaWithInputDescriptions(raw any, name string) any {
	switch typed := raw.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed)+1)
		for key, value := range typed {
			switch key {
			case "properties":
				if props, ok := value.(map[string]any); ok {
					nested := make(map[string]any, len(props))
					for nestedName, nestedValue := range props {
						nested[nestedName] = browserSchemaWithInputDescriptions(nestedValue, nestedName)
					}
					out[key] = nested
					continue
				}
			case "items", "additionalProperties":
				out[key] = browserSchemaWithInputDescriptions(value, name)
				continue
			case "anyOf", "oneOf", "allOf":
				out[key] = browserSchemaWithInputDescriptions(value, name)
				continue
			}
			out[key] = browserSchemaWithInputDescriptions(value, "")
		}
		if name != "" && browserSchemaLooksDescribable(out) && strings.TrimSpace(browserSchemaString(out["description"])) == "" {
			out["description"] = browserUnifiedInputPropertyDescription(name)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if described, ok := browserSchemaWithInputDescriptions(item, name).(map[string]any); ok {
				out = append(out, described)
			}
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, browserSchemaWithInputDescriptions(item, name))
		}
		return out
	default:
		return raw
	}
}

func browserSchemaLooksDescribable(schema map[string]any) bool {
	for _, key := range []string{"type", "properties", "items", "anyOf", "oneOf", "allOf", "additionalProperties", "enum"} {
		if _, ok := schema[key]; ok {
			return true
		}
	}
	return false
}

func browserSchemaString(raw any) string {
	value, _ := raw.(string)
	return value
}

func integerOutputSchema(description string) map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": description,
	}
}

func browserUnifiedInputPropertyDescription(name string) string {
	descriptions := map[string]string{
		"action":            "Unified browser action to run.",
		"operation":         "Compatibility alias for action.",
		"kind":              "Raw browser_act kind when using action=act.",
		"url":               "URL used by open, navigate, extract, snapshot, screenshot, or download actions.",
		"target":            "Browser target selector such as current, tab:N, or target:<id>.",
		"tab_index":         "One-based browser tab index.",
		"index":             "Compatibility alias for tab_index.",
		"selector":          "CSS selector or selector-like target for page actions.",
		"ref":               "Snapshot element reference for targeted page actions.",
		"element_ref":       "Compatibility alias for ref.",
		"input_ref":         "Compatibility alias for ref on input fields.",
		"text":              "Text to type, label hint to click, or text input for page actions.",
		"value":             "Value to type, select, store, or submit for page actions.",
		"values":            "Values to select, store, or submit for multi-value page actions.",
		"profile":           "Browser runtime profile.",
		"runtime_target":    "Browser runtime substrate target.",
		"browser_app":       "Host browser application name when using a host runtime.",
		"browser":           "Compatibility alias for browser_app.",
		"wait_ms":           "Maximum wait time for the browser action.",
		"post_wait_ms":      "Additional wait time after a state-changing page action.",
		"max_chars":         "Maximum text characters returned by extract/evaluate-style actions.",
		"max_elements":      "Maximum elements returned by snapshot-style actions.",
		"remember_target":   "Persist the selected target for later actions in this session.",
		"remember":          "Persist the selected runtime profile or target for later actions in this session.",
		"force":             "Bypass conservative review gates when the caller explicitly confirms the action.",
		"full_page":         "Capture the full page for screenshot or PDF-style actions.",
		"path":              "Workspace output path, input path, cookie path, or artifact path depending on action.",
		"paths":             "Workspace paths used by upload, download, or artifact actions.",
		"fields":            "Field descriptors for form-fill actions.",
		"cookies":           "Cookie descriptors for cookie set/clear/read actions.",
		"entries":           "Storage entries for storage set/clear/read actions.",
		"headers":           "Headers to set or inspect for browser request/header actions.",
		"script":            "JavaScript source for evaluate actions.",
		"javascript":        "Compatibility alias for script.",
		"js":                "Compatibility alias for script.",
		"coordination_goal": "Runtime coordination goal for workbench/control actions.",
		"files":             "Workspace file paths used by upload actions.",
		"file":              "Single workspace file path used by upload actions.",
		"output":            "Compatibility alias for output_path.",
		"output_path":       "Workspace output path for screenshots, PDFs, HTML captures, downloads, or traces.",
		"landscape":         "Print orientation flag for PDF capture actions.",
		"print_background":  "Whether PDF capture should include page background graphics.",
		"dialog_action":     "Dialog decision for alert, confirm, or prompt handling.",
		"accept":            "Accept a pending browser dialog when true.",
		"dismiss":           "Dismiss a pending browser dialog when true.",
		"prompt":            "Prompt text for dialog handling or compatibility routing.",
		"prompt_text":       "Text supplied to a prompt dialog.",
		"frame":             "Frame selector or frame hint for targeted page actions.",
		"element":           "Backend-neutral semantic hint for the target element.",
		"label":             "Visible label hint for the target element.",
		"format":            "Snapshot output format.",
		"snapshot_format":   "Compatibility alias for snapshot format.",
		"mode":              "Action mode or compatibility strategy for the selected browser operation.",
		"refs":              "Snapshot reference style to request when supported.",
		"interactive":       "Return or prioritize interactive elements in snapshot-style actions.",
		"compact":           "Request compact snapshot or extraction output when supported.",
		"efficient":         "Request efficient snapshot or extraction output when supported.",
		"depth":             "Maximum traversal depth for snapshot-style actions.",
		"level":             "Console, error, or trace level filter.",
		"filter":            "Text filter applied to debug streams such as console or requests.",
		"request_url":       "Request URL filter for network inspection actions.",
		"response_url":      "Response URL filter for response-body inspection actions.",
		"name":              "Cookie name, storage entry name, or debug stream filter name depending on action.",
		"key":               "Storage entry key.",
		"domain":            "Cookie domain scope.",
		"same_site":         "Cookie SameSite policy.",
		"http_only":         "Cookie HttpOnly flag.",
		"secure":            "Cookie Secure flag.",
		"expires":           "Cookie expiration as a Unix timestamp when supported by the backend.",
		"width":             "Viewport or device width in pixels.",
		"height":            "Viewport or device height in pixels.",
		"storage_kind":      "Browser storage namespace, such as local or session storage.",
		"enabled":           "Enable or disable the selected browser emulation or override.",
		"clear":             "Clear the selected browser state when true.",
		"headers_json":      "JSON object string containing headers for header override actions.",
		"username":          "HTTP authentication username for credentials actions.",
		"password":          "HTTP authentication password for credentials actions.",
		"origin":            "Origin URL for scoped browser permission or geolocation overrides.",
		"latitude":          "Latitude for geolocation emulation.",
		"longitude":         "Longitude for geolocation emulation.",
		"accuracy":          "Geolocation accuracy radius in meters.",
		"media":             "Preferred color-scheme media emulation value.",
		"timezone":          "IANA timezone identifier for timezone emulation.",
		"locale":            "Browser locale identifier for locale emulation.",
		"device":            "Device profile name for device emulation.",
		"delay_ms":          "Delay in milliseconds for wait or press actions.",
		"start_ref":         "Snapshot element reference for the drag start target.",
		"end_ref":           "Snapshot element reference for the drag end target.",
		"start_selector":    "CSS selector for the drag start target.",
		"end_selector":      "CSS selector for the drag end target.",
		"start_element":     "Semantic hint for the drag start element.",
		"end_element":       "Semantic hint for the drag end element.",
		"start_label":       "Visible label hint for the drag start element.",
		"end_label":         "Visible label hint for the drag end element.",
		"from":              "Drag source target, either as a string hint or an object with selector/ref/element/label.",
		"to":                "Drag destination target, either as a string hint or an object with selector/ref/element/label.",
		"submit":            "Submit the form or input after typing/filling when supported.",
	}
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return "Browser parameter `" + name + "` accepted by the unified browser entrypoint."
}

func browserUnifiedOutputPropertyDescription(name string) string {
	descriptions := map[string]string{
		"status":            "Execution status for the delegated browser runtime or page action.",
		"kind":              "Browser action kind for page/action results.",
		"action":            "Runtime or page action that produced this result.",
		"backend":           "Browser backend that handled the request.",
		"browser_app":       "Host browser application used by the request when applicable.",
		"profile":           "Browser runtime profile associated with the result.",
		"runtime_target":    "Browser runtime substrate associated with the result.",
		"url":               "Requested URL.",
		"final_url":         "Final page URL after navigation or extraction.",
		"title":             "Page title observed by the browser backend.",
		"content":           "Extracted page content or action-specific text payload.",
		"content_type":      "Content type for extracted or downloaded content.",
		"snapshot":          "Snapshot text for element/reference grounding.",
		"snapshot_refs":     "Reference format used by the snapshot.",
		"elements":          "Structured snapshot element summaries.",
		"path":              "Workspace artifact path produced or consumed by the action.",
		"files_touched":     "Workspace files touched by the action.",
		"artifacts":         "Artifact descriptors produced by the action.",
		"tabs":              "Browser tabs returned by tabs/session actions.",
		"target":            "Target selector used or selected by the action.",
		"target_id":         "Stable route-scoped target identifier.",
		"tab_index":         "One-based browser tab index.",
		"summary":           "Top-level browser result summary for model-facing follow-up.",
		"diagnostics":       "Top-level browser diagnostics summary.",
		"explanation":       "Top-level browser explanation summary.",
		"display":           "Display-oriented browser result projection.",
		"surface":           "Capability or review surface projection.",
		"view":              "Unified top-level browser view projection.",
		"review":            "Review state for popup, redirect, or policy-confirmation flows.",
		"workbench":         "Runtime workbench payload for inspection, coordination, and next steps.",
		"runtime_actions":   "Runtime actions exposed by the selected browser route.",
		"browser_tools":     "Browser tools exposed by the selected route.",
		"browser_act_kinds": "Browser action kinds exposed by the selected route.",
		"capabilities":      "Browser capability matrix for the selected route.",
		"note":              "Human-readable non-fatal note or fallback explanation.",
	}
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return "Browser result field `" + name + "` from delegated runtime or page action output."
}
