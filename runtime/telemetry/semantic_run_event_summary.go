package telemetry

import "strings"

type SemanticRunEventSummary struct {
	TotalEvents               int            `json:"total_events"`
	ByKind                    map[string]int `json:"by_kind,omitempty"`
	ByStage                   map[string]int `json:"by_stage,omitempty"`
	ByTool                    map[string]int `json:"by_tool,omitempty"`
	ByApprovalDecision        map[string]int `json:"by_approval_decision,omitempty"`
	ByApprovalReason          map[string]int `json:"by_approval_reason,omitempty"`
	ByInterruptionReason      map[string]int `json:"by_interruption_reason,omitempty"`
	RunInterrupted            int            `json:"run_interrupted,omitempty"`
	RunResumed                int            `json:"run_resumed,omitempty"`
	ApprovalRequested         int            `json:"approval_requested,omitempty"`
	ApprovalResolved          int            `json:"approval_resolved,omitempty"`
	ApprovalDenied            int            `json:"approval_denied,omitempty"`
	ApprovalRequiresConfirm   int            `json:"approval_requires_confirm,omitempty"`
	CheckpointInterruptions   int            `json:"checkpoint_interruptions,omitempty"`
	TerminationInterruptions  int            `json:"termination_interruptions,omitempty"`
	PendingToolCallInterrupts int            `json:"pending_tool_call_interrupts,omitempty"`
}

func SummarizeSemanticRunEvents(events []SemanticRunEvent) SemanticRunEventSummary {
	summary := SemanticRunEventSummary{
		ByKind:               map[string]int{},
		ByStage:              map[string]int{},
		ByTool:               map[string]int{},
		ByApprovalDecision:   map[string]int{},
		ByApprovalReason:     map[string]int{},
		ByInterruptionReason: map[string]int{},
	}
	for _, event := range events {
		event = normalizeSemanticRunEvent(event)
		if event.Kind == "" {
			continue
		}
		summary.TotalEvents++
		summary.ByKind[event.Kind]++
		if event.Stage != "" {
			summary.ByStage[event.Stage]++
		}
		if event.Tool != "" {
			summary.ByTool[event.Tool]++
		}
		if event.Reason != "" && event.Kind == SemanticRunEventKindRunInterrupted {
			summary.ByInterruptionReason[event.Reason]++
		}
		switch event.Kind {
		case SemanticRunEventKindRunInterrupted:
			summary.RunInterrupted++
			if event.Checkpoint != nil {
				summary.CheckpointInterruptions++
				if event.Checkpoint.HasPendingToolCalls {
					summary.PendingToolCallInterrupts++
				}
			}
			if event.Termination != nil {
				summary.TerminationInterruptions++
			}
		case SemanticRunEventKindRunResumed:
			summary.RunResumed++
		case SemanticRunEventKindApprovalRequested:
			summary.ApprovalRequested++
		case SemanticRunEventKindApprovalResolved:
			summary.ApprovalResolved++
		}
		if event.Approval != nil {
			if decision := strings.TrimSpace(event.Approval.Decision); decision != "" {
				summary.ByApprovalDecision[decision]++
			}
			if reason := strings.TrimSpace(firstNonEmptyString(event.Approval.Reason, event.Reason)); reason != "" {
				summary.ByApprovalReason[reason]++
			}
			if event.Approval.Denied {
				summary.ApprovalDenied++
			}
			if event.Approval.RequiresConfirm {
				summary.ApprovalRequiresConfirm++
			}
		}
	}
	if len(summary.ByKind) == 0 {
		summary.ByKind = nil
	}
	if len(summary.ByStage) == 0 {
		summary.ByStage = nil
	}
	if len(summary.ByTool) == 0 {
		summary.ByTool = nil
	}
	if len(summary.ByApprovalDecision) == 0 {
		summary.ByApprovalDecision = nil
	}
	if len(summary.ByApprovalReason) == 0 {
		summary.ByApprovalReason = nil
	}
	if len(summary.ByInterruptionReason) == 0 {
		summary.ByInterruptionReason = nil
	}
	return summary
}
