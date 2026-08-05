package mcp

import (
	"context"
	"encoding/json"
	"time"
)

const (
	// ProtocolVersion20251125 is the current stable MCP protocol revision used
	// by the first AgentX MCP client implementation.
	ProtocolVersion20251125 = "2025-11-25"
)

// RequestID is an integer JSON-RPC request identity generated once per MCP
// session.
type RequestID int64

// Request is one JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Notification is one JSON-RPC 2.0 notification without an ID.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCError is an untrusted remote JSON-RPC error payload.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Response is one JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// Transport owns message framing and concrete I/O. Implementations must be
// safe for concurrent Request calls, honor ctx, bound incoming messages, and
// make Shutdown idempotent.
type Transport interface {
	Request(context.Context, Request) (Response, error)
	Notify(context.Context, Notification) error
	Shutdown(context.Context) error
}

// Implementation identifies an MCP client or server implementation.
type Implementation struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// ClientCapabilities is deliberately empty in P7-E2. Sampling, roots,
// elicitation and experimental Tasks are not advertised.
type ClientCapabilities struct{}

type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

type ToolCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ServerCapabilities struct {
	Tools *ToolCapability `json:"tools,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// Tool is the MCP Tool declaration. Annotation and execution metadata remain
// untrusted descriptive data and do not grant authorization.
type Tool struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description,omitempty"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Annotations  map[string]any `json:"annotations,omitempty"`
	Execution    map[string]any `json:"execution,omitempty"`
}

type ListToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type CallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ContentBlock preserves MCP content fields as a detached JSON object. The
// Type and Text helpers provide the common text-content path without losing
// unknown protocol fields.
type ContentBlock map[string]any

func (b ContentBlock) Type() string {
	value, _ := b["type"].(string)
	return value
}

func (b ContentBlock) Text() string {
	value, _ := b["text"].(string)
	return value
}

type CallToolResult struct {
	Meta              map[string]any `json:"_meta,omitempty"`
	Content           []ContentBlock `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError,omitempty"`
}

// Config defines the supported protocol revisions and safety bounds. The
// first supported version is offered during initialization.
type Config struct {
	Transport          Transport
	ClientInfo         Implementation
	SupportedVersions  []string
	RequestTimeout     time.Duration
	CancellationWindow time.Duration
	MaxTools           int
	MaxListPages       int
}

// State is a detached lifecycle readback. Server Instructions are returned
// for Host review and are never injected into prompts automatically.
type State struct {
	Initialized     bool
	ProtocolVersion string
	ServerInfo      Implementation
	Capabilities    ServerCapabilities
	Instructions    string
}
