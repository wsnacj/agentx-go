package hostkit

import (
	"strings"

	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

// IntentFromParams converts model-supplied tool arguments into a neutral task frame.
func IntentFromParams(params map[string]any) globalcontracts.InvestigationIntent {
	requestedFields := StringSliceArg(firstValue(params, "requested_fields", "quote_fields", "profile_fields", "announcement_types", "research_fields"))
	requestedOutputs := StringSliceArg(params["requested_outputs"])
	intent := globalcontracts.InvestigationIntent{
		UserMessage:      stringArg(params["user_message"]),
		TaskKind:         globalcontracts.TaskKind(stringArg(params["task_kind"])),
		EntityName:       stringArg(params["entity_name"]),
		EntityMentions:   StringSliceArg(params["entity_mentions"]),
		StockCode:        stringArg(params["stock_code"]),
		Market:           globalcontracts.Market(strings.ToLower(stringArg(params["market"]))),
		RequestedFields:  mergeStringSlices(StringSliceArg(params["default_requested_fields"]), requestedFields),
		RequestedOutputs: mergeStringSlices(StringSliceArg(params["default_requested_outputs"]), requestedOutputs),
		SourceHint:       stringArg(params["source_hint"]),
		SourcePolicy:     stringArg(params["source_policy"]),
		OriginalIntent:   stringArg(params["original_intent"]),
		StopCondition:    stringArg(params["stop_condition"]),
	}
	if intent.Market == "" {
		intent.Market = globalcontracts.MarketAuto
	}
	if len(intent.EntityMentions) == 0 && intent.EntityName != "" {
		intent.EntityMentions = []string{intent.EntityName}
	}
	if freshness, ok := params["freshness"].(map[string]any); ok {
		intent.Freshness = globalcontracts.Freshness{
			Mode:                    globalcontracts.FreshnessMode(stringArg(freshness["mode"])),
			RelativeDateHint:        stringArg(freshness["relative_date_hint"]),
			TradeDate:               stringArg(freshness["trade_date"]),
			RequireRealtime:         boolArg(freshness["require_realtime"]),
			RequireLatestTradingDay: boolArg(freshness["require_latest_trading_day"]),
		}
	}
	return intent
}

func mergeStringSlices(groups ...[]string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, group := range groups {
		for _, value := range group {
			value = strings.TrimSpace(value)
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

// ParamsFromIntent overlays the normalized intent back onto tool arguments.
func ParamsFromIntent(base map[string]any, intent globalcontracts.InvestigationIntent) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	out["user_message"] = intent.UserMessage
	out["entity_name"] = intent.EntityName
	out["entity_mentions"] = intent.EntityMentions
	out["stock_code"] = intent.StockCode
	out["market"] = string(intent.Market)
	out["requested_fields"] = intent.RequestedFields
	out["requested_outputs"] = intent.RequestedOutputs
	out["source_policy"] = intent.SourcePolicy
	return out
}

// StringSliceArg converts loose JSON values into strings.
func StringSliceArg(value any) []string {
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				out = append(out, strings.TrimSpace(item))
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s := stringArg(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return nil
	}
}

func stringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func boolArg(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}

func firstValue(params map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := params[key]; ok {
			return value
		}
	}
	return nil
}
