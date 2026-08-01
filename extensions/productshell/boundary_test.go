package productshell

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestImportBoundary(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve package directory")
	}
	dir := filepath.Dir(current)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, declaration := range file.Imports {
			importPath, err := strconv.Unquote(declaration.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(importPath, "hs/") || strings.Contains(importPath, "/scene/") || strings.Contains(importPath, "/engine") {
				t.Fatalf("%s imports forbidden host/runtime owner %q", entry.Name(), importPath)
			}
		}
	}
}

func TestPortableObservationOwnerExcludesHostProjectionAndGovernance(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve package directory")
	}
	dir := filepath.Dir(current)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := [][]byte{
		[]byte("func BuildTerminalObservationSnapshot"),
		[]byte("func parseTerminalObservationInput"),
		[]byte("func parseBrowserOperationObservationInput"),
		[]byte("type ObservationSnapshot struct"),
		[]byte("type HostUIHandoffConsumerInventoryReport struct"),
		[]byte("type HostUIHandoffConsumerOnboardingReport struct"),
		[]byte("type HostUIHandoffReadbackParityReport struct"),
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, needle := range forbidden {
			if bytes.Contains(source, needle) {
				t.Fatalf("%s contains forbidden host projection/governance owner %q", entry.Name(), needle)
			}
		}
	}
}
