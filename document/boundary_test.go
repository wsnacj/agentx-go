package document_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentModuleDoesNotImportProductOwners(t *testing.T) {
	forbidden := []string{`"hs/`, `"scene/`, "engine.Runner", "core/agentx"}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, token := range forbidden {
			if strings.Contains(string(payload), token) {
				t.Errorf("%s contains forbidden product dependency token %q", path, token)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
