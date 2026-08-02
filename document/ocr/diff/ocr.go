package diff

import (
	"sync"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/wsnacj/agentx-go/document/ocr/internal/fuzzy"
)

type LineDiff struct {
	Op          string    `json:"op"`
	DiffText    string    `json:"diff_text"`
	BoundingBox []float64 `json:"bounding_box,omitempty"`
}

type ExtraPage struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
}

type PageDiff struct {
	PageNumber int        `json:"page_number"`
	DiffAreas  []LineDiff `json:"diff_areas"`
}

type DiffResult struct {
	Valid         bool        `json:"valid"`
	HasDiff       bool        `json:"has_diff"`
	PageDiffs     []PageDiff  `json:"page_diffs"`
	ExtraPages    []ExtraPage `json:"extra_pages"`
	MappingScheme string      `json:"mapping_scheme"`
}

func ExtractOCRCharsFromPage(page OCRPage) []OCRChar {
	var chars []OCRChar
	for _, line := range page.Lines {
		runes := []rune(line.Text)
		count := len(runes)
		if len(line.CharPositions) < count {
			count = len(line.CharPositions)
		}
		for i := 0; i < count; i++ {
			pos := line.CharPositions[i]
			if len(pos) < 4 {
				continue
			}
			chars = append(chars, OCRChar{Char: string(runes[i]), Position: pos})
		}
	}
	return chars
}

func NormalizeOCRPage(page OCRPage) (string, []int, []OCRChar) {
	chars := ExtractOCRCharsFromPage(page)
	normText, mapping := normalizeOCRChars(chars, true)
	return normText, mapping, chars
}

func PageDiffScore(a, b OCRPage) int {
	s1, _, _ := NormalizeOCRPage(a)
	s2, _, _ := NormalizeOCRPage(b)
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

func matchPagesForDiffOCR(orig, ret OCRResponse) (map[int]int, []int, string) {
	return matchPages(len(orig.Result.Pages), len(ret.Result.Pages), func(i, j int) int {
		return PageDiffScore(orig.Result.Pages[i], ret.Result.Pages[j])
	})
}

func CompareOCRResponses(orig, ret OCRResponse) DiffResult {
	result := DiffResult{Valid: true}
	origPages := orig.Result.Pages
	retPages := ret.Result.Pages
	if len(retPages) < len(origPages) {
		result.Valid = false
		result.HasDiff = true
		return result
	}
	mapping, extra, scheme := matchPagesForDiffOCR(orig, ret)
	result.MappingScheme = scheme
	var wg sync.WaitGroup
	pageDiffs := make([]PageDiff, len(origPages))
	for origIdx, retIdx := range mapping {
		wg.Add(1)
		go func(oi, ri int) {
			defer wg.Done()
			nOrig, mapOrig, charsOrig := NormalizeOCRPage(origPages[oi])
			nRet, mapRet, charsRet := NormalizeOCRPage(retPages[ri])
			segs := computeDiffSegments(nOrig, nRet)
			pd := PageDiff{PageNumber: oi + 1}
			for _, seg := range segs {
				var m []int
				var ch []OCRChar
				if seg.op == "insert" {
					m, ch = mapRet, charsRet
				} else {
					m, ch = mapOrig, charsOrig
				}
				x1, y1, x2, y2, ok := getBoundingBoxForSegment(seg, m, ch)
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
	if !result.HasDiff && len(extra) > 0 {
		result.HasDiff = true
	}
	for _, idx := range extra {
		text, _, _ := NormalizeOCRPage(retPages[idx])
		result.ExtraPages = append(result.ExtraPages, ExtraPage{PageNumber: idx + 1, Text: text})
	}
	return result
}

func CompareOCRJSON(orig, ret []byte) (DiffResult, error) {
	origResp, err := ParseOCRResponse(orig)
	if err != nil {
		return DiffResult{}, err
	}
	retResp, err := ParseOCRResponse(ret)
	if err != nil {
		return DiffResult{}, err
	}
	return CompareOCRResponses(*origResp, *retResp), nil
}

// FuzzyLocateText 在识别文本中查找与目标文本相近的位置。
func FuzzyLocateText(recognizedText, targetText string, segLength, margin, thresholdStart, thresholdEnd, overallThreshold, boundaryTolerance int, removePunct bool) []fuzzy.Candidate {
	return fuzzy.FuzzyFindAllTargetBoundaries(recognizedText, targetText, segLength, margin, thresholdStart, thresholdEnd, overallThreshold, boundaryTolerance, removePunct)
}
