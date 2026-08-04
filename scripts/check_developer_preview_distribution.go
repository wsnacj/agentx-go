//go:build ignore

// check_developer_preview_distribution composes the focused Developer Preview
// API, documentation, artifact and clean-room consumer gates.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type moduleSpec struct {
	path     string
	dir      string
	required string
}

type downloadInfo struct {
	Path    string
	Version string
	Zip     string
	Sum     string
	Origin  struct {
		Hash string
	}
}

var releaseModules = []moduleSpec{
	{path: "github.com/wsnacj/agentx-go", dir: ".", required: "README.md"},
	{path: "github.com/wsnacj/agentx-go/components", dir: "components", required: "llm/API.md"},
	{path: "github.com/wsnacj/agentx-go/runtime", dir: "runtime", required: "workflow/hostkit/API.md"},
	{path: "github.com/wsnacj/agentx-go/extensions", dir: "extensions", required: "domainkit/API.md"},
	{path: "github.com/wsnacj/agentx-go/providers", dir: "providers", required: "API.md"},
	{path: "github.com/wsnacj/agentx-go/tools", dir: "tools", required: "API.md"},
	{path: "github.com/wsnacj/agentx-go/browser", dir: "browser", required: "API.md"},
	{path: "github.com/wsnacj/agentx-go/document", dir: "document", required: "API.md"},
	{path: "github.com/wsnacj/agentx-go/scenes", dir: "scenes", required: "astock/API.md"},
}

func main() {
	freshCache := flag.Bool("fresh-cache", false, "fetch the consumer modules from VCS into an empty temporary cache")
	readOnlyCache := flag.Bool("read-only-cache", false, "freeze the fresh consumer module cache before verify, test and run")
	full := flag.Bool("full", false, "also run test, vet, tidy and list for all nine source modules")
	portal := flag.Bool("portal", false, "also build and validate the optional local Developer Portal (requires npm ci first)")
	flag.Parse()

	root, err := os.Getwd()
	check(err)
	checkRequiredFiles(root)
	checkGoMods(root)
	versions := readVersionMatrix(filepath.Join(root, "docs/reference/developer-preview-module-versions.txt"))
	rootVersion := versions["github.com/wsnacj/agentx-go"]
	if rootVersion == "" || len(versions) != len(releaseModules) {
		check(fmt.Errorf("Developer Preview version matrix has %d modules and root %q, want %d modules", len(versions), rootVersion, len(releaseModules)))
	}

	env := append(os.Environ(), "GOWORK=off")
	printRun(root, env, "go", "run", "./scripts/check_developer_preview_version.go")
	printRun(root, env, "go", "run", "./scripts/check_examples_version.go")
	printRun(root, env, "go", "run", "./scripts/check_developer_preview_api.go", "-check-platforms")
	printRun(root, env, "go", "run", "./scripts/check_docs_links.go")
	cleanroomArgs := []string{"run", "./scripts/check_cleanroom_consumer.go"}
	if *freshCache {
		cleanroomArgs = append(cleanroomArgs, "-fresh-cache")
	}
	if *readOnlyCache {
		if !*freshCache {
			check(fmt.Errorf("read-only-cache requires fresh-cache"))
		}
		cleanroomArgs = append(cleanroomArgs, "-read-only-cache")
	}
	printRun(root, env, "go", cleanroomArgs...)

	for _, module := range releaseModules {
		version := versions[module.path]
		if version == "" {
			check(fmt.Errorf("Developer Preview version matrix is missing %s", module.path))
		}
		checkArtifact(root, env, module, version, pseudoVersionRevision(version))
	}
	if *full {
		for _, module := range releaseModules {
			dir := filepath.Join(root, module.dir)
			printRun(dir, env, "go", "test", "./...")
			printRun(dir, env, "go", "vet", "./...")
			printRun(dir, env, "go", "mod", "tidy", "-diff")
			printRun(dir, env, "go", "list", "./...")
		}
	}
	if *portal {
		printRun(root, env, "npm", "run", "docs:check")
	}

	fmt.Printf("agentx-developer-preview-distribution-ok:root_version=%s:modules=%d:private_validation_ready=true:public_beta_ready=false:fresh_cache=%t:read_only_cache=%t:full=%t:portal=%t\n", rootVersion, len(releaseModules), *freshCache, *readOnlyCache, *full, *portal)
}

func checkRequiredFiles(root string) {
	for _, relative := range []string{
		"README.md",
		"CHANGELOG.md",
		"SECURITY.md",
		"SUPPORT.md",
		"CONTRIBUTING.md",
		".github/CODEOWNERS",
		"docs/guides/developer-preview-policy.md",
		"docs/reference/distribution-readiness.md",
		"docs/maturity.md",
		"docs/architecture/developer-portal-generator.md",
		"package.json",
		"package-lock.json",
		"portal/.vitepress/config.mjs",
		"portal/content/index.md",
	} {
		info, err := os.Stat(filepath.Join(root, relative))
		if err != nil || info.IsDir() || info.Size() == 0 {
			check(fmt.Errorf("missing non-empty distribution contract %s", relative))
		}
	}
	codeowners, err := os.ReadFile(filepath.Join(root, ".github/CODEOWNERS"))
	check(err)
	if !bytes.Contains(codeowners, []byte("@wsnacj")) {
		check(fmt.Errorf("CODEOWNERS does not name @wsnacj"))
	}
}

func checkGoMods(root string) {
	for _, module := range append(releaseModules, moduleSpec{dir: "conformance/consumer"}) {
		path := filepath.Join(root, module.dir, "go.mod")
		content, err := os.ReadFile(path)
		check(err)
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "replace ") || strings.TrimSpace(line) == "replace (" {
				check(fmt.Errorf("%s contains a replace directive", path))
			}
		}
	}
}

func readVersionMatrix(path string) map[string]string {
	file, err := os.Open(path)
	check(err)
	defer file.Close()
	versions := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || !strings.HasPrefix(fields[1], "v0.0.0-") {
			check(fmt.Errorf("invalid Developer Preview version matrix line %q", line))
		}
		if _, exists := versions[fields[0]]; exists {
			check(fmt.Errorf("duplicate Developer Preview module %s", fields[0]))
		}
		versions[fields[0]] = fields[1]
	}
	check(scanner.Err())
	return versions
}

func pseudoVersionRevision(version string) string {
	index := strings.LastIndex(version, "-")
	if index < 0 || len(version[index+1:]) < 12 {
		check(fmt.Errorf("cannot read revision from pseudo-version %q", version))
	}
	return version[index+1:]
}

func checkArtifact(root string, env []string, module moduleSpec, version, revision string) {
	command := exec.Command("go", "mod", "download", "-json", module.path+"@"+version)
	command.Dir = root
	command.Env = append(env,
		"GOPROXY=off",
		"GONOSUMDB=github.com/wsnacj/agentx-go",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		check(fmt.Errorf("download metadata for %s: %w: %s", module.path, err, bytes.TrimSpace(output)))
	}
	var info downloadInfo
	check(json.Unmarshal(output, &info))
	if info.Path != module.path || info.Version != version || info.Zip == "" || info.Sum == "" {
		check(fmt.Errorf("incomplete artifact metadata for %s: %#v", module.path, info))
	}
	if !strings.HasPrefix(info.Origin.Hash, revision) {
		check(fmt.Errorf("%s origin %q does not match pseudo-version revision %q", module.path, info.Origin.Hash, revision))
	}
	archive, err := zip.OpenReader(info.Zip)
	check(err)
	defer archive.Close()
	prefix := module.path + "@" + version + "/"
	found := false
	for _, file := range archive.File {
		relative := strings.TrimPrefix(file.Name, prefix)
		if relative == module.required {
			found = true
		}
		base := strings.ToLower(filepath.Base(relative))
		if base == ".env" || base == "go.work" {
			check(fmt.Errorf("%s module zip contains forbidden file %s", module.path, relative))
		}
	}
	if !found {
		check(fmt.Errorf("%s module zip is missing %s", module.path, module.required))
	}
	fmt.Printf("agentx-module-artifact-ok:path=%s:version=%s:origin=%s\n", module.path, version, info.Origin.Hash)
}

func printRun(dir string, env []string, name string, args ...string) {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = env
	output, err := command.CombinedOutput()
	if err != nil {
		check(fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, bytes.TrimSpace(output)))
	}
	if len(output) > 0 {
		fmt.Print(string(output))
	}
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Developer Preview distribution gate:", err)
	os.Exit(1)
}
