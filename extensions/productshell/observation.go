package productshell

import "strings"

// BuildHostDiagnosticOperatorLineObservation normalizes one typed operator-line
// input. Empty input produces no observation.
func BuildHostDiagnosticOperatorLineObservation(input HostDiagnosticOperatorLineObservationInput) *HostDiagnosticOperatorLineObservation {
	missingInputs := normalizeObservationStrings(input.MissingInputs)
	blockedReasons := normalizeObservationStrings(input.BlockedReasons)
	boundaries := normalizeObservationStrings(input.Boundaries)
	source := strings.TrimSpace(input.Source)
	key := strings.TrimSpace(input.Key)
	status := strings.TrimSpace(input.Status)
	operatorDisplayLine := strings.TrimSpace(input.OperatorDisplayLine)
	nextHostAction := strings.TrimSpace(input.NextHostAction)
	if source == "" &&
		key == "" &&
		status == "" &&
		operatorDisplayLine == "" &&
		nextHostAction == "" &&
		len(missingInputs) == 0 &&
		len(blockedReasons) == 0 &&
		len(boundaries) == 0 {
		return nil
	}
	if source == "" {
		source = "host_diagnostic_operator_line"
	}
	return &HostDiagnosticOperatorLineObservation{
		Source:              source,
		Key:                 key,
		Available:           input.Available,
		Status:              status,
		OperatorDisplayLine: operatorDisplayLine,
		MissingInputs:       missingInputs,
		BlockedReasons:      blockedReasons,
		Boundaries:          boundaries,
		NextHostAction:      nextHostAction,
	}
}

// BuildHostProcessProgressObservation normalizes typed host progress without
// interpreting host lifecycle or run-store semantics.
func BuildHostProcessProgressObservation(input HostProcessProgressObservationInput) *HostProcessProgressObservation {
	missingInputs := normalizeObservationStrings(input.MissingInputs)
	blockedReasons := normalizeObservationStrings(input.BlockedReasons)
	boundaries := normalizeObservationStrings(input.Boundaries)
	source := strings.TrimSpace(input.Source)
	status := strings.TrimSpace(input.Status)
	displayKind := strings.TrimSpace(input.DisplayKind)
	summaryCode := strings.TrimSpace(input.SummaryCode)
	displayLine := strings.TrimSpace(input.DisplayLine)
	processRef := strings.TrimSpace(input.ProcessRef)
	if source == "" &&
		status == "" &&
		displayKind == "" &&
		summaryCode == "" &&
		displayLine == "" &&
		processRef == "" &&
		len(missingInputs) == 0 &&
		len(blockedReasons) == 0 &&
		len(boundaries) == 0 {
		return nil
	}
	if source == "" {
		source = "productshellruntime_host_process_progress_display"
	}
	return &HostProcessProgressObservation{
		Source:                            source,
		Available:                         input.Available,
		Enabled:                           input.Enabled,
		Status:                            status,
		DisplayKind:                       displayKind,
		SummaryCode:                       summaryCode,
		DisplayLine:                       displayLine,
		SessionKey:                        strings.TrimSpace(input.SessionKey),
		ProductShellRef:                   strings.TrimSpace(input.ProductShellRef),
		WorkspaceRef:                      strings.TrimSpace(input.WorkspaceRef),
		ProcessRef:                        processRef,
		RunID:                             strings.TrimSpace(input.RunID),
		BranchID:                          strings.TrimSpace(input.BranchID),
		NodeExecID:                        strings.TrimSpace(input.NodeExecID),
		RuntimeDecisionRef:                strings.TrimSpace(input.RuntimeDecisionRef),
		RequestRef:                        strings.TrimSpace(input.RequestRef),
		ResultRef:                         strings.TrimSpace(input.ResultRef),
		ArtifactBundleRef:                 strings.TrimSpace(input.ArtifactBundleRef),
		StdoutRef:                         strings.TrimSpace(input.StdoutRef),
		StderrRef:                         strings.TrimSpace(input.StderrRef),
		ProcessStatus:                     strings.TrimSpace(input.ProcessStatus),
		LastKind:                          strings.TrimSpace(input.LastKind),
		ExitCode:                          input.ExitCode,
		ExitCodeKnown:                     input.ExitCodeKnown,
		ReadyForReadback:                  input.ReadyForReadback,
		Started:                           input.Started,
		Terminal:                          input.Terminal,
		Failed:                            input.Failed,
		Cancelled:                         input.Cancelled,
		TimedOut:                          input.TimedOut,
		ProcessCount:                      observationNonNegative(input.ProcessCount),
		ActiveCount:                       observationNonNegative(input.ActiveCount),
		TerminalCount:                     observationNonNegative(input.TerminalCount),
		HostProcessEventCount:             observationNonNegative(input.HostProcessEventCount),
		ViewReady:                         input.ViewReady,
		ProgressReady:                     input.ProgressReady,
		ConsumesHostProcessSessionView:    input.ConsumesHostProcessSessionView,
		ReadsToolOutput:                   input.ReadsToolOutput,
		BuildsProcessMapFromToolOutput:    input.BuildsProcessMapFromToolOutput,
		RunstoreProtocolAuthorizesProcess: input.RunstoreProtocolAuthorizesProcess,
		ProcessLifecycleControlsExecution: input.ProcessLifecycleControlsExecution,
		MissingInputs:                     missingInputs,
		BlockedReasons:                    blockedReasons,
		Boundaries:                        boundaries,
		NextHostAction:                    strings.TrimSpace(input.NextHostAction),
		RawOutputLoaded:                   input.RawOutputLoaded,
	}
}

// BuildSessionObservation normalizes typed session evidence into a compact
// observation. It does not read an engine session or transcript backend.
func BuildSessionObservation(input SessionObservationInput) *SessionObservation {
	sessionID := strings.TrimSpace(input.SessionID)
	compaction := buildSessionCompactionObservation(input.Compaction)
	branches := buildSessionBranchObservations(input.Branches)
	latestSummary := observationPreview(input.LatestSummary, 220)
	if sessionID == "" && len(input.Events) == 0 && len(branches) == 0 && compaction == nil && latestSummary == "" {
		return nil
	}
	out := &SessionObservation{
		Source:           "engine_session",
		SessionID:        sessionID,
		CreatedUnixMs:    input.CreatedUnixMs,
		LatestSummary:    latestSummary,
		SummaryVersion:   input.SummaryVersion,
		SummaryUpdatedAt: input.SummaryUpdatedAt,
		Branches:         branches,
		Compaction:       compaction,
	}
	for _, event := range input.Events {
		role := strings.ToLower(strings.TrimSpace(event.Role))
		content := observationPreview(event.Content, 180)
		toolCallID := strings.TrimSpace(event.ToolCallID)
		toolCallCount := event.ToolCallCount
		if toolCallCount < 0 {
			toolCallCount = 0
		}
		if role == "" && content == "" && toolCallID == "" && toolCallCount == 0 {
			continue
		}
		out.EventCount++
		switch role {
		case "user":
			out.UserMessageCount++
			if content != "" {
				out.LatestUserPreview = content
			}
		case "assistant":
			out.AssistantMessageCount++
			if content != "" {
				out.LatestAssistantPreview = content
			}
		case "tool":
			if toolCallID == "" {
				out.ToolResultCount++
			}
			if content != "" {
				out.LatestToolPreview = content
			}
		}
		if toolCallID != "" {
			out.ToolResultCount++
		}
		if toolCallCount > 0 {
			out.ToolCallMessageCount++
		}
	}
	out.BranchCount = len(out.Branches)
	out.Labels = sessionObservationLabels(*out)
	return out
}

func normalizeObservationStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func appendUniqueProductShellString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueProductShellStrings(values []string, additions ...string) []string {
	for _, addition := range additions {
		values = appendUniqueProductShellString(values, addition)
	}
	return values
}

func observationNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func buildSessionCompactionObservation(input SessionCompactionObservationInput) *SessionCompactionObservation {
	applied := input.CompactedToolOutputs > 0 ||
		input.CompactedHistoryBodies > 0 ||
		input.ProtocolAwareHistoryDrops > 0
	sanitized := input.Passes > 0 ||
		input.SynthesizedToolCallIDs > 0 ||
		input.RecoveredToolResults > 0 ||
		input.DowngradedToolResults > 0 ||
		input.StrippedReasoningMsgs > 0 ||
		input.MergedMessages > 0 ||
		applied
	if !sanitized && !input.StrictProvider {
		return nil
	}
	return &SessionCompactionObservation{
		Source:                    "transcript_sanitize",
		Sanitized:                 sanitized,
		Applied:                   applied,
		Passes:                    input.Passes,
		StrictProvider:            input.StrictProvider,
		SynthesizedToolCallIDs:    input.SynthesizedToolCallIDs,
		RecoveredToolResults:      input.RecoveredToolResults,
		DowngradedToolResults:     input.DowngradedToolResults,
		StrippedReasoningMsgs:     input.StrippedReasoningMsgs,
		MergedMessages:            input.MergedMessages,
		CompactedToolOutputs:      input.CompactedToolOutputs,
		CompactedHistoryBodies:    input.CompactedHistoryBodies,
		ProtocolAwareHistoryDrops: input.ProtocolAwareHistoryDrops,
	}
}

func buildSessionBranchObservations(inputs []SessionBranchObservationInput) []SessionBranchObservation {
	if len(inputs) == 0 {
		return nil
	}
	order := make([]string, 0, len(inputs))
	byID := map[string]*SessionBranchObservation{}
	for _, input := range inputs {
		branchID := strings.TrimSpace(input.BranchID)
		if branchID == "" {
			continue
		}
		item := byID[branchID]
		if item == nil {
			item = &SessionBranchObservation{BranchID: branchID}
			byID[branchID] = item
			order = append(order, branchID)
		}
		item.NodeCount++
		if input.StartedAt > 0 && (item.StartedAt <= 0 || input.StartedAt < item.StartedAt) {
			item.StartedAt = input.StartedAt
		}
		if input.FinishedAt > item.FinishedAt {
			item.FinishedAt = input.FinishedAt
		}
		item.LastNodeExecID = strings.TrimSpace(input.NodeExecID)
		item.LastNodeID = strings.TrimSpace(input.NodeID)
		item.LastStatus = strings.TrimSpace(input.Status)
		if sessionBranchStatusTerminal(item.LastStatus) || input.FinishedAt > 0 {
			item.Terminal = true
		}
	}
	if len(order) == 0 {
		return nil
	}
	out := make([]SessionBranchObservation, 0, len(order))
	for _, branchID := range order {
		if item := byID[branchID]; item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func sessionBranchStatusTerminal(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed", "cancelled", "canceled", "timeout", "timed_out", "skipped":
		return true
	default:
		return false
	}
}

func sessionObservationLabels(session SessionObservation) []string {
	labels := make([]string, 0, 7)
	if strings.TrimSpace(session.SessionID) != "" {
		labels = append(labels, "current_session")
	}
	if session.EventCount > 0 {
		labels = append(labels, "has_history")
	}
	if session.ToolResultCount > 0 {
		labels = append(labels, "has_tool_results")
	}
	if session.ToolCallMessageCount > 0 {
		labels = append(labels, "has_tool_calls")
	}
	if session.BranchCount > 0 {
		labels = append(labels, "branched")
	}
	if session.Compaction != nil {
		labels = append(labels, "transcript_sanitized")
		if session.Compaction.Applied {
			labels = append(labels, "compacted")
		}
	}
	return labels
}

func observationPreview(value string, max int) string {
	text := strings.Join(strings.Fields(value), " ")
	runes := []rune(text)
	if text == "" || max <= 0 || len(runes) <= max {
		return text
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return strings.TrimSpace(string(runes[:max-3])) + "..."
}
