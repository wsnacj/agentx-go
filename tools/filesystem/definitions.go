package filesystem

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

// ReadDefinition returns the stable read tool schema.
func ReadDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        ReadName,
		Description: "Read UTF-8 text from one workspace file or registered immutable assetfs:// URI. Use this for precise file inspection; use start_line/max_lines/max_chars to limit large files. This tool is read-only and never reports files_touched.",
		Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{
				"path":       stringSchema("Workspace-relative/absolute file path, or registered immutable assetfs:// URI. Prefer workspace-relative paths in normal use."),
				"start_line": intSchema("Zero-based line offset to start reading from. Omit to start at the beginning of the file.", 0),
				"max_lines":  intSchema("Maximum number of lines to return. Omit to allow the runtime default line selection.", 1),
				"max_chars":  intSchema("Maximum number of characters to return after line slicing. The runtime clamps this to its configured safety limit.", 1),
			}, "required": []string{"path"},
		},
		OutputSchema: readOutputSchema(),
	}}
}

// WriteDefinition returns the stable write tool schema.
func WriteDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        WriteName,
		Description: "Write UTF-8 text to a workspace file. This mutates the workspace, creates parent directories by default, and returns files_touched only after a successful write.",
		Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{
				"path":        stringSchema("Workspace-relative or absolute file path to write. Prefer workspace-relative paths in normal use."),
				"content":     stringSchema("Exact text content to write. In overwrite mode this replaces the whole file."),
				"append":      boolSchema("When true, append content to the file instead of overwriting it. Defaults to false."),
				"create_dirs": boolSchema("When true, create missing parent directories before writing. Defaults to true."),
			}, "required": []string{"path", "content"},
		},
		OutputSchema: writeOutputSchema(),
	}}
}

// EditDefinition returns the stable exact-text edit schema.
func EditDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        EditName,
		Description: "Replace exact text in a workspace file. This mutates the workspace and returns files_touched only after the replacement is written successfully.",
		Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{
				"path":        stringSchema("Workspace-relative or absolute file path to edit. Prefer workspace-relative paths in normal use."),
				"old_string":  stringSchema("Exact text to replace. The call fails if this text is not present."),
				"new_string":  stringSchema("Replacement text to write in place of old_string."),
				"replace_all": boolSchema("When true, replace every occurrence of old_string. Defaults to false, which replaces only the first occurrence."),
			}, "required": []string{"path", "old_string", "new_string"},
		},
		OutputSchema: editOutputSchema(),
	}}
}

// ApplyPatchDefinition returns the stable custom-patch schema.
func ApplyPatchDefinition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        ApplyPatchName,
		Description: "Apply a workspace patch using the custom *** Begin Patch / *** End Patch grammar. This is not git diff format: do not include diff --git, index, ---, or +++ headers. Use hunks that start with *** Add File:, *** Update File:, or *** Delete File:. This mutates files only when the patch is valid and returns added/modified/deleted/files_touched after a successful apply.",
		Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{
				"input": stringSchema("Patch text using the exact custom grammar: start with *** Begin Patch, then one or more *** Add File:/*** Update File:/*** Delete File: hunks, and finish with *** End Patch. Do not pass git diff output."),
			}, "required": []string{"input"},
		},
		OutputSchema: applyPatchOutputSchema(),
	}}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "string"}}
}

func readOutputSchema() map[string]any {
	return closedSchema(map[string]any{
		"path":       stringSchema("Resolved display path for the file that was read."),
		"start_line": intSchema("Zero-based start line represented in content.", 0),
		"line_count": intSchema("Number of lines represented in content.", 0),
		"content":    stringSchema("Text content returned after line and character limits."),
		"truncated":  boolSchema("True when content was trimmed to max_chars or the runtime read limit."),
	}, []string{"path", "start_line", "line_count", "content", "truncated"})
}

func writeOutputSchema() map[string]any {
	return closedSchema(map[string]any{
		"path":          stringSchema("Resolved display path for the file that was written."),
		"bytes_written": intSchema("Number of bytes written from the provided content.", 0),
		"mode":          stringSchema("Write mode used by the tool: overwrite or append."),
		"files_touched": stringArraySchema("Workspace files that were actually written by this operation."),
	}, []string{"path", "bytes_written", "mode", "files_touched"})
}

func editOutputSchema() map[string]any {
	return closedSchema(map[string]any{
		"path":          stringSchema("Resolved display path for the file that was edited."),
		"replacements":  intSchema("Number of replacements written to the file.", 0),
		"files_touched": stringArraySchema("Workspace files that were actually written by this operation."),
	}, []string{"path", "replacements", "files_touched"})
}

func applyPatchOutputSchema() map[string]any {
	return closedSchema(map[string]any{
		"added":         stringArraySchema("Files added by the patch."),
		"modified":      stringArraySchema("Files modified or moved by the patch."),
		"deleted":       stringArraySchema("Files deleted by the patch."),
		"files_touched": stringArraySchema("Workspace files that were actually added, modified, moved, or deleted by this operation."),
		"text":          stringSchema("Human-readable patch summary."),
	}, []string{"text"})
}

func closedSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": required}
}
