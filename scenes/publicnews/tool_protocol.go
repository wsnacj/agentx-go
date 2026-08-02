package publicnews

import (
	"context"
	"encoding/json"
	"fmt"

	llm "github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	ToolLatestNewsLookup  = "latest_news_lookup"
	ToolLatestNewsExtract = "latest_news_extract"
	ToolLatestNewsGuard   = "latest_news_guard"
)

type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

type ToolHandlers struct {
	Lookup  ToolPayloadHandler
	Extract ToolPayloadHandler
	Guard   ToolPayloadHandler
}

func RegisterTools(reg *agentxtools.Registry, handlers ToolHandlers) {
	RegisterLatestNewsLookupTool(reg, handlers.Lookup)
	registerPayloadTool(reg, LatestNewsExtractTool(), handlers.Extract)
	registerPayloadTool(reg, LatestNewsGuardTool(), handlers.Guard)
}

func RegisterLatestNewsLookupTool(reg *agentxtools.Registry, handler ToolPayloadHandler) {
	registerPayloadTool(reg, LatestNewsLookupTool(), handler)
}

func DecodeToolArguments(raw string) (map[string]any, error) {
	out := map[string]any{}
	err := json.Unmarshal([]byte(raw), &out)
	if err != nil {
		return nil, fmt.Errorf("decode public-news tool arguments: %w", err)
	}
	return out, nil
}

func registerPayloadTool(reg *agentxtools.Registry, tool llm.Tool, handler ToolPayloadHandler) {
	if reg == nil || handler == nil {
		return
	}
	reg.Register(tool, func(ctx context.Context, call llm.FunctionCall) (string, error) {
		params, err := DecodeToolArguments(call.Arguments)
		if err != nil {
			return "", err
		}
		payload, err := handler(ctx, params)
		if err != nil {
			return "", err
		}
		switch typed := payload.(type) {
		case string:
			return typed, nil
		case []byte:
			return string(typed), nil
		default:
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		}
	})
}

func LatestNewsExtractTool() llm.Tool {
	return functionTool(
		ToolLatestNewsExtract,
		"Host-wired extractor for latest-news tasks from an opened page or page text. In a latest-news-brief flow, call this immediately after a successful open_page/web_fetch/browser read and before any final answer. Prefer passing the open_page page_id so the host adapter can reuse cached page text, title, and final URL.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request for the latest-news brief."},
			"page_id":      map[string]any{"type": "string", "description": "page_id returned by open_page for the primary article. Prefer this over copying long page text."},
			"text":         map[string]any{"type": "string", "description": "Readable article text when the source came from web_fetch or browser instead of open_page cache."},
			"title":        map[string]any{"type": "string", "description": "Article title when not already available through page_id."},
			"source_url":   map[string]any{"type": "string", "description": "Final article URL when not already available through page_id."},
		},
	)
}

func LatestNewsLookupTool() llm.Tool {
	return functionTool(
		ToolLatestNewsLookup,
		"High-level latest-news lookup. Public-news-owned routing cues include 最新新闻、最新资讯、要闻、快讯和热点进展. Use this as the first tool for natural-language latest-news/latest-update requests when the host provides a lookup handler, including requests that ask to report provider/source unavailability. The model supplies only structured intent; host adapters must perform search/open/source selection, extract grounded article facts, cross-check independent sources, return guard status, and produce bounded answer contracts for provider-unavailable or source-quality diagnostic stops. If answer_contract.final_answer_recommended=true, use that boundary instead of retrying the same blocked source with low-level search/open/web_fetch/browser tools.",
		map[string]any{
			"user_message": map[string]any{
				"type":        "string",
				"description": "Original user request. Preserve relative-date and freshness hints verbatim.",
			},
			"task_kind": map[string]any{
				"type":        "string",
				"enum":        []string{"latest_news_brief", "latest_update_brief", "source_verification"},
				"description": "Structured user intent. This is not a factual conclusion; host adapters verify source freshness and evidence.",
			},
			"topic": map[string]any{
				"type":        "string",
				"description": "Topic/event/company/region mention inferred from the user request. Candidate only; source selection and facts require public evidence.",
			},
			"entity_mentions": stringArraySchema("Original topic/entity mentions from the request, including aliases or bilingual names when present."),
			"requested_fields": stringEnumArraySchema(
				"Fields requested by the user. Include impact/risks/next_steps only when requested by the user.",
				[]string{"headline", "published_at", "key_update", "source_url", "impact", "risks", "next_steps", "source_verification"},
			),
			"requested_outputs": stringEnumArraySchema(
				"Answer products requested by the user.",
				[]string{"brief", "timeline", "source_verification", "risk_summary", "impact_assessment"},
			),
			"freshness": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mode":               map[string]any{"type": "string", "description": "freshness mode such as latest, breaking, explicit_date_hint"},
					"relative_date_hint": map[string]any{"type": "string", "description": "Relative date phrase from the user, kept verbatim for source-date guard."},
					"published_after":    map[string]any{"type": "string", "description": "Earliest acceptable publication date or timestamp from the user or host policy."},
					"published_before":   map[string]any{"type": "string", "description": "Latest acceptable publication date or timestamp from the user or host policy."},
					"require_latest":     map[string]any{"type": "boolean", "description": "Whether selected sources must represent the latest public update for the requested topic."},
				},
				"description": "Freshness constraints from the user. Host adapters must verify selected sources against these constraints.",
			},
			"source_hint":        map[string]any{"type": "string", "description": "User-requested publisher, source class, region, or source hint. This is intent only; host adapters still rank and verify sources."},
			"source_policy":      map[string]any{"type": "string", "description": "Optional host/source policy hint for public news lookup, such as official-only, reputable media, or independent-source requirements. This is not an authorization source."},
			"cross_check_policy": map[string]any{"type": "string", "description": "Requested cross-check policy, for example independent_sources_required; host adapters decide source ranking and verification mechanics."},
			"original_intent":    map[string]any{"type": "string", "description": "Optional normalized intent from an upstream router, preserved for audit and guard checks."},
			"stop_condition":     map[string]any{"type": "string", "description": "Optional host or workflow stop condition for avoiding repeated calls after answer_contract or source-quality guard has produced a final boundary."},
		},
	)
}

func LatestNewsGuardTool() llm.Tool {
	return functionTool(
		ToolLatestNewsGuard,
		"Host-wired guard that confirms whether a latest-news page has enough grounded published_at/key_update/source_url facts for a brief. This is the mandatory final readiness gate for latest-news-brief: call it after latest_news_extract and before the final answer. For cross-check, pass the primary article page_id/source_url plus one or more independent sources in supporting_sources; do not rely on source_count alone. Primary and supporting sources need grounded page text through page_id, text, or a host-cached source_url; search snippets alone do not satisfy cross-check.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request for the latest-news brief."},
			"page_id":      map[string]any{"type": "string", "description": "page_id returned by open_page for the primary article. Required when the primary source came from open_page; omitting it may make the guard fail with grounded_page_text missing."},
			"text":         map[string]any{"type": "string", "description": "Primary article text when no open_page page_id is available."},
			"title":        map[string]any{"type": "string", "description": "Primary article title or headline."},
			"source_url":   map[string]any{"type": "string", "description": "Primary article final URL."},
			"headline":     map[string]any{"type": "string", "description": "Primary article headline extracted from latest_news_extract or the page title."},
			"source_site":  map[string]any{"type": "string", "description": "Primary article source site. The guard will prefer the URL host when source_url is present."},
			"published_at": map[string]any{"type": "string", "description": "Publication time from the primary article, ideally copied from latest_news_extract output."},
			"key_update":   map[string]any{"type": "string", "description": "Latest concrete update from the primary article, ideally copied from latest_news_extract output."},
			"source_count": map[string]any{"type": "number", "minimum": 1, "description": "Claimed number of sources. This does not satisfy cross-check by itself; supporting_sources must include independent-source evidence."},
			"supporting_sources": map[string]any{
				"type":        "array",
				"description": "Second and later independent sources used for cross-check. Include supporting published_at and key_update when available, use a different source site from the primary article, and include page_id/text or a host-cached source_url so the source is grounded beyond search snippets.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"page_id":      map[string]any{"type": "string", "description": "open_page page_id for this supporting article. Preferred for grounding."},
						"text":         map[string]any{"type": "string", "description": "Readable supporting article text when no page_id is available."},
						"headline":     map[string]any{"type": "string", "description": "Supporting source headline."},
						"source_url":   map[string]any{"type": "string", "description": "Supporting source final URL."},
						"source_site":  map[string]any{"type": "string", "description": "Supporting source site; should differ from the primary source site."},
						"published_at": map[string]any{"type": "string", "description": "Supporting source publication time."},
						"key_update":   map[string]any{"type": "string", "description": "Supporting source update confirming or qualifying the primary article."},
					},
				},
			},
		},
	)
}

func functionTool(name string, description string, properties map[string]any) llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        name,
			Description: description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   []string{"user_message"},
			},
		},
	}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": description,
	}
}

func stringEnumArraySchema(description string, values []string) map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "string",
			"enum": values,
		},
		"description": description,
	}
}
