package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	channelruntime "github.com/wsnacj/agentx-go/runtime/channel"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/diffs"
	"github.com/wsnacj/agentx-go/tools/filesystem"
	"github.com/wsnacj/agentx-go/tools/httprequest"
	"github.com/wsnacj/agentx-go/tools/memory"
	"github.com/wsnacj/agentx-go/tools/message"
	"github.com/wsnacj/agentx-go/tools/scheduler"
)

type result struct {
	Registered []string `json:"registered"`
	Executed   []string `json:"executed"`
	Verified   bool     `json:"verified"`
}

type recordingSender struct {
	target channelruntime.TextTarget
	text   string
}

func (s *recordingSender) SendText(ctx context.Context, target channelruntime.TextTarget, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.target, s.text = target, text
	return nil
}

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

type memoryWorkspace struct{ files map[string]string }

func newMemoryWorkspace() *memoryWorkspace {
	return &memoryWorkspace{files: map[string]string{"note.txt": "before\n"}}
}

func (w *memoryWorkspace) Read(ctx context.Context, request filesystem.ReadRequest) (filesystem.ReadResult, error) {
	if err := ctx.Err(); err != nil {
		return filesystem.ReadResult{}, err
	}
	content, ok := w.files[request.Path]
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
	var summary filesystem.PatchSummary
	for _, hunk := range request.Hunks {
		if hunk.Kind != filesystem.PatchUpdate {
			return filesystem.PatchSummary{}, fmt.Errorf("consumer accepts update hunks only")
		}
		content, ok := w.files[hunk.Path]
		if !ok {
			return filesystem.PatchSummary{}, fmt.Errorf("file not found")
		}
		var err error
		for _, chunk := range hunk.Chunks {
			content, err = filesystem.ApplyUpdateChunk(content, chunk)
			if err != nil {
				return filesystem.PatchSummary{}, err
			}
		}
		w.files[hunk.Path] = content
		summary.Modified = append(summary.Modified, hunk.Path)
		summary.Touched = append(summary.Touched, hunk.Path)
	}
	return summary, nil
}

func registerPortableTools(registry *tools.Registry, sender *recordingSender, workspace *memoryWorkspace) {
	diffs.Register(registry)
	message.Register(registry, message.Options{
		Sender: sender, Target: channelruntime.TextTarget{ChatID: "consumer-chat"}, Platform: "memory",
	})
	httprequest.Register(registry, httprequest.Options{Prepare: func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		parsed, err := url.Parse(input.RawURL)
		if err != nil || parsed.Hostname() != "example.test" {
			return httprequest.PreparedRequest{}, fmt.Errorf("URL is outside the consumer fixture")
		}
		return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"fixture":true}`)),
				Request:    request,
			}, nil
		})}, nil
	}})
	filesystem.Register(registry, filesystem.Options{Workspace: workspace})
	memory.Register(registry, memory.Options{Backend: memory.BackendFuncs{
		SearchFunc: func(ctx context.Context, request memory.SearchRequest) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			return fmt.Sprintf(`{"query":%q,"hits":[]}`, request.Query), nil
		},
		GetFunc: func(ctx context.Context, request memory.GetRequest) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			return fmt.Sprintf(`{"path":%q,"text":"ready"}`, request.Path), nil
		},
	}})
	handleSchedule := func(ctx context.Context, request scheduler.Request) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"action":%q,"jobs":[]}`, request.Action), nil
	}
	scheduler.Register(registry, scheduler.BackendFuncs{
		AddFunc: handleSchedule, ListFunc: handleSchedule, StatusFunc: handleSchedule,
		RunFunc: handleSchedule, RemoveFunc: handleSchedule,
	})
}

func execute(ctx context.Context, registry *tools.Registry, name, arguments string) (string, error) {
	value, err := registry.Execute(ctx, toolcontract.Call{Name: name, Arguments: arguments})
	if err != nil {
		return "", fmt.Errorf("execute %s: %w", name, err)
	}
	return value, nil
}

func run(ctx context.Context) (result, error) {
	registry := tools.NewRegistry()
	sender := &recordingSender{}
	workspace := newMemoryWorkspace()
	registerPortableTools(registry, sender, workspace)

	definitions := registry.Definitions()
	registered := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		registered = append(registered, definition.Function.Name)
	}

	executed := make([]string, 0, 10)
	run := func(name, arguments string) (string, error) {
		value, err := execute(ctx, registry, name, arguments)
		if err == nil {
			executed = append(executed, name)
		}
		return value, err
	}
	if _, err := run(message.Name, `{"action":"send","text":"ready"}`); err != nil {
		return result{}, err
	}
	if _, err := run(httprequest.Name, `{"url":"https://example.test/status"}`); err != nil {
		return result{}, err
	}
	if _, err := run(filesystem.WriteName, `{"path":"draft.txt","content":"before\n"}`); err != nil {
		return result{}, err
	}
	if _, err := run(filesystem.EditName, `{"path":"draft.txt","old_string":"before","new_string":"after"}`); err != nil {
		return result{}, err
	}
	readResult, err := run(filesystem.ReadName, `{"path":"draft.txt"}`)
	if err != nil {
		return result{}, err
	}
	patch, _ := json.Marshal(map[string]string{"input": "*** Begin Patch\n*** Update File: note.txt\n@@\n-before\n+after\n*** End Patch"})
	if _, err := run(filesystem.ApplyPatchName, string(patch)); err != nil {
		return result{}, err
	}
	if _, err := run(memory.SearchName, `{"query":"release"}`); err != nil {
		return result{}, err
	}
	if _, err := run(memory.GetName, `{"path":"MEMORY.md"}`); err != nil {
		return result{}, err
	}
	if _, err := run(scheduler.Name, `{"action":"list"}`); err != nil {
		return result{}, err
	}
	diffResult, err := run("diffs", `{"before":"old\n","after":"new\n","path":"sample.txt"}`)
	if err != nil {
		return result{}, err
	}

	sort.Strings(executed)
	verified := len(registered) == 10 && len(executed) == 10 &&
		sender.target.ChatID == "consumer-chat" && sender.text == "ready" &&
		workspace.files["draft.txt"] == "after\n" && workspace.files["note.txt"] == "after\n" &&
		strings.Contains(readResult, `"content":"after`) && strings.Contains(diffResult, `"tool":"diffs"`)
	return result{Registered: registered, Executed: executed, Verified: verified}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
