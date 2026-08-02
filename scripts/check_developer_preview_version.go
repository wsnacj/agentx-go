//go:build ignore

// check_developer_preview_version verifies the focused version facts used by
// the Developer Preview consumer and Chinese adoption documentation.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const versionPath = "docs/reference/developer-preview-version.txt"

var modules = []string{
	"github.com/wsnacj/agentx-go",
	"github.com/wsnacj/agentx-go/components",
	"github.com/wsnacj/agentx-go/runtime",
	"github.com/wsnacj/agentx-go/extensions",
}

func main() {
	root, err := os.Getwd()
	check(err)
	versionBytes, err := os.ReadFile(filepath.Join(root, versionPath))
	check(err)
	version := strings.TrimSpace(string(versionBytes))
	if !strings.HasPrefix(version, "v0.0.0-") || strings.ContainsAny(version, " \t\r\n") {
		check(fmt.Errorf("invalid Developer Preview pseudo-version %q", version))
	}

	selected, err := readRequiredModules(filepath.Join(root, "conformance/consumer/go.mod"))
	check(err)
	for _, module := range modules {
		if got := selected[module]; got != version {
			check(fmt.Errorf("conformance consumer %s version = %q, want %q", module, got, version))
		}
	}

	for _, relative := range []string{
		"README.md",
		"CHANGELOG.md",
		"docs/guides/installation-and-modules.md",
		"docs/guides/versioning-and-upgrades.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, relative))
		check(err)
		if !strings.Contains(string(content), version) {
			check(fmt.Errorf("%s does not reference accepted version %s", relative, version))
		}
	}

	fmt.Printf("agentx-developer-preview-version-ok:version=%s:modules=%d\n", version, len(modules))
}

func readRequiredModules(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	selected := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		if fields[0] == "require" {
			fields = fields[1:]
		}
		if len(fields) < 2 {
			continue
		}
		for _, module := range modules {
			if fields[0] == module && strings.HasPrefix(fields[1], "v") {
				selected[module] = fields[1]
			}
		}
	}
	return selected, scanner.Err()
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Developer Preview version gate:", err)
	os.Exit(1)
}
