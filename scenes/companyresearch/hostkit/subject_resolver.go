package hostkit

import (
	"context"
	"strings"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

type SubjectResolver func(context.Context, research.SubjectResolutionRequest) (research.SubjectResolution, error)

func resolveSubject(ctx context.Context, resolver SubjectResolver, intent research.CompanyResearchIntent) (*research.SubjectResolution, []string) {
	if resolver == nil {
		return nil, nil
	}
	resolution, err := resolver(ctx, research.SubjectResolutionRequestFromIntent(intent))
	if err != nil {
		return &research.SubjectResolution{
			AdapterStatus: "error",
			FailureCode:   "subject_resolution_error",
			InputTerm:     firstNonEmpty(intent.EntityName, firstString(intent.EntityMentions), intent.UserMessage),
			Warnings:      []string{err.Error()},
		}, []string{"subject_resolution_error"}
	}
	resolution = normalizeSubjectResolution(resolution, intent)
	if subjectResolutionEmpty(resolution) {
		return nil, nil
	}
	if resolution.FailureCode != "" {
		return &resolution, []string{subjectResolutionWarningCode(resolution.FailureCode)}
	}
	return &resolution, nil
}

func subjectResolutionWarningCode(failureCode string) string {
	failureCode = strings.TrimSpace(failureCode)
	if failureCode == "" {
		return ""
	}
	if strings.HasPrefix(failureCode, "subject_resolution_") {
		return failureCode
	}
	return "subject_resolution_" + failureCode
}

func normalizeSubjectResolution(resolution research.SubjectResolution, intent research.CompanyResearchIntent) research.SubjectResolution {
	resolution.AdapterStatus = strings.TrimSpace(resolution.AdapterStatus)
	if resolution.AdapterStatus == "" {
		if candidate, ok := verifiedSubjectCandidate(&resolution); ok && candidate != nil {
			resolution.AdapterStatus = "ok"
		} else {
			resolution.AdapterStatus = "unknown"
		}
	}
	resolution.InputTerm = firstNonEmpty(resolution.InputTerm, intent.EntityName, firstString(intent.EntityMentions), intent.UserMessage)
	resolution.PreferredMarket = firstNonEmpty(normalizeFinanceMarketHint(resolution.PreferredMarket), normalizeFinanceMarketHint(intent.MarketHint))
	if resolution.SelectedCandidate != nil {
		normalized := normalizeSubjectCandidate(*resolution.SelectedCandidate)
		resolution.SelectedCandidate = &normalized
	}
	for i := range resolution.Candidates {
		resolution.Candidates[i] = normalizeSubjectCandidate(resolution.Candidates[i])
	}
	return resolution
}

func normalizeSubjectCandidate(candidate research.SubjectResolutionCandidate) research.SubjectResolutionCandidate {
	candidate.EntityName = strings.TrimSpace(candidate.EntityName)
	candidate.DisplayName = strings.TrimSpace(candidate.DisplayName)
	candidate.StockCode = normalizeStockCodeHint(candidate.StockCode)
	candidate.Ticker = strings.TrimSpace(candidate.Ticker)
	candidate.Market = normalizeFinanceMarketHint(candidate.Market)
	candidate.Exchange = strings.TrimSpace(candidate.Exchange)
	candidate.Source = strings.TrimSpace(candidate.Source)
	candidate.EvidenceURL = strings.TrimSpace(candidate.EvidenceURL)
	candidate.MatchReason = strings.TrimSpace(candidate.MatchReason)
	candidate.MismatchReason = strings.TrimSpace(candidate.MismatchReason)
	return candidate
}

func subjectResolutionEmpty(resolution research.SubjectResolution) bool {
	return strings.TrimSpace(resolution.AdapterStatus) == "" &&
		strings.TrimSpace(resolution.FailureCode) == "" &&
		strings.TrimSpace(resolution.InputTerm) == "" &&
		resolution.SelectedCandidate == nil &&
		len(resolution.Candidates) == 0 &&
		len(resolution.QueryVariants) == 0 &&
		len(resolution.Warnings) == 0
}

func verifiedSubjectCandidate(resolution *research.SubjectResolution) (*research.SubjectResolutionCandidate, bool) {
	if resolution == nil || resolution.SelectedCandidate == nil {
		return nil, false
	}
	candidate := normalizeSubjectCandidate(*resolution.SelectedCandidate)
	if !candidate.Verified {
		return nil, false
	}
	if strings.TrimSpace(candidate.MismatchReason) != "" {
		return nil, false
	}
	if candidate.EntityName == "" && candidate.DisplayName == "" && candidate.StockCode == "" && candidate.Ticker == "" {
		return nil, false
	}
	return &candidate, true
}

func intentWithSubjectResolution(intent research.CompanyResearchIntent, resolution *research.SubjectResolution) research.CompanyResearchIntent {
	candidate, ok := verifiedSubjectCandidate(resolution)
	if !ok {
		return intent
	}
	originalName := intent.EntityName
	canonicalName := firstNonEmpty(candidate.EntityName, candidate.DisplayName, intent.EntityName)
	if canonicalName != "" {
		intent.EntityName = canonicalName
	}
	mention := firstNonEmpty(canonicalName, originalName, firstString(intent.EntityMentions))
	if mention != "" {
		intent.EntityMentions = []string{mention}
	}
	if candidate.Market != "" {
		intent.MarketHint = candidate.Market
	}
	return intent
}

func subjectIdentityCheckIntent(original research.CompanyResearchIntent, downstream research.CompanyResearchIntent, resolution *research.SubjectResolution) research.CompanyResearchIntent {
	out := downstream
	mentions := []string{}
	add := func(values ...string) {
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				mentions = append(mentions, value)
			}
		}
	}
	add(out.EntityName)
	add(out.EntityMentions...)
	add(original.EntityName)
	add(original.EntityMentions...)
	if resolution != nil {
		add(resolution.InputTerm)
		if candidate, ok := verifiedSubjectCandidate(resolution); ok && candidate != nil {
			add(candidate.EntityName, candidate.DisplayName, candidate.StockCode, candidate.Ticker)
		}
	}
	out.EntityMentions = cleanStrings(mentions)
	return out
}

func downstreamParamsWithSubjectResolution(intent research.CompanyResearchIntent, target string, resolution *research.SubjectResolution) map[string]any {
	out := downstreamParams(intentWithSubjectResolution(intent, resolution), target)
	applySubjectResolutionParams(out, resolution)
	return out
}

func downstreamParamsWithSubjectAndFinanceIdentity(intent research.CompanyResearchIntent, target string, resolution *research.SubjectResolution, finance map[string]any) map[string]any {
	out := downstreamParamsWithSubjectResolution(intent, target, resolution)
	applyFinanceIdentityParams(out, finance, false)
	return out
}

func downstreamNewsParamsWithSubjectAndFinanceIdentity(intent research.CompanyResearchIntent, resolution *research.SubjectResolution, finance map[string]any) map[string]any {
	out := downstreamParamsWithSubjectResolution(intent, "news", resolution)
	applyFinanceIdentityParams(out, finance, true)
	return out
}

func applySubjectResolutionParams(out map[string]any, resolution *research.SubjectResolution) {
	candidate, ok := verifiedSubjectCandidate(resolution)
	if !ok {
		return
	}
	if name := firstNonEmpty(candidate.EntityName, candidate.DisplayName); name != "" {
		out["entity_name"] = name
	}
	if name := firstNonEmpty(research.StringArg(out["entity_name"]), candidate.EntityName, candidate.DisplayName); name != "" {
		out["entity_mentions"] = []string{name}
	}
	if code := firstNonEmpty(candidate.StockCode, candidate.Ticker); code != "" {
		out["stock_code"] = code
	}
	if ticker := firstNonEmpty(candidate.Ticker, candidate.StockCode); ticker != "" {
		out["ticker"] = ticker
	}
	if candidate.Market != "" {
		out["market_hint"] = candidate.Market
		out["market"] = candidate.Market
	}
}

func applyFinanceIdentityParams(out map[string]any, finance map[string]any, includeEntityMentions bool) {
	hint := financeResolvedIdentityHint(finance)
	if hint.StockCode != "" {
		out["stock_code"] = hint.StockCode
		out["ticker"] = hint.StockCode
	}
	if hint.MarketHint != "" {
		out["market_hint"] = hint.MarketHint
		out["market"] = hint.MarketHint
	}
	if hint.CompanyName != "" {
		out["entity_name"] = hint.CompanyName
		if includeEntityMentions {
			out["entity_mentions"] = []string{hint.CompanyName}
		}
	}
}

func firstString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
