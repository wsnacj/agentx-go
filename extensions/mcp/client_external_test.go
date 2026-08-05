package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/extensions/mcp"
)

func TestClientLifecycleDiscoversAndExecutesTools(t *testing.T) {
	transport := newFakeTransport()
	client := newClient(t, transport)
	state, err := client.Initialize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !state.Initialized || state.ProtocolVersion != mcp.ProtocolVersion20251125 || state.Capabilities.Tools == nil {
		t.Fatalf("state=%#v", state)
	}
	set, err := client.DiscoverTools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	definitions := set.Definitions()
	if len(definitions) != 2 || definitions[0].Function.Name != "Echo" || definitions[1].Function.Name != "sum" {
		t.Fatalf("definitions=%#v", definitions)
	}
	output, err := set.Execute(context.Background(), toolcontract.Call{Name: "Echo", Arguments: `{"text":"hello"}`})
	if err != nil {
		t.Fatal(err)
	}
	var result mcp.CallToolResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 || result.Content[0].Type() != "text" || result.Content[0].Text() != "hello" {
		t.Fatalf("result=%#v", result)
	}
	if transport.notificationMethods()[0] != "notifications/initialized" {
		t.Fatalf("notifications=%#v", transport.notificationMethods())
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Snapshot(context.Background()); !errors.Is(err, &mcp.Error{Code: mcp.ErrorCodeClosed}) {
		t.Fatalf("post-close err=%v", err)
	}
}

func TestClientRejectsVersionRemoteErrorAndUnavailableTool(t *testing.T) {
	transport := newFakeTransport()
	transport.version = "2099-01-01"
	client := newClient(t, transport)
	if _, err := client.Initialize(context.Background()); !errors.Is(err, &mcp.Error{Code: mcp.ErrorCodeProtocol}) {
		t.Fatalf("version err=%v", err)
	}
	transport = newFakeTransport()
	client = newClient(t, transport)
	if _, err := client.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	set, err := client.DiscoverTools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := set.Execute(context.Background(), toolcontract.Call{Name: "missing"}); !errors.Is(err, &mcp.Error{Code: mcp.ErrorCodeToolUnavailable}) {
		t.Fatalf("unavailable err=%v", err)
	}
	transport.remoteError = true
	if _, err := set.Execute(context.Background(), toolcontract.Call{Name: "Echo", Arguments: `{}`}); !errors.Is(err, &mcp.Error{Code: mcp.ErrorCodeRemote}) {
		t.Fatalf("remote err=%v", err)
	}
}

func TestClientCancellationSendsNotification(t *testing.T) {
	transport := newFakeTransport()
	client := newClient(t, transport)
	if _, err := client.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	set, err := client.DiscoverTools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	transport.blockCalls = true
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := set.Execute(ctx, toolcontract.Call{Name: "Echo", Arguments: `{}`}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	methods := transport.notificationMethods()
	if methods[len(methods)-1] != "notifications/cancelled" {
		t.Fatalf("notifications=%#v", methods)
	}
}

func newClient(t *testing.T, transport mcp.Transport) *mcp.Client {
	t.Helper()
	client, err := mcp.New(mcp.Config{
		Transport:          transport,
		ClientInfo:         mcp.Implementation{Name: "agentx-test", Version: "0.1.0"},
		RequestTimeout:     time.Second,
		CancellationWindow: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

type fakeTransport struct {
	mu            sync.Mutex
	version       string
	notifications []mcp.Notification
	remoteError   bool
	blockCalls    bool
	shutdowns     int
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{version: mcp.ProtocolVersion20251125}
}

func (f *fakeTransport) Request(ctx context.Context, request mcp.Request) (mcp.Response, error) {
	f.mu.Lock()
	remoteError := f.remoteError
	blockCalls := f.blockCalls
	version := f.version
	f.mu.Unlock()
	if request.Method == "tools/call" && blockCalls {
		<-ctx.Done()
		return mcp.Response{}, ctx.Err()
	}
	if request.Method == "tools/call" && remoteError {
		return mcp.Response{JSONRPC: "2.0", ID: request.ID, Error: &mcp.RPCError{Code: -32000, Message: "untrusted remote detail"}}, nil
	}
	var result any
	switch request.Method {
	case "initialize":
		result = mcp.InitializeResult{
			ProtocolVersion: version,
			Capabilities:    mcp.ServerCapabilities{Tools: &mcp.ToolCapability{ListChanged: true}},
			ServerInfo:      mcp.Implementation{Name: "fixture", Version: "1.0.0"},
			Instructions:    "untrusted server instructions",
		}
	case "tools/list":
		var params mcp.ListToolsParams
		_ = json.Unmarshal(request.Params, &params)
		if params.Cursor == "" {
			result = mcp.ListToolsResult{Tools: []mcp.Tool{{Name: "sum", InputSchema: objectSchema()}}, NextCursor: "page-2"}
		} else {
			result = mcp.ListToolsResult{Tools: []mcp.Tool{{Name: "Echo", Description: "echo text", InputSchema: objectSchema()}}}
		}
	case "tools/call":
		var params mcp.CallToolParams
		_ = json.Unmarshal(request.Params, &params)
		text, _ := params.Arguments["text"].(string)
		result = mcp.CallToolResult{Content: []mcp.ContentBlock{{"type": "text", "text": text}}}
	default:
		return mcp.Response{JSONRPC: "2.0", ID: request.ID, Error: &mcp.RPCError{Code: -32601, Message: "method not found"}}, nil
	}
	raw, _ := json.Marshal(result)
	return mcp.Response{JSONRPC: "2.0", ID: request.ID, Result: raw}, nil
}

func (f *fakeTransport) Notify(_ context.Context, notification mcp.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifications = append(f.notifications, notification)
	return nil
}

func (f *fakeTransport) Shutdown(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shutdowns++
	return nil
}

func (f *fakeTransport) notificationMethods() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.notifications))
	for _, notification := range f.notifications {
		out = append(out, notification.Method)
	}
	return out
}

func objectSchema() map[string]any { return map[string]any{"type": "object"} }
