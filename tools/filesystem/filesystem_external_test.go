package filesystem_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/filesystem"
)

type memoryWorkspace struct {
	mu    sync.Mutex
	files map[string]string
}

func newMemoryWorkspace(files map[string]string) *memoryWorkspace {
	copyFiles := make(map[string]string, len(files))
	for path, content := range files {
		copyFiles[path] = content
	}
	return &memoryWorkspace{files: copyFiles}
}

func (w *memoryWorkspace) Read(ctx context.Context, request filesystem.ReadRequest) (filesystem.ReadResult, error) {
	if err := ctx.Err(); err != nil {
		return filesystem.ReadResult{}, err
	}
	w.mu.Lock()
	content, ok := w.files[request.Path]
	w.mu.Unlock()
	if !ok {
		return filesystem.ReadResult{}, fmt.Errorf("file not found")
	}
	selected, start, lines, truncated, err := filesystem.SelectText(bytes.NewBufferString(content), request.StartLine, request.MaxLines, request.MaxChars)
	return filesystem.ReadResult{Path: request.Path, StartLine: start, LineCount: lines, Content: selected, Truncated: truncated}, err
}

func (w *memoryWorkspace) Write(ctx context.Context, request filesystem.WriteRequest) (filesystem.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return filesystem.WriteResult{}, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	mode := "overwrite"
	if request.Append {
		w.files[request.Path] += request.Content
		mode = "append"
	} else {
		w.files[request.Path] = request.Content
	}
	return filesystem.WriteResult{Path: request.Path, BytesWritten: len(request.Content), Mode: mode, FilesTouched: []string{request.Path}}, nil
}

func (w *memoryWorkspace) Edit(ctx context.Context, request filesystem.EditRequest) (filesystem.EditResult, error) {
	if err := ctx.Err(); err != nil {
		return filesystem.EditResult{}, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	content, ok := w.files[request.Path]
	if !ok {
		return filesystem.EditResult{}, fmt.Errorf("file not found")
	}
	next, replacements, err := filesystem.EditText(content, request.OldText, request.NewText, request.ReplaceAll, request.MaxOutputChars)
	if err != nil {
		return filesystem.EditResult{}, err
	}
	w.files[request.Path] = next
	return filesystem.EditResult{Path: request.Path, Replacements: replacements, FilesTouched: []string{request.Path}}, nil
}

func (w *memoryWorkspace) ApplyPatch(ctx context.Context, request filesystem.ApplyPatchRequest) (filesystem.PatchSummary, error) {
	if err := ctx.Err(); err != nil {
		return filesystem.PatchSummary{}, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	next := make(map[string]string, len(w.files))
	for path, content := range w.files {
		next[path] = content
	}
	var summary filesystem.PatchSummary
	for _, hunk := range request.Hunks {
		switch hunk.Kind {
		case filesystem.PatchAdd:
			if _, exists := next[hunk.Path]; exists {
				return filesystem.PatchSummary{}, fmt.Errorf("target already exists")
			}
			content := strings.Join(hunk.AddBody, "\n")
			if len(hunk.AddBody) > 0 {
				content += "\n"
			}
			next[hunk.Path] = content
			summary.Added = append(summary.Added, hunk.Path)
			summary.Touched = append(summary.Touched, hunk.Path)
		case filesystem.PatchDelete:
			if _, exists := next[hunk.Path]; !exists {
				return filesystem.PatchSummary{}, fmt.Errorf("source does not exist")
			}
			delete(next, hunk.Path)
			summary.Deleted = append(summary.Deleted, hunk.Path)
			summary.Touched = append(summary.Touched, hunk.Path)
		case filesystem.PatchUpdate:
			content, exists := next[hunk.Path]
			if !exists {
				return filesystem.PatchSummary{}, fmt.Errorf("source does not exist")
			}
			var err error
			for _, chunk := range hunk.Chunks {
				content, err = filesystem.ApplyUpdateChunk(content, chunk)
				if err != nil {
					return filesystem.PatchSummary{}, err
				}
			}
			target := hunk.Path
			if strings.TrimSpace(hunk.MoveTo) != "" {
				target = hunk.MoveTo
				delete(next, hunk.Path)
				summary.Touched = append(summary.Touched, hunk.Path)
			}
			next[target] = content
			summary.Modified = append(summary.Modified, target)
			summary.Touched = append(summary.Touched, target)
		}
	}
	w.files = next
	return summary, nil
}

func (w *memoryWorkspace) content(path string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.files[path]
}

func TestExternalConsumerRunsAllFilesystemTools(t *testing.T) {
	workspace := newMemoryWorkspace(map[string]string{"notes.txt": "one\ntwo\n", "old.txt": "remove\n"})
	registry := agentxtools.NewRegistry()
	filesystem.Register(registry, filesystem.Options{Workspace: workspace})

	if _, err := registry.Execute(context.Background(), toolcontract.Call{Name: "write", Arguments: `{"file_path":"new.txt","text":"before\n"}`}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := registry.Execute(context.Background(), toolcontract.Call{Name: "edit", Arguments: `{"path":"new.txt","old_string":"before","new_string":"after"}`}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	read, err := registry.Execute(context.Background(), toolcontract.Call{Name: "read", Arguments: `{"path":"new.txt","max_lines":1}`})
	if err != nil || !strings.Contains(read, `"content":"after"`) {
		t.Fatalf("read=%s err=%v", read, err)
	}
	patch := "*** Begin Patch\n*** Update File: notes.txt\n@@\n-two\n+second\n*** Delete File: old.txt\n*** End Patch"
	result, err := registry.Execute(context.Background(), toolcontract.Call{Name: "apply_patch", Arguments: `{"input":` + fmt.Sprintf("%q", patch) + `}`})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !strings.Contains(result, `"modified":["notes.txt"]`) || workspace.content("notes.txt") != "one\nsecond\n" || workspace.content("old.txt") != "" {
		t.Fatalf("result=%s files=%#v", result, workspace.files)
	}
	if got := workspace.content("new.txt"); got != "after\n" {
		t.Fatalf("new.txt=%q", got)
	}
}

func TestDefinitionsAndCancellationRemainStable(t *testing.T) {
	registry := agentxtools.NewRegistry()
	filesystem.Register(registry, filesystem.Options{Workspace: newMemoryWorkspace(nil)})
	got := make([]string, 0, 4)
	for _, definition := range registry.Definitions() {
		got = append(got, definition.Function.Name)
	}
	if want := []string{"apply_patch", "edit", "read", "write"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("definitions=%v want=%v", got, want)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := registry.Execute(ctx, toolcontract.Call{Name: "write", Arguments: `{"path":"x","content":"x"}`}); err != context.Canceled {
		t.Fatalf("cancel error=%v", err)
	}
}
