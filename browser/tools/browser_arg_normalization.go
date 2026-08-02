package tools

import "strings"

func browserStripControlCharacters(raw string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, raw)
}

func browserStripWrappingDelimiters(raw string) string {
	current := strings.TrimSpace(raw)
	for i := 0; i < 8; i++ {
		if len(current) < 2 {
			return current
		}
		first := current[0]
		last := current[len(current)-1]
		if (first == '"' && last == '"') ||
			(first == '\'' && last == '\'') ||
			(first == '`' && last == '`') {
			current = strings.TrimSpace(current[1 : len(current)-1])
			continue
		}
		return current
	}
	return current
}

func browserSanitizeLooseArgumentString(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	sanitized := strings.TrimSpace(browserStripControlCharacters(raw))
	if sanitized == "" {
		return ""
	}
	return browserStripWrappingDelimiters(sanitized)
}

func browserNormalizeToolToken(raw string) string {
	return strings.ToLower(strings.TrimSpace(browserSanitizeLooseArgumentString(raw)))
}

func browserCanonicalSnapshotFormat(raw string) string {
	switch normalized := browserNormalizeToolToken(raw); normalized {
	case "":
		return ""
	case "aria", "ax", "a11y", "accessible", "accessibility", "role", "roles":
		return "aria"
	case "ai", "text", "txt", "plain", "default", "minimal", "snapshot":
		return "ai"
	default:
		// Model-planned snapshot format values are internal hints rather than a
		// user-facing contract; prefer a stable default over failing the whole turn.
		return "ai"
	}
}

func browserSanitizeElementRefPayload(payload browserElementRefPayload) browserElementRefPayload {
	payload.Selector = browserSanitizeLooseArgumentString(payload.Selector)
	payload.FramePath = browserSanitizeLooseArgumentString(payload.FramePath)
	payload.NativeRef = browserSanitizeLooseArgumentString(payload.NativeRef)
	payload.Role = browserSanitizeLooseArgumentString(payload.Role)
	payload.Tag = browserSanitizeLooseArgumentString(payload.Tag)
	payload.Label = browserSanitizeLooseArgumentString(payload.Label)
	payload.Type = browserSanitizeLooseArgumentString(payload.Type)
	payload.Href = browserSanitizeLooseArgumentString(payload.Href)
	payload.Placeholder = browserSanitizeLooseArgumentString(payload.Placeholder)
	payload.PageURL = browserSanitizeLooseArgumentString(payload.PageURL)
	payload.PageOrigin = browserSanitizeLooseArgumentString(payload.PageOrigin)
	payload.PagePath = browserSanitizeLooseArgumentString(payload.PagePath)
	payload.PageTitle = browserSanitizeLooseArgumentString(payload.PageTitle)
	return payload
}
