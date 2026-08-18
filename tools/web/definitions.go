package web

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

// SearchDefinition returns the preferred web search schema.
func SearchDefinition() toolcontract.Definition {
	return searchDefinition(SearchName, "Search the web and return ranked source candidates with provider diagnostics.")
}

// WebSearchDefinition returns the compatibility web_search schema.
func WebSearchDefinition() toolcontract.Definition {
	return searchDefinition(WebSearchName, "Search the web using an explicitly configured provider and return ranked results.")
}

func searchDefinition(name, description string) toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: name, Description: description, Parameters: object(map[string]any{
		"query":       text("Natural-language search query."),
		"provider":    enum("Optional provider override. ark is a compatibility alias for doubao_custom.", "brave", "perplexity", "openrouter", "doubao_custom", "ark", "baidu"),
		"max_results": integer("Maximum results to return.", 1, 10),
		"count":       integer("Compatibility alias for max_results.", 1, 10),
		"country":     text("Optional country preference."), "search_lang": text("Optional search language."), "ui_lang": text("Optional UI language."),
		"freshness": text("Optional provider-specific recency filter."), "date_after": text("Optional start date YYYY-MM-DD."), "date_before": text("Optional end date YYYY-MM-DD."),
		"domain_filter": array(text("Domain allow/deny entry.")), "timeout_ms": integer("Request timeout in milliseconds.", 1, 120000),
		"authoritative_only": map[string]any{"type": "boolean", "description": "Limit results to the provider's highest authority tier when supported."},
		"query_rewrite":      map[string]any{"type": "boolean", "description": "Allow the provider to rewrite a conversational query when supported."},
	}, []string{"query"})}}
}

// WebFetchDefinition returns the low-level fetch schema.
func WebFetchDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: WebFetchName, Description: "Fetch and extract an explicit URL through Host-owned outbound policy.", Parameters: fetchParameters(true)}}
}

// OpenPageDefinition returns the readable-page schema.
func OpenPageDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: OpenPageName, Description: "Open a URL as a readable page and cache it for find_in_page.", Parameters: fetchParameters(false)}}
}

// FindInPageDefinition returns the cache-only page search schema.
func FindInPageDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: FindInPageName, Description: "Find matching snippets in a page_id returned by open_page.", Parameters: object(map[string]any{
		"page_id": text("Cached page identifier."), "query": text("Text to find."),
		"max_matches": integer("Maximum snippets.", 1, 20), "context_chars": integer("Surrounding characters.", 20, 400),
	}, []string{"page_id", "query"})}}
}

func fetchParameters(includeMode bool) map[string]any {
	properties := map[string]any{"url": text("Absolute HTTP or HTTPS URL."), "max_chars": integer("Maximum extracted characters.", 100, 80000), "max_response_bytes": integer("Maximum response bytes.", 16000, 10000000), "timeout_ms": integer("Request timeout in milliseconds.", 1, 120000)}
	if includeMode {
		properties["extract_mode"] = enum("Extraction mode.", "auto", "text", "markdown", "raw")
	}
	return object(properties, []string{"url"})
}

func object(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": required}
}
func text(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func enum(description string, values ...string) map[string]any {
	out := text(description)
	out["enum"] = values
	return out
}
func integer(description string, min, max int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": min, "maximum": max}
}
func array(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
