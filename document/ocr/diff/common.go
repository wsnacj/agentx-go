package diff

import (
	"math"
	"strings"
	"unicode"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type OCRChar struct {
	Char     string
	Position []int
}

type diffSegment struct {
	start    int
	end      int
	diffText string
	op       string
}

func normalizeOCRChars(chars []OCRChar, removePunct bool) (string, []int) {
	var b strings.Builder
	mapping := make([]int, 0, len(chars))
	isPunct := func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。、．！？(),.!?（）", r)
	}
	for idx, oc := range chars {
		runes := []rune(oc.Char)
		if len(runes) == 0 {
			continue
		}
		ch := runes[0]
		if removePunct && isPunct(ch) {
			continue
		}
		b.WriteRune(ch)
		mapping = append(mapping, idx)
	}
	return b.String(), mapping
}

func computeDiffSegments(origText, retText string) []diffSegment {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(origText, retText, false)
	var segs []diffSegment
	var oi, ri int
	for _, d := range diffs {
		length := len([]rune(d.Text))
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			oi += length
			ri += length
		case diffmatchpatch.DiffInsert:
			segs = append(segs, diffSegment{start: ri, end: ri + length, diffText: d.Text, op: "insert"})
			ri += length
		case diffmatchpatch.DiffDelete:
			segs = append(segs, diffSegment{start: oi, end: oi + length, diffText: d.Text, op: "delete"})
			oi += length
		}
	}
	return segs
}

func getBoundingBoxForSegment(seg diffSegment, mapping []int, chars []OCRChar) (float64, float64, float64, float64, bool) {
	if seg.start < 0 || seg.end > len(mapping) || seg.start >= seg.end {
		return 0, 0, 0, 0, false
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for idx := seg.start; idx < seg.end; idx++ {
		ci := mapping[idx]
		if ci < 0 || ci >= len(chars) {
			continue
		}
		pos := chars[ci].Position
		if len(pos) < 4 {
			continue
		}
		x, y, w, h := float64(pos[0]), float64(pos[1]), float64(pos[2]), float64(pos[3])
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+w > maxX {
			maxX = x + w
		}
		if y+h > maxY {
			maxY = y + h
		}
	}
	if minX == math.MaxFloat64 {
		return 0, 0, 0, 0, false
	}
	return minX, minY, maxX, maxY, true
}

func matchPages(nOrig, nRet int, score func(i, j int) int) (map[int]int, []int, string) {
	mapping := make(map[int]int, nOrig)
	if nRet < nOrig {
		return mapping, nil, ""
	}
	delta := nRet - nOrig
	scoreTail := score(0, 0)
	scoreHead := scoreTail
	if delta > 0 {
		scoreHead = score(0, delta)
	}
	var extra []int
	if delta > 0 && scoreHead < scoreTail {
		for i := 0; i < nOrig; i++ {
			mapping[i] = i + delta
		}
		for i := 0; i < delta; i++ {
			extra = append(extra, i)
		}
		return mapping, extra, "head"
	}
	for i := 0; i < nOrig; i++ {
		mapping[i] = i
	}
	for i := nOrig; i < nRet; i++ {
		extra = append(extra, i)
	}
	return mapping, extra, "tail"
}
