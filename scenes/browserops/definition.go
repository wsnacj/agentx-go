package browserops

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID                        = "browser-ops-pack"
	DefaultWorkflow               = "browser_form_submit_v1"
	ActionFailurePayloadWorkflow  = "browser_action_failure_payload_v1"
	VerifyPageStateWorkflow       = "browser_verify_page_state_v1"
	ExtractStructuredDataWorkflow = "browser_extract_structured_data_v1"
	SiteSearchWorkflow            = "browser_site_search_v1"
	DownloadFileWorkflow          = "browser_download_file_v1"
)

func browserFormLikeCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"form": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"submit": map[string]any{"type": "boolean"},
					"fields": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"selector": map[string]any{"type": "string"},
								"value":    map[string]any{"type": "string"},
								"type":     map[string]any{"type": "string"},
							},
							"required": []string{"selector", "value"},
						},
					},
				},
				"required": []string{"fields"},
			},
		},
		"required": []string{"target_url", "form"},
	}
}

func browserActionFailurePayloadCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"payloads": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": true,
				},
			},
			"required_failed_checks": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"require_trace_artifact":     map[string]any{"type": "boolean"},
			"require_snapshot_evidence":  map[string]any{"type": "boolean"},
			"min_distinct_failed_checks": map[string]any{"type": "integer"},
		},
		"required": []string{"payloads", "required_failed_checks"},
	}
}

func browserVerifyPageStateCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"expectations": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"required_text": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"forbidden_text": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"url_contains":       map[string]any{"type": "string"},
					"require_screenshot": map[string]any{"type": "boolean"},
					"require_final_url":  map[string]any{"type": "boolean"},
					"min_snapshot_chars": map[string]any{"type": "integer"},
				},
				"additionalProperties": false,
			},
		},
		"required": []string{"target_url", "expectations"},
	}
}

func browserExtractStructuredDataCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"extraction": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fields": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":          map[string]any{"type": "string"},
								"type":          map[string]any{"type": "string"},
								"description":   map[string]any{"type": "string"},
								"required":      map[string]any{"type": "boolean"},
								"expected_text": map[string]any{"type": "string"},
							},
							"required":             []string{"name"},
							"additionalProperties": false,
						},
					},
					"url_contains":       map[string]any{"type": "string"},
					"require_screenshot": map[string]any{"type": "boolean"},
					"require_final_url":  map[string]any{"type": "boolean"},
					"min_snapshot_chars": map[string]any{"type": "integer"},
				},
				"required":             []string{"fields"},
				"additionalProperties": false,
			},
		},
		"required": []string{"target_url", "extraction"},
	}
}

func browserSiteSearchCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"search": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
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
							"additionalProperties": true,
						},
					},
					"submit":                map[string]any{"type": "boolean"},
					"expected_results":      browserSiteSearchExpectedResultsSchema(),
					"required_text":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"forbidden_text":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"url_contains":          map[string]any{"type": "string"},
					"require_query_visible": map[string]any{"type": "boolean"},
					"require_screenshot":    map[string]any{"type": "boolean"},
					"require_final_url":     map[string]any{"type": "boolean"},
					"require_search_action": map[string]any{"type": "boolean"},
					"require_submitted":     map[string]any{"type": "boolean"},
					"min_snapshot_chars":    map[string]any{"type": "integer"},
				},
				"required":             []string{"query", "fields", "expected_results"},
				"additionalProperties": false,
			},
		},
		"required": []string{"target_url", "search"},
	}
}

func browserSiteSearchExpectedResultsSchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":            map[string]any{"type": "string"},
				"url_contains":     map[string]any{"type": "string"},
				"snippet_contains": map[string]any{"type": "string"},
			},
			"additionalProperties": false,
		},
	}
}

func browserDownloadFileCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"download": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url":                   map[string]any{"type": "string"},
					"mode":                  map[string]any{"type": "string"},
					"expected_filename":     map[string]any{"type": "string"},
					"expected_content_type": map[string]any{"type": "string"},
					"url_contains":          map[string]any{"type": "string"},
					"min_bytes":             map[string]any{"type": "integer"},
					"wait_ms":               map[string]any{"type": "integer"},
					"require_final_url":     map[string]any{"type": "boolean"},
					"require_screenshot":    map[string]any{"type": "boolean"},
				},
				"required":             []string{"url", "expected_filename"},
				"additionalProperties": false,
			},
		},
		"required": []string{"target_url", "download"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:                 PackID,
			Version:            "0.6.0",
			Domain:             "browser_operations",
			RouteHints:         []string{"浏览器自动化", "浏览器表单", "自动提单", "网页录入", "页面状态验证", "网页结构化抽取", "站内搜索", "文件下载", "browser form", "form submission", "record update", "verify page state", "extract structured data", "site search", "download file"},
			SupportedCaseTypes: []string{"browser.form_submission", "browser.record_update", "browser.action_failure_payload", "browser.verify_page_state", "browser.extract_structured_data", "browser.site_search", "browser.download_file"},
			DefaultWorkflow:    DefaultWorkflow,
			RequiredPlugins:    []string{"browser-runtime"},
			PolicyProfiles:     []string{"browser_readonly", "browser_page_state_review", "browser_extract_readonly_review", "browser_site_search_review", "browser_download_review", "browser_fill_review", "browser_failure_payload_review"},
			ArtifactTypes:      []string{BrowserArtifactTypePageSnapshot, BrowserArtifactTypePageScreenshot, BrowserArtifactTypeActionTrace, BrowserArtifactTypeSubmitReport, BrowserArtifactTypeDownloadedFile},
			Evaluators:         []string{"browser_submit_evidence_gate", "browser_action_failure_payload_gate", "browser_page_state_gate", "browser_structured_data_gate", "browser_site_search_gate", "browser_download_file_gate"},
			EvalSuites:         []string{"browser_submit_success_suite", "browser_action_failure_payload_suite", "browser_page_state_success_suite", "browser_structured_data_success_suite", "browser_site_search_success_suite", "browser_download_file_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    "browser.form_submission",
				Description: "浏览器表单提交任务。提供目标页面 URL 和待填写字段，pack 会驱动浏览器打开页面、填写表单、抓取提交证据并给出验收判断。",
				RouteHints:  []string{"自动提单", "提交表单", "填写表单", "网页录入", "browser form", "submit form"},
				Schema:      browserFormLikeCaseSchema(),
			},
			{
				CaseType:    "browser.record_update",
				Description: "浏览器记录更新任务。当前复用同一条表单填报工作流，因此仍然要求提供目标页面 URL 和待填写字段，适合 CRM/工单/后台录入类页面更新。",
				RouteHints:  []string{"更新记录", "后台录入", "工单录入", "crm update", "record update"},
				Schema:      browserFormLikeCaseSchema(),
			},
			{
				CaseType:    "browser.action_failure_payload",
				Description: "浏览器失败动作 payload 回归任务。验证 actionability、failure_evidence 和 trace-like artifact 是否保留机器可读失败原因。",
				RouteHints:  []string{"browser action failure", "actionability failure", "失败动作证据", "browser regression"},
				Schema:      browserActionFailurePayloadCaseSchema(),
			},
			{
				CaseType:    "browser.verify_page_state",
				Description: "浏览器页面状态验证任务。打开目标页面，抓取 snapshot 和截图证据，并按必含文本、禁含文本和最终 URL 约束做只读验收。",
				RouteHints:  []string{"页面状态验证", "验证页面内容", "page state", "verify page state", "browser page check"},
				Schema:      browserVerifyPageStateCaseSchema(),
			},
			{
				CaseType:    "browser.extract_structured_data",
				Description: "浏览器结构化数据抽取任务。打开目标页面，基于 snapshot 和截图证据抽取指定字段，并校验字段值是否有页面证据支撑。",
				RouteHints:  []string{"网页结构化抽取", "提取页面字段", "extract structured data", "browser extraction", "page data extraction"},
				Schema:      browserExtractStructuredDataCaseSchema(),
			},
			{
				CaseType:    "browser.site_search",
				Description: "浏览器站内搜索任务。打开目标页面，按 case 提供的搜索字段执行 host-approved 搜索，抓取结果页证据，并校验预期结果是否出现。",
				RouteHints:  []string{"站内搜索", "搜索站点", "site search", "browser search", "search results"},
				Schema:      browserSiteSearchCaseSchema(),
			},
			{
				CaseType:    "browser.download_file",
				Description: "浏览器文件下载任务。打开目标页面，通过 host-approved 下载 runtime 获取文件 artifact，并校验下载路径、文件名、content-type、大小和 URL 证据。",
				RouteHints:  []string{"文件下载", "下载附件", "download file", "browser download", "download artifact"},
				Schema:      browserDownloadFileCaseSchema(),
			},
		},
		CaseLibrary: []agentxpack.CaseLibraryCase{
			{
				ID:          "browser_action_failure_payload.core_actionability_regression",
				CaseType:    "browser.action_failure_payload",
				Locale:      "en-US",
				Description: "Pack-owned regression case for machine-readable browser actionability failure payloads.",
				Input: map[string]any{
					"payloads":                   []any{"{{payloads}}"},
					"required_failed_checks":     []any{"visible", "enabled", "editable", "receives_events", "attached"},
					"require_trace_artifact":     true,
					"require_snapshot_evidence":  true,
					"min_distinct_failed_checks": 5,
				},
				InputPlaceholders: []agentxpack.CaseInputPlaceholder{
					{
						Name:        "payloads",
						Path:        "payloads",
						Description: "Array of browser_act action_failed payloads captured from focused browser actionability regression fixtures.",
						Required:    true,
						Example:     []any{},
					},
				},
				ExpectedOutput: map[string]any{
					"review.passed":                               true,
					"failure_payload.valid_payload_count":         5,
					"failure_payload.distinct_failed_check_count": 5,
				},
				ReviewStatus: agentxpack.CaseReviewStatusApproved,
				Tags:         []string{"browser", "actionability", "regression"},
			},
		},
		PromptTemplates: []agentxpack.PromptTemplate{
			{
				Name:        "browser_action_failure_payload_gate_instruction",
				Description: "Pack-owned evaluator instruction for browser actionability failure payload regressions.",
				Locale:      "en-US",
				Template:    "Review the browser failed-action payloads. Pass only when each required failed_check is represented by an action_failed payload with failed actionability, matching failure_evidence.reason_code, snapshot evidence when required, and a trace_like failure artifact when required.",
				SourceAttributions: []agentxpack.SourceAttribution{
					{
						SourceType: agentxpack.SourceAttributionTypePack,
						SourceID:   "browserops.workflow.browser_action_failure_payload_v1.failure_payload_gate",
						Title:      "browserops pack failure payload evaluator instruction",
						Notes:      "Pack-owned prompt pattern; kept in browserops metadata rather than generic runtime or builtin tools.",
					},
				},
				CaseTypes: []string{"browser.action_failure_payload"},
				Tags:      []string{"browser", "actionability", "evaluator"},
			},
			{
				Name:        "browser_page_state_gate_instruction",
				Description: "Pack-owned evaluator instruction for readonly browser page-state verification.",
				Locale:      "en-US",
				Template:    "Review the browser page-state evidence. Pass only when snapshot evidence is present, required_text terms are visible in the snapshot, forbidden_text terms are absent, the final URL matches the target/url_contains constraints when requested, and screenshot evidence is present when required.",
				SourceAttributions: []agentxpack.SourceAttribution{
					{
						SourceType: agentxpack.SourceAttributionTypePack,
						SourceID:   "browserops.workflow.browser_verify_page_state_v1.page_state_gate",
						Title:      "browserops pack page-state evaluator instruction",
						Notes:      "Pack-owned page-state gate; host owns browser runtime, tab/session state, login state, and target-site policy.",
					},
				},
				CaseTypes: []string{"browser.verify_page_state"},
				Tags:      []string{"browser", "page-state", "evaluator"},
			},
			{
				Name:        "browser_structured_data_gate_instruction",
				Description: "Pack-owned evaluator instruction for browser structured data extraction.",
				Locale:      "en-US",
				Template:    "Review the browser structured-data extraction evidence. Extract only the requested fields from the page snapshot, return extracted_data keyed by field name, and pass only when every required field has an extracted value and each expected_text value is visible in the page evidence.",
				SourceAttributions: []agentxpack.SourceAttribution{
					{
						SourceType: agentxpack.SourceAttributionTypePack,
						SourceID:   "browserops.workflow.browser_extract_structured_data_v1.structured_data_gate",
						Title:      "browserops pack structured-data evaluator instruction",
						Notes:      "Pack-owned extraction gate; host owns browser runtime, tab/session state, login state, target-site policy, and private field rules.",
					},
				},
				CaseTypes: []string{"browser.extract_structured_data"},
				Tags:      []string{"browser", "structured-data", "evaluator"},
			},
			{
				Name:        "browser_site_search_gate_instruction",
				Description: "Pack-owned evaluator instruction for browser site search.",
				Locale:      "en-US",
				Template:    "Review the browser site-search evidence. Pass only when the query was supplied, search action evidence is ready when requested, the result page snapshot is present, every expected result is visible in page evidence, forbidden text is absent, and screenshot/final URL evidence satisfies the case constraints.",
				SourceAttributions: []agentxpack.SourceAttribution{
					{
						SourceType: agentxpack.SourceAttributionTypePack,
						SourceID:   "browserops.workflow.browser_site_search_v1.site_search_gate",
						Title:      "browserops pack site-search evaluator instruction",
						Notes:      "Pack-owned site-search gate; host owns browser runtime, login state, selector policy, target-site policy, and source-specific search behavior.",
					},
				},
				CaseTypes: []string{"browser.site_search"},
				Tags:      []string{"browser", "site-search", "evaluator"},
			},
			{
				Name:        "browser_download_file_gate_instruction",
				Description: "Pack-owned evaluator instruction for browser file downloads.",
				Locale:      "en-US",
				Template:    "Review the browser download evidence. Pass only when a downloaded-file artifact path is present, status is downloaded or ok, expected filename/content type/min bytes are satisfied, screenshot/final URL evidence satisfies the case constraints when requested, and no runtime failure reasons are present.",
				SourceAttributions: []agentxpack.SourceAttribution{
					{
						SourceType: agentxpack.SourceAttributionTypePack,
						SourceID:   "browserops.workflow.browser_download_file_v1.download_gate",
						Title:      "browserops pack download-file evaluator instruction",
						Notes:      "Pack-owned download gate; host owns download roots, artifact retention, approval policy, target-site policy, browser profile, and login state.",
					},
				},
				CaseTypes: []string{"browser.download_file"},
				Tags:      []string{"browser", "download", "evaluator"},
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "Browser Form Submission",
				Description:     "打开浏览器页面、填写表单、抓取页面证据并做最终提交验收。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.form_submission", "browser.record_update"},
				RouteHints:      []string{"自动提单", "浏览器提单", "browser form", "submit form", "网页录入"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "open_target",
				DefaultContract: "browser_fill_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "form.field_count", Type: "integer", Required: true},
					{Name: "review.snapshot", Type: "string", Required: true},
					{Name: "review.final_url", Type: "string"},
					{Name: "review.evidence_path", Type: "string", Required: true},
					{Name: "review.evidence_final_url", Type: "string"},
					{Name: "review.trace_path", Type: "string"},
					{Name: "review.failure_reasons", Type: "array"},
					{Name: "review.passed", Type: "boolean", Required: true},
					{Name: "review.score", Type: "number", Required: true},
					{Name: "review.summary", Type: "string", Required: true},
				},
				ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
					{Type: BrowserArtifactTypePageScreenshot},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_submit_evidence_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "open_target",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open browser target",
						Description: "打开业务提交页面，建立后续浏览器上下文。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.target_url", To: "args.url"},
						},
						Config: map[string]any{
							"tool": "browser_open_target",
						},
					},
					{
						ID:          "fill_fields",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Fill form fields",
						Description: "按业务 case 中提供的字段列表填写页面表单，可选自动提交。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.form.fields", To: "args.fields"},
							{From: "case.input.form.submit", To: "args.submit"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.field_count", To: "state.form.field_count"},
							{From: "result.submitted", To: "state.form.submitted"},
						},
						Config: map[string]any{
							"tool": "browser_fill_fields",
						},
					},
					{
						ID:          "capture_snapshot",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture page snapshot",
						Description: "抓取提交后的页面快照，供 evaluator 做语义判断。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.snapshot", To: "state.review.snapshot"},
							{From: "result.final_url", To: "state.review.final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_page_snapshot",
						},
					},
					{
						ID:          "capture_evidence",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture submit evidence",
						Description: "保存最终页面截图，进入 artifact 与 eventlog 证据链。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.review.evidence_path"},
							{From: "result.final_url", To: "state.review.evidence_final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_submission_evidence",
						},
					},
					{
						ID:          "final_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Evaluate submission evidence",
						Description: "结合页面快照、截图路径和填表结果，对浏览器任务完成度做结构化验收。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.review.snapshot", To: "args.input"},
							{From: "state.review.evidence_path", To: "args.context.evidence_path"},
							{From: "state.review.final_url", To: "args.context.final_url"},
							{From: "state.form.field_count", To: "args.context.field_count"},
							{From: "state.form.submitted", To: "args.context.submitted"},
							{From: "case.input.target_url", To: "args.context.target_url"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.review.passed"},
							{From: "result.score", To: "state.review.score"},
							{From: "result.summary", To: "state.review.summary"},
						},
						Config: map[string]any{
							"evaluator":   "browser_submit_evidence_gate",
							"instruction": "Review the captured browser snapshot and screenshot evidence metadata. Confirm whether visual_evidence_ready, snapshot_ready, and screenshot_ready are true before passing; then judge whether the intended browser form interaction appears complete, whether the page likely reflects a successful post-submit state, and summarize the strongest evidence.",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "open_target", To: "fill_fields", On: "success"},
					{From: "fill_fields", To: "capture_snapshot", On: "success"},
					{From: "capture_snapshot", To: "capture_evidence", On: "success"},
					{From: "capture_evidence", To: "final_gate", On: "success"},
				},
			},
			{
				ID:              VerifyPageStateWorkflow,
				Title:           "Browser Page State Verification",
				Description:     "打开目标页面、抓取页面状态证据，并根据只读页面状态期望做结构化验收。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.verify_page_state"},
				RouteHints:      []string{"页面状态验证", "browser page state", "verify page state", "page check"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "open_target",
				DefaultContract: "browser_page_state_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "review.snapshot", Type: "string", Required: true},
					{Name: "review.final_url", Type: "string"},
					{Name: "review.evidence_path", Type: "string", Required: true},
					{Name: "review.evidence_final_url", Type: "string"},
					{Name: "page_state.passed", Type: "boolean", Required: true},
					{Name: "page_state.score", Type: "number", Required: true},
					{Name: "page_state.summary", Type: "string", Required: true},
					{Name: "page_state.failure_reasons", Type: "array"},
					{Name: "page_state.evidence_bundle", Type: "object"},
				},
				ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
					{Type: BrowserArtifactTypePageScreenshot},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_page_state_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "open_target",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open browser target",
						Description: "打开待验证页面，建立只读页面状态检查上下文。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.target_url", To: "args.url"},
						},
						Config: map[string]any{
							"tool": "browser_open_target",
						},
					},
					{
						ID:          "capture_snapshot",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture page snapshot",
						Description: "抓取页面结构化 snapshot，作为页面状态判断的主要文本证据。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.snapshot", To: "state.review.snapshot"},
							{From: "result.final_url", To: "state.review.final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_page_snapshot",
						},
					},
					{
						ID:          "capture_evidence",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture page evidence",
						Description: "保存页面截图，补充 snapshot 的视觉证据与 artifact 引用。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.review.evidence_path"},
							{From: "result.final_url", To: "state.review.evidence_final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_submission_evidence",
						},
					},
					{
						ID:          "page_state_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Evaluate page state evidence",
						Description: "根据页面 snapshot、截图证据和 case expectations 判断页面状态是否满足预期。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.review.snapshot", To: "args.input"},
							{From: "state.review.evidence_path", To: "args.context.evidence_path"},
							{From: "state.review.final_url", To: "args.context.final_url"},
							{From: "case.input.target_url", To: "args.context.target_url"},
							{From: "case.input.expectations.required_text", To: "args.context.required_text"},
							{From: "case.input.expectations.forbidden_text", To: "args.context.forbidden_text"},
							{From: "case.input.expectations.url_contains", To: "args.context.url_contains"},
							{From: "case.input.expectations.require_screenshot", To: "args.context.require_screenshot"},
							{From: "case.input.expectations.require_final_url", To: "args.context.require_final_url"},
							{From: "case.input.expectations.min_snapshot_chars", To: "args.context.min_snapshot_chars"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.page_state.passed"},
							{From: "result.score", To: "state.page_state.score"},
							{From: "result.summary", To: "state.page_state.summary"},
							{From: "result.failure_reasons", To: "state.page_state.failure_reasons"},
							{From: "result.evidence_bundle", To: "state.page_state.evidence_bundle"},
						},
						Config: map[string]any{
							"evaluator":   "browser_page_state_gate",
							"instruction": "Review the captured browser page-state evidence. Pass only when snapshot evidence is ready, required_text terms are present, forbidden_text terms are absent, final_url satisfies the target/url_contains constraints when requested, and screenshot evidence is present when require_screenshot is true.",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "open_target", To: "capture_snapshot", On: "success"},
					{From: "capture_snapshot", To: "capture_evidence", On: "success"},
					{From: "capture_evidence", To: "page_state_gate", On: "success"},
				},
			},
			{
				ID:              ExtractStructuredDataWorkflow,
				Title:           "Browser Structured Data Extraction",
				Description:     "打开目标页面、抓取页面证据，并按 case 中声明的字段抽取结构化数据。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.extract_structured_data"},
				RouteHints:      []string{"网页结构化抽取", "browser structured data", "extract page data", "page data extraction"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "open_target",
				DefaultContract: "browser_extract_readonly_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "review.snapshot", Type: "string", Required: true},
					{Name: "review.final_url", Type: "string"},
					{Name: "review.evidence_path", Type: "string", Required: true},
					{Name: "review.evidence_final_url", Type: "string"},
					{Name: "extract.passed", Type: "boolean", Required: true},
					{Name: "extract.score", Type: "number", Required: true},
					{Name: "extract.summary", Type: "string", Required: true},
					{Name: "extract.data", Type: "object", Required: true},
					{Name: "extract.field_results", Type: "array"},
					{Name: "extract.failure_reasons", Type: "array"},
					{Name: "extract.evidence_bundle", Type: "object"},
				},
				ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
					{Type: BrowserArtifactTypePageScreenshot},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_structured_data_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "open_target",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open browser target",
						Description: "打开待抽取页面，建立只读页面抽取上下文。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.target_url", To: "args.url"},
						},
						Config: map[string]any{
							"tool": "browser_open_target",
						},
					},
					{
						ID:          "capture_snapshot",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture page snapshot",
						Description: "抓取页面结构化 snapshot，作为字段抽取与证据校验的主要来源。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.snapshot", To: "state.review.snapshot"},
							{From: "result.final_url", To: "state.review.final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_page_snapshot",
						},
					},
					{
						ID:          "capture_evidence",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture extraction evidence",
						Description: "保存页面截图，补充结构化抽取的视觉证据与 artifact 引用。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.review.evidence_path"},
							{From: "result.final_url", To: "state.review.evidence_final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_submission_evidence",
						},
					},
					{
						ID:          "structured_data_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Extract and evaluate structured data",
						Description: "按 case extraction.fields 抽取结构化字段，并确认字段值有 snapshot 证据支撑。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.review.snapshot", To: "args.input"},
							{From: "state.review.evidence_path", To: "args.context.evidence_path"},
							{From: "state.review.final_url", To: "args.context.final_url"},
							{From: "case.input.target_url", To: "args.context.target_url"},
							{From: "case.input.extraction.fields", To: "args.context.fields"},
							{From: "case.input.extraction.url_contains", To: "args.context.url_contains"},
							{From: "case.input.extraction.require_screenshot", To: "args.context.require_screenshot"},
							{From: "case.input.extraction.require_final_url", To: "args.context.require_final_url"},
							{From: "case.input.extraction.min_snapshot_chars", To: "args.context.min_snapshot_chars"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.extract.passed"},
							{From: "result.score", To: "state.extract.score"},
							{From: "result.summary", To: "state.extract.summary"},
							{From: "result.extracted_data", To: "state.extract.data"},
							{From: "result.field_results", To: "state.extract.field_results"},
							{From: "result.failure_reasons", To: "state.extract.failure_reasons"},
							{From: "result.evidence_bundle", To: "state.extract.evidence_bundle"},
						},
						Config: map[string]any{
							"evaluator":   "browser_structured_data_gate",
							"instruction": "Extract the requested fields from the browser page snapshot. Return extracted_data keyed by field name, field_results for each requested field, and pass only when every required field has an extracted value and each expected_text value is visible in the page evidence. Keep browser profile, login state, and target-site policy host-owned.",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "open_target", To: "capture_snapshot", On: "success"},
					{From: "capture_snapshot", To: "capture_evidence", On: "success"},
					{From: "capture_evidence", To: "structured_data_gate", On: "success"},
				},
			},
			{
				ID:              SiteSearchWorkflow,
				Title:           "Browser Site Search",
				Description:     "打开目标页面、执行 host-approved 站内搜索、抓取结果页证据并验证期望结果。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.site_search"},
				RouteHints:      []string{"站内搜索", "browser site search", "search results", "site search"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "open_target",
				DefaultContract: "browser_site_search_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "search.field_count", Type: "integer", Required: true},
					{Name: "search.submitted", Type: "boolean"},
					{Name: "review.snapshot", Type: "string", Required: true},
					{Name: "review.final_url", Type: "string"},
					{Name: "review.evidence_path", Type: "string", Required: true},
					{Name: "review.evidence_final_url", Type: "string"},
					{Name: "site_search.passed", Type: "boolean", Required: true},
					{Name: "site_search.score", Type: "number", Required: true},
					{Name: "site_search.summary", Type: "string", Required: true},
					{Name: "site_search.result_evaluations", Type: "array"},
					{Name: "site_search.failure_reasons", Type: "array"},
					{Name: "site_search.evidence_bundle", Type: "object"},
				},
				ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
					{Type: BrowserArtifactTypePageScreenshot},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_site_search_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "open_target",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open browser target",
						Description: "打开待搜索页面，建立 host-approved 站内搜索上下文。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.target_url", To: "args.url"},
						},
						Config: map[string]any{
							"tool": "browser_open_target",
						},
					},
					{
						ID:          "submit_search",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Submit site search",
						Description: "按 case.search.fields 填写搜索字段，并根据 submit 设置触发站内搜索。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.search.fields", To: "args.fields"},
							{From: "case.input.search.submit", To: "args.submit"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.field_count", To: "state.search.field_count"},
							{From: "result.submitted", To: "state.search.submitted"},
						},
						Config: map[string]any{
							"tool": "browser_fill_fields",
						},
					},
					{
						ID:          "capture_snapshot",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture search result snapshot",
						Description: "抓取搜索结果页 snapshot，作为结果验证的主要文本证据。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.snapshot", To: "state.review.snapshot"},
							{From: "result.final_url", To: "state.review.final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_page_snapshot",
						},
					},
					{
						ID:          "capture_evidence",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture search evidence",
						Description: "保存搜索结果页截图，补充 snapshot 的视觉证据与 artifact 引用。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.review.evidence_path"},
							{From: "result.final_url", To: "state.review.evidence_final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_submission_evidence",
						},
					},
					{
						ID:          "site_search_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Evaluate site search evidence",
						Description: "根据搜索结果页 snapshot、截图证据和 case.search.expected_results 判断搜索是否达成预期。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.review.snapshot", To: "args.input"},
							{From: "state.review.evidence_path", To: "args.context.evidence_path"},
							{From: "state.review.final_url", To: "args.context.final_url"},
							{From: "state.search.field_count", To: "args.context.field_count"},
							{From: "state.search.submitted", To: "args.context.submitted"},
							{From: "case.input.target_url", To: "args.context.target_url"},
							{From: "case.input.search.query", To: "args.context.query"},
							{From: "case.input.search.expected_results", To: "args.context.expected_results"},
							{From: "case.input.search.required_text", To: "args.context.required_text"},
							{From: "case.input.search.forbidden_text", To: "args.context.forbidden_text"},
							{From: "case.input.search.url_contains", To: "args.context.url_contains"},
							{From: "case.input.search.require_query_visible", To: "args.context.require_query_visible"},
							{From: "case.input.search.require_screenshot", To: "args.context.require_screenshot"},
							{From: "case.input.search.require_final_url", To: "args.context.require_final_url"},
							{From: "case.input.search.require_search_action", To: "args.context.require_search_action"},
							{From: "case.input.search.require_submitted", To: "args.context.require_submitted"},
							{From: "case.input.search.min_snapshot_chars", To: "args.context.min_snapshot_chars"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.site_search.passed"},
							{From: "result.score", To: "state.site_search.score"},
							{From: "result.summary", To: "state.site_search.summary"},
							{From: "result.result_evaluations", To: "state.site_search.result_evaluations"},
							{From: "result.failure_reasons", To: "state.site_search.failure_reasons"},
							{From: "result.evidence_bundle", To: "state.site_search.evidence_bundle"},
						},
						Config: map[string]any{
							"evaluator":   "browser_site_search_gate",
							"instruction": "Review the captured browser site-search result evidence. Pass only when the query is supplied, search action state is ready when requested, expected_results appear in the snapshot, forbidden text is absent, final_url satisfies the target/url_contains constraints when requested, and screenshot evidence is present when require_screenshot is true. Keep target-site policy, selector policy, login state, and source-specific search behavior host-owned.",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "open_target", To: "submit_search", On: "success"},
					{From: "submit_search", To: "capture_snapshot", On: "success"},
					{From: "capture_snapshot", To: "capture_evidence", On: "success"},
					{From: "capture_evidence", To: "site_search_gate", On: "success"},
				},
			},
			{
				ID:              DownloadFileWorkflow,
				Title:           "Browser File Download",
				Description:     "打开目标页面、执行 host-approved 文件下载、抓取页面证据并验证下载 artifact。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.download_file"},
				RouteHints:      []string{"文件下载", "browser download", "download file", "download artifact"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "open_target",
				DefaultContract: "browser_download_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "review.snapshot", Type: "string", Required: true},
					{Name: "review.final_url", Type: "string"},
					{Name: "review.evidence_path", Type: "string"},
					{Name: "review.evidence_final_url", Type: "string"},
					{Name: "download.path", Type: "string", Required: true},
					{Name: "download.status", Type: "string"},
					{Name: "download.bytes", Type: "integer"},
					{Name: "download.content_type", Type: "string"},
					{Name: "download.final_url", Type: "string"},
					{Name: "download.note", Type: "string"},
					{Name: "download.passed", Type: "boolean", Required: true},
					{Name: "download.score", Type: "number", Required: true},
					{Name: "download.summary", Type: "string", Required: true},
					{Name: "download.failure_reasons", Type: "array"},
					{Name: "download.evidence_bundle", Type: "object"},
				},
				ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
					{Type: BrowserArtifactTypeDownloadedFile},
					{Type: BrowserArtifactTypePageScreenshot},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_download_file_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "open_target",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open browser target",
						Description: "打开包含下载入口的目标页面，建立 host-approved 下载上下文。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.target_url", To: "args.url"},
						},
						Config: map[string]any{
							"tool": "browser_open_target",
						},
					},
					{
						ID:          "capture_snapshot",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture download page snapshot",
						Description: "抓取下载页 snapshot，保留下载入口和页面状态的文本证据。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.snapshot", To: "state.review.snapshot"},
							{From: "result.final_url", To: "state.review.final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_page_snapshot",
						},
					},
					{
						ID:          "download_file",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Download file artifact",
						Description: "通过 host-injected browser runtime 下载文件 artifact。下载根目录、保留策略、审批与站点策略均由 host 控制。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.download.url", To: "args.url"},
							{From: "case.input.download.mode", To: "args.mode"},
							{From: "case.input.download.wait_ms", To: "args.wait_ms"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.download.path"},
							{From: "result.status", To: "state.download.status"},
							{From: "result.bytes", To: "state.download.bytes"},
							{From: "result.content_type", To: "state.download.content_type"},
							{From: "result.final_url", To: "state.download.final_url"},
							{From: "result.note", To: "state.download.note"},
						},
						Config: map[string]any{
							"tool": "browser_download_file",
						},
					},
					{
						ID:          "capture_evidence",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Capture download page evidence",
						Description: "保存下载页截图，补充下载 artifact 之外的页面证据。",
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.path", To: "state.review.evidence_path"},
							{From: "result.final_url", To: "state.review.evidence_final_url"},
						},
						Config: map[string]any{
							"tool": "browser_capture_submission_evidence",
						},
					},
					{
						ID:          "download_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Evaluate download artifact evidence",
						Description: "根据下载 artifact 元数据、最终 URL 和页面截图证据判断下载是否达成预期。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.download.path", To: "args.context.downloaded_path"},
							{From: "state.download.status", To: "args.context.status"},
							{From: "state.download.bytes", To: "args.context.bytes"},
							{From: "state.download.content_type", To: "args.context.content_type"},
							{From: "state.download.final_url", To: "args.context.final_url"},
							{From: "state.review.evidence_path", To: "args.context.evidence_path"},
							{From: "case.input.target_url", To: "args.context.target_url"},
							{From: "case.input.download.url", To: "args.context.download_url"},
							{From: "case.input.download.expected_filename", To: "args.context.expected_filename"},
							{From: "case.input.download.expected_content_type", To: "args.context.expected_content_type"},
							{From: "case.input.download.url_contains", To: "args.context.url_contains"},
							{From: "case.input.download.min_bytes", To: "args.context.min_bytes"},
							{From: "case.input.download.require_final_url", To: "args.context.require_final_url"},
							{From: "case.input.download.require_screenshot", To: "args.context.require_screenshot"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.download.passed"},
							{From: "result.score", To: "state.download.score"},
							{From: "result.summary", To: "state.download.summary"},
							{From: "result.failure_reasons", To: "state.download.failure_reasons"},
							{From: "result.evidence_bundle", To: "state.download.evidence_bundle"},
						},
						Config: map[string]any{
							"evaluator":   "browser_download_file_gate",
							"instruction": "Review the browser download artifact evidence. Pass only when a downloaded file path is present, status is downloaded or ok, expected filename/content type/min bytes are satisfied, final_url satisfies the target/url_contains constraints when requested, screenshot evidence is present when require_screenshot is true, and no runtime failure reasons are present. Keep download root, retention policy, target-site policy, and approvals host-owned.",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "open_target", To: "capture_snapshot", On: "success"},
					{From: "capture_snapshot", To: "download_file", On: "success"},
					{From: "download_file", To: "capture_evidence", On: "success"},
					{From: "capture_evidence", To: "download_gate", On: "success"},
				},
			},
			{
				ID:              ActionFailurePayloadWorkflow,
				Title:           "Browser Action Failure Payload Gate",
				Description:     "验证浏览器失败动作 payload 是否保留 actionability、failure evidence 和 trace-like artifact。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{"browser.action_failure_payload"},
				RouteHints:      []string{"browser action failure", "actionability failure", "browser regression"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "failure_payload_gate",
				DefaultContract: "browser_failure_payload_review",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "review.passed", Type: "boolean", Required: true},
					{Name: "review.summary", Type: "string", Required: true},
					{Name: "failure_payload.valid_payload_count", Type: "integer", Required: true},
					{Name: "failure_payload.distinct_failed_check_count", Type: "integer", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "browser_action_failure_payload_gate"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "failure_payload_gate",
						Kind:        agentxworkflow.NodeEvaluate,
						Title:       "Evaluate browser action failure payloads",
						Description: "检查失败动作 payload 是否保留 status、failed_check、reason_code、snapshot evidence 和 trace-like artifact。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.payloads", To: "args.input.payloads"},
							{From: "case.input.required_failed_checks", To: "args.input.required_failed_checks"},
							{From: "case.input.require_trace_artifact", To: "args.input.require_trace_artifact"},
							{From: "case.input.require_snapshot_evidence", To: "args.input.require_snapshot_evidence"},
							{From: "case.input.min_distinct_failed_checks", To: "args.input.min_distinct_failed_checks"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.passed", To: "state.review.passed"},
							{From: "result.summary", To: "state.review.summary"},
							{From: "result.valid_payload_count", To: "state.failure_payload.valid_payload_count"},
							{From: "result.distinct_failed_check_count", To: "state.failure_payload.distinct_failed_check_count"},
						},
						Config: map[string]any{
							"evaluator":   "browser_action_failure_payload_gate",
							"instruction": "Review the browser failed-action payloads. Pass only when each required failed_check is represented by an action_failed payload with failed actionability, matching failure_evidence.reason_code, snapshot evidence when required, and a trace_like failure artifact when required.",
						},
					},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "browser_open_target",
				Description: "打开业务目标页面并让浏览器状态停留在当前目标页。",
				RuntimeTool: "browser_act",
				RuntimeArgs: map[string]any{
					"kind":    "open",
					"wait_ms": 800,
				},
				Tags: []string{"browser", "navigation"},
			},
			{
				Name:        "browser_fill_fields",
				Description: "按字段列表批量填写浏览器表单，并可在同一动作里执行 submit。",
				RuntimeTool: "browser_act",
				RuntimeArgs: map[string]any{
					"kind":   "fill",
					"target": "current",
				},
				Tags: []string{"browser", "form", "mutating"},
			},
			{
				Name:        "browser_capture_page_snapshot",
				Description: "抓取页面 aria/AI snapshot，用于 evaluator 语义判断。",
				RuntimeTool: "browser_act",
				RuntimeArgs: map[string]any{
					"kind":         "snapshot",
					"target":       "current",
					"format":       "aria",
					"mode":         "efficient",
					"interactive":  true,
					"compact":      true,
					"max_chars":    6000,
					"max_elements": 80,
				},
				Tags: []string{"browser", "snapshot", "evidence"},
			},
			{
				Name:        "browser_capture_submission_evidence",
				Description: "保存最终页面截图，进入 artifact registry 与 event log。",
				RuntimeTool: "browser_screenshot",
				RuntimeArgs: map[string]any{
					"target":    "current",
					"full_page": true,
				},
				ArtifactTypes: []string{BrowserArtifactTypePageScreenshot},
				Tags:          []string{"browser", "screenshot", "evidence"},
			},
			{
				Name:        "browser_download_file",
				Description: "通过 host-injected browser runtime 下载文件 artifact。下载根目录、保留策略、审批和目标站点策略由 host 持有。",
				RuntimeTool: "browser_act",
				RuntimeArgs: map[string]any{
					"kind":   "download",
					"target": "current",
				},
				ArtifactTypes: []string{BrowserArtifactTypeDownloadedFile},
				Tags:          []string{"browser", "download", "artifact", "evidence"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "browser_submit_evidence_gate",
				Description: "检查浏览器页面快照与截图证据是否足以支撑“表单已正确填写并进入预期状态”的判断。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                map[string]any{"type": "boolean"},
						"score":                 map[string]any{"type": "number"},
						"summary":               map[string]any{"type": "string"},
						"visual_evidence_ready": map[string]any{"type": "boolean"},
						"snapshot_ready":        map[string]any{"type": "boolean"},
						"screenshot_ready":      map[string]any{"type": "boolean"},
						"final_url_ready":       map[string]any{"type": "boolean"},
						"action_trace_ready":    map[string]any{"type": "boolean"},
						"artifact_refs_ready":   map[string]any{"type": "boolean"},
						"failure_reasons": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"evidence": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"evidence_bundle": BrowserEvidenceBundleSchema(),
					},
				},
			},
			{
				Name:         "browser_page_state_gate",
				Description:  "检查浏览器页面 snapshot、截图证据、最终 URL 和页面状态期望是否一致。",
				OutputSchema: BrowserPageStateEvaluationSchema(),
			},
			{
				Name:         "browser_structured_data_gate",
				Description:  "检查浏览器页面结构化字段抽取结果是否完整，并确认字段值有页面 evidence 支撑。",
				OutputSchema: BrowserStructuredDataEvaluationSchema(),
			},
			{
				Name:         "browser_site_search_gate",
				Description:  "检查浏览器站内搜索结果页证据是否完整，并确认期望搜索结果出现在页面 snapshot 中。",
				OutputSchema: BrowserSiteSearchEvaluationSchema(),
			},
			{
				Name:         "browser_download_file_gate",
				Description:  "检查浏览器下载 artifact 元数据是否完整，并确认文件名、content-type、大小和 URL 证据满足 case 约束。",
				OutputSchema: BrowserDownloadFileEvaluationSchema(),
			},
			{
				Name:        "browser_action_failure_payload_gate",
				Description: "检查 browser action_failed payload 是否包含稳定的 actionability、failure_evidence、failed_check、reason_code、snapshot evidence 和 trace-like artifact。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                      map[string]any{"type": "boolean"},
						"summary":                     map[string]any{"type": "string"},
						"valid_payload_count":         map[string]any{"type": "integer"},
						"distinct_failed_check_count": map[string]any{"type": "integer"},
						"missing_failed_checks": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"failure_reasons": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
					},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:              "browser_submit_success_suite",
				Description:       "要求浏览器提交流程产出截图证据，并由 final gate 明确判定通过。",
				WorkflowIDs:       []string{DefaultWorkflow},
				RequiredArtifacts: []string{BrowserArtifactTypePageScreenshot},
				RequiredState:     []string{"review.passed", "review.summary", "review.snapshot", "review.evidence_path"},
				PassPath:          "review.passed",
				SummaryPath:       "review.summary",
				Default:           true,
			},
			{
				Name:          "browser_action_failure_payload_suite",
				Description:   "要求 failed browser action regression case 产出可机读失败 payload，并覆盖至少一组 actionability failed_check。",
				WorkflowIDs:   []string{ActionFailurePayloadWorkflow},
				RequiredState: []string{"review.passed", "review.summary", "failure_payload.valid_payload_count", "failure_payload.distinct_failed_check_count"},
				PassPath:      "review.passed",
				SummaryPath:   "review.summary",
			},
			{
				Name:              "browser_page_state_success_suite",
				Description:       "要求只读页面状态验证产出 snapshot、截图证据，并由 page-state gate 明确判定通过。",
				WorkflowIDs:       []string{VerifyPageStateWorkflow},
				RequiredArtifacts: []string{BrowserArtifactTypePageScreenshot},
				RequiredState:     []string{"page_state.passed", "page_state.summary", "review.snapshot", "review.evidence_path"},
				PassPath:          "page_state.passed",
				SummaryPath:       "page_state.summary",
			},
			{
				Name:              "browser_structured_data_success_suite",
				Description:       "要求结构化抽取产出 snapshot、截图证据、抽取字段和明确通过结果。",
				WorkflowIDs:       []string{ExtractStructuredDataWorkflow},
				RequiredArtifacts: []string{BrowserArtifactTypePageScreenshot},
				RequiredState:     []string{"extract.passed", "extract.summary", "extract.data", "review.snapshot", "review.evidence_path"},
				PassPath:          "extract.passed",
				SummaryPath:       "extract.summary",
			},
			{
				Name:              "browser_site_search_success_suite",
				Description:       "要求站内搜索产出 snapshot、截图证据、搜索结果验收和明确通过结果。",
				WorkflowIDs:       []string{SiteSearchWorkflow},
				RequiredArtifacts: []string{BrowserArtifactTypePageScreenshot},
				RequiredState:     []string{"site_search.passed", "site_search.summary", "review.snapshot", "review.evidence_path"},
				PassPath:          "site_search.passed",
				SummaryPath:       "site_search.summary",
			},
			{
				Name:              "browser_download_file_success_suite",
				Description:       "要求下载任务产出 downloaded-file artifact、页面证据和明确通过结果。",
				WorkflowIDs:       []string{DownloadFileWorkflow},
				RequiredArtifacts: []string{BrowserArtifactTypeDownloadedFile, BrowserArtifactTypePageScreenshot},
				RequiredState:     []string{"download.passed", "download.summary", "download.path", "review.snapshot", "review.evidence_path"},
				PassPath:          "download.passed",
				SummaryPath:       "download.summary",
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "browser_readonly",
				Contract: agentxexecution.Contract{
					ID:      "browser-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_runtime", "browser_extract", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_runtime", "browser_extract", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
			{
				Name: "browser_page_state_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-page-state-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_act", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_act", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:           agentxexecution.SideEffectBrowserMutate,
						StrictRecovery:     true,
						CrossSystemConfirm: false,
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
			{
				Name: "browser_extract_readonly_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-extract-readonly-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_act", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_act", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:           agentxexecution.SideEffectBrowserMutate,
						StrictRecovery:     true,
						CrossSystemConfirm: false,
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
			{
				Name: "browser_site_search_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-site-search-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_act", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_act", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:           agentxexecution.SideEffectBrowserMutate,
						StrictRecovery:     true,
						CrossSystemConfirm: false,
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
			{
				Name: "browser_download_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-download-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_act", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_act", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:           agentxexecution.SideEffectBrowserMutate,
						StrictRecovery:     true,
						CrossSystemConfirm: false,
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"download", "screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
			{
				Name: "browser_fill_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-fill-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"browser_act", "browser_screenshot", "llm_task"},
						DeclaredTools:   []string{"browser_act", "browser_screenshot", "llm_task"},
						RequireDeclared: true,
						MaxRisk:         "high",
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:           agentxexecution.SideEffectBrowserMutate,
						StrictRecovery:     true,
						CrossSystemConfirm: false,
					},
					Evidence: agentxexecution.EvidencePolicy{
						RequiredArtifacts: []string{"screenshot"},
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
				Default: true,
			},
			{
				Name: "browser_failure_payload_review",
				Contract: agentxexecution.Contract{
					ID:      "browser-failure-payload-review",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools:      []string{"llm_task"},
						DeclaredTools:   []string{"llm_task"},
						RequireDeclared: true,
						MaxRisk:         "medium",
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
			},
		},
		MemorySchemas: []agentxpack.MemorySchema{
			{
				Name:        "browser_run_memory",
				Description: "沉淀浏览器自动化运行中的目标、结果、证据与失败模式，供后续相似页面回放和问题归因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":       map[string]any{"type": "string"},
						"case_type":     map[string]any{"type": "string"},
						"workflow_id":   map[string]any{"type": "string"},
						"status":        map[string]any{"type": "string"},
						"summary":       map[string]any{"type": "string"},
						"intent":        map[string]any{"type": "string"},
						"target_url":    map[string]any{"type": "string"},
						"evidence_path": map[string]any{"type": "string"},
						"error":         map[string]any{"type": "string"},
						"state":         map[string]any{"type": "object"},
						"case_input":    map[string]any{"type": "object"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "status", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{
				"浏览器自动化",
				"表单提交",
				"页面证据",
				"selector 漂移",
			},
			Limit:      4,
			MaxChars:   1200,
			ScopedOnly: true,
		},
	}
}

func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}

func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
