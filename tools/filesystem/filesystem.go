// Package filesystem provides portable filesystem tool coordination.
package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

const (
	// ReadName is the catalog name of the read tool.
	ReadName = "read"
	// WriteName is the catalog name of the write tool.
	WriteName = "write"
	// EditName is the catalog name of the edit tool.
	EditName = "edit"
	// ApplyPatchName is the catalog name of the apply-patch tool.
	ApplyPatchName = "apply_patch"
)

const (
	defaultMaxReadChars         = 120_000
	defaultMaxWriteChars        = 200_000
	defaultMaxPatchChars        = 280_000
	defaultReadFileMaxBytes     = int64(64 << 20)
	defaultMutationFileMaxBytes = int64(4 << 20)
)

// Options configures portable filesystem coordination. Workspace owns every
// side effect and all Host policy; this package owns model-facing semantics.
type Options struct {
	Workspace            Workspace
	MaxReadChars         int
	MaxWriteChars        int
	MaxPatchChars        int
	MaxReadFileBytes     int64
	MaxMutationFileBytes int64
}

// Workspace is the narrow Host port for protected, root-bound filesystem IO.
// Implementations own workspace selection, authorization, symlink policy and
// atomic/durable mutation. They must not reinterpret model-facing arguments.
type Workspace interface {
	Read(context.Context, ReadRequest) (ReadResult, error)
	Write(context.Context, WriteRequest) (WriteResult, error)
	Edit(context.Context, EditRequest) (EditResult, error)
	ApplyPatch(context.Context, ApplyPatchRequest) (PatchSummary, error)
}

// ReadRequest is a normalized text selection request.
type ReadRequest struct {
	Path         string
	StartLine    int
	MaxLines     int
	MaxChars     int
	MaxFileBytes int64
}

// ReadResult is the Host-resolved text selection returned to the model.
type ReadResult struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	LineCount int    `json:"line_count"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated"`
}

// WriteRequest is a normalized mutation request.
type WriteRequest struct {
	Path       string
	Content    string
	Append     bool
	CreateDirs bool
}

// WriteResult describes a committed write.
type WriteResult struct {
	Path         string   `json:"path"`
	BytesWritten int      `json:"bytes_written"`
	Mode         string   `json:"mode"`
	FilesTouched []string `json:"files_touched,omitempty"`
}

// EditRequest is a normalized exact-text mutation request.
type EditRequest struct {
	Path           string
	OldText        string
	NewText        string
	ReplaceAll     bool
	MaxInputBytes  int64
	MaxOutputChars int
}

// EditResult describes a committed edit.
type EditResult struct {
	Path         string   `json:"path"`
	Replacements int      `json:"replacements"`
	FilesTouched []string `json:"files_touched,omitempty"`
}

// ApplyPatchRequest carries a canonical parsed patch to the Host transaction.
type ApplyPatchRequest struct {
	Hunks         []PatchHunk
	MaxInputBytes int64
}

// Register adds all four filesystem tools when a Workspace is available.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.Workspace == nil {
		return
	}
	reg.Register(ReadDefinition(), NewReadHandler(opts))
	reg.Register(WriteDefinition(), NewWriteHandler(opts))
	reg.Register(EditDefinition(), NewEditHandler(opts))
	reg.Register(ApplyPatchDefinition(), NewApplyPatchHandler(opts))
}

// NewReadHandler constructs the read tool implementation.
func NewReadHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		startLine := firstInt(params, "start_line", "from")
		if startLine < 0 {
			startLine = 0
		}
		maxReadChars := positiveOr(opts.MaxReadChars, defaultMaxReadChars)
		maxChars := firstInt(params, "max_chars")
		if maxChars <= 0 || maxChars > maxReadChars {
			maxChars = maxReadChars
		}
		result, err := opts.Workspace.Read(ctx, ReadRequest{
			Path: firstString(params, "path", "file_path"), StartLine: startLine,
			MaxLines: firstInt(params, "max_lines", "lines"), MaxChars: maxChars,
			MaxFileBytes: positiveInt64Or(opts.MaxReadFileBytes, defaultReadFileMaxBytes),
		})
		if err != nil {
			return "", fmt.Errorf("%s: %w", ReadName, err)
		}
		return marshal(result)
	}
}

// NewWriteHandler constructs the write tool implementation.
func NewWriteHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		content := firstRawString(params, "content", "text")
		maxWriteChars := positiveOr(opts.MaxWriteChars, defaultMaxWriteChars)
		if len(content) > maxWriteChars {
			return "", fmt.Errorf("%s: content too large (%d > %d)", WriteName, len(content), maxWriteChars)
		}
		createDirs := true
		if _, ok := params["create_dirs"]; ok {
			createDirs = readBool(params, "create_dirs")
		}
		result, err := opts.Workspace.Write(ctx, WriteRequest{
			Path: firstString(params, "path", "file_path"), Content: content,
			Append: readBool(params, "append"), CreateDirs: createDirs,
		})
		if err != nil {
			return "", err
		}
		return marshal(result)
	}
}

// NewEditHandler constructs the exact-text edit implementation.
func NewEditHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		oldText := firstRawString(params, "old_string", "oldText", "old_text")
		if oldText == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(EditName, []string{"old_string"}, "edit: old_string is required")
		}
		result, err := opts.Workspace.Edit(ctx, EditRequest{
			Path: firstString(params, "path", "file_path"), OldText: oldText,
			NewText:        firstRawString(params, "new_string", "newText", "new_text"),
			ReplaceAll:     readBool(params, "replace_all") || readBool(params, "all"),
			MaxInputBytes:  positiveInt64Or(opts.MaxMutationFileBytes, defaultMutationFileMaxBytes),
			MaxOutputChars: positiveOr(opts.MaxWriteChars, defaultMaxWriteChars),
		})
		if err != nil {
			return "", err
		}
		return marshal(result)
	}
}

// NewApplyPatchHandler constructs the custom-patch implementation.
func NewApplyPatchHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		input := firstString(params, "input", "patch")
		if strings.TrimSpace(input) == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(ApplyPatchName, []string{"input"}, "apply_patch: input is required")
		}
		maxPatchChars := positiveOr(opts.MaxPatchChars, defaultMaxPatchChars)
		if len(input) > maxPatchChars {
			return "", fmt.Errorf("%s: patch too large (%d > %d)", ApplyPatchName, len(input), maxPatchChars)
		}
		hunks, err := ParsePatch(input)
		if err != nil {
			return "", err
		}
		summary, err := opts.Workspace.ApplyPatch(ctx, ApplyPatchRequest{
			Hunks: hunks, MaxInputBytes: positiveInt64Or(opts.MaxMutationFileBytes, defaultMutationFileMaxBytes),
		})
		if err != nil {
			return "", err
		}
		return marshal(struct {
			Added        []string `json:"added,omitempty"`
			Modified     []string `json:"modified,omitempty"`
			Deleted      []string `json:"deleted,omitempty"`
			FilesTouched []string `json:"files_touched,omitempty"`
			Text         string   `json:"text"`
		}{summary.Added, summary.Modified, summary.Deleted, summary.FilesTouched(), summary.Text()})
	}
}

func marshal(value any) (toolcontract.Result, error) {
	blob, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}

func positiveOr(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func positiveInt64Or(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}
