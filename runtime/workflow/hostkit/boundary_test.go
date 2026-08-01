package hostkit

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProductionImportsStayInsideCanonicalWorkflow(t *testing.T) {
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	allowed := map[string]bool{
		"github.com/wsnacj/agentx-go/runtime/workflow":               true,
		"github.com/wsnacj/agentx-go/runtime/workflow/composition":   true,
		"github.com/wsnacj/agentx-go/runtime/workflow/journal":       true,
		"github.com/wsnacj/agentx-go/runtime/workflow/lowering":      true,
		"github.com/wsnacj/agentx-go/runtime/workflow/nodeexec":      true,
		"github.com/wsnacj/agentx-go/runtime/workflow/orchestration": true,
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", entry, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquote import: %v", err)
			}
			if strings.HasPrefix(path, "github.com/wsnacj/agentx-go/") && !allowed[path] {
				t.Errorf("%s imports non-approved owner %q", entry, path)
			}
			if path == "hs" || strings.HasPrefix(path, "hs/") || strings.HasPrefix(path, "scene/") {
				t.Errorf("%s imports forbidden substrate %q", entry, path)
			}
		}
	}
}

func TestChineseReferenceExists(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatalf("ReadFile(API.md): %v", err)
	}
	for _, required := range []string{
		"Developer Preview candidate",
		"Config",
		"New",
		"Run",
		"并发",
		"取消",
		"durable",
		"非目标",
	} {
		if !strings.Contains(string(content), required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
