package tools

import "strings"

type browserTopLevelSummary struct {
	Category              string                        `json:"category,omitempty"`
	State                 string                        `json:"state,omitempty"`
	SummaryCode           string                        `json:"summary_code,omitempty"`
	RepairCommand         string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	NextStepAlias         string                        `json:"next_step_alias,omitempty"`
	ManualRetryHint       string                        `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback   bool                          `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction  string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction     string                        `json:"primary_node_action,omitempty"`
	NextStep              string                        `json:"next_step,omitempty"`
}

type browserTopLevelDisplaySummary struct {
	Ready                 bool                          `json:"ready,omitempty"`
	Sections              []string                      `json:"sections,omitempty"`
	Category              string                        `json:"category,omitempty"`
	State                 string                        `json:"state,omitempty"`
	SummaryCode           string                        `json:"summary_code,omitempty"`
	RepairCommand         string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	NextStepAlias         string                        `json:"next_step_alias,omitempty"`
	ManualRetryHint       string                        `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback   bool                          `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction  string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction     string                        `json:"primary_node_action,omitempty"`
	NextStep              string                        `json:"next_step,omitempty"`
}

type browserTopLevelSurfaceSummary struct {
	Ready                 bool                          `json:"ready,omitempty"`
	Sections              []string                      `json:"sections,omitempty"`
	Category              string                        `json:"category,omitempty"`
	State                 string                        `json:"state,omitempty"`
	SummaryCode           string                        `json:"summary_code,omitempty"`
	RepairCommand         string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	NextStepAlias         string                        `json:"next_step_alias,omitempty"`
	ManualRetryHint       string                        `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback   bool                          `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction  string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction     string                        `json:"primary_node_action,omitempty"`
	NextStep              string                        `json:"next_step,omitempty"`
	ReviewPolicyState     string                        `json:"review_policy_state,omitempty"`
	ReviewDecision        string                        `json:"review_decision,omitempty"`
	ReviewReady           bool                          `json:"review_ready,omitempty"`
	BrowserTools          []string                      `json:"browser_tools,omitempty"`
	ArtifactTools         []string                      `json:"artifact_tools,omitempty"`
	ArtifactKinds         []string                      `json:"artifact_kinds,omitempty"`
	ArtifactContract      string                        `json:"artifact_contract,omitempty"`
	BrowserActKinds       []string                      `json:"browser_act_kinds,omitempty"`
	BrowserSurface        string                        `json:"browser_surface,omitempty"`
	BrowserOptInTargets   []string                      `json:"browser_opt_in_targets,omitempty"`
}

type browserTopLevelViewSummary struct {
	Kind                  string                        `json:"kind,omitempty"`
	Ready                 bool                          `json:"ready,omitempty"`
	Sections              []string                      `json:"sections,omitempty"`
	Category              string                        `json:"category,omitempty"`
	State                 string                        `json:"state,omitempty"`
	SummaryCode           string                        `json:"summary_code,omitempty"`
	RepairCommand         string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	NextStepAlias         string                        `json:"next_step_alias,omitempty"`
	ManualRetryHint       string                        `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback   bool                          `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction  string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction     string                        `json:"primary_node_action,omitempty"`
	NextStep              string                        `json:"next_step,omitempty"`
	BrowserTools          []string                      `json:"browser_tools,omitempty"`
	ArtifactTools         []string                      `json:"artifact_tools,omitempty"`
	ArtifactKinds         []string                      `json:"artifact_kinds,omitempty"`
	ArtifactContract      string                        `json:"artifact_contract,omitempty"`
	BrowserActKinds       []string                      `json:"browser_act_kinds,omitempty"`
	BrowserSurface        string                        `json:"browser_surface,omitempty"`
	BrowserOptInTargets   []string                      `json:"browser_opt_in_targets,omitempty"`
	Review                *browserReviewSurfaceSummary  `json:"review,omitempty"`
}

type browserReviewSurfaceSummary struct {
	PolicyState string                         `json:"policy_state,omitempty"`
	Decision    string                         `json:"decision,omitempty"`
	Ready       bool                           `json:"ready,omitempty"`
	Explanation *browserTopLevelSummary        `json:"explanation,omitempty"`
	Diagnostics *browserTopLevelSummary        `json:"diagnostics,omitempty"`
	Summary     *browserTopLevelSummary        `json:"summary,omitempty"`
	Display     *browserTopLevelDisplaySummary `json:"display,omitempty"`
}

func browserCloneTopLevelSummary(summary *browserTopLevelSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	return &cloned
}

func browserCloneTopLevelDisplaySummary(summary *browserTopLevelDisplaySummary) *browserTopLevelDisplaySummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	if len(summary.Sections) > 0 {
		cloned.Sections = append([]string(nil), summary.Sections...)
	}
	return &cloned
}

func browserCloneTopLevelSurfaceSummary(summary *browserTopLevelSurfaceSummary) *browserTopLevelSurfaceSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	if len(summary.Sections) > 0 {
		cloned.Sections = append([]string(nil), summary.Sections...)
	}
	if len(summary.BrowserTools) > 0 {
		cloned.BrowserTools = append([]string(nil), summary.BrowserTools...)
	}
	if len(summary.ArtifactTools) > 0 {
		cloned.ArtifactTools = append([]string(nil), summary.ArtifactTools...)
	}
	if len(summary.ArtifactKinds) > 0 {
		cloned.ArtifactKinds = append([]string(nil), summary.ArtifactKinds...)
	}
	cloned.ArtifactContract = strings.TrimSpace(summary.ArtifactContract)
	if len(summary.BrowserActKinds) > 0 {
		cloned.BrowserActKinds = append([]string(nil), summary.BrowserActKinds...)
	}
	if len(summary.BrowserOptInTargets) > 0 {
		cloned.BrowserOptInTargets = append([]string(nil), summary.BrowserOptInTargets...)
	}
	return &cloned
}

func browserCloneTopLevelViewSummary(summary *browserTopLevelViewSummary) *browserTopLevelViewSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	if len(summary.Sections) > 0 {
		cloned.Sections = append([]string(nil), summary.Sections...)
	}
	if len(summary.BrowserTools) > 0 {
		cloned.BrowserTools = append([]string(nil), summary.BrowserTools...)
	}
	if len(summary.ArtifactTools) > 0 {
		cloned.ArtifactTools = append([]string(nil), summary.ArtifactTools...)
	}
	if len(summary.ArtifactKinds) > 0 {
		cloned.ArtifactKinds = append([]string(nil), summary.ArtifactKinds...)
	}
	cloned.ArtifactContract = strings.TrimSpace(summary.ArtifactContract)
	if len(summary.BrowserActKinds) > 0 {
		cloned.BrowserActKinds = append([]string(nil), summary.BrowserActKinds...)
	}
	if len(summary.BrowserOptInTargets) > 0 {
		cloned.BrowserOptInTargets = append([]string(nil), summary.BrowserOptInTargets...)
	}
	cloned.Review = browserCloneReviewSurfaceSummary(summary.Review)
	return &cloned
}

func browserCloneReviewSurfaceSummary(summary *browserReviewSurfaceSummary) *browserReviewSurfaceSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	cloned.Explanation = browserCloneTopLevelSummary(summary.Explanation)
	cloned.Diagnostics = browserCloneTopLevelSummary(summary.Diagnostics)
	cloned.Summary = browserCloneTopLevelSummary(summary.Summary)
	cloned.Display = browserCloneTopLevelDisplaySummary(summary.Display)
	return &cloned
}

func browserReviewSurfaceSummaryEmpty(summary browserReviewSurfaceSummary) bool {
	return strings.TrimSpace(summary.PolicyState) == "" &&
		strings.TrimSpace(summary.Decision) == "" &&
		!summary.Ready &&
		summary.Explanation == nil &&
		summary.Diagnostics == nil &&
		summary.Summary == nil &&
		summary.Display == nil
}

func browserTopLevelSurfaceEmpty(summary browserTopLevelSurfaceSummary) bool {
	return !summary.Ready &&
		len(summary.Sections) == 0 &&
		strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == "" &&
		strings.TrimSpace(summary.ReviewPolicyState) == "" &&
		strings.TrimSpace(summary.ReviewDecision) == "" &&
		!summary.ReviewReady &&
		len(summary.BrowserTools) == 0 &&
		len(summary.ArtifactTools) == 0 &&
		len(summary.ArtifactKinds) == 0 &&
		strings.TrimSpace(summary.ArtifactContract) == "" &&
		len(summary.BrowserActKinds) == 0 &&
		strings.TrimSpace(summary.BrowserSurface) == "" &&
		len(summary.BrowserOptInTargets) == 0
}

func browserTopLevelViewEmpty(summary browserTopLevelViewSummary) bool {
	return !summary.Ready &&
		len(summary.Sections) == 0 &&
		strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == "" &&
		len(summary.BrowserTools) == 0 &&
		len(summary.ArtifactTools) == 0 &&
		len(summary.ArtifactKinds) == 0 &&
		strings.TrimSpace(summary.ArtifactContract) == "" &&
		len(summary.BrowserActKinds) == 0 &&
		strings.TrimSpace(summary.BrowserSurface) == "" &&
		len(summary.BrowserOptInTargets) == 0 &&
		summary.Review == nil
}

func browserTopLevelDisplayFromSummary(summary *browserTopLevelSummary) *browserTopLevelDisplaySummary {
	if summary == nil {
		return nil
	}
	display := &browserTopLevelDisplaySummary{
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: summary.DefaultCandidateRoute,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserTopLevelDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserTopLevelDisplayFromWorkbenchDisplay(summary *browserRuntimeWorkbenchDisplaySummary) *browserTopLevelDisplaySummary {
	if summary == nil {
		return nil
	}
	display := &browserTopLevelDisplaySummary{
		Ready:                 summary.Ready,
		Sections:              append([]string(nil), summary.Sections...),
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: summary.DefaultCandidateRoute,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserTopLevelDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserTopLevelDisplayFromSurface(summary *browserTopLevelSurfaceSummary) *browserTopLevelDisplaySummary {
	if summary == nil {
		return nil
	}
	display := &browserTopLevelDisplaySummary{
		Ready:                 summary.Ready,
		Sections:              append([]string(nil), summary.Sections...),
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: summary.DefaultCandidateRoute,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserTopLevelDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserTopLevelDisplayFromView(summary *browserTopLevelViewSummary) *browserTopLevelDisplaySummary {
	if summary == nil {
		return nil
	}
	route := browserRuntimeRouteDescriptorFromTopLevelView(summary)
	display := &browserTopLevelDisplaySummary{
		Ready:                 summary.Ready,
		Sections:              append([]string(nil), summary.Sections...),
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: route,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserTopLevelDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserTopLevelSummaryFromDisplay(summary *browserTopLevelDisplaySummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: summary.DefaultCandidateRoute,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSummaryFromSurface(summary *browserTopLevelSurfaceSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: summary.DefaultCandidateRoute,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSummaryFromView(summary *browserTopLevelViewSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	route := browserRuntimeRouteDescriptorFromTopLevelView(summary)
	out := &browserTopLevelSummary{
		Category:              strings.TrimSpace(summary.Category),
		State:                 strings.TrimSpace(summary.State),
		SummaryCode:           strings.TrimSpace(summary.SummaryCode),
		RepairCommand:         strings.TrimSpace(summary.RepairCommand),
		DefaultCandidateRoute: route,
		NextStepAlias:         strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:       strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:   summary.ResolvedViaFallback,
		PrimaryBrowserAction:  strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:     strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:              strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSummaryFromReview(summary *browserReviewSurfaceSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	route := browserRuntimeRouteDescriptorFromReviewSurface(summary)
	if out := browserCloneTopLevelSummary(summary.Summary); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	if out := browserCloneTopLevelSummary(summary.Diagnostics); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	if out := browserCloneTopLevelSummary(summary.Explanation); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	if out := browserTopLevelSummaryFromDisplay(summary.Display); out != nil {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	return nil
}

func browserTopLevelSummaryFromWorkbench(summary *browserRuntimeWorkbenchSurfaceSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	route := browserRuntimeRouteDescriptorFromWorkbenchSurface(summary)
	if out := browserCloneTopLevelSummary(summary.Summary); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	if out := browserCloneTopLevelSummary(summary.Diagnostics); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	if out := browserCloneTopLevelSummary(summary.Explanation); out != nil && !browserUnifiedSummaryEmpty(*out) {
		out.DefaultCandidateRoute = firstBrowserRuntimeRouteDescriptor(out.DefaultCandidateRoute, route)
		return out
	}
	return nil
}

func browserReviewSurfaceSummaryFromView(summary *browserTopLevelViewSummary) *browserReviewSurfaceSummary {
	if summary == nil {
		return nil
	}
	return browserReviewSurfaceSummaryWithDefaultCandidateRoute(
		summary.Review,
		browserRuntimeRouteDescriptorFromTopLevelView(summary),
	)
}

func browserReviewSurfaceSummaryFromWorkbench(summary *browserRuntimeWorkbenchSurfaceSummary) *browserReviewSurfaceSummary {
	if summary == nil {
		return nil
	}
	return browserReviewSurfaceSummaryWithDefaultCandidateRoute(
		summary.Review,
		browserRuntimeRouteDescriptorFromWorkbenchSurface(summary),
	)
}

func browserTopLevelDisplayEmpty(summary browserTopLevelDisplaySummary) bool {
	return !summary.Ready &&
		len(summary.Sections) == 0 &&
		strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == ""
}

func browserTopLevelNextStepAliasFromAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return ""
	}
	lower := strings.ToLower(action)
	switch {
	case strings.HasPrefix(lower, "browser action="):
		return strings.TrimSpace(action[len("browser action="):])
	case strings.HasPrefix(lower, "nodes action="):
		return strings.TrimSpace(action[len("nodes action="):])
	default:
		return action
	}
}

func browserTopLevelSummaryFromDiagnosticsExplanation(explanation *browserDiagnosticsExplanationSummary) *browserTopLevelSummary {
	if explanation == nil {
		return nil
	}
	return browserTopLevelSummaryFromDiagnosticFields(
		explanation.Category,
		explanation.State,
		explanation.SummaryCode,
		explanation.NextStepAlias,
		explanation.ManualRetryHint,
	)
}

func browserTopLevelSummaryFromRuntimeDiagnosticsExplanation(explanation *browserRuntimeDiagnosticsExplanationSummary) *browserTopLevelSummary {
	if explanation == nil {
		return nil
	}
	return browserTopLevelSummaryFromDiagnosticFields(
		explanation.Category,
		explanation.State,
		explanation.SummaryCode,
		explanation.NextStepAlias,
		explanation.ManualRetryHint,
	)
}

func browserTopLevelSummaryFromResolverExplanation(explanation *browserRuntimeResolverExplanationSummary) *browserTopLevelSummary {
	if explanation == nil {
		return nil
	}
	summary := &browserTopLevelSummary{
		State:           strings.TrimSpace(explanation.State),
		SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
		NextStepAlias:   strings.TrimSpace(explanation.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
	}
	if strings.EqualFold(summary.State, "resolved_via_fallback") {
		summary.Category = "resolver_fallback"
		summary.ResolvedViaFallback = true
	} else if summary.State != "" || summary.SummaryCode != "" || summary.NextStepAlias != "" || summary.ManualRetryHint != "" {
		summary.Category = "resolver"
	}
	if browserUnifiedSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func browserTopLevelSummaryFromWorkbenchDiagnostics(summary *browserRuntimeWorkbenchDiagnosticsSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:             strings.TrimSpace(summary.Category),
		State:                strings.TrimSpace(summary.State),
		SummaryCode:          strings.TrimSpace(summary.SummaryCode),
		RepairCommand:        strings.TrimSpace(summary.RepairCommand),
		NextStepAlias:        strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:  summary.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSummaryFromDiagnosticFields(category, state, summaryCode, nextStepAlias, manualRetryHint string) *browserTopLevelSummary {
	summary := &browserTopLevelSummary{
		Category:        strings.TrimSpace(category),
		State:           strings.TrimSpace(state),
		SummaryCode:     strings.TrimSpace(summaryCode),
		NextStepAlias:   strings.TrimSpace(nextStepAlias),
		ManualRetryHint: strings.TrimSpace(manualRetryHint),
	}
	if strings.EqualFold(summary.State, "resolved_via_fallback") {
		summary.ResolvedViaFallback = true
	}
	if summary.Category == "" {
		if summary.ResolvedViaFallback {
			summary.Category = "resolver_fallback"
		} else if summary.State != "" || summary.SummaryCode != "" || summary.NextStepAlias != "" || summary.ManualRetryHint != "" {
			summary.Category = "resolver"
		}
	}
	if browserUnifiedSummaryEmpty(*summary) {
		return nil
	}
	return summary
}
