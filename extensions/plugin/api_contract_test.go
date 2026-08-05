package plugin_test

import (
	"os"
	"strings"
	"testing"
)

func TestPluginAPIReferenceCoversContractAndBoundaries(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, required := range []string{
		"Manifest", "Dependency", "PermissionRequest", "Parse", "Normalize",
		"ErrorCode", "不表示Host已经授权", "不扫描默认目录", "不运行command/hook",
		"不实例化", "Experimental",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
