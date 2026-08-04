package hostkit

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

func TestExactExportSurfaceAndChineseReference(t *testing.T) {
	types, funcs, methods := collectContract(t)
	assertStrings(t, "types", types, []string{
		"ChatClientConfig",
		"Config",
		"Factory",
		"ModelResult",
		"ModelToolRoundAdapter",
		"ModelToolClientConfig",
		"ModelToolRoundConfig",
		"ModelToolRoundExchange",
		"ModelToolRoundResult",
		"RunConfig",
		"RunResult",
		"ToolResult",
		"ToolDirectAnswer",
	})
	assertStrings(t, "functions", funcs, []string{"Execute", "New", "NewChatClient", "NewModelToolClient", "NewModelToolRoundAdapter"})
	assertStrings(t, "methods", methods, []string{"Execute", "ExecuteRound", "ExecutionResult"})

	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatalf("read API.md: %v", err)
	}
	text := string(content)
	for _, required := range []string{
		"v0.1.0 Developer Preview candidate",
		"Config",
		"Factory",
		"RunConfig",
		"RunResult",
		"Execute",
		"New",
		"ModelToolRoundAdapter",
		"ModelToolRoundConfig",
		"NewModelToolRoundAdapter",
		"ModelToolClientConfig",
		"NewModelToolClient",
		"ChatClientConfig",
		"NewChatClient",
		"ToolDirectAnswer",
		"DirectAnswer",
		"ExecutionResult",
		"ExecuteRound",
		"RequestModel",
		"ExecuteTools",
		"明确 non-goal",
		"Shutdown",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}

func TestProductionImportsStaySubstrateNeutral(t *testing.T) {
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", entry, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquote import: %v", err)
			}
			if path == "hs" || strings.HasPrefix(path, "hs/") || strings.Contains(path, "/scene/") || strings.Contains(path, "/engine") {
				t.Fatalf("forbidden production import %q", path)
			}
		}
	}
}

func collectContract(t *testing.T) (types, funcs, methods []string) {
	t.Helper()
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry, err)
		}
		for _, decl := range file.Decls {
			switch value := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					typed, ok := spec.(*ast.TypeSpec)
					if ok && ast.IsExported(typed.Name.Name) {
						types = append(types, typed.Name.Name)
					}
				}
			case *ast.FuncDecl:
				if !ast.IsExported(value.Name.Name) {
					continue
				}
				if value.Recv == nil {
					funcs = append(funcs, value.Name.Name)
				} else if exportedReceiver(value.Recv.List[0].Type) {
					methods = append(methods, value.Name.Name)
				}
			}
		}
	}
	sort.Strings(types)
	sort.Strings(funcs)
	sort.Strings(methods)
	return types, funcs, methods
}

func exportedReceiver(expr ast.Expr) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		return ast.IsExported(value.Name)
	case *ast.StarExpr:
		return exportedReceiver(value.X)
	default:
		return false
	}
}

func assertStrings(t *testing.T, name string, got, want []string) {
	t.Helper()
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s mismatch\ngot: %q\nwant: %q", name, got, want)
	}
}
