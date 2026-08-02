package tools

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestBrowserToolsDoNotDependOnProductRuntime(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	root := filepath.Dir(file)
	forbiddenImports := []string{"hs/", "/scene/", "/core/agentx/", "/engine"}
	forbiddenSource := []string{"engine.Runner", "core/agentx/browserdaemon"}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range parsed.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			for _, forbidden := range forbiddenImports {
				if strings.Contains(value, forbidden) {
					t.Fatalf("browser tools boundary violation: import %q in %s", value, path)
				}
			}
		}
		blob, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, forbidden := range forbiddenSource {
			if strings.Contains(string(blob), forbidden) {
				t.Fatalf("browser tools boundary violation: found %q in %s", forbidden, path)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
