package controlcontract

func normalizeAttemptRefs(in []AttemptRef) []AttemptRef {
	out := make([]AttemptRef, 0, len(in))
	seen := map[AttemptRef]struct{}{}
	for _, value := range in {
		normalized, ok := NormalizeAttemptRef(string(value))
		if !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneAttemptRefs(in []AttemptRef) []AttemptRef {
	if len(in) == 0 {
		return nil
	}
	out := make([]AttemptRef, len(in))
	copy(out, in)
	return out
}

func observationSliceUnsafePayload(values []Observation) bool {
	for _, value := range values {
		if value.RawOutputLoaded ||
			displaySafeRefRejected(value.Source) ||
			displaySafeRefRejected(value.Subject) ||
			displaySafeRefSliceRejected(value.DisplaySafeRefs) ||
			evidenceRefRejected(value.EvidenceRefs) ||
			ContainsUnsafeRawOutput(value.Name, value.Value, value.Unit, value.ObservedAt, value.DegradationReason) {
			return true
		}
	}
	return false
}

func runtimeAdapterObservationEvidenceRefs(values []Observation) []EvidenceRef {
	out := []EvidenceRef{}
	for _, observation := range normalizeObservations(values) {
		out = MergeEvidenceRefs(out, observation.EvidenceRefs)
	}
	return out
}
