package baidu

import (
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/document/ocr/model"
)

func TestBuildPayload(t *testing.T) {
	resp := baiduResponse{WordsResult: []struct {
		Words    string        `json:"words"`
		Location locationBlock `json:"location"`
	}{
		{Words: "你好", Location: locationBlock{Left: 1, Top: 2, Width: 3, Height: 4}},
		{Words: "世界", Location: locationBlock{Left: 5, Top: 6, Width: 7, Height: 8}},
	}}
	data, _ := json.Marshal(resp)
	op := ocrOperation{}
	payloadAny, err := op.Build([][]byte{data}, []string{"sample.png"})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	payload, ok := payloadAny.(model.OCRPayload)
	if !ok {
		t.Fatalf("unexpected payload type %T", payloadAny)
	}
	if payload.RecognizedText != "你好\n世界" {
		t.Fatalf("recognized text mismatch: %q", payload.RecognizedText)
	}
	if len(payload.Coordinates) != 2 {
		t.Fatalf("coordinates mismatch: %v", payload.Coordinates)
	}
	if len(payload.TextBoxes) != 2 || payload.TextBoxes[0].Text != "你好" || payload.TextBoxes[0].Page != 1 {
		t.Fatalf("text boxes mismatch: %#v", payload.TextBoxes)
	}
	if len(payload.HTMLPages) != 1 {
		t.Fatalf("expected 1 html page")
	}
}

func TestDiffNoChange(t *testing.T) {
	resp := baiduResponse{WordsResult: []struct {
		Words    string        `json:"words"`
		Location locationBlock `json:"location"`
	}{{Words: "hello"}}}
	data, _ := json.Marshal(resp)
	op := ocrOperation{}
	summary, err := op.Diff([][]byte{data}, data, 2)
	if err != nil {
		t.Fatalf("diff err: %v", err)
	}
	if summary != nil {
		t.Fatalf("expected nil summary")
	}
}

func TestTableBuild(t *testing.T) {
	resp := baiduTableResponse{
		TablesResult: []tableResult{
			{
				Body: []tableCell{
					{RowStart: 0, RowEnd: 1, ColStart: 0, ColEnd: 1, Words: "A"},
					{RowStart: 0, RowEnd: 1, ColStart: 1, ColEnd: 2, Words: "B"},
				},
				TableLocation: []point{{X: 0, Y: 0}, {X: 10, Y: 10}},
			},
		},
	}
	data, _ := json.Marshal(resp)
	payloadAny, err := tableOperation{}.Build([][]byte{data}, []string{"table.png"})
	if err != nil {
		t.Fatalf("table build err: %v", err)
	}
	payload, ok := payloadAny.(model.TablePayload)
	if !ok {
		t.Fatalf("unexpected payload type %T", payloadAny)
	}
	if len(payload.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(payload.Pages))
	}
	if len(payload.HTMLPages) != 1 {
		t.Fatalf("expected 1 html page, got %d", len(payload.HTMLPages))
	}
	if len(payload.Text) != 1 {
		t.Fatalf("expected 1 text entry, got %d", len(payload.Text))
	}
}
