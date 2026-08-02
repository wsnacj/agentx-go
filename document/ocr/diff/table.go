package diff

import (
	"sync"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type TablePageDiff struct {
	PageNumber int        `json:"page_number"`
	DiffAreas  []LineDiff `json:"diff_areas"`
}

type TableDiffResult struct {
	Valid         bool            `json:"valid"`
	HasDiff       bool            `json:"has_diff"`
	PageDiffs     []TablePageDiff `json:"page_diffs"`
	ExtraPages    []ExtraPage     `json:"extra_pages"`
	MappingScheme string          `json:"mapping_scheme"`
}

func ExtractTableCharsFromPage(page TPage) []OCRChar {
	var chars []OCRChar
	for _, tbl := range page.Tables {
		for _, line := range tbl.Lines {
			runes := []rune(line.Text)
			for i, r := range runes {
				var pos []int
				if i < len(line.CharPositions) && len(line.CharPositions[i]) >= 4 {
					pos = line.CharPositions[i]
				} else {
					pos = []int{0, 0, 0, 0}
				}
				chars = append(chars, OCRChar{Char: string(r), Position: pos})
			}
		}
		for _, cell := range tbl.TableCells {
			for _, line := range cell.Lines {
				runes := []rune(line.Text)
				for i, r := range runes {
					var pos []int
					if i < len(line.CharPositions) && len(line.CharPositions[i]) >= 4 {
						pos = line.CharPositions[i]
					} else {
						pos = []int{0, 0, 0, 0}
					}
					chars = append(chars, OCRChar{Char: string(r), Position: pos})
				}
			}
		}
	}
	return chars
}

func NormalizeTablePage(page TPage) (string, []int, []OCRChar) {
	chars := ExtractTableCharsFromPage(page)
	norm, mapping := normalizeOCRChars(chars, true)
	return norm, mapping, chars
}

func PageDiffScoreOrigTable(a, b TPage) int {
	s1, _, _ := NormalizeTablePage(a)
	s2, _, _ := NormalizeTablePage(b)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(s1, s2, false)
	score := 0
	for _, d := range diffs {
		if d.Type != diffmatchpatch.DiffEqual {
			score += len([]rune(d.Text))
		}
	}
	return score
}

func matchPagesForDiffTable(orig, ret TableResponse) (map[int]int, []int, string) {
	return matchPages(len(orig.Result.Pages), len(ret.Result.Pages), func(i, j int) int {
		return PageDiffScoreOrigTable(orig.Result.Pages[i], ret.Result.Pages[j])
	})
}

func CompareTableResponses(orig, ret TableResponse) TableDiffResult {
	result := TableDiffResult{Valid: true}
	origPages := orig.Result.Pages
	retPages := ret.Result.Pages
	if len(retPages) < len(origPages) {
		result.Valid = false
		result.HasDiff = true
		return result
	}
	pageMap, extraIdx, scheme := matchPagesForDiffTable(orig, ret)
	result.MappingScheme = scheme
	var wg sync.WaitGroup
	pageDiffs := make([]TablePageDiff, len(origPages))
	for origIdx, retIdx := range pageMap {
		wg.Add(1)
		go func(oi, ri int) {
			defer wg.Done()
			nOrig, mapOrig, charsOrig := NormalizeTablePage(origPages[oi])
			nRet, mapRet, charsRet := NormalizeTablePage(retPages[ri])
			segs := computeDiffSegments(nOrig, nRet)
			pd := TablePageDiff{PageNumber: oi + 1}
			for _, seg := range segs {
				var mapping []int
				var chars []OCRChar
				if seg.op == "insert" {
					mapping, chars = mapRet, charsRet
				} else {
					mapping, chars = mapOrig, charsOrig
				}
				x1, y1, x2, y2, ok := getBoundingBoxForSegment(seg, mapping, chars)
				diff := LineDiff{Op: seg.op, DiffText: seg.diffText}
				if ok {
					diff.BoundingBox = []float64{x1, y1, x2, y2}
				}
				pd.DiffAreas = append(pd.DiffAreas, diff)
			}
			pageDiffs[oi] = pd
		}(origIdx, retIdx)
	}
	wg.Wait()
	result.PageDiffs = pageDiffs
	for _, pd := range pageDiffs {
		if len(pd.DiffAreas) > 0 {
			result.HasDiff = true
			break
		}
	}
	if !result.HasDiff && len(extraIdx) > 0 {
		result.HasDiff = true
	}
	for _, idx := range extraIdx {
		text, _, _ := NormalizeTablePage(retPages[idx])
		result.ExtraPages = append(result.ExtraPages, ExtraPage{PageNumber: idx + 1, Text: text})
	}
	return result
}

func CompareTableJSON(orig, ret []byte) (TableDiffResult, error) {
	origResp, err := ParseTableResponse(orig)
	if err != nil {
		return TableDiffResult{}, err
	}
	retResp, err := ParseTableResponse(ret)
	if err != nil {
		return TableDiffResult{}, err
	}
	return CompareTableResponses(*origResp, *retResp), nil
}
