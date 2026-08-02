package docparse

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	docparsehostkit "github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
)

func TestImplicitAssetDiscoveryFailsClosed(t *testing.T) {
	if _, err := LocateAssets(); err == nil {
		t.Fatal("LocateAssets should reject implicit source-checkout discovery")
	}
	if _, err := DomainRoot(); err == nil {
		t.Fatal("DomainRoot should reject implicit source-checkout discovery")
	}
}

func TestLocateAssetsAt(t *testing.T) {
	assets, err := LocateAssetsAt(".")
	if err != nil {
		t.Fatalf("LocateAssetsAt returned error: %v", err)
	}
	for name, path := range map[string]string{
		"domain": assets.DomainRoot,
		"skills": assets.SkillsRoot,
		"tools":  assets.ToolsRoot,
	} {
		if path == "" {
			t.Fatalf("%s path is empty", name)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s path %q is not accessible: %v", name, path, err)
		}
	}
	if got := assets.ExtensionRoot(); got != assets.DomainRoot {
		t.Fatalf("ExtensionRoot = %q, want %q", got, assets.DomainRoot)
	}
	if got := assets.SkillPath("document-operations"); got != filepath.Join(assets.SkillsRoot, "document-operations", "SKILL.md") {
		t.Fatalf("SkillPath mismatch: %q", got)
	}
	if got := assets.ToolManifestPath("docparse_extract_fields"); got != filepath.Join(assets.ToolsRoot, "docparse_extract_fields.tool.json") {
		t.Fatalf("ToolManifestPath mismatch: %q", got)
	}
	if got := assets.ToolManifestPath("docparse_profile_probe"); got != filepath.Join(assets.ToolsRoot, "docparse_profile_probe.tool.json") {
		t.Fatalf("ToolManifestPath mismatch: %q", got)
	}
}

func TestEmbeddedAssetsExposeSkillAndToolSurface(t *testing.T) {
	source := ExtensionFS()
	if _, err := fs.ReadFile(source, path.Join("skills", "document-operations", "SKILL.md")); err != nil {
		t.Fatalf("read embedded document-operations skill: %v", err)
	}
	for _, name := range docparsehostkit.ToolNames() {
		if _, err := fs.ReadFile(source, path.Join("tools", name+".tool.json")); err != nil {
			t.Fatalf("read embedded tool manifest %q: %v", name, err)
		}
	}
}

func TestExtractionToolManifestsAcceptExistingParseResults(t *testing.T) {
	assets, err := LocateAssetsAt(".")
	if err != nil {
		t.Fatalf("LocateAssetsAt returned error: %v", err)
	}
	for _, name := range []string{"docparse_extract_fields", "docparse_extract_table"} {
		blob, err := os.ReadFile(assets.ToolManifestPath(name))
		if err != nil {
			t.Fatalf("read %s manifest: %v", name, err)
		}
		var manifest struct {
			Parameters struct {
				Properties map[string]any `json:"properties"`
				Required   []string       `json:"required"`
			} `json:"parameters"`
		}
		if err := json.Unmarshal(blob, &manifest); err != nil {
			t.Fatalf("decode %s manifest: %v", name, err)
		}
		if _, ok := manifest.Parameters.Properties["parse_result"]; !ok {
			t.Fatalf("%s manifest should expose parse_result for local evidence projection", name)
		}
		for _, required := range manifest.Parameters.Required {
			if required == "document_path" {
				t.Fatalf("%s manifest should not force document_path when result_path/parse_result is supported", name)
			}
		}
	}
}

func TestToolManifestsDescribeParameters(t *testing.T) {
	assets, err := LocateAssetsAt(".")
	if err != nil {
		t.Fatalf("LocateAssetsAt returned error: %v", err)
	}
	entries, err := os.ReadDir(assets.ToolsRoot)
	if err != nil {
		t.Fatalf("read tools root: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(assets.ToolsRoot, entry.Name())
		blob, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var manifest map[string]any
		if err := json.Unmarshal(blob, &manifest); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if strings.TrimSpace(readString(manifest["description"])) == "" {
			t.Fatalf("%s missing top-level description", path)
		}
		params, _ := manifest["parameters"].(map[string]any)
		if params == nil {
			t.Fatalf("%s missing parameters", path)
		}
		assertSchemaDescriptions(t, path, params, "")
	}
}

func assertSchemaDescriptions(t *testing.T, owner string, schema map[string]any, prefix string) {
	t.Helper()
	props, _ := schema["properties"].(map[string]any)
	for name, raw := range props {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s property %s should be an object schema", owner, path)
		}
		if strings.TrimSpace(readString(prop["description"])) == "" {
			t.Fatalf("%s property %s missing description", owner, path)
		}
		if nested, _ := prop["properties"].(map[string]any); len(nested) != 0 {
			assertSchemaDescriptions(t, owner, prop, path)
		}
		if items, _ := prop["items"].(map[string]any); items != nil {
			if nested, _ := items["properties"].(map[string]any); len(nested) != 0 {
				assertSchemaDescriptions(t, owner, items, path+"[]")
			}
		}
	}
}

func readString(value any) string {
	out, _ := value.(string)
	return out
}

func assertSameStrings(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings length mismatch: got=%#v want=%#v", got, want)
	}
	gotSet := make(map[string]bool, len(got))
	for _, value := range got {
		gotSet[value] = true
	}
	for _, value := range want {
		if !gotSet[value] {
			t.Fatalf("strings mismatch: got=%#v want=%#v", got, want)
		}
	}
}
