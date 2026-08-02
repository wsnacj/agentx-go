package tools

import (
	"os"
	"path/filepath"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

type browserArtifactPayload struct {
	Role  string                  `json:"role,omitempty"`
	Kind  string                  `json:"kind,omitempty"`
	Path  string                  `json:"path,omitempty"`
	Bytes int64                   `json:"bytes,omitempty"`
	Media *agentxmedia.Descriptor `json:"media,omitempty"`
}

const browserArtifactContract = "artifacts+media"

func browserArtifactKind(kind string, path string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "screenshot":
		return "screenshot"
	case "save_pdf", "print_pdf", "pdf":
		return "pdf"
	case "download":
		return "download"
	case "wait_download":
		return "download"
	case "trace_stop", "trace":
		return "trace"
	case "save_html", "save_page", "html":
		return "html"
	}
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(path))) {
	case ".pdf":
		return "pdf"
	case ".html", ".htm", ".mhtml":
		return "html"
	case ".png", ".jpg", ".jpeg", ".webp":
		return "screenshot"
	default:
		if strings.TrimSpace(path) != "" {
			return "file"
		}
		return ""
	}
}

func browserScreenshotMediaDescriptor(path string, bytes int64, finalURL string, result agentxbrowserruntime.BrowserScreenshotResult) *agentxmedia.Descriptor {
	path = strings.TrimSpace(path)
	format := mediaDescriptorFormatFromPath(path)
	return &agentxmedia.Descriptor{
		Source:        "browser",
		Kind:          "screenshot",
		Path:          path,
		URL:           strings.TrimSpace(finalURL),
		MIMEType:      mediaDescriptorMIMETypeFromPath(path),
		Format:        format,
		Bytes:         bytes,
		CaptureScope:  strings.TrimSpace(result.CaptureScope),
		CaptureWidth:  result.CaptureWidth,
		CaptureHeight: result.CaptureHeight,
	}
}

func browserArtifactsForMedia(role string, media *agentxmedia.Descriptor) []browserArtifactPayload {
	if media == nil {
		return nil
	}
	return []browserArtifactPayload{{
		Role:  strings.TrimSpace(role),
		Kind:  strings.TrimSpace(media.Kind),
		Path:  strings.TrimSpace(media.Path),
		Bytes: media.Bytes,
		Media: media,
	}}
}

func browserScreenshotArtifacts(path string, bytes int64, finalURL string, result agentxbrowserruntime.BrowserScreenshotResult) []browserArtifactPayload {
	return browserArtifactsForMedia("primary", browserScreenshotMediaDescriptor(path, bytes, finalURL, result))
}

func browserArtifactTouchedPaths(paths ...string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserFileMediaDescriptor(kind string, path string, bytes int64, finalURL string) *agentxmedia.Descriptor {
	path = strings.TrimSpace(path)
	artifactKind := browserArtifactKind(kind, path)
	if path == "" || artifactKind == "" {
		return nil
	}
	return &agentxmedia.Descriptor{
		Source:   "browser",
		Kind:     artifactKind,
		Path:     path,
		URL:      strings.TrimSpace(finalURL),
		MIMEType: mediaDescriptorMIMETypeFromPath(path),
		Format:   mediaDescriptorFormatFromPath(path),
		Bytes:    bytes,
	}
}

func browserActResultMediaDescriptor(result agentxbrowserruntime.BrowserActResult) *agentxmedia.Descriptor {
	if strings.TrimSpace(result.Path) == "" {
		return nil
	}
	if strings.ToLower(strings.TrimSpace(result.Kind)) != "screenshot" {
		return browserFileMediaDescriptor(result.Kind, result.Path, result.Bytes, result.FinalURL)
	}
	return &agentxmedia.Descriptor{
		Source:        "browser",
		Kind:          "screenshot",
		Path:          strings.TrimSpace(result.Path),
		URL:           strings.TrimSpace(result.FinalURL),
		MIMEType:      mediaDescriptorMIMETypeFromPath(result.Path),
		Format:        mediaDescriptorFormatFromPath(result.Path),
		Bytes:         result.Bytes,
		CaptureScope:  strings.TrimSpace(result.CaptureScope),
		CaptureWidth:  result.CaptureWidth,
		CaptureHeight: result.CaptureHeight,
	}
}

func browserActArtifacts(result agentxbrowserruntime.BrowserActResult) []browserArtifactPayload {
	return browserArtifactsForMedia("primary", browserActResultMediaDescriptor(result))
}

func pdfRenderedPageMediaDescriptor(path string, bytes int64, page int) agentxmedia.Descriptor {
	path = strings.TrimSpace(path)
	return agentxmedia.Descriptor{
		Source:   "pdf",
		Kind:     "rendered_page",
		Path:     path,
		MIMEType: mediaDescriptorMIMETypeFromPath(path),
		Format:   mediaDescriptorFormatFromPath(path),
		Bytes:    bytes,
		Index:    page,
	}
}

func mediaDescriptorFormatFromPath(path string) string {
	switch strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(path))), ".") {
	case "jpeg":
		return "jpg"
	case "jpg", "png", "gif", "webp", "bmp", "tiff", "mp4", "mov", "m4v", "webm":
		return strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(path))), ".")
	default:
		return ""
	}
}

func mediaDescriptorMIMETypeFromPath(path string) string {
	format := mediaDescriptorFormatFromPath(path)
	if format == "" {
		return ""
	}
	if format == "jpg" {
		return "image/jpeg"
	}
	if format == "png" || format == "gif" || format == "webp" || format == "bmp" || format == "tiff" {
		return "image/" + format
	}
	return "video/" + format
}

func toolMediaDescriptorFileBytes(path string) int64 {
	info, err := os.Stat(strings.TrimSpace(path))
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}
