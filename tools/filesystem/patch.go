package filesystem

import (
	"fmt"
	"strings"
)

// PatchKind identifies a custom patch hunk operation.
type PatchKind string

const (
	// PatchAdd adds a new file.
	PatchAdd PatchKind = "add"
	// PatchDelete removes an existing file.
	PatchDelete PatchKind = "delete"
	// PatchUpdate edits or moves an existing file.
	PatchUpdate PatchKind = "update"
)

// PatchChunk contains the exact old/new line context for one update.
type PatchChunk struct {
	OldLines []string
	NewLines []string
}

// PatchHunk is one parsed custom patch operation.
type PatchHunk struct {
	Kind    PatchKind
	Path    string
	MoveTo  string
	AddBody []string
	Chunks  []PatchChunk
}

// PatchSummary records successfully committed files.
type PatchSummary struct {
	Added    []string
	Modified []string
	Deleted  []string
	Touched  []string
}

// FilesTouched returns a stable first-seen union of changed paths.
func (s PatchSummary) FilesTouched() []string {
	out := make([]string, 0, len(s.Touched)+len(s.Added)+len(s.Modified)+len(s.Deleted))
	seen := map[string]bool{}
	for _, group := range [][]string{s.Touched, s.Added, s.Modified, s.Deleted} {
		for _, path := range group {
			path = strings.TrimSpace(path)
			if path == "" || seen[path] {
				continue
			}
			seen[path] = true
			out = append(out, path)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Text returns the historical human-readable apply-patch summary.
func (s PatchSummary) Text() string {
	lines := []string{"Success. Updated the following files:"}
	for _, path := range s.Added {
		lines = append(lines, "A "+path)
	}
	for _, path := range s.Modified {
		lines = append(lines, "M "+path)
	}
	for _, path := range s.Deleted {
		lines = append(lines, "D "+path)
	}
	return strings.Join(lines, "\n")
}

// ParsePatch parses the AgentX custom *** Begin Patch grammar.
func ParsePatch(input string) ([]PatchHunk, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("apply_patch: input is empty")
	}
	lines := strings.Split(strings.ReplaceAll(trimmed, "\r\n", "\n"), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "*** Begin Patch" || strings.TrimSpace(lines[len(lines)-1]) != "*** End Patch" {
		return nil, fmt.Errorf("apply_patch: invalid boundaries")
	}
	last := len(lines) - 1
	idx := 1
	hunks := make([]PatchHunk, 0)
	for idx < last {
		line := strings.TrimSpace(lines[idx])
		if line == "" {
			idx++
			continue
		}
		switch {
		case strings.HasPrefix(line, "*** Add File: "):
			pathValue := strings.TrimSpace(strings.TrimPrefix(line, "*** Add File: "))
			idx++
			addBody := make([]string, 0)
			for idx < last {
				raw := lines[idx]
				if strings.HasPrefix(raw, "*** ") {
					break
				}
				if !strings.HasPrefix(raw, "+") {
					return nil, fmt.Errorf("apply_patch: invalid add line %d", idx+1)
				}
				addBody = append(addBody, raw[1:])
				idx++
			}
			hunks = append(hunks, PatchHunk{Kind: PatchAdd, Path: pathValue, AddBody: addBody})
		case strings.HasPrefix(line, "*** Delete File: "):
			pathValue := strings.TrimSpace(strings.TrimPrefix(line, "*** Delete File: "))
			idx++
			hunks = append(hunks, PatchHunk{Kind: PatchDelete, Path: pathValue})
		case strings.HasPrefix(line, "*** Update File: "):
			pathValue := strings.TrimSpace(strings.TrimPrefix(line, "*** Update File: "))
			idx++
			moveTo := ""
			if idx < last && strings.HasPrefix(strings.TrimSpace(lines[idx]), "*** Move to: ") {
				moveTo = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[idx]), "*** Move to: "))
				idx++
			}
			chunks := make([]PatchChunk, 0)
			current := PatchChunk{}
			hasChanges := false
			for idx < last {
				raw := lines[idx]
				if strings.HasPrefix(raw, "*** ") {
					break
				}
				trimmedLine := strings.TrimSpace(raw)
				if trimmedLine == "*** End of File" {
					idx++
					continue
				}
				if strings.HasPrefix(trimmedLine, "@@") {
					if hasChanges {
						chunks = append(chunks, current)
						current = PatchChunk{}
						hasChanges = false
					}
					idx++
					continue
				}
				if raw == "" {
					return nil, fmt.Errorf("apply_patch: invalid update line %d", idx+1)
				}
				prefix := raw[0]
				value := ""
				if len(raw) > 1 {
					value = raw[1:]
				}
				switch prefix {
				case ' ':
					current.OldLines = append(current.OldLines, value)
					current.NewLines = append(current.NewLines, value)
					hasChanges = true
				case '-':
					current.OldLines = append(current.OldLines, value)
					hasChanges = true
				case '+':
					current.NewLines = append(current.NewLines, value)
					hasChanges = true
				default:
					return nil, fmt.Errorf("apply_patch: invalid update line %d", idx+1)
				}
				idx++
			}
			if hasChanges {
				chunks = append(chunks, current)
			}
			if len(chunks) == 0 {
				return nil, fmt.Errorf("apply_patch: update hunk without changes")
			}
			hunks = append(hunks, PatchHunk{Kind: PatchUpdate, Path: pathValue, MoveTo: moveTo, Chunks: chunks})
		default:
			return nil, fmt.Errorf("apply_patch: unknown hunk header at line %d", idx+1)
		}
	}
	if len(hunks) == 0 {
		return nil, fmt.Errorf("apply_patch: no hunks")
	}
	return hunks, nil
}

// ApplyUpdateChunk applies one exact and unambiguous text chunk.
func ApplyUpdateChunk(content string, chunk PatchChunk) (string, error) {
	oldText := strings.Join(chunk.OldLines, "\n")
	newText := strings.Join(chunk.NewLines, "\n")
	if oldText == "" {
		if newText == "" {
			return content, nil
		}
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if content == "" {
			return newText, nil
		}
		return content + newText, nil
	}
	replace := func(src, from, to string) (string, bool, error) {
		count := strings.Count(src, from)
		if count == 0 {
			return src, false, nil
		}
		if count > 1 {
			return src, false, fmt.Errorf("ambiguous patch context")
		}
		return strings.Replace(src, from, to, 1), true, nil
	}
	if next, ok, err := replace(content, oldText, newText); err != nil {
		return "", err
	} else if ok {
		return next, nil
	}
	if next, ok, err := replace(content, oldText+"\n", newText+"\n"); err != nil {
		return "", err
	} else if ok {
		return next, nil
	}
	return "", fmt.Errorf("patch context not found")
}
