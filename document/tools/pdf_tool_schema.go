package tools

func pdfUnifiedParametersSchema() map[string]any {
	return closedInputSchema(map[string]any{
		"prompt":              stringSchema("Question, summarization request, comparison request, or analysis instruction for the PDF model."),
		"query":               stringSchema("Compatibility alias for prompt."),
		"instruction":         stringSchema("Compatibility alias for prompt."),
		"task":                stringSchema("Compatibility alias for prompt."),
		"pdf":                 stringSchema("Single workspace path, file:// URL, or http(s) PDF URL."),
		"pdfs":                stringArraySchema("Multiple workspace paths, file:// URLs, or http(s) PDF URLs for comparison or document-set analysis."),
		"path":                stringSchema("Compatibility alias for pdf when the input is a local workspace path."),
		"file_path":           stringSchema("Compatibility alias for pdf when the input is a local workspace path."),
		"url":                 stringSchema("Compatibility alias for pdf when the input is a file:// or http(s) URL."),
		"source_url":          stringSchema("Compatibility alias for pdf when the input is a source URL."),
		"model":               stringSchema("Optional preferred PDF-capable model or configured submodel name. Use a concrete model/config name, not route labels such as vision or text_chat."),
		"vision_model":        stringSchema("Compatibility alias for model; use a concrete model/config name, not route labels."),
		"pages":               pdfUnifiedPagesInputSchema("Selected pages as an integer list or range string such as 1-3,5."),
		"page_range":          stringSchema("Selected page range such as 1-3,5."),
		"pageRange":           stringSchema("Compatibility alias for page_range."),
		"start_page":          intSchema("First selected page for a contiguous range.", 1),
		"end_page":            intSchema("Last selected page for a contiguous range.", 1),
		"max_pages":           intSchema("Maximum pages to inspect per PDF. Runtime clamps to its configured hard limit.", 1),
		"max_context_chars":   intSchema("Maximum text context characters sent to the model. Runtime clamps to its configured hard limit.", 200),
		"max_visual_pages":    intSchema("Maximum rendered pages sent to a vision-capable route. Runtime clamps to its configured hard limit.", 1),
		"max_response_bytes":  intSchema("Maximum bytes to read when fetching a remote PDF. Runtime clamps to its configured hard limit.", 1),
		"max_bytes_mb":        pdfUnifiedNumberSchema("Maximum bytes to read when fetching a remote PDF, expressed in megabytes. Runtime clamps to its configured hard limit.", 0),
		"maxBytesMb":          pdfUnifiedNumberSchema("Compatibility alias for max_bytes_mb.", 0),
		"timeout_ms":          intSchema("Maximum remote fetch runtime in milliseconds. Runtime clamps to its configured hard limit.", 1),
		"include_diagnostics": boolSchema("Include route, capability, and analysis-plan diagnostics in successful output."),
	}, nil)
}

func pdfUnifiedPagesInputSchema(description string) map[string]any {
	return map[string]any{
		"description": description,
		"anyOf": []any{
			map[string]any{
				"type":        "array",
				"description": "Explicit page numbers.",
				"items":       map[string]any{"type": "integer", "minimum": 1},
			},
			stringSchema("Page range string such as 1-3,5."),
		},
	}
}

func pdfUnifiedNumberSchema(description string, minimum float64) map[string]any {
	schema := numberSchema(description)
	schema["minimum"] = minimum
	return schema
}

func pdfUnifiedOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"status":             stringEnumSchema("Execution status for the unified PDF route.", "success", "failed", "unavailable", "invalid_input"),
		"prompt":             stringSchema("Resolved prompt used for PDF analysis."),
		"model":              stringSchema("Selected PDF-capable model or configured submodel name."),
		"client":             stringSchema("Selected model client/provider name."),
		"native_pdf":         boolSchema("True when the final answer came from a native PDF provider route."),
		"document_count":     intSchema("Number of resolved PDF inputs.", 0),
		"documents":          pdfUnifiedDocumentsOutputSchema(),
		"analysis_plan":      pdfUnifiedAnalysisPlanOutputSchema(),
		"attempted_models":   stringArraySchema("PDF model candidates attempted in order."),
		"fallback_used":      boolSchema("True when a non-primary PDF model produced the final response."),
		"focus_enabled":      boolSchema("True when document focus classification shaped evidence selection."),
		"focus_query_class":  stringEnumSchema("Focus query class inferred from prompt and document structure.", pdfUnifiedQueryClassGeneric, pdfUnifiedQueryClassFieldCompare, pdfUnifiedQueryClassChartSummary),
		"focus_reason_codes": stringArraySchema("Reason codes that enabled focused PDF evidence selection."),
		"focus_confidence":   stringSchema("Confidence for the focused evidence selection."),
		"route":              pdfUnifiedRouteTraceOutputSchema(),
		"capability_matrix":  pdfUnifiedCapabilityMatrixOutputSchema(),
		"answer":             stringSchema("Grounded PDF answer or summary."),
		"answer_scope":       stringSchema("Downstream scope contract for interpreting the resolved answer versus supporting diagnostics."),
		"warning":            stringSchema("Non-fatal routing, page-selection, fallback, OCR, rendering, or backend warning."),
	}, []string{"status"})
}

func pdfUnifiedDocumentsOutputSchema() map[string]any {
	return objectArraySchema("Resolved PDF documents and evidence projections.", map[string]any{
		"path":                stringSchema("Human-readable PDF input path or URL label."),
		"page_count":          intSchema("Total PDF page count when available.", 0),
		"selected_pages":      pdfUnifiedIntArrayOutputSchema("Selected pages applied to this PDF."),
		"page_limit_applied":  boolSchema("True when the configured page limit truncated inspection."),
		"selection_strategy":  stringEnumSchema("How the bounded page set was selected.", pdfUnifiedSelectionAll, pdfUnifiedSelectionExplicit, pdfUnifiedSelectionPrefix, pdfUnifiedSelectionQuery),
		"text_chars":          intSchema("Total extracted text characters used for this PDF.", 0),
		"visual_pages":        pdfUnifiedIntArrayOutputSchema("Rendered pages supplied to a vision-capable route."),
		"evidence_pages":      pdfAnalyzePageItemsOutputSchema("Page evidence excerpts selected for the prompt."),
		"structure_items":     pdfUnifiedStructureItemsOutputSchema("Structure and layout signals inferred for this PDF."),
		"segments":            pdfUnifiedSegmentsOutputSchema("Document segments inferred for focus routing."),
		"primary_segment":     pdfUnifiedSegmentOutputSchema("Primary document segment selected for focus routing."),
		"supporting_segments": pdfUnifiedSegmentsOutputSchema("Supporting document segments selected for focus routing."),
	})
}

func pdfUnifiedAnalysisPlanOutputSchema() map[string]any {
	schema := closedOutputSchema(map[string]any{
		"mode":                    stringSchema("Resolved PDF analysis mode."),
		"needs_vision":            boolSchema("True when rendered visual pages are needed."),
		"needs_ocr":               boolSchema("True when OCR enrichment is recommended or used."),
		"preferred_backend":       stringSchema("Preferred PDF extraction backend class."),
		"preferred_clients":       stringArraySchema("Preferred model clients for this PDF route."),
		"provider_routing":        stringSchema("Provider routing strategy used for non-native PDF analysis."),
		"native_provider_routing": stringSchema("Provider routing strategy used for native PDF analysis."),
		"reason":                  stringSchema("Human-readable routing rationale."),
		"warning":                 stringSchema("Analysis-plan warning."),
		"candidate_models":        pdfUnifiedCandidateModelsOutputSchema("Candidate PDF-capable models considered by the route."),
		"suggested_next_steps":    stringArraySchema("Suggested follow-up PDF tools or actions."),
	}, nil)
	schema["description"] = "PDF analysis plan and model-routing hints."
	return schema
}

func pdfUnifiedRouteTraceOutputSchema() map[string]any {
	schema := closedOutputSchema(map[string]any{
		"selected_route":               stringSchema("Final PDF route, such as text_chat, rendered_vision, or native_pdf."),
		"selected_model":               stringSchema("Model selected for the final route."),
		"attempted_routes":             stringArraySchema("Routes attempted for the selected model."),
		"available_routes":             stringArraySchema("Routes available for the selected model and inputs."),
		"limitations":                  stringArraySchema("Route limitations or downgraded capability reasons."),
		"page_selection_requested":     boolSchema("True when the caller requested specific pages."),
		"page_selection_downgrade":     boolSchema("True when page selection forced downgrade from native PDF routing."),
		"visual_input_prepared":        boolSchema("True when rendered page visuals were prepared."),
		"text_input_available":         boolSchema("True when extracted text was available."),
		"native_page_selection_policy": stringSchema("Policy applied when page selection and native PDF routing conflict."),
		"policy_decision":              stringSchema("Specific policy decision made by the route."),
	}, nil)
	schema["description"] = "Selected route and PDF capability diagnostics."
	return schema
}

func pdfUnifiedCapabilityMatrixOutputSchema() map[string]any {
	return objectArraySchema("Per-model PDF route capability matrix.", map[string]any{
		"model":            stringSchema("Candidate model or submodel name."),
		"client":           stringSchema("Candidate client/provider name."),
		"native_pdf":       boolSchema("True when the candidate supports native PDF input."),
		"supports_vision":  boolSchema("True when the candidate supports rendered-page vision input."),
		"available_routes": stringArraySchema("Available routes for this candidate."),
		"limitations":      stringArraySchema("Route limitations for this candidate."),
	})
}

func pdfUnifiedCandidateModelsOutputSchema(description string) map[string]any {
	return objectArraySchema(description, map[string]any{
		"name":       stringSchema("Candidate configured submodel name."),
		"client":     stringSchema("Candidate client/provider name."),
		"model":      stringSchema("Provider model name."),
		"native_pdf": boolSchema("True when this candidate supports native PDF input."),
	})
}

func pdfAnalyzePageItemsOutputSchema(description string) map[string]any {
	return objectArraySchema(description, map[string]any{
		"page":    intSchema("PDF page number.", 1),
		"chars":   intSchema("Extracted characters on this page.", 0),
		"empty":   boolSchema("True when no usable text was extracted for this page."),
		"excerpt": stringSchema("Grounding excerpt from this page."),
	})
}

func pdfUnifiedStructureItemsOutputSchema(description string) map[string]any {
	return objectArraySchema(description, map[string]any{
		"id":            stringSchema("Stable structure item identifier."),
		"page":          intSchema("PDF page number.", 1),
		"content_layer": stringSchema("Content layer such as body or furniture."),
		"block_kind":    stringSchema("Inferred block kind such as text, chart/table, key-value, or signature."),
		"role":          stringSchema("Inferred document role represented by this structure item."),
		"confidence":    stringSchema("Confidence label for this structure item."),
		"anchors":       stringArraySchema("Evidence anchors for this structure item."),
		"excerpt":       stringSchema("Text excerpt backing this structure item."),
		"signal_codes":  stringArraySchema("Signal codes that produced this structure item."),
	})
}

func pdfUnifiedSegmentsOutputSchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": description,
		"items":       pdfUnifiedSegmentOutputSchema("Document segment item."),
	}
}

func pdfUnifiedSegmentOutputSchema(description string) map[string]any {
	schema := closedOutputSchema(map[string]any{
		"id":           stringSchema("Stable segment identifier."),
		"kind":         stringSchema("Inferred segment kind."),
		"pages":        pdfUnifiedIntArrayOutputSchema("Pages covered by this segment."),
		"page_start":   intSchema("First page covered by this segment.", 0),
		"page_end":     intSchema("Last page covered by this segment.", 0),
		"confidence":   stringSchema("Confidence label for this segment."),
		"anchors":      stringArraySchema("Evidence anchors for this segment."),
		"excerpt":      stringSchema("Text excerpt backing this segment."),
		"signal_codes": stringArraySchema("Signal codes that produced this segment."),
	}, nil)
	schema["description"] = description
	return schema
}

func pdfUnifiedIntArrayOutputSchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": description,
		"items":       map[string]any{"type": "integer"},
	}
}
