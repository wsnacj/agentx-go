// pdf_parser_advanced.go - 高级功能扩展（清理版）

package pdfparser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PDFParserOptions 解析器选项（清理版）
type PDFParserOptions struct {
	NeedCharacter    bool   // 是否需要字符级信息
	ExtractImages    bool   // 是否提取图片
	OutputFormat     string // 输出格式: json, text, html
	PageRange        string // 页面范围: "1-5", "1,3,5", "all"
	TableEngine      string // 表格引擎: "pdfplumber", "pymupdf", "hybrid"
	HighAccuracyMode bool   // 高精度模式（启用Camelot兜底，速度较慢但精度更高）
}

// DefaultOptions 返回默认选项
func DefaultOptions() *PDFParserOptions {
	return &PDFParserOptions{
		NeedCharacter:    false,
		ExtractImages:    false,
		OutputFormat:     "json",
		PageRange:        "all",
		TableEngine:      "hybrid", // 默认使用混合模式
		HighAccuracyMode: false,    // 默认关闭高精度模式（Camelot兜底）
	}
}

// ValidateOptions 验证选项参数
func (opts *PDFParserOptions) ValidateOptions() error {
	validFormats := map[string]bool{"json": true, "text": true, "html": true}
	if !validFormats[opts.OutputFormat] {
		return fmt.Errorf("invalid output format: %s", opts.OutputFormat)
	}

	validEngines := map[string]bool{"pdfplumber": true, "pymupdf": true, "hybrid": true}
	if opts.TableEngine != "" && !validEngines[opts.TableEngine] {
		return fmt.Errorf("invalid table engine: %s", opts.TableEngine)
	}

	return nil
}

// ParsePDFWithOptions 使用自定义选项解析PDF
func (p *PDFParser) ParsePDFWithOptions(pdfPath string, opts *PDFParserOptions) (*TableResponse, error) {
	return p.ParsePDFWithOptionsContext(context.Background(), pdfPath, opts)
}

// ParsePDFWithOptionsContext 使用显式 options 并传播调用方 context。
func (p *PDFParser) ParsePDFWithOptionsContext(ctx context.Context, pdfPath string, opts *PDFParserOptions) (*TableResponse, error) {
	// 预检查
	if err := IsValidPDFPath(pdfPath); err != nil {
		return nil, err
	}

	if opts == nil {
		opts = DefaultOptions()
	}

	// 验证选项
	if err := opts.ValidateOptions(); err != nil {
		return nil, err
	}
	if p == nil || p.runner == nil {
		return nil, fmt.Errorf("python script execution failed: pdfparser runner is required")
	}
	ctx, cancel := p.withTimeout(ctx)
	defer cancel()
	result, err := p.runner.Run(ctx, RunRequest{PDFPath: pdfPath, Options: *opts})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("PDF parsing timeout after 5 minutes")
		}
		// 尝试解析stderr中的错误JSON
		if len(result.Stderr) > 0 {
			var errorResp TableResponse
			if jsonErr := json.Unmarshal(result.Stderr, &errorResp); jsonErr == nil {
				return &errorResp, nil
			}
		}

		// 记录stderr内容但不返回给用户
		if len(result.Stderr) > 0 {
			log.Printf("Python script stderr: %s", string(result.Stderr))
		}

		return nil, fmt.Errorf("python script execution failed: %w", err)
	}

	output := result.Stdout

	// 只有JSON格式才需要解析
	if opts.OutputFormat == "json" {
		var response TableResponse
		if err := json.Unmarshal(output, &response); err != nil {
			log.Printf("Failed to parse JSON response. Raw output: %s", string(output))
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}

		// 检查响应状态
		if response.Code != 0 {
			return &response, fmt.Errorf("parsing failed: %s", response.Msg)
		}

		return &response, nil
	} else {
		// 对于text和html格式，创建一个包含原始输出的响应
		return &TableResponse{
			Msg:     "success",
			Code:    0,
			Version: "1.0.0",
			Result:  TResult{Pages: []TPage{}}, // 空结构，实际内容在输出中
		}, nil
	}
}

// NewPdfplumberOptions 创建使用pdfplumber的选项
func NewPdfplumberOptions() *PDFParserOptions {
	opts := DefaultOptions()
	opts.TableEngine = "pdfplumber"
	return opts
}

// NewPyMuPDFOptions 创建只使用PyMuPDF的选项
func NewPyMuPDFOptions() *PDFParserOptions {
	opts := DefaultOptions()
	opts.TableEngine = "pymupdf"
	return opts
}

// NewHybridOptions 创建混合模式选项
func NewHybridOptions() *PDFParserOptions {
	opts := DefaultOptions()
	opts.TableEngine = "hybrid"
	return opts
}

// ParsePDFWithPdfplumber 使用pdfplumber解析PDF的便捷方法
func (p *PDFParser) ParsePDFWithPdfplumber(pdfPath string, needCharacter bool) (*TableResponse, error) {
	opts := NewPdfplumberOptions()
	opts.NeedCharacter = needCharacter
	return p.ParsePDFWithOptions(pdfPath, opts)
}

// ParsePDFWithHybrid 使用混合模式解析PDF的便捷方法
func (p *PDFParser) ParsePDFWithHybrid(pdfPath string, needCharacter bool) (*TableResponse, error) {
	opts := NewHybridOptions()
	opts.NeedCharacter = needCharacter
	return p.ParsePDFWithOptions(pdfPath, opts)
}

// TableExporter 表格导出器
type TableExporter struct {
	response *TableResponse
}

// NewTableExporter 创建表格导出器
func NewTableExporter(response *TableResponse) *TableExporter {
	if response == nil {
		log.Println("Warning: creating exporter with nil response")
		return &TableExporter{response: &TableResponse{}}
	}
	return &TableExporter{response: response}
}

// ExportToCSV 导出表格到CSV文件
func (e *TableExporter) ExportToCSV(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	exportCount := 0
	for pageIdx, page := range e.response.Result.Pages {
		tableCount := 0
		for _, table := range page.Tables {
			// 只导出真正的表格，跳过plain类型
			if table.Type == "table_with_line" || table.Type == "table_without_line" {
				tableCount++
				filename := filepath.Join(outputDir,
					fmt.Sprintf("page_%d_table_%d.csv", pageIdx+1, tableCount))

				if err := e.exportTableToCSV(table, filename); err != nil {
					return fmt.Errorf("failed to export table %d on page %d: %v", tableCount, pageIdx+1, err)
				}
				exportCount++
			}
		}
	}

	log.Printf("Exported %d tables to CSV files in %s", exportCount, outputDir)
	return nil
}

func (e *TableExporter) exportTableToCSV(table Table, filename string) error {
	if table.TableRows <= 0 || table.TableCols <= 0 {
		return fmt.Errorf("invalid table dimensions: %dx%d", table.TableRows, table.TableCols)
	}

	// 创建二维数组存储表格数据
	grid := make([][]string, table.TableRows)
	for i := range grid {
		grid[i] = make([]string, table.TableCols)
	}

	// 填充表格数据
	for _, cell := range table.TableCells {
		// 边界检查
		if cell.StartRow < 0 || cell.StartRow >= table.TableRows ||
			cell.StartCol < 0 || cell.StartCol >= table.TableCols {
			log.Printf("Warning: cell position out of bounds: [%d,%d] in %dx%d table",
				cell.StartRow, cell.StartCol, table.TableRows, table.TableCols)
			continue
		}

		// 处理跨行跨列的情况
		for r := cell.StartRow; r <= cell.EndRow && r < table.TableRows; r++ {
			for c := cell.StartCol; c <= cell.EndCol && c < table.TableCols; c++ {
				grid[r][c] = cell.Text
			}
		}
	}

	// 写入CSV文件
	var buf bytes.Buffer
	for _, row := range grid {
		for i, cell := range row {
			// 转义特殊字符
			cell = strings.ReplaceAll(cell, "\"", "\"\"")
			if strings.ContainsAny(cell, ",\n\"") {
				cell = "\"" + cell + "\""
			}
			buf.WriteString(cell)
			if i < len(row)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteString("\n")
	}

	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// ExportToHTML 导出到HTML格式，参考table.go中的buildPageHTML方法
func (e *TableExporter) ExportToHTML(outputFile string) error {
	var html bytes.Buffer

	// HTML头部
	html.WriteString(`<!DOCTYPE html>
<html>
<body>
<h1>PDF Document Analysis</h1>
`)

	// 遍历页面
	for pageIdx, page := range e.response.Result.Pages {
		html.WriteString(fmt.Sprintf(`<div class="page-header">Page %d (宽度: %d, 高度: %d)</div>`,
			pageIdx+1, page.Width, page.Height))

		// 遍历所有元素
		elemCount := 0
		for _, table := range page.Tables {
			elemCount++
			switch table.Type {
			case "plain":
				// 输出文本内容
				html.WriteString(fmt.Sprintf(`<div class="text-content">
					<h4>Text Block %d</h4>`, elemCount))
				html.WriteString("<ul>")
				for _, line := range table.Lines {
					if strings.TrimSpace(line.Text) != "" {
						html.WriteString(fmt.Sprintf("<li>%s</li>", escapeHTML(line.Text)))
					}
				}
				html.WriteString("</ul>")
				html.WriteString(`</div>`)

			case "table_with_line", "table_without_line":
				// 输出表格
				tableType := "Table"
				engineInfo := ""
				if table.Type == "table_without_line" {
					tableType = "Table (no borders)"
					engineInfo = " <span class=\"engine-info\">[text-based detection]</span>"
				} else {
					engineInfo = " <span class=\"engine-info\">[line-based detection]</span>"
				}
				html.WriteString(fmt.Sprintf(`<h3>%s %d%s</h3>`, tableType, elemCount, engineInfo))
				html.WriteString(fmt.Sprintf(`<p class="table-info">Size: %d×%d, Cells: %d</p>`,
					table.TableRows, table.TableCols, len(table.TableCells)))

				// 注意：对于有table_cells的表格，不再输出table.Lines，避免重复显示
				// 只有当表格没有cells或cells为空时，才显示lines（这种情况通常是检测错误）
				if len(table.TableCells) == 0 && len(table.Lines) > 0 {
					html.WriteString("<div style='color: #888; font-style: italic;'>表格检测到文本但未形成单元格结构：</div>")
					html.WriteString("<ul>")
					for _, line := range table.Lines {
						if strings.TrimSpace(line.Text) != "" {
							html.WriteString(fmt.Sprintf("<li>%s</li>", escapeHTML(line.Text)))
						}
					}
					html.WriteString("</ul>")
				}

				if table.TableRows > 0 && table.TableCols > 0 {
					html.WriteString(`<table>`)

					// 利用 cellMap 处理合并单元格，参考table.go的逻辑
					cellMap := make(map[int]map[int]*TableCell)
					for _, cell := range table.TableCells {
						if _, exists := cellMap[cell.StartRow]; !exists {
							cellMap[cell.StartRow] = make(map[int]*TableCell)
						}
						cellMap[cell.StartRow][cell.StartCol] = &cell
					}

					// 生成HTML表格，处理合并单元格
					for row := 0; row < table.TableRows; row++ {
						html.WriteString("<tr>")
						for col := 0; col < table.TableCols; col++ {
							if cell, exists := cellMap[row][col]; exists {
								rowSpan := cell.EndRow - cell.StartRow + 1
								colSpan := cell.EndCol - cell.StartCol + 1
								// 只在起始位置输出单元格
								if row == cell.StartRow && col == cell.StartCol {
									cellText := escapeHTML(cell.Text)
									if strings.TrimSpace(cellText) == "" {
										cellText = "&nbsp;"
									}
									html.WriteString(fmt.Sprintf("<td rowspan='%d' colspan='%d'>%s</td>",
										rowSpan, colSpan, cellText))
								}
							} else {
								// 检查是否被合并单元格覆盖
								covered := false
								for _, rMap := range cellMap {
									for _, mergedCell := range rMap {
										if row >= mergedCell.StartRow && row <= mergedCell.EndRow &&
											col >= mergedCell.StartCol && col <= mergedCell.EndCol {
											covered = true
											break
										}
									}
									if covered {
										break
									}
								}
								// 如果没有被覆盖，输出空单元格
								if !covered {
									html.WriteString(`<td class="empty-cell">&nbsp;</td>`)
								}
							}
						}
						html.WriteString("</tr>")
					}

					html.WriteString(`</table>`)
				}
			}
		}
	}

	// HTML尾部
	html.WriteString(`</body></html>`)

	return os.WriteFile(outputFile, html.Bytes(), 0644)
}

// 辅助函数：HTML转义
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// TextAnalyzer 文本分析器
type TextAnalyzer struct {
	response *TableResponse
}

// NewTextAnalyzer 创建文本分析器
func NewTextAnalyzer(response *TableResponse) *TextAnalyzer {
	if response == nil {
		log.Println("Warning: creating analyzer with nil response")
		return &TextAnalyzer{response: &TableResponse{}}
	}
	return &TextAnalyzer{response: response}
}

// AnalyzeStructure 分析文档结构
func (a *TextAnalyzer) AnalyzeStructure() *DocumentStructure {
	structure := &DocumentStructure{
		TotalPages:   len(a.response.Result.Pages),
		TotalTables:  0,
		TotalLines:   0,
		TableContent: make([]TableSummary, 0),
		TextBlocks:   make([]TextBlock, 0),
	}

	for pageIdx, page := range a.response.Result.Pages {
		// 分析每个元素
		for tableIdx, table := range page.Tables {
			switch table.Type {
			case "plain":
				// 统计文本行
				structure.TotalLines += len(table.Lines)

				// 创建文本块
				if len(table.Lines) > 0 {
					block := TextBlock{
						PageNumber:    pageIdx + 1,
						ElementNumber: tableIdx + 1,
						LineCount:     len(table.Lines),
						Lines:         table.Lines,
						Position:      table.Position,
					}
					structure.TextBlocks = append(structure.TextBlocks, block)
				}

			case "table_with_line", "table_without_line":
				// 统计表格
				structure.TotalTables++

				// 创建表格摘要
				summary := TableSummary{
					PageNumber:    pageIdx + 1,
					ElementNumber: tableIdx + 1,
					Rows:          table.TableRows,
					Cols:          table.TableCols,
					Position:      table.Position,
					CellCount:     len(table.TableCells),
					TableType:     table.Type,
				}
				structure.TableContent = append(structure.TableContent, summary)
			}
		}
	}

	return structure
}

// DocumentStructure 文档结构分析结果
type DocumentStructure struct {
	TotalPages   int            `json:"total_pages"`
	TotalTables  int            `json:"total_tables"`
	TotalLines   int            `json:"total_lines"`
	TableContent []TableSummary `json:"table_content"`
	TextBlocks   []TextBlock    `json:"text_blocks"`
}

// TableSummary 表格摘要
type TableSummary struct {
	PageNumber    int    `json:"page_number"`
	ElementNumber int    `json:"element_number"`
	Rows          int    `json:"rows"`
	Cols          int    `json:"cols"`
	Position      []int  `json:"position"`
	CellCount     int    `json:"cell_count"`
	TableType     string `json:"table_type"`
}

// TextBlock 文本块
type TextBlock struct {
	PageNumber    int     `json:"page_number"`
	ElementNumber int     `json:"element_number"`
	LineCount     int     `json:"line_count"`
	Lines         []TLine `json:"lines"`
	Position      []int   `json:"position"`
}

// SearchText 在文档中搜索文本
func (a *TextAnalyzer) SearchText(keyword string) []SearchResult {
	results := make([]SearchResult, 0)
	if keyword == "" {
		return results
	}

	keyword = strings.ToLower(keyword)

	for pageIdx, page := range a.response.Result.Pages {
		for tableIdx, table := range page.Tables {
			switch table.Type {
			case "plain":
				// 搜索普通文本
				for lineIdx, line := range table.Lines {
					if strings.Contains(strings.ToLower(line.Text), keyword) {
						results = append(results, SearchResult{
							PageNumber:    pageIdx + 1,
							ElementNumber: tableIdx + 1,
							LineNumber:    lineIdx + 1,
							Text:          line.Text,
							Position:      line.Position,
							InTable:       false,
							ElementType:   "plain",
						})
					}
				}

			case "table_with_line", "table_without_line":
				// 搜索表格内容
				for _, cell := range table.TableCells {
					if strings.Contains(strings.ToLower(cell.Text), keyword) {
						results = append(results, SearchResult{
							PageNumber:    pageIdx + 1,
							ElementNumber: tableIdx + 1,
							TableNumber:   tableIdx + 1,
							CellRow:       cell.StartRow,
							CellCol:       cell.StartCol,
							Text:          cell.Text,
							Position:      cell.Position,
							InTable:       true,
							ElementType:   table.Type,
						})
					}
				}
			}
		}
	}

	return results
}

// SearchResult 搜索结果
type SearchResult struct {
	PageNumber    int    `json:"page_number"`
	ElementNumber int    `json:"element_number"`        // 元素序号（在页面中的顺序）
	LineNumber    int    `json:"line_number,omitempty"` // 在plain元素中的行号
	TableNumber   int    `json:"table_number,omitempty"`
	CellRow       int    `json:"cell_row,omitempty"`
	CellCol       int    `json:"cell_col,omitempty"`
	Text          string `json:"text"`
	Position      []int  `json:"position"`
	InTable       bool   `json:"in_table"`
	ElementType   string `json:"element_type"` // plain, table_with_line, table_without_line
}
