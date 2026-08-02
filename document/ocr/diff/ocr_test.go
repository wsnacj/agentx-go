package diff

import "testing"

func TestCompareOCRJSON(t *testing.T) {
	orig := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello","char_candidates":[["H"],["e"],["l"],["l"],["o"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]}]}}`)
	ret := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello world","char_candidates":[["H"],["e"],["l"],["l"],["o"],[" "],["w"],["o"],["r"],["l"],["d"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9],[1],[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1],[5,0,6,1],[6,0,7,1],[7,0,8,1],[8,0,9,1],[9,0,10,1],[10,0,11,1]]}]}]}}`)
	res, err := CompareOCRJSON(orig, ret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasDiff {
		t.Fatalf("expected diff")
	}
	if len(res.PageDiffs) == 0 {
		t.Fatalf("expected page diffs")
	}
}
