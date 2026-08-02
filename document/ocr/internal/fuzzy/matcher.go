package fuzzy

import (
	"sort"
	"strings"
	"unicode"

	"github.com/wsnacj/agentx-go/document/ocr/internal/ratio"
)

type Candidate struct {
	OrigStart     int
	OrigEnd       int
	MatchedSubstr string
	StartScore    int
	EndScore      int
	OverallScore  int
	NormCandidate string
}

func normalizeText(text string, removePunct bool) (string, []int) {
	runes := []rune(text)
	var builder strings.Builder
	mapping := make([]int, 0, len(runes))
	builder.Grow(len(text))
	filter := func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。、．！？(),.!?（）", r)
	}
	for i, r := range runes {
		if removePunct && filter(r) {
			continue
		}
		builder.WriteRune(r)
		mapping = append(mapping, i)
	}
	return builder.String(), mapping
}

type segmentMatch struct {
	start int
	end   int
	score int
}

func optimizedFuzzyFindSegments(normText []rune, segment string, margin, threshold int) []segmentMatch {
	var matches []segmentMatch
	segmentRunes := []rune(segment)
	segLen := len(segmentRunes)
	n := len(normText)
	if n == 0 || segLen == 0 {
		return matches
	}
	minWin := max(1, segLen-margin)
	maxWin := min(n, segLen+margin)
	var builder strings.Builder
	for winLen := minWin; winLen <= maxWin; winLen++ {
		for i := 0; i <= n-winLen; i++ {
			builder.Reset()
			builder.Grow(winLen * 4)
			for j := i; j < i+winLen; j++ {
				builder.WriteRune(normText[j])
			}
			candidate := builder.String()
			score := rapidfuzz.Ratio(candidate, segment)
			if score >= threshold {
				matches = append(matches, segmentMatch{start: i, end: i + winLen, score: score})
			}
		}
	}
	return matches
}

func quickScreening(normText []rune, target []rune, startPos, endPos int) bool {
	if startPos >= endPos {
		return false
	}
	candidateLen := endPos - startPos
	targetLen := len(target)
	if abs(candidateLen-targetLen) > max(candidateLen, targetLen)/2 {
		return false
	}
	if candidateLen > 10 && targetLen > 10 {
		commonChars := 0
		checkLen := min(5, min(candidateLen, targetLen))
		for i := 0; i < checkLen; i++ {
			for j := 0; j < checkLen; j++ {
				if normText[startPos+i] == target[j] {
					commonChars++
					break
				}
			}
		}
		if commonChars == 0 {
			return false
		}
	}
	return true
}

func FuzzyFindAllTargetBoundaries(recognizedText, targetText string, segLength, margin, thresholdStart, thresholdEnd, overallThreshold, boundaryTolerance int, removePunct bool) []Candidate {
	recognizedRunes := []rune(recognizedText)
	normText, mapping := normalizeText(recognizedText, removePunct)
	normTarget, _ := normalizeText(targetText, removePunct)
	if len(normTarget) == 0 {
		return nil
	}
	targetRunes := []rune(normTarget)
	normTextRunes := []rune(normText)
	segL := min(segLength, len(targetRunes))
	if segL == 0 {
		return nil
	}
	startSeg := string(targetRunes[:segL])
	endSeg := string(targetRunes[len(targetRunes)-segL:])
	startMatches := optimizedFuzzyFindSegments(normTextRunes, startSeg, margin, thresholdStart)
	endMatches := optimizedFuzzyFindSegments(normTextRunes, endSeg, margin, thresholdEnd)
	var candidates []Candidate
	candidateMap := make(map[string]bool)
	var builder strings.Builder
	for _, s := range startMatches {
		for _, e := range endMatches {
			if s.start >= e.end {
				continue
			}
			if !quickScreening(normTextRunes, targetRunes, s.start, e.end) {
				continue
			}
			builder.Reset()
			builder.Grow((e.end - s.start) * 4)
			for i := s.start; i < e.end; i++ {
				builder.WriteRune(normTextRunes[i])
			}
			candidateStr := builder.String()
			if candidateMap[candidateStr] {
				continue
			}
			candidateMap[candidateStr] = true
			overallScore := rapidfuzz.Ratio(candidateStr, normTarget)
			if overallScore >= overallThreshold {
				origStart := mapping[s.start]
				origEnd := mapping[e.end-1] + 1
				scoreStart := s.score
				scoreEnd := 0
				for _, em := range endMatches {
					if em.start == e.start && em.end == e.end {
						scoreEnd = em.score
						break
					}
				}
				candidates = append(candidates, Candidate{
					OrigStart:     mapping[s.start],
					OrigEnd:       mapping[e.end-1] + 1,
					MatchedSubstr: string(recognizedRunes[origStart:origEnd]),
					StartScore:    scoreStart,
					EndScore:      scoreEnd,
					OverallScore:  overallScore,
					NormCandidate: candidateStr,
				})
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].OverallScore > candidates[j].OverallScore
	})
	var final []Candidate
	for _, cand := range candidates {
		skip := false
		for _, sel := range final {
			if abs(cand.OrigStart-sel.OrigStart) <= boundaryTolerance && abs(cand.OrigEnd-sel.OrigEnd) <= boundaryTolerance {
				skip = true
				break
			}
		}
		if !skip {
			final = append(final, cand)
		}
	}
	if len(final) > 0 && final[0].OverallScore < 80 {
		return final[:1]
	}
	if len(final) > 10 {
		return final[:10]
	}
	return final
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
