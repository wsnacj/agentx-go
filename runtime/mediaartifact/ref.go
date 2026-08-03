package mediaartifact

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Ref 是从 browser、PDF、video、nodes 等松散工具结果中提取出的媒体引用。
// 它只描述输入引用及其来源提示，不承诺引用可访问，也不执行持久化。
type Ref struct {
	Raw            string   `json:"raw"`
	Display        string   `json:"display,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	ModeHint       string   `json:"mode_hint,omitempty"`
	ArtifactSource string   `json:"artifact_source,omitempty"`
	ArtifactKind   string   `json:"artifact_kind,omitempty"`
}

// RefsFromValue 将工具输出中常见的字符串、数组和对象形态归一化为媒体引用。
// 该函数不访问文件或网络，也不会修改传入的 map、slice 或其子值。
func RefsFromValue(value any) ([]Ref, error) {
	refs, err := refsFromValue(value)
	if err != nil {
		return nil, err
	}
	out := make([]Ref, 0, len(refs))
	for _, ref := range refs {
		out = append(out, Ref{
			Raw:            strings.TrimSpace(ref.Raw),
			Display:        strings.TrimSpace(ref.Display),
			Labels:         append([]string(nil), ref.Labels...),
			ModeHint:       strings.TrimSpace(ref.ModeHint),
			ArtifactSource: strings.TrimSpace(ref.ArtifactSource),
			ArtifactKind:   strings.TrimSpace(ref.ArtifactKind),
		})
	}
	return out, nil
}

func refsFromValue(value any) ([]Ref, error) {
	if value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case string:
		return refsFromString(typed)
	case []string:
		out := make([]Ref, 0, len(typed))
		for _, item := range typed {
			refs, err := refsFromString(item)
			if err != nil {
				return nil, err
			}
			out = append(out, refs...)
		}
		return out, nil
	case []any:
		out := make([]Ref, 0, len(typed))
		for _, item := range typed {
			refs, err := refsFromValue(item)
			if err != nil {
				return nil, err
			}
			out = append(out, refs...)
		}
		return out, nil
	case map[string]any:
		return refsFromObject(typed, Ref{})
	default:
		// 保留既有 HS consumer 的错误文本，避免迁移改变可观察行为。
		return nil, fmt.Errorf("image_analyze: unsupported artifact type %T", value)
	}
}

func refsFromString(raw string) ([]Ref, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var decoded any
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			return refsFromValue(decoded)
		}
	}
	return []Ref{{Raw: trimmed}}, nil
}

func refsFromObject(obj map[string]any, inherited Ref) ([]Ref, error) {
	if len(obj) == 0 {
		return nil, nil
	}
	base := mergeRef(inherited, contextFromObject(obj))
	out := make([]Ref, 0, 4)
	for _, key := range []string{
		"media",
		"artifacts",
		"visual_analysis",
		"rendered_pages",
		"renderedPages",
		"files",
		"frames",
		"documents",
	} {
		children, err := refsFromChildValue(obj, key, base)
		if err != nil {
			return nil, err
		}
		out = append(out, children...)
	}
	if refs, ok := primaryRefsFromObject(obj, base); ok {
		out = append(out, refs...)
	}
	return out, nil
}

func refsFromChildValue(obj map[string]any, key string, base Ref) ([]Ref, error) {
	value, ok := obj[key]
	if !ok || value == nil {
		return nil, nil
	}
	children, err := refsFromValue(value)
	if err != nil {
		return nil, err
	}
	out := make([]Ref, 0, len(children))
	for _, child := range children {
		out = append(out, mergeRef(base, child))
	}
	return out, nil
}

func primaryRefsFromObject(obj map[string]any, base Ref) ([]Ref, bool) {
	raw := strings.TrimSpace(firstString(
		obj,
		"path",
		"frame_path",
		"image",
		"file",
		"file_path",
		"url",
		"data_url",
		"dataURI",
	))
	if raw == "" {
		return nil, false
	}
	ref := base
	ref.Raw = raw
	if strings.TrimSpace(ref.Display) == "" {
		ref.Display = firstNonEmpty(strings.TrimSpace(firstString(obj, "display")), raw)
	}
	return []Ref{ref}, true
}

func contextFromObject(obj map[string]any) Ref {
	action := strings.ToLower(strings.TrimSpace(firstString(obj, "action")))
	kind := strings.ToLower(strings.TrimSpace(firstString(obj, "kind")))
	toolName := normalizeToolName(firstString(obj, "tool", "source_tool", "sourceTool"))
	ref := Ref{
		Display:        strings.TrimSpace(firstString(obj, "display")),
		ArtifactSource: artifactSource(toolName, action, kind, obj),
		ArtifactKind:   artifactKind(action, kind, obj),
	}
	if rawMode := strings.TrimSpace(firstString(obj, "mode", "analysis_mode", "analysisMode")); rawMode != "" {
		ref.ModeHint = normalizeMode(rawMode)
	}
	if ref.ModeHint == "" {
		ref.ModeHint = modeHint(ref.ArtifactSource, ref.ArtifactKind, obj)
	}
	ref.Labels = labels(toolName, action, kind, obj)
	return ref
}

func artifactSource(toolName string, action string, kind string, obj map[string]any) string {
	if source := strings.ToLower(strings.TrimSpace(firstString(obj, "source"))); source == "browser" || source == "nodes" || source == "pdf" || source == "video" {
		return source
	}
	if source := browserArtifactSourceForTool(toolName); source != "" {
		return source
	}
	switch toolName {
	case "nodes":
		return "nodes"
	case "pdf", "pdf_analyze", "pdf_extract_structured":
		return "pdf"
	case "video_frames", "video-frames":
		return "video"
	}
	if kind == "screenshot" || strings.TrimSpace(firstString(obj, "capture_scope")) != "" {
		return "browser"
	}
	if strings.HasPrefix(action, "camera_") || strings.HasPrefix(action, "photos_") || strings.HasPrefix(action, "screen_") {
		return "nodes"
	}
	if strings.Contains(action, "frame") || strings.TrimSpace(firstString(obj, "frame_path")) != "" || strings.TrimSpace(firstString(obj, "strategy")) != "" {
		return "video"
	}
	return ""
}

func browserArtifactSourceForTool(toolName string) string {
	switch normalizeToolName(toolName) {
	case "browser_act", "browser_runtime", "browser_extract", "browser_screenshot":
		return "browser"
	default:
		return ""
	}
}

func artifactKind(action string, kind string, obj map[string]any) string {
	switch {
	case kind == "screenshot" || strings.TrimSpace(firstString(obj, "capture_scope")) != "":
		return "screenshot"
	case kind == "frame" || strings.TrimSpace(firstString(obj, "frame_path")) != "":
		return "frame"
	case kind != "":
		return kind
	case action != "":
		return action
	default:
		return ""
	}
}

func modeHint(artifactSource string, artifactKind string, obj map[string]any) string {
	if rawMode := strings.TrimSpace(firstString(obj, "mode", "analysis_mode", "analysisMode")); rawMode != "" {
		return normalizeMode(rawMode)
	}
	if artifactKind == "screenshot" {
		return "screenshot"
	}
	if artifactKind == "rendered_page" || artifactSource == "pdf" {
		return "document"
	}
	return ""
}

func labels(toolName string, action string, kind string, obj map[string]any) []string {
	out := make([]string, 0, 8)
	for _, item := range []string{
		toolName,
		action,
		kind,
		strings.TrimSpace(firstString(obj, "strategy")),
		strings.TrimSpace(firstString(obj, "note")),
		strings.TrimSpace(firstString(obj, "timestamp_sec", "timestampSec", "timestamp_ms", "timestampMs")),
		strings.TrimSpace(firstString(obj, "title")),
		strings.TrimSpace(firstString(obj, "final_url", "url")),
		strings.TrimSpace(firstString(obj, "target")),
		strings.TrimSpace(firstString(obj, "capture_scope")),
		strings.TrimSpace(firstString(obj, "facing")),
		strings.TrimSpace(firstString(obj, "created_at", "createdAt")),
	} {
		if strings.TrimSpace(item) != "" {
			out = append(out, item)
		}
	}
	return uniqueLabels(out)
}

func mergeRef(base Ref, incoming Ref) Ref {
	if strings.TrimSpace(incoming.Raw) != "" {
		base.Raw = strings.TrimSpace(incoming.Raw)
	}
	if strings.TrimSpace(base.Display) == "" {
		base.Display = incoming.Display
	}
	if strings.TrimSpace(base.ModeHint) == "" {
		base.ModeHint = incoming.ModeHint
	}
	if strings.TrimSpace(base.ArtifactSource) == "" {
		base.ArtifactSource = incoming.ArtifactSource
	}
	if strings.TrimSpace(base.ArtifactKind) == "" {
		base.ArtifactKind = incoming.ArtifactKind
	}
	base.Labels = uniqueLabels(append(base.Labels, incoming.Labels...))
	return base
}

func normalizeToolName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch normalized {
	case "bash":
		return "exec"
	case "apply-patch":
		return "apply_patch"
	default:
		return normalized
	}
}

func normalizeMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "ocr":
		return "ocr"
	case "document", "doc":
		return "document"
	case "chart", "graph", "dashboard":
		return "chart"
	case "screenshot", "screen", "ui":
		return "screenshot"
	case "", "general", "auto":
		return "general"
	default:
		return ""
	}
}

func uniqueLabels(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, item := range in {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	return out
}

func firstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		if params == nil {
			return ""
		}
		value, ok := params[key].(string)
		if ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonEmpty(parts ...string) string {
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
