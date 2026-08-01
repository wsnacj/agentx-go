package executionpolicy

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const (
	ExecutionLoopReportContractVersion = "agentx.execution_loop_report.v1"
	ExecutionLoopReportKind            = "execution_loop_report"

	ExecutionLoopStatusBlocked               = "execution_loop_blocked"
	ExecutionLoopStatusReadyForHostExecution = "execution_loop_ready_for_host_execution"
	ExecutionLoopStatusReadyForRetry         = "execution_loop_ready_for_retry"

	ExecutionLoopActionHostExecution     = "host_execution"
	ExecutionLoopActionDeniedReadReview  = "denied_read_review"
	ExecutionLoopActionSandboxEscalation = "sandbox_escalation"
	ExecutionLoopActionNetworkApproval   = "network_approval"
	ExecutionLoopActionHostApproval      = "host_approval"
	ExecutionLoopActionExecutionReview   = "execution_review"
)

type ExecutionLoopReportInput struct {
	Enabled              bool
	ExecutionLoopRef     string
	DecisionPacketRef    string
	AttemptRef           string
	RetryRequestRef      string
	DeniedReadReviewRef  string
	SandboxEscalationRef string
	NetworkApprovalRef   string
	HostApprovalRef      string
	ExecutionReviewRef   string
	DecisionPacket       DecisionPacket
	Boundaries           []string
	NextHostAction       string
}

type ExecutionLoopReportContract struct {
	Version                 string `json:"version"`
	Owner                   string `json:"owner"`
	ObservationOnly         bool   `json:"observation_only"`
	PolicySource            bool   `json:"policy_source"`
	AuthorizationSource     bool   `json:"authorization_source"`
	RuntimeInvocationSource bool   `json:"runtime_invocation_source"`
	HostAdapterOwner        string `json:"host_adapter_owner"`
	ExecutionPolicyOwner    string `json:"execution_policy_owner"`
}

type ExecutionLoopReport struct {
	Available                          bool                        `json:"available"`
	Enabled                            bool                        `json:"enabled"`
	Status                             string                      `json:"status,omitempty"`
	ReportKind                         string                      `json:"report_kind,omitempty"`
	Contract                           ExecutionLoopReportContract `json:"contract"`
	ExecutionLoopRef                   string                      `json:"execution_loop_ref,omitempty"`
	DecisionPacketRef                  string                      `json:"decision_packet_ref,omitempty"`
	AttemptRef                         string                      `json:"attempt_ref,omitempty"`
	RetryRequestRef                    string                      `json:"retry_request_ref,omitempty"`
	DeniedReadReviewRef                string                      `json:"denied_read_review_ref,omitempty"`
	SandboxEscalationRef               string                      `json:"sandbox_escalation_ref,omitempty"`
	NetworkApprovalRef                 string                      `json:"network_approval_ref,omitempty"`
	HostApprovalRef                    string                      `json:"host_approval_ref,omitempty"`
	ExecutionReviewRef                 string                      `json:"execution_review_ref,omitempty"`
	DecisionPacketAvailable            bool                        `json:"decision_packet_available"`
	DecisionPacketFinalAction          DecisionAction              `json:"decision_packet_final_action,omitempty"`
	DecisionPacketSteps                int                         `json:"decision_packet_steps,omitempty"`
	DecisionPacketSummary              DecisionPacketSummary       `json:"decision_packet_summary,omitempty"`
	ExecutionAllowed                   bool                        `json:"execution_allowed"`
	ExecutionBlocked                   bool                        `json:"execution_blocked"`
	ReadyForHostExecution              bool                        `json:"ready_for_host_execution"`
	ReadyForRetry                      bool                        `json:"ready_for_retry"`
	RetryRequired                      bool                        `json:"retry_required"`
	RequiresDeniedReadReview           bool                        `json:"requires_denied_read_review"`
	ReadyForDeniedReadReview           bool                        `json:"ready_for_denied_read_review"`
	RequiresSandboxEscalation          bool                        `json:"requires_sandbox_escalation"`
	ReadyForSandboxEscalation          bool                        `json:"ready_for_sandbox_escalation"`
	RequiresNetworkApproval            bool                        `json:"requires_network_approval"`
	ReadyForNetworkApproval            bool                        `json:"ready_for_network_approval"`
	RequiresHostApproval               bool                        `json:"requires_host_approval"`
	ReadyForHostApproval               bool                        `json:"ready_for_host_approval"`
	RequiresExecutionReview            bool                        `json:"requires_execution_review"`
	ReadyForExecutionReview            bool                        `json:"ready_for_execution_review"`
	Actions                            []ExecutionLoopAction       `json:"actions,omitempty"`
	MissingInputs                      []string                    `json:"missing_inputs,omitempty"`
	BlockedReasons                     []string                    `json:"blocked_reasons,omitempty"`
	Boundaries                         []string                    `json:"boundaries,omitempty"`
	NextHostAction                     string                      `json:"next_host_action,omitempty"`
	RawOutputLoaded                    bool                        `json:"raw_output_loaded"`
	NoCoreExecution                    bool                        `json:"no_core_execution"`
	NoToolInvocation                   bool                        `json:"no_tool_invocation"`
	NoHostAdapterInvocation            bool                        `json:"no_host_adapter_invocation"`
	DecisionPacketFinalActionPreserved bool                        `json:"decision_packet_final_action_preserved"`
}

type ExecutionLoopAction struct {
	Kind            string           `json:"kind,omitempty"`
	Required        bool             `json:"required"`
	Ready           bool             `json:"ready"`
	ActionRef       string           `json:"action_ref,omitempty"`
	StepKind        DecisionStepKind `json:"step_kind,omitempty"`
	StepAction      DecisionAction   `json:"step_action,omitempty"`
	Tool            string           `json:"tool,omitempty"`
	Reason          string           `json:"reason,omitempty"`
	DecisionSubject string           `json:"decision_subject,omitempty"`
	PolicySource    string           `json:"policy_source,omitempty"`
	ControlSource   string           `json:"control_source,omitempty"`
	NextHostAction  string           `json:"next_host_action,omitempty"`
}

func BuildExecutionLoopReport(input ExecutionLoopReportInput) ExecutionLoopReport {
	packet := normalizeExecutionLoopDecisionPacket(input.DecisionPacket)
	packetAvailable := executionLoopDecisionPacketAvailable(packet)
	token := executionLoopDecisionPacketToken(packet)
	report := ExecutionLoopReport{
		Enabled:                            input.Enabled,
		Status:                             ExecutionLoopStatusBlocked,
		ReportKind:                         ExecutionLoopReportKind,
		Contract:                           DefaultExecutionLoopReportContract(),
		DecisionPacketAvailable:            packetAvailable,
		DecisionPacketFinalAction:          packet.FinalAction,
		DecisionPacketSteps:                len(packet.Steps),
		DecisionPacketSummary:              packet.Summary,
		NoCoreExecution:                    true,
		NoToolInvocation:                   true,
		NoHostAdapterInvocation:            true,
		DecisionPacketFinalActionPreserved: true,
		Boundaries: executionLoopBoundaries(input.Boundaries,
			"execution_loop_report",
			"decision_packet_execution_loop_evidence",
			"execution_loop_does_not_invoke_tools",
			"execution_loop_does_not_call_host_adapter",
			"execution_loop_does_not_mutate_policy",
			"hostruntime_consumes_decision_ref_only",
			"display_safe_refs_only",
			"no_core_execution",
		),
		NextHostAction: executionLoopNextHostAction(input.NextHostAction, "review_execution_loop_report"),
	}
	report.ExecutionLoopRef = executionLoopRefWithFallback(input.ExecutionLoopRef, "execution_loop", token, &report.RawOutputLoaded)
	report.DecisionPacketRef = executionLoopRefWithFallback(input.DecisionPacketRef, "decision_packet", token, &report.RawOutputLoaded)
	report.AttemptRef = executionLoopRef(input.AttemptRef, &report.RawOutputLoaded)
	report.RetryRequestRef = executionLoopRef(input.RetryRequestRef, &report.RawOutputLoaded)
	report.DeniedReadReviewRef = executionLoopRef(input.DeniedReadReviewRef, &report.RawOutputLoaded)
	report.SandboxEscalationRef = executionLoopRef(input.SandboxEscalationRef, &report.RawOutputLoaded)
	report.NetworkApprovalRef = executionLoopRef(input.NetworkApprovalRef, &report.RawOutputLoaded)
	report.HostApprovalRef = executionLoopRef(input.HostApprovalRef, &report.RawOutputLoaded)
	report.ExecutionReviewRef = executionLoopRef(input.ExecutionReviewRef, &report.RawOutputLoaded)
	report.MissingInputs = executionLoopRequireRef(report.MissingInputs, report.ExecutionLoopRef, "host:execution_loop_ref")
	report.MissingInputs = executionLoopRequireRef(report.MissingInputs, report.DecisionPacketRef, "host:execution_decision_packet_ref")
	if report.RawOutputLoaded {
		return report.Normalize()
	}
	if !input.Enabled {
		report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_enabled")
		report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_default_off")
		report.Boundaries = executionLoopAppendUnique(report.Boundaries, "execution_loop_default_off")
		report.NextHostAction = "enable_execution_loop_report"
		return report.Normalize()
	}
	if !packetAvailable {
		report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_decision_packet")
		report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_decision_packet_missing")
		report.NextHostAction = "provide_execution_decision_packet"
		return report.Normalize()
	}
	report.Actions = executionLoopActionsForPacket(packet, report)
	report = executionLoopApplyActions(report)
	switch {
	case packet.FinalAction == DecisionActionAllow && len(report.Actions) == 0:
		report.Available = true
		report.ExecutionAllowed = true
		report.ReadyForHostExecution = true
		report.Status = ExecutionLoopStatusReadyForHostExecution
		report.NextHostAction = "invoke_host_runtime_with_decision_packet_consumer"
		report.Boundaries = executionLoopAppendUnique(report.Boundaries, "execution_loop_ready_for_host_execution")
	case len(report.Actions) > 0:
		report.ExecutionBlocked = true
		report.RetryRequired = true
		report.MissingInputs = executionLoopRequireRef(report.MissingInputs, report.RetryRequestRef, "host:execution_loop_retry_request_ref")
		if executionLoopActionsReady(report.Actions) && report.RetryRequestRef != "" && len(report.MissingInputs) == 0 && len(report.BlockedReasons) == 0 {
			report.Available = true
			report.ReadyForRetry = true
			report.Status = ExecutionLoopStatusReadyForRetry
			report.NextHostAction = "retry_with_new_execution_decision_packet"
			report.Boundaries = executionLoopAppendUnique(report.Boundaries, "execution_loop_ready_for_retry")
		}
	default:
		report.ExecutionBlocked = true
		report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_decision_packet_final_action_not_executable")
		report.NextHostAction = "review_execution_decision_packet_final_action"
	}
	return report.Normalize()
}

func DefaultExecutionLoopReportContract() ExecutionLoopReportContract {
	return ExecutionLoopReportContract{
		Version:                 ExecutionLoopReportContractVersion,
		Owner:                   "execution",
		ObservationOnly:         true,
		PolicySource:            false,
		AuthorizationSource:     false,
		RuntimeInvocationSource: false,
		HostAdapterOwner:        "hostruntime",
		ExecutionPolicyOwner:    "execution",
	}
}

func (r ExecutionLoopReport) Normalize() ExecutionLoopReport {
	out := r
	if strings.TrimSpace(out.Status) == "" {
		out.Status = ExecutionLoopStatusBlocked
	}
	if strings.TrimSpace(out.ReportKind) == "" {
		out.ReportKind = ExecutionLoopReportKind
	}
	out.Contract = normalizeExecutionLoopReportContract(out.Contract)
	out.DecisionPacketFinalAction = NormalizeDecisionAction(out.DecisionPacketFinalAction)
	out.Actions = normalizeExecutionLoopActions(out.Actions)
	out.MissingInputs = executionLoopAppendUnique(nil, out.MissingInputs...)
	out.BlockedReasons = executionLoopAppendUnique(nil, out.BlockedReasons...)
	out.Boundaries = executionLoopBoundaries(out.Boundaries,
		"execution_loop_report",
		"decision_packet_execution_loop_evidence",
		"execution_loop_does_not_invoke_tools",
		"execution_loop_does_not_call_host_adapter",
		"no_core_execution",
	)
	out.NextHostAction = executionLoopNextHostAction(out.NextHostAction, "review_execution_loop_report")
	out.NoCoreExecution = true
	out.NoToolInvocation = true
	out.NoHostAdapterInvocation = true
	out.DecisionPacketFinalActionPreserved = true
	if out.RawOutputLoaded {
		out.Available = false
		out.ReadyForHostExecution = false
		out.ReadyForRetry = false
		out.Status = ExecutionLoopStatusBlocked
		out.MissingInputs = executionLoopAppendUnique(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = executionLoopAppendUnique(out.BlockedReasons, "unsafe_execution_loop_ref")
		out.Boundaries = executionLoopAppendUnique(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "remove_raw_execution_loop_refs"
	}
	if !out.ReadyForHostExecution && !out.ReadyForRetry {
		out.Available = false
		if out.Status == ExecutionLoopStatusReadyForHostExecution || out.Status == ExecutionLoopStatusReadyForRetry {
			out.Status = ExecutionLoopStatusBlocked
		}
	}
	return out
}

func executionLoopActionsForPacket(packet DecisionPacket, report ExecutionLoopReport) []ExecutionLoopAction {
	actions := []ExecutionLoopAction{}
	for _, step := range packet.Steps {
		step = NormalizeDecisionStep(step)
		if step.Action == DecisionActionAllow || step.Action == DecisionActionNone {
			continue
		}
		actions = append(actions, executionLoopActionForStep(step, report))
	}
	if len(actions) == 0 && packet.FinalAction == DecisionActionPrompt {
		actions = append(actions, executionLoopActionForStep(DecisionStep{
			Kind:            DecisionStepKindApproval,
			Action:          DecisionActionPrompt,
			Reason:          "execution_loop_prompt_requires_host_approval",
			DecisionSubject: packet.Subject,
			PolicySource:    RuntimePolicySourceApprovalHook,
			ControlSource:   RuntimeControlSourceExecutionContract,
		}, report))
	}
	return dedupeExecutionLoopActions(actions)
}

func executionLoopActionForStep(step DecisionStep, report ExecutionLoopReport) ExecutionLoopAction {
	kind := executionLoopActionKindForStep(step)
	action := ExecutionLoopAction{
		Kind:            kind,
		Required:        true,
		StepKind:        step.Kind,
		StepAction:      step.Action,
		Tool:            step.Tool,
		Reason:          step.Reason,
		DecisionSubject: step.DecisionSubject,
		PolicySource:    step.PolicySource,
		ControlSource:   step.ControlSource,
		NextHostAction:  executionLoopActionNextHostAction(kind),
	}
	switch kind {
	case ExecutionLoopActionDeniedReadReview:
		action.ActionRef = report.DeniedReadReviewRef
	case ExecutionLoopActionSandboxEscalation:
		action.ActionRef = report.SandboxEscalationRef
	case ExecutionLoopActionNetworkApproval:
		action.ActionRef = report.NetworkApprovalRef
	case ExecutionLoopActionHostApproval:
		action.ActionRef = report.HostApprovalRef
	default:
		action.ActionRef = report.ExecutionReviewRef
	}
	action.Ready = strings.TrimSpace(action.ActionRef) != ""
	return action
}

func executionLoopActionKindForStep(step DecisionStep) string {
	if executionLoopIsDeniedReadStep(step) {
		return ExecutionLoopActionDeniedReadReview
	}
	switch NormalizeDecisionStepKind(step.Kind) {
	case DecisionStepKindSandbox:
		return ExecutionLoopActionSandboxEscalation
	case DecisionStepKindNetwork:
		return ExecutionLoopActionNetworkApproval
	case DecisionStepKindApproval:
		return ExecutionLoopActionHostApproval
	default:
		return ExecutionLoopActionExecutionReview
	}
}

func executionLoopIsDeniedReadStep(step DecisionStep) bool {
	text := strings.ToLower(strings.TrimSpace(step.Tool + " " + step.RuntimeAction + " " + step.Reason + " " + step.DecisionSubject + " " + step.TargetKind))
	if !strings.Contains(text, "read") {
		return false
	}
	return step.Action == DecisionActionForbidden ||
		strings.Contains(text, "denied") ||
		strings.Contains(text, "forbidden") ||
		strings.Contains(text, "not_allowed")
}

func executionLoopApplyActions(report ExecutionLoopReport) ExecutionLoopReport {
	for _, action := range report.Actions {
		switch action.Kind {
		case ExecutionLoopActionDeniedReadReview:
			report.RequiresDeniedReadReview = true
			report.ReadyForDeniedReadReview = report.ReadyForDeniedReadReview || action.Ready
			if !action.Ready {
				report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_denied_read_review_ref")
				report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_denied_read_review_required")
				report.NextHostAction = "provide_denied_read_review_ref"
			}
		case ExecutionLoopActionSandboxEscalation:
			report.RequiresSandboxEscalation = true
			report.ReadyForSandboxEscalation = report.ReadyForSandboxEscalation || action.Ready
			if !action.Ready {
				report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_sandbox_escalation_ref")
				report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_sandbox_escalation_required")
				report.NextHostAction = "provide_sandbox_escalation_ref"
			}
		case ExecutionLoopActionNetworkApproval:
			report.RequiresNetworkApproval = true
			report.ReadyForNetworkApproval = report.ReadyForNetworkApproval || action.Ready
			if !action.Ready {
				report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_network_approval_ref")
				report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_network_approval_required")
				report.NextHostAction = "provide_network_approval_ref"
			}
		case ExecutionLoopActionHostApproval:
			report.RequiresHostApproval = true
			report.ReadyForHostApproval = report.ReadyForHostApproval || action.Ready
			if !action.Ready {
				report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_host_approval_ref")
				report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_host_approval_required")
				report.NextHostAction = "provide_host_approval_ref"
			}
		default:
			report.RequiresExecutionReview = true
			report.ReadyForExecutionReview = report.ReadyForExecutionReview || action.Ready
			if !action.Ready {
				report.MissingInputs = executionLoopAppendUnique(report.MissingInputs, "host:execution_loop_execution_review_ref")
				report.BlockedReasons = executionLoopAppendUnique(report.BlockedReasons, "execution_loop_execution_review_required")
				report.NextHostAction = "provide_execution_review_ref"
			}
		}
		report.Boundaries = executionLoopAppendUnique(report.Boundaries, "execution_loop_action_"+action.Kind)
	}
	return report
}

func executionLoopActionNextHostAction(kind string) string {
	switch kind {
	case ExecutionLoopActionDeniedReadReview:
		return "review_denied_read_and_retry_with_new_decision_packet"
	case ExecutionLoopActionSandboxEscalation:
		return "request_sandbox_escalation_and_retry_with_new_decision_packet"
	case ExecutionLoopActionNetworkApproval:
		return "request_network_approval_and_retry_with_new_decision_packet"
	case ExecutionLoopActionHostApproval:
		return "request_host_approval_and_retry_with_new_decision_packet"
	default:
		return "review_execution_decision_and_retry_with_new_decision_packet"
	}
}

func executionLoopActionsReady(actions []ExecutionLoopAction) bool {
	if len(actions) == 0 {
		return false
	}
	for _, action := range actions {
		if action.Required && !action.Ready {
			return false
		}
	}
	return true
}

func normalizeExecutionLoopActions(actions []ExecutionLoopAction) []ExecutionLoopAction {
	out := make([]ExecutionLoopAction, 0, len(actions))
	for _, action := range actions {
		action.Kind = strings.TrimSpace(action.Kind)
		action.StepKind = NormalizeDecisionStepKind(action.StepKind)
		action.StepAction = NormalizeDecisionAction(action.StepAction)
		action.Tool = strings.ToLower(strings.TrimSpace(action.Tool))
		action.Reason = strings.TrimSpace(action.Reason)
		action.DecisionSubject = strings.TrimSpace(action.DecisionSubject)
		action.PolicySource = strings.TrimSpace(action.PolicySource)
		action.ControlSource = strings.TrimSpace(action.ControlSource)
		action.ActionRef = strings.TrimSpace(action.ActionRef)
		action.NextHostAction = executionLoopNextHostAction(action.NextHostAction, executionLoopActionNextHostAction(action.Kind))
		action.Ready = strings.TrimSpace(action.ActionRef) != ""
		if action.Kind == "" {
			continue
		}
		out = append(out, action)
	}
	return out
}

func dedupeExecutionLoopActions(actions []ExecutionLoopAction) []ExecutionLoopAction {
	out := make([]ExecutionLoopAction, 0, len(actions))
	seen := map[string]struct{}{}
	for _, action := range normalizeExecutionLoopActions(actions) {
		key := strings.Join([]string{
			action.Kind,
			string(action.StepKind),
			string(action.StepAction),
			action.Tool,
			action.Reason,
			action.DecisionSubject,
			action.ActionRef,
		}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, action)
	}
	return out
}

func normalizeExecutionLoopReportContract(contract ExecutionLoopReportContract) ExecutionLoopReportContract {
	out := contract
	if strings.TrimSpace(out.Version) == "" {
		out.Version = ExecutionLoopReportContractVersion
	}
	if strings.TrimSpace(out.Owner) == "" {
		out.Owner = "execution"
	}
	if strings.TrimSpace(out.HostAdapterOwner) == "" {
		out.HostAdapterOwner = "hostruntime"
	}
	if strings.TrimSpace(out.ExecutionPolicyOwner) == "" {
		out.ExecutionPolicyOwner = "execution"
	}
	out.ObservationOnly = true
	out.PolicySource = false
	out.AuthorizationSource = false
	out.RuntimeInvocationSource = false
	return out
}

func normalizeExecutionLoopDecisionPacket(packet DecisionPacket) DecisionPacket {
	if !executionLoopDecisionPacketAvailable(packet) {
		return DecisionPacket{}
	}
	return NewDecisionPacket(DecisionPacketInput{
		ContractID: packet.ContractID,
		Source:     packet.Source,
		Subject:    packet.Subject,
		Steps:      packet.Steps,
	})
}

func executionLoopDecisionPacketAvailable(packet DecisionPacket) bool {
	return strings.TrimSpace(packet.SchemaVersion) != "" ||
		strings.TrimSpace(packet.ContractID) != "" ||
		strings.TrimSpace(packet.Source) != "" ||
		strings.TrimSpace(packet.Subject) != "" ||
		len(packet.Steps) > 0
}

func executionLoopRefWithFallback(value, prefix, token string, raw *bool) string {
	if strings.TrimSpace(value) != "" {
		return executionLoopRef(value, raw)
	}
	return executionLoopDerivedRef(prefix, token)
}

func executionLoopRef(value string, raw *bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !executionLoopDisplaySafeRef(value) {
		if raw != nil {
			*raw = true
		}
		return ""
	}
	return value
}

func executionLoopDerivedRef(prefix, token string) string {
	prefix = strings.TrimSpace(prefix)
	token = strings.TrimSpace(token)
	if prefix == "" {
		return ""
	}
	if token == "" {
		token = "execution_loop"
	}
	ref := prefix + ":" + token
	if !executionLoopDisplaySafeRef(ref) {
		return ""
	}
	return ref
}

var (
	executionLoopDisplaySafeRefPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,127}$`)
	executionLoopURLPattern            = regexp.MustCompile(`(?i)\b[a-z][a-z0-9+.-]*://`)
	executionLoopUnixPathPattern       = regexp.MustCompile(`(^|[\s"'(])/(Users|home|var|etc|tmp|private|opt|usr|Volumes)/[^\s"' )]+`)
	executionLoopWindowsPathPattern    = regexp.MustCompile(`(?i)\b[A-Z]:\\[^\s]+`)
	executionLoopSecretPattern         = regexp.MustCompile(`(?i)\b(api[_-]?key|access[_-]?token|refresh[_-]?token|auth[_-]?token|password|passwd|secret|credential)\s*[:=]\s*\S+`)
	executionLoopPEMPattern            = regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----`)
)

func executionLoopDisplaySafeRef(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" &&
		!executionLoopContainsUnsafeRawOutput(value) &&
		executionLoopDisplaySafeRefPattern.MatchString(value)
}

func executionLoopContainsUnsafeRawOutput(value string) bool {
	return executionLoopURLPattern.MatchString(value) ||
		executionLoopUnixPathPattern.MatchString(value) ||
		executionLoopWindowsPathPattern.MatchString(value) ||
		executionLoopSecretPattern.MatchString(value) ||
		executionLoopPEMPattern.MatchString(value)
}

func executionLoopDecisionPacketToken(packet DecisionPacket) string {
	packet = normalizeExecutionLoopDecisionPacket(packet)
	h := sha256.New()
	executionLoopHashWrite(h, packet.SchemaVersion)
	executionLoopHashWrite(h, packet.ContractID)
	executionLoopHashWrite(h, packet.Source)
	executionLoopHashWrite(h, packet.Subject)
	executionLoopHashWrite(h, string(packet.FinalAction))
	for _, step := range packet.Steps {
		step = NormalizeDecisionStep(step)
		executionLoopHashWrite(h, string(step.Kind))
		executionLoopHashWrite(h, string(step.Action))
		executionLoopHashWrite(h, step.Tool)
		executionLoopHashWrite(h, step.RuntimeAction)
		executionLoopHashWrite(h, step.Reason)
		executionLoopHashWrite(h, step.DecisionSubject)
		executionLoopHashWrite(h, step.TargetKind)
		executionLoopHashWrite(h, step.PolicySource)
		executionLoopHashWrite(h, step.ControlSource)
		executionLoopHashWrite(h, step.EnforcementSurface)
	}
	sum := h.Sum(nil)
	return "decision_" + hex.EncodeToString(sum[:8])
}

func executionLoopHashWrite(h interface{ Write([]byte) (int, error) }, value string) {
	_, _ = h.Write([]byte(strings.TrimSpace(value)))
	_, _ = h.Write([]byte{0})
}

func executionLoopRequireRef(inputs []string, ref string, missing string) []string {
	if strings.TrimSpace(ref) == "" {
		return executionLoopAppendUnique(inputs, missing)
	}
	return inputs
}

func executionLoopBoundaries(in []string, required ...string) []string {
	return executionLoopAppendUnique(executionLoopAppendUnique(nil, required...), in...)
}

func executionLoopNextHostAction(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func executionLoopAppendUnique(values []string, next ...string) []string {
	out := make([]string, 0, len(values)+len(next))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	for _, value := range next {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
