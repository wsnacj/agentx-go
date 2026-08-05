package plugin_test

import (
	"errors"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/plugin"
)

func TestParseNormalizesPortableManifest(t *testing.T) {
	manifest, err := plugin.Parse([]byte(`{
		"name":"Research-Kit",
		"schema_version":"v1",
		"version":"0.1.0",
		"trust_boundary":"reviewed",
		"roots":[".","."],
		"entrypoints":{"skills":"skills","tools":"tools"},
		"dependencies":[{"kind":"connector","id":"Public-Web"}],
		"requested_permissions":[{"capability":"network-read","reason":"fetch public sources"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "research-kit" || manifest.TrustBoundary != plugin.TrustBoundaryReviewed {
		t.Fatalf("manifest=%#v", manifest)
	}
	if len(manifest.Dependencies) != 1 || manifest.Dependencies[0].ID != "public-web" {
		t.Fatalf("dependencies=%#v", manifest.Dependencies)
	}
	if len(manifest.RequestedPermissions) != 1 || manifest.RequestedPermissions[0].Capability != "network-read" {
		t.Fatalf("permissions=%#v", manifest.RequestedPermissions)
	}
}

func TestParseRejectsPolicyCredentialsAndEscapingPaths(t *testing.T) {
	for name, payload := range map[string]string{
		"policy":     `{"name":"bad","policy":"allow"}`,
		"credential": `{"name":"bad","credentials":{"token":"secret"}}`,
		"nested":     `{"name":"bad","metadata":{"approval":"always"}}`,
		"path":       `{"name":"bad","entrypoints":{"skills":"../skills"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := plugin.Parse([]byte(payload))
			if err == nil {
				t.Fatal("expected error")
			}
			if name == "path" && !errors.Is(err, &plugin.Error{Code: plugin.ErrorCodeInvalidPath}) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestNormalizeDoesNotMutateCaller(t *testing.T) {
	raw := plugin.Manifest{Name: "sample", Roots: []string{"z", "a", "z"}}
	manifest, err := plugin.Normalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw.Roots) != 3 || raw.Roots[0] != "z" {
		t.Fatalf("caller mutated: %#v", raw.Roots)
	}
	if len(manifest.Roots) != 2 || manifest.Roots[0] != "a" || manifest.Roots[1] != "z" {
		t.Fatalf("roots=%#v", manifest.Roots)
	}
}
