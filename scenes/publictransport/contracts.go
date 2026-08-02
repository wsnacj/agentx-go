package publictransport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

const (
	DefaultAdapterRef      control.DisplaySafeRef = "adapter:public_transport_ticket_lookup"
	DefaultStrategyRef     control.DisplaySafeRef = "strategy:public_transport_ticket_lookup"
	DefaultRunRef          control.DisplaySafeRef = "adapter_run:public_transport_ticket_lookup"
	DefaultEvidenceRef     control.DisplaySafeRef = "evidence:public_transport_ticket_inventory"
	DefaultCapabilityRef   control.DisplaySafeRef = "capability:public_transport_ticket_lookup"
	DefaultSourcePolicyRef control.DisplaySafeRef = "source_policy:host_public_transport_ticket_lookup_required"
)

// Collector is the only runtime port owned by the Domain Kit. Implementations
// may call a provider, fixture, cache, or another host service.
type Collector interface {
	CollectPublicTransportTicketEvidence(context.Context, Request) (Report, error)
}

// Request carries only display-safe runtime identity and policy references.
// Provider endpoints, credentials, station catalogs, and HTTP policy remain host-owned.
type Request struct {
	RuntimeRequest control.RuntimeAdapterExecutionRequest `json:"runtime_request,omitempty"`
	QueryRefs      []control.DisplaySafeRef               `json:"query_refs,omitempty"`
	RouteRefs      []control.DisplaySafeRef               `json:"route_refs,omitempty"`
	TravelDateRefs []control.DisplaySafeRef               `json:"travel_date_refs,omitempty"`
	SourceRefs     []control.DisplaySafeRef               `json:"source_refs,omitempty"`
	PolicyRefs     []control.DisplaySafeRef               `json:"policy_refs,omitempty"`
	ObservedAt     string                                 `json:"observed_at,omitempty"`
	Boundaries     []control.Boundary                     `json:"boundaries,omitempty"`
}

// Evidence identifies one display-safe inventory observation.
type Evidence struct {
	EvidenceRef    control.DisplaySafeRef   `json:"evidence_ref,omitempty"`
	QueryRef       control.DisplaySafeRef   `json:"query_ref,omitempty"`
	RouteRef       control.DisplaySafeRef   `json:"route_ref,omitempty"`
	TravelDateRef  control.DisplaySafeRef   `json:"travel_date_ref,omitempty"`
	SourceRef      control.DisplaySafeRef   `json:"source_ref,omitempty"`
	InventoryRef   control.DisplaySafeRef   `json:"inventory_ref,omitempty"`
	Kind           string                   `json:"kind,omitempty"`
	Strength       control.EvidenceStrength `json:"strength,omitempty"`
	ObservedAt     string                   `json:"observed_at,omitempty"`
	Confidence     string                   `json:"confidence,omitempty"`
	DisplaySafeRef control.DisplaySafeRef   `json:"display_safe_ref,omitempty"`
}

// InventoryRow is a provider-neutral, display-safe row. FareLookupTokens are
// opaque host adapter tokens and must never contain credentials.
type InventoryRow struct {
	RowRef             control.DisplaySafeRef `json:"row_ref,omitempty"`
	TrainRef           control.DisplaySafeRef `json:"train_ref,omitempty"`
	TrainNo            string                 `json:"train_no,omitempty"`
	DepartureTime      string                 `json:"departure_time,omitempty"`
	ArrivalTime        string                 `json:"arrival_time,omitempty"`
	Duration           string                 `json:"duration,omitempty"`
	SeatSummary        string                 `json:"seat_summary,omitempty"`
	FareSummary        string                 `json:"fare_summary,omitempty"`
	FareLookupTokens   map[string]string      `json:"fare_lookup_tokens,omitempty"`
	AvailabilityStatus string                 `json:"availability_status,omitempty"`
}

// Report is the canonical provider-neutral readback contract.
type Report struct {
	Status                           control.VerificationStatus `json:"status,omitempty"`
	FailureClass                     control.FailureClass       `json:"failure_class,omitempty"`
	FailureReason                    string                     `json:"failure_reason,omitempty"`
	ObservedAt                       string                     `json:"observed_at,omitempty"`
	InventoryObserved                bool                       `json:"inventory_observed"`
	TicketResultClaimed              bool                       `json:"ticket_result_claimed"`
	BookingAttempted                 bool                       `json:"booking_attempted"`
	PurchaseAttempted                bool                       `json:"purchase_attempted"`
	SourceRefs                       []control.DisplaySafeRef   `json:"source_refs,omitempty"`
	QueryEvidenceRefs                []control.DisplaySafeRef   `json:"query_evidence_refs,omitempty"`
	Evidence                         []Evidence                 `json:"evidence,omitempty"`
	InventoryRows                    []InventoryRow             `json:"inventory_rows,omitempty"`
	FareLookupAttemptCount           int                        `json:"fare_lookup_attempt_count,omitempty"`
	FareLookupObservedCount          int                        `json:"fare_lookup_observed_count,omitempty"`
	FareLookupBlockedCount           int                        `json:"fare_lookup_blocked_count,omitempty"`
	StationCandidatePairCount        int                        `json:"station_candidate_pair_count,omitempty"`
	StationCandidatePairAttemptCount int                        `json:"station_candidate_pair_attempt_count,omitempty"`
	StationCandidatePairLimit        int                        `json:"station_candidate_pair_limit,omitempty"`
	StationCandidateInventoryCount   int                        `json:"station_candidate_inventory_count,omitempty"`
	StationCandidateEvidenceMisses   int                        `json:"station_candidate_evidence_misses,omitempty"`
	StationCandidateEmptyCount       int                        `json:"station_candidate_empty_count,omitempty"`
	StationCandidateFailureCount     int                        `json:"station_candidate_failure_count,omitempty"`
	SelectedFromStationCode          string                     `json:"selected_from_station_code,omitempty"`
	SelectedToStationCode            string                     `json:"selected_to_station_code,omitempty"`
	MissingInputs                    []control.MissingInput     `json:"missing_inputs,omitempty"`
	UnavailableReasons               []string                   `json:"unavailable_reasons,omitempty"`
	Boundaries                       []control.Boundary         `json:"boundaries,omitempty"`
	RawOutputLoaded                  bool                       `json:"raw_output_loaded"`
}

// Normalize returns a detached, fail-closed report.
func (report Report) Normalize() Report {
	out := report
	out.Status = control.NormalizeVerificationStatus(string(out.Status))
	if out.Status == control.VerificationNotEvaluated {
		out.Status = control.VerificationBlocked
	}
	out.FailureClass = control.NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = controlToken(out.FailureReason)
	out.ObservedAt = strings.TrimSpace(out.ObservedAt)
	out.SourceRefs = normalizeRefs(out.SourceRefs)
	out.QueryEvidenceRefs = normalizeRefs(out.QueryEvidenceRefs)
	out.Evidence = normalizeEvidence(out.Evidence, out.ObservedAt)
	out.InventoryRows = normalizeRows(out.InventoryRows)
	out.FareLookupAttemptCount = nonNegative(out.FareLookupAttemptCount)
	out.FareLookupObservedCount = nonNegative(out.FareLookupObservedCount)
	out.FareLookupBlockedCount = nonNegative(out.FareLookupBlockedCount)
	out.StationCandidatePairCount = nonNegative(out.StationCandidatePairCount)
	out.StationCandidatePairAttemptCount = nonNegative(out.StationCandidatePairAttemptCount)
	out.StationCandidatePairLimit = nonNegative(out.StationCandidatePairLimit)
	out.StationCandidateInventoryCount = nonNegative(out.StationCandidateInventoryCount)
	out.StationCandidateEvidenceMisses = nonNegative(out.StationCandidateEvidenceMisses)
	out.StationCandidateEmptyCount = nonNegative(out.StationCandidateEmptyCount)
	out.StationCandidateFailureCount = nonNegative(out.StationCandidateFailureCount)
	out.SelectedFromStationCode = queryToken(out.SelectedFromStationCode)
	out.SelectedToStationCode = queryToken(out.SelectedToStationCode)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.UnavailableReasons = controlTokens(out.UnavailableReasons)
	out.Boundaries = control.AppendBoundaries(nil, out.Boundaries...)
	out.InventoryObserved = out.InventoryObserved && len(out.Evidence) > 0
	out.TicketResultClaimed = out.TicketResultClaimed && out.InventoryObserved && len(out.Evidence) > 0
	if !out.InventoryObserved {
		out.InventoryRows = nil
	}
	if out.BookingAttempted || out.PurchaseAttempted {
		out.Status = control.VerificationReviewRequired
		out.FailureClass = control.FailurePolicyBlocked
		out.FailureReason = firstString(out.FailureReason, "ticket_purchase_not_allowed")
		out.Boundaries = control.AppendBoundaries(out.Boundaries, "booking_or_purchase_attempt_not_allowed")
	}
	if out.RawOutputLoaded {
		out.Status = control.VerificationReviewRequired
		if out.FailureClass == control.FailureNone {
			out.FailureClass = control.FailureEvidenceWeak
		}
		out.MissingInputs = control.AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = control.AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	if out.Status == control.VerificationSatisfied && len(out.Evidence) == 0 {
		out.Status = control.VerificationBlocked
		out.FailureClass = control.FailureEvidenceMissing
		out.FailureReason = "official_ticket_inventory_evidence_missing"
	}
	if out.Status != control.VerificationSatisfied {
		out.MissingInputs = control.AppendMissingInputs(out.MissingInputs, "host:official_ticket_inventory_evidence")
		if out.FailureClass == control.FailureNone {
			out.FailureClass = control.FailureTargetUnavailable
		}
	}
	return out
}

func normalizeEvidence(values []Evidence, fallbackObservedAt string) []Evidence {
	out := make([]Evidence, 0, len(values))
	for _, value := range values {
		item := value
		item.EvidenceRef = normalizeRef(item.EvidenceRef)
		item.QueryRef = normalizeRef(item.QueryRef)
		item.RouteRef = normalizeRef(item.RouteRef)
		item.TravelDateRef = normalizeRef(item.TravelDateRef)
		item.SourceRef = normalizeRef(item.SourceRef)
		item.InventoryRef = normalizeRef(item.InventoryRef)
		item.Kind = firstString(controlToken(item.Kind), "public_transport_ticket_inventory")
		item.Strength = control.NormalizeEvidenceStrength(string(item.Strength))
		if item.Strength == control.EvidenceMissing {
			item.Strength = control.EvidenceAdequate
		}
		item.ObservedAt = firstString(strings.TrimSpace(item.ObservedAt), fallbackObservedAt)
		item.Confidence = firstString(controlToken(item.Confidence), "confidence:medium")
		item.DisplaySafeRef = normalizeRef(item.DisplaySafeRef)
		if item.EvidenceRef == "" || item.InventoryRef == "" {
			continue
		}
		if item.DisplaySafeRef == "" {
			item.DisplaySafeRef = makeRef("result:public_transport_ticket_inventory", string(item.EvidenceRef))
		}
		out = append(out, item)
	}
	return out
}

func normalizeRows(values []InventoryRow) []InventoryRow {
	out := make([]InventoryRow, 0, len(values))
	seen := map[control.DisplaySafeRef]bool{}
	for _, value := range values {
		row := value
		row.TrainNo = displayText(row.TrainNo, 32)
		row.DepartureTime = displayText(row.DepartureTime, 16)
		row.ArrivalTime = displayText(row.ArrivalTime, 16)
		row.Duration = displayText(row.Duration, 24)
		row.SeatSummary = displayText(row.SeatSummary, 160)
		row.FareSummary = displayText(row.FareSummary, 160)
		row.FareLookupTokens = cloneStringMap(row.FareLookupTokens)
		row.AvailabilityStatus = controlToken(firstString(row.AvailabilityStatus, availabilityStatus(row.SeatSummary)))
		if row.AvailabilityStatus == "" {
			row.AvailabilityStatus = "unknown"
		}
		if row.TrainNo == "" || (row.DepartureTime == "" && row.ArrivalTime == "" && row.SeatSummary == "" && row.FareSummary == "") {
			continue
		}
		row.TrainRef = normalizeRef(row.TrainRef)
		if row.TrainRef == "" {
			row.TrainRef = makeRef("train", row.TrainNo)
		}
		row.RowRef = normalizeRef(row.RowRef)
		if row.RowRef == "" {
			row.RowRef = makeRef("ticket_inventory_row", row.TrainNo+":"+row.DepartureTime+":"+row.ArrivalTime+":"+row.SeatSummary+":"+row.FareSummary)
		}
		if seen[row.RowRef] {
			continue
		}
		seen[row.RowRef] = true
		out = append(out, row)
	}
	return out
}

func normalizeRefs(values []control.DisplaySafeRef) []control.DisplaySafeRef {
	out := []control.DisplaySafeRef{}
	for _, value := range values {
		value = normalizeRef(value)
		if value == "" || containsRef(out, value) {
			continue
		}
		out = append(out, value)
	}
	return out
}

func normalizeRef(value control.DisplaySafeRef) control.DisplaySafeRef {
	ref, ok := control.NormalizeDisplaySafeRef(string(value))
	if !ok {
		return ""
	}
	return ref
}

func normalizeMissingInputs(values []control.MissingInput) []control.MissingInput {
	out := []control.MissingInput{}
	for _, value := range values {
		value = control.MissingInput(controlToken(string(value)))
		if value != "" {
			out = control.AppendMissingInputs(out, value)
		}
	}
	return out
}

func controlTokens(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = controlToken(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func controlToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(" ", "_", "-", "_", ".", "_", "/", "_", ":", "_").Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}

func queryToken(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-', r == '.', r == ':':
			return r
		default:
			return -1
		}
	}, value)
	return value
}

func makeRef(prefix, raw string) control.DisplaySafeRef {
	prefix = strings.TrimSpace(prefix)
	raw = strings.TrimSpace(raw)
	if prefix == "" || raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return control.DisplaySafeRef(prefix + ":" + hex.EncodeToString(sum[:8]))
}

func displayText(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(strings.ToLower(value), "http://") || strings.Contains(strings.ToLower(value), "https://") {
		return ""
	}
	value = strings.Join(strings.Fields(value), " ")
	if maxLen > 0 && len([]rune(value)) > maxLen {
		value = string([]rune(value)[:maxLen])
	}
	return value
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = controlToken(key)
		value = displayText(value, 128)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func availabilityStatus(summary string) string {
	if strings.TrimSpace(summary) == "" {
		return "unknown"
	}
	for _, part := range strings.Split(summary, ",") {
		value := strings.TrimSpace(part)
		if idx := strings.Index(value, "="); idx >= 0 {
			value = strings.TrimSpace(value[idx+1:])
		}
		if seatAvailable(value) {
			return "available"
		}
	}
	return "sold_out"
}

func seatAvailable(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "none", "no", "false", "sold_out", "soldout", "not_available", "unavailable", "wu", "--", "*", "无", "无票", "售完", "已售完", "候补":
		return false
	default:
		return true
	}
}

func containsRef(values []control.DisplaySafeRef, want control.DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func nonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
