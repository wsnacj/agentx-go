package productshell

import "strings"

// HostUIHandoffConsumerConformanceSchemaV1 identifies the host UI handoff consumer conformance report schema.
const HostUIHandoffConsumerConformanceSchemaV1 = "agentx.productshell.host_ui_handoff.consumer_conformance.v1"

// HostUIHandoffConsumerConformanceInput describes the expected consumer envelope contract.
type HostUIHandoffConsumerConformanceInput struct {
	Consumer       string
	ExpectedTarget string
	ExpectedSource string
	Envelope       *HostUIHandoffEnvelope
}

// HostUIHandoffConsumerConformanceReport records display-safe evidence for a host UI handoff consumer.
type HostUIHandoffConsumerConformanceReport struct {
	Schema                            string   `json:"schema,omitempty"`
	Consumer                          string   `json:"consumer,omitempty"`
	Status                            string   `json:"status,omitempty"`
	Passed                            bool     `json:"passed"`
	ExpectedTarget                    string   `json:"expected_target,omitempty"`
	ExpectedSource                    string   `json:"expected_source,omitempty"`
	EnvelopeSchema                    string   `json:"envelope_schema,omitempty"`
	EnvelopeSurface                   string   `json:"envelope_surface,omitempty"`
	EnvelopeTarget                    string   `json:"envelope_target,omitempty"`
	EnvelopeSource                    string   `json:"envelope_source,omitempty"`
	EntryCount                        int      `json:"entry_count,omitempty"`
	EnvelopePresent                   bool     `json:"envelope_present"`
	SchemaMatched                     bool     `json:"schema_matched"`
	SurfaceMatched                    bool     `json:"surface_matched"`
	TargetMatched                     bool     `json:"target_matched"`
	SourceMatched                     bool     `json:"source_matched"`
	EntryCountMatched                 bool     `json:"entry_count_matched"`
	LatestEntryMatched                bool     `json:"latest_entry_matched"`
	EntriesDisplaySafe                bool     `json:"entries_display_safe"`
	DeliveryBoundaryPresent           bool     `json:"delivery_boundary_present"`
	NoHostDiagnosticsDecodeBoundary   bool     `json:"no_host_diagnostics_decode_boundary"`
	DisplaySafeHandoffBoundaryPresent bool     `json:"display_safe_handoff_boundary_present"`
	MissingInputs                     []string `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string `json:"blocked_reasons,omitempty"`
	Boundaries                        []string `json:"boundaries,omitempty"`
}

// BuildHostUIHandoffConsumerConformanceReport checks a handoff envelope
// without reading raw host diagnostics.
func BuildHostUIHandoffConsumerConformanceReport(input HostUIHandoffConsumerConformanceInput) HostUIHandoffConsumerConformanceReport {
	report := HostUIHandoffConsumerConformanceReport{
		Schema:         HostUIHandoffConsumerConformanceSchemaV1,
		Consumer:       hostUIHandoffRenderToken(input.Consumer),
		ExpectedTarget: hostUIHandoffRenderToken(input.ExpectedTarget),
		ExpectedSource: hostUIHandoffRenderToken(input.ExpectedSource),
		Boundaries: []string{
			"productshell_host_ui_handoff_consumer_conformance",
			"conformance_reads_handoff_envelope_only",
			"host_surface_owns_delivery",
			"no_host_diagnostics_json_decode",
		},
	}
	if report.Consumer == "" {
		report.MissingInputs = appendUniqueProductShellString(report.MissingInputs, "consumer")
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "consumer_missing")
	}
	if input.Envelope == nil {
		report.MissingInputs = appendUniqueProductShellString(report.MissingInputs, "host_ui_handoff:envelope")
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_envelope_missing")
		report.Status = "blocked"
		return report
	}
	envelope := *input.Envelope
	report.EnvelopePresent = true
	report.EnvelopeSchema = strings.TrimSpace(envelope.Schema)
	report.EnvelopeSurface = strings.TrimSpace(envelope.Surface)
	report.EnvelopeTarget = strings.TrimSpace(envelope.Target)
	report.EnvelopeSource = strings.TrimSpace(envelope.Source)
	report.EntryCount = len(envelope.Entries)
	report.SchemaMatched = envelope.Schema == HostUIHandoffSchemaV1
	report.SurfaceMatched = envelope.Surface == HostUIHandoffSurface
	report.TargetMatched = report.ExpectedTarget == "" || envelope.Target == report.ExpectedTarget
	report.SourceMatched = report.ExpectedSource == "" || envelope.Source == report.ExpectedSource
	report.EntryCountMatched = envelope.EntryCount == len(envelope.Entries) && len(envelope.Entries) > 0
	report.LatestEntryMatched = hostUIHandoffLatestEntryMatches(envelope)
	report.EntriesDisplaySafe = hostUIHandoffEntriesDisplaySafe(envelope)
	report.DeliveryBoundaryPresent = productShellStringSliceContains(envelope.Boundaries, "host_adapter_owns_delivery")
	report.NoHostDiagnosticsDecodeBoundary = productShellStringSliceContains(envelope.Boundaries, "no_host_diagnostics_json_decode")
	report.DisplaySafeHandoffBoundaryPresent = productShellStringSliceContains(envelope.Boundaries, "display_safe_handoff_fields")
	if !report.SchemaMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_schema_mismatch")
	}
	if !report.SurfaceMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_surface_mismatch")
	}
	if !report.TargetMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_target_mismatch")
	}
	if !report.SourceMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_source_mismatch")
	}
	if !report.EntryCountMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_entries_missing_or_count_mismatch")
	}
	if !report.LatestEntryMatched {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_latest_entry_mismatch")
	}
	if !report.EntriesDisplaySafe {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_entry_not_display_safe")
	}
	if !report.DeliveryBoundaryPresent {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_delivery_boundary_missing")
	}
	if !report.NoHostDiagnosticsDecodeBoundary {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_no_host_diagnostics_decode_boundary_missing")
	}
	if !report.DisplaySafeHandoffBoundaryPresent {
		report.BlockedReasons = appendUniqueProductShellString(report.BlockedReasons, "host_ui_handoff_display_safe_boundary_missing")
	}
	report.Passed = report.Consumer != "" &&
		report.EnvelopePresent &&
		report.SchemaMatched &&
		report.SurfaceMatched &&
		report.TargetMatched &&
		report.SourceMatched &&
		report.EntryCountMatched &&
		report.LatestEntryMatched &&
		report.EntriesDisplaySafe &&
		report.DeliveryBoundaryPresent &&
		report.NoHostDiagnosticsDecodeBoundary &&
		report.DisplaySafeHandoffBoundaryPresent
	if report.Passed {
		report.Status = "passed"
	} else {
		report.Status = "blocked"
	}
	return report
}

func hostUIHandoffLatestEntryMatches(envelope HostUIHandoffEnvelope) bool {
	if len(envelope.Entries) == 0 || envelope.LatestEntry == nil {
		return false
	}
	latest := envelope.Entries[len(envelope.Entries)-1]
	return envelope.LatestEntry.Target == latest.Target &&
		envelope.LatestEntry.Source == latest.Source &&
		envelope.LatestEntry.Kind == latest.Kind &&
		envelope.LatestEntry.Key == latest.Key &&
		envelope.LatestEntry.Status == latest.Status &&
		envelope.LatestEntry.DisplayLine == latest.DisplayLine
}

func hostUIHandoffEntriesDisplaySafe(envelope HostUIHandoffEnvelope) bool {
	if len(envelope.Entries) == 0 {
		return false
	}
	for _, entry := range envelope.Entries {
		if entry.Target != envelope.Target || !hostUIHandoffConformanceTokenSafe(entry.Target) {
			return false
		}
		if !hostUIHandoffConformanceTokenSafe(entry.Source) ||
			!hostUIHandoffConformanceTokenSafe(entry.Kind) ||
			!hostUIHandoffConformanceTokenSafe(entry.Key) ||
			!hostUIHandoffConformanceTokenSafe(entry.Status) ||
			!hostUIHandoffConformanceTokenSafe(entry.NextHostAction) ||
			!hostUIHandoffConformanceTokenListSafe(entry.MissingInputs) ||
			!hostUIHandoffConformanceTokenListSafe(entry.BlockedReasons) ||
			!hostUIHandoffConformanceTokenListSafe(entry.Boundaries) ||
			!hostUIHandoffConformanceDisplayLineSafe(entry.DisplayLine) {
			return false
		}
		if entry.Kind == HostUIHandoffKindHostDiagnosticOperatorLine &&
			!productShellStringSliceContains(entry.Boundaries, "display_safe_operator_line_only") {
			return false
		}
	}
	return true
}

func hostUIHandoffConformanceTokenListSafe(values []string) bool {
	for _, value := range values {
		if !hostUIHandoffConformanceTokenSafe(value) {
			return false
		}
	}
	return true
}

func hostUIHandoffConformanceTokenSafe(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || value == "redacted" {
		return true
	}
	return hostUIHandoffRenderToken(value) == value
}

func hostUIHandoffConformanceDisplayLineSafe(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || value == "redacted" {
		return true
	}
	return hostUIHandoffRenderDisplayLine(value) == value
}

func productShellStringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
