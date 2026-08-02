package filesystem

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// SelectText applies the stable line/rune selection and binary rejection used
// by the read tool. Hosts can stream any root-bound or immutable asset source.
func SelectText(source io.Reader, startLine, maxLines, maxChars int) (selected string, selectedStart, selectedLines int, truncated bool, err error) {
	if source == nil {
		return "", startLine, 0, false, fmt.Errorf("read: source is unavailable")
	}
	if startLine < 0 {
		startLine = 0
	}
	if maxChars <= 0 {
		return "", startLine, 0, false, fmt.Errorf("read: max chars must be positive")
	}
	if maxLines == 0 {
		maxLines = -1
	}
	reader := bufio.NewReaderSize(source, 64<<10)
	head, peekErr := reader.Peek(512)
	if peekErr != nil && peekErr != io.EOF && peekErr != bufio.ErrBufferFull {
		return "", startLine, 0, false, peekErr
	}
	if IsLikelyBinary(head) {
		return "", startLine, 0, false, fmt.Errorf("binary file is not supported")
	}

	var builder strings.Builder
	lineNo := 0
	selectedRunes := 0
	lineSelected := false
	lastRuneWasNewline := false
	beginSelectedLine := func() {
		if lineSelected || lineNo < startLine || (maxLines > 0 && selectedLines >= maxLines) {
			return
		}
		if selectedLines > 0 {
			if selectedRunes < maxChars {
				builder.WriteByte('\n')
				selectedRunes++
			} else {
				truncated = true
			}
		}
		selectedLines++
		lineSelected = true
	}

	for {
		r, _, readErr := reader.ReadRune()
		if readErr != nil {
			if readErr != io.EOF {
				return "", startLine, 0, false, readErr
			}
			if lastRuneWasNewline && lineNo >= startLine && (maxLines <= 0 || selectedLines < maxLines) {
				beginSelectedLine()
			}
			break
		}
		if r == '\n' {
			if lineNo >= startLine && (maxLines <= 0 || selectedLines < maxLines) {
				beginSelectedLine()
			}
			lineNo++
			lineSelected = false
			lastRuneWasNewline = true
			if maxLines > 0 && selectedLines >= maxLines {
				break
			}
			continue
		}
		lastRuneWasNewline = false
		if lineNo < startLine || (maxLines > 0 && selectedLines >= maxLines && !lineSelected) {
			continue
		}
		beginSelectedLine()
		if selectedRunes < maxChars {
			builder.WriteRune(r)
			selectedRunes++
			continue
		}
		truncated = true
		if maxLines <= 0 {
			break
		}
	}
	return builder.String(), startLine, selectedLines, truncated, nil
}

// EditText applies the exact replacement contract without performing IO.
func EditText(content, oldText, newText string, replaceAll bool, maxOutputChars int) (string, int, error) {
	occurrences := strings.Count(content, oldText)
	if occurrences == 0 {
		return "", 0, fmt.Errorf("edit: old_string not found")
	}
	replacements := 1
	next := ""
	if replaceAll {
		replacements = occurrences
		next = strings.ReplaceAll(content, oldText, newText)
	} else {
		next = strings.Replace(content, oldText, newText, 1)
	}
	if maxOutputChars > 0 && len(next) > maxOutputChars {
		return "", 0, fmt.Errorf("edit: result too large (%d > %d)", len(next), maxOutputChars)
	}
	return next, replacements, nil
}

// IsLikelyBinary detects NUL bytes in the same bounded prefix used by read.
func IsLikelyBinary(blob []byte) bool {
	limit := len(blob)
	if limit > 512 {
		limit = 512
	}
	for i := 0; i < limit; i++ {
		if blob[i] == 0 {
			return true
		}
	}
	return false
}
