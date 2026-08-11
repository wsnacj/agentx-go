//go:build ignore

// check_cleanroom_consumer copies the fixed-version consumer outside the source
// checkout and verifies it only from the local immutable module cache.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var expectedModules = []string{
	"github.com/wsnacj/agentx-go",
	"github.com/wsnacj/agentx-go/components",
	"github.com/wsnacj/agentx-go/runtime",
	"github.com/wsnacj/agentx-go/extensions",
	"github.com/wsnacj/agentx-go/scenes",
}

func main() {
	freshCache := flag.Bool("fresh-cache", false, "download fixed modules into an empty temporary module and build cache")
	readOnlyCache := flag.Bool("read-only-cache", false, "freeze the fresh module cache before verify, test and run")
	flag.Parse()
	if *readOnlyCache && !*freshCache {
		check(fmt.Errorf("read-only-cache requires fresh-cache"))
	}

	root, err := os.Getwd()
	check(err)
	versions, err := readVersionMatrix(filepath.Join(root, "docs/reference/developer-preview-module-versions.txt"))
	check(err)
	rootVersion := versions["github.com/wsnacj/agentx-go"]
	if rootVersion == "" {
		check(fmt.Errorf("fixed-version matrix is missing root module"))
	}

	consumer := filepath.Join(root, "conformance/consumer")
	goMod, err := os.ReadFile(filepath.Join(consumer, "go.mod"))
	check(err)
	if bytes.Contains(goMod, []byte("replace ")) || bytes.Contains(goMod, []byte("replace(")) || bytes.Contains(goMod, []byte("replace (")) {
		check(fmt.Errorf("conformance consumer must not contain replace directives"))
	}

	temporary, err := os.MkdirTemp("", "agentx-go-cleanroom-consumer-")
	check(err)
	defer os.RemoveAll(temporary)
	consumerCopy := filepath.Join(temporary, "consumer")
	check(copyConsumer(consumer, consumerCopy))

	mode := "module-cache"
	env := append(os.Environ(),
		"GOWORK=off",
		"GOPROXY=off",
		"GOFLAGS=-mod=readonly",
	)
	if *freshCache {
		mode = "fresh-vcs-cache"
		moduleCache := filepath.Join(temporary, "gomodcache")
		env = append(env,
			"GOPROXY=https://proxy.golang.org,direct",
			"GOSUMDB=sum.golang.org",
			"GOMODCACHE="+moduleCache,
			"GOCACHE="+filepath.Join(temporary, "gocache"),
		)
		// Warm the complete transitive graph before freezing the cache. The
		// no-argument form only guarantees modules named directly by go.mod.
		run(consumerCopy, env, "go", "mod", "download", "all")
		run(consumerCopy, env, "go", "list", "-deps", "./...")
		if *readOnlyCache {
			check(setTreeWritable(moduleCache, false))
			defer func() {
				_ = setTreeWritable(moduleCache, true)
			}()
			// A frozen cache is useful only if the consumer also proves it does
			// not fall back to VCS or the public proxy.
			env = append(env, "GOPROXY=off")
			mode = "fresh-vcs-read-only-cache"
		}
	}
	run(consumerCopy, env, "go", "mod", "verify")
	modules := run(consumerCopy, env, "go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "all")
	for _, module := range expectedModules {
		version := versions[module]
		if version == "" {
			check(fmt.Errorf("fixed-version matrix is missing %s", module))
		}
		want := module + " " + version
		if !containsLine(modules, want) {
			check(fmt.Errorf("clean-room module list missing %q", want))
		}
	}
	run(consumerCopy, env, "go", "test", "./...")
	output := strings.TrimSpace(run(consumerCopy, env, "go", "run", "."))
	if !strings.HasPrefix(output, "agentx-core-developer-preview-ok:") {
		check(fmt.Errorf("unexpected clean-room output %q", output))
	}
	fmt.Printf("agentx-cleanroom-consumer-ok:root=%s:modules=%d:source=%s\n", rootVersion, len(expectedModules), mode)
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
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid version matrix line %q", line)
		}
		versions[fields[0]] = fields[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return versions, nil
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

func copyConsumer(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, input)
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
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

func containsLine(output, want string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "clean-room consumer gate:", err)
	os.Exit(1)
}
