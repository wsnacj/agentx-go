package publicsource

import (
	"context"
	"strings"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

// Coordinator owns one exact Collector invocation and generic AgentX projection.
type Coordinator struct {
	Collector  Collector
	Descriptor control.ProductionAdapterDescriptor
	Now        func() time.Time
}

type Execution struct {
	Result control.RuntimeAdapterExecutionResult `json:"result"`
	Report Report                                `json:"report"`
}

func NewCoordinator(collector Collector) Coordinator { return Coordinator{Collector: collector} }

func Descriptor() control.ProductionAdapterDescriptor {
	return control.ProductionAdapterDescriptor{
		AdapterRef: DefaultAdapterRef, Owner: "host", OwnerRef: "host:public_source", Version: "v1",
		Kind: control.ProductionAdapterSourceApply, SupportedCandidateRefs: []control.DisplaySafeRef{DefaultStrategyRef},
		ProvidesCapabilityRefs: []control.DisplaySafeRef{"capability:public_source_retrieval"}, RequiresCapabilityRefs: []control.DisplaySafeRef{"capability:public_source_retrieval"},
		InputContractRef: "contract:public_source_query_ref", OutputContractRef: "contract:public_source_evidence_and_display_summary_refs", ReadbackContractRef: "contract:public_source_retrieval_readback",
		RequiredPolicyRefs: []control.DisplaySafeRef{"policy:public_source_readonly"}, IdempotencyContractRef: "contract:public_source_retrieval_idempotency",
		RiskRef: "risk:public_web_readonly", SideEffectClass: "read_only", TimeoutPolicyRef: "timeout:public_source_retrieval", RedactionPolicyRef: "redaction:display_safe_refs_only",
		DisplaySafeInputRefs: []control.DisplaySafeRef{"input:public_source_query_ref"}, DisplaySafeOutputRefs: []control.DisplaySafeRef{"output:public_source_evidence_refs", "output:public_source_display_summary_refs"},
		Boundaries: []control.Boundary{"public_source_runtime_adapter_descriptor", "public_source_readonly", "no_transactional_side_effects"},
	}.Normalize()
}

func Registry() control.HostAdapterRegistrySnapshot {
	return control.BuildHostAdapterRegistry(control.HostAdapterRegistryInput{RegistryRef: "registry:public_source_runtime_adapters", Descriptors: []control.ProductionAdapterDescriptor{Descriptor()}, PolicyRefs: []control.DisplaySafeRef{"policy:intensity", "contract:execution", "policy:public_source_readonly"}})
}

func Strategy(required []control.EvidenceRef) control.StrategyCandidate {
	expected := control.MergeEvidenceRefs(required)
	if len(expected) == 0 {
		expected = []control.EvidenceRef{{Kind: "public_source_result", Strength: control.EvidenceAdequate, Ref: DefaultEvidenceRef}}
	}
	return control.StrategyCandidate{ID: string(DefaultStrategyRef), Kind: "host_adapter", ControlMode: control.ControlModeObjective, MinIntensity: control.IntensityL3ManagedObjective, MaxIntensity: control.IntensityL3ManagedObjective, CapabilityRefs: []control.DisplaySafeRef{"capability:public_source_retrieval"}, ExpectedEvidence: expected, Risk: "read_only", SideEffectClass: "read_only", Owner: "host", Boundaries: []control.Boundary{"public_source_strategy_metadata", "source_refs_required", "read_only_source_strategy"}}
}

func (coordinator Coordinator) Execute(ctx context.Context, request control.RuntimeAdapterExecutionRequest) Execution {
	descriptor := coordinator.Descriptor.Normalize()
	if descriptor.AdapterRef == "" {
		descriptor = Descriptor()
	}
	request = request.Normalize()
	boundaries := []control.Boundary{"public_source_runtime_adapter", "host_owned_public_source_adapter", "read_only_public_source_retrieval", "display_safe_refs_only", "no_transactional_side_effects"}
	if !request.ReadyForHostExecution {
		return blocked(request, request.AdapterRef, request.StrategyRef, control.FailureConfigMissing, "public_source_request_not_ready", append(boundaries, "public_source_request_not_ready"))
	}
	if request.AdapterRef != descriptor.AdapterRef {
		return blocked(request, descriptor.AdapterRef, request.StrategyRef, control.FailureUnsupportedOperation, "public_source_adapter_ref_mismatch", append(boundaries, "public_source_adapter_ref_mismatch"))
	}
	if !containsRef(descriptor.SupportedCandidateRefs, request.StrategyRef) {
		return blocked(request, request.AdapterRef, firstRef(descriptor.SupportedCandidateRefs...), control.FailureUnsupportedOperation, "public_source_strategy_ref_mismatch", append(boundaries, "public_source_strategy_ref_mismatch"))
	}
	if coordinator.Collector == nil {
		return blocked(request, request.AdapterRef, request.StrategyRef, control.FailureHostAdapterMissing, "public_source_collector_missing", append(boundaries, "public_source_collector_missing"))
	}
	now := coordinator.Now
	if now == nil {
		now = time.Now
	}
	domainRequest := Request{RuntimeRequest: request, QueryRefs: queryRefs(request.InputRefs), SourceRefs: sourceRefs(request.Frame.SourceContext), PolicyRefs: refs(request.PolicyRefs), ObservedAt: now().UTC().Format(time.RFC3339), Boundaries: append([]control.Boundary(nil), boundaries...)}
	report, err := coordinator.Collector.CollectPublicSourceEvidence(ctx, domainRequest)
	if err != nil {
		return blocked(request, request.AdapterRef, request.StrategyRef, control.FailureExternalDependencyUnavailable, "public_source_collector_error", append(boundaries, "public_source_collector_error"))
	}
	report = report.Normalize()
	status, failure, reason := executionStatus(report)
	return Execution{Report: report, Result: control.BuildRuntimeAdapterExecutionResult(control.RuntimeAdapterExecutionResultInput{
		Request: request, AdapterRef: request.AdapterRef, StrategyRef: request.StrategyRef, HostAdapterRunRef: runRef(request),
		Status: status, FailureClass: failure, FailureReason: reason, Observations: observations(request, report), EvidenceRefs: evidenceRefs(report),
		OutputRefs: descriptor.DisplaySafeOutputRefs, MissingInputs: missingInputs(report), Boundaries: control.AppendBoundaries(boundaries, report.Boundaries...), RawOutputLoaded: report.RawOutputLoaded,
	})}
}

func blocked(request control.RuntimeAdapterExecutionRequest, adapterRef, strategyRef control.DisplaySafeRef, failure control.FailureClass, reason string, boundaries []control.Boundary) Execution {
	return Execution{Result: control.BuildRuntimeAdapterExecutionResult(control.RuntimeAdapterExecutionResultInput{Request: request, AdapterRef: adapterRef, StrategyRef: strategyRef, Status: control.VerificationBlocked, FailureClass: failure, FailureReason: reason, Boundaries: boundaries})}
}

func executionStatus(report Report) (control.VerificationStatus, control.FailureClass, string) {
	report = report.Normalize()
	if report.Status == control.VerificationSatisfied && len(report.Evidence) > 0 {
		return control.VerificationSatisfied, control.FailureNone, ""
	}
	reason := report.FailureReason
	if reason == "" && len(report.UnavailableReasons) > 0 {
		reason = report.UnavailableReasons[0]
	}
	return report.Status, report.FailureClass, first(reason, "public_source_unavailable")
}

func observations(request control.RuntimeAdapterExecutionRequest, report Report) []control.Observation {
	report = report.Normalize()
	out := make([]control.Observation, 0, len(report.Evidence)+len(report.DisplaySummaries)+1)
	for _, item := range report.Evidence {
		evidence := control.EvidenceRef{Ref: item.EvidenceRef, Kind: first(item.Kind, "public_source_result"), Strength: item.Strength, Source: runRef(request), ObservedAt: item.ObservedAt}
		out = append(out, control.Observation{Kind: first(item.Kind, "public_source_result"), Source: item.SourceRef, Subject: objectiveRef(request), Name: "public_source_result", Value: string(item.DisplaySafeRef), Unit: item.Confidence, Strength: item.Strength, ObservedAt: item.ObservedAt, EvidenceRefs: []control.EvidenceRef{evidence}, DisplaySafeRefs: refs([]control.DisplaySafeRef{item.SourceRef, item.QueryRef, item.EvidenceRef, item.ConfidenceRef, item.DisplaySafeRef})})
	}
	for _, item := range report.DisplaySummaries {
		strength := summaryStrength(item)
		evidence := control.EvidenceRef{Ref: item.EvidenceRef, Kind: "public_source_display_summary", Strength: strength, Source: runRef(request), ObservedAt: item.ObservedAt}
		out = append(out, control.Observation{Kind: "public_source_display_summary", Source: item.SourceRef, Subject: objectiveRef(request), Name: "public_source_display_summary", Value: string(item.SummaryRef), Unit: item.Confidence, Strength: strength, ObservedAt: item.ObservedAt, EvidenceRefs: []control.EvidenceRef{evidence}, DisplaySafeRefs: refs([]control.DisplaySafeRef{item.SourceRef, item.QueryRef, item.EvidenceRef, item.SummaryRef, item.AttestationRef, item.RedactionRef, item.DisplayPolicyRef})})
	}
	if len(out) == 0 && len(report.UnavailableReasons) > 0 {
		out = append(out, control.Observation{Kind: "public_source_unavailable", Source: runRef(request), Subject: objectiveRef(request), Name: "public_source_unavailable", Value: report.UnavailableReasons[0], Strength: control.EvidenceMissing, ObservedAt: report.ObservedAt, DisplaySafeRefs: []control.DisplaySafeRef{runRef(request)}, DegradationReason: report.UnavailableReasons[0]})
	}
	return out
}

func evidenceRefs(report Report) []control.EvidenceRef {
	report = report.Normalize()
	out := []control.EvidenceRef{}
	for _, ref := range report.QueryEvidenceRefs {
		out = append(out, control.EvidenceRef{Ref: ref, Kind: "public_source_query", Strength: control.EvidenceAdequate, Source: DefaultRunRef, ObservedAt: report.ObservedAt})
	}
	for _, item := range report.Evidence {
		out = append(out, control.EvidenceRef{Ref: item.EvidenceRef, Kind: first(item.Kind, "public_source_result"), Strength: item.Strength, Source: DefaultRunRef, ObservedAt: item.ObservedAt})
	}
	for _, item := range report.DisplaySummaries {
		out = append(out, control.EvidenceRef{Ref: item.EvidenceRef, Kind: "public_source_display_summary", Strength: summaryStrength(item), Source: DefaultRunRef, ObservedAt: item.ObservedAt})
	}
	return control.MergeEvidenceRefs(out)
}

func missingInputs(report Report) []control.MissingInput {
	if report.Normalize().Status == control.VerificationSatisfied {
		return nil
	}
	return []control.MissingInput{"host:public_source_evidence"}
}
func queryRefs(values []control.DisplaySafeRef) []control.DisplaySafeRef {
	out := []control.DisplaySafeRef{}
	for _, ref := range refs(values) {
		if strings.HasPrefix(string(ref), "query:") || strings.HasPrefix(string(ref), "input:") {
			out = appendRef(out, ref)
		}
	}
	if len(out) == 0 {
		return refs(values)
	}
	return out
}
func sourceRefs(values []control.DisplaySafeRef) []control.DisplaySafeRef {
	out := []control.DisplaySafeRef{}
	for _, ref := range refs(values) {
		if strings.HasPrefix(string(ref), "source:") || strings.HasPrefix(string(ref), "host_source:") {
			out = appendRef(out, ref)
		}
	}
	return out
}
func runRef(request control.RuntimeAdapterExecutionRequest) control.DisplaySafeRef {
	if request.Frame.ID != "" {
		return makeRef("adapter_run:public_source", request.Frame.ID+":"+string(request.StrategyRef))
	}
	return DefaultRunRef
}
func objectiveRef(request control.RuntimeAdapterExecutionRequest) control.DisplaySafeRef {
	if ref, ok := control.NormalizeDisplaySafeRef(request.Frame.ID); ok {
		return ref
	}
	return "objective:public_source"
}
