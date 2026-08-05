package catalog_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCatalogImportDirection(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Clean(entry.Name())
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		for _, spec := range parsed.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(imported, "hs/") || strings.HasPrefix(imported, "scene/") ||
				strings.Contains(imported, "agentx-platform") || strings.Contains(imported, "/engine") ||
				strings.Contains(imported, "provider") || strings.Contains(imported, "sqlite") {
				t.Errorf("%s imports forbidden owner %q", path, imported)
			}
		}
	}
}
