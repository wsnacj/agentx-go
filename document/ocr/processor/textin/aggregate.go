package textin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/wsnacj/agentx-go/document/ocr/model"
)

// --- OCR payload helpers ---

type textInOCRResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Result  textInOCRResult `json:"result"`
}

type textInOCRResult struct {
	Pages []textInOCRPage `json:"pages"`
}

type textInOCRPage struct {
	Angle  int             `json:"angle"`
	Width  int             `json:"width"`
	Height int             `json:"height"`
	Lines  []textInOCRLine `json:"lines"`
}

type textInOCRLine struct {
	Text                string      `json:"text"`
	CharCandidates      [][]string  `json:"char_candidates"`
	CharCandidatesScore [][]float64 `json:"char_candidates_score"`
	CharPositions       [][]int     `json:"char_positions"`
}

func buildOCRPayload(raw [][]byte, files []string) (model.OCRPayload, error) {
	payload := model.OCRPayload{
		Files:        append([]string(nil), files...),
		RawResponses: append([][]byte(nil), raw...),
	}
	if len(raw) == 0 {
		return payload, nil
	}

	var combinedHTML strings.Builder
	var combinedText strings.Builder
	var normalizedText strings.Builder

	for idx, data := range raw {
		var resp textInOCRResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return payload, fmt.Errorf("parse ocr response[%d]: %w", idx, err)
		}
		if resp.Code != 200 {
			return payload, fmt.Errorf("ocr response[%d] failed: %s (code=%d)", idx, resp.Message, resp.Code)
		}

		pagesHTML := buildOCRHTMLPages(&resp)
		for _, html := range pagesHTML {
			payload.HTMLPages = append(payload.HTMLPages, html)
			combinedHTML.WriteString(html)
		}

		for _, page := range resp.Result.Pages {
			pageNumber := len(payload.PageTexts) + 1
			text, normalizedPageText, coords, boxes := extractOCRText(page.Lines, pageNumber, page.Width, page.Height)
			combinedText.WriteString(text)
			if normalizedPageText != "" {
				if normalizedText.Len() > 0 {
					normalizedText.WriteString("\n\n")
				}
				normalizedText.WriteString(normalizedPageText)
			}
			payload.Coordinates = append(payload.Coordinates, coords...)
			payload.TextBoxes = append(payload.TextBoxes, boxes...)
			payload.PageTexts = append(payload.PageTexts, text)
			payload.NormalizedPageTexts = append(payload.NormalizedPageTexts, normalizedPageText)
		}
	}

	payload.CombinedHTML = combinedHTML.String()
	payload.RecognizedText = combinedText.String()
	payload.NormalizedText = normalizedText.String()
	return payload, nil
}

func buildOCRHTML(resp *textInOCRResponse) string {
	var html bytes.Buffer
	for idx, page := range resp.Result.Pages {
		html.WriteString(buildOCRPageHTML(page, idx))
	}
	return html.String()
}

func buildOCRHTMLPages(resp *textInOCRResponse) []string {
	var pages []string
	for idx, page := range resp.Result.Pages {
		pages = append(pages, buildOCRPageHTML(page, idx))
	}
	return pages
}

func buildOCRPageHTML(page textInOCRPage, index int) string {
	var buf bytes.Buffer
	buf.WriteString("<html><body>")
	fmt.Fprintf(&buf, "<h2>OCR 识别结果 - 页面 %d (角度: %d, 宽度: %d, 高度: %d)</h2>", index+1, page.Angle, page.Width, page.Height)
	buf.WriteString("<ul>")
	for _, line := range page.Lines {
		buf.WriteString("<li>")
		buf.WriteString(html.EscapeString(line.Text))
		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	buf.WriteString("</body></html>")
	return buf.String()
}

func extractOCRText(lines []textInOCRLine, page int, pageWidth int, pageHeight int) (string, string, []model.Coordinate, []model.TextBox) {
	var rawBuilder strings.Builder
	var normalizedBuilder strings.Builder
	var coords []model.Coordinate
	var boxes []model.TextBox

	for _, line := range lines {
		lineText, lineCoords := extractOCRLine(line)
		rawBuilder.WriteString(lineText)
		coords = append(coords, lineCoords...)
		if trimmed := strings.TrimSpace(lineText); trimmed != "" {
			if normalizedBuilder.Len() > 0 {
				normalizedBuilder.WriteByte('\n')
			}
			normalizedBuilder.WriteString(trimmed)
			if bbox := unionCoordinates(lineCoords); bbox != (model.Coordinate{}) {
				boxes = append(boxes, model.TextBox{
					Page:       page,
					Text:       trimmed,
					Coordinate: bbox,
					PageWidth:  pageWidth,
					PageHeight: pageHeight,
					Level:      "line",
				})
			}
		}
	}
	return rawBuilder.String(), normalizedBuilder.String(), coords, boxes
}

func extractOCRLine(line textInOCRLine) (string, []model.Coordinate) {
	var builder strings.Builder
	var coords []model.Coordinate
	added := false

	if len(line.CharCandidates) > 0 && len(line.CharPositions) > 0 {
		for i, candidates := range line.CharCandidates {
			if i >= len(line.CharPositions) {
				continue
			}
			var scores []float64
			if i < len(line.CharCandidatesScore) {
				scores = line.CharCandidatesScore[i]
			}
			bestChar, ok := selectBestCandidate(candidates, scores)
			if !ok {
				continue
			}
			builder.WriteString(bestChar)
			if bbox := computeBoundingBox(line.CharPositions[i]); bbox != (model.Coordinate{}) {
				coords = append(coords, bbox)
			}
			added = true
		}
	}
	if !added && strings.TrimSpace(line.Text) != "" {
		builder.WriteString(line.Text)
	}
	return builder.String(), coords
}

// --- Table payload helpers ---

type textInTableResponse struct {
	Code   int              `json:"code"`
	Msg    string           `json:"message"`
	Result textInTablePages `json:"result"`
}

type textInTablePages struct {
	Pages []textInTablePage `json:"pages"`
}

type textInTablePage struct {
	Angle  int           `json:"angle"`
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Tables []textInTable `json:"tables"`
}

type textInTable struct {
	Lines      []textInTableLine `json:"lines"`
	TableCells []textInTableCell `json:"table_cells"`
	TableRows  int               `json:"table_rows"`
	TableCols  int               `json:"table_cols"`
	Position   []int             `json:"position"`
}

type textInTableLine struct {
	Text                string      `json:"text"`
	CharCandidates      [][]string  `json:"char_candidates"`
	CharCandidatesScore [][]float64 `json:"char_candidates_score"`
	CharPositions       [][]int     `json:"char_positions"`
}

type textInTableCell struct {
	StartRow int               `json:"start_row"`
	StartCol int               `json:"start_col"`
	EndRow   int               `json:"end_row"`
	EndCol   int               `json:"end_col"`
	Text     string            `json:"text"`
	Lines    []textInTableLine `json:"lines"`
}

func buildTablePayload(raw [][]byte, files []string) (model.TablePayload, error) {
	payload := model.TablePayload{
		Files: append([]string(nil), files...),
		Raw:   append([][]byte(nil), raw...),
	}
	if len(raw) == 0 {
		return payload, nil
	}

	var combinedHTML strings.Builder
	var normalizedCombined strings.Builder

	for idx, data := range raw {
		var resp textInTableResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return payload, fmt.Errorf("parse table response[%d]: %w", idx, err)
		}
		if resp.Code != 200 {
			return payload, fmt.Errorf("table response[%d] failed: %s (code=%d)", idx, resp.Msg, resp.Code)
		}

		pagesHTML := buildTableHTMLPages(&resp)
		for _, html := range pagesHTML {
			payload.HTMLPages = append(payload.HTMLPages, html)
			combinedHTML.WriteString(html)
		}

		for _, page := range resp.Result.Pages {
			text, normalizedText, coords := extractTableText(page)
			payload.Text = append(payload.Text, text)
			payload.NormalizedTexts = append(payload.NormalizedTexts, normalizedText)
			if normalizedText != "" {
				if normalizedCombined.Len() > 0 {
					normalizedCombined.WriteString("\n\n")
				}
				normalizedCombined.WriteString(normalizedText)
			}
			payload.Pages = append(payload.Pages, model.TablePage{
				Index:       len(payload.Pages) + 1,
				Width:       page.Width,
				Height:      page.Height,
				Recognized:  text,
				Coordinates: coords,
			})
		}
	}

	payload.CombinedHTML = combinedHTML.String()
	payload.NormalizedCombinedText = normalizedCombined.String()
	return payload, nil
}

func buildTablePageHTML(page textInTablePage, index int) string {
	var buf bytes.Buffer
	buf.WriteString("<html><body>")
	fmt.Fprintf(&buf, "<h2>表格 OCR 识别结果 - 页面 %d (宽度: %d, 高度: %d)</h2>", index+1, page.Width, page.Height)
	for _, tbl := range page.Tables {
		buf.WriteString("<table border='1'>")
		buf.WriteString("<ul>")
		for _, line := range tbl.Lines {
			buf.WriteString("<li>")
			buf.WriteString(html.EscapeString(line.Text))
			buf.WriteString("</li>")
		}
		buf.WriteString("</ul>")

		cellMap := make(map[int]map[int]*textInTableCell)
		for i := range tbl.TableCells {
			cell := &tbl.TableCells[i]
			if _, exists := cellMap[cell.StartRow]; !exists {
				cellMap[cell.StartRow] = make(map[int]*textInTableCell)
			}
			cellMap[cell.StartRow][cell.StartCol] = cell
		}
		for row := 0; row < tbl.TableRows; row++ {
			buf.WriteString("<tr>")
			for col := 0; col < tbl.TableCols; col++ {
				if cellRow, ok := cellMap[row]; ok {
					if cell, exists := cellRow[col]; exists {
						rowSpan := cell.EndRow - cell.StartRow + 1
						colSpan := cell.EndCol - cell.StartCol + 1
						if row == cell.StartRow && col == cell.StartCol {
							buf.WriteString(fmt.Sprintf("<td rowspan='%d' colspan='%d'>%s</td>", rowSpan, colSpan, html.EscapeString(cell.Text)))
						}
						continue
					}
				}
				covered := false
				for _, rowMap := range cellMap {
					for _, cell := range rowMap {
						if row >= cell.StartRow && row <= cell.EndRow && col >= cell.StartCol && col <= cell.EndCol {
							covered = true
							break
						}
					}
					if covered {
						break
					}
				}
				if !covered {
					buf.WriteString("<td></td>")
				}
			}
			buf.WriteString("</tr>")
		}
		buf.WriteString("</table><br/>")
	}
	buf.WriteString("</body></html>")
	return buf.String()
}

func buildTableHTML(resp *textInTableResponse) string {
	var buf bytes.Buffer
	for idx, page := range resp.Result.Pages {
		buf.WriteString(buildTablePageHTML(page, idx))
	}
	return buf.String()
}

func buildTableHTMLPages(resp *textInTableResponse) []string {
	var pages []string
	for idx, page := range resp.Result.Pages {
		pages = append(pages, buildTablePageHTML(page, idx))
	}
	return pages
}

func extractTableText(page textInTablePage) (string, string, []model.Coordinate) {
	var rawBuilder strings.Builder
	var normalizedBuilder strings.Builder
	var coords []model.Coordinate

	for _, tbl := range page.Tables {
		for _, line := range tbl.Lines {
			text, cs := extractTableLine(line)
			rawBuilder.WriteString(text)
			coords = append(coords, cs...)
			if trimmed := strings.TrimSpace(text); trimmed != "" {
				if normalizedBuilder.Len() > 0 {
					normalizedBuilder.WriteByte('\n')
				}
				normalizedBuilder.WriteString(trimmed)
			}
		}
		for _, cell := range tbl.TableCells {
			if len(cell.Lines) == 0 && strings.TrimSpace(cell.Text) != "" {
				rawBuilder.WriteString(cell.Text)
				if normalizedBuilder.Len() > 0 {
					normalizedBuilder.WriteByte('\n')
				}
				normalizedBuilder.WriteString(strings.TrimSpace(cell.Text))
				continue
			}
			for _, line := range cell.Lines {
				text, cs := extractTableLine(line)
				rawBuilder.WriteString(text)
				coords = append(coords, cs...)
				if trimmed := strings.TrimSpace(text); trimmed != "" {
					if normalizedBuilder.Len() > 0 {
						normalizedBuilder.WriteByte('\n')
					}
					normalizedBuilder.WriteString(trimmed)
				}
			}
		}
	}
	return rawBuilder.String(), normalizedBuilder.String(), coords
}

func extractTableLine(line textInTableLine) (string, []model.Coordinate) {
	var builder strings.Builder
	var coords []model.Coordinate
	added := false

	if len(line.CharCandidates) > 0 && len(line.CharPositions) > 0 {
		for i, candidates := range line.CharCandidates {
			if i >= len(line.CharPositions) {
				continue
			}
			var scores []float64
			if i < len(line.CharCandidatesScore) {
				scores = line.CharCandidatesScore[i]
			}
			bestChar, ok := selectBestCandidate(candidates, scores)
			if !ok {
				continue
			}
			builder.WriteString(bestChar)
			if bbox := computeBoundingBox(line.CharPositions[i]); bbox != (model.Coordinate{}) {
				coords = append(coords, bbox)
			}
			added = true
		}
	}

	if !added && strings.TrimSpace(line.Text) != "" {
		builder.WriteString(line.Text)
	}
	return builder.String(), coords
}

// --- Stamp payload helpers ---

type textInStampResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Result  textInStampResult `json:"result"`
}

type textInStampResult struct {
	Details struct {
		Stamp []textInStampDetail `json:"stamp"`
	} `json:"details"`
}

type textInStampDetail struct {
	Value    string `json:"value"`
	Type     string `json:"type"`
	Shape    string `json:"stamp_shape"`
	Color    string `json:"color"`
	Position []int  `json:"position"`
}

func buildStampPayload(raw [][]byte, files []string) (model.StampPayload, error) {
	payload := model.StampPayload{
		Files: append([]string(nil), files...),
		Raw:   append([][]byte(nil), raw...),
	}

	for idx, data := range raw {
		var resp textInStampResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return payload, fmt.Errorf("parse stamp response[%d]: %w", idx, err)
		}
		if resp.Code != 200 {
			return payload, fmt.Errorf("stamp response[%d] failed: %s (code=%d)", idx, resp.Message, resp.Code)
		}

		page := model.StampPage{Index: len(payload.Pages) + 1}
		for _, detail := range resp.Result.Details.Stamp {
			page.Stamp = append(page.Stamp, model.StampDetail{
				Text:     detail.Value,
				Type:     detail.Type,
				Shape:    detail.Shape,
				Color:    detail.Color,
				Position: append([]int(nil), detail.Position...),
			})
		}
		payload.Pages = append(payload.Pages, page)
	}

	return payload, nil
}

// --- Shared helpers ---

func selectBestCandidate(candidates []string, scores []float64) (string, bool) {
	if len(candidates) == 0 {
		return "", false
	}
	limit := len(candidates)
	if len(scores) > 0 && len(scores) < limit {
		limit = len(scores)
	}
	if limit > 0 && len(scores) > 0 {
		bestIdx := 0
		bestScore := scores[0]
		for i := 1; i < limit; i++ {
			if scores[i] > bestScore {
				bestScore = scores[i]
				bestIdx = i
			}
		}
		if bestIdx < len(candidates) && candidates[bestIdx] != "" {
			return candidates[bestIdx], true
		}
	}
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate, true
		}
	}
	return "", false
}

func mergeTextInOCRResponses(raw [][]byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) == 1 {
		return append([]byte(nil), raw[0]...), nil
	}

	merged := textInOCRResponse{Code: 200, Message: "OK"}
	for idx, data := range raw {
		var resp textInOCRResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("parse ocr response[%d]: %w", idx, err)
		}
		if idx == 0 {
			if resp.Code != 0 {
				merged.Code = resp.Code
			}
			if strings.TrimSpace(resp.Message) != "" {
				merged.Message = resp.Message
			}
		}
		merged.Result.Pages = append(merged.Result.Pages, resp.Result.Pages...)
	}
	data, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged ocr response: %w", err)
	}
	return data, nil
}

func mergeTextInTableResponses(raw [][]byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) == 1 {
		return append([]byte(nil), raw[0]...), nil
	}

	merged := textInTableResponse{Code: 200, Msg: "OK"}
	for idx, data := range raw {
		var resp textInTableResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("parse table response[%d]: %w", idx, err)
		}
		if idx == 0 {
			if resp.Code != 0 {
				merged.Code = resp.Code
			}
			if strings.TrimSpace(resp.Msg) != "" {
				merged.Msg = resp.Msg
			}
		}
		merged.Result.Pages = append(merged.Result.Pages, resp.Result.Pages...)
	}
	data, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged table response: %w", err)
	}
	return data, nil
}

func computeBoundingBox(pos []int) model.Coordinate {
	if len(pos) < 2 {
		return model.Coordinate{}
	}
	left, top := math.MaxInt, math.MaxInt
	right, bottom := 0, 0
	for i := 0; i < len(pos)/2; i++ {
		x, y := pos[2*i], pos[2*i+1]
		if x < left {
			left = x
		}
		if x > right {
			right = x
		}
		if y < top {
			top = y
		}
		if y > bottom {
			bottom = y
		}
	}
	return model.Coordinate{Left: left, Top: top, Right: right, Bottom: bottom}
}

func unionCoordinates(coords []model.Coordinate) model.Coordinate {
	if len(coords) == 0 {
		return model.Coordinate{}
	}
	left, top := math.MaxInt, math.MaxInt
	right, bottom := 0, 0
	seen := false
	for _, coord := range coords {
		if coord == (model.Coordinate{}) {
			continue
		}
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
		seen = true
	}
	if !seen {
		return model.Coordinate{}
	}
	return model.Coordinate{Left: left, Top: top, Right: right, Bottom: bottom}
}
