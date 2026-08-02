package browserops

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductionSourcesDoNotDependOnHSSceneOrRunner(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve browserops source path")
	}
	root := filepath.Dir(filename)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imported := range parsed.Imports {
			importPath := strings.Trim(imported.Path.Value, `"`)
			if importPath == "hs" || strings.HasPrefix(importPath, "hs/") ||
				importPath == "scene" || strings.HasPrefix(importPath, "scene/") {
				t.Errorf("%s imports forbidden owner %q", path, importPath)
			}
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			if identifier, ok := node.(*ast.Ident); ok && identifier.Name == "Runner" {
				t.Errorf("%s references forbidden Runner identifier", path)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan browserops production boundary: %v", err)
	}
}
