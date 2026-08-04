//go:build ignore

// check_developer_preview_version verifies the nine-module fixed-version facts
// used by representative external consumers. Public adoption documentation
// intentionally does not embed these fast-moving validation versions.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	rootVersionPath = "docs/reference/developer-preview-version.txt"
	matrixPath      = "docs/reference/developer-preview-module-versions.txt"
)

type moduleSpec struct {
	path     string
	consumer string
}

var modules = []moduleSpec{
	{path: "github.com/wsnacj/agentx-go", consumer: "conformance/consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/components", consumer: "conformance/consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/runtime", consumer: "conformance/consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/extensions", consumer: "conformance/consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/providers", consumer: "providers/conformance/provider-cohort-consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/tools", consumer: "browser/conformance/browser-platform-consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/browser", consumer: "browser/conformance/browser-platform-consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/document", consumer: "document/conformance/tools-consumer/go.mod"},
	{path: "github.com/wsnacj/agentx-go/scenes", consumer: "conformance/consumer/go.mod"},
}

func main() {
	root, err := os.Getwd()
	check(err)
	expected, err := readVersionMatrix(filepath.Join(root, matrixPath))
	check(err)
	if len(expected) != len(modules) {
		check(fmt.Errorf("version matrix has %d modules, want %d", len(expected), len(modules)))
	}

	rootVersionBytes, err := os.ReadFile(filepath.Join(root, rootVersionPath))
	check(err)
	rootVersion := strings.TrimSpace(string(rootVersionBytes))
	if rootVersion != expected["github.com/wsnacj/agentx-go"] {
		check(fmt.Errorf("legacy root version = %q, matrix has %q", rootVersion, expected["github.com/wsnacj/agentx-go"]))
	}

	for _, module := range modules {
		version, ok := expected[module.path]
		if !ok {
			check(fmt.Errorf("version matrix is missing %s", module.path))
		}
		selected, err := readRequiredModule(filepath.Join(root, module.consumer), module.path)
		check(err)
		if selected != version {
			check(fmt.Errorf("%s selects %s at %q, want %q", module.consumer, module.path, selected, version))
		}
	}

	fmt.Printf("agentx-developer-preview-version-ok:root=%s:modules=%d\n", rootVersion, len(modules))
}

func readVersionMatrix(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	versions := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != "v0.1.0" || strings.ContainsAny(fields[1], "\t\r\n ") {
			return nil, fmt.Errorf("invalid version matrix line %q", line)
		}
		if _, exists := versions[fields[0]]; exists {
			return nil, fmt.Errorf("duplicate version matrix module %s", fields[0])
		}
		versions[fields[0]] = fields[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func readRequiredModule(path, module string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	selected := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[0] == "require" {
			fields = fields[1:]
		}
		if len(fields) >= 2 && fields[0] == module && strings.HasPrefix(fields[1], "v") {
			selected = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return selected, nil
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Developer Preview version gate:", err)
	os.Exit(1)
}
