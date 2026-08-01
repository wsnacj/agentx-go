package executionpolicy

import "strings"

const (
	DecisionPacketSchemaV1 = "agentx.execution_decision_packet.v1"
)

type DecisionAction string

const (
	DecisionActionNone      DecisionAction = ""
	DecisionActionAllow     DecisionAction = "allow"
	DecisionActionPrompt    DecisionAction = "prompt"
	DecisionActionForbidden DecisionAction = "forbidden"
)

type DecisionStepKind string

const (
	DecisionStepKindRuntime  DecisionStepKind = "runtime"
	DecisionStepKindApproval DecisionStepKind = "approval"
	DecisionStepKindSandbox  DecisionStepKind = "sandbox"
	DecisionStepKindNetwork  DecisionStepKind = "network"
	DecisionStepKindRetry    DecisionStepKind = "retry"
	DecisionStepKindAudit    DecisionStepKind = "audit"
)

type DecisionPacketInput struct {
	ContractID       string            `json:"contract_id,omitempty"`
	Source           string            `json:"source,omitempty"`
	Subject          string            `json:"subject,omitempty"`
	RuntimeDecisions []RuntimeDecision `json:"runtime_decisions,omitempty"`
	Steps            []DecisionStep    `json:"steps,omitempty"`
}

type DecisionPacket struct {
	SchemaVersion string                `json:"schema_version"`
	ContractID    string                `json:"contract_id,omitempty"`
	Source        string                `json:"source,omitempty"`
	Subject       string                `json:"subject,omitempty"`
	FinalAction   DecisionAction        `json:"final_action,omitempty"`
	Steps         []DecisionStep        `json:"steps,omitempty"`
	Summary       DecisionPacketSummary `json:"summary,omitempty"`
	Audit         *DecisionAuditRecord  `json:"audit,omitempty"`
}

type DecisionStep struct {
	Kind               DecisionStepKind `json:"kind,omitempty"`
	Action             DecisionAction   `json:"action,omitempty"`
	Tool               string           `json:"tool,omitempty"`
	RuntimeAction      string           `json:"runtime_action,omitempty"`
	Reason             string           `json:"reason,omitempty"`
	Detail             string           `json:"detail,omitempty"`
	DecisionSubject    string           `json:"decision_subject,omitempty"`
	TargetKind         string           `json:"target_kind,omitempty"`
	PolicySource       string           `json:"policy_source,omitempty"`
	ControlSource      string           `json:"control_source,omitempty"`
	EnforcementSurface string           `json:"enforcement_surface,omitempty"`
	Degraded           bool             `json:"degraded,omitempty"`
}

type DecisionPacketSummary struct {
	Steps          int            `json:"steps,omitempty"`
	Allowed        int            `json:"allowed,omitempty"`
	Prompted       int            `json:"prompted,omitempty"`
	Forbidden      int            `json:"forbidden,omitempty"`
	Degraded       int            `json:"degraded,omitempty"`
	ByKind         map[string]int `json:"by_kind,omitempty"`
	ByAction       map[string]int `json:"by_action,omitempty"`
	ByPolicySource map[string]int `json:"by_policy_source,omitempty"`
	ByReason       map[string]int `json:"by_reason,omitempty"`
	BySubject      map[string]int `json:"by_subject,omitempty"`
	ByTool         map[string]int `json:"by_tool,omitempty"`
}

type DecisionAuditRecord struct {
	ContractID     string         `json:"contract_id,omitempty"`
	Source         string         `json:"source,omitempty"`
	Subject        string         `json:"subject,omitempty"`
	FinalAction    DecisionAction `json:"final_action,omitempty"`
	Reasons        []string       `json:"reasons,omitempty"`
	PolicySources  []string       `json:"policy_sources,omitempty"`
	ControlSources []string       `json:"control_sources,omitempty"`
	Tools          []string       `json:"tools,omitempty"`
}

func NewDecisionPacket(input DecisionPacketInput) DecisionPacket {
	packet := DecisionPacket{
		SchemaVersion: DecisionPacketSchemaV1,
		ContractID:    strings.TrimSpace(input.ContractID),
		Source:        strings.TrimSpace(input.Source),
		Subject:       strings.TrimSpace(input.Subject),
	}
	for _, decision := range input.RuntimeDecisions {
		step := DecisionStepFromRuntimeDecision(decision)
		if isDecisionStepEmpty(step) {
			continue
		}
		packet.Steps = append(packet.Steps, step)
	}
	for _, step := range input.Steps {
		step = NormalizeDecisionStep(step)
		if isDecisionStepEmpty(step) {
			continue
		}
		packet.Steps = append(packet.Steps, step)
	}
	packet.Summary = SummarizeDecisionSteps(packet.Steps)
	packet.FinalAction = finalDecisionAction(packet.Steps)
	packet.Audit = buildDecisionAuditRecord(packet)
	return packet
}

func DecisionStepFromRuntimeDecision(decision RuntimeDecision) DecisionStep {
	return DecisionStepFromRuntimeDecisionKind(DecisionStepKindForRuntimeDecision(decision), decision)
}

func DecisionStepFromRuntimeDecisionKind(kind DecisionStepKind, decision RuntimeDecision) DecisionStep {
	if !decision.Checked {
		return DecisionStep{}
	}
	return NormalizeDecisionStep(DecisionStep{
		Kind:               kind,
		Action:             DecisionActionFromRuntimeDecision(decision),
		Tool:               decision.Tool,
		RuntimeAction:      decision.Action,
		Reason:             decision.Reason,
		Detail:             decision.Detail,
		DecisionSubject:    decision.DecisionSubject,
		TargetKind:         decision.TargetKind,
		PolicySource:       decision.PolicySource,
		ControlSource:      decision.ControlSource,
		EnforcementSurface: decision.EnforcementSurface,
		Degraded:           decision.Degraded,
	})
}

func DecisionActionFromRuntimeDecision(decision RuntimeDecision) DecisionAction {
	if !decision.Checked {
		return DecisionActionNone
	}
	if decision.Denied {
		return DecisionActionForbidden
	}
	if decision.RequiresConfirm {
		return DecisionActionPrompt
	}
	if decision.Allowed || decision.Degraded {
		return DecisionActionAllow
	}
	return DecisionActionNone
}

func DecisionStepKindForRuntimeDecision(decision RuntimeDecision) DecisionStepKind {
	if !decision.Checked {
		return ""
	}
	policySource := strings.ToLower(strings.TrimSpace(decision.PolicySource))
	reason := strings.ToLower(strings.TrimSpace(decision.Reason))
	subject := strings.ToLower(strings.TrimSpace(decision.DecisionSubject))
	targetKind := strings.ToLower(strings.TrimSpace(decision.TargetKind))
	tool := strings.ToLower(strings.TrimSpace(decision.Tool))
	runtimeAction := strings.ToLower(strings.TrimSpace(decision.Action))
	if policySource == RuntimePolicySourceApprovalHook || strings.Contains(reason, "approval") {
		return DecisionStepKindApproval
	}
	if targetKind == "url" ||
		targetKind == "endpoint" ||
		strings.Contains(subject, "url") ||
		strings.Contains(subject, "network") ||
		strings.Contains(subject, "gateway_route") ||
		strings.Contains(subject, "browser_proxy_route") {
		return DecisionStepKindNetwork
	}
	if tool == "exec" ||
		tool == "process" ||
		runtimeAction == "exec" ||
		runtimeAction == "kill" ||
		targetKind == "runtime_target" ||
		strings.Contains(subject, "command") ||
		strings.Contains(subject, "workdir") ||
		strings.Contains(subject, "metadata") ||
		strings.Contains(reason, "metadata") ||
		strings.Contains(subject, "sandbox") {
		return DecisionStepKindSandbox
	}
	return DecisionStepKindRuntime
}

func NormalizeDecisionStep(step DecisionStep) DecisionStep {
	step.Kind = NormalizeDecisionStepKind(step.Kind)
	step.Action = NormalizeDecisionAction(step.Action)
	step.Tool = strings.ToLower(strings.TrimSpace(step.Tool))
	step.RuntimeAction = strings.ToLower(strings.TrimSpace(step.RuntimeAction))
	step.Reason = strings.TrimSpace(step.Reason)
	step.Detail = strings.TrimSpace(step.Detail)
	step.DecisionSubject = strings.TrimSpace(step.DecisionSubject)
	step.TargetKind = strings.ToLower(strings.TrimSpace(step.TargetKind))
	step.PolicySource = strings.TrimSpace(step.PolicySource)
	step.ControlSource = strings.TrimSpace(step.ControlSource)
	step.EnforcementSurface = strings.TrimSpace(step.EnforcementSurface)
	return step
}

func NormalizeDecisionAction(action DecisionAction) DecisionAction {
	switch strings.ToLower(strings.TrimSpace(string(action))) {
	case string(DecisionActionAllow):
		return DecisionActionAllow
	case string(DecisionActionPrompt):
		return DecisionActionPrompt
	case string(DecisionActionForbidden):
		return DecisionActionForbidden
	default:
		return DecisionActionNone
	}
}

func NormalizeDecisionStepKind(kind DecisionStepKind) DecisionStepKind {
	switch strings.ToLower(strings.TrimSpace(string(kind))) {
	case string(DecisionStepKindApproval):
		return DecisionStepKindApproval
	case string(DecisionStepKindSandbox):
		return DecisionStepKindSandbox
	case string(DecisionStepKindNetwork):
		return DecisionStepKindNetwork
	case string(DecisionStepKindRetry):
		return DecisionStepKindRetry
	case string(DecisionStepKindAudit):
		return DecisionStepKindAudit
	case "", string(DecisionStepKindRuntime):
		return DecisionStepKindRuntime
	default:
		return DecisionStepKindRuntime
	}
}

func SummarizeDecisionSteps(steps []DecisionStep) DecisionPacketSummary {
	var summary DecisionPacketSummary
	for _, step := range steps {
		step = NormalizeDecisionStep(step)
		if isDecisionStepEmpty(step) {
			continue
		}
		summary.Steps++
		switch step.Action {
		case DecisionActionAllow:
			summary.Allowed++
		case DecisionActionPrompt:
			summary.Prompted++
		case DecisionActionForbidden:
			summary.Forbidden++
		}
		if step.Degraded {
			summary.Degraded++
		}
		incrementDecisionPacketCount(&summary.ByKind, string(step.Kind))
		incrementDecisionPacketCount(&summary.ByAction, string(step.Action))
		incrementDecisionPacketCount(&summary.ByPolicySource, step.PolicySource)
		incrementDecisionPacketCount(&summary.ByReason, step.Reason)
		incrementDecisionPacketCount(&summary.BySubject, step.DecisionSubject)
		incrementDecisionPacketCount(&summary.ByTool, step.Tool)
	}
	return summary
}

func finalDecisionAction(steps []DecisionStep) DecisionAction {
	final := DecisionActionNone
	for _, step := range steps {
		switch NormalizeDecisionAction(step.Action) {
		case DecisionActionForbidden:
			return DecisionActionForbidden
		case DecisionActionPrompt:
			if final != DecisionActionForbidden {
				final = DecisionActionPrompt
			}
		case DecisionActionAllow:
			if final == DecisionActionNone {
				final = DecisionActionAllow
			}
		}
	}
	return final
}

func buildDecisionAuditRecord(packet DecisionPacket) *DecisionAuditRecord {
	if len(packet.Steps) == 0 &&
		strings.TrimSpace(packet.ContractID) == "" &&
		strings.TrimSpace(packet.Source) == "" &&
		strings.TrimSpace(packet.Subject) == "" {
		return nil
	}
	record := DecisionAuditRecord{
		ContractID:  strings.TrimSpace(packet.ContractID),
		Source:      strings.TrimSpace(packet.Source),
		Subject:     strings.TrimSpace(packet.Subject),
		FinalAction: NormalizeDecisionAction(packet.FinalAction),
	}
	for _, step := range packet.Steps {
		step = NormalizeDecisionStep(step)
		record.Reasons = appendUniqueDecisionValue(record.Reasons, step.Reason)
		record.PolicySources = appendUniqueDecisionValue(record.PolicySources, step.PolicySource)
		record.ControlSources = appendUniqueDecisionValue(record.ControlSources, step.ControlSource)
		record.Tools = appendUniqueDecisionValue(record.Tools, step.Tool)
	}
	if len(record.Reasons) == 0 {
		record.Reasons = nil
	}
	if len(record.PolicySources) == 0 {
		record.PolicySources = nil
	}
	if len(record.ControlSources) == 0 {
		record.ControlSources = nil
	}
	if len(record.Tools) == 0 {
		record.Tools = nil
	}
	return &record
}

func incrementDecisionPacketCount(target *map[string]int, key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if *target == nil {
		*target = map[string]int{}
	}
	(*target)[key]++
}

func appendUniqueDecisionValue(values []string, value string) []string {
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

func isDecisionStepEmpty(step DecisionStep) bool {
	step = NormalizeDecisionStep(step)
	return step.Action == DecisionActionNone &&
		step.Tool == "" &&
		step.RuntimeAction == "" &&
		step.Reason == "" &&
		step.Detail == "" &&
		step.DecisionSubject == "" &&
		step.TargetKind == "" &&
		step.PolicySource == "" &&
		step.ControlSource == "" &&
		step.EnforcementSurface == "" &&
		!step.Degraded
}
