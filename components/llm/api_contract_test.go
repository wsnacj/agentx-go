package llm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestW301ExactExportSurface(t *testing.T) {
	t.Helper()
	types, funcs, methods, constants, imports := collectPackageContract(t)

	assertExactStrings(t, "types", types, []string{
		"BotInput", "BotReference", "BotResponse", "BotUsage", "BotUsageAction",
		"BotUsageModel", "ChatInput", "ChatRequest", "ChatResponse", "Conversation",
		"EmbedInput", "EmbeddingOptions", "EmbeddingRequest", "EmbeddingResponse",
		"EventStreamResult", "Function", "FunctionCall", "FunctionCallDelta",
		"FunctionResult", "Message", "PayloadHook", "ReasoningOptions",
		"RequestOptions", "ResponseHook", "ResponseMetadata", "SimpleStreamChunk",
		"SimpleStreamResult", "SparseEntry", "StreamChunk", "StreamEvent",
		"StreamEventType", "StreamMessageSnapshot", "StreamResult", "StreamStopReason",
		"ThinkingOptions", "Tool", "ToolChoice", "ToolChoiceFunction", "Usage",
		"UsageRecord", "VisionInput", "VisualContent", "VisualOption", "VisualRequest",
		"VisualResponse",
	})
	assertExactStrings(t, "functions", funcs, []string{
		"BridgeEventStreamResult", "BridgeEventStreamToSimple", "BridgeLegacyStreamResult",
		"BuildStreamMessageSnapshot", "EmbeddingOptionsFromMap", "MergeToolCallSnapshot",
		"NewImageList", "NewImageURL", "NewLocalImage", "NewTextBlock", "NewVideoURL",
		"NormalizeStreamStopReason", "RequestOptionsFromMap",
		"SanitizeFunctionParametersSchema", "SanitizeProviderOptionMap",
		"SanitizeToolSchemas", "SortedFunctionCallIndexes", "SortedStringSnapshotIndexes",
		"WithDetail", "WithFPS", "WithLabels",
	})
	assertExactStrings(t, "methods", methods, []string{
		"BotInput.Clone", "ChatInput.Clone", "EmbedInput.Clone",
		"EmbeddingOptions.Clone", "EmbeddingOptions.ToMap",
		"FunctionCallDelta.HasArguments", "FunctionCallDelta.HasName",
		"RequestOptions.Clone", "RequestOptions.ToMap", "VisionInput.Clone",
	})
	assertExactStrings(t, "constants", constants, []string{
		"DetailAuto", "DetailHigh", "StreamEventDone", "StreamEventError",
		"StreamEventStart", "StreamEventTextDelta", "StreamEventTextEnd",
		"StreamEventTextStart", "StreamEventThinkingDelta", "StreamEventThinkingEnd",
		"StreamEventThinkingStart", "StreamEventToolCallDelta", "StreamEventToolCallEnd",
		"StreamEventToolCallStart", "StreamEventUsage", "StreamStopReasonContentFilter",
		"StreamStopReasonLength", "StreamStopReasonStop", "StreamStopReasonToolUse",
	})
	assertExactStrings(t, "production imports", imports, []string{
		"context", "slices", "strings", "time",
	})
}

func TestW301ChineseReferenceCoversExactSurface(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatalf("read API.md: %v", err)
	}
	text := string(content)
	types, funcs, methods, constants, _ := collectPackageContract(t)
	for _, name := range append(append(append(types, funcs...), methods...), constants...) {
		if !strings.Contains(text, name) {
			t.Errorf("API.md does not mention exported API %s", name)
		}
	}
	for _, required := range []string{
		"private validation / Experimental",
		"不创建 AgentX `Client` 或 Runtime",
		"Callback 与 context",
		"Stream 与 Cancel",
		"明确 non-goal",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("API.md missing required boundary %q", required)
		}
	}
}

func collectPackageContract(t *testing.T) (
	types []string,
	funcs []string,
	methods []string,
	constants []string,
	imports []string,
) {
	t.Helper()
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob package files: %v", err)
	}
	importSet := map[string]struct{}{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquote import %s in %s: %v", spec.Path.Value, entry, err)
			}
			importSet[path] = struct{}{}
		}
		for _, decl := range file.Decls {
			switch value := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					switch typed := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(typed.Name.Name) {
							types = append(types, typed.Name.Name)
						}
					case *ast.ValueSpec:
						if value.Tok != token.CONST {
							continue
						}
						for _, name := range typed.Names {
							if ast.IsExported(name.Name) {
								constants = append(constants, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if !ast.IsExported(value.Name.Name) {
					continue
				}
				if value.Recv == nil {
					funcs = append(funcs, value.Name.Name)
					continue
				}
				receiver := receiverName(value.Recv.List[0].Type)
				methods = append(methods, receiver+"."+value.Name.Name)
			}
		}
	}
	for path := range importSet {
		imports = append(imports, path)
	}
	sort.Strings(types)
	sort.Strings(funcs)
	sort.Strings(methods)
	sort.Strings(constants)
	sort.Strings(imports)
	return types, funcs, methods, constants, imports
}

func receiverName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverName(value.X)
	default:
		return ""
	}
}

func assertExactStrings(t *testing.T, name string, got []string, want []string) {
	t.Helper()
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s mismatch\ngot:  %q\nwant: %q", name, got, want)
	}
}
