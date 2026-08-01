package pack

import "testing"

func TestRuntimeMetadataReadersDoNotNormalize(t *testing.T) {
	config := map[string]any{
		semanticToolNameConfigKey: " semantic ",
		artifactTypesConfigKey:    []string{" artifact.a ", "artifact.a"},
		semanticToolTagsConfigKey: []any{" tag ", "tag"},
	}

	if got := SemanticToolNameFromConfig(config); got != " semantic " {
		t.Fatalf("expected semantic tool name to remain unnormalized, got %q", got)
	}
	gotTypes := ArtifactTypesFromConfig(config)
	if len(gotTypes) != 2 || gotTypes[0] != " artifact.a " || gotTypes[1] != "artifact.a" {
		t.Fatalf("expected artifact metadata to remain unnormalized, got %#v", gotTypes)
	}
	gotTags := SemanticToolTagsFromConfig(config)
	if len(gotTags) != 2 || gotTags[0] != " tag " || gotTags[1] != "tag" {
		t.Fatalf("expected semantic tag metadata to remain unnormalized, got %#v", gotTags)
	}
}

func TestApplySemanticToolRuntimeMetadataDoesNotNormalize(t *testing.T) {
	config := map[string]any{}
	applySemanticToolRuntimeMetadata(config, SemanticTool{
		Name:          " semantic ",
		ArtifactTypes: []string{" artifact.a ", "artifact.a"},
		Tags:          []string{" tag ", "tag"},
	})

	if got := config[semanticToolNameConfigKey]; got != " semantic " {
		t.Fatalf("expected semantic tool name to remain unnormalized, got %#v", got)
	}
	gotTypes, _ := config[artifactTypesConfigKey].([]string)
	if len(gotTypes) != 2 || gotTypes[0] != " artifact.a " || gotTypes[1] != "artifact.a" {
		t.Fatalf("expected artifact metadata write to remain unnormalized, got %#v", gotTypes)
	}
	gotTags, _ := config[semanticToolTagsConfigKey].([]string)
	if len(gotTags) != 2 || gotTags[0] != " tag " || gotTags[1] != "tag" {
		t.Fatalf("expected semantic tag metadata write to remain unnormalized, got %#v", gotTags)
	}
}

func TestRuntimeMetadataNormalizeHelpersStopTrimming(t *testing.T) {
	gotTypes := NormalizeArtifactTypes([]string{" artifact.a ", "artifact.a", " artifact.a "})
	if len(gotTypes) != 2 || gotTypes[0] != " artifact.a " || gotTypes[1] != "artifact.a" {
		t.Fatalf("expected artifact normalize helper to stop trimming and exact-dedupe only, got %#v", gotTypes)
	}
	gotTags := NormalizeSemanticToolTags([]any{" tag ", "tag", " tag "})
	if len(gotTags) != 2 || gotTags[0] != " tag " || gotTags[1] != "tag" {
		t.Fatalf("expected semantic tag normalize helper to stop trimming and exact-dedupe only, got %#v", gotTags)
	}
}
