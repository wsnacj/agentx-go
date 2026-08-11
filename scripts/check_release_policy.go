//go:build ignore

// check_release_policy verifies the current Developer Preview candidate
// compatibility and named responsibility decisions before artifacts or tags.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var releaseVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)

var releaseModuleDirs = []string{
	".",
	"components",
	"runtime",
	"extensions",
	"providers",
	"tools",
	"browser",
	"document",
	"scenes",
}

var compatibilityPackages = map[string]bool{
	".":                         true,
	"components/llm":            true,
	"runtime/execution":         true,
	"runtime/hostkit":           true,
	"runtime/objective/hostkit": true,
	"runtime/session/hostkit":   true,
	"runtime/toolloop":          true,
	"runtime/workflow":          true,
	"runtime/workflow/hostkit":  true,
}

func main() {
	root, err := os.Getwd()
	check(err)
	releaseVersion := strings.TrimSpace(string(mustRead(filepath.Join(root, "docs/reference/developer-preview-version.txt"))))
	if !releaseVersionPattern.MatchString(releaseVersion) {
		check(fmt.Errorf("invalid Developer Preview release version %q", releaseVersion))
	}

	checkVersionMatrix(filepath.Join(root, "docs/reference/developer-preview-module-versions.txt"), releaseVersion)
	checkCompatibility(filepath.Join(root, "docs/reference/developer-preview-packages.tsv"))
	checkReleaseModules(root, releaseVersion)
	checkMarkers(filepath.Join(root, "SECURITY.md"), []string{
		"@wsnacj",
		"3个工作日",
		"不是修复SLA",
		"暂无backup security owner",
	})
	checkMarkers(filepath.Join(root, "SUPPORT.md"), []string{
		"@wsnacj",
		"release approver",
		"rollback owner",
		"暂无backup owner",
	})
	checkMarkers(filepath.Join(root, "CHANGELOG.md"), []string{
		"## [0.2.2] - 2026-08-11",
		"root module checksum",
		"## [0.2.0] - 2026-08-06",
		"首版核心兼容候选面收窄",
	})

	fmt.Printf("agentx-release-policy-ok:version=%s:modules=%d:compatibility_packages=%d:release_owner=@wsnacj:public_beta_ready=false\n",
		releaseVersion, len(releaseModuleDirs), len(compatibilityPackages))
}

func checkVersionMatrix(path, releaseVersion string) {
	seen := 0
	scanner := bufio.NewScanner(bytes.NewReader(mustRead(path)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != releaseVersion {
			check(fmt.Errorf("%s has invalid release row %q", path, line))
		}
		seen++
	}
	check(scanner.Err())
	if seen != len(releaseModuleDirs) {
		check(fmt.Errorf("%s has %d modules, want %d", path, seen, len(releaseModuleDirs)))
	}
}

func checkCompatibility(path string) {
	actual := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(mustRead(path)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			check(fmt.Errorf("%s has malformed row %q", path, line))
		}
		if fields[1] == "developer_preview_candidate" {
			actual[fields[0]] = true
		}
	}
	check(scanner.Err())
	if len(actual) != len(compatibilityPackages) {
		check(fmt.Errorf("compatibility package count = %d, want %d", len(actual), len(compatibilityPackages)))
	}
	for packageDir := range compatibilityPackages {
		if !actual[packageDir] {
			check(fmt.Errorf("compatibility package %s is missing", packageDir))
		}
	}
}

func checkReleaseModules(root, releaseVersion string) {
	for _, moduleDir := range releaseModuleDirs {
		path := filepath.Join(root, moduleDir, "go.mod")
		content := string(mustRead(path))
		if strings.Contains(content, "replace ") || strings.Contains(content, "replace(") {
			check(fmt.Errorf("%s contains replace", path))
		}
		if !strings.Contains(content, "\ngo 1.25.0\n") {
			check(fmt.Errorf("%s must declare go 1.25.0", path))
		}
		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 3 && fields[0] == "require" {
				fields = fields[1:]
			}
			if len(fields) < 2 || !strings.HasPrefix(fields[0], "github.com/wsnacj/agentx-go") {
				continue
			}
			if fields[1] != releaseVersion {
				check(fmt.Errorf("%s selects %s at %s, want %s", path, fields[0], fields[1], releaseVersion))
			}
		}
		check(scanner.Err())
	}
}

func checkMarkers(path string, markers []string) {
	content := string(mustRead(path))
	for _, marker := range markers {
		if !strings.Contains(content, marker) {
			check(fmt.Errorf("%s is missing %q", path, marker))
		}
	}
}

func mustRead(path string) []byte {
	content, err := os.ReadFile(path)
	check(err)
	if len(bytes.TrimSpace(content)) == 0 {
		check(fmt.Errorf("%s is empty", path))
	}
	return content
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "release policy gate:", err)
	os.Exit(1)
}
