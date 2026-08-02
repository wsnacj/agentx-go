package companyresearch

import (
	"strings"
)

func IntentFromParams(params map[string]any) CompanyResearchIntent {
	intent := CompanyResearchIntent{
		UserMessage:         strings.TrimSpace(StringArg(params["user_message"])),
		TaskKind:            strings.TrimSpace(StringArg(params["task_kind"])),
		EntityName:          strings.TrimSpace(StringArg(params["entity_name"])),
		EntityMentions:      StringListArg(params["entity_mentions"]),
		MarketHint:          strings.TrimSpace(StringArg(params["market_hint"])),
		ComparisonSubjects:  CompanySubjectsArg(params["comparison_subjects"]),
		RequestedDimensions: NormalizeStrings(StringListArg(params["requested_dimensions"])),
		RequestedOutputs:    NormalizeStrings(StringListArg(params["requested_outputs"])),
		Freshness:           ObjectArg(params["freshness"]),
		RiskScope:           strings.TrimSpace(StringArg(params["risk_scope"])),
		SourcePolicy:        strings.TrimSpace(StringArg(params["source_policy"])),
		OriginalIntent:      strings.TrimSpace(StringArg(params["original_intent"])),
		StopCondition:       strings.TrimSpace(StringArg(params["stop_condition"])),
	}
	if intent.OriginalIntent == "" {
		intent.OriginalIntent = intent.UserMessage
	}
	if len(intent.RequestedDimensions) == 0 {
		intent.RequestedDimensions = []string{"financials", "market_data", "news", "risk"}
	}
	if len(intent.RequestedOutputs) == 0 {
		intent.RequestedOutputs = []string{"brief", "risk_summary", "investment_boundary"}
	}
	return intent
}

func ParamsFromIntent(intent CompanyResearchIntent) map[string]any {
	out := map[string]any{
		"user_message":         intent.UserMessage,
		"task_kind":            intent.TaskKind,
		"entity_name":          intent.EntityName,
		"entity_mentions":      intent.EntityMentions,
		"market_hint":          intent.MarketHint,
		"requested_dimensions": intent.RequestedDimensions,
		"requested_outputs":    intent.RequestedOutputs,
		"freshness":            intent.Freshness,
		"risk_scope":           intent.RiskScope,
		"source_policy":        intent.SourcePolicy,
		"original_intent":      intent.OriginalIntent,
		"stop_condition":       intent.StopCondition,
	}
	return out
}

func StringArg(raw any) string {
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func StringListArg(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		return NormalizeStrings(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := StringArg(item); text != "" {
				out = append(out, text)
			}
		}
		return NormalizeStrings(out)
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return nil
	}
}

func ObjectArg(raw any) map[string]any {
	value, _ := raw.(map[string]any)
	return value
}

func CompanySubjectsArg(raw any) []CompanySubject {
	switch typed := raw.(type) {
	case []CompanySubject:
		return typed
	case []any:
		out := make([]CompanySubject, 0, len(typed))
		for _, item := range typed {
			object, _ := item.(map[string]any)
			if len(object) == 0 {
				continue
			}
			subject := CompanySubject{
				EntityName:     StringArg(object["entity_name"]),
				EntityMentions: StringListArg(object["entity_mentions"]),
				MarketHint:     StringArg(object["market_hint"]),
			}
			if subject.EntityName != "" || len(subject.EntityMentions) > 0 {
				out = append(out, subject)
			}
		}
		return out
	default:
		return nil
	}
}

func NormalizeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ContainsToken(values []string, tokens ...string) bool {
	allowed := map[string]bool{}
	for _, token := range tokens {
		allowed[strings.ToLower(strings.TrimSpace(token))] = true
	}
	for _, value := range values {
		if allowed[strings.ToLower(strings.TrimSpace(value))] {
			return true
		}
	}
	return false
}
