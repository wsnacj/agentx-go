package productshell

import "strings"

// HostUIHandoffRuntimeUseSchemaV1 identifies runtime-use evidence for a concrete handoff consumer.
const HostUIHandoffRuntimeUseSchemaV1 = "agentx.productshell.host_ui_handoff.runtime_use.v1"

// HostUIHandoffRuntimeUseInput records one concrete host adapter consuming a
// handoff envelope. Its fields are independent of inventory or onboarding
// governance contracts.
type HostUIHandoffRuntimeUseInput struct {
	Consumer               string
	HostAdapter            string
	Target                 string
	Source                 string
	HandoffSchema          string
	ConformanceSchema      string
	ConformanceRefs        []string
	Envelope               *HostUIHandoffEnvelope
	ConsumedEntryCount     int
	DecodedHostDiagnostics bool
	RuntimeDeliverySource  bool
	Boundaries             []string
}

// HostUIHandoffRuntimeUseReport is display-safe evidence that a concrete consumer used a handoff envelope.
type HostUIHandoffRuntimeUseReport struct {
	Schema                   string                                  `json:"schema,omitempty"`
	Status                   string                                  `json:"status,omitempty"`
	Ready                    bool                                    `json:"ready"`
	Consumer                 string                                  `json:"consumer,omitempty"`
	HostAdapter              string                                  `json:"host_adapter,omitempty"`
	Target                   string                                  `json:"target,omitempty"`
	Source                   string                                  `json:"source,omitempty"`
	EnvelopePresent          bool                                    `json:"envelope_present"`
	EnvelopeSchema           string                                  `json:"envelope_schema,omitempty"`
	EnvelopeTarget           string                                  `json:"envelope_target,omitempty"`
	EnvelopeSource           string                                  `json:"envelope_source,omitempty"`
	EnvelopeEntryCount       int                                     `json:"envelope_entry_count,omitempty"`
	ConsumedEntryCount       int                                     `json:"consumed_entry_count,omitempty"`
	ConsumedAllEntries       bool                                    `json:"consumed_all_entries"`
	ConformancePassed        bool                                    `json:"conformance_passed"`
	ConsumesHandoffEnvelope  bool                                    `json:"consumes_handoff_envelope"`
	DecodedHostDiagnostics   bool                                    `json:"decoded_host_diagnostics"`
	HostAdapterOwnsDelivery  bool                                    `json:"host_adapter_owns_delivery"`
	EngineOwnsDelivery       bool                                    `json:"engine_owns_delivery"`
	ProductShellOwnsDelivery bool                                    `json:"product_shell_owns_delivery"`
	RuntimeDeliverySource    bool                                    `json:"runtime_delivery_source"`
	MissingInputs            []string                                `json:"missing_inputs,omitempty"`
	BlockedReasons           []string                                `json:"blocked_reasons,omitempty"`
	UnsafeDescriptorRefs     []string                                `json:"unsafe_descriptor_refs,omitempty"`
	Boundaries               []string                                `json:"boundaries,omitempty"`
	NextHostAction           string                                  `json:"next_host_action,omitempty"`
	ConsumerConformance      *HostUIHandoffConsumerConformanceReport `json:"consumer_conformance,omitempty"`
}

// BuildHostUIHandoffRuntimeUseReport checks runtime-use evidence without
// reading raw host diagnostics or consulting an inventory registry.
func BuildHostUIHandoffRuntimeUseReport(input HostUIHandoffRuntimeUseInput) HostUIHandoffRuntimeUseReport {
	normalized, unsafeRefs := normalizeHostUIHandoffRuntimeUseInput(input)
	report := HostUIHandoffRuntimeUseReport{
		Schema:                  HostUIHandoffRuntimeUseSchemaV1,
		Status:                  "blocked",
		Consumer:                normalized.Consumer,
		HostAdapter:             normalized.HostAdapter,
		Target:                  normalized.Target,
		Source:                  normalized.Source,
		ConsumedEntryCount:      input.ConsumedEntryCount,
		DecodedHostDiagnostics:  input.DecodedHostDiagnostics,
		HostAdapterOwnsDelivery: true,
		RuntimeDeliverySource:   input.RuntimeDeliverySource,
		UnsafeDescriptorRefs:    unsafeRefs,
		NextHostAction:          "consume_host_ui_handoff_envelope",
		Boundaries: []string{
			"productshell_host_ui_handoff_runtime_use",
			"consumer_reads_handoff_envelope_only",
			"host_adapter_owns_delivery",
			"no_host_diagnostics_json_decode",
			"display_safe_handoff_fields",
			"engine_does_not_deliver_host_ui_handoff",
			"productshell_does_not_deliver_host_ui_handoff",
			"not_runtime_delivery_source",
		},
	}
	report.Boundaries = appendUniqueProductShellStrings(report.Boundaries, normalized.Boundaries...)
	report.MissingInputs = requireHostUIHandoffRuntimeUseInput(report.MissingInputs, report.Consumer, "host_ui_handoff:consumer")
	report.MissingInputs = requireHostUIHandoffRuntimeUseInput(report.MissingInputs, report.HostAdapter, "host_ui_handoff:host_adapter")
	report.MissingInputs = requireHostUIHandoffRuntimeUseInput(report.MissingInputs, report.Target, "host_ui_handoff:target")
	report.MissingInputs = requireHostUIHandoffRuntimeUseInput(report.MissingInputs, report.Source, "host_ui_handoff:source")
	if len(report.UnsafeDescriptorRefs) > 0 {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_descriptor_unsafe_token")
	}
	if input.Envelope == nil {
		report.MissingInputs = appendUniqueProductShellString(report.MissingInputs, "host_ui_handoff:envelope")
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_envelope_missing")
		return report
	}
	envelope := input.Envelope
	report.EnvelopePresent = true
	report.EnvelopeSchema = strings.TrimSpace(envelope.Schema)
	report.EnvelopeTarget = strings.TrimSpace(envelope.Target)
	report.EnvelopeSource = strings.TrimSpace(envelope.Source)
	report.EnvelopeEntryCount = len(envelope.Entries)
	conformance := BuildHostUIHandoffConsumerConformanceReport(HostUIHandoffConsumerConformanceInput{
		Consumer:       report.Consumer,
		ExpectedTarget: report.Target,
		ExpectedSource: report.Source,
		Envelope:       envelope,
	})
	report.ConsumerConformance = &conformance
	report.ConformancePassed = conformance.Passed
	if !report.ConformancePassed {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_conformance_blocked")
		report.BlockedReasons = appendUniqueProductShellStrings(report.BlockedReasons, conformance.BlockedReasons...)
	}
	if input.ConsumedEntryCount <= 0 {
		report.MissingInputs = appendUniqueProductShellString(report.MissingInputs, "host_ui_handoff:consumer_entry_use")
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_consumer_entry_missing")
	}
	report.ConsumedAllEntries = input.ConsumedEntryCount > 0 && input.ConsumedEntryCount == report.EnvelopeEntryCount
	if !report.ConsumedAllEntries {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_entry_count_mismatch")
	}
	if input.DecodedHostDiagnostics {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_raw_host_diagnostics_decode")
	}
	if input.RuntimeDeliverySource {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_runtime_use_delivery_source")
	}
	report.Ready = report.EnvelopePresent &&
		report.ConformancePassed &&
		report.ConsumedAllEntries &&
		!report.DecodedHostDiagnostics &&
		!report.RuntimeDeliverySource &&
		report.HostAdapterOwnsDelivery &&
		!report.EngineOwnsDelivery &&
		!report.ProductShellOwnsDelivery &&
		len(report.MissingInputs) == 0 &&
		len(report.BlockedReasons) == 0 &&
		len(report.UnsafeDescriptorRefs) == 0
	report.ConsumesHandoffEnvelope = report.Ready
	if report.Ready {
		report.Status = "ready"
		report.NextHostAction = "host_adapter_continue_owning_handoff_delivery"
		report.Boundaries = appendUniqueProductShellString(report.Boundaries, "host_ui_handoff_runtime_use_ready")
	}
	return report
}

func normalizeHostUIHandoffRuntimeUseInput(input HostUIHandoffRuntimeUseInput) (HostUIHandoffRuntimeUseInput, []string) {
	var unsafeRefs []string
	out := HostUIHandoffRuntimeUseInput{
		Consumer:               hostUIHandoffRuntimeUseToken(input.Consumer, "consumer", &unsafeRefs),
		HostAdapter:            hostUIHandoffRuntimeUseToken(input.HostAdapter, "host_adapter", &unsafeRefs),
		Target:                 hostUIHandoffRuntimeUseToken(input.Target, "target", &unsafeRefs),
		Source:                 hostUIHandoffRuntimeUseToken(input.Source, "source", &unsafeRefs),
		HandoffSchema:          hostUIHandoffRuntimeUseToken(hostUIHandoffFirstNonEmptyString(input.HandoffSchema, HostUIHandoffSchemaV1), "handoff_schema", &unsafeRefs),
		ConformanceSchema:      hostUIHandoffRuntimeUseToken(hostUIHandoffFirstNonEmptyString(input.ConformanceSchema, HostUIHandoffConsumerConformanceSchemaV1), "conformance_schema", &unsafeRefs),
		ConformanceRefs:        hostUIHandoffRuntimeUseTokenList(input.ConformanceRefs, "conformance_refs", &unsafeRefs),
		Envelope:               input.Envelope,
		ConsumedEntryCount:     input.ConsumedEntryCount,
		DecodedHostDiagnostics: input.DecodedHostDiagnostics,
		RuntimeDeliverySource:  input.RuntimeDeliverySource,
		Boundaries:             hostUIHandoffRuntimeUseTokenList(input.Boundaries, "boundaries", &unsafeRefs),
	}
	return out, unsafeRefs
}

func hostUIHandoffRuntimeUseToken(value string, field string, unsafeRefs *[]string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	token := hostUIHandoffRenderToken(value)
	if token != value {
		*unsafeRefs = appendUniqueProductShellString(*unsafeRefs, field)
	}
	return token
}

func hostUIHandoffRuntimeUseTokenList(values []string, field string, unsafeRefs *[]string) []string {
	var out []string
	for _, value := range values {
		token := hostUIHandoffRuntimeUseToken(value, field, unsafeRefs)
		if token == "" {
			continue
		}
		out = appendUniqueProductShellString(out, token)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func requireHostUIHandoffRuntimeUseInput(missing []string, value string, ref string) []string {
	if strings.TrimSpace(value) != "" {
		return missing
	}
	return appendUniqueProductShellString(missing, ref)
}

func hostUIHandoffFirstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
