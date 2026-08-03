package retrieval

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestKnownRetrievalControlFlowUsesTypedErrors(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve retrieval typed error guard path")
	}
	root := filepath.Dir(current)
	targets := map[string][]string{
		"web_fetch_execute.go": {"formatWebFetchRequestError"},
		"search_engine.go":     {"formatSearchRequestError"},
	}
	for name, functions := range targets {
		path := filepath.Join(root, name)
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, path, payload, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, function := range functions {
			body := retrievalGuardFunctionBody(t, fileSet, parsed, payload, function, name)
			for _, forbidden := range []string{"err.Error()", "strings.Contains("} {
				if strings.Contains(body, forbidden) {
					t.Errorf("%s %s retains text-driven control token %q", name, function, forbidden)
				}
			}
		}
	}
}

func retrievalGuardFunctionBody(t *testing.T, fileSet *token.FileSet, file *ast.File, payload []byte, name, path string) string {
	t.Helper()
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != name || function.Body == nil {
			continue
		}
		start := fileSet.Position(function.Body.Pos()).Offset
		end := fileSet.Position(function.Body.End()).Offset
		if start >= 0 && end >= start && end <= len(payload) {
			return string(payload[start:end])
		}
	}
	t.Fatalf("function %s not found in %s", name, path)
	return ""
}
