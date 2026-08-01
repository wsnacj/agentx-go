package runstore_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const nodeexecImportPath = "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"

func TestOwnerImportDirection(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Clean(entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("Unquote import in %s: %v", path, err)
			}
			if imported == nodeexecImportPath {
				continue
			}
			firstSegment := strings.Split(imported, "/")[0]
			if strings.Contains(firstSegment, ".") ||
				strings.HasPrefix(imported, "hs/") ||
				strings.HasPrefix(imported, "scene/") {
				t.Errorf("%s imports non-approved owner %q", path, imported)
			}
		}
	}
}
