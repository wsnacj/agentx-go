// Package domainkit provides a model-free execution boundary for compiled
// AgentX domain modules.
package domainkit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/wsnacj/agentx-go/extensions/domainmodule"
)

// ErrorCode is a stable machine-readable failure classification.
type ErrorCode string

const (
	ErrorCodeInvalidConfig  ErrorCode = "invalid_config"
	ErrorCodeInvalidRequest ErrorCode = "invalid_request"
	ErrorCodeModuleNotFound ErrorCode = "module_not_found"
	ErrorCodeToolNotFound   ErrorCode = "tool_not_found"
	ErrorCodeHandlerFailed  ErrorCode = "handler_failed"
	ErrorCodeEncodingFailed ErrorCode = "encoding_failed"
)

var (
	ErrInvalidConfig  = &Error{Code: ErrorCodeInvalidConfig}
	ErrInvalidRequest = &Error{Code: ErrorCodeInvalidRequest}
	ErrModuleNotFound = &Error{Code: ErrorCodeModuleNotFound}
	ErrToolNotFound   = &Error{Code: ErrorCodeToolNotFound}
	ErrHandlerFailed  = &Error{Code: ErrorCodeHandlerFailed}
	ErrEncodingFailed = &Error{Code: ErrorCodeEncodingFailed}
)

// Error is the typed error returned by Runtime. Error keeps the wrapped
// cause's text for compatibility; DisplaySafeMessage is the value suitable for
// logs or user-visible status.
type Error struct {
	Code               ErrorCode
	Retryable          bool
	DisplaySafeMessage string
	Cause              error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.DisplaySafeMessage != "" {
		return e.DisplaySafeMessage
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && e.Code != "" && e.Code == other.Code
}

// Handler executes one explicitly selected module tool. The host owns all
// authorization, credentials, providers and side effects behind the handler.
type Handler func(context.Context, map[string]any) (any, error)

// Module binds a portable manifest to host-provided handlers. Every manifest
// tool must have exactly one non-nil handler.
type Module struct {
	Manifest domainmodule.Manifest
	Handlers map[string]Handler
}

// Config is the complete immutable construction input for a Runtime.
type Config struct {
	Modules []Module
}

// RunRequest selects one module-owned tool. CaseID is an opaque display-safe
// conformance or host correlation identity and does not affect dispatch.
type RunRequest struct {
	ModuleID  string
	CaseID    string
	Tool      string
	Arguments map[string]any
}

// RunResult is the stable model-free execution projection. OutputDigest is the
// lowercase SHA-256 of Output and allows callers to compare fixture runs.
type RunResult struct {
	ModuleID     string `json:"module_id"`
	CaseID       string `json:"case_id,omitempty"`
	Tool         string `json:"tool"`
	Output       string `json:"output"`
	OutputDigest string `json:"output_digest"`
}

type moduleEntry struct {
	manifest domainmodule.Manifest
	handlers map[string]Handler
}

// Runtime is an immutable, concurrency-safe dispatch table. Handler
// concurrency remains a host responsibility.
type Runtime struct {
	modules map[string]moduleEntry
}

// New validates all manifests and handlers before publishing the runtime.
func New(cfg Config) (*Runtime, error) {
	modules := make(map[string]moduleEntry, len(cfg.Modules))
	for _, candidate := range cfg.Modules {
		manifest, err := domainmodule.NormalizeManifest(candidate.Manifest)
		if err != nil {
			return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", err)
		}
		if _, exists := modules[manifest.ID]; exists {
			return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", fmt.Errorf("duplicate domain kit module id %q", manifest.ID))
		}
		declared := make(map[string]bool, len(manifest.Tools))
		for _, name := range manifest.Tools {
			declared[name] = true
		}
		handlers := make(map[string]Handler, len(candidate.Handlers))
		for rawName, handler := range candidate.Handlers {
			name := strings.TrimSpace(rawName)
			if name == "" || handler == nil {
				return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", fmt.Errorf("domain kit module %q has an empty tool handler", manifest.ID))
			}
			if !declared[name] {
				return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", fmt.Errorf("domain kit module %q handler %q is outside its manifest tool surface", manifest.ID, name))
			}
			if _, exists := handlers[name]; exists {
				return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", fmt.Errorf("domain kit module %q has duplicate tool handler %q", manifest.ID, name))
			}
			handlers[name] = handler
		}
		missing := make([]string, 0)
		for _, name := range manifest.Tools {
			if handlers[name] == nil {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", fmt.Errorf("domain kit module %q is missing tool handlers: %s", manifest.ID, strings.Join(missing, ", ")))
		}
		modules[manifest.ID] = moduleEntry{manifest: manifest, handlers: handlers}
	}
	if len(modules) == 0 {
		return nil, typedError(ErrorCodeInvalidConfig, "domain kit configuration is invalid", errors.New("domain kit requires at least one module"))
	}
	return &Runtime{modules: modules}, nil
}

// Run invokes exactly one explicit handler and never calls a model or selects
// a provider. The same request and deterministic handler produce the same
// output digest.
func (r *Runtime) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	if err := ctx.Err(); err != nil {
		return RunResult{}, err
	}
	moduleID := domainmodule.NormalizeID(req.ModuleID)
	tool := strings.TrimSpace(req.Tool)
	caseID := strings.TrimSpace(req.CaseID)
	if r == nil || moduleID == "" || tool == "" {
		return RunResult{}, typedError(ErrorCodeInvalidRequest, "domain kit request is invalid", errors.New("domain kit module id and tool are required"))
	}
	module, ok := r.modules[moduleID]
	if !ok {
		return RunResult{}, typedError(ErrorCodeModuleNotFound, "domain kit module is unavailable", fmt.Errorf("domain kit module %q is not registered", moduleID))
	}
	handler, ok := module.handlers[tool]
	if !ok {
		return RunResult{}, typedError(ErrorCodeToolNotFound, "domain kit tool is unavailable", fmt.Errorf("domain kit module %q tool %q is not registered", moduleID, tool))
	}
	payload, err := handler(ctx, cloneArguments(req.Arguments))
	if err != nil {
		return RunResult{}, typedError(ErrorCodeHandlerFailed, "domain kit handler failed", err)
	}
	output, err := encodeOutput(payload)
	if err != nil {
		return RunResult{}, typedError(ErrorCodeEncodingFailed, "domain kit output encoding failed", err)
	}
	digest := sha256.Sum256([]byte(output))
	return RunResult{
		ModuleID:     module.manifest.ID,
		CaseID:       caseID,
		Tool:         tool,
		Output:       output,
		OutputDigest: hex.EncodeToString(digest[:]),
	}, nil
}

func typedError(code ErrorCode, display string, cause error) error {
	return &Error{Code: code, DisplaySafeMessage: display, Cause: cause}
}

func encodeOutput(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(append([]byte(nil), typed...)), nil
	default:
		blob, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(blob), nil
	}
}

func cloneArguments(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneArguments(typed)
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = cloneValue(typed[i])
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	case map[string]string:
		out := make(map[string]string, len(typed))
		for key, item := range typed {
			out[key] = item
		}
		return out
	default:
		return value
	}
}
