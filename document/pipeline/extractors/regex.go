package extractors

import (
	"regexp"
	"strings"
)

type RegexInput struct {
	Text        string // chapter text
	Scope       string // header|full|footer
	Pattern     string
	HeaderLines int
	FooterLines int
}

type RegexResult struct {
	Value      string
	Confidence float64
	Snippet    string
	LineNumber int
}

// RunRegex 提供一个简单的正则抽取（取第一个捕获组，若无则整体匹配）
func RunRegex(in RegexInput) (RegexResult, bool) {
	candidates := RunRegexCandidates(in)
	if len(candidates) == 0 {
		return RegexResult{}, false
	}
	return candidates[0], true
}

func RunRegexCandidates(in RegexInput) []RegexResult {
	pat := strings.TrimSpace(in.Pattern)
	text := regexScopeText(in)
	if pat == "" || text == "" {
		return nil
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil
	}
	matches := re.FindAllStringSubmatchIndex(text, 32)
	if len(matches) == 0 {
		return nil
	}
	out := make([]RegexResult, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 || m[0] < 0 || m[1] < m[0] {
			continue
		}
		// 如果存在捕获组，取第一个捕获组
		val := text[m[0]:m[1]]
		if len(m) >= 4 && m[2] >= 0 && m[3] >= m[2] {
			val = text[m[2]:m[3]]
		}
		// 简单置信度：根据匹配长度/文本长度估计
		conf := 0.6
		if len(val) > 0 && len([]rune(val)) <= 64 {
			conf += 0.2
		}
		out = append(out, RegexResult{
			Value:      strings.TrimSpace(val),
			Confidence: conf,
			Snippet:    text[m[0]:m[1]],
			LineNumber: lineNumberAtByteOffset(text, m[0]),
		})
	}
	return out
}

func lineNumberAtByteOffset(text string, offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	line := 1
	for i := 0; i < offset; i++ {
		if text[i] == '\n' {
			line++
		}
	}
	return line
}

func regexScopeText(in RegexInput) string {
	text := in.Text
	if strings.TrimSpace(text) == "" {
		return ""
	}
	scope := strings.ToLower(strings.TrimSpace(in.Scope))
	switch scope {
	case "", "full", "body":
		return text
	case "header":
		n := in.HeaderLines
		if n <= 0 {
			n = 6
		}
		return firstLines(text, n)
	case "footer":
		n := in.FooterLines
		if n <= 0 {
			n = 4
		}
		return lastLines(text, n)
	default:
		return text
	}
}

func firstLines(text string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func lastLines(text string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
