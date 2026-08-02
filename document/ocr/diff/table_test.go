package diff

import "testing"

func TestCompareTableJSON(t *testing.T) {
	orig := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"tables":[{"lines":[{"text":"A","char_candidates":[["A"]],"char_candidates_score":[[0.9]],"char_positions":[[0,0,1,1]]}]}]}]}}`)
	ret := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"tables":[{"lines":[{"text":"AB","char_candidates":[["A"],["B"]],"char_candidates_score":[[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1]]}]}]}]}}`)
	res, err := CompareTableJSON(orig, ret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasDiff {
		t.Fatalf("expected diff")
	}
}
