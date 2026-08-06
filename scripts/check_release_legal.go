//go:build ignore

// check_release_legal verifies the Apache-2.0 distribution files shared by
// every AgentX Go module. It does not replace dependency or legal review.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var moduleRoots = []string{
	".",
	"components",
	"runtime",
	"extensions",
	"providers",
	"tools",
	"browser",
	"document",
	"scenes",
	"examples",
}

var directDependencyNotices = []string{
	"github.com/fsnotify/fsnotify",
	"gopkg.in/yaml.v3",
	"codeberg.org/readeck/go-readability/v2",
	"github.com/PuerkitoBio/goquery",
	"github.com/sergi/go-diff",
	"golang.org/x/net",
	"github.com/agext/levenshtein",
	"github.com/cenkalti/backoff/v5",
	"github.com/prometheus/client_golang",
	"github.com/stretchr/testify",
	"go.uber.org/zap",
	"playwright",
	"vitepress",
}

func main() {
	root, err := os.Getwd()
	check(err)

	license := mustRead(filepath.Join(root, "LICENSE"))
	for _, marker := range []string{
		"Apache License",
		"Version 2.0, January 2004",
		"2. Grant of Copyright License.",
		"3. Grant of Patent License.",
		"END OF TERMS AND CONDITIONS",
	} {
		if !bytes.Contains(license, []byte(marker)) {
			check(fmt.Errorf("LICENSE is missing Apache-2.0 marker %q", marker))
		}
	}

	notice := mustRead(filepath.Join(root, "NOTICE"))
	for _, marker := range []string{
		"AgentX Go",
		"Copyright 2026 wsnacj and AgentX Go contributors",
	} {
		if !bytes.Contains(notice, []byte(marker)) {
			check(fmt.Errorf("NOTICE is missing %q", marker))
		}
	}

	for _, moduleRoot := range moduleRoots {
		moduleLicense := mustRead(filepath.Join(root, moduleRoot, "LICENSE"))
		if !bytes.Equal(moduleLicense, license) {
			check(fmt.Errorf("%s/LICENSE differs from root LICENSE", moduleRoot))
		}
		moduleNotice := mustRead(filepath.Join(root, moduleRoot, "NOTICE"))
		if !bytes.Equal(moduleNotice, notice) {
			check(fmt.Errorf("%s/NOTICE differs from root NOTICE", moduleRoot))
		}
	}

	thirdParty := string(mustRead(filepath.Join(root, "THIRD_PARTY_NOTICES.md")))
	for _, dependency := range directDependencyNotices {
		if !strings.Contains(thirdParty, dependency) {
			check(fmt.Errorf("THIRD_PARTY_NOTICES.md is missing %s", dependency))
		}
	}

	packageJSON := struct {
		License string `json:"license"`
	}{}
	check(json.Unmarshal(mustRead(filepath.Join(root, "package.json")), &packageJSON))
	if packageJSON.License != "Apache-2.0" {
		check(fmt.Errorf("package.json license = %q, want Apache-2.0", packageJSON.License))
	}

	fmt.Printf("agentx-release-legal-ok:license=Apache-2.0:module_roots=%d:release_modules=9:examples_release=false:third_party_direct=%d\n", len(moduleRoots), len(directDependencyNotices))
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
	fmt.Fprintln(os.Stderr, "release legal gate:", err)
	os.Exit(1)
}
