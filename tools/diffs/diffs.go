// Package diffs implements the portable text-only diffs tool.
package diffs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/runtime/toolerrors"
)

// Request is the normalized input to the text-only diff implementation.
type Request struct {
	Before string
	After  string
	Path   string
	Format string
}

// Definition returns the stable model-facing tool declaration.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{
		Type: "function",
		Function: toolcontract.Function{
			Name:        "diffs",
			Description: "Compute a text diff between two provided strings and return a unified diff with line statistics. This tool does not inspect files, write or mutate files, inspect git status, staged changes, or worktree diffs by itself; read the relevant before/after text first, and use write/edit/apply_patch when the task requires changing files. Useful for code review, explaining changes, and comparing versions.",
			Parameters: closedSchema(map[string]any{
				"before": stringSchema("The original text to compare. This must be provided by the caller; the tool will not read or write files or git state."),
				"after":  stringSchema("The modified text to compare. This must be provided by the caller; the tool will not read or write files or git state."),
				"path":   stringSchema("Optional display filename shown in the diff header."),
				"format": enumSchema("Output format: 'unified' (default) for line-by-line diff, 'semantic' for cleanup-optimised output.", "unified", "semantic"),
			}, []string{"before", "after"}),
			OutputSchema: closedSchema(map[string]any{
				"tool":      stringSchema("Tool name that produced the diff."),
				"format":    enumSchema("Effective diff format used by the runtime.", "unified", "semantic"),
				"unified":   stringSchema("Unified diff text with optional display path headers."),
				"additions": intSchema("Number of inserted lines."),
				"deletions": intSchema("Number of deleted lines."),
				"changes":   intSchema("Number of changed hunks or standalone insert/delete changes."),
				"unchanged": intSchema("Number of unchanged lines represented in the diff."),
				"path":      stringSchema("Optional display filename included in diff headers."),
			}, []string{"tool", "format", "unified", "additions", "deletions", "changes", "unchanged"}),
		},
	}
}

// Register adds the diffs implementation to a compatible catalog.
func Register(registrar toolcontract.Registrar) {
	if registrar == nil {
		return
	}
	registrar.Register(Definition(), Execute)
}

// Execute decodes a canonical JSON call and runs the text-only diff.
func Execute(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) {
	parameters := map[string]any{}
	if strings.TrimSpace(call.Arguments) != "" {
		if err := json.Unmarshal([]byte(call.Arguments), &parameters); err != nil {
			return "", toolerrors.NewInvalidJSONToolArgumentError("diffs", fmt.Errorf("decode tool args: %w", err))
		}
	}
	return Run(Request{Before: readString(parameters, "before"), After: readString(parameters, "after"), Path: readString(parameters, "path"), Format: readString(parameters, "format")})
}

// Run executes an already normalized request and returns the existing JSON result bytes.
func Run(request Request) (toolcontract.Result, error) {
	if request.Before == "" && request.After == "" {
		return "", toolerrors.NewMissingRequiredToolArgumentError("diffs", []string{"before", "after"}, "diffs: 'before' and 'after' are required")
	}
	format := strings.TrimSpace(request.Format)
	if format == "" {
		format = "unified"
	}
	if format != "unified" && format != "semantic" {
		return "", toolerrors.NewInvalidToolArgumentError("diffs", []string{"format"}, fmt.Sprintf("diffs: unsupported format %q; use 'unified' or 'semantic'", format))
	}
	unified, additions, deletions, changes, unchanged := compute(request.Before, request.After, request.Path, format)
	payload := map[string]any{
		"tool": "diffs", "format": format, "unified": unified,
		"additions": additions, "deletions": deletions, "changes": changes, "unchanged": unchanged,
	}
	if request.Path != "" {
		payload["path"] = request.Path
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func compute(before, after, path, format string) (string, int, int, int, int) {
	dmp := diffmatchpatch.New()
	beforeChars, afterChars, lineArray := dmp.DiffLinesToChars(before, after)
	differences := dmp.DiffCharsToLines(dmp.DiffMain(beforeChars, afterChars, false), lineArray)
	if format == "semantic" {
		differences = dmp.DiffCleanupSemantic(differences)
	}
	displayA, displayB := "a", "b"
	if path != "" {
		displayA, displayB = "a/"+path, "b/"+path
	}
	var output strings.Builder
	fmt.Fprintf(&output, "--- %s\n+++ %s\n", displayA, displayB)
	additions, deletions, changes, unchanged := 0, 0, 0, 0
	previousDelete := false
	for _, difference := range differences {
		parts := strings.Split(difference.Text, "\n")
		for index, line := range parts {
			if index == len(parts)-1 && line == "" {
				continue
			}
			switch difference.Type {
			case diffmatchpatch.DiffInsert:
				fmt.Fprintf(&output, "+%s\n", line)
				additions++
			case diffmatchpatch.DiffDelete:
				fmt.Fprintf(&output, "-%s\n", line)
				deletions++
			case diffmatchpatch.DiffEqual:
				fmt.Fprintf(&output, " %s\n", line)
				unchanged++
			}
		}
		switch difference.Type {
		case diffmatchpatch.DiffDelete:
			if previousDelete {
				changes++
			}
			previousDelete = true
		case diffmatchpatch.DiffInsert:
			changes++
			previousDelete = false
		case diffmatchpatch.DiffEqual:
			if previousDelete {
				changes++
				previousDelete = false
			}
		}
	}
	if previousDelete {
		changes++
	}
	return output.String(), additions, deletions, changes, unchanged
}

func readString(parameters map[string]any, key string) string {
	value, _ := parameters[key].(string)
	return strings.TrimSpace(value)
}

func closedSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": required}
}
func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func intSchema(description string) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": 0}
}
func enumSchema(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": values}
}
