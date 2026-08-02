package browserd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBrowserdHostDoesNotDependOnProductRuntime(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	root := filepath.Dir(file)
	forbidden := []string{"hs/", "core/agentx", "scene/", "agentx/engine", "Runner"}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		blob, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, token := range forbidden {
			if strings.Contains(string(blob), token) {
				t.Fatalf("browserd host boundary violation: found %q in %s", token, path)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
