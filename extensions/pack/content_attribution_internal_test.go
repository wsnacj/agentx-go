package pack

import (
	"strings"
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestValidateDefinitionAcceptsSourceAttributedContent(t *testing.T) {
	def := contentAttributionTestDefinition()
	if err := testCoordinator(t).ValidateDefinition(def); err != nil {
		t.Fatalf("validate definition with source-attributed content: %v", err)
	}

	tmpl, ok := def.PromptTemplateByName("content_review")
	if !ok {
		t.Fatalf("expected prompt template lookup to succeed")
	}
	if tmpl.SourceAttributions[0].SourceType != SourceAttributionTypeReferenceProject || tmpl.CaseTypes[0] != "content.review" {
		t.Fatalf("unexpected prompt template metadata: %#v", tmpl)
	}
	media, ok := def.MediaArtifactByID("content_reference_media")
	if !ok {
		t.Fatalf("expected media artifact lookup to succeed")
	}
	if media.ArtifactType != "content.reference.media" || media.SourceAttributions[0].SourceID != "reference-project.content-case.001" {
		t.Fatalf("unexpected media artifact metadata: %#v", media)
	}
}

func TestValidateDefinitionRejectsMalformedSourceAttributedContent(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Definition)
		want string
	}{
		{
			name: "prompt_missing_attribution",
			edit: func(def *Definition) {
				def.PromptTemplates[0].SourceAttributions = nil
			},
			want: "prompt template \"content_review\" source_attributions is required",
		},
		{
			name: "prompt_unsupported_source_type",
			edit: func(def *Definition) {
				def.PromptTemplates[0].SourceAttributions[0].SourceType = "website"
			},
			want: "source_type \"website\" is not supported",
		},
		{
			name: "prompt_duplicate_variable",
			edit: func(def *Definition) {
				def.PromptTemplates[0].Variables = append(def.PromptTemplates[0].Variables, PromptTemplateVariable{Name: "subject"})
			},
			want: "duplicate prompt template",
		},
		{
			name: "prompt_undeclared_case_type",
			edit: func(def *Definition) {
				def.PromptTemplates[0].CaseTypes = []string{"media.unknown"}
			},
			want: "is not declared in manifest.supported_case_types",
		},
		{
			name: "media_missing_path_or_url",
			edit: func(def *Definition) {
				def.MediaArtifacts[0].Path = ""
				def.MediaArtifacts[0].URL = ""
			},
			want: "path or url is required",
		},
		{
			name: "media_undeclared_artifact_type",
			edit: func(def *Definition) {
				def.MediaArtifacts[0].ArtifactType = "media.image.unknown"
			},
			want: "references undeclared artifact type",
		},
		{
			name: "media_duplicate_source_attribution",
			edit: func(def *Definition) {
				def.MediaArtifacts[0].SourceAttributions = append(def.MediaArtifacts[0].SourceAttributions, def.MediaArtifacts[0].SourceAttributions[0])
			},
			want: "duplicate media artifact",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := contentAttributionTestDefinition()
			tt.edit(&def)
			err := testCoordinator(t).ValidateDefinition(def)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q validation error, got %v", tt.want, err)
			}
		})
	}
}

func TestContentAttributionLookupsReturnDeepCopies(t *testing.T) {
	def := contentAttributionTestDefinition()

	tmpl, ok := def.PromptTemplateByName("content_review")
	if !ok {
		t.Fatalf("expected prompt template lookup")
	}
	tmpl.Variables[0].Example = "mutated"
	tmpl.SourceAttributions[0].Title = "mutated"
	tmpl.CaseTypes[0] = "mutated"
	tmpl.Tags[0] = "mutated"
	if def.PromptTemplates[0].Variables[0].Example != "a sample content item" ||
		def.PromptTemplates[0].SourceAttributions[0].Title != "reference content examples" ||
		def.PromptTemplates[0].CaseTypes[0] != "content.review" ||
		def.PromptTemplates[0].Tags[0] != "content" {
		t.Fatalf("expected original prompt template to remain unchanged, got %#v", def.PromptTemplates[0])
	}

	media, ok := def.MediaArtifactByID("content_reference_media")
	if !ok {
		t.Fatalf("expected media artifact lookup")
	}
	media.SourceAttributions[0].Title = "mutated"
	media.CaseTypes[0] = "mutated"
	media.Tags[0] = "mutated"
	if def.MediaArtifacts[0].SourceAttributions[0].Title != "reference media output" ||
		def.MediaArtifacts[0].CaseTypes[0] != "content.review" ||
		def.MediaArtifacts[0].Tags[0] != "reference" {
		t.Fatalf("expected original media artifact to remain unchanged, got %#v", def.MediaArtifacts[0])
	}
}

func contentAttributionTestDefinition() Definition {
	return Definition{
		Manifest: Manifest{
			ID:                 "content-attribution-test-pack",
			Version:            "0.1.0",
			Domain:             "media_prompting",
			SupportedCaseTypes: []string{"content.review"},
			DefaultWorkflow:    "content_review_v1",
			ArtifactTypes:      []string{"content.reference.media"},
		},
		CaseSchemas: []CaseSchema{
			{CaseType: "content.review"},
		},
		PromptTemplates: []PromptTemplate{
			{
				Name:        "content_review",
				Description: "Review generic content inputs against pack-owned expectations.",
				Locale:      "en-US",
				Template:    "Review {{subject}} using the pack-owned content rubric.",
				Variables: []PromptTemplateVariable{
					{
						Name:        "subject",
						Description: "Primary review subject.",
						Required:    true,
						Example:     "a sample content item",
					},
				},
				SourceAttributions: []SourceAttribution{
					{
						SourceType: SourceAttributionTypeReferenceProject,
						SourceID:   "reference-project.content-patterns",
						Title:      "reference content examples",
						URL:        "https://example.invalid/reference-project/content-patterns",
						License:    "reference-only",
					},
				},
				CaseTypes: []string{"content.review"},
				Tags:      []string{"content", "prompt"},
			},
		},
		MediaArtifacts: []PackMediaArtifact{
			{
				ID:           "content_reference_media",
				ArtifactType: "content.reference.media",
				Kind:         "image",
				Description:  "Reference media output used by the pack-local content review case.",
				Path:         "examples/reference-output.png",
				MIMEType:     "image/png",
				SourceAttributions: []SourceAttribution{
					{
						SourceType: SourceAttributionTypeReferenceProject,
						SourceID:   "reference-project.content-case.001",
						Title:      "reference media output",
						URL:        "https://example.invalid/reference-project/content-case-001",
						License:    "reference-only",
					},
				},
				CaseTypes: []string{"content.review"},
				Tags:      []string{"reference", "media"},
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:        "content_review_v1",
				Pack:      "content-attribution-test-pack",
				CaseTypes: []string{"content.review"},
				EntryNode: "review_prompt",
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:     "review_prompt",
						Kind:   agentxworkflow.NodeTool,
						Config: map[string]any{"tool": "review_prompt"},
					},
				},
			},
		},
	}
}
