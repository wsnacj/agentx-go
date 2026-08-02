package publictransport

import "strings"

const (
	SeatEvidenceModeAvailable = "available"
	SeatEvidenceModeObserved  = "observed"
)

// InventoryEvaluationInput describes deterministic acceptance requirements.
type InventoryEvaluationInput struct {
	Report                Report
	MinimumRows           int
	RequiredTrainPrefixes []string
	RequiredSeatTokens    []string
	SeatEvidenceMode      string
	RequireFareEvidence   bool
}

// InventoryEvaluation is a provider-neutral acceptance result.
type InventoryEvaluation struct {
	Passed                 bool           `json:"passed"`
	InventoryObserved      bool           `json:"inventory_observed"`
	MinimumRowsSatisfied   bool           `json:"minimum_rows_satisfied"`
	TrainEvidenceSatisfied bool           `json:"train_evidence_satisfied"`
	SeatEvidenceSatisfied  bool           `json:"seat_evidence_satisfied"`
	FareEvidenceSatisfied  bool           `json:"fare_evidence_satisfied"`
	NoBookingOrPurchase    bool           `json:"no_booking_or_purchase"`
	MatchingRows           []InventoryRow `json:"matching_rows,omitempty"`
	ObservedRows           []InventoryRow `json:"observed_rows,omitempty"`
	FailureReasons         []string       `json:"failure_reasons,omitempty"`
}

// EvaluateInventory applies the Domain Kit's deterministic evidence gate.
func EvaluateInventory(input InventoryEvaluationInput) InventoryEvaluation {
	report := input.Report.Normalize()
	minimumRows := nonNegative(input.MinimumRows)
	prefixes := normalizeTrainPrefixes(input.RequiredTrainPrefixes)
	seats := normalizeSeatTokens(input.RequiredSeatTokens)
	matching := FilterInventoryRowsByEvidence(report.InventoryRows, prefixes, seats)
	observed := FilterInventoryRowsByObservedEvidence(report.InventoryRows, prefixes, seats)
	out := InventoryEvaluation{
		InventoryObserved:      report.InventoryObserved,
		MinimumRowsSatisfied:   report.InventoryObserved && len(report.InventoryRows) >= minimumRows,
		TrainEvidenceSatisfied: len(prefixes) == 0 || RowsHaveTrainPrefix(report.InventoryRows, prefixes),
		FareEvidenceSatisfied:  !input.RequireFareEvidence || fareEvidenceComplete(report.InventoryRows),
		NoBookingOrPurchase:    !report.BookingAttempted && !report.PurchaseAttempted,
		MatchingRows:           matching,
		ObservedRows:           observed,
	}
	if NormalizeSeatEvidenceMode(input.SeatEvidenceMode) == SeatEvidenceModeObserved {
		out.SeatEvidenceSatisfied = len(seats) == 0 || RowsHaveSeatTokenObservation(report.InventoryRows, seats)
	} else {
		out.SeatEvidenceSatisfied = len(seats) == 0 || RowsHaveSeatToken(report.InventoryRows, seats)
	}
	if !out.InventoryObserved {
		out.FailureReasons = append(out.FailureReasons, "inventory_not_observed")
	}
	if !out.MinimumRowsSatisfied {
		out.FailureReasons = append(out.FailureReasons, "minimum_rows_not_satisfied")
	}
	if !out.TrainEvidenceSatisfied {
		out.FailureReasons = append(out.FailureReasons, "train_evidence_missing")
	}
	if !out.SeatEvidenceSatisfied {
		out.FailureReasons = append(out.FailureReasons, "seat_evidence_missing")
	}
	if !out.FareEvidenceSatisfied {
		out.FailureReasons = append(out.FailureReasons, "fare_evidence_missing")
	}
	if !out.NoBookingOrPurchase {
		out.FailureReasons = append(out.FailureReasons, "booking_or_purchase_not_allowed")
	}
	out.Passed = len(out.FailureReasons) == 0
	return out
}

// NormalizeSeatEvidenceMode returns the fail-closed evidence mode.
func NormalizeSeatEvidenceMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), SeatEvidenceModeObserved) {
		return SeatEvidenceModeObserved
	}
	return SeatEvidenceModeAvailable
}

// FilterInventoryRowsByEvidence returns rows satisfying all requested available evidence.
func FilterInventoryRowsByEvidence(rows []InventoryRow, trainPrefixes, seatTokens []string) []InventoryRow {
	rows = normalizeRows(rows)
	prefixes := normalizeTrainPrefixes(trainPrefixes)
	tokens := normalizeSeatTokens(seatTokens)
	if len(prefixes) == 0 && len(tokens) == 0 {
		return rows
	}
	out := make([]InventoryRow, 0, len(rows))
	for _, row := range rows {
		if InventoryRowMatchesEvidence(row, prefixes, tokens) {
			out = append(out, row)
		}
	}
	return out
}

// FilterInventoryRowsByObservedEvidence accepts sold-out but observed seat fields.
func FilterInventoryRowsByObservedEvidence(rows []InventoryRow, trainPrefixes, seatTokens []string) []InventoryRow {
	rows = normalizeRows(rows)
	prefixes := normalizeTrainPrefixes(trainPrefixes)
	tokens := normalizeSeatTokens(seatTokens)
	if len(prefixes) == 0 && len(tokens) == 0 {
		return rows
	}
	out := make([]InventoryRow, 0, len(rows))
	for _, row := range rows {
		if InventoryRowMatchesObservedEvidence(row, prefixes, tokens) {
			out = append(out, row)
		}
	}
	return out
}

func RowsHaveMatchingEvidence(rows []InventoryRow, trainPrefixes, seatTokens []string) bool {
	return len(FilterInventoryRowsByEvidence(rows, trainPrefixes, seatTokens)) > 0
}

func RowsHaveMatchingObservedEvidence(rows []InventoryRow, trainPrefixes, seatTokens []string) bool {
	return len(FilterInventoryRowsByObservedEvidence(rows, trainPrefixes, seatTokens)) > 0
}

func RowsHaveTrainPrefix(rows []InventoryRow, trainPrefixes []string) bool {
	prefixes := normalizeTrainPrefixes(trainPrefixes)
	if len(prefixes) == 0 {
		return false
	}
	for _, row := range normalizeRows(rows) {
		if rowMatchesTrainPrefix(row, prefixes) {
			return true
		}
	}
	return false
}

func RowsHaveSeatToken(rows []InventoryRow, seatTokens []string) bool {
	tokens := normalizeSeatTokens(seatTokens)
	if len(tokens) == 0 {
		return false
	}
	for _, row := range normalizeRows(rows) {
		if rowMatchesSeatToken(row, tokens) {
			return true
		}
	}
	return false
}

func RowsHaveSeatTokenObservation(rows []InventoryRow, seatTokens []string) bool {
	tokens := normalizeSeatTokens(seatTokens)
	if len(tokens) == 0 {
		return false
	}
	for _, row := range normalizeRows(rows) {
		if rowHasSeatTokenObservation(row, tokens) {
			return true
		}
	}
	return false
}

func InventoryRowMatchesEvidence(row InventoryRow, trainPrefixes, seatTokens []string) bool {
	prefixes := normalizeTrainPrefixes(trainPrefixes)
	tokens := normalizeSeatTokens(seatTokens)
	if len(prefixes) > 0 && !rowMatchesTrainPrefix(row, prefixes) {
		return false
	}
	if len(tokens) > 0 && !rowMatchesSeatToken(row, tokens) {
		return false
	}
	if len(prefixes) > 0 && len(tokens) == 0 && !rowHasAvailableSeat(row) {
		return false
	}
	return len(prefixes) > 0 || len(tokens) > 0
}

func InventoryRowMatchesObservedEvidence(row InventoryRow, trainPrefixes, seatTokens []string) bool {
	prefixes := normalizeTrainPrefixes(trainPrefixes)
	tokens := normalizeSeatTokens(seatTokens)
	if len(prefixes) > 0 && !rowMatchesTrainPrefix(row, prefixes) {
		return false
	}
	if len(tokens) > 0 && !rowHasSeatTokenObservation(row, tokens) {
		return false
	}
	return len(prefixes) > 0 || len(tokens) > 0
}

func rowMatchesTrainPrefix(row InventoryRow, prefixes []string) bool {
	trainNo := strings.ToUpper(strings.TrimSpace(row.TrainNo))
	for _, prefix := range normalizeTrainPrefixes(prefixes) {
		if trainNo != "" && strings.HasPrefix(trainNo, prefix) {
			return true
		}
	}
	return false
}

func rowMatchesSeatToken(row InventoryRow, tokens []string) bool {
	available := observedAvailableSeatTokens(row.SeatSummary)
	for _, token := range normalizeSeatTokens(tokens) {
		if available[token] {
			return true
		}
	}
	return false
}

func rowHasSeatTokenObservation(row InventoryRow, tokens []string) bool {
	observed := observedSeatTokens(row.SeatSummary)
	for _, token := range normalizeSeatTokens(tokens) {
		if observed[token] {
			return true
		}
	}
	return false
}

func rowHasAvailableSeat(row InventoryRow) bool {
	switch controlToken(row.AvailabilityStatus) {
	case "available":
		return true
	case "sold_out":
		return false
	default:
		return len(observedAvailableSeatTokens(row.SeatSummary)) > 0
	}
}

func observedSeatTokens(summary string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(summary, ",") {
		key, _, ok := strings.Cut(part, "=")
		if key = controlToken(key); ok && key != "" {
			out[key] = true
		}
	}
	return out
}

func observedAvailableSeatTokens(summary string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(summary, ",") {
		key, value, ok := strings.Cut(part, "=")
		key = controlToken(key)
		if ok && key != "" && seatAvailable(value) {
			out[key] = true
		}
	}
	return out
}

func normalizeTrainPrefixes(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.Trim(strings.ToUpper(strings.TrimSpace(value)), " .:-_/")
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func normalizeSeatTokens(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = controlToken(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func fareEvidenceComplete(rows []InventoryRow) bool {
	rows = normalizeRows(rows)
	if len(rows) == 0 {
		return true
	}
	for _, row := range rows {
		if strings.TrimSpace(row.FareSummary) == "" {
			return false
		}
	}
	return true
}
