package llm

import "testing"

func TestNewImageURL(t *testing.T) {
	vc := NewImageURL("http://example.com/img.png")
	if vc.Type != "image_url" || vc.ImageURL != "http://example.com/img.png" || vc.Detail != DetailAuto {
		t.Fatalf("unexpected visual content: %+v", vc)
	}

	vc = NewImageURL("http://example.com/img.png", WithDetail(DetailHigh), WithLabels("poster"))
	if vc.Detail != DetailHigh {
		t.Fatalf("detail override failed: %v", vc.Detail)
	}
	if len(vc.Labels) != 1 || vc.Labels[0] != "poster" {
		t.Fatalf("labels not applied: %+v", vc.Labels)
	}
}

func TestNewLocalImage(t *testing.T) {
	vc := NewLocalImage("/tmp/a.png")
	if vc.ImageURL != "/tmp/a.png" || vc.Detail != DetailAuto {
		t.Fatalf("unexpected local visual content: %+v", vc)
	}
}

func TestNewTextBlock(t *testing.T) {
	vc := NewTextBlock("hello", WithLabels("caption"))
	if vc.Type != "text" || vc.Text != "hello" {
		t.Fatalf("unexpected text visual content: %+v", vc)
	}
	if len(vc.Labels) != 1 || vc.Labels[0] != "caption" {
		t.Fatalf("labels not applied to text content")
	}
}

func TestNewImageList(t *testing.T) {
	visuals := NewImageList([]string{"a.png", "b.png"}, WithDetail(DetailHigh))
	if len(visuals) != 2 {
		t.Fatalf("unexpected length: %d", len(visuals))
	}
	for _, v := range visuals {
		if v.Detail != DetailHigh {
			t.Fatalf("detail not applied: %+v", v)
		}
	}
}
