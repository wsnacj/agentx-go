package mediaartifact

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDescriptorJSONRoundTrip(t *testing.T) {
	hasAudio := true
	input := Descriptor{
		Source:        "browser",
		Kind:          "screenshot",
		Path:          ".agentx/browser/capture.png",
		URL:           "https://example.com",
		MIMEType:      "image/png",
		Format:        "png",
		Bytes:         128,
		Width:         1440,
		Height:        900,
		DurationMs:    2500,
		TimestampMs:   1200,
		FPS:           30,
		HasAudio:      &hasAudio,
		ScreenIndex:   1,
		CaptureScope:  "page",
		CaptureWidth:  1440,
		CaptureHeight: 900,
		Facing:        "front",
		Index:         2,
		CreatedAt:     "2026-03-14T09:30:00Z",
	}

	blob, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal descriptor: %v", err)
	}
	var got Descriptor
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("unmarshal descriptor: %v", err)
	}
	if got.Source != input.Source ||
		got.Kind != input.Kind ||
		got.Path != input.Path ||
		got.URL != input.URL ||
		got.MIMEType != input.MIMEType ||
		got.Format != input.Format ||
		got.Bytes != input.Bytes ||
		got.Width != input.Width ||
		got.Height != input.Height ||
		got.DurationMs != input.DurationMs ||
		got.TimestampMs != input.TimestampMs ||
		got.FPS != input.FPS ||
		got.ScreenIndex != input.ScreenIndex ||
		got.CaptureScope != input.CaptureScope ||
		got.CaptureWidth != input.CaptureWidth ||
		got.CaptureHeight != input.CaptureHeight ||
		got.Facing != input.Facing ||
		got.Index != input.Index ||
		got.CreatedAt != input.CreatedAt {
		t.Fatalf("expected round-trip descriptor %#v, got %#v", input, got)
	}
	if got.HasAudio == nil || input.HasAudio == nil || *got.HasAudio != *input.HasAudio {
		t.Fatalf("expected round-trip has_audio=%v, got %#v", *input.HasAudio, got.HasAudio)
	}
}

func TestDescriptorJSONOmitsZeroValueOptionalFields(t *testing.T) {
	blob, err := json.Marshal(Descriptor{
		Source: "pdf",
		Kind:   "page",
	})
	if err != nil {
		t.Fatalf("marshal descriptor: %v", err)
	}
	text := string(blob)
	if !strings.Contains(text, `"source":"pdf"`) || !strings.Contains(text, `"kind":"page"`) {
		t.Fatalf("expected required fields to be present, got %s", text)
	}
	for _, forbidden := range []string{
		`"path"`,
		`"url"`,
		`"mime_type"`,
		`"duration_ms"`,
		`"timestamp_ms"`,
		`"has_audio"`,
		`"capture_scope"`,
		`"created_at"`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("expected zero-value optional field %s to be omitted, got %s", forbidden, text)
		}
	}
}
