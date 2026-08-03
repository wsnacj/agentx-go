package videoframes

import (
	"os"
	"strings"
	"testing"
)

func TestChineseAPIReferenceCoversCandidateSurface(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	reference := string(content)
	for _, required := range []string{
		"NewLocalAdapter", "Available", "Register", "Definition",
		"LocalOptions", "LocalAdapter", "Result", "Probe", "Frame",
		"FilesTouched", "ErrUnsafeFile", "ErrFileTooLarge",
		"授权", "sandbox", "取消", "并发",
	} {
		if !strings.Contains(reference, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
