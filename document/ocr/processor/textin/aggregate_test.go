package textin

import (
	"strings"
	"testing"
)

func TestBuildOCRPayloadFallback(t *testing.T) {
	raw := []byte(`{
        "code":200,
        "message":"OK",
        "result":{
            "pages":[{
                "angle":0,
                "width":100,
                "height":100,
                "lines":[
                    {
                        "text":"Hello",
                        "char_candidates":[["H"]],
                        "char_candidates_score":[[0.9]],
                        "char_positions":[[0,0,10,10]]
                    },
                    {
                        "text":"Fallback",
                        "char_candidates":[],
                        "char_candidates_score":[],
                        "char_positions":[]
                    }
                ]
            }]
        }
    }`)

	payload, err := buildOCRPayload([][]byte{raw}, []string{"p1.png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.RecognizedText == "" || payload.PageTexts[0] == "" {
		t.Fatalf("expected recognized text")
	}
	if want := "Fallback"; !strings.Contains(payload.RecognizedText, want) {
		t.Fatalf("recognized text missing %q", want)
	}
	if got, want := payload.NormalizedText, "H\nFallback"; got != want {
		t.Fatalf("unexpected normalized text: got %q want %q", got, want)
	}
	if len(payload.Coordinates) != 1 {
		t.Fatalf("expected coordinate for candidate char")
	}
	if len(payload.TextBoxes) != 1 || payload.TextBoxes[0].Text != "H" || payload.TextBoxes[0].PageWidth != 100 || payload.TextBoxes[0].PageHeight != 100 {
		t.Fatalf("expected text box for recognized line, got %#v", payload.TextBoxes)
	}
}

func TestBuildOCRPayloadEscapesHTMLAndHandlesMismatchedCandidates(t *testing.T) {
	raw := []byte(`{
        "code":200,
        "message":"OK",
        "result":{
            "pages":[{
                "angle":0,
                "width":100,
                "height":100,
                "lines":[
                    {
                        "text":"A < B & C",
                        "char_candidates":[["A"]],
                        "char_candidates_score":[[0.1,0.9]],
                        "char_positions":[[0,0,10,10]]
                    }
                ]
            }]
        }
    }`)

	payload, err := buildOCRPayload([][]byte{raw}, []string{"p1.png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := payload.RecognizedText; got != "A" {
		t.Fatalf("unexpected recognized text: %q", got)
	}
	if len(payload.HTMLPages) != 1 {
		t.Fatalf("expected one html page")
	}
	if want := "A &lt; B &amp; C"; !strings.Contains(payload.HTMLPages[0], want) {
		t.Fatalf("html page missing escaped text %q: %s", want, payload.HTMLPages[0])
	}
}

func TestBuildTablePayloadFallback(t *testing.T) {
	raw := []byte(`{
        "code":200,
        "message":"OK",
        "result":{
            "pages":[{
                "angle":0,
                "width":100,
                "height":100,
                "tables":[
                    {
                        "lines":[
                            {
                                "text":"Row",
                                "char_candidates":[["R"],["o"],["w"]],
                                "char_candidates_score":[[0.9],[0.9],[0.9]],
                                "char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1]]
                            },
                            {
                                "text":"Tail",
                                "char_candidates":[],
                                "char_candidates_score":[],
                                "char_positions":[]
                            }
                        ],
                        "table_cells":[]
                    }
                ]
            }]
        }
    }`)

	payload, err := buildTablePayload([][]byte{raw}, []string{"t1.png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.Pages) != 1 {
		t.Fatalf("expected one page")
	}
	if want := "Tail"; !strings.Contains(payload.Pages[0].Recognized, want) {
		t.Fatalf("recognized text missing fallback %q", want)
	}
	if got, want := payload.NormalizedCombinedText, "Row\nTail"; got != want {
		t.Fatalf("unexpected normalized combined text: got %q want %q", got, want)
	}
}

func TestBuildTablePayloadEscapesHTMLAndFallsBackToCellText(t *testing.T) {
	raw := []byte(`{
        "code":200,
        "message":"OK",
        "result":{
            "pages":[{
                "angle":0,
                "width":100,
                "height":100,
                "tables":[
                    {
                        "lines":[
                            {
                                "text":"Header < Unsafe",
                                "char_candidates":[],
                                "char_candidates_score":[],
                                "char_positions":[]
                            }
                        ],
                        "table_rows":1,
                        "table_cols":1,
                        "table_cells":[
                            {
                                "start_row":0,
                                "start_col":0,
                                "end_row":0,
                                "end_col":0,
                                "text":"Cell <Value>",
                                "lines":[]
                            }
                        ]
                    }
                ]
            }]
        }
    }`)

	payload, err := buildTablePayload([][]byte{raw}, []string{"t1.png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := payload.Text[0]; !strings.Contains(got, "Cell <Value>") {
		t.Fatalf("expected raw text to include cell fallback, got %q", got)
	}
	if len(payload.HTMLPages) != 1 {
		t.Fatalf("expected one html page")
	}
	if want := "Header &lt; Unsafe"; !strings.Contains(payload.HTMLPages[0], want) {
		t.Fatalf("html page missing escaped header %q: %s", want, payload.HTMLPages[0])
	}
	if want := "Cell &lt;Value&gt;"; !strings.Contains(payload.HTMLPages[0], want) {
		t.Fatalf("html page missing escaped cell %q: %s", want, payload.HTMLPages[0])
	}
}

func TestMergeTextInOCRResponses(t *testing.T) {
	raw := [][]byte{
		[]byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"A"}]}]}}`),
		[]byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"B"}]}]}}`),
	}

	merged, err := mergeTextInOCRResponses(raw)
	if err != nil {
		t.Fatalf("mergeTextInOCRResponses failed: %v", err)
	}
	payload, err := buildOCRPayload([][]byte{merged}, []string{"merged.png"})
	if err != nil {
		t.Fatalf("build merged payload: %v", err)
	}
	if len(payload.PageTexts) != 2 {
		t.Fatalf("expected merged payload to keep both pages, got %d", len(payload.PageTexts))
	}
}

func TestOCRDiffSupportsMultiPageResponses(t *testing.T) {
	baseline := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Alpha","char_candidates":[["A"],["l"],["p"],["h"],["a"]],"char_candidates_score":[[1],[1],[1],[1],[1]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]},{"angle":0,"width":100,"height":100,"lines":[{"text":"Beta","char_candidates":[["B"],["e"],["t"],["a"]],"char_candidates_score":[[1],[1],[1],[1]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1]]}]}]}}`)
	raw := [][]byte{
		[]byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Alpha","char_candidates":[["A"],["l"],["p"],["h"],["a"]],"char_candidates_score":[[1],[1],[1],[1],[1]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]}]}}`),
		[]byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Beta2","char_candidates":[["B"],["e"],["t"],["a"],["2"]],"char_candidates_score":[[1],[1],[1],[1],[1]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]}]}}`),
	}

	summary, err := ocrOperation{}.Diff(raw, baseline, 2)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if summary == nil || summary.OCRDiff == nil {
		t.Fatalf("expected multi-page diff summary")
	}
	if !summary.OCRDiff.HasDiff {
		t.Fatalf("expected diff to be detected")
	}
}
