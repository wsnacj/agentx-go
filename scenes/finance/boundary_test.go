package finance_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestFinanceHasNoHostOrLegacySceneImports(t *testing.T) {
	_, current, _, _ := runtime.Caller(0)
	root := filepath.Dir(current)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, item := range file.Imports {
			value, _ := strconv.Unquote(item.Path.Value)
			if strings.HasPrefix(value, "hs/") || strings.Contains(value, "/scene/agentx_") {
				t.Errorf("%s imports forbidden host path %q", path, value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
