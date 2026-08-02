package hostkit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterToolsDelegatesOpenTargetToBrowserAct(t *testing.T) {
	exec := &recordingExecutor{output: `{"status":"opened","final_url":"https://example.com/form"}`}
	observer := &recordingTaskObserver{}
	reg := agentxtools.NewRegistry()
	RegisterTools(reg, BuildStandardToolHandlers(Config{Executor: exec, TaskObserver: observer}))

	out, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserOpenTarget,
		Arguments: `{"target_url":"https://example.com/form","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("execute open target: %v", err)
	}
	if out != exec.output {
		t.Fatalf("unexpected output: %s", out)
	}
	call := exec.requireSingleCall(t)
	if call.Name != RuntimeToolBrowserAct {
		t.Fatalf("runtime tool = %q, want %q", call.Name, RuntimeToolBrowserAct)
	}
	args := decodeCallArgs(t, call)
	if args["kind"] != "open" || args["url"] != "https://example.com/form" || args["runtime_target"] != "node" {
		t.Fatalf("unexpected runtime args: %#v", args)
	}
	if args["wait_ms"] != float64(defaultOpenWaitMS) {
		t.Fatalf("wait_ms = %#v, want %d", args["wait_ms"], defaultOpenWaitMS)
	}
	if len(observer.events) != 1 {
		t.Fatalf("expected one task observation, got %#v", observer.events)
	}
	event := observer.events[0]
	if event.SemanticTool != ToolBrowserOpenTarget || event.RuntimeTool != RuntimeToolBrowserAct || event.RuntimeKind != "open" || event.Status != "observed" || event.AdapterStatus != "opened" {
		t.Fatalf("unexpected task observation: %#v", event)
	}
}

func TestRegisterToolsDelegatesFillFieldsToBrowserAct(t *testing.T) {
	exec := &recordingExecutor{output: `{"status":"filled","field_count":1,"submitted":true}`}
	reg := agentxtools.NewRegistry()
	RegisterTools(reg, BuildStandardToolHandlers(Config{Executor: exec}))

	_, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserFillFields,
		Arguments: `{"fields":[{"selector":"input[name=email]","value":"a@example.com"}],"submit":true,"target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("execute fill fields: %v", err)
	}
	args := decodeCallArgs(t, exec.requireSingleCall(t))
	if args["kind"] != "fill" || args["target"] != "tab:2" || args["submit"] != true {
		t.Fatalf("unexpected runtime args: %#v", args)
	}
	fields, ok := args["fields"].([]any)
	if !ok || len(fields) != 1 {
		t.Fatalf("unexpected fields payload: %#v", args["fields"])
	}
}

func TestCaptureEvidenceCanUseBrowserActScreenshotRuntime(t *testing.T) {
	exec := &recordingExecutor{output: `{"status":"captured","path":".agentx/browser/shot.png"}`}
	reg := agentxtools.NewRegistry()
	RegisterTools(reg, BuildStandardToolHandlers(Config{
		Executor:              exec,
		RuntimeScreenshotTool: RuntimeToolBrowserAct,
	}))

	_, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserCaptureSubmissionEvidence,
		Arguments: `{"target":"current","path":"shots/final.png"}`,
	})
	if err != nil {
		t.Fatalf("execute capture evidence: %v", err)
	}
	call := exec.requireSingleCall(t)
	if call.Name != RuntimeToolBrowserAct {
		t.Fatalf("runtime tool = %q, want %q", call.Name, RuntimeToolBrowserAct)
	}
	args := decodeCallArgs(t, call)
	if args["kind"] != "screenshot" || args["full_page"] != true || args["path"] != "shots/final.png" {
		t.Fatalf("unexpected runtime args: %#v", args)
	}
}

func TestRegisterToolsDelegatesDownloadFileToBrowserAct(t *testing.T) {
	exec := &recordingExecutor{output: `{"status":"downloaded","path":".agentx/browserd/artifacts/download.csv"}`}
	reg := agentxtools.NewRegistry()
	RegisterTools(reg, BuildStandardToolHandlers(Config{Executor: exec}))

	_, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserDownloadFile,
		Arguments: `{"mode":"download","url":"https://example.com/report.csv","output_path":"reports/report.csv","force":true}`,
	})
	if err != nil {
		t.Fatalf("execute download file: %v", err)
	}
	call := exec.requireSingleCall(t)
	if call.Name != RuntimeToolBrowserAct {
		t.Fatalf("runtime tool = %q, want %q", call.Name, RuntimeToolBrowserAct)
	}
	args := decodeCallArgs(t, call)
	if args["kind"] != "download" || args["url"] != "https://example.com/report.csv" || args["output_path"] != "reports/report.csv" || args["force"] != true {
		t.Fatalf("unexpected runtime args: %#v", args)
	}
}

func TestRegisterToolsDelegatesWaitDownloadFileToBrowserAct(t *testing.T) {
	exec := &recordingExecutor{output: `{"status":"downloaded","path":".agentx/browserd/artifacts/waited.csv"}`}
	reg := agentxtools.NewRegistry()
	RegisterTools(reg, BuildStandardToolHandlers(Config{Executor: exec}))

	_, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserDownloadFile,
		Arguments: `{"mode":"wait_download","wait_ms":30000,"allow_recent_download_reuse":true}`,
	})
	if err != nil {
		t.Fatalf("execute wait download file: %v", err)
	}
	args := decodeCallArgs(t, exec.requireSingleCall(t))
	if args["kind"] != "wait_download" || args["wait_ms"] != float64(30000) || args["allow_recent_download_reuse"] != true {
		t.Fatalf("unexpected runtime args: %#v", args)
	}
}

func TestMissingExecutorReturnsStructuredUnsupportedPayload(t *testing.T) {
	observer := &recordingTaskObserver{}
	reg := agentxtools.NewRegistry()
	cfg := DefaultConfig()
	cfg.TaskObserver = observer
	RegisterTools(reg, BuildStandardToolHandlers(cfg))

	out, err := reg.Execute(context.Background(), llm.FunctionCall{
		Name:      ToolBrowserFillFields,
		Arguments: `{"fields":[{"selector":"input","value":"x"}],"submit":true}`,
	})
	if err != nil {
		t.Fatalf("execute without runtime executor: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode payload: %v raw=%s", err, out)
	}
	if payload["status"] != "unsupported" || payload["failure_code"] != "browser_runtime_executor_not_configured" || payload["missing"] != "livekit.executor" {
		t.Fatalf("unexpected unsupported payload: %#v", payload)
	}
	if payload["field_count"] != float64(0) || payload["submitted"] != true {
		t.Fatalf("expected fill compatibility fields, got %#v", payload)
	}
	if len(observer.events) != 1 || observer.events[0].Status != "unsupported" || observer.events[0].FailureCode != "browser_runtime_executor_not_configured" {
		t.Fatalf("unexpected unsupported observation: %#v", observer.events)
	}
}

func TestToolDefinitionsExposeBrowserOpsSurface(t *testing.T) {
	if BrowserOpenTargetTool().Function.Name != ToolBrowserOpenTarget {
		t.Fatalf("unexpected open target definition")
	}
	if BrowserFillFieldsTool().Function.Name != ToolBrowserFillFields {
		t.Fatalf("unexpected fill fields definition")
	}
	if BrowserDownloadFileTool().Function.Name != ToolBrowserDownloadFile {
		t.Fatalf("unexpected download file definition")
	}
	if got := ToolNames(); len(got) != 5 || got[0] != ToolBrowserOpenTarget || got[4] != ToolBrowserDownloadFile {
		t.Fatalf("unexpected tool names: %#v", got)
	}
}

func TestToolDefinitionsDescribeParameters(t *testing.T) {
	for _, tool := range []llm.Tool{
		BrowserOpenTargetTool(),
		BrowserFillFieldsTool(),
		BrowserCapturePageSnapshotTool(),
		BrowserCaptureSubmissionEvidenceTool(),
		BrowserDownloadFileTool(),
	} {
		if strings.TrimSpace(tool.Function.Description) == "" {
			t.Fatalf("%s missing description", tool.Function.Name)
		}
		assertToolSchemaDescriptions(t, tool.Function.Name, tool.Function.Parameters, "")
	}
}

func assertToolSchemaDescriptions(t *testing.T, owner string, schema map[string]any, prefix string) {
	t.Helper()
	props, _ := schema["properties"].(map[string]any)
	for name, raw := range props {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s property %s should be an object schema", owner, path)
		}
		if strings.TrimSpace(readToolSchemaString(prop["description"])) == "" {
			t.Fatalf("%s property %s missing description", owner, path)
		}
		if nested, _ := prop["properties"].(map[string]any); len(nested) != 0 {
			assertToolSchemaDescriptions(t, owner, prop, path)
		}
		if items, _ := prop["items"].(map[string]any); items != nil {
			if nested, _ := items["properties"].(map[string]any); len(nested) != 0 {
				assertToolSchemaDescriptions(t, owner, items, path+"[]")
			}
		}
	}
}

func readToolSchemaString(value any) string {
	out, _ := value.(string)
	return out
}

type recordingExecutor struct {
	calls  []llm.FunctionCall
	output string
	err    error
}

type recordingTaskObserver struct {
	events []TaskObservation
}

func (o *recordingTaskObserver) ObserveBrowserTask(_ context.Context, observation TaskObservation) {
	o.events = append(o.events, observation)
}

func (e *recordingExecutor) Execute(_ context.Context, call llm.FunctionCall) (string, error) {
	e.calls = append(e.calls, call)
	return e.output, e.err
}

func (e *recordingExecutor) requireSingleCall(t *testing.T) llm.FunctionCall {
	t.Helper()
	if len(e.calls) != 1 {
		t.Fatalf("expected one runtime call, got %#v", e.calls)
	}
	return e.calls[0]
}

func decodeCallArgs(t *testing.T, call llm.FunctionCall) map[string]any {
	t.Helper()
	var args map[string]any
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		t.Fatalf("decode runtime args: %v raw=%s", err, call.Arguments)
	}
	return args
}
