//go:build ignore

// check_pre_beta_candidate builds a disposable, same-version four-module
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
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	candidateVersion     = "v0.0.0-m6d.0"
	govulncheckModule    = "golang.org/x/vuln/cmd/govulncheck"
	govulncheckVersion   = "v1.6.0"
	candidateGoToolchain = "go1.25.12"
	fixedVersionFile     = "docs/reference/developer-preview-version.txt"
	manifestFile         = "pre-beta-candidate-manifest.txt"
	dependencyGraphFile  = "pre-beta-candidate-modules.txt"
	securityLogName      = "govulncheck-%s.log"
	technicalReadyMarker = "pre_beta_technical_candidate_ready=true"
)

type moduleSpec struct {
	path string
	dir  string
	name string
}

type moduleArtifact struct {
	path   string
	sha256 string
	size   int64
	files  int
}

var candidateModules = []moduleSpec{
	{path: "github.com/wsnacj/agentx-go", dir: ".", name: "root"},
	{path: "github.com/wsnacj/agentx-go/components", dir: "components", name: "components"},
	{path: "github.com/wsnacj/agentx-go/runtime", dir: "runtime", name: "runtime"},
	{path: "github.com/wsnacj/agentx-go/extensions", dir: "extensions", name: "extensions"},
}

func main() {
	artifactDir := flag.String("artifact-dir", "", "write the value-safe manifest and scan logs to this directory")
	flag.Parse()

	root, err := os.Getwd()
	check(err)
	checkRepositoryRoot(root)
	checkCleanTree(root)
	revision := strings.TrimSpace(run(root, os.Environ(), "git", "rev-parse", "HEAD"))
	commitTime := readCommitTime(root)
	fixedVersion := readTrimmed(filepath.Join(root, fixedVersionFile))

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

	baseEnv := append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN="+candidateGoToolchain)
	printRun(root, baseEnv, "go", "run", "./scripts/check_developer_preview_distribution.go")

	stagingRoot := filepath.Join(work, "staging")
	proxyRoot := filepath.Join(work, "proxy")
	moduleCache := filepath.Join(work, "gomodcache")
	buildCache := filepath.Join(work, "gocache")
	check(os.MkdirAll(stagingRoot, 0o755))
	check(os.MkdirAll(proxyRoot, 0o755))

	nestedRoots := trackedModuleRoots(root)
	staged := make(map[string]string, len(candidateModules))
	for _, module := range candidateModules {
		target := filepath.Join(stagingRoot, module.name)
		copyTrackedModule(root, module.dir, target, nestedRoots)
		rewriteCandidateRequirements(target)
		staged[module.path] = target
	}

	candidateEnv := append(baseEnv,
		"GOPROXY=file://"+proxyRoot+",https://proxy.golang.org",
		"GONOSUMDB=github.com/wsnacj/agentx-go",
		"GONOPROXY=none",
		"GOSUMDB=sum.golang.org",
		"GOMODCACHE="+moduleCache,
		"GOCACHE="+buildCache,
	)
	toolBin := filepath.Join(work, "bin")
	check(os.MkdirAll(toolBin, 0o755))
	toolEnv := replaceEnv(candidateEnv, "GOBIN", toolBin)
	printRun(root, toolEnv, "go", "install", govulncheckModule+"@"+govulncheckVersion)
	govulncheckPath := filepath.Join(toolBin, "govulncheck")
	if info, statErr := os.Stat(govulncheckPath); statErr != nil || !info.Mode().IsRegular() {
		check(fmt.Errorf("pinned govulncheck binary was not installed"))
	}
	printRun(root, candidateEnv, govulncheckPath, "-version")

	// The first pass materializes a complete candidate graph. A module near the
	// end of the graph can otherwise retain sums that were needed only while an
	// upstream candidate zip was still absent.
	for _, module := range candidateModules {
		dir := staged[module.path]
		printRun(dir, candidateEnv, "go", "mod", "tidy")
		checkNoReplace(filepath.Join(dir, "go.mod"))
		writeProxyModule(proxyRoot, dir, module, commitTime, false)
	}
	// With all four provisional zips present, stabilize in dependency order and
	// overwrite each zip. Only this second pass is a candidate artifact.
	var artifacts []moduleArtifact
	for _, module := range candidateModules {
		dir := staged[module.path]
		printRun(dir, candidateEnv, "go", "mod", "tidy")
		printRun(dir, candidateEnv, "go", "mod", "tidy", "-diff")
		checkNoReplace(filepath.Join(dir, "go.mod"))
		artifact := writeProxyModule(proxyRoot, dir, module, commitTime, true)
		artifacts = append(artifacts, artifact)
	}

	for _, module := range candidateModules {
		dir := staged[module.path]
		printRun(dir, candidateEnv, "go", "mod", "download", "all")
		printRun(dir, candidateEnv, "go", "mod", "verify")
		printRun(dir, candidateEnv, "go", "test", "./...")
		printRun(dir, candidateEnv, "go", "vet", "./...")
		printRun(dir, candidateEnv, "go", "mod", "tidy", "-diff")
		packages := nonEmptyLines(run(dir, candidateEnv, "go", "list", "./..."))
		if packages == 0 {
			check(fmt.Errorf("%s candidate enumerated zero packages", module.path))
		}
		logPath := filepath.Join(outputDir, fmt.Sprintf(securityLogName, module.name))
		runGovulncheck(govulncheckPath, dir, candidateEnv, work, logPath)
		fmt.Printf("agentx-pre-beta-module-ok:path=%s:packages=%d:known_reachable_vulnerabilities=0\n", module.path, packages)
	}

	consumer := filepath.Join(work, "consumer")
	copyTrackedModule(root, "conformance/consumer", consumer, nestedRoots)
	for _, module := range candidateModules {
		printRun(consumer, candidateEnv, "go", "mod", "edit", "-require="+module.path+"@"+candidateVersion)
	}
	checkNoReplace(filepath.Join(consumer, "go.mod"))
	printRun(consumer, candidateEnv, "go", "mod", "tidy")
	checkNoReplace(filepath.Join(consumer, "go.mod"))
	printRun(consumer, candidateEnv, "go", "mod", "download", "all")
	printRun(consumer, candidateEnv, "go", "list", "-deps", "./...")
	printRun(consumer, candidateEnv, "go", "mod", "verify")
	moduleGraph := run(consumer, candidateEnv, "go", "list", "-m", "-f", "{{.Path}} {{.Version}} {{.Sum}} {{.GoModSum}}", "all")
	checkCandidateSelection(moduleGraph)
	check(os.WriteFile(filepath.Join(outputDir, dependencyGraphFile), []byte(sanitize(moduleGraph, work)), 0o644))
	printRun(consumer, candidateEnv, "go", "test", "./...")
	consumerOutput := strings.TrimSpace(run(consumer, candidateEnv, "go", "run", "."))
	if !strings.HasPrefix(consumerOutput, "agentx-core-developer-preview-ok:") {
		check(fmt.Errorf("unexpected candidate consumer output %q", consumerOutput))
	}

	check(setTreeWritable(moduleCache, false))
	defer func() { _ = setTreeWritable(moduleCache, true) }()
	offlineEnv := replaceEnv(candidateEnv,
		"GOPROXY", "off",
		"GOFLAGS", "-mod=readonly",
	)
	printRun(consumer, offlineEnv, "go", "mod", "verify")
	checkCandidateSelection(run(consumer, offlineEnv, "go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "all"))
	printRun(consumer, offlineEnv, "go", "test", "./...")
	offlineOutput := strings.TrimSpace(run(consumer, offlineEnv, "go", "run", "."))
	if !strings.HasPrefix(offlineOutput, "agentx-core-developer-preview-ok:") {
		check(fmt.Errorf("unexpected offline candidate consumer output %q", offlineOutput))
	}

	manifest := buildManifest(revision, commitTime, fixedVersion, artifacts)
	check(os.WriteFile(filepath.Join(outputDir, manifestFile), []byte(manifest), 0o644))
	checkCleanTree(root)

	fmt.Printf("agentx-pre-beta-candidate-ok:version=%s:modules=%d:source=%s:scanner=%s@%s:%s:public_beta_ready=false\n",
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

func rewriteCandidateRequirements(dir string) {
	goMod := filepath.Join(dir, "go.mod")
	content := readTrimmed(goMod)
	for _, module := range candidateModules {
		if strings.Contains(content, module.path+" ") {
			printRun(dir, append(os.Environ(), "GOWORK=off"), "go", "mod", "edit", "-require="+module.path+"@"+candidateVersion)
		}
	}
	checkNoReplace(goMod)
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

func runGovulncheck(binary, dir string, env []string, work, logPath string) {
	command := exec.Command(binary, "./...")
	command.Dir = dir
	command.Env = env
	output, err := command.CombinedOutput()
	safeOutput := sanitize(string(output), work)
	check(os.WriteFile(logPath, []byte(safeOutput), 0o644))
	if err != nil {
		check(fmt.Errorf("govulncheck %s: %w: %s", dir, err, strings.TrimSpace(safeOutput)))
	}
	if safeOutput != "" {
		fmt.Print(safeOutput)
	}
}

func checkCandidateSelection(output string) {
	for _, module := range candidateModules {
		want := module.path + " " + candidateVersion
		if !containsLinePrefix(output, want) {
			check(fmt.Errorf("candidate module list missing %q", want))
		}
	}
}

func buildManifest(revision string, commitTime time.Time, fixedVersion string, artifacts []moduleArtifact) string {
	var builder strings.Builder
	fmt.Fprintln(&builder, "agentx_pre_beta_candidate_manifest")
	fmt.Fprintf(&builder, "source_revision=%s\n", revision)
	fmt.Fprintf(&builder, "source_commit_time=%s\n", commitTime.Format(time.RFC3339))
	fmt.Fprintf(&builder, "gate_build_go_version=%s\n", runtime.Version())
	fmt.Fprintf(&builder, "candidate_go_toolchain=%s\n", candidateGoToolchain)
	fmt.Fprintf(&builder, "candidate_version=%s\n", candidateVersion)
	fmt.Fprintf(&builder, "candidate_version_scope=disposable_validation_only\n")
	fmt.Fprintf(&builder, "rollback_version=%s\n", fixedVersion)
	fmt.Fprintf(&builder, "security_scanner=%s@%s\n", govulncheckModule, govulncheckVersion)
	fmt.Fprintf(&builder, "known_reachable_vulnerabilities=0\n")
	fmt.Fprintf(&builder, "license_notice_status=pending_owner_decision\n")
	fmt.Fprintf(&builder, "named_security_approval_status=pending\n")
	fmt.Fprintf(&builder, "release_authorization_status=pending\n")
	fmt.Fprintf(&builder, "compatibility_promotion_status=developer_preview_only\n")
	for _, artifact := range artifacts {
		fmt.Fprintf(&builder, "module=%s version=%s zip_sha256=%s zip_bytes=%d zip_files=%d\n",
			artifact.path, candidateVersion, artifact.sha256, artifact.size, artifact.files)
	}
	fmt.Fprintln(&builder, technicalReadyMarker)
	fmt.Fprintln(&builder, "public_beta_ready=false")
	return builder.String()
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

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Pre-Beta candidate gate:", err)
	os.Exit(1)
}
