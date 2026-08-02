package publictransport

import (
	"context"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

// Coordinator owns provider-neutral dispatch from an AgentX runtime request to
// one host-supplied Collector. It never retries, chooses providers, or performs I/O itself.
type Coordinator struct {
	Collector  Collector
	Descriptor control.ProductionAdapterDescriptor
	Now        func() time.Time
}

// Execution keeps both the generic AgentX projection and the domain readback.
type Execution struct {
	Result control.RuntimeAdapterExecutionResult `json:"result"`
	Report Report                                `json:"report"`
}

// NewCoordinator constructs a coordinator around an explicit host Collector.
func NewCoordinator(collector Collector) Coordinator {
	return Coordinator{Collector: collector}
}

// Descriptor returns the canonical read-only adapter descriptor.
func Descriptor() control.ProductionAdapterDescriptor {
	return control.ProductionAdapterDescriptor{
		AdapterRef:             DefaultAdapterRef,
		Owner:                  "host",
		OwnerRef:               "host:public_transport",
		Version:                "v1",
		Kind:                   control.ProductionAdapterSourceApply,
		SupportedCandidateRefs: []control.DisplaySafeRef{DefaultStrategyRef},
		ProvidesCapabilityRefs: []control.DisplaySafeRef{DefaultCapabilityRef},
		RequiresCapabilityRefs: []control.DisplaySafeRef{DefaultCapabilityRef},
		InputContractRef:       "contract:public_transport_ticket_query_refs",
		OutputContractRef:      "contract:official_ticket_inventory_evidence_refs",
		ReadbackContractRef:    "contract:public_transport_ticket_readback",
		RequiredPolicyRefs: []control.DisplaySafeRef{
			"policy:public_source_readonly",
			DefaultSourcePolicyRef,
		},
		IdempotencyContractRef: "contract:public_transport_ticket_lookup_idempotency",
		RiskRef:                "risk:public_transport_readonly",
		SideEffectClass:        "read_only",
		TimeoutPolicyRef:       "timeout:public_transport_ticket_lookup",
		RedactionPolicyRef:     "redaction:display_safe_refs_only",
		DisplaySafeInputRefs:   []control.DisplaySafeRef{"input:public_transport_ticket_query_refs"},
		DisplaySafeOutputRefs:  []control.DisplaySafeRef{"output:official_ticket_inventory_evidence_refs"},
		Boundaries: []control.Boundary{
			"public_transport_ticket_runtime_adapter_descriptor",
			"host_owned_public_transport_ticket_lookup_required",
			"official_ticket_inventory_evidence_required",
			"read_only_public_transport_ticket_lookup",
			"display_safe_refs_only",
			"no_core_ticket_lookup",
			"no_booking_attempt",
			"no_purchase_attempt",
		},
	}.Normalize()
}

// Registry returns a one-entry registry snapshot for explicit Host composition.
func Registry() control.HostAdapterRegistrySnapshot {
	return control.BuildHostAdapterRegistry(control.HostAdapterRegistryInput{
		RegistryRef: "registry:public_transport_ticket_runtime_adapters",
		Descriptors: []control.ProductionAdapterDescriptor{Descriptor()},
		PolicyRefs: []control.DisplaySafeRef{
			"policy:intensity",
			"contract:execution",
			"policy:public_source_readonly",
			DefaultSourcePolicyRef,
		},
	})
}

// Strategy returns the host-adapter strategy metadata for the supplied evidence contract.
func Strategy(requiredEvidence []control.EvidenceRef) control.StrategyCandidate {
	expected := control.MergeEvidenceRefs(requiredEvidence)
	if len(expected) == 0 {
		expected = []control.EvidenceRef{{
			Ref:      DefaultEvidenceRef,
			Kind:     "public_transport_ticket_inventory",
			Strength: control.EvidenceAdequate,
		}}
	}
	return control.StrategyCandidate{
		ID:               string(DefaultStrategyRef),
		Kind:             "host_adapter",
		ControlMode:      control.ControlModeObjective,
		MinIntensity:     control.IntensityL3ManagedObjective,
		MaxIntensity:     control.IntensityL3ManagedObjective,
		CapabilityRefs:   []control.DisplaySafeRef{DefaultCapabilityRef},
		ExpectedEvidence: expected,
		Risk:             "read_only",
		SideEffectClass:  "read_only",
		Owner:            "host",
		Boundaries: []control.Boundary{
			"public_transport_ticket_strategy_metadata",
			"official_ticket_inventory_evidence_required",
			"no_core_ticket_lookup",
			"no_booking_or_purchase_strategy",
		},
	}
}

// InventoryEvidenceRef derives a deterministic display-safe evidence identity.
func InventoryEvidenceRef(queryRef control.DisplaySafeRef) control.DisplaySafeRef {
	queryRef = normalizeRef(queryRef)
	if queryRef == "" {
		return DefaultEvidenceRef
	}
	return makeRef("evidence:public_transport_ticket_inventory", string(queryRef))
}

// Execute coordinates one exact Collector invocation.
func (c Coordinator) Execute(ctx context.Context, request control.RuntimeAdapterExecutionRequest) Execution {
	descriptor := c.Descriptor.Normalize()
	if descriptor.AdapterRef == "" {
		descriptor = Descriptor()
	}
	request = request.Normalize()
	boundaries := []control.Boundary{
		"public_transport_ticket_runtime_adapter",
		"host_owned_public_transport_ticket_adapter",
		"read_only_public_transport_ticket_lookup",
		"display_safe_refs_only",
		"no_core_ticket_lookup",
		"no_booking_attempt",
		"no_purchase_attempt",
	}
	if !request.ReadyForHostExecution {
		return blockedExecution(request, request.AdapterRef, request.StrategyRef, control.FailureConfigMissing, "public_transport_ticket_request_not_ready", append(boundaries, "public_transport_ticket_request_not_ready"))
	}
	if request.AdapterRef != descriptor.AdapterRef {
		return blockedExecution(request, descriptor.AdapterRef, request.StrategyRef, control.FailureUnsupportedOperation, "public_transport_ticket_adapter_ref_mismatch", append(boundaries, "public_transport_ticket_adapter_ref_mismatch"))
	}
	if !containsRef(descriptor.SupportedCandidateRefs, request.StrategyRef) {
		return blockedExecution(request, request.AdapterRef, firstRef(descriptor.SupportedCandidateRefs), control.FailureUnsupportedOperation, "public_transport_ticket_strategy_ref_mismatch", append(boundaries, "public_transport_ticket_strategy_ref_mismatch"))
	}
	if c.Collector == nil {
		return blockedExecution(request, request.AdapterRef, request.StrategyRef, control.FailureHostAdapterMissing, "public_transport_ticket_collector_missing", append(boundaries, "public_transport_ticket_collector_missing"))
	}
	now := c.Now
	if now == nil {
		now = time.Now
	}
	domainRequest := Request{
		RuntimeRequest: request,
		QueryRefs:      refsWithPrefixes(request.InputRefs, "query:", "input:"),
		RouteRefs:      refsWithPrefixes(request.InputRefs, "route:"),
		TravelDateRefs: refsWithPrefixes(request.InputRefs, "date:", "travel_date:"),
		SourceRefs:     refsWithPrefixes(request.Frame.SourceContext, "source:", "host_source:"),
		PolicyRefs:     append([]control.DisplaySafeRef(nil), request.PolicyRefs...),
		ObservedAt:     now().UTC().Format(time.RFC3339),
		Boundaries:     append([]control.Boundary(nil), boundaries...),
	}
	if len(domainRequest.QueryRefs) == 0 {
		domainRequest.QueryRefs = normalizeRefs(request.InputRefs)
	}
	report, err := c.Collector.CollectPublicTransportTicketEvidence(ctx, domainRequest)
	if err != nil {
		return blockedExecution(request, request.AdapterRef, request.StrategyRef, control.FailureExternalDependencyUnavailable, "public_transport_ticket_collector_error", append(boundaries, "public_transport_ticket_collector_error"))
	}
	report = report.Normalize()
	status, failure, reason := adapterStatus(report)
	return Execution{
		Report: report,
		Result: control.BuildRuntimeAdapterExecutionResult(control.RuntimeAdapterExecutionResultInput{
			Request:           request,
			AdapterRef:        request.AdapterRef,
			StrategyRef:       request.StrategyRef,
			HostAdapterRunRef: runtimeRunRef(request),
			Status:            status,
			FailureClass:      failure,
			FailureReason:     reason,
			Observations:      observations(request, report),
			EvidenceRefs:      evidenceRefs(request, report),
			OutputRefs:        descriptor.DisplaySafeOutputRefs,
			MissingInputs:     report.MissingInputs,
			Boundaries:        control.AppendBoundaries(boundaries, report.Boundaries...),
			RawOutputLoaded:   report.RawOutputLoaded,
		}),
	}
}

func blockedExecution(request control.RuntimeAdapterExecutionRequest, adapterRef, strategyRef control.DisplaySafeRef, failure control.FailureClass, reason string, boundaries []control.Boundary) Execution {
	return Execution{Result: control.BuildRuntimeAdapterExecutionResult(control.RuntimeAdapterExecutionResultInput{
		Request: request, AdapterRef: adapterRef, StrategyRef: strategyRef,
		Status: control.VerificationBlocked, FailureClass: failure, FailureReason: reason, Boundaries: boundaries,
	})}
}

func adapterStatus(report Report) (control.VerificationStatus, control.FailureClass, string) {
	report = report.Normalize()
	if report.Status == control.VerificationSatisfied && len(report.Evidence) > 0 {
		return control.VerificationSatisfied, control.FailureNone, ""
	}
	reason := report.FailureReason
	if reason == "" && len(report.UnavailableReasons) > 0 {
		reason = report.UnavailableReasons[0]
	}
	return report.Status, report.FailureClass, firstString(reason, "official_ticket_inventory_unavailable")
}

func observations(request control.RuntimeAdapterExecutionRequest, report Report) []control.Observation {
	report = report.Normalize()
	out := make([]control.Observation, 0, len(report.Evidence)+len(report.InventoryRows)+1)
	for _, item := range report.Evidence {
		evidence := control.EvidenceRef{Ref: item.EvidenceRef, Kind: firstString(item.Kind, "public_transport_ticket_inventory"), Strength: item.Strength, Source: runtimeRunRef(request), ObservedAt: item.ObservedAt}
		refs := []control.DisplaySafeRef{item.SourceRef, item.QueryRef, item.RouteRef, item.TravelDateRef, item.InventoryRef, item.EvidenceRef, item.DisplaySafeRef}
		out = append(out, control.Observation{
			Kind: firstString(item.Kind, "public_transport_ticket_inventory"), Source: firstRef([]control.DisplaySafeRef{item.SourceRef, runtimeRunRef(request)}),
			Subject: objectiveRef(request), Name: "public_transport_ticket_inventory", Value: string(item.InventoryRef), Unit: item.Confidence,
			Strength: item.Strength, ObservedAt: item.ObservedAt, EvidenceRefs: []control.EvidenceRef{evidence}, DisplaySafeRefs: normalizeRefs(refs),
		})
	}
	for _, row := range report.InventoryRows {
		out = append(out, control.Observation{
			Kind: "public_transport_ticket_inventory_row", Source: firstRef(report.SourceRefs), Subject: objectiveRef(request),
			Name: "public_transport_ticket_inventory_row", Value: string(row.RowRef), Unit: row.AvailabilityStatus,
			Strength: control.EvidenceAdequate, ObservedAt: report.ObservedAt, DisplaySafeRefs: normalizeRefs([]control.DisplaySafeRef{row.RowRef, row.TrainRef}),
		})
	}
	if len(out) == 0 {
		reason := firstString(report.FailureReason, "official_ticket_inventory_unavailable")
		if len(report.UnavailableReasons) > 0 {
			reason = report.UnavailableReasons[0]
		}
		out = append(out, control.Observation{
			Kind: "public_transport_ticket_unavailable", Source: runtimeRunRef(request), Subject: objectiveRef(request),
			Name: "public_transport_ticket_unavailable", Value: reason, Strength: control.EvidenceMissing, ObservedAt: report.ObservedAt,
			DisplaySafeRefs: []control.DisplaySafeRef{runtimeRunRef(request)}, DegradationReason: reason,
		})
	}
	return out
}

func evidenceRefs(request control.RuntimeAdapterExecutionRequest, report Report) []control.EvidenceRef {
	report = report.Normalize()
	out := make([]control.EvidenceRef, 0, len(report.QueryEvidenceRefs)+len(report.Evidence))
	for _, ref := range report.QueryEvidenceRefs {
		out = append(out, control.EvidenceRef{Ref: ref, Kind: "public_transport_ticket_query", Strength: control.EvidenceAdequate, Source: runtimeRunRef(request), ObservedAt: report.ObservedAt})
	}
	for _, item := range report.Evidence {
		out = append(out, control.EvidenceRef{Ref: item.EvidenceRef, Kind: firstString(item.Kind, "public_transport_ticket_inventory"), Strength: item.Strength, Source: runtimeRunRef(request), ObservedAt: item.ObservedAt})
	}
	return control.MergeEvidenceRefs(out)
}

func refsWithPrefixes(values []control.DisplaySafeRef, prefixes ...string) []control.DisplaySafeRef {
	out := []control.DisplaySafeRef{}
	for _, value := range values {
		for _, prefix := range prefixes {
			if len(value) >= len(prefix) && string(value[:len(prefix)]) == prefix {
				out = append(out, value)
				break
			}
		}
	}
	return normalizeRefs(out)
}

func runtimeRunRef(request control.RuntimeAdapterExecutionRequest) control.DisplaySafeRef {
	if request.Frame.ID != "" {
		return makeRef("adapter_run:public_transport_ticket", request.Frame.ID+":"+string(request.StrategyRef))
	}
	return DefaultRunRef
}

func objectiveRef(request control.RuntimeAdapterExecutionRequest) control.DisplaySafeRef {
	if ref, ok := control.NormalizeDisplaySafeRef(request.Frame.ID); ok {
		return ref
	}
	return "objective:public_transport_ticket"
}

func firstRef(values []control.DisplaySafeRef) control.DisplaySafeRef {
	for _, value := range values {
		if value = normalizeRef(value); value != "" {
			return value
		}
	}
	return ""
}
