// Package memory provides portable long-term memory lifecycle coordination.
//
// It owns validation, explicit scope and provenance propagation, bounded recall,
// idempotent write/archive requests, revision compare-and-swap expectations and
// typed errors. Concrete storage, ranking, visibility and retention policies
// remain Host-owned behind Backend.
package memory
