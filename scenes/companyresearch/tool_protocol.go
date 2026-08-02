package companyresearch

import (
	"context"
	"encoding/json"
	"fmt"

	llm "github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

type StandardToolHandlers struct {
	Lookup  ToolPayloadHandler
	Compare ToolPayloadHandler
	Guard   ToolPayloadHandler
}

func RegisterStandardTools(reg *agentxtools.Registry, handlers StandardToolHandlers) {
	registerPayloadTool(reg, CompanyResearchLookupTool(), handlers.Lookup)
	registerPayloadTool(reg, CompanyCompareLookupTool(), handlers.Compare)
	registerPayloadTool(reg, CompanyResearchGuardTool(), handlers.Guard)
}

func DecodeToolArguments(raw string) (map[string]any, error) {
	out := map[string]any{}
	err := json.Unmarshal([]byte(raw), &out)
	if err != nil {
		return nil, fmt.Errorf("decode company-research tool arguments: %w", err)
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

func CompanyResearchLookupTool() llm.Tool {
	return functionTool(
		ToolCompanyResearchLookup,
		"High-level single-company research lookup. Use for composite company questions that may require financial reports, market/stock data, current share-price or valuation context, public news, announcements, research metadata, or risk framing. 中文组合问题如公司财报、股价/行情估值、最近新闻、风险边界和整体情况也应优先使用此入口。 The model supplies structured intent; downstream tools verify facts from public evidence.",
		toolProperties(false),
	)
}

func CompanyCompareLookupTool() llm.Tool {
	props := toolProperties(true)
	return functionTool(
		ToolCompanyCompareLookup,
		"High-level multi-company comparison lookup. Use when the user asks to compare two or more companies across financials, market/stock data, public news, announcements, research evidence, valuation context, or risk boundaries. 中文多公司财报、股价、新闻风险、估值和投资边界比较也应优先使用此入口。If the result contains answer_contract.recovery_recommended=true and final_answer_recommended=false, do not close the answer yet and do not call company_research_guard as the recovery step; call the suggested recovery tools for the listed recovery_targets when those tools are visible and budget remains, then synthesize only from verified evidence. If recovery tools are unavailable or still fail, return a bounded partial answer with the recovery targets and evidence limits.",
		props,
	)
}

func CompanyResearchGuardTool() llm.Tool {
	return functionTool(
		ToolCompanyResearchGuard,
		"Guard company research evidence before final answer. Checks evidence readiness, missing dimensions, downstream degradation, and answer-boundary risk. Do not use this as a standalone repair tool for a lookup/compare result with answer_contract.recovery_recommended=true; it cannot fetch missing evidence, so use the suggested recovery tools first.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request used to verify that evidence covers the requested company research task."},
			"intent":       map[string]any{"type": "object", "description": "Structured intent produced by company_research_lookup or company_compare_lookup."},
			"evidence":     map[string]any{"type": "object", "description": "Evidence payload gathered from host adapters and downstream finance, stock, news, or public-web tools."},
			"warnings":     toolStringArraySchema("Warnings collected during research orchestration."),
			"source_policy": map[string]any{
				"type":        "string",
				"description": "Host/source policy hint. This is not an authorization source.",
			},
		},
	)
}

func toolProperties(includeSubjects bool) map[string]any {
	props := map[string]any{
		"user_message":    map[string]any{"type": "string", "description": "Original user request. Preserve freshness and relative-date hints verbatim."},
		"task_kind":       map[string]any{"type": "string", "enum": []string{"company_research", "investment_risk_review", "business_performance_review", "reliability_review", "company_compare", "peer_review", "investment_risk_compare"}, "description": "Structured user intent label. This guides orchestration but is not a factual conclusion."},
		"entity_name":     map[string]any{"type": "string", "description": "Company/entity mention inferred from the request. Candidate only; downstream tools verify identity."},
		"entity_mentions": toolStringArraySchema("Original company/entity mentions, aliases, and tickers from the request."),
		"market_hint":     map[string]any{"type": "string", "description": "Optional structured market hint such as A-share, HK, US, or unknown. Candidate only."},
		"requested_dimensions": stringEnumArraySchema("Evidence dimensions requested by the user.", []string{
			"financials", "market_data", "news", "announcements", "research", "risk", "comparison", "valuation",
		}),
		"requested_outputs": stringEnumArraySchema("Answer products requested by the user.", []string{
			"brief", "evidence_table", "risk_summary", "positive_factors", "investment_boundary", "comparison",
		}),
		"freshness": map[string]any{
			"type":        "object",
			"description": "Freshness constraints from the user; host adapters decide how to satisfy or degrade them.",
			"properties": map[string]any{
				"mode":               map[string]any{"type": "string", "description": "Freshness mode such as latest, recent, period, or any."},
				"relative_date_hint": map[string]any{"type": "string", "description": "User's relative-date phrase, kept verbatim for host date resolution."},
				"published_after":    map[string]any{"type": "string", "description": "Lower publication-date bound when the user supplied an absolute or resolved date."},
				"published_before":   map[string]any{"type": "string", "description": "Upper publication-date bound when the user supplied an absolute or resolved date."},
				"require_latest":     map[string]any{"type": "boolean", "description": "Whether the answer must prefer the latest available public evidence."},
			},
		},
		"risk_scope":      map[string]any{"type": "string", "description": "Risk or assessment boundary requested by the user, such as investment risk, operational risk, or reliability."},
		"source_policy":   map[string]any{"type": "string", "description": "Optional host/source policy hint. This is not an authorization source."},
		"original_intent": map[string]any{"type": "string", "description": "Original natural-language intent preserved for audit and downstream routing."},
		"stop_condition":  map[string]any{"type": "string", "description": "Caller-provided condition for when the research lookup has enough evidence to stop."},
	}
	if includeSubjects {
		props["comparison_subjects"] = map[string]any{
			"type":        "array",
			"description": "Companies/entities to compare. Candidate only; downstream tools verify identity.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"entity_name":       map[string]any{"type": "string", "description": "Company/entity mention for this comparison subject. Candidate only."},
					"entity_mentions":   toolStringArraySchema("Aliases or tickers for this subject."),
					"market_hint":       map[string]any{"type": "string", "description": "Optional market hint for this subject, such as A-share, HK, US, or unknown."},
					"source_policy":     map[string]any{"type": "string", "description": "Optional per-subject source policy hint owned by the host."},
					"requested_outputs": toolStringArraySchema("Optional per-subject output hints."),
				},
			},
		}
	}
	return props
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

func toolStringArraySchema(description string) map[string]any {
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
