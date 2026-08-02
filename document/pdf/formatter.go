// text_formatter.go - 文本规整器，将PDF解析结果按行规整输出

package pdfparser

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

var (
	// 匹配任意汉字或数字之间的空格
	reCJKNumSpace = regexp.MustCompile(`([\p{Han}\p{Nd}])\s+([\p{Han}\p{Nd}])`)
	// 匹配连续多个空格
	reMultiSP = regexp.MustCompile(`\s{2,}`)
)

// TextElement 表示一个可排序的文本元素
type TextElement struct {
	Text     string  // 文本内容
	X        float64 // X坐标
	Y        float64 // Y坐标
	Width    float64 // 宽度
	Height   float64 // 高度
	FontSize float64 // 字体大小
	IsTable  bool    // 是否为表格
}

// LineGroup 表示一行文本
type LineGroup struct {
	Y        float64       // 行的Y坐标
	Elements []TextElement // 该行的所有元素
}

// TextFormatter 文本格式化器
type TextFormatter struct {
	response *TableResponse
}

// NewTextFormatter 创建文本格式化器
func NewTextFormatter(response *TableResponse) *TextFormatter {
	if response == nil {
		return &TextFormatter{response: &TableResponse{}}
	}
	return &TextFormatter{response: response}
}

// FormatToText 将TableResponse格式化为按行规整的文本
func (f *TextFormatter) FormatToText() string {
	if f.response == nil || len(f.response.Result.Pages) == 0 {
		return ""
	}

	var pageTexts []string
	for pageIdx, page := range f.response.Result.Pages {
		pageText := f.formatPageToText(page, pageIdx+1)
		if pageText != "" {
			pageTexts = append(pageTexts, pageText)
		}
	}

	return strings.Join(pageTexts, "\n\n")
}

// formatPageToText 格式化单个页面的文本
func (f *TextFormatter) formatPageToText(page TPage, pageNum int) string {
	if len(page.Tables) == 0 {
		return ""
	}

	// 1. 收集所有文本元素
	elements := f.collectTextElements(page)
	if len(elements) == 0 {
		return ""
	}

	// 2. 计算分行阈值
	threshold := f.calculateLineThreshold(elements)

	// 3. 按行分组
	lineGroups := f.groupElementsByLine(elements, threshold)

	// 4. 排序并格式化输出
	return f.formatLineGroups(lineGroups)
}

// collectTextElements 收集页面中的所有文本元素
func (f *TextFormatter) collectTextElements(page TPage) []TextElement {
	var elements []TextElement

	for _, table := range page.Tables {
		switch table.Type {
		case "plain":
			// 处理普通文本
			for _, line := range table.Lines {
				if strings.TrimSpace(line.Text) == "" {
					continue
				}

				element := TextElement{
					Text:     line.Text,
					X:        float64(line.Position[0]),
					Y:        float64(line.Position[1]),
					Width:    float64(line.Position[2] - line.Position[0]),
					Height:   float64(line.Position[3] - line.Position[1]),
					FontSize: f.estimateFontSize(line),
					IsTable:  false,
				}
				elements = append(elements, element)
			}

		case "table_with_line", "table_without_line":
			// 处理表格 - 按行处理
			tableText := f.formatTableToText(table)
			if tableText != "" {
				element := TextElement{
					Text:     tableText,
					X:        float64(table.Position[0]),
					Y:        float64(table.Position[1]),
					Width:    float64(table.Position[2] - table.Position[0]),
					Height:   float64(table.Position[3] - table.Position[1]),
					FontSize: f.estimateTableFontSize(table),
					IsTable:  true,
				}
				elements = append(elements, element)
			}
		}
	}

	return elements
}

// formatTableToText 将表格格式化为文本，每行用换行符分隔，单元格用\t分隔
func (f *TextFormatter) formatTableToText(table Table) string {
	if table.TableRows <= 0 || table.TableCols <= 0 {
		return ""
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
			continue
		}

		// 处理跨行跨列的情况 - 只在起始位置填充文本，避免重复
		if cell.StartRow == cell.EndRow && cell.StartCol == cell.EndCol {
			// 普通单元格，直接填充
			grid[cell.StartRow][cell.StartCol] = strings.TrimSpace(cell.Text)
		} else {
			// 合并单元格，只在起始位置填充文本，其他位置标记为占位符
			grid[cell.StartRow][cell.StartCol] = strings.TrimSpace(cell.Text)
			// 其他被合并的位置标记为空字符串但已占用
			for r := cell.StartRow; r <= cell.EndRow && r < table.TableRows; r++ {
				for c := cell.StartCol; c <= cell.EndCol && c < table.TableCols; c++ {
					if !(r == cell.StartRow && c == cell.StartCol) {
						grid[r][c] = "" // 合并单元格的非起始位置留空
					}
				}
			}
		}
	}

	// 将表格转换为文本
	var rows []string
	for _, row := range grid {
		// 清理每个单元格的文本
		cleanedRow := make([]string, len(row))
		for i, cell := range row {
			cleanedRow[i] = f.cleanCellText(cell)
		}

		// 用制表符连接单元格
		rowText := strings.Join(cleanedRow, "\t")

		// 只添加非空行
		if strings.TrimSpace(strings.ReplaceAll(rowText, "\t", "")) != "" {
			rows = append(rows, rowText)
		}
	}

	return strings.Join(rows, "\n")
}

// cleanCellText 清理单元格文本
func (f *TextFormatter) cleanCellText(text string) string {
	if text == "" {
		return ""
	}

	// 去除多余的空格和换行
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// 去掉汉字/数字间的空格
	text = reCJKNumSpace.ReplaceAllString(text, "$1$2")

	// 合并多个空格
	text = reMultiSP.ReplaceAllString(text, " ")

	return text
}

// estimateFontSize 估算字体大小
func (f *TextFormatter) estimateFontSize(line TLine) float64 {
	if len(line.Position) >= 4 {
		height := float64(line.Position[3] - line.Position[1])
		if height > 0 {
			return height * 0.8 // 估算字体大小约为行高的80%
		}
	}
	return 12.0 // 默认字体大小
}

// estimateTableFontSize 估算表格字体大小
func (f *TextFormatter) estimateTableFontSize(table Table) float64 {
	if len(table.Lines) == 0 {
		return 12.0
	}

	// 取第一个非空行的字体大小
	for _, line := range table.Lines {
		if strings.TrimSpace(line.Text) != "" {
			return f.estimateFontSize(line)
		}
	}

	return 12.0
}

// calculateLineThreshold 计算分行阈值
func (f *TextFormatter) calculateLineThreshold(elements []TextElement) float64 {
	if len(elements) == 0 {
		return 5.0
	}

	// 收集所有字体大小
	sizes := make([]float64, len(elements))
	for i, elem := range elements {
		sizes[i] = elem.FontSize
	}

	// 计算中位数
	sort.Float64s(sizes)
	median := sizes[len(sizes)/2]

	// 阈值为中位字体大小的30%
	threshold := median * 0.3
	if threshold < 2.0 {
		threshold = 2.0
	}

	return threshold
}

// groupElementsByLine 按行分组元素
func (f *TextFormatter) groupElementsByLine(elements []TextElement, threshold float64) []LineGroup {
	if len(elements) == 0 {
		return nil
	}

	var groups []LineGroup

	for _, elem := range elements {
		placed := false

		// 尝试加入现有行
		for i := range groups {
			if math.Abs(elem.Y-groups[i].Y) <= threshold {
				groups[i].Elements = append(groups[i].Elements, elem)
				placed = true
				break
			}
		}

		// 如果没有合适的行，创建新行
		if !placed {
			groups = append(groups, LineGroup{
				Y:        elem.Y,
				Elements: []TextElement{elem},
			})
		}
	}

	return groups
}

// formatLineGroups 格式化行组为最终文本
func (f *TextFormatter) formatLineGroups(lineGroups []LineGroup) string {
	if len(lineGroups) == 0 {
		return ""
	}

	// 按Y坐标排序（从上到下）
	sort.Slice(lineGroups, func(i, j int) bool {
		return lineGroups[i].Y < lineGroups[j].Y
	})

	var lines []string

	for _, group := range lineGroups {
		// 每行内部按X坐标排序（从左到右）
		sort.Slice(group.Elements, func(i, j int) bool {
			return group.Elements[i].X < group.Elements[j].X
		})

		// 组合该行的文本
		var lineParts []string
		for _, elem := range group.Elements {
			if elem.IsTable {
				// 表格直接添加，已经包含换行
				lineParts = append(lineParts, elem.Text)
			} else {
				// 普通文本清理后添加
				cleanText := f.cleanText(elem.Text)
				if cleanText != "" {
					lineParts = append(lineParts, cleanText)
				}
			}
		}

		if len(lineParts) > 0 {
			lineText := strings.Join(lineParts, " ")
			lines = append(lines, lineText)
		}
	}

	return strings.Join(lines, "\n")
}

// cleanText 清理普通文本
func (f *TextFormatter) cleanText(text string) string {
	if text == "" {
		return ""
	}

	// 去除多余的空格和换行
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// 去掉汉字/数字间的空格
	text = reCJKNumSpace.ReplaceAllString(text, "$1$2")

	// 合并多个空格
	text = reMultiSP.ReplaceAllString(text, " ")

	return text
}

// FormatToLines 将结果格式化为行数组
func (f *TextFormatter) FormatToLines() []string {
	text := f.FormatToText()
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")

	// 过滤空行
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return result
}

// FormatPageToLines 格式化指定页面为行数组
func (f *TextFormatter) FormatPageToLines(pageNum int) []string {
	if f.response == nil || pageNum < 1 || pageNum > len(f.response.Result.Pages) {
		return nil
	}

	page := f.response.Result.Pages[pageNum-1]
	pageText := f.formatPageToText(page, pageNum)

	if pageText == "" {
		return nil
	}

	lines := strings.Split(pageText, "\n")

	// 过滤空行
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return result
}

// FormatToPages 将TableResponse格式化为按页面分组的文本数组
func (f *TextFormatter) FormatToPages() []string {
	if f.response == nil || len(f.response.Result.Pages) == 0 {
		return nil
	}

	var pageTexts []string
	for pageIdx, page := range f.response.Result.Pages {
		pageText := f.formatPageToText(page, pageIdx+1)
		pageTexts = append(pageTexts, pageText)
	}

	return pageTexts
}

// FormatToPagesWithEmptyFilter 将TableResponse格式化为按页面分组的文本数组，可选择是否包含空页面
func (f *TextFormatter) FormatToPagesWithEmptyFilter(includeEmpty bool) []string {
	if f.response == nil || len(f.response.Result.Pages) == 0 {
		return nil
	}

	var pageTexts []string
	for pageIdx, page := range f.response.Result.Pages {
		pageText := f.formatPageToText(page, pageIdx+1)

		if includeEmpty || strings.TrimSpace(pageText) != "" {
			pageTexts = append(pageTexts, pageText)
		}
	}

	return pageTexts
}

// ExtractTextOnly 提取纯文本内容（不包含表格格式）
func (f *TextFormatter) ExtractTextOnly() string {
	if f.response == nil || len(f.response.Result.Pages) == 0 {
		return ""
	}

	var pageTexts []string

	for _, page := range f.response.Result.Pages {
		var elements []TextElement

		// 只收集普通文本元素
		for _, table := range page.Tables {
			if table.Type == "plain" {
				for _, line := range table.Lines {
					if strings.TrimSpace(line.Text) == "" {
						continue
					}

					element := TextElement{
						Text:     line.Text,
						X:        float64(line.Position[0]),
						Y:        float64(line.Position[1]),
						FontSize: f.estimateFontSize(line),
					}
					elements = append(elements, element)
				}
			}
		}

		if len(elements) > 0 {
			threshold := f.calculateLineThreshold(elements)
			lineGroups := f.groupElementsByLine(elements, threshold)
			pageText := f.formatLineGroupsTextOnly(lineGroups)

			if pageText != "" {
				pageTexts = append(pageTexts, pageText)
			}
		}
	}

	return strings.Join(pageTexts, "\n\n")
}

// formatLineGroupsTextOnly 格式化行组为纯文本（无表格）
func (f *TextFormatter) formatLineGroupsTextOnly(lineGroups []LineGroup) string {
	if len(lineGroups) == 0 {
		return ""
	}

	// 按Y坐标排序（从上到下）
	sort.Slice(lineGroups, func(i, j int) bool {
		return lineGroups[i].Y < lineGroups[j].Y
	})

	var lines []string

	for _, group := range lineGroups {
		// 每行内部按X坐标排序（从左到右）
		sort.Slice(group.Elements, func(i, j int) bool {
			return group.Elements[i].X < group.Elements[j].X
		})

		// 组合该行的文本
		var lineParts []string
		for _, elem := range group.Elements {
			cleanText := f.cleanText(elem.Text)
			if cleanText != "" {
				lineParts = append(lineParts, cleanText)
			}
		}

		if len(lineParts) > 0 {
			lineText := strings.Join(lineParts, " ")
			lines = append(lines, lineText)
		}
	}

	return strings.Join(lines, "\n")
}
