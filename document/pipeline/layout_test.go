package pipeline

import (
	"github.com/wsnacj/agentx-go/document/pdf"
	"github.com/wsnacj/agentx-go/document/pipeline/extractors"
	"testing"
)

func TestTableLayoutEvidenceForResultUsesPDFParserCellPosition(t *testing.T) {
	resp := &pdfparser.TableResponse{
		Result: pdfparser.TResult{
			Pages: []pdfparser.TPage{{
				Width:  600,
				Height: 800,
				Tables: []pdfparser.Table{{
					Type:      "table_with_line",
					TableRows: 2,
					TableCols: 3,
					Position:  []int{10, 20, 300, 120},
					TableCells: []pdfparser.TableCell{
						{StartRow: 0, StartCol: 0, EndRow: 0, EndCol: 0, Text: "Item", Position: []int{10, 20, 80, 40}},
						{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 1, Text: "2025", Position: []int{80, 20, 160, 40}},
						{StartRow: 0, StartCol: 2, EndRow: 0, EndCol: 2, Text: "2024", Position: []int{160, 20, 240, 40}},
						{StartRow: 1, StartCol: 0, EndRow: 1, EndCol: 0, Text: "Revenue", Position: []int{10, 40, 80, 60}},
						{StartRow: 1, StartCol: 1, EndRow: 1, EndCol: 1, Text: "100", Position: []int{80, 40, 160, 60}},
						{StartRow: 1, StartCol: 2, EndRow: 1, EndCol: 2, Text: "90", Position: []int{160, 40, 240, 60}},
					},
				}},
			}},
		},
	}
	layout := buildDocumentLayoutFromPDFResponse(resp)
	boxes, refs := tableLayoutEvidenceForResult(layout, []int{1}, extractors.TableResult{
		Value:       "100",
		RowLabel:    "Revenue",
		ColumnLabel: "2025",
	})

	if len(boxes) != 1 || boxes[0].X0 != 80 || boxes[0].Y0 != 40 || boxes[0].Source != "pdfparser" {
		t.Fatalf("expected pdfparser cell bbox evidence, got %#v", boxes)
	}
	if len(refs) != 1 || refs[0].TableIndex != 1 || refs[0].StartRow != 1 || refs[0].StartCol != 1 {
		t.Fatalf("expected pdfparser table cell reference, got %#v", refs)
	}
}

func TestTableLayoutEvidenceForResultRequiresRealLayout(t *testing.T) {
	boxes, refs := tableLayoutEvidenceForResult(nil, []int{1}, extractors.TableResult{
		Value:    "100",
		RowLabel: "Revenue",
	})
	if len(boxes) != 0 || len(refs) != 0 {
		t.Fatalf("expected no evidence without layout source, got boxes=%#v refs=%#v", boxes, refs)
	}
}

func TestTableLayoutEvidenceForResultDoesNotGuessMissingColumnLabel(t *testing.T) {
	resp := &pdfparser.TableResponse{
		Result: pdfparser.TResult{
			Pages: []pdfparser.TPage{{
				Tables: []pdfparser.Table{{
					Type:      "table_with_line",
					TableRows: 1,
					TableCols: 2,
					TableCells: []pdfparser.TableCell{
						{StartRow: 0, StartCol: 0, EndRow: 0, EndCol: 0, Text: "Revenue", Position: []int{10, 40, 80, 60}},
						{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 1, Text: "100", Position: []int{80, 40, 160, 60}},
					},
				}},
			}},
		},
	}
	layout := buildDocumentLayoutFromPDFResponse(resp)
	boxes, refs := tableLayoutEvidenceForResult(layout, []int{1}, extractors.TableResult{
		Value:       "100",
		RowLabel:    "Revenue",
		ColumnLabel: "2025",
	})
	if len(boxes) != 0 || len(refs) != 0 {
		t.Fatalf("expected no bbox when requested column label is not present, got boxes=%#v refs=%#v", boxes, refs)
	}
}

func TestTableLayoutEvidenceForResultStitchesAdjacentPageLabelAndValueCells(t *testing.T) {
	resp := &pdfparser.TableResponse{
		Result: pdfparser.TResult{
			Pages: []pdfparser.TPage{
				{
					Tables: []pdfparser.Table{{
						Type:      "table_with_line",
						TableRows: 2,
						TableCols: 3,
						TableCells: []pdfparser.TableCell{
							{StartRow: 0, StartCol: 0, EndRow: 0, EndCol: 0, Text: "Item", Position: []int{10, 20, 80, 40}},
							{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 1, Text: "2025", Position: []int{80, 20, 160, 40}},
							{StartRow: 0, StartCol: 2, EndRow: 0, EndCol: 2, Text: "2024", Position: []int{160, 20, 240, 40}},
							{StartRow: 1, StartCol: 0, EndRow: 1, EndCol: 0, Text: "Operating cash flow", Position: []int{10, 40, 160, 60}},
						},
					}},
				},
				{
					Tables: []pdfparser.Table{{
						Type:      "table_with_line",
						TableRows: 2,
						TableCols: 3,
						TableCells: []pdfparser.TableCell{
							{StartRow: 0, StartCol: 0, EndRow: 0, EndCol: 0, Text: "Item", Position: []int{10, 20, 80, 40}},
							{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 1, Text: "2025", Position: []int{80, 20, 160, 40}},
							{StartRow: 0, StartCol: 2, EndRow: 0, EndCol: 2, Text: "2024", Position: []int{160, 20, 240, 40}},
							{StartRow: 1, StartCol: 1, EndRow: 1, EndCol: 1, Text: "100", Position: []int{80, 40, 160, 60}},
							{StartRow: 1, StartCol: 2, EndRow: 1, EndCol: 2, Text: "90", Position: []int{160, 40, 240, 60}},
						},
					}},
				},
			},
		},
	}
	layout := buildDocumentLayoutFromPDFResponse(resp)
	boxes, refs := tableLayoutEvidenceForResult(layout, []int{2, 1}, extractors.TableResult{
		Value:       "100",
		RowLabel:    "Operating cash flow",
		ColumnLabel: "2025",
	})

	if len(boxes) != 2 || boxes[0].Page != 1 || boxes[1].Page != 2 {
		t.Fatalf("expected stitched label/value bounding boxes, got %#v", boxes)
	}
	if len(refs) != 2 || refs[0].Page != 1 || refs[0].StartCol != 0 || refs[1].Page != 2 || refs[1].StartCol != 1 {
		t.Fatalf("expected stitched label/value table cell refs, got %#v", refs)
	}
}

func TestTableLayoutEvidenceForResultDoesNotStitchWithoutColumnEvidence(t *testing.T) {
	resp := &pdfparser.TableResponse{
		Result: pdfparser.TResult{
			Pages: []pdfparser.TPage{
				{Tables: []pdfparser.Table{{
					Type: "table_with_line",
					TableCells: []pdfparser.TableCell{
						{StartRow: 1, StartCol: 0, EndRow: 1, EndCol: 0, Text: "Operating cash flow", Position: []int{10, 40, 160, 60}},
					},
				}}},
				{Tables: []pdfparser.Table{{
					Type: "table_with_line",
					TableCells: []pdfparser.TableCell{
						{StartRow: 1, StartCol: 1, EndRow: 1, EndCol: 1, Text: "100", Position: []int{80, 40, 160, 60}},
					},
				}}},
			},
		},
	}
	layout := buildDocumentLayoutFromPDFResponse(resp)
	boxes, refs := tableLayoutEvidenceForResult(layout, []int{1, 2}, extractors.TableResult{
		Value:    "100",
		RowLabel: "Operating cash flow",
	})
	if len(boxes) != 0 || len(refs) != 0 {
		t.Fatalf("expected no stitched evidence without column label, got boxes=%#v refs=%#v", boxes, refs)
	}
}
