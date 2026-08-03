//go:build ignore

// check_package_api_docs verifies that every externally importable production
// package in the nine AgentX library modules has one non-empty Chinese API.md.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
			if err != nil || len(bytes.TrimSpace(content)) == 0 {
				check(fmt.Errorf("%s: missing non-empty Chinese API reference %s", filepath.ToSlash(relative), filepath.ToSlash(reference)))
			}
			if !containsHan(content) {
				check(fmt.Errorf("%s: API reference has no Chinese text", filepath.ToSlash(relative)))
			}
			packages++
		}
	}

	fmt.Printf("agentx-package-api-docs-ok:modules=%d:packages=%d\n", len(modules), packages)
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
