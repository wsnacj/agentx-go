//go:build ignore

// check_examples_version keeps the external-style examples module and its
// README on one fixed source revision. It does not update versions.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pseudoVersionPattern = regexp.MustCompile(`v0\.0\.0-[0-9]{14}-[0-9a-f]{12}`)

func main() {
	root, err := os.Getwd()
	fail(err)
	goModPath := filepath.Join(root, "examples", "go.mod")
	readmePath := filepath.Join(root, "examples", "README.md")

	goMod, err := os.ReadFile(goModPath)
	fail(err)
	if strings.Contains(string(goMod), "replace ") || strings.Contains(string(goMod), "replace(") {
		fail(fmt.Errorf("examples/go.mod must not contain replace"))
	}
	selected := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(goMod)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "github.com/wsnacj/agentx-go") || !pseudoVersionPattern.MatchString(fields[1]) {
			continue
		}
		selected[fields[0]] = fields[1]
	}
	fail(scanner.Err())
	if len(selected) == 0 {
		fail(fmt.Errorf("examples/go.mod has no fixed agentx-go modules"))
	}
	version := ""
	for module, current := range selected {
		if version == "" {
			version = current
		}
		if current != version {
			fail(fmt.Errorf("examples module version drift: %s=%s want %s", module, current, version))
		}
	}

	readme, err := os.ReadFile(readmePath)
	fail(err)
	matches := pseudoVersionPattern.FindAllString(string(readme), -1)
	if len(matches) != 1 || matches[0] != version {
		fail(fmt.Errorf("examples README version = %v, want exactly [%s]", matches, version))
	}
	fmt.Printf("agentx-examples-version-ok:version=%s:modules=%d\n", version, len(selected))
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
