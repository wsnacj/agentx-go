package execution_test

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

func TestExactExperimentalSurface(t *testing.T) {
	exports := map[string]bool{}
	for _, path := range productionGoFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			switch value := declaration.(type) {
			case *ast.FuncDecl:
				if value.Recv == nil && ast.IsExported(value.Name.Name) {
					exports[value.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(spec.Name.Name) {
							exports[spec.Name.Name] = true
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								exports[name.Name] = true
							}
						}
					}
				}
			}
		}
	}
	got := sortedKeys(exports)
	want := []string{"Host", "New", "Request", "Result", "Runtime"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("exports = %v, want %v", got, want)
	}
}

func TestOwnerImportDirection(t *testing.T) {
	for _, path := range productionGoFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		for _, spec := range parsed.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("Unquote(%s): %v", spec.Path.Value, err)
			}
			if strings.HasPrefix(imported, "hs/") ||
				strings.HasPrefix(imported, "scene/") ||
				(strings.HasPrefix(imported, "github.com/") && imported != "github.com/wsnacj/agentx-go") {
				t.Errorf("%s imports forbidden owner %q", path, imported)
			}
		}
	}
}

func productionGoFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		files = append(files, filepath.Clean(entry.Name()))
	}
	sort.Strings(files)
	return files
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
