package transition_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const transitionImportPath = "github.com/wsnacj/agentx-go/runtime/workflow/transition"

func TestOwnerImportDirection(t *testing.T) {
	for _, path := range productionGoFiles(t, ".") {
		for _, imported := range importsOf(t, path) {
			if imported == "github.com/wsnacj/agentx-go/runtime/workflow" {
				continue
			}
			if strings.HasPrefix(imported, "github.com/wsnacj/agentx-go/") ||
				strings.HasPrefix(imported, "hs/") ||
				strings.HasPrefix(imported, "scene/") {
				t.Errorf("%s imports non-approved owner %q", path, imported)
			}
		}
	}

	for _, path := range productionGoFiles(t, "..") {
		for _, imported := range importsOf(t, path) {
			if imported == transitionImportPath {
				t.Errorf("parent workflow owner reverse-imports %q in %s", imported, path)
			}
		}
	}
}

func productionGoFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", root, err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		paths = append(paths, filepath.Join(root, entry.Name()))
	}
	return paths
}

func importsOf(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}
	imports := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("Unquote import in %s: %v", path, err)
		}
		imports = append(imports, imported)
	}
	return imports
}
