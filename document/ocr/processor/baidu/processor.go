package baidu

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/wsnacj/agentx-go/document/ocr/config"
	diffpkg "github.com/wsnacj/agentx-go/document/ocr/diff"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// NewProcessor 构造 Baidu ProviderProcessor。
func NewProcessor(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
	if cfg.Kind != "baidu" {
		return nil, fmt.Errorf("baidu processor received provider %s", cfg.Kind)
	}
	return provider{}, nil
}

type provider struct{}

type ocrOperation struct{}
type tableOperation struct{}

func (provider) For(kind model.OperationKind) (processor.OperationProcessor, error) {
	switch kind {
	case model.OperationKindOCR:
		return ocrOperation{}, nil
	case model.OperationKindTable:
		return tableOperation{}, nil
	default:
		return nil, fmt.Errorf("baidu processor: unsupported operation %s", kind)
	}
}

func (ocrOperation) Build(raw [][]byte, files []string) (any, error) {
	payload := model.OCRPayload{
		Files:        append([]string(nil), files...),
		RawResponses: append([][]byte(nil), raw...),
	}
	if len(raw) == 0 {
		return payload, nil
	}

	var combined strings.Builder
	var htmlPages []string
	var pageTexts []string
	var coords []model.Coordinate
	var boxes []model.TextBox

	for idx, data := range raw {
		var resp baiduResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return payload, fmt.Errorf("parse baidu response[%d]: %w", idx, err)
		}
		text := resp.Text()
		if combined.Len() > 0 {
			combined.WriteByte('\n')
		}
		combined.WriteString(text)
		pageTexts = append(pageTexts, text)
		htmlPages = append(htmlPages, resp.HTML(idx))
		coords = append(coords, resp.Coordinates()...)
		boxes = append(boxes, resp.TextBoxes(idx+1)...)
	}

	payload.RecognizedText = combined.String()
	payload.PageTexts = pageTexts
	payload.HTMLPages = htmlPages
	payload.CombinedHTML = strings.Join(htmlPages, "")
	payload.Coordinates = coords
	payload.TextBoxes = boxes
	return payload, nil
}

func (ocrOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	if len(baseline) == 0 || len(raw) == 0 {
		return nil, nil
	}
	baseText, err := recognizedTextFromRaw([][]byte{baseline})
	if err != nil {
		return nil, fmt.Errorf("parse baidu baseline: %w", err)
	}
	curText, err := recognizedTextFromRaw(raw)
	if err != nil {
		return nil, fmt.Errorf("parse baidu diff payload: %w", err)
	}
	if baseText == curText {
		return nil, nil
	}
	if preview <= 0 {
		preview = 3
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(baseText, curText, false)
	dmp.DiffCleanupSemantic(diffs)

	segments := countDiffSegments(diffs)
	if segments == 0 {
		return nil, nil
	}
	previews := collectDiffPreview(diffs, preview)
	notes := collectFuzzyNotes(baseText, curText, diffs, preview)

	return &model.DiffSummary{
		OCRDiff: &model.DiffResultSummary{
			HasDiff:       true,
			ExtraPages:    0,
			DiffPageCount: segments,
			MappingScheme: "baidu-text",
			Preview:       previews,
		},
		Notes: notes,
	}, nil
}

type baiduResponse struct {
	WordsResult []struct {
		Words    string        `json:"words"`
		Location locationBlock `json:"location"`
	} `json:"words_result"`
}

type locationBlock struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (br baiduResponse) Text() string {
	var builder strings.Builder
	for idx, item := range br.WordsResult {
		if idx > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(item.Words)
	}
	return builder.String()
}

func (br baiduResponse) Coordinates() []model.Coordinate {
	coords := make([]model.Coordinate, 0, len(br.WordsResult))
	for _, item := range br.WordsResult {
		coords = append(coords, locationCoordinate(item.Location))
	}
	return coords
}

func (br baiduResponse) TextBoxes(page int) []model.TextBox {
	boxes := make([]model.TextBox, 0, len(br.WordsResult))
	for _, item := range br.WordsResult {
		text := strings.TrimSpace(item.Words)
		if text == "" {
			continue
		}
		boxes = append(boxes, model.TextBox{
			Page:       page,
			Text:       text,
			Coordinate: locationCoordinate(item.Location),
			Level:      "line",
		})
	}
	return boxes
}

func locationCoordinate(location locationBlock) model.Coordinate {
	return model.Coordinate{
		Left:   location.Left,
		Top:    location.Top,
		Right:  location.Left + location.Width,
		Bottom: location.Top + location.Height,
	}
}

func (br baiduResponse) HTML(idx int) string {
	var builder strings.Builder
	builder.WriteString("<html><body>")
	fmt.Fprintf(&builder, "<h2>Baidu OCR 识别结果 - 页面 %d</h2>", idx+1)
	builder.WriteString("<ul>")
	for _, item := range br.WordsResult {
		builder.WriteString("<li>")
		builder.WriteString(html.EscapeString(item.Words))
		builder.WriteString("</li>")
	}
	builder.WriteString("</ul></body></html>")
	return builder.String()
}

func recognizedTextFromRaw(raw [][]byte) (string, error) {
	var texts []string
	for idx, data := range raw {
		var resp baiduResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return "", fmt.Errorf("parse baidu response[%d]: %w", idx, err)
		}
		texts = append(texts, resp.Text())
	}
	return strings.Join(texts, "\n"), nil
}

// --- Table specific processing ---

type baiduTableResponse struct {
	TablesResult []tableResult `json:"tables_result"`
	TableNum     int           `json:"table_num"`
	ExcelFile    string        `json:"excel_file"`
}

type tableResult struct {
	TableLocation []point     `json:"table_location"`
	Header        []tableText `json:"header"`
	Body          []tableCell `json:"body"`
	Footer        []tableText `json:"footer"`
}

type tableText struct {
	Location []point `json:"location"`
	Words    string  `json:"words"`
}

type tableCell struct {
	CellLocation []point `json:"cell_location"`
	RowStart     int     `json:"row_start"`
	RowEnd       int     `json:"row_end"`
	ColStart     int     `json:"col_start"`
	ColEnd       int     `json:"col_end"`
	Words        string  `json:"words"`
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (tableOperation) Build(raw [][]byte, files []string) (any, error) {
	payload := model.TablePayload{
		Files: append([]string(nil), files...),
		Raw:   append([][]byte(nil), raw...),
	}
	if len(raw) == 0 {
		return payload, nil
	}

	var (
		htmlPages []string
		texts     []string
		pages     []model.TablePage
		pageIndex int
	)

	for idx, data := range raw {
		var resp baiduTableResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return payload, fmt.Errorf("parse baidu table response[%d]: %w", idx, err)
		}
		tableTexts := make([]string, 0, len(resp.TablesResult))
		tableHTML := make([]string, 0, len(resp.TablesResult))
		for _, tbl := range resp.TablesResult {
			tableTexts = append(tableTexts, tableToText(tbl))
			tableHTML = append(tableHTML, tableToHTML(tbl, pageIndex))
			pages = append(pages, model.TablePage{
				Index:      pageIndex,
				Recognized: tableToText(tbl),
				Coordinates: []model.Coordinate{
					tableCoordinate(tbl.TableLocation),
				},
			})
			pageIndex++
		}
		if len(tableTexts) == 0 {
			continue
		}
		texts = append(texts, strings.Join(tableTexts, "\n\n"))
		htmlPages = append(htmlPages, strings.Join(tableHTML, "\n"))
	}

	payload.Pages = pages
	payload.HTMLPages = htmlPages
	payload.CombinedHTML = strings.Join(htmlPages, "\n")
	payload.Text = texts
	return payload, nil
}

func (tableOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	return nil, nil
}

func tableCoordinate(points []point) model.Coordinate {
	if len(points) == 0 {
		return model.Coordinate{}
	}
	left, top := points[0].X, points[0].Y
	right, bottom := left, top
	for _, pt := range points {
		if pt.X < left {
			left = pt.X
		}
		if pt.X > right {
			right = pt.X
		}
		if pt.Y < top {
			top = pt.Y
		}
		if pt.Y > bottom {
			bottom = pt.Y
		}
	}
	return model.Coordinate{Left: left, Top: top, Right: right, Bottom: bottom}
}

func tableToText(tbl tableResult) string {
	var builder strings.Builder
	if len(tbl.Header) > 0 {
		builder.WriteString("[Header]\n")
		for _, h := range tbl.Header {
			builder.WriteString(" - ")
			builder.WriteString(h.Words)
			builder.WriteByte('\n')
		}
	}
	if len(tbl.Body) > 0 {
		builder.WriteString("[Body]\n")
		for _, cell := range tbl.Body {
			fmt.Fprintf(&builder, "r%d-%d c%d-%d %s\n", cell.RowStart, cell.RowEnd, cell.ColStart, cell.ColEnd, cell.Words)
		}
	}
	if len(tbl.Footer) > 0 {
		builder.WriteString("[Footer]\n")
		for _, f := range tbl.Footer {
			builder.WriteString(" - ")
			builder.WriteString(f.Words)
			builder.WriteByte('\n')
		}
	}
	return strings.TrimSpace(builder.String())
}

func tableToHTML(tbl tableResult, index int) string {
	rows, cols := estimateTableSize(tbl.Body)
	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}
	for _, cell := range tbl.Body {
		r := clamp(cell.RowStart, 0, rows-1)
		c := clamp(cell.ColStart, 0, cols-1)
		if grid[r][c] == "" {
			grid[r][c] = cell.Words
		} else {
			grid[r][c] = grid[r][c] + " | " + cell.Words
		}
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "<h3>表格 %d</h3>", index+1)
	builder.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"4\">")
	for r := 0; r < rows; r++ {
		builder.WriteString("<tr>")
		for c := 0; c < cols; c++ {
			builder.WriteString("<td>")
			builder.WriteString(html.EscapeString(grid[r][c]))
			builder.WriteString("</td>")
		}
		builder.WriteString("</tr>")
	}
	builder.WriteString("</table>")
	return builder.String()
}

func estimateTableSize(cells []tableCell) (int, int) {
	maxRow, maxCol := 1, 1
	for _, cell := range cells {
		if cell.RowEnd > maxRow {
			maxRow = cell.RowEnd
		}
		if cell.ColEnd > maxCol {
			maxCol = cell.ColEnd
		}
	}
	if maxRow <= 0 {
		maxRow = 1
	}
	if maxCol <= 0 {
		maxCol = 1
	}
	return maxRow, maxCol
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
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
				notes = append(notes, fmt.Sprintf("新增片段“%s”未定位", clipped))
				continue
			}
			c := cands[0]
			notes = append(notes, fmt.Sprintf("新增片段“%s”疑似位置[%d,%d) 相似度=%d", clipped, c.OrigStart, c.OrigEnd, c.OverallScore))
		case diffmatchpatch.DiffDelete:
			cands := diffpkg.FuzzyLocateText(baseText, target, 6, 4, 60, 60, 70, 3, true)
			if len(cands) == 0 {
				notes = append(notes, fmt.Sprintf("缺失片段“%s”未定位", clipped))
				continue
			}
			c := cands[0]
			notes = append(notes, fmt.Sprintf("缺失片段“%s”原位置[%d,%d) 相似度=%d", clipped, c.OrigStart, c.OrigEnd, c.OverallScore))
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
