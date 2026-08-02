package browserruntime

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBrowserRuntimeRemainsSiteAgnostic(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	root := filepath.Dir(file)
	forbidden := []string{
		"doubao", "doubao.com", "豆包",
		"xiaohongshu", "xiaohongshu.com", "小红书",
		"taobao", "taobao.com", "淘宝",
		"jd.com", "京东",
		"volcengine", "volcengine.com", "火山引擎",
	}
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
		body := strings.ToLower(string(blob))
		for _, token := range forbidden {
			if strings.Contains(body, strings.ToLower(token)) {
				t.Fatalf("browser runtime boundary violation: found %q in %s", token, path)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
