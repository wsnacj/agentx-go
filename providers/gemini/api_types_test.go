package gemini

import (
	"testing"
)

func TestResponseText(t *testing.T) {
	resp := &GenerateContentResponse{
		Candidates: []Candidate{
			{Content: Content{Parts: []Part{{Text: "hello"}, {Text: " "}, {Text: "world"}}}},
		},
	}

	if got := ResponseText(resp); got != "hello world" {
		t.Fatalf("ResponseText() = %q", got)
	}
}

func TestNewVideoPartFromURI(t *testing.T) {
	fps := float32(1.5)
	part := NewVideoPartFromURI("https://example.com/video.mp4", "video/mp4", &fps)

	if part.FileData == nil || part.FileData.FileURI != "https://example.com/video.mp4" || part.FileData.MimeType != "video/mp4" {
		t.Fatalf("FileData = %#v", part.FileData)
	}
	if part.VideoMetadata == nil || part.VideoMetadata.FPS == nil || *part.VideoMetadata.FPS != fps {
		t.Fatalf("VideoMetadata = %#v", part.VideoMetadata)
	}
}
