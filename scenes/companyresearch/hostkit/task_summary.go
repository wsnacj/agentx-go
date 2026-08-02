package hostkit

import (
	"strings"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

type downstreamIdentityHint struct {
	Role        research.CompanyResearchTaskRole
	Source      string
	CompanyName string
	StockCode   string
	Ticker      string
	MarketHint  string
	EvidenceURL string
}

func applyDownstreamSubjectConfirmation(payload *research.CompanyResearchPayload, plan *research.CompanyResearchTaskPlan) {
	if payload == nil || plan == nil || plan.Subject.Verified {
		return
	}
	hint, sourceRoles, ok := downstreamSubjectConfirmation(payload.Intent, payload.Evidence)
	if !ok {
		return
	}
	plan.Subject.CanonicalName = firstNonEmpty(hint.CompanyName, plan.Subject.CanonicalName)
	plan.Subject.StockCode = firstNonEmpty(hint.StockCode, plan.Subject.StockCode)
	plan.Subject.Ticker = firstNonEmpty(hint.Ticker, plan.Subject.Ticker, hint.StockCode)
	plan.Subject.MarketHint = firstNonEmpty(hint.MarketHint, plan.Subject.MarketHint)
	plan.Subject.Source = firstNonEmpty(hint.Source, "downstream_evidence")
	plan.Subject.EvidenceURL = firstNonEmpty(hint.EvidenceURL, plan.Subject.EvidenceURL)
	plan.Subject.Verified = true
	plan.Subject.Confidence = 0.8
	plan.Subject.FailureCode = ""
	plan.Subject.ResolutionState = "confirmed_by_downstream_evidence"
	plan.Warnings = removeTaskSummaryString(plan.Warnings, "subject_not_verified_by_task_plan")
	payload.TaskPlan = plan
	if result, ok := research.SubjectResolutionTaskResult(*plan); ok {
		result.Summary = "subject confirmed by downstream evidence"
		result.Diagnostics = map[string]string{
			"identity_source":       "downstream_evidence",
			"identity_source_roles": strings.Join(sourceRoles, ","),
		}
		upsertTaskResult(payload, result)
	}
}

func downstreamSubjectConfirmation(intent research.CompanyResearchIntent, evidence research.CompanyResearchEvidence) (downstreamIdentityHint, []string, bool) {
	hints := matchingDownstreamIdentityHints(intent, evidence)
	if len(hints) == 0 || downstreamIdentityHintsConflict(hints) {
		return downstreamIdentityHint{}, nil, false
	}
	merged := mergeDownstreamIdentityHints(hints)
	roles := make([]string, 0, len(hints))
	seen := map[string]bool{}
	for _, hint := range hints {
		role := string(hint.Role)
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		roles = append(roles, role)
	}
	return merged, roles, true
}

func matchingDownstreamIdentityHints(intent research.CompanyResearchIntent, evidence research.CompanyResearchEvidence) []downstreamIdentityHint {
	out := []downstreamIdentityHint{}
	for _, hint := range verifiedDownstreamIdentityHints(evidence) {
		if identityHintMatchesIntent(intent, hint) {
			out = append(out, hint)
		}
	}
	return out
}

func verifiedDownstreamIdentityHints(evidence research.CompanyResearchEvidence) []downstreamIdentityHint {
	out := []downstreamIdentityHint{}
	if hint, ok := strictFinanceIdentityHint(evidence.Finance); ok {
		out = append(out, hint)
	}
	if hint, ok := strictMarketIdentityHint(research.CompanyResearchRoleMarketAnalyst, evidence.AStock); ok {
		out = append(out, hint)
	}
	if hint, ok := strictMarketIdentityHint(research.CompanyResearchRoleMarketAnalyst, evidence.GlobalStock); ok {
		out = append(out, hint)
	}
	return out
}

func strictFinanceIdentityHint(finance map[string]any) (downstreamIdentityHint, bool) {
	if !financeEvidenceReady(finance) {
		return downstreamIdentityHint{}, false
	}
	rawCode := firstNonEmpty(
		deepString(finance, "candidates", "resolved_code"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "code_or_ticker"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "stock_code"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "ticker"),
		deepString(finance, "metrics", "stock_code"),
		deepString(finance, "metrics", "evidence", "stock_code"),
	)
	companyName := financeResolvedCompanyName(finance)
	marketHint := normalizeFinanceMarketHint(firstNonEmpty(
		deepString(finance, "candidates", "resolved_market"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "market"),
		deepString(finance, "metrics", "evidence", "market"),
		marketHintFromCode(rawCode),
	))
	hint := downstreamIdentityHint{
		Role:        research.CompanyResearchRoleFinanceAnalyst,
		Source:      firstNonEmpty(deepString(finance, "tool"), deepString(finance, "source"), "finance"),
		CompanyName: companyName,
		StockCode:   normalizeStockCodeHint(rawCode),
		Ticker:      strings.TrimSpace(rawCode),
		MarketHint:  marketHint,
		EvidenceURL: firstNonEmpty(deepString(finance, "evidence_url"), deepString(finance, "brief", "evidence", "source_url"), deepString(finance, "metrics", "evidence", "source_url")),
	}
	if hint.CompanyName == "" && hint.StockCode == "" && hint.Ticker == "" {
		return downstreamIdentityHint{}, false
	}
	return hint, true
}

func strictMarketIdentityHint(role research.CompanyResearchTaskRole, evidence map[string]any) (downstreamIdentityHint, bool) {
	if !research.EvidencePayloadReady(evidence) {
		return downstreamIdentityHint{}, false
	}
	quoteEvidence := primaryMarketQuoteEvidence(evidence)
	rawCode := firstNonEmpty(
		deepString(evidence, "quote", "subject", "ticker"),
		deepString(evidence, "quote", "subject", "stock_code"),
		deepString(evidence, "quote", "subject", "code"),
		deepString(evidence, "subject", "ticker"),
		deepString(evidence, "subject", "stock_code"),
		deepString(evidence, "subject", "code"),
		deepString(evidence, "quote", "symbol"),
		deepString(evidence, "ticker"),
		deepString(evidence, "stock_code"),
		deepString(quoteEvidence, "subject", "ticker"),
		deepString(quoteEvidence, "subject", "stock_code"),
		deepString(quoteEvidence, "subject", "code"),
		deepString(quoteEvidence, "identity_resolution", "selected_candidate", "code"),
		deepString(quoteEvidence, "identity_resolution", "selected_candidate", "ticker"),
	)
	companyName := firstNonEmpty(
		deepString(evidence, "quote", "subject", "entity_name"),
		deepString(evidence, "quote", "subject", "display_name"),
		deepString(evidence, "subject", "entity_name"),
		deepString(evidence, "subject", "display_name"),
		deepString(evidence, "profile", "company_name"),
		deepString(quoteEvidence, "subject", "entity_name"),
		deepString(quoteEvidence, "subject", "display_name"),
		deepString(quoteEvidence, "identity_resolution", "selected_candidate", "name"),
	)
	marketHint := normalizeFinanceMarketHint(firstNonEmpty(
		deepString(evidence, "quote", "subject", "market"),
		deepString(evidence, "quote", "subject", "exchange"),
		deepString(evidence, "subject", "market"),
		deepString(evidence, "subject", "exchange"),
		deepString(evidence, "quote", "market"),
		deepString(evidence, "market"),
		deepString(evidence, "market_hint"),
		deepString(quoteEvidence, "subject", "market"),
		deepString(quoteEvidence, "subject", "exchange"),
		deepString(quoteEvidence, "identity_resolution", "selected_candidate", "market"),
		deepString(quoteEvidence, "identity_resolution", "selected_candidate", "exchange"),
		marketHintFromCode(rawCode),
	))
	hint := downstreamIdentityHint{
		Role:        role,
		Source:      firstNonEmpty(deepString(evidence, "tool"), deepString(evidence, "source"), string(role)),
		CompanyName: companyName,
		StockCode:   normalizeStockCodeHint(rawCode),
		Ticker:      strings.TrimSpace(rawCode),
		MarketHint:  marketHint,
		EvidenceURL: firstNonEmpty(deepString(evidence, "evidence_url"), deepString(evidence, "quote", "evidence_url"), deepString(evidence, "quote", "source_url"), deepString(quoteEvidence, "evidence", "source_url"), deepString(quoteEvidence, "source_url")),
	}
	if hint.CompanyName == "" && hint.StockCode == "" && hint.Ticker == "" {
		return downstreamIdentityHint{}, false
	}
	return hint, true
}

func identityHintMatchesIntent(intent research.CompanyResearchIntent, hint downstreamIdentityHint) bool {
	if !marketHintsOverlap(intent.MarketHint, hint.MarketHint) {
		return false
	}
	if hint.CompanyName != "" {
		if companyNameMatchesIntent(intent, hint.CompanyName) {
			return true
		}
	}
	return identityCodeMatchesIntent(intent, hint.StockCode, hint.Ticker)
}

func identityCodeMatchesIntent(intent research.CompanyResearchIntent, codes ...string) bool {
	candidates := append([]string{}, intent.EntityName)
	candidates = append(candidates, intent.EntityMentions...)
	expected := map[string]bool{}
	for _, candidate := range candidates {
		if normalized := normalizeIdentityCode(candidate); normalized != "" {
			expected[normalized] = true
		}
	}
	if len(expected) == 0 {
		return false
	}
	for _, code := range codes {
		if expected[normalizeIdentityCode(code)] {
			return true
		}
	}
	return false
}

func downstreamIdentityHintsConflict(hints []downstreamIdentityHint) bool {
	for i := 0; i < len(hints); i++ {
		for j := i + 1; j < len(hints); j++ {
			if identityHintsConflict(hints[i], hints[j]) {
				return true
			}
		}
	}
	return false
}

func identityHintsConflict(left, right downstreamIdentityHint) bool {
	if identityHintsSameSecurity(left, right) {
		return false
	}
	if left.CompanyName != "" && right.CompanyName != "" && !companyNamesCompatible(left.CompanyName, right.CompanyName) {
		return true
	}
	leftCode := normalizeIdentityCodeForMarket(firstNonEmpty(left.StockCode, left.Ticker), left.MarketHint)
	rightCode := normalizeIdentityCodeForMarket(firstNonEmpty(right.StockCode, right.Ticker), right.MarketHint)
	if leftCode == "" || rightCode == "" || leftCode == rightCode {
		return false
	}
	leftMarket := strings.TrimSpace(left.MarketHint)
	rightMarket := strings.TrimSpace(right.MarketHint)
	if leftMarket == "" || rightMarket == "" {
		return true
	}
	return marketHintsOverlap(leftMarket, rightMarket)
}

func identityHintsSameSecurity(left, right downstreamIdentityHint) bool {
	leftCode := normalizeIdentityCodeForMarket(firstNonEmpty(left.StockCode, left.Ticker), left.MarketHint)
	rightCode := normalizeIdentityCodeForMarket(firstNonEmpty(right.StockCode, right.Ticker), right.MarketHint)
	if leftCode == "" || rightCode == "" || leftCode != rightCode {
		return false
	}
	leftMarket := strings.TrimSpace(left.MarketHint)
	rightMarket := strings.TrimSpace(right.MarketHint)
	if leftMarket == "" || rightMarket == "" {
		return true
	}
	return marketHintsOverlap(leftMarket, rightMarket)
}

func mergeDownstreamIdentityHints(hints []downstreamIdentityHint) downstreamIdentityHint {
	merged := downstreamIdentityHint{}
	sources := []string{}
	for _, hint := range hints {
		merged.CompanyName = firstNonEmpty(merged.CompanyName, hint.CompanyName)
		merged.StockCode = firstNonEmpty(merged.StockCode, hint.StockCode)
		merged.Ticker = firstNonEmpty(merged.Ticker, hint.Ticker, hint.StockCode)
		merged.MarketHint = firstNonEmpty(merged.MarketHint, hint.MarketHint)
		merged.EvidenceURL = firstNonEmpty(merged.EvidenceURL, hint.EvidenceURL)
		if hint.Source != "" {
			sources = append(sources, hint.Source)
		}
	}
	merged.Source = strings.Join(cleanStrings(sources), ",")
	return merged
}

func buildCompanyResearchTaskSummary(payload research.CompanyResearchPayload) *research.CompanyResearchTaskSummary {
	summary := &research.CompanyResearchTaskSummary{
		ReadyDimensions:   cleanStrings(payload.AnswerReadiness.ReadyDimensions),
		MissingDimensions: cleanStrings(payload.AnswerReadiness.MissingDimensions),
		Warnings:          cleanStrings(payload.Warnings),
	}
	roleSeen := map[research.CompanyResearchTaskRole]research.CompanyResearchTaskStatus{}
	for _, result := range payload.TaskResults {
		if result.Role == "" {
			continue
		}
		roleSeen[result.Role] = result.Status
	}
	for _, role := range taskSummaryRoleOrder() {
		status, ok := roleSeen[role]
		if !ok {
			continue
		}
		switch status {
		case research.CompanyResearchTaskStatusReady:
			summary.ReadyRoles = append(summary.ReadyRoles, role)
		case research.CompanyResearchTaskStatusDegraded:
			summary.DegradedRoles = append(summary.DegradedRoles, role)
		case research.CompanyResearchTaskStatusFailed:
			summary.FailedRoles = append(summary.FailedRoles, role)
		case research.CompanyResearchTaskStatusSkipped:
			summary.SkippedRoles = append(summary.SkippedRoles, role)
		}
	}
	summary.Conflicts = append(summary.Conflicts, taskConflicts(payload)...)
	for _, subject := range payload.Subjects {
		if subject.TaskSummary == nil {
			continue
		}
		name := firstNonEmpty(subject.Intent.EntityName, taskPlanSubjectName(subject), "unknown_subject")
		for _, conflict := range subject.TaskSummary.Conflicts {
			if conflict.Subject == "" {
				conflict.Subject = name
			}
			summary.Conflicts = append(summary.Conflicts, conflict)
		}
	}
	if len(summary.ReadyRoles) == 0 &&
		len(summary.DegradedRoles) == 0 &&
		len(summary.FailedRoles) == 0 &&
		len(summary.SkippedRoles) == 0 &&
		len(summary.ReadyDimensions) == 0 &&
		len(summary.MissingDimensions) == 0 &&
		len(summary.Conflicts) == 0 &&
		len(summary.Warnings) == 0 {
		return nil
	}
	return summary
}

func taskSummaryRoleOrder() []research.CompanyResearchTaskRole {
	return []research.CompanyResearchTaskRole{
		research.CompanyResearchRoleSubjectResolver,
		research.CompanyResearchRoleFinanceAnalyst,
		research.CompanyResearchRoleMarketAnalyst,
		research.CompanyResearchRoleNewsAnalyst,
		research.CompanyResearchRoleRiskReviewer,
		research.CompanyResearchRoleEvidenceGuard,
		research.CompanyResearchRoleSynthesisEditor,
	}
}

func taskConflicts(payload research.CompanyResearchPayload) []research.CompanyResearchTaskConflict {
	conflicts := []research.CompanyResearchTaskConflict{}
	identityIntent := subjectIdentityCheckIntent(payload.Intent, intentWithSubjectResolution(payload.Intent, payload.SubjectResolution), payload.SubjectResolution)
	hints := verifiedDownstreamIdentityHints(payload.Evidence)
	for _, hint := range hints {
		if identityIntent.MarketHint != "" && hint.MarketHint != "" && !marketHintsOverlap(identityIntent.MarketHint, hint.MarketHint) {
			conflicts = append(conflicts, research.CompanyResearchTaskConflict{
				Code:      "market_mismatch",
				Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
				Role:      hint.Role,
				Dimension: dimensionForIdentityRole(hint.Role),
				Expected:  identityIntent.MarketHint,
				Observed:  identityHintLabel(hint),
				Summary:   "downstream evidence market does not match requested market hint",
			})
			continue
		}
		if !identityHintMatchesIntent(identityIntent, hint) {
			conflicts = append(conflicts, research.CompanyResearchTaskConflict{
				Code:      "subject_identity_mismatch",
				Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
				Role:      hint.Role,
				Dimension: dimensionForIdentityRole(hint.Role),
				Expected:  intentIdentityLabel(identityIntent),
				Observed:  identityHintLabel(hint),
				Summary:   "downstream evidence identity does not match the requested subject",
			})
		}
	}
	for i := 0; i < len(hints); i++ {
		for j := i + 1; j < len(hints); j++ {
			left := hints[i]
			right := hints[j]
			if left.CompanyName != "" && right.CompanyName != "" && !identityHintsSameSecurity(left, right) && !companyNamesCompatible(left.CompanyName, right.CompanyName) {
				conflicts = append(conflicts, research.CompanyResearchTaskConflict{
					Code:      "cross_task_subject_mismatch",
					Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
					Role:      left.Role,
					OtherRole: right.Role,
					Dimension: "identity",
					Expected:  identityHintLabel(left),
					Observed:  identityHintLabel(right),
					Summary:   "downstream tasks resolved different company identities",
				})
			}
			if identityCodeConflict(left, right) {
				conflicts = append(conflicts, research.CompanyResearchTaskConflict{
					Code:      "cross_task_code_mismatch",
					Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
					Role:      left.Role,
					OtherRole: right.Role,
					Dimension: "identity",
					Expected:  identityHintLabel(left),
					Observed:  identityHintLabel(right),
					Summary:   "downstream tasks resolved incompatible security codes for the same market scope",
				})
			}
		}
	}
	conflicts = append(conflicts, freshnessConflicts(payload)...)
	conflicts = append(conflicts, sourceFactConflicts(payload)...)
	return conflicts
}

func freshnessConflicts(payload research.CompanyResearchPayload) []research.CompanyResearchTaskConflict {
	if !structuredFreshnessRequired(payload.Intent) {
		return nil
	}
	conflicts := []research.CompanyResearchTaskConflict{}
	if evidenceFreshnessRejected(payload.Evidence.Finance,
		[]string{"metrics", "evaluation", "period_latest"},
		[]string{"brief", "evaluation", "period_latest"},
		[]string{"freshness_confirmed"},
	) {
		conflicts = append(conflicts, research.CompanyResearchTaskConflict{
			Code:      "freshness_not_confirmed",
			Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
			Role:      research.CompanyResearchRoleFinanceAnalyst,
			Dimension: "financials",
			Expected:  "latest_or_recent",
			Observed:  "finance evidence freshness not confirmed",
			Summary:   "downstream finance evidence did not confirm requested freshness",
		})
	}
	if evidenceFreshnessRejected(payload.Evidence.AStock,
		[]string{"quote", "freshness", "confirmed"},
		[]string{"freshness_confirmed"},
	) || evidenceFreshnessRejected(payload.Evidence.GlobalStock,
		[]string{"quote", "freshness", "confirmed"},
		[]string{"freshness_confirmed"},
	) {
		conflicts = append(conflicts, research.CompanyResearchTaskConflict{
			Code:      "freshness_not_confirmed",
			Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
			Role:      research.CompanyResearchRoleMarketAnalyst,
			Dimension: "market_data",
			Expected:  "latest_or_recent",
			Observed:  "market evidence freshness not confirmed",
			Summary:   "downstream market evidence did not confirm requested freshness",
		})
	}
	if evidenceFreshnessRejected(payload.Evidence.News,
		[]string{"freshness_confirmed"},
		[]string{"evaluator_report", "freshness_confirmed"},
		[]string{"guard", "freshness_confirmed"},
	) {
		conflicts = append(conflicts, research.CompanyResearchTaskConflict{
			Code:      "freshness_not_confirmed",
			Subject:   firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage),
			Role:      research.CompanyResearchRoleNewsAnalyst,
			Dimension: "news",
			Expected:  "latest_or_recent",
			Observed:  "news evidence freshness not confirmed",
			Summary:   "downstream news evidence did not confirm requested freshness",
		})
	}
	return conflicts
}

func sourceFactConflicts(payload research.CompanyResearchPayload) []research.CompanyResearchTaskConflict {
	conflicts := []research.CompanyResearchTaskConflict{}
	subject := firstNonEmpty(payload.Intent.EntityName, payload.Intent.UserMessage)
	if evidenceSourceRejected(payload.Evidence.Finance,
		[]string{"metrics", "evaluation", "source_accepted"},
		[]string{"brief", "evaluation", "source_accepted"},
		[]string{"source_accepted"},
	) {
		conflicts = append(conflicts, research.CompanyResearchTaskConflict{
			Code:      "source_not_accepted",
			Subject:   subject,
			Role:      research.CompanyResearchRoleFinanceAnalyst,
			Dimension: "financials",
			Expected:  "accepted_source",
			Observed:  "finance evidence source not accepted",
			Summary:   "downstream finance evidence source was not accepted",
		})
	}
	if evidenceSourceRejected(payload.Evidence.News,
		[]string{"source_accepted"},
		[]string{"evaluator_report", "source_accepted"},
		[]string{"guard", "source_accepted"},
	) {
		conflicts = append(conflicts, research.CompanyResearchTaskConflict{
			Code:      "source_not_accepted",
			Subject:   subject,
			Role:      research.CompanyResearchRoleNewsAnalyst,
			Dimension: "news",
			Expected:  "accepted_source",
			Observed:  "news evidence source not accepted",
			Summary:   "downstream news evidence source was not accepted",
		})
	}
	conflicts = append(conflicts, financeSourceFactConflicts(payload.Evidence.Finance, subject)...)
	if conflict, ok := newsSourceFactConflict(payload.Evidence.News, subject); ok {
		conflicts = append(conflicts, conflict)
	}
	return conflicts
}

func identityCodeConflict(left, right downstreamIdentityHint) bool {
	leftCode := normalizeIdentityCodeForMarket(firstNonEmpty(left.StockCode, left.Ticker), left.MarketHint)
	rightCode := normalizeIdentityCodeForMarket(firstNonEmpty(right.StockCode, right.Ticker), right.MarketHint)
	if leftCode == "" || rightCode == "" || leftCode == rightCode {
		return false
	}
	leftMarket := strings.TrimSpace(left.MarketHint)
	rightMarket := strings.TrimSpace(right.MarketHint)
	if leftMarket == "" || rightMarket == "" {
		return true
	}
	return marketHintsOverlap(leftMarket, rightMarket)
}

func structuredFreshnessRequired(intent research.CompanyResearchIntent) bool {
	if boolMapValue(intent.Freshness, "require_latest") {
		return true
	}
	mode := strings.ToLower(strings.TrimSpace(research.StringArg(intent.Freshness["mode"])))
	switch mode {
	case "latest", "recent", "current", "realtime":
		return true
	default:
		return false
	}
}

func evidenceFreshnessRejected(evidence map[string]any, paths ...[]string) bool {
	if !research.EvidencePayloadReady(evidence) {
		return false
	}
	return anyExplicitFalse(evidence, paths...)
}

func evidenceSourceRejected(evidence map[string]any, paths ...[]string) bool {
	if !research.EvidencePayloadReady(evidence) {
		return false
	}
	return anyExplicitFalse(evidence, paths...)
}

func anyExplicitFalse(object map[string]any, paths ...[]string) bool {
	for _, path := range paths {
		if value, ok := deepBool(object, path...); ok && !value {
			return true
		}
	}
	return false
}

func deepBool(object map[string]any, path ...string) (bool, bool) {
	var current any = object
	for _, key := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return false, false
		}
		current = next[key]
	}
	value, ok := current.(bool)
	return value, ok
}

func boolMapValue(object map[string]any, key string) bool {
	if object == nil {
		return false
	}
	value, _ := object[key].(bool)
	return value
}

func financeSourceFactConflicts(finance map[string]any, subject string) []research.CompanyResearchTaskConflict {
	if !financeEvidenceReady(finance) {
		return nil
	}
	valuesByField := map[string][]sourceFactValue{}
	addFact := func(field, value, period, source string) {
		field = strings.TrimSpace(field)
		value = strings.TrimSpace(value)
		if field == "" || value == "" {
			return
		}
		valuesByField[field] = append(valuesByField[field], sourceFactValue{
			Field:  field,
			Value:  value,
			Period: strings.TrimSpace(period),
			Source: strings.TrimSpace(source),
		})
	}
	for _, fact := range contractObjectListAt(finance, "assessment_projection", "verified_facts") {
		addFact(
			research.StringArg(fact["field"]),
			research.StringArg(fact["value"]),
			research.StringArg(fact["period"]),
			research.StringArg(fact["source"]),
		)
	}
	metricEvidence := contractMapAt(finance, "metrics", "evidence", "metric_evidence")
	for field, raw := range metricEvidence {
		object, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		addFact(
			field,
			research.StringArg(object["value"]),
			research.StringArg(object["period"]),
			research.StringArg(object["source"]),
		)
	}
	for _, metric := range contractObjectListAt(finance, "brief", "evidence", "metrics") {
		addFact(
			research.StringArg(metric["name"]),
			research.StringArg(metric["value"]),
			"",
			research.StringArg(metric["source"]),
		)
	}
	conflicts := []research.CompanyResearchTaskConflict{}
	for field, values := range valuesByField {
		if expected, observed, ok := sourceFactMismatch(values); ok {
			conflicts = append(conflicts, research.CompanyResearchTaskConflict{
				Code:      "source_fact_mismatch",
				Subject:   subject,
				Role:      research.CompanyResearchRoleFinanceAnalyst,
				Dimension: "financials",
				Expected:  expected.Label(),
				Observed:  observed.Label(),
				Summary:   "finance source-backed values disagree for field " + field,
			})
		}
	}
	return conflicts
}

func newsSourceFactConflict(news map[string]any, subject string) (research.CompanyResearchTaskConflict, bool) {
	if !research.EvidencePayloadReady(news) {
		return research.CompanyResearchTaskConflict{}, false
	}
	topPublishedAt := firstNonEmpty(
		deepString(news, "published_at"),
		deepString(news, "extract", "published_at"),
	)
	sourcePublishedAt := firstNonEmpty(
		deepString(news, "sources", "primary_source", "published_at"),
		deepString(news, "primary_source", "published_at"),
	)
	if topPublishedAt == "" || sourcePublishedAt == "" || normalizeSourceFactValue(topPublishedAt) == normalizeSourceFactValue(sourcePublishedAt) {
		return research.CompanyResearchTaskConflict{}, false
	}
	return research.CompanyResearchTaskConflict{
		Code:      "source_fact_mismatch",
		Subject:   subject,
		Role:      research.CompanyResearchRoleNewsAnalyst,
		Dimension: "news",
		Expected:  topPublishedAt,
		Observed:  sourcePublishedAt,
		Summary:   "news source-backed published_at values disagree",
	}, true
}

type sourceFactValue struct {
	Field  string
	Value  string
	Period string
	Source string
}

func (value sourceFactValue) Label() string {
	parts := []string{value.Value}
	if value.Period != "" {
		parts = append(parts, value.Period)
	}
	if value.Source != "" {
		parts = append(parts, value.Source)
	}
	return strings.Join(parts, " @ ")
}

func sourceFactMismatch(values []sourceFactValue) (sourceFactValue, sourceFactValue, bool) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if normalizeSourceFactValue(values[i].Value) != normalizeSourceFactValue(values[j].Value) {
				return values[i], values[j], true
			}
		}
	}
	return sourceFactValue{}, sourceFactValue{}, false
}

func normalizeSourceFactValue(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), "")
	value = strings.NewReplacer(",", "", "，", "", " ", "", "\t", "", "\n", "", "\r", "").Replace(value)
	return value
}

func companyNamesCompatible(left, right string) bool {
	leftKey := normalizeSubjectIdentity(left)
	rightKey := normalizeSubjectIdentity(right)
	if leftKey == "" || rightKey == "" {
		return true
	}
	if leftKey == rightKey || strings.Contains(leftKey, rightKey) || strings.Contains(rightKey, leftKey) {
		return true
	}
	leftPrefix := normalizeSubjectIdentity(cjkSubjectPrefix(left))
	rightPrefix := normalizeSubjectIdentity(cjkSubjectPrefix(right))
	return leftPrefix != "" && rightPrefix != "" && leftPrefix == rightPrefix
}

func normalizeIdentityCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = normalizeStockCodeHint(value)
	var builder strings.Builder
	for _, r := range strings.ToUpper(value) {
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func normalizeIdentityCodeForMarket(value string, market string) string {
	normalized := normalizeIdentityCode(value)
	if normalized == "" {
		return ""
	}
	if normalizeFinanceMarketHint(market) != "hk" || !identityCodeAllDigits(normalized) {
		return normalized
	}
	trimmed := strings.TrimLeft(normalized, "0")
	if trimmed == "" {
		return "0"
	}
	return trimmed
}

func identityCodeAllDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func identityHintLabel(hint downstreamIdentityHint) string {
	parts := []string{}
	if hint.CompanyName != "" {
		parts = append(parts, hint.CompanyName)
	}
	if code := firstNonEmpty(hint.Ticker, hint.StockCode); code != "" {
		parts = append(parts, code)
	}
	if hint.MarketHint != "" {
		parts = append(parts, hint.MarketHint)
	}
	if len(parts) == 0 {
		return string(hint.Role)
	}
	return strings.Join(parts, "/")
}

func intentIdentityLabel(intent research.CompanyResearchIntent) string {
	parts := []string{}
	if intent.EntityName != "" {
		parts = append(parts, intent.EntityName)
	}
	if intent.MarketHint != "" {
		parts = append(parts, intent.MarketHint)
	}
	return strings.Join(parts, "/")
}

func dimensionForIdentityRole(role research.CompanyResearchTaskRole) string {
	switch role {
	case research.CompanyResearchRoleFinanceAnalyst:
		return "financials"
	case research.CompanyResearchRoleMarketAnalyst:
		return "market_data"
	case research.CompanyResearchRoleNewsAnalyst:
		return "news"
	default:
		return "identity"
	}
}

func taskSummaryContractLines(payload research.CompanyResearchPayload) []string {
	if payload.TaskSummary == nil {
		return nil
	}
	lines := []string{}
	if len(payload.TaskSummary.Conflicts) > 0 {
		lines = append(lines, "任务冲突摘要：")
		for _, conflict := range payload.TaskSummary.Conflicts {
			text := firstNonEmpty(conflict.Summary, conflict.Code)
			if conflict.Expected != "" || conflict.Observed != "" {
				text += "（期望：" + firstNonEmpty(conflict.Expected, "未指定") + "；观测：" + firstNonEmpty(conflict.Observed, "未指定") + "）"
			}
			lines = append(lines, "- "+text)
		}
	}
	if len(payload.TaskSummary.DegradedRoles) > 0 {
		roleNames := make([]string, 0, len(payload.TaskSummary.DegradedRoles))
		for _, role := range payload.TaskSummary.DegradedRoles {
			roleNames = append(roleNames, string(role))
		}
		lines = append(lines, "未完全就绪的任务角色："+strings.Join(cleanStrings(roleNames), "、")+"。")
	}
	if len(payload.TaskSummary.FailedRoles) > 0 {
		roleNames := make([]string, 0, len(payload.TaskSummary.FailedRoles))
		for _, role := range payload.TaskSummary.FailedRoles {
			roleNames = append(roleNames, string(role))
		}
		lines = append(lines, "执行失败的任务角色："+strings.Join(cleanStrings(roleNames), "、")+"。")
	}
	return lines
}

func removeTaskSummaryString(values []string, target string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			continue
		}
		out = append(out, value)
	}
	return out
}

func taskPlanSubjectName(payload research.CompanyResearchPayload) string {
	if payload.TaskPlan == nil {
		return ""
	}
	return payload.TaskPlan.Subject.CanonicalName
}
