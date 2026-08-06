//go:build ignore

// check_pre_beta_candidate builds a disposable, same-version nine-module
// candidate from tracked source and verifies it through a local Go proxy.
// It never creates tags, releases, or tracked version changes.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	candidateVersion     = "v0.2.0"
	govulncheckModule    = "golang.org/x/vuln/cmd/govulncheck"
	govulncheckVersion   = "v1.6.0"
	candidateGoToolchain = "go1.25.12"
	manifestFile         = "pre-beta-candidate-manifest.txt"
	dependencyGraphFile  = "pre-beta-candidate-modules.txt"
	securityLogName      = "govulncheck-%s.log"
	technicalReadyMarker = "developer_preview_release_candidate_ready=true"
	candidateConsumerOK  = "agentx-pre-beta-consumer-ok:"
)

type moduleSpec struct {
	path         string
	dir          string
	name         string
	consumerPath string
}

type moduleArtifact struct {
	path   string
	sha256 string
	size   int64
	files  int
}

type securitySummary struct {
	path        string
	reachable   int
	imported    int
	required    int
	residualIDs []string
}

var (
	securityCountPattern = regexp.MustCompile(`This scan also found ([0-9]+) vulnerabilit(y|ies) in packages you import and ([0-9]+)[[:space:]]+vulnerabilit(y|ies) in modules you require`)
	vulnerabilityPattern = regexp.MustCompile(`GO-[0-9]{4}-[0-9]+`)
)

var candidateModules = []moduleSpec{
	{path: "github.com/wsnacj/agentx-go", dir: ".", name: "root", consumerPath: "github.com/wsnacj/agentx-go"},
	{path: "github.com/wsnacj/agentx-go/components", dir: "components", name: "components", consumerPath: "github.com/wsnacj/agentx-go/components/llm"},
	{path: "github.com/wsnacj/agentx-go/runtime", dir: "runtime", name: "runtime", consumerPath: "github.com/wsnacj/agentx-go/runtime/protocol"},
	{path: "github.com/wsnacj/agentx-go/extensions", dir: "extensions", name: "extensions", consumerPath: "github.com/wsnacj/agentx-go/extensions/domainkit"},
	{path: "github.com/wsnacj/agentx-go/providers", dir: "providers", name: "providers", consumerPath: "github.com/wsnacj/agentx-go/providers/fault"},
	{path: "github.com/wsnacj/agentx-go/tools", dir: "tools", name: "tools", consumerPath: "github.com/wsnacj/agentx-go/tools/diffs"},
	{path: "github.com/wsnacj/agentx-go/browser", dir: "browser", name: "browser", consumerPath: "github.com/wsnacj/agentx-go/browser/runtime"},
	{path: "github.com/wsnacj/agentx-go/document", dir: "document", name: "document", consumerPath: "github.com/wsnacj/agentx-go/document/contracts"},
	{path: "github.com/wsnacj/agentx-go/scenes", dir: "scenes", name: "scenes", consumerPath: "github.com/wsnacj/agentx-go/scenes/publictransport"},
}

func main() {
	artifactDir := flag.String("artifact-dir", "", "write the value-safe manifest and scan logs to this directory")
	persistentProxyDir := flag.String("proxy-dir", "", "write the verified local Go proxy to this empty directory instead of deleting it")
	prepareProxyOnly := flag.Bool("prepare-proxy-only", false, "prepare the same-version local Go proxy, then stop before tests and security scans")
	flag.Parse()
	if *prepareProxyOnly && *persistentProxyDir == "" {
		check(fmt.Errorf("prepare-proxy-only requires proxy-dir"))
	}
	checkSecurityParser()

	root, err := os.Getwd()
	check(err)
	checkRepositoryRoot(root)
	checkCleanTree(root)
	revision := strings.TrimSpace(run(root, os.Environ(), "git", "rev-parse", "HEAD"))
	commitTime := readCommitTime(root)
	rollbackRevision := strings.TrimSpace(run(root, os.Environ(), "git", "rev-parse", "HEAD^"))
	goCommand := resolveCandidateGo(root)

	outputDir := *artifactDir
	removeOutput := false
	if outputDir == "" {
		outputDir, err = os.MkdirTemp("", "agentx-m6d-artifacts-")
		check(err)
		removeOutput = true
	}
	check(os.MkdirAll(outputDir, 0o755))
	if removeOutput {
		defer os.RemoveAll(outputDir)
	}

	work, err := os.MkdirTemp("", "agentx-m6d-candidate-")
	check(err)
	defer os.RemoveAll(work)

	baseEnv := append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	printRun(root, baseEnv, goCommand, "run", "./scripts/check_release_legal.go")
	printRun(root, baseEnv, goCommand, "run", "./scripts/check_release_policy.go")
	printRun(root, baseEnv, goCommand, "run", "./scripts/check_developer_preview_version.go")
	printRun(root, baseEnv, goCommand, "run", "./scripts/check_public_docs.go")

	stagingRoot := filepath.Join(work, "staging")
	proxyRoot := filepath.Join(work, "proxy")
	if *persistentProxyDir != "" {
		proxyRoot, err = filepath.Abs(*persistentProxyDir)
		check(err)
		entries, readErr := os.ReadDir(proxyRoot)
		if readErr == nil && len(entries) > 0 {
			check(fmt.Errorf("proxy-dir must be empty: %s", proxyRoot))
		}
		if readErr != nil && !os.IsNotExist(readErr) {
			check(readErr)
		}
	}
	moduleCache := filepath.Join(work, "gomodcache")
	buildCache := filepath.Join(work, "gocache")
	check(os.MkdirAll(stagingRoot, 0o755))
	check(os.MkdirAll(proxyRoot, 0o755))

	nestedRoots := trackedModuleRoots(root)
	staged := make(map[string]string, len(candidateModules))
	var securitySummaries []securitySummary
	for _, module := range candidateModules {
		target := filepath.Join(stagingRoot, module.name)
		copyTrackedModule(root, module.dir, target, nestedRoots)
		rewriteCandidateRequirements(target, goCommand, baseEnv)
		staged[module.path] = target
	}

	candidatePath := filepath.Dir(goCommand) + string(os.PathListSeparator) + os.Getenv("PATH")
	candidateEnv := append(replaceEnv(baseEnv, "PATH", candidatePath),
		// govulncheck intentionally accepts GOVERSION for source-mode standard
		// library analysis. Keep this explicit: a newer launcher alone does not
		// change the version modeled by the scanner.
		"GOVERSION="+candidateGoToolchain,
		"GOPROXY=file://"+proxyRoot+",https://proxy.golang.org",
		"GONOSUMDB=github.com/wsnacj/agentx-go",
		"GONOPROXY=none",
		"GOSUMDB=sum.golang.org",
		"GOMODCACHE="+moduleCache,
		"GOCACHE="+buildCache,
	)
	// The first pass materializes a complete candidate graph. A module near the
	// end of the graph can otherwise retain sums that were needed only while an
	// upstream candidate zip was still absent.
	for _, module := range candidateModules {
		dir := staged[module.path]
		printRunTransient(dir, candidateEnv, work, goCommand, "mod", "tidy")
		checkNoReplace(filepath.Join(dir, "go.mod"))
		writeProxyModule(proxyRoot, dir, module, commitTime, false)
	}
	// With all nine provisional zips present, stabilize in dependency order and
	// overwrite each zip. Only this second pass is a candidate artifact.
	var artifacts []moduleArtifact
	for _, module := range candidateModules {
		dir := staged[module.path]
		printRunTransient(dir, candidateEnv, work, goCommand, "mod", "tidy")
		printRun(dir, candidateEnv, goCommand, "mod", "tidy", "-diff")
		checkNoReplace(filepath.Join(dir, "go.mod"))
		artifact := writeProxyModule(proxyRoot, dir, module, commitTime, true)
		artifacts = append(artifacts, artifact)
	}
	if *prepareProxyOnly {
		checkCleanTree(root)
		fmt.Printf("agentx-release-proxy-prepared:version=%s:modules=%d:source=%s:proxy=%s\n",
			candidateVersion, len(candidateModules), revision, proxyRoot)
		return
	}

	toolBin := filepath.Join(work, "bin")
	check(os.MkdirAll(toolBin, 0o755))
	toolEnv := replaceEnv(candidateEnv, "GOBIN", toolBin)
	printRunTransient(root, toolEnv, work, goCommand, "install", govulncheckModule+"@"+govulncheckVersion)
	govulncheckPath := filepath.Join(toolBin, "govulncheck")
	if info, statErr := os.Stat(govulncheckPath); statErr != nil || !info.Mode().IsRegular() {
		check(fmt.Errorf("pinned govulncheck binary was not installed"))
	}
	printRun(root, candidateEnv, govulncheckPath, "-version")

	printRun(root, candidateEnv, goCommand, "run", "./scripts/check_developer_preview_api.go", "-check-platforms")
	printRun(root, candidateEnv, goCommand, "run", "./scripts/check_package_api_docs.go")
	printRun(root, candidateEnv, goCommand, "run", "./scripts/check_docs_links.go")

	for _, module := range candidateModules {
		dir := staged[module.path]
		printRun(dir, candidateEnv, goCommand, "test", "./...")
		printRun(dir, candidateEnv, goCommand, "test", "-race", "./...")
		printRun(dir, candidateEnv, goCommand, "vet", "./...")
		printRun(dir, candidateEnv, goCommand, "mod", "verify")
		printRun(dir, candidateEnv, goCommand, "mod", "tidy", "-diff")
		packages := nonEmptyLines(run(dir, candidateEnv, goCommand, "list", "./..."))
		if packages == 0 {
			check(fmt.Errorf("%s candidate enumerated zero packages", module.path))
		}
		logPath := filepath.Join(outputDir, fmt.Sprintf(securityLogName, module.name))
		security := runGovulncheck(govulncheckPath, dir, candidateEnv, work, logPath)
		security.path = module.path
		securitySummaries = append(securitySummaries, security)
		fmt.Printf("agentx-pre-beta-module-ok:path=%s:packages=%d:known_reachable_vulnerabilities=%d:imported_unreachable=%d:module_unreachable=%d\n",
			module.path, packages, security.reachable, security.imported, security.required)
	}

	consumer := filepath.Join(work, "consumer")
	writeCandidateConsumer(consumer)
	checkNoReplace(filepath.Join(consumer, "go.mod"))
	printRunTransient(consumer, candidateEnv, work, goCommand, "mod", "tidy")
	checkNoReplace(filepath.Join(consumer, "go.mod"))
	printRunTransient(consumer, candidateEnv, work, goCommand, "mod", "download", "all")
	if packages := nonEmptyLines(run(consumer, candidateEnv, goCommand, "list", "-deps", "./...")); packages == 0 {
		check(fmt.Errorf("candidate consumer enumerated zero dependencies"))
	}
	printRun(consumer, candidateEnv, goCommand, "mod", "verify")
	moduleGraph := run(consumer, candidateEnv, goCommand, "list", "-m", "-f", "{{.Path}} {{.Version}} {{.Sum}} {{.GoModSum}}", "all")
	checkCandidateSelection(moduleGraph)
	check(os.WriteFile(filepath.Join(outputDir, dependencyGraphFile), []byte(sanitize(moduleGraph, work)), 0o644))
	printRun(consumer, candidateEnv, goCommand, "test", "./...")
	consumerOutput := strings.TrimSpace(run(consumer, candidateEnv, goCommand, "run", "."))
	if !strings.HasPrefix(consumerOutput, candidateConsumerOK) {
		check(fmt.Errorf("unexpected candidate consumer output %q", consumerOutput))
	}

	check(setTreeWritable(moduleCache, false))
	defer func() { _ = setTreeWritable(moduleCache, true) }()
	offlineEnv := replaceEnv(candidateEnv,
		"GOPROXY", "off",
		"GOFLAGS", "-mod=readonly",
	)
	printRun(consumer, offlineEnv, goCommand, "mod", "verify")
	checkCandidateSelection(run(consumer, offlineEnv, goCommand, "list", "-m", "-f", "{{.Path}} {{.Version}}", "all"))
	printRun(consumer, offlineEnv, goCommand, "test", "./...")
	offlineOutput := strings.TrimSpace(run(consumer, offlineEnv, goCommand, "run", "."))
	if !strings.HasPrefix(offlineOutput, candidateConsumerOK) {
		check(fmt.Errorf("unexpected offline candidate consumer output %q", offlineOutput))
	}

	manifest := buildManifest(revision, commitTime, rollbackRevision, artifacts, securitySummaries, releaseLegalStatus(root))
	check(os.WriteFile(filepath.Join(outputDir, manifestFile), []byte(manifest), 0o644))
	checkCleanTree(root)

	fmt.Printf("agentx-developer-preview-release-candidate-ok:version=%s:modules=%d:source=%s:scanner=%s@%s:%s:public_beta_ready=false\n",
		candidateVersion, len(candidateModules), revision, govulncheckModule, govulncheckVersion, technicalReadyMarker)
}

func checkRepositoryRoot(root string) {
	module := readTrimmed(filepath.Join(root, "go.mod"))
	if !strings.HasPrefix(module, "module github.com/wsnacj/agentx-go\n") {
		check(fmt.Errorf("run from the agentx-go repository root"))
	}
}

func checkCleanTree(root string) {
	if output := strings.TrimSpace(run(root, os.Environ(), "git", "status", "--porcelain")); output != "" {
		check(fmt.Errorf("candidate gate requires a clean tracked source tree: %s", output))
	}
}

func readCommitTime(root string) time.Time {
	value := strings.TrimSpace(run(root, os.Environ(), "git", "show", "-s", "--format=%cI", "HEAD"))
	parsed, err := time.Parse(time.RFC3339, value)
	check(err)
	return parsed.UTC()
}

func resolveCandidateGo(root string) string {
	bootstrapEnv := replaceEnv(os.Environ(), "GOTOOLCHAIN", candidateGoToolchain)
	goRoot := strings.TrimSpace(run(root, bootstrapEnv, "go", "env", "GOROOT"))
	goCommand := filepath.Join(goRoot, "bin", "go")
	info, err := os.Stat(goCommand)
	if err != nil || !info.Mode().IsRegular() {
		check(fmt.Errorf("cannot resolve %s go command", candidateGoToolchain))
	}
	version := strings.TrimSpace(run(root, replaceEnv(bootstrapEnv, "GOTOOLCHAIN", "local"), goCommand, "version"))
	if !strings.Contains(version, "go version "+candidateGoToolchain+" ") {
		check(fmt.Errorf("resolved candidate go command = %q, want %s", version, candidateGoToolchain))
	}
	fmt.Printf("agentx-pre-beta-toolchain-ok:version=%s\n", candidateGoToolchain)
	return goCommand
}

func trackedModuleRoots(root string) []string {
	output := run(root, os.Environ(), "git", "ls-files", "-z", "--", "**/go.mod")
	var roots []string
	for _, path := range strings.Split(output, "\x00") {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" || path == "go.mod" {
			continue
		}
		roots = append(roots, strings.TrimSuffix(path, "/go.mod"))
	}
	sort.Strings(roots)
	return roots
}

func copyTrackedModule(root, moduleDir, target string, nestedRoots []string) {
	pathspec := moduleDir
	if moduleDir == "." {
		pathspec = "."
	}
	output := run(root, os.Environ(), "git", "ls-files", "-z", "--", pathspec)
	var copied int
	modulePrefix := ""
	if moduleDir != "." {
		modulePrefix = filepath.ToSlash(moduleDir) + "/"
	}
	for _, tracked := range strings.Split(output, "\x00") {
		tracked = filepath.ToSlash(tracked)
		if tracked == "" || !strings.HasPrefix(tracked, modulePrefix) {
			continue
		}
		relative := strings.TrimPrefix(tracked, modulePrefix)
		if relative == "" || belongsToNestedModule(tracked, moduleDir, nestedRoots) {
			continue
		}
		source := filepath.Join(root, filepath.FromSlash(tracked))
		info, err := os.Lstat(source)
		check(err)
		if !info.Mode().IsRegular() {
			check(fmt.Errorf("candidate source must be a regular file: %s", tracked))
		}
		destination := filepath.Join(target, filepath.FromSlash(relative))
		check(os.MkdirAll(filepath.Dir(destination), 0o755))
		check(copyFile(source, destination, info.Mode().Perm()))
		copied++
	}
	if copied == 0 {
		check(fmt.Errorf("tracked module %s copied zero files", moduleDir))
	}
	if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
		check(fmt.Errorf("tracked module %s is missing go.mod: %w", moduleDir, err))
	}
}

func belongsToNestedModule(tracked, moduleDir string, nestedRoots []string) bool {
	moduleDir = filepath.ToSlash(moduleDir)
	for _, nested := range nestedRoots {
		if moduleDir != "." && nested == moduleDir {
			continue
		}
		if moduleDir != "." && !strings.HasPrefix(nested, moduleDir+"/") {
			continue
		}
		if tracked == nested || strings.HasPrefix(tracked, nested+"/") {
			return true
		}
	}
	return false
}

func rewriteCandidateRequirements(dir, goCommand string, env []string) {
	goMod := filepath.Join(dir, "go.mod")
	content := readTrimmed(goMod)
	for _, module := range candidateModules {
		if strings.Contains(content, module.path+" ") {
			printRun(dir, env, goCommand, "mod", "edit", "-require="+module.path+"@"+candidateVersion)
		}
	}
	checkNoReplace(goMod)
	// Candidate modules are rebuilt from this exact tracked source revision.
	// Any previously committed checksum for the same candidate version belongs
	// to an older local artifact and would make the next revision impossible to
	// stage. Remove only AgentX-owned sums in the temporary copy; third-party
	// checksums remain authoritative and the final artifact sums are recalculated.
	resetAgentXSums(filepath.Join(dir, "go.sum"))
}

func resetAgentXSums(path string) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return
	}
	check(err)
	var kept []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "github.com/wsnacj/agentx-go ") ||
			strings.HasPrefix(line, "github.com/wsnacj/agentx-go/") {
			continue
		}
		kept = append(kept, line)
	}
	check(os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644))
}

func writeCandidateConsumer(dir string) {
	check(os.MkdirAll(dir, 0o755))
	var goMod strings.Builder
	fmt.Fprintln(&goMod, "module agentx-pre-beta-consumer")
	fmt.Fprintln(&goMod)
	fmt.Fprintln(&goMod, "go 1.25.0")
	fmt.Fprintln(&goMod)
	fmt.Fprintln(&goMod, "require (")
	for _, module := range candidateModules {
		fmt.Fprintf(&goMod, "\t%s %s\n", module.path, candidateVersion)
	}
	fmt.Fprintln(&goMod, ")")
	check(os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod.String()), 0o644))

	var source strings.Builder
	fmt.Fprintln(&source, "package main")
	fmt.Fprintln(&source)
	fmt.Fprintln(&source, "import (")
	fmt.Fprintln(&source, "\t\"fmt\"")
	for _, module := range candidateModules {
		fmt.Fprintf(&source, "\t_ %q\n", module.consumerPath)
	}
	fmt.Fprintln(&source, ")")
	fmt.Fprintln(&source)
	fmt.Fprintln(&source, "func main() {")
	fmt.Fprintf(&source, "\tfmt.Println(%q)\n", candidateConsumerOK+"modules=9")
	fmt.Fprintln(&source, "}")
	check(os.WriteFile(filepath.Join(dir, "main.go"), []byte(source.String()), 0o644))
}

func writeProxyModule(proxyRoot, source string, module moduleSpec, commitTime time.Time, announce bool) moduleArtifact {
	versionDir := filepath.Join(proxyRoot, filepath.FromSlash(module.path), "@v")
	check(os.MkdirAll(versionDir, 0o755))
	modBytes, err := os.ReadFile(filepath.Join(source, "go.mod"))
	check(err)
	check(os.WriteFile(filepath.Join(versionDir, candidateVersion+".mod"), modBytes, 0o644))
	info := fmt.Sprintf("{\"Version\":%q,\"Time\":%q}\n", candidateVersion, commitTime.Format(time.RFC3339))
	check(os.WriteFile(filepath.Join(versionDir, candidateVersion+".info"), []byte(info), 0o644))
	check(os.WriteFile(filepath.Join(versionDir, "list"), []byte(candidateVersion+"\n"), 0o644))

	zipPath := filepath.Join(versionDir, candidateVersion+".zip")
	files := writeModuleZip(zipPath, source, module.path, commitTime)
	hash, size := hashFile(zipPath)
	if announce {
		fmt.Printf("agentx-pre-beta-artifact-ok:path=%s:version=%s:sha256=%s:bytes=%d:files=%d\n",
			module.path, candidateVersion, hash, size, files)
	}
	return moduleArtifact{path: module.path, sha256: hash, size: size, files: files}
}

func writeModuleZip(target, source, modulePath string, commitTime time.Time) int {
	output, err := os.Create(target)
	check(err)
	archive := zip.NewWriter(output)
	var files []string
	check(filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("module zip source contains symlink %s", path)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	}))
	sort.Strings(files)
	for _, relative := range files {
		path := filepath.Join(source, filepath.FromSlash(relative))
		info, err := os.Stat(path)
		check(err)
		header, err := zip.FileInfoHeader(info)
		check(err)
		header.Name = modulePath + "@" + candidateVersion + "/" + relative
		header.Method = zip.Deflate
		header.SetModTime(commitTime)
		writer, err := archive.CreateHeader(header)
		check(err)
		input, err := os.Open(path)
		check(err)
		_, copyErr := io.Copy(writer, input)
		closeErr := input.Close()
		check(copyErr)
		check(closeErr)
	}
	check(archive.Close())
	check(output.Close())
	return len(files)
}

func runGovulncheck(binary, dir string, env []string, work, logPath string) securitySummary {
	var log strings.Builder
	for attempt := 1; attempt <= 3; attempt++ {
		command := exec.Command(binary, "-show", "verbose", "./...")
		command.Dir = dir
		command.Env = env
		output, err := command.CombinedOutput()
		safeOutput := sanitize(string(output), work)
		fmt.Fprintf(&log, "attempt=%d\n%s", attempt, safeOutput)
		if err == nil {
			check(os.WriteFile(logPath, []byte(log.String()), 0o644))
			return parseSecuritySummary(safeOutput)
		}
		if !isTransientScanFailure(safeOutput) {
			check(os.WriteFile(logPath, []byte(log.String()), 0o644))
			check(fmt.Errorf("govulncheck %s: %w: %s", dir, err, strings.TrimSpace(safeOutput)))
		}
		if attempt < 3 {
			fmt.Printf("agentx-govulncheck-transient-retry:attempt=%d\n", attempt)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		check(os.WriteFile(logPath, []byte(log.String()), 0o644))
		check(fmt.Errorf("govulncheck %s exhausted transient retries: %w: %s", dir, err, strings.TrimSpace(safeOutput)))
	}
	return securitySummary{}
}

func parseSecuritySummary(output string) securitySummary {
	summary := securitySummary{}
	if !strings.Contains(output, "the "+candidateGoToolchain+" standard library") {
		check(fmt.Errorf("govulncheck output did not prove %s standard library analysis", candidateGoToolchain))
	}
	if strings.Contains(output, "Your code is affected by 0 vulnerabilities.") || strings.Contains(output, "No vulnerabilities found.") {
		summary.reachable = 0
	} else {
		check(fmt.Errorf("govulncheck success output did not prove zero reachable vulnerabilities"))
	}
	if match := securityCountPattern.FindStringSubmatch(output); len(match) == 5 {
		summary.imported = mustAtoi(match[1])
		summary.required = mustAtoi(match[3])
	}
	seen := map[string]bool{}
	for _, id := range vulnerabilityPattern.FindAllString(output, -1) {
		if seen[id] {
			continue
		}
		seen[id] = true
		summary.residualIDs = append(summary.residualIDs, id)
	}
	sort.Strings(summary.residualIDs)
	if summary.imported+summary.required != len(summary.residualIDs) {
		check(fmt.Errorf("govulncheck residual count mismatch: imported=%d required=%d ids=%v",
			summary.imported, summary.required, summary.residualIDs))
	}
	return summary
}

func checkSecurityParser() {
	output := `Govulncheck scanned the following 6 modules and the go1.25.12 standard library:
Vulnerability #1: GO-2026-5024
Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.`
	summary := parseSecuritySummary(output)
	if summary.reachable != 0 || summary.imported != 0 || summary.required != 1 || len(summary.residualIDs) != 1 || summary.residualIDs[0] != "GO-2026-5024" {
		check(fmt.Errorf("govulncheck parser self-check failed: %+v", summary))
	}
}

func isTransientScanFailure(output string) bool {
	lower := strings.ToLower(output)
	if strings.Contains(lower, "vulnerability #") || strings.Contains(lower, "your code is affected by") {
		return false
	}
	for _, marker := range []string{
		"connection reset",
		"connection refused",
		"i/o timeout",
		"tls handshake timeout",
		"temporary failure",
		"no such host",
		"unexpected eof",
		"server misbehaving",
		"502 bad gateway",
		"503 service unavailable",
		"504 gateway timeout",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func checkCandidateSelection(output string) {
	for _, module := range candidateModules {
		want := module.path + " " + candidateVersion
		if !containsLinePrefix(output, want) {
			check(fmt.Errorf("candidate module list missing %q", want))
		}
	}
}

func buildManifest(revision string, commitTime time.Time, rollbackRevision string, artifacts []moduleArtifact, security []securitySummary, legalStatus string) string {
	var builder strings.Builder
	fmt.Fprintln(&builder, "agentx_pre_beta_candidate_manifest")
	fmt.Fprintf(&builder, "source_revision=%s\n", revision)
	fmt.Fprintf(&builder, "source_commit_time=%s\n", commitTime.Format(time.RFC3339))
	fmt.Fprintf(&builder, "candidate_go_toolchain=%s\n", candidateGoToolchain)
	fmt.Fprintf(&builder, "security_standard_library_version=%s\n", candidateGoToolchain)
	fmt.Fprintf(&builder, "candidate_version=%s\n", candidateVersion)
	fmt.Fprintf(&builder, "candidate_version_scope=public_v0.2.0_developer_preview_candidate\n")
	fmt.Fprintf(&builder, "rollback_revision=%s\n", rollbackRevision)
	fmt.Fprintf(&builder, "rollback_strategy=withdraw_release_and_restore_pre_release_branch\n")
	fmt.Fprintf(&builder, "security_scanner=%s@%s\n", govulncheckModule, govulncheckVersion)
	fmt.Fprintf(&builder, "known_reachable_vulnerabilities=0\n")
	fmt.Fprintf(&builder, "license_notice_status=%s\n", legalStatus)
	fmt.Fprintf(&builder, "named_security_approval_status=approved_at_wsnacj\n")
	fmt.Fprintf(&builder, "release_authorization_status=candidate_validation_authorized_tags_pending\n")
	fmt.Fprintf(&builder, "public_visibility_authorization_status=pending\n")
	fmt.Fprintf(&builder, "compatibility_promotion_status=developer_preview_9_packages\n")
	for _, artifact := range artifacts {
		fmt.Fprintf(&builder, "module=%s version=%s zip_sha256=%s zip_bytes=%d zip_files=%d\n",
			artifact.path, candidateVersion, artifact.sha256, artifact.size, artifact.files)
	}
	for _, summary := range security {
		ids := "none"
		if len(summary.residualIDs) > 0 {
			ids = strings.Join(summary.residualIDs, ",")
		}
		fmt.Fprintf(&builder, "security=%s reachable=%d imported_unreachable=%d module_unreachable=%d residual_ids=%s\n",
			summary.path, summary.reachable, summary.imported, summary.required, ids)
	}
	fmt.Fprintln(&builder, technicalReadyMarker)
	fmt.Fprintln(&builder, "public_beta_ready=false")
	fmt.Fprintln(&builder, "public_release_ready=false")
	return builder.String()
}

func releaseLegalStatus(root string) string {
	license := hasRegularFile(filepath.Join(root, "LICENSE")) || hasRegularFile(filepath.Join(root, "LICENSE.txt")) || hasRegularFile(filepath.Join(root, "LICENSE.md"))
	notice := hasRegularFile(filepath.Join(root, "NOTICE")) || hasRegularFile(filepath.Join(root, "NOTICE.txt")) || hasRegularFile(filepath.Join(root, "NOTICE.md"))
	switch {
	case license && notice:
		return "apache_2_0_owner_approved"
	case license:
		return "license_present_notice_missing_owner_decision"
	default:
		return "missing_owner_decision"
	}
}

func hasRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

func mustAtoi(value string) int {
	parsed, err := strconv.Atoi(value)
	check(err)
	return parsed
}

func checkNoReplace(path string) {
	content, err := os.ReadFile(path)
	check(err)
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "replace ") || trimmed == "replace(" || trimmed == "replace (" {
			check(fmt.Errorf("%s contains replace directive", path))
		}
	}
}

func setTreeWritable(root string, writable bool) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		if writable {
			mode |= 0o200
		} else {
			mode &^= 0o222
		}
		return os.Chmod(path, mode)
	})
}

func replaceEnv(env []string, pairs ...string) []string {
	result := append([]string{}, env...)
	for index := 0; index < len(pairs); index += 2 {
		key, value := pairs[index], pairs[index+1]
		prefix := key + "="
		updated := false
		for i := range result {
			if strings.HasPrefix(result[i], prefix) {
				result[i] = prefix + value
				updated = true
			}
		}
		if !updated {
			result = append(result, prefix+value)
		}
	}
	return result
}

func copyFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func hashFile(path string) (string, int64) {
	file, err := os.Open(path)
	check(err)
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	check(err)
	return hex.EncodeToString(hash.Sum(nil)), size
}

func readTrimmed(path string) string {
	content, err := os.ReadFile(path)
	check(err)
	return strings.TrimSpace(string(content))
}

func sanitize(value, work string) string {
	return strings.ReplaceAll(value, work, "<candidate-work>")
}

func nonEmptyLines(value string) int {
	count := 0
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func containsLinePrefix(output, want string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), want) {
			return true
		}
	}
	return false
}

func run(dir string, env []string, name string, args ...string) string {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = env
	output, err := command.CombinedOutput()
	if err != nil {
		check(fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, bytes.TrimSpace(output)))
	}
	return string(output)
}

func printRun(dir string, env []string, name string, args ...string) {
	output := run(dir, env, name, args...)
	if output != "" {
		fmt.Print(output)
	}
}

func printRunTransient(dir string, env []string, work, name string, args ...string) {
	for attempt := 1; attempt <= 3; attempt++ {
		command := exec.Command(name, args...)
		command.Dir = dir
		command.Env = env
		output, err := command.CombinedOutput()
		if err == nil {
			if len(output) > 0 {
				fmt.Print(sanitize(string(output), work))
			}
			return
		}
		safeOutput := sanitize(string(output), work)
		if !isTransientScanFailure(safeOutput) {
			check(fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(safeOutput)))
		}
		if attempt == 3 {
			check(fmt.Errorf("%s %s exhausted transient retries: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(safeOutput)))
		}
		fmt.Printf("agentx-candidate-network-transient-retry:command=%s:attempt=%d\n", strings.Join(append([]string{name}, args...), " "), attempt)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Pre-Beta candidate gate:", err)
	os.Exit(1)
}
