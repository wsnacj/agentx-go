// Package contextwindow provides provider-neutral context-window preparation.
//
// It composes transcript sanitation and deterministic compaction with a
// Host-supplied semantic summarizer. The package owns no model selection,
// network client, credential, persistence, retry policy, or global state.
//
// Current maturity: Experimental / private validation.
package contextwindow
