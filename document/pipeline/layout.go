package pipeline

import (
	"github.com/wsnacj/agentx-go/document/pdf"
	"github.com/wsnacj/agentx-go/document/pipeline/extractors"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"regexp"
	"sort"
	"strings"
)

type documentLayout struct {
	Pages map[int]pageLayout
}

type pageLayout struct {
	Page   int
	Width  int
	Height int
	Tables []tableLayout
}

type tableLayout struct {
	Page       int
	TableIndex int
	Type       string
	Rows       int
	Cols       int
	Cells      []tableCellLayout
	Box        *types.BoundingBox
}

type tableCellLayout struct {
	Text     string
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
	Box      *types.BoundingBox
}

var layoutNonAlnumRe = regexp.MustCompile(`[^\p{Han}\p{L}\p{N}]+`)

func buildDocumentLayoutFromPDFResponse(resp *pdfparser.TableResponse) *documentLayout {
	if resp == nil || len(resp.Result.Pages) == 0 {
		return nil
	}
	layout := &documentLayout{Pages: map[int]pageLayout{}}
	for pageIdx, page := range resp.Result.Pages {
		pageNumber := pageIdx + 1
		pageLayout := pageLayout{
			Page:   pageNumber,
			Width:  page.Width,
			Height: page.Height,
		}
		for tableIdx, table := range page.Tables {
			if table.Type != "table_with_line" && table.Type != "table_without_line" {
				continue
			}
			t := tableLayout{
				Page:       pageNumber,
				TableIndex: tableIdx + 1,
				Type:       table.Type,
				Rows:       table.TableRows,
				Cols:       table.TableCols,
			}
			if box, ok := pdfPositionBoundingBox(pageNumber, table.Position); ok {
				t.Box = &box
			}
			for _, cell := range table.TableCells {
				text := strings.TrimSpace(cell.Text)
				if text == "" {
					continue
				}
				box, ok := pdfPositionBoundingBox(pageNumber, cell.Position)
				if !ok {
					continue
				}
				c := tableCellLayout{
					Text:     text,
					StartRow: cell.StartRow,
					StartCol: cell.StartCol,
					EndRow:   cell.EndRow,
					EndCol:   cell.EndCol,
					Box:      &box,
				}
				t.Cells = append(t.Cells, c)
			}
			if len(t.Cells) > 0 {
				pageLayout.Tables = append(pageLayout.Tables, t)
			}
		}
		if len(pageLayout.Tables) > 0 {
			layout.Pages[pageNumber] = pageLayout
		}
	}
	if len(layout.Pages) == 0 {
		return nil
	}
	return layout
}

func pdfPositionBoundingBox(page int, pos []int) (types.BoundingBox, bool) {
	if len(pos) < 4 {
		return types.BoundingBox{}, false
	}
	return types.BoundingBox{
		Page:             page,
		X0:               float64(pos[0]),
		Y0:               float64(pos[1]),
		X1:               float64(pos[2]),
		Y1:               float64(pos[3]),
		Unit:             "source_unit",
		CoordinateSystem: "page_top_left",
		Source:           "pdfparser",
	}, true
}

func tableLayoutEvidenceForResult(layout *documentLayout, pageRefs []int, result extractors.TableResult) ([]types.BoundingBox, []types.TableCellRef) {
	if layout == nil || len(layout.Pages) == 0 || strings.TrimSpace(result.Value) == "" {
		return nil, nil
	}
	targetValue := compactLayoutValue(result.Value)
	if targetValue == "" {
		return nil, nil
	}
	pages := uniquePositiveInts(pageRefs)
	for _, pageNumber := range pages {
		page, ok := layout.Pages[pageNumber]
		if !ok {
			continue
		}
		for _, table := range page.Tables {
			for _, cell := range table.Cells {
				if !layoutValueMatches(cell.Text, targetValue) {
					continue
				}
				if !layoutRowLabelMatches(table, cell, result.RowLabel) {
					continue
				}
				if !layoutColumnLabelMatches(table, cell, result.ColumnLabel) {
					continue
				}
				return tableCellLayoutEvidence(table, cell)
			}
		}
	}
	if boxes, refs := multiPageTableLayoutEvidenceForResult(layout, pages, result, targetValue); len(refs) > 0 {
		return boxes, refs
	}
	return nil, nil
}

func multiPageTableLayoutEvidenceForResult(layout *documentLayout, pageRefs []int, result extractors.TableResult, targetValue string) ([]types.BoundingBox, []types.TableCellRef) {
	if len(pageRefs) < 2 || compactLayoutLabel(result.RowLabel) == "" || compactLayoutLabel(result.ColumnLabel) == "" {
		return nil, nil
	}
	for _, pageNumber := range pageRefs {
		page, ok := layout.Pages[pageNumber]
		if !ok {
			continue
		}
		for _, table := range page.Tables {
			for _, valueCell := range table.Cells {
				if !layoutValueMatches(valueCell.Text, targetValue) {
					continue
				}
				if !layoutColumnLabelMatches(table, valueCell, result.ColumnLabel) {
					continue
				}
				labelTable, labelCell, ok := priorPageRowLabelCell(layout, pageRefs, pageNumber, result.RowLabel)
				if !ok {
					continue
				}
				return stitchedTableCellLayoutEvidence(labelTable, labelCell, table, valueCell)
			}
		}
	}
	return nil, nil
}

func priorPageRowLabelCell(layout *documentLayout, pageRefs []int, valuePageNumber int, label string) (tableLayout, tableCellLayout, bool) {
	target := compactLayoutLabel(label)
	if target == "" || !containsInt(pageRefs, valuePageNumber-1) {
		return tableLayout{}, tableCellLayout{}, false
	}
	page, ok := layout.Pages[valuePageNumber-1]
	if !ok {
		return tableLayout{}, tableCellLayout{}, false
	}
	var bestTable tableLayout
	var bestCell tableCellLayout
	found := false
	for _, table := range page.Tables {
		for _, cell := range table.Cells {
			actual := compactLayoutLabel(cell.Text)
			if actual == "" {
				continue
			}
			if actual != target && !strings.Contains(actual, target) && !strings.Contains(target, actual) {
				continue
			}
			if !found || cell.StartRow > bestCell.StartRow || (cell.StartRow == bestCell.StartRow && cell.StartCol < bestCell.StartCol) {
				bestTable = table
				bestCell = cell
				found = true
			}
		}
	}
	return bestTable, bestCell, found
}

func tableCellLayoutEvidence(table tableLayout, cell tableCellLayout) ([]types.BoundingBox, []types.TableCellRef) {
	if cell.Box == nil {
		return nil, nil
	}
	var boxes []types.BoundingBox
	boxes = append(boxes, *cell.Box)
	ref := types.TableCellRef{
		Page:          table.Page,
		TableIndex:    table.TableIndex,
		StartRow:      cell.StartRow,
		StartCol:      cell.StartCol,
		EndRow:        cell.EndRow,
		EndCol:        cell.EndCol,
		BoundingBoxes: boxes,
		Source:        "pdfparser",
	}
	return boxes, []types.TableCellRef{ref}
}

func stitchedTableCellLayoutEvidence(labelTable tableLayout, labelCell tableCellLayout, valueTable tableLayout, valueCell tableCellLayout) ([]types.BoundingBox, []types.TableCellRef) {
	labelBoxes, labelRefs := tableCellLayoutEvidence(labelTable, labelCell)
	valueBoxes, valueRefs := tableCellLayoutEvidence(valueTable, valueCell)
	if len(labelRefs) == 0 || len(valueRefs) == 0 {
		return nil, nil
	}
	boxes := append([]types.BoundingBox{}, labelBoxes...)
	boxes = append(boxes, valueBoxes...)
	refs := append([]types.TableCellRef{}, labelRefs...)
	refs = append(refs, valueRefs...)
	return boxes, refs
}

func layoutRowLabelMatches(table tableLayout, valueCell tableCellLayout, label string) bool {
	target := compactLayoutLabel(label)
	if target == "" {
		return true
	}
	actual := compactLayoutLabel(layoutRowLabelText(table, valueCell.StartRow, valueCell.StartCol))
	if actual == "" {
		return false
	}
	return actual == target || strings.Contains(actual, target) || strings.Contains(target, actual)
}

func layoutColumnLabelMatches(table tableLayout, valueCell tableCellLayout, label string) bool {
	target := compactLayoutLabel(label)
	if target == "" {
		return true
	}
	actual := compactLayoutLabel(layoutColumnLabelText(table, valueCell.StartRow, valueCell.StartCol))
	if actual == "" {
		return false
	}
	return actual == target || strings.Contains(actual, target) || strings.Contains(target, actual)
}

func layoutRowLabelText(table tableLayout, row int, beforeCol int) string {
	cells := make([]tableCellLayout, 0, len(table.Cells))
	for _, cell := range table.Cells {
		if cell.StartRow <= row && cell.EndRow >= row && cell.StartCol < beforeCol {
			cells = append(cells, cell)
		}
	}
	sort.SliceStable(cells, func(i, j int) bool {
		return cells[i].StartCol < cells[j].StartCol
	})
	parts := make([]string, 0, len(cells))
	for _, cell := range cells {
		if text := strings.TrimSpace(cell.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func layoutColumnLabelText(table tableLayout, row int, col int) string {
	cells := make([]tableCellLayout, 0, len(table.Cells))
	for _, cell := range table.Cells {
		if cell.EndRow < row && cell.StartCol <= col && cell.EndCol >= col {
			cells = append(cells, cell)
		}
	}
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].EndRow == cells[j].EndRow {
			return cells[i].StartCol < cells[j].StartCol
		}
		return cells[i].EndRow > cells[j].EndRow
	})
	parts := make([]string, 0, len(cells))
	for _, cell := range cells {
		if text := strings.TrimSpace(cell.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func layoutValueMatches(text string, targetValue string) bool {
	actual := compactLayoutValue(text)
	if actual == "" || targetValue == "" {
		return false
	}
	return actual == targetValue
}

func compactLayoutValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "*†‡§")
	value = strings.Trim(value, "()")
	value = strings.ReplaceAll(value, ",", "")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "\t", "")
	return strings.ToLower(value)
}

func compactLayoutLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = layoutNonAlnumRe.ReplaceAllString(value, "")
	return value
}

func uniquePositiveInts(values []int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
