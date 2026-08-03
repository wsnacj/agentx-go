package ark

import (
	"encoding/json"
	"strings"
)

// ToolResultPayload provides a structured tool result view.
type ToolResultPayload struct {
	Result  ToolResult
	Text    string
	Payload map[string]any
}

// WebSearchItem captures a normalized search hit.
type WebSearchItem struct {
	Title       string
	URL         string
	Snippet     string
	Source      string
	PublishedAt string
	Raw         map[string]any
}

// WebSearchResult represents web_search output.
type WebSearchResult struct {
	ToolResultPayload
	Results []map[string]any
	Items   []WebSearchItem
}

// KnowledgeSearchChunk captures a normalized knowledge hit.
type KnowledgeSearchChunk struct {
	DocID  string
	Title  string
	Text   string
	Score  float64
	Source string
	Raw    map[string]any
}

// KnowledgeSearchResult represents knowledge_search output.
type KnowledgeSearchResult struct {
	ToolResultPayload
	Results []map[string]any
	Chunks  []KnowledgeSearchChunk
}

// BoundingBox captures normalized box data.
type BoundingBox struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
	W  float64
	H  float64
}

// ImageProcessResult represents image_process output.
type ImageProcessResult struct {
	ToolResultPayload
	Artifacts []map[string]any
	Boxes     []BoundingBox
}

// MCPResult represents mcp output.
type MCPResult struct {
	ToolResultPayload
}

// DoubaoAppResult represents doubao_app output.
type DoubaoAppResult struct {
	ToolResultPayload
}

// ToolResultText joins output_text content for a tool result.
func ToolResultText(result ToolResult) string {
	var out strings.Builder
	for _, content := range result.Content {
		if content.Text == "" {
			continue
		}
		out.WriteString(content.Text)
	}
	if result.Extra != nil {
		if value, ok := result.Extra["content"]; ok {
			if text := stringifyContent(value); text != "" {
				out.WriteString(text)
			}
		}
	}
	return out.String()
}

// ToolResultPayloadMap returns the raw payload map for a tool result.
func ToolResultPayloadMap(result ToolResult) map[string]any {
	if result.Extra == nil {
		return nil
	}
	out := make(map[string]any, len(result.Extra))
	for k, v := range result.Extra {
		out[k] = v
	}
	return out
}

// ParseWebSearchResult converts a ToolResult into a WebSearchResult.
func ParseWebSearchResult(result ToolResult) WebSearchResult {
	payload := ToolResultPayload{
		Result:  result,
		Text:    ToolResultText(result),
		Payload: ToolResultPayloadMap(result),
	}
	results := extractResults(payload.Payload)
	return WebSearchResult{
		ToolResultPayload: payload,
		Results:           results,
		Items:             parseWebSearchItems(results),
	}
}

// ParseKnowledgeSearchResult converts a ToolResult into a KnowledgeSearchResult.
func ParseKnowledgeSearchResult(result ToolResult) KnowledgeSearchResult {
	payload := ToolResultPayload{
		Result:  result,
		Text:    ToolResultText(result),
		Payload: ToolResultPayloadMap(result),
	}
	results := extractResults(payload.Payload)
	return KnowledgeSearchResult{
		ToolResultPayload: payload,
		Results:           results,
		Chunks:            parseKnowledgeChunks(results),
	}
}

// ParseImageProcessResult converts a ToolResult into an ImageProcessResult.
func ParseImageProcessResult(result ToolResult) ImageProcessResult {
	payload := ToolResultPayload{
		Result:  result,
		Text:    ToolResultText(result),
		Payload: ToolResultPayloadMap(result),
	}
	artifacts := extractArtifacts(payload.Payload)
	return ImageProcessResult{
		ToolResultPayload: payload,
		Artifacts:         artifacts,
		Boxes:             parseBoundingBoxes(artifacts),
	}
}

// ParseMCPResult converts a ToolResult into an MCPResult.
func ParseMCPResult(result ToolResult) MCPResult {
	payload := ToolResultPayload{
		Result:  result,
		Text:    ToolResultText(result),
		Payload: ToolResultPayloadMap(result),
	}
	return MCPResult{ToolResultPayload: payload}
}

// ParseDoubaoAppResult converts a ToolResult into a DoubaoAppResult.
func ParseDoubaoAppResult(result ToolResult) DoubaoAppResult {
	payload := ToolResultPayload{
		Result:  result,
		Text:    ToolResultText(result),
		Payload: ToolResultPayloadMap(result),
	}
	return DoubaoAppResult{ToolResultPayload: payload}
}

func extractResults(payload map[string]any) []map[string]any {
	if payload == nil {
		return nil
	}
	for _, key := range []string{"results", "data", "items", "chunks", "documents"} {
		if value, ok := payload[key]; ok {
			return asMapSlice(value)
		}
	}
	return nil
}

func extractArtifacts(payload map[string]any) []map[string]any {
	if payload == nil {
		return nil
	}
	for _, key := range []string{"artifacts", "outputs", "boxes", "points", "lines"} {
		if value, ok := payload[key]; ok {
			return asMapSlice(value)
		}
	}
	return nil
}

func parseWebSearchItems(results []map[string]any) []WebSearchItem {
	if len(results) == 0 {
		return nil
	}
	items := make([]WebSearchItem, 0, len(results))
	for _, result := range results {
		items = append(items, WebSearchItem{
			Title:       pickString(result, "title", "name"),
			URL:         pickString(result, "url", "link", "source_url"),
			Snippet:     pickString(result, "snippet", "summary", "content", "abstract", "text"),
			Source:      pickString(result, "source", "site", "site_name"),
			PublishedAt: pickString(result, "date", "publish_time", "published_at", "time"),
			Raw:         result,
		})
	}
	return items
}

func parseKnowledgeChunks(results []map[string]any) []KnowledgeSearchChunk {
	if len(results) == 0 {
		return nil
	}
	chunks := make([]KnowledgeSearchChunk, 0, len(results))
	for _, result := range results {
		score, _ := pickFloat(result, "score", "similarity", "rerank_score")
		chunks = append(chunks, KnowledgeSearchChunk{
			DocID:  pickString(result, "doc_id", "document_id", "id"),
			Title:  pickString(result, "title", "doc_title", "name"),
			Text:   pickString(result, "text", "content", "chunk"),
			Score:  score,
			Source: pickString(result, "source", "url"),
			Raw:    result,
		})
	}
	return chunks
}

func parseBoundingBoxes(artifacts []map[string]any) []BoundingBox {
	if len(artifacts) == 0 {
		return nil
	}
	var boxes []BoundingBox
	for _, artifact := range artifacts {
		if box, ok := extractBoundingBox(artifact); ok {
			boxes = append(boxes, box)
		}
	}
	return boxes
}

func extractBoundingBox(payload map[string]any) (BoundingBox, bool) {
	if payload == nil {
		return BoundingBox{}, false
	}
	if box, ok := payload["box"].(map[string]any); ok {
		return parseBoxMap(box)
	}
	if box, ok := payload["bbox"].(map[string]any); ok {
		return parseBoxMap(box)
	}
	if box, ok := payload["rectangle"].(map[string]any); ok {
		return parseBoxMap(box)
	}
	x1, ok1 := pickFloat(payload, "x1", "left")
	y1, ok2 := pickFloat(payload, "y1", "top")
	x2, ok3 := pickFloat(payload, "x2", "right")
	y2, ok4 := pickFloat(payload, "y2", "bottom")
	if ok1 && ok2 && ok3 && ok4 {
		return BoundingBox{X1: x1, Y1: y1, X2: x2, Y2: y2}, true
	}
	x, okX := pickFloat(payload, "x")
	y, okY := pickFloat(payload, "y")
	w, okW := pickFloat(payload, "width", "w")
	h, okH := pickFloat(payload, "height", "h")
	if okX && okY && okW && okH {
		return BoundingBox{X1: x, Y1: y, X2: x + w, Y2: y + h, W: w, H: h}, true
	}
	return BoundingBox{}, false
}

func parseBoxMap(payload map[string]any) (BoundingBox, bool) {
	if payload == nil {
		return BoundingBox{}, false
	}
	x1, ok1 := pickFloat(payload, "x1", "left", "x_min")
	y1, ok2 := pickFloat(payload, "y1", "top", "y_min")
	x2, ok3 := pickFloat(payload, "x2", "right", "x_max")
	y2, ok4 := pickFloat(payload, "y2", "bottom", "y_max")
	if ok1 && ok2 && ok3 && ok4 {
		return BoundingBox{X1: x1, Y1: y1, X2: x2, Y2: y2}, true
	}
	x, okX := pickFloat(payload, "x")
	y, okY := pickFloat(payload, "y")
	w, okW := pickFloat(payload, "width", "w")
	h, okH := pickFloat(payload, "height", "h")
	if okX && okY && okW && okH {
		return BoundingBox{X1: x, Y1: y, X2: x + w, Y2: y + h, W: w, H: h}, true
	}
	return BoundingBox{}, false
}

func asMapSlice(value any) []map[string]any {
	switch v := value.(type) {
	case []map[string]any:
		return v
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if cast, ok := item.(map[string]any); ok {
				out = append(out, cast)
			}
		}
		return out
	case map[string]any:
		return []map[string]any{v}
	default:
		return nil
	}
}

func stringifyContent(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		var out strings.Builder
		for _, item := range v {
			if text := stringifyContent(item); text != "" {
				out.WriteString(text)
			}
		}
		return out.String()
	case map[string]any:
		if text, ok := v["text"].(string); ok {
			return text
		}
		if delta, ok := v["delta"].(string); ok {
			return delta
		}
		return ""
	default:
		return ""
	}
}

func pickString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}
	return ""
}

func pickFloat(payload map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case float64:
			return v, true
		case float32:
			return float64(v), true
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		case json.Number:
			if f, err := v.Float64(); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}
