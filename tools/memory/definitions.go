package memory

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

// SearchDefinition returns the stable unified memory search schema.
func SearchDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        SearchName,
		Description: "Search durable memory files, structured memory records, and optionally persisted sessions, for relevant recall snippets. Prefer this when you want a unified memory entrypoint.",
		Parameters: closedSchema(map[string]any{
			"query":            stringSchema("Recall query to search across selected memory sources."),
			"limit":            intSchema("Maximum number of hits to return per source. The runtime clamps this to the configured limit.", 1),
			"include_sessions": boolSchema("Include persisted session recall in addition to default memory sources."),
			"sources": map[string]any{
				"description": "Memory sources to query. Supported canonical sources are memory, structured, and sessions; durable memory path aliases are normalized to memory.",
				"oneOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
			"session_id":       stringSchema("Persisted session id to focus when sessions are included."),
			"status":           stringSchema("Single session status filter applied when sessions are included."),
			"statuses":         stringArraySchema("Session status filters applied when sessions are included."),
			"model":            stringSchema("Single model filter applied when sessions are included."),
			"models":           stringArraySchema("Model filters applied when sessions are included."),
			"tag":              stringSchema("Single session tag filter applied when sessions are included."),
			"tags":             stringArraySchema("Session tag filters applied when sessions are included."),
			"candidate_limit":  intSchema("Maximum session candidates to hydrate before recall ranking.", 1),
			"rerank_limit":     intSchema("Maximum session hits to rerank after initial scoring.", 1),
			"include_clusters": boolSchema("Include session recall cluster summaries when supported by the recall backend."),
			"cluster_limit":    intSchema("Maximum session recall clusters to return when clusters are included.", 1),
			"history_lines":    intSchema("Maximum recent message lines to include while hydrating session recall context.", 1),
		}, []string{"query"}),
		OutputSchema: searchOutputSchema(),
	}}
}

// GetDefinition returns the stable memory snippet schema.
func GetDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        GetName,
		Description: "Read a specific snippet from MEMORY.md or memory/*.md.",
		Parameters: closedSchema(map[string]any{
			"path":  stringSchema("Memory file path to read, such as MEMORY.md or memory/topic.md."),
			"from":  intSchema("One-based starting line to read from. Defaults to 1.", 1),
			"lines": intSchema("Maximum number of lines to return. The runtime clamps this to the configured limit.", 1),
		}, []string{"path"}),
		OutputSchema: closedSchema(map[string]any{
			"path":       stringSchema("Resolved memory file path relative to the memory root."),
			"line_start": intSchema("One-based first line represented in text.", 1),
			"line_end":   intSchema("One-based last line represented in text. May be less than line_start when no content was returned.", 0),
			"text":       stringSchema("Memory snippet text joined with newline characters."),
		}, []string{"path", "line_start", "line_end", "text"}),
	}}
}

func searchOutputSchema() map[string]any {
	return closedSchema(map[string]any{
		"query":             stringSchema("Recall query that was executed."),
		"query_tokens":      stringArraySchema("Expanded query tokens used for lexical or hybrid recall."),
		"query_token_kinds": looseObjectSchema("Counts grouped by token expansion kind."),
		"backend":           backendStatusSchema("Durable memory backend diagnostics for memory-only responses."),
		"backends":          looseObjectSchema("Per-source backend diagnostics for unified multi-source responses."),
		"diagnostics":       searchDiagnosticsSchema(),
		"hits":              searchHitArraySchema("Ranked memory, structured memory, or session recall hits."),
		"hit_count":         intSchema("Number of unified hits returned.", 0),
		"ranking_mode":      stringSchema("Unified ranking strategy used across sources."),
		"top_hit":           looseObjectSchema("Best-ranked hit copied from hits for quick inspection."),
		"top_hit_source":    stringSchema("Source name for the best-ranked hit."),
		"top_hit_score":     numberSchema("Score for the best-ranked hit."),
		"sources_requested": stringArraySchema("Canonical memory sources requested by the caller."),
		"sources":           looseObjectSchema("Per-source payloads and availability diagnostics."),
		"source_status":     stringMapSchema("Per-source status summary such as ok, partial, degraded, or unavailable."),
		"status":            stringSchema("Aggregated source status for unified responses."),
		"warnings":          stringArraySchema("Warnings explaining partial, degraded, disabled, or unavailable sources."),
		"actions":           stringArraySchema("Operator actions suggested for unavailable or degraded sources."),
		"disabled":          boolSchema("True when all requested sources are disabled or unavailable."),
		"unavailable":       boolSchema("True when at least one requested source is unavailable."),
		"degraded":          boolSchema("True when at least one requested source is running in degraded mode."),
	}, []string{"query"})
}

func backendStatusSchema(description string) map[string]any {
	schema := closedSchema(map[string]any{
		"configured": stringSchema("Configured backend or scorer name."),
		"active":     stringSchema("Backend or scorer actually used for the request."),
		"kind":       stringSchema("Backend class such as lexical, semantic, hybrid, or structured."),
		"fallback":   stringSchema("Fallback backend used when the configured backend was unavailable."),
		"available":  boolSchema("True when the backend was available for the request."),
		"degraded":   boolSchema("True when the backend served a fallback or degraded result."),
		"reason":     stringSchema("Backend availability or fallback reason."),
	}, nil)
	schema["description"] = description
	return schema
}

func searchDiagnosticsSchema() map[string]any {
	schema := closedSchema(map[string]any{
		"query_tokens":      stringArraySchema("Expanded query tokens used by durable memory search."),
		"query_token_kinds": looseObjectSchema("Counts grouped by token expansion kind."),
		"memory_sources":    looseObjectSchema("Matched durable memory source counts."),
		"memory_sections":   looseObjectSchema("Matched durable memory section counts."),
		"index":             indexStatusSchema("Durable memory index diagnostics."),
		"backend":           backendStatusSchema("Durable memory backend diagnostics."),
		"hit_count":         intSchema("Number of durable memory hits returned.", 0),
	}, nil)
	schema["description"] = "Durable memory search diagnostics for memory-only responses."
	return schema
}

func indexStatusSchema(description string) map[string]any {
	schema := closedSchema(map[string]any{
		"mode":               stringSchema("Index strategy used for candidate narrowing."),
		"available":          boolSchema("True when an index backend was available."),
		"used":               boolSchema("True when the index participated in this query."),
		"candidate_files":    intSchema("Number of candidate files selected by the index.", 0),
		"candidate_chunks":   intSchema("Number of candidate chunks selected by the index.", 0),
		"scanned_files":      intSchema("Number of files scanned after candidate selection.", 0),
		"total_files":        intSchema("Total durable memory files considered by the search engine.", 0),
		"full_scan_fallback": boolSchema("True when the search engine fell back to a full scan."),
		"matched_sources":    looseObjectSchema("Matched durable memory source counts."),
		"matched_sections":   looseObjectSchema("Matched durable memory section counts."),
		"ranking_signals":    stringArraySchema("Ranking signals applied to durable memory hits."),
		"reason":             stringSchema("Index availability or fallback reason."),
	}, nil)
	schema["description"] = description
	return schema
}

func searchHitArraySchema(description string) map[string]any {
	return map[string]any{
		"type": "array", "description": description,
		"items": map[string]any{
			"type": "object", "additionalProperties": true,
			"properties": map[string]any{
				"source":           stringSchema("Source name for the hit, such as memory, structured, or sessions."),
				"path":             stringSchema("Durable memory file path for memory hits."),
				"line":             intSchema("Primary one-based line number for durable memory hits.", 0),
				"line_start":       intSchema("One-based first line represented by a durable memory hit.", 0),
				"line_end":         intSchema("One-based last line represented by a durable memory hit.", 0),
				"section":          stringSchema("Durable memory section label when available."),
				"snippet":          stringSchema("Durable memory snippet text."),
				"session_id":       stringSchema("Persisted session id for session recall hits."),
				"excerpt":          stringSchema("Session recall excerpt text."),
				"memory_record_id": stringSchema("Structured memory record id for structured hits."),
				"score":            numberSchema("Hit score reported by the source or unified ranker."),
				"merged_rank":      intSchema("Rank after cross-source merge.", 0),
				"source_rank":      intSchema("Rank within the source before cross-source merge.", 0),
				"updated_at":       intSchema("Source update time in Unix milliseconds when available.", 0),
				"updated_at_iso":   stringSchema("Source update time as an ISO timestamp when available."),
			},
		},
	}
}

func closedSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": required}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}
func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}
func numberSchema(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}
func stringArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "string"}}
}
func looseObjectSchema(description string) map[string]any {
	return map[string]any{"type": "object", "description": description, "additionalProperties": true}
}
func stringMapSchema(description string) map[string]any {
	return map[string]any{"type": "object", "description": description, "additionalProperties": map[string]any{"type": "string"}}
}
