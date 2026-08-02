package publicnews

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewsbriefKitDoesNotImportProjectOrConcreteNetworkAdapters(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
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
		text := string(blob)
		for _, forbidden := range []string{
			"core/agentx/examples/project-integration",
			"core/agentx/network/retrieval",
			"core/agentx/browserruntime",
			"github.com/wsnacj/agentx-go/browser/runtime",
			"scene/agentx_public_news/adapters",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("newsbrief kit must stay host/source neutral; found %q in %s", forbidden, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
