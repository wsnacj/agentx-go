package hostkit

import (
	"fmt"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
)

// IntentFromParams converts model-supplied tool arguments into a neutral task frame.
func IntentFromParams(params map[string]any) astockcontracts.InvestigationIntent {
	intent := astockcontracts.InvestigationIntent{
		UserMessage:      StringArg(params["user_message"]),
		TaskKind:         astockcontracts.TaskKind(StringArg(params["task_kind"])),
		EntityName:       StringArg(params["entity_name"]),
		EntityMentions:   StringSliceArg(params["entity_mentions"]),
		StockCode:        StringArg(params["stock_code"]),
		Market:           astockcontracts.Market(StringArg(params["market"])),
		RequestedFields:  StringSliceArg(params["requested_fields"]),
		RequestedOutputs: StringSliceArg(params["requested_outputs"]),
		Assessment:       StringArg(params["assessment"]),
		SourceHint:       StringArg(params["source_hint"]),
		SourcePolicy:     StringArg(params["source_policy"]),
		Freshness:        FreshnessArg(params["freshness"]),
	}
	if intent.Market == "auto" {
		intent.Market = ""
	}
	return intent
}

// ParamsFromIntent returns a shallow copy of params enriched with normalized intent fields.
func ParamsFromIntent(params map[string]any, intent astockcontracts.InvestigationIntent) map[string]any {
	out := map[string]any{}
	for key, value := range params {
		out[key] = value
	}
	putString(out, "user_message", intent.UserMessage)
	putString(out, "task_kind", string(intent.TaskKind))
	putString(out, "entity_name", intent.EntityName)
	putString(out, "stock_code", intent.StockCode)
	putString(out, "market", string(intent.Market))
	if len(intent.EntityMentions) > 0 {
		out["entity_mentions"] = append([]string(nil), intent.EntityMentions...)
	}
	if len(intent.RequestedFields) > 0 {
		out["requested_fields"] = append([]string(nil), intent.RequestedFields...)
	}
	if len(intent.RequestedOutputs) > 0 {
		out["requested_outputs"] = append([]string(nil), intent.RequestedOutputs...)
	}
	if intent.SourcePolicy != "" {
		out["source_policy"] = intent.SourcePolicy
	}
	return out
}

func StringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func StringSliceArg(value any) []string {
	switch typed := value.(type) {
	case []string:
		return compactStrings(typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, StringArg(item))
		}
		return compactStrings(items)
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		parts := strings.FieldsFunc(typed, func(r rune) bool {
			return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n' || r == '\t'
		})
		return compactStrings(parts)
	default:
		return nil
	}
}

func FreshnessArg(value any) astockcontracts.Freshness {
	raw, ok := value.(map[string]any)
	if !ok {
		return astockcontracts.Freshness{}
	}
	return astockcontracts.Freshness{
		Mode:                    astockcontracts.FreshnessMode(StringArg(raw["mode"])),
		RelativeDateHint:        StringArg(raw["relative_date_hint"]),
		TradeDate:               StringArg(raw["trade_date"]),
		AsOf:                    StringArg(raw["as_of"]),
		RequireRealtime:         BoolArg(raw["require_realtime"]),
		RequireLatestTradingDay: BoolArg(raw["require_latest_trading_day"]),
	}
}

func BoolArg(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || strings.TrimSpace(typed) == "1"
	default:
		return false
	}
}

func putString(out map[string]any, key string, value string) {
	if value != "" {
		out[key] = value
	}
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
