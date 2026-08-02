package astock_test

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	astock "github.com/wsnacj/agentx-go/scenes/astock"
)

type embeddedToolContract struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// TestToolDefinitionsHaveEmbeddedToolContracts keeps the two existing A-share
// surfaces paired without pretending they are the same contract. The Go
// definitions are the runtime LLM schema; the declarative files additionally
// carry install/catalog metadata and historically have different parameter
// wording and fields. M5D preserves both byte/signature surfaces and checks
// their differential against HS during cutover instead of silently rewriting
// either one.
func TestToolDefinitionsHaveEmbeddedToolContracts(t *testing.T) {
	assetPaths, err := fs.Glob(astock.ExtensionFS(), "tools/*.tool.json")
	if err != nil {
		t.Fatalf("glob embedded tool contracts: %v", err)
	}
	sort.Strings(assetPaths)

	assetsByName := make(map[string]embeddedToolContract, len(assetPaths))
	for _, path := range assetPaths {
		content, err := fs.ReadFile(astock.ExtensionFS(), path)
		if err != nil {
			t.Fatalf("read embedded tool contract %s: %v", path, err)
		}
		var contract embeddedToolContract
		if err := json.Unmarshal(content, &contract); err != nil {
			t.Fatalf("decode embedded tool contract %s: %v", path, err)
		}
		if strings.TrimSpace(contract.Name) == "" {
			t.Fatalf("embedded tool contract %s has no name", path)
		}
		if _, exists := assetsByName[contract.Name]; exists {
			t.Fatalf("duplicate embedded tool contract name %q", contract.Name)
		}
		assetsByName[contract.Name] = contract
	}

	definitions := astock.ToolDefinitions()
	if len(definitions) != len(assetsByName) {
		t.Fatalf("tool contract count mismatch: definitions=%d assets=%d", len(definitions), len(assetsByName))
	}
	seen := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		assertToolDefinitionHasAsset(t, definition, assetsByName, seen)
	}
	for name := range assetsByName {
		if !seen[name] {
			t.Errorf("embedded tool contract %q has no ToolDefinitions entry", name)
		}
	}
}

func assertToolDefinitionHasAsset(
	t *testing.T,
	definition llm.Tool,
	assetsByName map[string]embeddedToolContract,
	seen map[string]bool,
) {
	t.Helper()
	name := strings.TrimSpace(definition.Function.Name)
	if name == "" {
		t.Fatal("ToolDefinitions returned a tool with an empty function name")
	}
	if seen[name] {
		t.Fatalf("ToolDefinitions returned duplicate tool name %q", name)
	}
	seen[name] = true

	asset, ok := assetsByName[name]
	if !ok {
		t.Fatalf("ToolDefinitions returned %q without an embedded tool contract", name)
	}
	if strings.TrimSpace(definition.Function.Description) == "" {
		t.Errorf("ToolDefinitions entry %q has no description", name)
	}
	if len(definition.Function.Parameters) == 0 {
		t.Errorf("ToolDefinitions entry %q has no parameters", name)
	}
	if strings.TrimSpace(asset.Description) == "" {
		t.Errorf("embedded tool contract %q has no description", name)
	}
	if len(asset.Parameters) == 0 {
		t.Errorf("embedded tool contract %q has no parameters", name)
	}
}

func TestManifestAndNameListsReturnFreshCopies(t *testing.T) {
	wantTools := astock.ToolNames()
	wantSkills := astock.SkillNames()
	wantManifest := astock.Manifest()
	wantDefinitions := astock.ToolDefinitions()

	mutatedTools := astock.ToolNames()
	mutatedSkills := astock.SkillNames()
	mutatedManifest := astock.Manifest()
	mutateFirst(mutatedTools)
	mutateFirst(mutatedSkills)
	mutateFirst(mutatedManifest.Skills)
	mutateFirst(mutatedManifest.Tools)
	mutateFirst(mutatedManifest.Packs)
	mutateFirst(mutatedManifest.Workflows)
	mutatedDefinitions := astock.ToolDefinitions()
	mutatedDefinitions[0].Function.Name = "caller-mutation"
	mutatedDefinitions[0].Function.Parameters["caller-mutation"] = true

	if got := astock.ToolNames(); !reflect.DeepEqual(got, wantTools) {
		t.Errorf("ToolNames shared caller mutation: got=%v want=%v", got, wantTools)
	}
	if got := astock.SkillNames(); !reflect.DeepEqual(got, wantSkills) {
		t.Errorf("SkillNames shared caller mutation: got=%v want=%v", got, wantSkills)
	}
	if got := astock.Manifest(); !reflect.DeepEqual(got, wantManifest) {
		t.Errorf("Manifest shared caller mutation: got=%#v want=%#v", got, wantManifest)
	}
	if got := astock.ToolDefinitions(); !reflect.DeepEqual(got, wantDefinitions) {
		t.Errorf("ToolDefinitions shared caller mutation: got=%#v want=%#v", got, wantDefinitions)
	}
}

func mutateFirst(values []string) {
	if len(values) > 0 {
		values[0] = "caller-mutation"
	}
}

func TestProductionSourcesDoNotDependOnHSSceneOrRunner(t *testing.T) {
	root := astockSourceRoot(t)
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
			if importPath == "hs" || strings.HasPrefix(importPath, "hs/") {
				t.Errorf("%s imports forbidden HS package %q", relativeAStockPath(root, path), importPath)
			}
			if importPath == "scene" || strings.HasPrefix(importPath, "scene/") {
				t.Errorf("%s imports forbidden Scene package %q", relativeAStockPath(root, path), importPath)
			}
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok && identifier.Name == "Runner" {
				t.Errorf("%s references forbidden Runner identifier", relativeAStockPath(root, path))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan astock production boundary: %v", err)
	}
}

func astockSourceRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve astock test source path")
	}
	return filepath.Dir(filename)
}

func relativeAStockPath(root string, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}
