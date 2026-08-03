package mediaartifact_test

import (
	"reflect"
	"strings"
	"testing"

	mediaartifact "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

func TestRefsFromValueNormalizesNestedToolArtifacts(t *testing.T) {
	input := map[string]any{
		"tool":  " Browser_Runtime ",
		"title": "Quarterly report",
		"media": []any{
			map[string]any{
				"path":          " captures/page.png ",
				"kind":          "screenshot",
				"capture_scope": "page",
				"mode":          "ui",
			},
			map[string]any{
				"frame_path":  " frames/0001.jpg ",
				"source_tool": "video_frames",
				"strategy":    "interval",
			},
		},
	}

	got, err := mediaartifact.RefsFromValue(input)
	if err != nil {
		t.Fatalf("RefsFromValue(): %v", err)
	}
	want := []mediaartifact.Ref{
		{
			Raw:            "captures/page.png",
			Display:        "captures/page.png",
			Labels:         []string{"browser_runtime", "Quarterly report", "screenshot", "page"},
			ModeHint:       "screenshot",
			ArtifactSource: "browser",
			ArtifactKind:   "screenshot",
		},
		{
			Raw:            "frames/0001.jpg",
			Display:        "frames/0001.jpg",
			Labels:         []string{"browser_runtime", "Quarterly report", "video_frames", "interval"},
			ArtifactSource: "browser",
			ArtifactKind:   "frame",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RefsFromValue() = %#v, want %#v", got, want)
	}
	if got[0].Labels[0] == "" || input["tool"] != " Browser_Runtime " {
		t.Fatalf("expected a copied result without mutating input: got=%#v input=%#v", got, input)
	}
}

func TestRefsFromValueAcceptsJSONStringAndPreservesOrdering(t *testing.T) {
	got, err := mediaartifact.RefsFromValue(`{
		"tool":"pdf_analyze",
		"rendered_pages":[
			{"path":"page-1.png","kind":"rendered_page"},
			{"path":"page-2.png","kind":"rendered_page"}
		],
		"path":"report.pdf"
	}`)
	if err != nil {
		t.Fatalf("RefsFromValue(): %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("RefsFromValue() count = %d, want 3: %#v", len(got), got)
	}
	if got[0].Raw != "page-1.png" || got[1].Raw != "page-2.png" || got[2].Raw != "report.pdf" {
		t.Fatalf("RefsFromValue() ordering = %#v", got)
	}
	for _, ref := range got {
		if ref.ArtifactSource != "pdf" || ref.ModeHint != "document" {
			t.Fatalf("expected inherited PDF context, got %#v", ref)
		}
	}
}

func TestRefsFromValueKeepsMalformedJSONLookingStringAsRawReference(t *testing.T) {
	got, err := mediaartifact.RefsFromValue(`{"path":`)
	if err != nil {
		t.Fatalf("RefsFromValue(): %v", err)
	}
	want := []mediaartifact.Ref{{Raw: `{"path":`}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RefsFromValue() = %#v, want %#v", got, want)
	}
}

func TestRefsFromValueRejectsUnsupportedTopLevelType(t *testing.T) {
	_, err := mediaartifact.RefsFromValue(42)
	if err == nil || !strings.Contains(err.Error(), "image_analyze: unsupported artifact type int") {
		t.Fatalf("RefsFromValue() error = %v", err)
	}
}

func TestRefsFromValueReturnsDetachedLabels(t *testing.T) {
	input := map[string]any{
		"path": "capture.png",
		"tool": "browser_screenshot",
	}
	first, err := mediaartifact.RefsFromValue(input)
	if err != nil {
		t.Fatalf("RefsFromValue(): %v", err)
	}
	first[0].Labels[0] = "changed"
	second, err := mediaartifact.RefsFromValue(input)
	if err != nil {
		t.Fatalf("RefsFromValue() second call: %v", err)
	}
	if second[0].Labels[0] != "browser_screenshot" {
		t.Fatalf("expected detached label storage, got %#v", second)
	}
}
