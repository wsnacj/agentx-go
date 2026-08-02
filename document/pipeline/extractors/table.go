package extractors

import (
	"regexp"
	"strings"
	"unicode"
)

type TableInput struct {
	Text          string
	FieldKey      string
	RowLabels     []string
	ColumnLabels  []string
	ValueColumn   int
	MaxCandidates int
}

type TableResult struct {
	Value       string
	Confidence  float64
	Snippet     string
	RowLabel    string
	ColumnLabel string
	Unit        string
	UnitSource  string
	Period      string
	LineNumber  int
}

type tableLine struct {
	Raw              string
	Cells            []string
	RowLabel         string
	Values           []string
	ValueCellIndexes []int
}

var (
	tableValueRe           = regexp.MustCompile(`^\(?[+-]?(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?\)?%?$`)
	tablePeriodRe          = regexp.MustCompile(`(?i)(?:19|20)\d{2}|current|previous|comparative|本期|上期|本年|上年|本年度|上年度`)
	tableUnitRe            = regexp.MustCompile(`(?i)(rmb|cny|hkd|hk\$|usd|us\$|u\.s\.\$|\$|renminbi|yuan|dollars?|million|billion|thousand|mn|bn|单位|人民币|港币|美元|元|万元|亿元|千元|百万|百万元|千)`)
	tablePageBoundaryRe    = regexp.MustCompile(`(?i)^(?:page|p\.)\s*\d+\s*(?:of\s*\d+)?$`)
	tableCJKPageBoundaryRe = regexp.MustCompile(`^第\s*\d+\s*页$`)
)

func RunTableCandidates(in TableInput) []TableResult {
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil
	}
	rowLabels := normalizeTableRowLabels(in)
	if len(rowLabels) == 0 {
		return nil
	}
	limit := in.MaxCandidates
	if limit <= 0 {
		limit = 16
	}
	lines := strings.Split(text, "\n")
	out := make([]TableResult, 0, 4)
	recentUnit := ""
	recentUnitSource := ""
	recentUnitIndex := -1
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if unit := detectTableUnit(line); unit != "" {
			recentUnit = unit
			recentUnitSource = line
			recentUnitIndex = i
			continue
		}
		activeUnit := ""
		activeUnitSource := ""
		if recentUnit != "" && recentUnitIndex >= 0 && i-recentUnitIndex <= 12 {
			activeUnit = recentUnit
			activeUnitSource = recentUnitSource
		}
		rows := tableDataLineCandidates(lines, i)
		for _, row := range rows {
			if !matchesAnyTableLabel(row.RowLabel, rowLabels) {
				continue
			}
			header := findTableHeaderLine(lines, i)
			valueIndexes := chooseTableValueIndexes(row, header, in)
			for _, valueIndex := range valueIndexes {
				if valueIndex < 0 || valueIndex >= len(row.Values) {
					continue
				}
				value := row.Values[valueIndex]
				if isNullishTableValue(value) {
					continue
				}
				columnLabel := tableColumnLabel(row, header, valueIndex)
				confidence := 0.68
				if tableLabelsEquivalent(row.RowLabel, bestMatchedTableLabel(row.RowLabel, rowLabels)) {
					confidence += 0.08
				}
				if matchesAnyTableLabel(columnLabel, in.ColumnLabels) {
					confidence += 0.12
				}
				if activeUnit != "" {
					confidence += 0.04
				}
				if confidence > 0.95 {
					confidence = 0.95
				}
				out = append(out, TableResult{
					Value:       value,
					Confidence:  confidence,
					Snippet:     tableEvidenceSnippet(activeUnitSource, header.Raw, row.Raw),
					RowLabel:    row.RowLabel,
					ColumnLabel: columnLabel,
					Unit:        activeUnit,
					UnitSource:  activeUnitSource,
					Period:      columnLabel,
					LineNumber:  i + 1,
				})
				if len(out) >= limit {
					return out
				}
			}
		}
	}
	return out
}

func normalizeTableRowLabels(in TableInput) []string {
	labels := append([]string{}, in.RowLabels...)
	if strings.TrimSpace(in.FieldKey) != "" {
		labels = append(labels, in.FieldKey)
	}
	out := []string{}
	seen := map[string]bool{}
	for _, label := range labels {
		label = strings.TrimSpace(label)
		key := compactTableLabel(label)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, label)
	}
	return out
}

func parseTableDataLine(line string) (tableLine, bool) {
	parsed := parseTableLine(line)
	if len(parsed.Values) == 0 || strings.TrimSpace(parsed.RowLabel) == "" {
		return tableLine{}, false
	}
	return parsed, true
}

func tableDataLineCandidates(lines []string, rowIndex int) []tableLine {
	line := strings.TrimSpace(lines[rowIndex])
	if line == "" {
		return nil
	}
	out := []tableLine{}
	seen := map[string]bool{}
	appendCandidate := func(raw string) {
		row, ok := parseTableDataLine(raw)
		if !ok {
			return
		}
		key := compactTableLabel(row.RowLabel) + "|" + strings.Join(row.Values, "|")
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, row)
	}
	appendCandidate(line)
	for _, prefix := range tableWrappedLabelPrefixes(lines, rowIndex) {
		appendCandidate(prefix + "\n" + line)
	}
	return out
}

func tableWrappedLabelPrefixes(lines []string, rowIndex int) []string {
	const maxLookback = 8
	const maxSkippedSeparators = 4
	const maxLabelLines = 3
	prefixes := []string{}
	labelLines := []string{}
	skipped := 0
	skippedPageBoundary := false
	skippedHeader := false
	for i := rowIndex - 1; i >= 0 && rowIndex-i <= maxLookback; i-- {
		prev := strings.TrimSpace(lines[i])
		if prev == "" || isTablePageBoundaryLine(prev) {
			if isTablePageBoundaryLine(prev) && prev != "" {
				skippedPageBoundary = true
			}
			skipped++
			if skipped > maxSkippedSeparators {
				break
			}
			continue
		}
		if len(labelLines) == 0 && detectTableUnit(prev) != "" {
			skipped++
			if skipped > maxSkippedSeparators {
				break
			}
			continue
		}
		if len(labelLines) == 0 && looksLikeTableHeaderLine(prev) {
			skippedHeader = true
			skipped++
			if skipped > maxSkippedSeparators {
				break
			}
			continue
		}
		if !looksLikeWrappedTableLabelLine(prev) {
			break
		}
		if skippedHeader && !skippedPageBoundary {
			break
		}
		labelLines = append([]string{prev}, labelLines...)
		prefixes = append(prefixes, strings.Join(labelLines, "\n"))
		if len(labelLines) >= maxLabelLines {
			break
		}
	}
	return prefixes
}

func looksLikeTableHeaderLine(line string) bool {
	parsed := parseTableLine(line)
	return tableHeaderLooksUseful(parsed.Cells)
}

func isTablePageBoundaryLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}
	normalized := strings.ToLower(strings.Trim(line, "-_ \t"))
	if normalized == "page" || normalized == "page break" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(line), "---PAGE---") {
		return true
	}
	if tablePageBoundaryRe.MatchString(line) {
		return true
	}
	if tableCJKPageBoundaryRe.MatchString(line) {
		return true
	}
	return false
}

func looksLikeWrappedTableLabelLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || len([]rune(line)) > 120 || detectTableUnit(line) != "" {
		return false
	}
	parsed := parseTableLine(line)
	if len(parsed.Values) > 0 {
		return false
	}
	return len(parsed.Cells) > 0 && !tableHeaderLooksUseful(parsed.Cells)
}

func parseTableLine(line string) tableLine {
	raw := strings.TrimSpace(line)
	cells := tableCells(raw)
	if len(cells) == 0 {
		return tableLine{Raw: raw}
	}
	firstValue := -1
	for i, cell := range cells {
		if isTableValueToken(cell) {
			firstValue = i
			break
		}
	}
	if firstValue < 0 {
		return tableLine{Raw: raw, Cells: cells}
	}
	values := []string{}
	valueIndexes := []int{}
	for i := firstValue; i < len(cells); i++ {
		if !isTableValueToken(cells[i]) {
			continue
		}
		value, ok := cleanTableValue(cells[i])
		if !ok {
			continue
		}
		values = append(values, value)
		valueIndexes = append(valueIndexes, i)
	}
	return tableLine{
		Raw:              raw,
		Cells:            cells,
		RowLabel:         strings.Join(cells[:firstValue], " "),
		Values:           values,
		ValueCellIndexes: valueIndexes,
	}
}

func tableCells(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	if strings.Count(line, "|") >= 2 {
		line = strings.Trim(line, "|")
		parts := strings.Split(line, "|")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if cell := cleanTableCell(part); cell != "" {
				out = append(out, cell)
			}
		}
		return out
	}
	fields := strings.Fields(line)
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if cell := cleanTableCell(field); cell != "" {
			out = append(out, cell)
		}
	}
	return out
}

func cleanTableCell(cell string) string {
	cell = strings.TrimSpace(cell)
	cell = strings.Trim(cell, "|")
	return strings.TrimSpace(cell)
}

func cleanTableValue(value string) (string, bool) {
	value = cleanTableCell(value)
	if value == "" {
		return "", false
	}
	if isNullishTableValue(value) {
		return value, true
	}
	value = strings.Trim(value, "*†‡§")
	value = strings.TrimSpace(value)
	return value, value != ""
}

func isTableValueToken(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	if isNullishTableValue(token) {
		return true
	}
	return tableValueRe.MatchString(token)
}

func isNullishTableValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "-", "--", "—", "–", "－", "nil", "null", "n/a", "na", "不适用":
		return true
	default:
		return false
	}
}

func findTableHeaderLine(lines []string, rowIndex int) tableLine {
	start := rowIndex - 1
	stop := rowIndex - 12
	if stop < 0 {
		stop = 0
	}
	for i := start; i >= stop; i-- {
		raw := strings.TrimSpace(lines[i])
		if raw == "" || isTablePageBoundaryLine(raw) || detectTableUnit(raw) != "" {
			continue
		}
		candidate := parseTableLine(raw)
		if len(candidate.Cells) == 0 {
			continue
		}
		if tableHeaderLooksUseful(candidate.Cells) {
			return candidate
		}
	}
	return tableLine{}
}

func tableHeaderLooksUseful(cells []string) bool {
	if len(cells) < 2 {
		return false
	}
	for _, cell := range cells {
		if tablePeriodRe.MatchString(cell) {
			return true
		}
	}
	return false
}

func chooseTableValueIndexes(row tableLine, header tableLine, in TableInput) []int {
	if len(row.Values) == 0 {
		return nil
	}
	if in.ValueColumn > 0 {
		idx := in.ValueColumn - 1
		if idx >= 0 && idx < len(row.Values) {
			return []int{idx}
		}
	}
	if len(in.ColumnLabels) > 0 {
		out := []int{}
		for i := range row.Values {
			if matchesAnyTableLabel(tableColumnLabel(row, header, i), in.ColumnLabels) {
				out = append(out, i)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []int{0}
}

func tableColumnLabel(row tableLine, header tableLine, valueIndex int) string {
	if len(header.Cells) == 0 || valueIndex < 0 || valueIndex >= len(row.ValueCellIndexes) {
		return ""
	}
	absoluteIndex := row.ValueCellIndexes[valueIndex]
	if absoluteIndex >= 0 && absoluteIndex < len(header.Cells) {
		return header.Cells[absoluteIndex]
	}
	if len(header.Cells) == len(row.Values) {
		return header.Cells[valueIndex]
	}
	if len(header.Cells) > len(row.Values) {
		offset := len(header.Cells) - len(row.Values)
		idx := offset + valueIndex
		if idx >= 0 && idx < len(header.Cells) {
			return header.Cells[idx]
		}
	}
	return ""
}

func detectTableUnit(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || len([]rune(line)) > 160 || !tableUnitRe.MatchString(line) {
		return ""
	}
	numericTokens := 0
	for _, cell := range tableCells(line) {
		if tableValueRe.MatchString(cell) {
			numericTokens++
		}
	}
	if numericTokens > 1 {
		return ""
	}
	return line
}

func tableEvidenceSnippet(unitLine string, headerLine string, rowLine string) string {
	lines := []string{}
	for _, line := range []string{unitLine, headerLine, rowLine} {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(lines) > 0 && lines[len(lines)-1] == line {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func matchesAnyTableLabel(value string, labels []string) bool {
	for _, label := range labels {
		if tableLabelsEquivalent(value, label) {
			return true
		}
	}
	return false
}

func bestMatchedTableLabel(value string, labels []string) string {
	for _, label := range labels {
		if tableLabelsEquivalent(value, label) {
			return label
		}
	}
	return ""
}

func tableLabelsEquivalent(value string, label string) bool {
	value = strings.TrimSpace(value)
	label = strings.TrimSpace(label)
	if value == "" || label == "" {
		return false
	}
	valueCompact := compactTableLabel(value)
	labelCompact := compactTableLabel(label)
	if valueCompact == "" || labelCompact == "" {
		return false
	}
	if valueCompact == labelCompact || trimTablePlural(valueCompact) == trimTablePlural(labelCompact) {
		return true
	}
	if isLongNumericLabel(labelCompact) && strings.Contains(valueCompact, labelCompact) {
		return true
	}
	if containsCJK(label) {
		return strings.Contains(valueCompact, labelCompact)
	}
	if strings.ContainsAny(label, " \t-_/") {
		return strings.Contains(valueCompact, labelCompact)
	}
	return false
}

func compactTableLabel(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func trimTablePlural(value string) string {
	if len(value) > 3 && strings.HasSuffix(value, "s") {
		return strings.TrimSuffix(value, "s")
	}
	return value
}

func isLongNumericLabel(value string) bool {
	if len(value) < 4 {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func containsCJK(value string) bool {
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
