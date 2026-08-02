// Package tools adapts the provider-neutral Browser Runtime to AgentX tool
// definitions and handlers. Callers explicitly inject browser backends and
// optional host ports; the package does not select a provider, discover
// credentials, or start a concrete browser service by itself.
package tools
