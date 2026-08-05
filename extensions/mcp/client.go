package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultRequestTimeout     = 30 * time.Second
	defaultCancellationWindow = time.Second
	defaultMaxTools           = 256
	defaultMaxListPages       = 16
)

// Client implements the stable MCP lifecycle and Tool subset over an injected
// transport. It does not own process, HTTP, credentials, or authorization.
type Client struct {
	config Config

	initMu sync.Mutex
	mu     sync.RWMutex
	state  State
	closed bool

	nextID atomic.Int64
	active sync.WaitGroup

	drained   chan struct{}
	drainOnce sync.Once
}

func New(config Config) (*Client, error) {
	applyConfigDefaults(&config)
	if config.Transport == nil || strings.TrimSpace(config.ClientInfo.Name) == "" ||
		strings.TrimSpace(config.ClientInfo.Version) == "" || len(config.SupportedVersions) == 0 ||
		config.RequestTimeout <= 0 || config.CancellationWindow <= 0 ||
		config.MaxTools <= 0 || config.MaxListPages <= 0 {
		return nil, &Error{Code: ErrorCodeInvalidConfig}
	}
	seen := map[string]bool{}
	versions := make([]string, 0, len(config.SupportedVersions))
	for _, version := range config.SupportedVersions {
		version = strings.TrimSpace(version)
		if version == "" || seen[version] {
			continue
		}
		seen[version] = true
		versions = append(versions, version)
	}
	if len(versions) == 0 {
		return nil, &Error{Code: ErrorCodeInvalidConfig}
	}
	config.SupportedVersions = versions
	config.ClientInfo.Name = strings.TrimSpace(config.ClientInfo.Name)
	config.ClientInfo.Title = strings.TrimSpace(config.ClientInfo.Title)
	config.ClientInfo.Version = strings.TrimSpace(config.ClientInfo.Version)
	config.ClientInfo.Description = strings.TrimSpace(config.ClientInfo.Description)
	return &Client{config: config, drained: make(chan struct{})}, nil
}

func applyConfigDefaults(config *Config) {
	if len(config.SupportedVersions) == 0 {
		config.SupportedVersions = []string{ProtocolVersion20251125}
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	if config.CancellationWindow == 0 {
		config.CancellationWindow = defaultCancellationWindow
	}
	if config.MaxTools == 0 {
		config.MaxTools = defaultMaxTools
	}
	if config.MaxListPages == 0 {
		config.MaxListPages = defaultMaxListPages
	}
}

// Initialize performs the mandatory version/capability handshake and sends
// notifications/initialized before making operation methods available.
func (c *Client) Initialize(ctx context.Context) (State, error) {
	if err := validateContext(ctx); err != nil {
		return State{}, err
	}
	if !c.begin() {
		return State{}, &Error{Code: ErrorCodeClosed}
	}
	defer c.active.Done()
	c.initMu.Lock()
	defer c.initMu.Unlock()
	c.mu.RLock()
	if c.state.Initialized {
		state := cloneState(c.state)
		c.mu.RUnlock()
		return state, nil
	}
	c.mu.RUnlock()
	params := InitializeParams{
		ProtocolVersion: c.config.SupportedVersions[0],
		Capabilities:    ClientCapabilities{},
		ClientInfo:      c.config.ClientInfo,
	}
	var result InitializeResult
	if err := c.request(ctx, "initialize", params, &result); err != nil {
		return State{}, err
	}
	if !contains(c.config.SupportedVersions, result.ProtocolVersion) ||
		strings.TrimSpace(result.ServerInfo.Name) == "" || strings.TrimSpace(result.ServerInfo.Version) == "" {
		return State{}, &Error{Code: ErrorCodeProtocol, Cause: errors.New("unsupported version or invalid server info")}
	}
	if err := c.notify(ctx, "notifications/initialized", nil); err != nil {
		return State{}, err
	}
	state := State{
		Initialized:     true,
		ProtocolVersion: strings.TrimSpace(result.ProtocolVersion),
		ServerInfo:      cloneImplementation(result.ServerInfo),
		Capabilities:    cloneCapabilities(result.Capabilities),
		Instructions:    strings.TrimSpace(result.Instructions),
	}
	c.mu.Lock()
	c.state = state
	c.mu.Unlock()
	return cloneState(state), nil
}

// DiscoverTools returns a stable Tool snapshot backed by this Client.
func (c *Client) DiscoverTools(ctx context.Context) (*ToolSet, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if !c.begin() {
		return nil, &Error{Code: ErrorCodeClosed}
	}
	defer c.active.Done()
	if err := c.requireTools(); err != nil {
		return nil, err
	}
	tools := make([]Tool, 0)
	cursor := ""
	seenCursors := map[string]bool{}
	seenTools := map[string]bool{}
	for page := 0; page < c.config.MaxListPages; page++ {
		var result ListToolsResult
		if err := c.request(ctx, "tools/list", ListToolsParams{Cursor: cursor}, &result); err != nil {
			return nil, err
		}
		for _, candidate := range result.Tools {
			normalized, err := normalizeTool(candidate)
			if err != nil {
				return nil, err
			}
			if seenTools[normalized.Name] || len(tools) >= c.config.MaxTools {
				return nil, &Error{Code: ErrorCodeInvalidTool, Cause: errors.New("duplicate tool or tool limit exceeded")}
			}
			seenTools[normalized.Name] = true
			tools = append(tools, normalized)
		}
		next := strings.TrimSpace(result.NextCursor)
		if next == "" {
			sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
			return newToolSet(c, tools), nil
		}
		if seenCursors[next] {
			return nil, &Error{Code: ErrorCodeProtocol, Cause: errors.New("repeated tools cursor")}
		}
		seenCursors[next] = true
		cursor = next
	}
	return nil, &Error{Code: ErrorCodeProtocol, Cause: errors.New("tools page limit exceeded")}
}

// CallTool invokes one MCP Tool without applying Host authorization. Product
// callers should normally expose a reviewed ToolSet through their Tool owner.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (CallToolResult, error) {
	if err := validateContext(ctx); err != nil {
		return CallToolResult{}, err
	}
	if !c.begin() {
		return CallToolResult{}, &Error{Code: ErrorCodeClosed}
	}
	defer c.active.Done()
	if err := c.requireTools(); err != nil {
		return CallToolResult{}, err
	}
	if !validToolName(name) {
		return CallToolResult{}, &Error{Code: ErrorCodeInvalidTool}
	}
	var result CallToolResult
	if err := c.request(ctx, "tools/call", CallToolParams{Name: name, Arguments: cloneAnyMap(arguments)}, &result); err != nil {
		return CallToolResult{}, err
	}
	result = cloneCallToolResult(result)
	return result, nil
}

// Snapshot returns detached negotiated state while the Client is open.
func (c *Client) Snapshot(ctx context.Context) (State, error) {
	if err := validateContext(ctx); err != nil {
		return State{}, err
	}
	if !c.begin() {
		return State{}, &Error{Code: ErrorCodeClosed}
	}
	defer c.active.Done()
	c.mu.RLock()
	defer c.mu.RUnlock()
	return cloneState(c.state), nil
}

// Shutdown rejects new calls, waits for in-flight operations and delegates
// graceful termination to the injected Transport. It is bounded and
// idempotent when the Transport follows its contract.
func (c *Client) Shutdown(ctx context.Context) error {
	if c == nil {
		return nil
	}
	if err := validateContext(ctx); err != nil {
		return err
	}
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	c.drainOnce.Do(func() {
		go func() {
			c.active.Wait()
			close(c.drained)
		}()
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.drained:
	}
	if err := c.config.Transport.Shutdown(ctx); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &Error{Code: ErrorCodeTransport, Cause: err}
	}
	return nil
}

func (c *Client) request(ctx context.Context, method string, params any, output any) error {
	requestCtx, cancel := withRequestTimeout(ctx, c.config.RequestTimeout)
	defer cancel()
	raw, err := marshalOptional(params)
	if err != nil {
		return &Error{Code: ErrorCodeProtocol, Cause: err}
	}
	id := RequestID(c.nextID.Add(1))
	response, err := c.config.Transport.Request(requestCtx, Request{JSONRPC: "2.0", ID: id, Method: method, Params: raw})
	if err != nil {
		if requestCtx.Err() != nil {
			c.sendCancellation(ctx, id, requestCtx.Err())
			return requestCtx.Err()
		}
		return &Error{Code: ErrorCodeTransport, Cause: err}
	}
	if response.JSONRPC != "2.0" || response.ID != id || (response.Error == nil && len(response.Result) == 0) {
		return &Error{Code: ErrorCodeProtocol, Cause: errors.New("invalid JSON-RPC response")}
	}
	if response.Error != nil {
		return &Error{Code: ErrorCodeRemote, RemoteCode: response.Error.Code, Cause: errors.New(response.Error.Message)}
	}
	if err := json.Unmarshal(response.Result, output); err != nil {
		return &Error{Code: ErrorCodeProtocol, Cause: err}
	}
	return nil
}

func (c *Client) notify(ctx context.Context, method string, params any) error {
	raw, err := marshalOptional(params)
	if err != nil {
		return &Error{Code: ErrorCodeProtocol, Cause: err}
	}
	if err := c.config.Transport.Notify(ctx, Notification{JSONRPC: "2.0", Method: method, Params: raw}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &Error{Code: ErrorCodeTransport, Cause: err}
	}
	return nil
}

func (c *Client) sendCancellation(parent context.Context, id RequestID, reason error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), c.config.CancellationWindow)
	defer cancel()
	_ = c.notify(ctx, "notifications/cancelled", map[string]any{"requestId": id, "reason": reason.Error()})
}

func (c *Client) requireTools() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.state.Initialized {
		return &Error{Code: ErrorCodeNotInitialized}
	}
	if c.state.Capabilities.Tools == nil {
		return &Error{Code: ErrorCodeProtocol, Cause: errors.New("server did not negotiate tools")}
	}
	return nil
}

func (c *Client) begin() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	c.active.Add(1)
	return true
}

func withRequestTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func marshalOptional(value any) (json.RawMessage, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("MCP context is required")
	}
	return ctx.Err()
}

func contains(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneState(state State) State {
	state.ServerInfo = cloneImplementation(state.ServerInfo)
	state.Capabilities = cloneCapabilities(state.Capabilities)
	return state
}

func cloneImplementation(value Implementation) Implementation { return value }

func cloneCapabilities(value ServerCapabilities) ServerCapabilities {
	if value.Tools != nil {
		tools := *value.Tools
		value.Tools = &tools
	}
	return value
}
