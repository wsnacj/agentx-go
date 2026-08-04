//go:build ignore

// check_package_api_docs verifies that every externally importable production
// package in the nine AgentX library modules has one non-empty Chinese API.md.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func main() {
	root, err := os.Getwd()
	check(err)

	modules := []string{".", "components", "runtime", "extensions", "providers", "tools", "browser", "document", "scenes"}
	packages := 0
	for _, module := range modules {
		directory := filepath.Join(root, module)
		command := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")
		command.Dir = directory
		command.Env = append(os.Environ(), "GOWORK=off")
		output, err := command.CombinedOutput()
		if err != nil {
			check(fmt.Errorf("go list %s: %w: %s", module, err, strings.TrimSpace(string(output))))
		}
		for _, packageDir := range strings.Fields(string(output)) {
			relative, err := filepath.Rel(root, packageDir)
			check(err)
			if hasInternalSegment(relative) {
				continue
			}
			reference := filepath.Join(packageDir, "API.md")
			if module == "." && filepath.Clean(packageDir) == filepath.Clean(root) {
				reference = filepath.Join(root, "docs", "reference", "agentx.md")
			}
			content, err := os.ReadFile(reference)
			if err != nil || len(bytes.TrimSpace(content)) < 128 {
				check(fmt.Errorf("%s: missing non-empty Chinese API reference %s", filepath.ToSlash(relative), filepath.ToSlash(reference)))
			}
			if !containsHan(content) {
				check(fmt.Errorf("%s: API reference has no Chinese text", filepath.ToSlash(relative)))
			}
			if !bytes.HasPrefix(bytes.TrimSpace(content), []byte("# ")) {
				check(fmt.Errorf("%s: API reference must start with a level-one heading", filepath.ToSlash(relative)))
			}
			exports, err := exportedIdentifiers(packageDir)
			check(err)
			if len(exports) > 0 && !mentionsAnyIdentifier(string(content), exports) {
				check(fmt.Errorf("%s: API reference does not mention any of %d exported identifiers", filepath.ToSlash(relative), len(exports)))
			}
			packages++
		}
	}

	fmt.Printf("agentx-package-api-docs-ok:modules=%d:packages=%d\n", len(modules), packages)
}

func exportedIdentifiers(directory string) ([]string, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, parsed := range packages {
		for _, file := range parsed.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(value.Name.Name) {
						names[value.Name.Name] = true
					}
				case *ast.ValueSpec:
					for _, name := range value.Names {
						if ast.IsExported(name.Name) {
							names[name.Name] = true
						}
					}
				case *ast.FuncDecl:
					if ast.IsExported(value.Name.Name) {
						names[value.Name.Name] = true
					}
				}
				return true
			})
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func mentionsAnyIdentifier(content string, identifiers []string) bool {
	for _, identifier := range identifiers {
		if strings.Contains(content, identifier) {
			return true
		}
	}
	return false
}

func hasInternalSegment(path string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(path), "/") {
		if segment == "internal" {
			return true
		}
	}
	return false
}

func containsHan(content []byte) bool {
	for _, value := range string(content) {
		if unicode.Is(unicode.Han, value) {
			return true
		}
	}
	return false
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "package API docs gate:", err)
	os.Exit(1)
}
