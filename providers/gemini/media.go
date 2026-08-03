package gemini

import "strings"

func parseDataURI(raw string) (string, string, bool) {
	if !strings.HasPrefix(raw, "data:") {
		return "", "", false
	}
	comma := strings.Index(raw, ",")
	if comma == -1 {
		return "", "", false
	}
	header := raw[5:comma]
	data := raw[comma+1:]
	if data == "" {
		return "", "", false
	}
	mimeType := "application/octet-stream"
	if header != "" {
		parts := strings.Split(header, ";")
		if len(parts) > 0 && parts[0] != "" {
			mimeType = parts[0]
		}
	}
	return mimeType, data, true
}
