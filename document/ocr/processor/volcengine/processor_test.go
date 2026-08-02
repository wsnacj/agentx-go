package volcengine

import (
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/document/ocr/model"
)

func TestOCROperationBuild(t *testing.T) {
	data := buildResponse(t, []string{"你好", "世界"}, [][]volcChar{
		{
			{X: 10, Y: 20, Width: 5, Height: 10, Score: 0.9, Char: "你"},
			{X: 20, Y: 20, Width: 5, Height: 10, Score: 0.9, Char: "好"},
		},
	})

	payloadAny, err := ocrOperation{}.Build([][]byte{data}, []string{"sample.png"})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	payload, ok := payloadAny.(model.OCRPayload)
	if !ok {
		t.Fatalf("unexpected payload type %T", payloadAny)
	}

	expectedText := "你好\n世界"
	if payload.RecognizedText != expectedText {
		t.Fatalf("recognized text mismatch: got %q want %q", payload.RecognizedText, expectedText)
	}
	if len(payload.PageTexts) != 1 || payload.PageTexts[0] != expectedText {
		t.Fatalf("page texts mismatch: %#v", payload.PageTexts)
	}
	if len(payload.HTMLPages) != 1 {
		t.Fatalf("expected 1 html page, got %d", len(payload.HTMLPages))
	}
	if len(payload.Coordinates) != 2 {
		t.Fatalf("expected 2 coordinates from chars, got %d", len(payload.Coordinates))
	}
	if len(payload.TextBoxes) != 1 || payload.TextBoxes[0].Text != "你好" || payload.TextBoxes[0].Confidence <= 0 {
		t.Fatalf("expected line text box from chars, got %#v", payload.TextBoxes)
	}
}

func TestNormalizeCoordForRelativeValue(t *testing.T) {
	got := normalizeCoord(0.5)
	if got != 5000 {
		t.Fatalf("expected 0.5 to scale to 5000, got %d", got)
	}
}

func TestOCROperationDiff(t *testing.T) {
	baseline := buildResponse(t, []string{"你好", "世界"}, nil)
	current := buildResponse(t, []string{"你好", "宇宙"}, nil)

	summary, err := ocrOperation{}.Diff([][]byte{current}, baseline, 2)
	if err != nil {
		t.Fatalf("diff computation failed: %v", err)
	}
	if summary == nil {
		t.Fatalf("expected diff summary, got nil")
	}
	if summary.OCRDiff == nil || !summary.OCRDiff.HasDiff {
		t.Fatalf("expected has diff, summary=%+v", summary)
	}
	if len(summary.OCRDiff.Preview) == 0 {
		t.Fatalf("expected preview entries")
	}
	if len(summary.Notes) == 0 {
		t.Fatalf("expected fuzzy notes to be populated")
	}
}

func TestOCROperationDiffNoChange(t *testing.T) {
	raw := buildResponse(t, []string{"平安", "喜乐"}, nil)
	summary, err := ocrOperation{}.Diff([][]byte{raw}, raw, 3)
	if err != nil {
		t.Fatalf("diff no change failed: %v", err)
	}
	if summary != nil {
		t.Fatalf("expected nil summary when no diff, got %+v", summary)
	}
}

func buildResponse(t *testing.T, texts []string, chars [][]volcChar) []byte {
	t.Helper()
	resp := apiResponse{
		Code:    10000,
		Message: "Success",
		Data: respData{
			LineTexts: append([]string(nil), texts...),
			LineRects: []volcRect{
				{X: 10, Y: 20, Width: 30, Height: 40},
			},
			Chars: chars,
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	return data
}
