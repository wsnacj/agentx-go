package volcengine

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/wsnacj/agentx-go/document/ocr/config"
	diffpkg "github.com/wsnacj/agentx-go/document/ocr/diff"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// NewProcessor 构造 Volcengine ProviderProcessor。
func NewProcessor(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
	if cfg.Kind != "volcengine" {
		return nil, fmt.Errorf("volcengine processor received provider %s", cfg.Kind)
	}
	return &provider{}, nil
}

type provider struct{}

func (provider) For(kind model.OperationKind) (processor.OperationProcessor, error) {
	switch kind {
	case model.OperationKindOCR:
		return ocrOperation{}, nil
	default:
		return nil, fmt.Errorf("volcengine processor: unsupported operation %s", kind)
	}
}

type ocrOperation struct{}

func (ocrOperation) Build(raw [][]byte, files []string) (any, error) {
	payload := model.OCRPayload{
		Files:        append([]string(nil), files...),
		RawResponses: append([][]byte(nil), raw...),
	}
	if len(raw) == 0 {
		return payload, nil
	}

	var combinedHTML strings.Builder
	var combinedText strings.Builder

	var collectedTexts []string
	for idx, data := range raw {
		page, err := parseResponse(data)
		if err != nil {
			return payload, fmt.Errorf("parse volcengine response[%d]: %w", idx, err)
		}

		pageText := recognizedTextFromResponse(page)
		collectedTexts = append(collectedTexts, pageText)
		payload.PageTexts = append(payload.PageTexts, pageText)
		if combinedText.Len() > 0 {
			combinedText.WriteByte('\n')
		}
		combinedText.WriteString(pageText)

		htmlPage := buildHTMLPage(idx, page.Data.LineTexts)
		payload.HTMLPages = append(payload.HTMLPages, htmlPage)
		combinedHTML.WriteString(htmlPage)

		payload.Coordinates = append(payload.Coordinates, extractCoordinates(page.Data)...)
		payload.TextBoxes = append(payload.TextBoxes, extractTextBoxes(page.Data, idx+1)...)
	}

	payload.CombinedHTML = combinedHTML.String()
	if len(collectedTexts) == 1 {
		payload.RecognizedText = collectedTexts[0]
	} else {
		payload.RecognizedText = combinedText.String()
	}
	return payload, nil
}

func (ocrOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	if len(baseline) == 0 || len(raw) == 0 {
		return nil, nil
	}

	baseText, err := recognizedTextFromRawSlices([][]byte{baseline})
	if err != nil {
		return nil, fmt.Errorf("parse volcengine baseline: %w", err)
	}
	currentText, err := recognizedTextFromRawSlices(raw)
	if err != nil {
		return nil, fmt.Errorf("parse volcengine diff payload: %w", err)
	}
	if baseText == currentText {
		return nil, nil
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(baseText, currentText, false)
	dmp.DiffCleanupSemantic(diffs)

	diffSegments := countDiffSegments(diffs)
	if diffSegments == 0 {
		return nil, nil
	}

	if preview <= 0 {
		preview = 3
	}
	previews := collectDiffPreview(diffs, preview)
	notes := collectFuzzyNotes(baseText, currentText, diffs, preview)

	return &model.DiffSummary{
		OCRDiff: &model.DiffResultSummary{
			HasDiff:       true,
			ExtraPages:    0,
			DiffPageCount: diffSegments,
			MappingScheme: "volcengine-text",
			Preview:       previews,
		},
		Notes: notes,
	}, nil
}

type apiResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    respData `json:"data"`
}

type respData struct {
	LineTexts []string     `json:"line_texts"`
	LineRects []volcRect   `json:"line_rects"`
	Chars     [][]volcChar `json:"chars"`
	Polygons  [][][]int    `json:"polygons"`
	LineProbs []float64    `json:"line_probs"`
}

type volcRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type volcChar struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Score  float64 `json:"score"`
	Char   string  `json:"char"`
}

func parseResponse(data []byte) (apiResponse, error) {
	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return apiResponse{}, err
	}
	if resp.Code != 10000 {
		return apiResponse{}, fmt.Errorf("unexpected response code %d message %s", resp.Code, resp.Message)
	}
	return resp, nil
}

func buildHTMLPage(idx int, lines []string) string {
	var buf strings.Builder
	buf.WriteString("<html><body>")
	fmt.Fprintf(&buf, "<h2>Volcengine OCR 识别结果 - 页面 %d</h2>", idx+1)
	buf.WriteString("<ul>")
	for _, line := range lines {
		buf.WriteString("<li>")
		buf.WriteString(html.EscapeString(line))
		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	buf.WriteString("</body></html>")
	return buf.String()
}

func extractCoordinates(data respData) []model.Coordinate {
	var coords []model.Coordinate
	for _, lineChars := range data.Chars {
		for _, ch := range lineChars {
			coords = append(coords, toCoordinate(ch.X, ch.Y, ch.Width, ch.Height))
		}
	}
	if len(coords) == 0 {
		for _, rect := range data.LineRects {
			coords = append(coords, toCoordinate(rect.X, rect.Y, rect.Width, rect.Height))
		}
	}
	return coords
}

func extractTextBoxes(data respData, page int) []model.TextBox {
	boxes := []model.TextBox{}
	for idx, text := range data.LineTexts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		var coord model.Coordinate
		confidence := 0.0
		if idx < len(data.Chars) && len(data.Chars[idx]) > 0 {
			coord = unionVolcChars(data.Chars[idx])
			confidence = averageVolcCharScore(data.Chars[idx])
		}
		if coord == (model.Coordinate{}) && idx < len(data.LineRects) {
			rect := data.LineRects[idx]
			coord = toCoordinate(rect.X, rect.Y, rect.Width, rect.Height)
		}
		if coord == (model.Coordinate{}) {
			continue
		}
		if confidence == 0 && idx < len(data.LineProbs) {
			confidence = data.LineProbs[idx]
		}
		boxes = append(boxes, model.TextBox{
			Page:       page,
			Text:       text,
			Coordinate: coord,
			Confidence: confidence,
			Level:      "line",
		})
	}
	return boxes
}

func unionVolcChars(chars []volcChar) model.Coordinate {
	if len(chars) == 0 {
		return model.Coordinate{}
	}
	left, top := math.MaxInt, math.MaxInt
	right, bottom := 0, 0
	for _, ch := range chars {
		coord := toCoordinate(ch.X, ch.Y, ch.Width, ch.Height)
		if coord.Left < left {
			left = coord.Left
		}
		if coord.Top < top {
			top = coord.Top
		}
		if coord.Right > right {
			right = coord.Right
		}
		if coord.Bottom > bottom {
			bottom = coord.Bottom
		}
	}
	return model.Coordinate{Left: left, Top: top, Right: right, Bottom: bottom}
}

func averageVolcCharScore(chars []volcChar) float64 {
	if len(chars) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for _, ch := range chars {
		if ch.Score <= 0 {
			continue
		}
		total += ch.Score
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func toCoordinate(x, y, width, height float64) model.Coordinate {
	left := normalizeCoord(x)
	top := normalizeCoord(y)
	right := normalizeCoord(x + width)
	bottom := normalizeCoord(y + height)
	if right < left {
		right = left
	}
	if bottom < top {
		bottom = top
	}
	return model.Coordinate{
		Left:   left,
		Top:    top,
		Right:  right,
		Bottom: bottom,
	}
}

func normalizeCoord(v float64) int {
	if math.Abs(v) <= 1 {
		return int(math.Round(v * 10000))
	}
	return int(math.Round(v))
}

func recognizedTextFromResponse(resp apiResponse) string {
	return strings.Join(resp.Data.LineTexts, "\n")
}

func recognizedTextFromRawSlices(raw [][]byte) (string, error) {
	var texts []string
	for idx, data := range raw {
		resp, err := parseResponse(data)
		if err != nil {
			return "", fmt.Errorf("parse volcengine response[%d]: %w", idx, err)
		}
		texts = append(texts, recognizedTextFromResponse(resp))
	}
	return strings.Join(texts, "\n"), nil
}

func countDiffSegments(diffs []diffmatchpatch.Diff) int {
	count := 0
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual {
			continue
		}
		if strings.TrimSpace(d.Text) == "" {
			continue
		}
		count++
	}
	return count
}

func collectDiffPreview(diffs []diffmatchpatch.Diff, limit int) []string {
	var previews []string
	for _, d := range diffs {
		var prefix string
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			prefix = "新增"
		case diffmatchpatch.DiffDelete:
			prefix = "缺失"
		default:
			continue
		}
		text := clipText(d.Text, 48)
		if strings.TrimSpace(text) == "" {
			continue
		}
		previews = append(previews, fmt.Sprintf("%s：%s", prefix, text))
		if len(previews) >= limit {
			break
		}
	}
	return previews
}

func collectFuzzyNotes(baseText, currentText string, diffs []diffmatchpatch.Diff, limit int) []string {
	var notes []string
	for _, d := range diffs {
		if len(notes) >= limit {
			break
		}
		target := strings.TrimSpace(d.Text)
		if target == "" {
			continue
		}
		clipped := clipText(target, 32)
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			cands := diffpkg.FuzzyLocateText(currentText, target, 6, 4, 60, 60, 70, 3, true)
			if len(cands) == 0 {
				notes = append(notes, fmt.Sprintf("新增片段“%s”未在当前文本中定位", clipped))
				continue
			}
			c := cands[0]
			notes = append(notes, fmt.Sprintf("新增片段“%s”疑似位置[%d,%d) 相似度=%d", clipped, c.OrigStart, c.OrigEnd, c.OverallScore))
		case diffmatchpatch.DiffDelete:
			cands := diffpkg.FuzzyLocateText(baseText, target, 6, 4, 60, 60, 70, 3, true)
			if len(cands) == 0 {
				notes = append(notes, fmt.Sprintf("缺失片段“%s”未在基线文本中定位", clipped))
				continue
			}
			c := cands[0]
			notes = append(notes, fmt.Sprintf("缺失片段“%s”原位置[%d,%d) 相似度=%d", clipped, c.OrigStart, c.OrigEnd, c.OverallScore))
		default:
		}
	}
	return notes
}

func clipText(s string, maxLen int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}
