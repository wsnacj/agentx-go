// Package pipeline provides the experimental portable document parsing runtime.
//
// Runtime owns deterministic orchestration, extraction policies, diagnostics,
// caching and artifact formatting. Hosts explicitly provide document loading,
// section splitting and model execution; the package never discovers provider
// credentials, product configuration or concrete backends.
package pipeline
