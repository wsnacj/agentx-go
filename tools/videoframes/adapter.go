// Package videoframes provides an explicit, opt-in local video-frame adapter.
//
// The package owns portable argument, probing, extraction, artifact and result
// semantics. Hosts still own authorization, approval, sandboxing, dependency
// installation and whether the tool is registered at all.
package videoframes

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

const (
	defaultIntervalSec = 5.0
	defaultMaxFrames   = 24
	hardMaxFrames      = 120
	stdoutLimitBytes   = 8 << 20
	stderrLimitBytes   = 1 << 20
)

// LocalOptions configures a host-selected local ffmpeg/ffprobe adapter.
// Root is a containment boundary, not an authorization or sandbox policy.
type LocalOptions struct {
	Root          string
	MaxFrames     int
	MaxInputBytes int64
	FFmpegPath    string
	FFprobePath   string
}

// LocalAdapter coordinates one opt-in video_frames implementation.
// It is safe for concurrent calls after construction.
type LocalAdapter struct {
	root          string
	maxFrames     int
	maxInputBytes int64
	ffmpegPath    string
	ffprobePath   string
}

// NewLocalAdapter constructs an adapter without registering tools or running
// commands. Missing binary paths use ffmpeg and ffprobe from the Host PATH.
func NewLocalAdapter(opts LocalOptions) *LocalAdapter {
	ffmpegPath := strings.TrimSpace(opts.FFmpegPath)
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	ffprobePath := strings.TrimSpace(opts.FFprobePath)
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	maxInputBytes := opts.MaxInputBytes
	if maxInputBytes <= 0 {
		maxInputBytes = defaultMaxInputBytes
	}
	return &LocalAdapter{
		root:          opts.Root,
		maxFrames:     clampLimit(opts.MaxFrames, defaultMaxFrames, hardMaxFrames),
		maxInputBytes: maxInputBytes,
		ffmpegPath:    ffmpegPath,
		ffprobePath:   ffprobePath,
	}
}

// Available reports whether both configured command binaries can be resolved.
func (a *LocalAdapter) Available() bool {
	if a == nil {
		return false
	}
	for _, binary := range []string{a.ffprobePath, a.ffmpegPath} {
		if _, err := exec.LookPath(binary); err != nil {
			return false
		}
	}
	return true
}

// Register adds video_frames only when the explicit adapter and both binaries
// are available. It does not install dependencies or select a Host policy.
func Register(reg toolcontract.Registrar, adapter *LocalAdapter) {
	if reg == nil || adapter == nil || !adapter.Available() {
		return
	}
	reg.Register(Definition(), adapter.handle)
}

// Definition returns the stable model-facing video_frames declaration.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{
		Type: "function",
		Function: toolcontract.Function{
			Name:        "video_frames",
			Description: "Probe a local video and extract keyframes or interval frames into stable workspace artifacts with media descriptors.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":         map[string]any{"type": "string", "description": "Workspace-relative local video path."},
					"video":        map[string]any{"type": "string", "description": "Alias for path."},
					"strategy":     stringEnumSchema("Frame extraction strategy. Use interval for evenly sampled frames or keyframe for encoded keyframes.", "interval", "keyframe"),
					"mode":         stringEnumSchema("Compatibility alias for strategy.", "interval", "keyframe"),
					"interval_sec": map[string]any{"type": "number", "description": "Interval in seconds when strategy=interval."},
					"max_frames":   map[string]any{"type": "integer", "description": "Maximum number of frames to keep."},
				},
				"required": []string{"path"},
			},
			OutputSchema: outputSchema(),
		},
	}
}

// Result is the stable JSON result returned by video_frames.
type Result struct {
	Tool         string   `json:"tool"`
	Action       string   `json:"action"`
	Status       string   `json:"status"`
	SourceVideo  string   `json:"source_video,omitempty"`
	Strategy     string   `json:"strategy,omitempty"`
	IntervalSec  float64  `json:"interval_sec,omitempty"`
	OutputDir    string   `json:"output_dir,omitempty"`
	FrameCount   int      `json:"frame_count,omitempty"`
	FilesTouched []string `json:"files_touched,omitempty"`
	Probe        *Probe   `json:"probe,omitempty"`
	Frames       []Frame  `json:"frames,omitempty"`
	Warning      string   `json:"warning,omitempty"`
}

// Probe contains provider-neutral metadata reported by ffprobe.
type Probe struct {
	DurationSec float64 `json:"duration_sec,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	FPS         float64 `json:"fps,omitempty"`
}

// Frame describes one retained frame artifact.
type Frame struct {
	FramePath    string                  `json:"frame_path"`
	Index        int                     `json:"index"`
	TimestampSec float64                 `json:"timestamp_sec,omitempty"`
	Media        *agentxmedia.Descriptor `json:"media,omitempty"`
}

// FilesTouched returns stable, de-duplicated non-empty frame paths.
func FilesTouched(items []Frame) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		path := strings.TrimSpace(item.FramePath)
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

func (a *LocalAdapter) handle(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
	params, err := decodeArgs(call.Arguments)
	if err != nil {
		return "", err
	}
	if err := rejectHostOwnedArguments(params, "root", "out_dir", "output_dir"); err != nil {
		return "", err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	pathValue := firstString(params, "path", "video", "file", "file_path")
	if strings.TrimSpace(pathValue) == "" {
		return "", agentxtoolerrors.NewMissingRequiredToolArgumentError("video_frames", []string{"path"}, "video_frames: path/video is required")
	}
	resolvedPath, displayPath, err := resolvePathWithinRoot(a.root, pathValue)
	if err != nil {
		return "", fmt.Errorf("video_frames: %w", err)
	}
	strategy := normalizeStrategy(firstString(params, "strategy", "mode"))
	if strategy == "" {
		strategy = "interval"
	}
	intervalSec := readFloatDefault(params, defaultIntervalSec, "interval_sec", "interval")
	if strategy == "interval" && intervalSec <= 0 {
		return "", fmt.Errorf("video_frames: interval_sec must be > 0")
	}
	requestMaxFrames := clampLimit(firstInt(params, "max_frames"), a.maxFrames, hardMaxFrames)
	ownedOutput, err := createOwnedOutput(a.root, resolvedPath, strategy, intervalSec)
	if err != nil {
		return "", fmt.Errorf("video_frames: create owned artifact directory: %w", err)
	}
	defer ownedOutput.cleanup()
	snapshotPath, err := snapshotInput(ctx, resolvedPath, ownedOutput.path, a.maxInputBytes)
	if err != nil {
		return "", fmt.Errorf("video_frames: snapshot input: %w", err)
	}
	probe, err := a.probe(ctx, snapshotPath)
	if err != nil {
		return "", fmt.Errorf("video_frames: %w", err)
	}
	var keyframeTimestamps []float64
	if strategy == "keyframe" {
		keyframeTimestamps, err = a.probeKeyframeTimestamps(ctx, snapshotPath)
		if err != nil {
			return "", fmt.Errorf("video_frames: %w", err)
		}
	}
	if err := a.extract(ctx, snapshotPath, ownedOutput.path, strategy, intervalSec, requestMaxFrames); err != nil {
		return "", fmt.Errorf("video_frames: %w", err)
	}
	if err := removeInputSnapshot(snapshotPath); err != nil {
		return "", fmt.Errorf("video_frames: remove input snapshot: %w", err)
	}
	frames, warning, err := collectArtifacts(a.root, ownedOutput.path, strategy, intervalSec, requestMaxFrames, keyframeTimestamps)
	if err != nil {
		return "", fmt.Errorf("video_frames: %w", err)
	}
	payload := Result{
		Tool: "video_frames", Action: "extract", Status: "success", SourceVideo: displayPath,
		Strategy: strategy, IntervalSec: intervalSec, OutputDir: ownedOutput.display,
		FrameCount: len(frames), FilesTouched: FilesTouched(frames), Probe: probe, Frames: frames, Warning: warning,
	}
	if strategy != "interval" {
		payload.IntervalSec = 0
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	ownedOutput.retain()
	return string(blob), nil
}

type ffprobeJSON struct {
	Streams []struct {
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		FrameRate string `json:"r_frame_rate"`
		Duration  string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func (a *LocalAdapter) probe(ctx context.Context, path string) (*Probe, error) {
	stdout, stderr, err := a.runCommand(ctx, a.ffprobePath, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,r_frame_rate,duration:format=duration", "-of", "json", path)
	if err != nil {
		return nil, fmt.Errorf("ffprobe video metadata failed: %s", firstNonEmpty(strings.TrimSpace(stderr), err.Error()))
	}
	var payload ffprobeJSON
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		return nil, fmt.Errorf("decode ffprobe metadata: %w", err)
	}
	out := &Probe{}
	if len(payload.Streams) > 0 {
		out.Width = payload.Streams[0].Width
		out.Height = payload.Streams[0].Height
		out.FPS = parseRate(payload.Streams[0].FrameRate)
		if out.DurationSec == 0 {
			out.DurationSec = parseFloat(payload.Streams[0].Duration)
		}
	}
	if out.DurationSec == 0 {
		out.DurationSec = parseFloat(payload.Format.Duration)
	}
	return out, nil
}

func (a *LocalAdapter) probeKeyframeTimestamps(ctx context.Context, path string) ([]float64, error) {
	stdout, stderr, err := a.runCommand(ctx, a.ffprobePath, "-v", "error", "-select_streams", "v:0", "-skip_frame", "nokey", "-show_entries", "frame=best_effort_timestamp_time", "-of", "csv=p=0", path)
	if err != nil {
		return nil, fmt.Errorf("ffprobe keyframe timestamps failed: %s", firstNonEmpty(strings.TrimSpace(stderr), err.Error()))
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	out := make([]float64, 0, len(lines))
	for _, line := range lines {
		value := parseFloat(line)
		if value >= 0 {
			out = append(out, value)
		}
	}
	return out, nil
}

func (a *LocalAdapter) extract(ctx context.Context, inputPath, outputDir, strategy string, intervalSec float64, maxFrames int) error {
	outputPattern := filepath.Join(outputDir, "frame_%05d.jpg")
	args := []string{"-hide_banner", "-loglevel", "error", "-i", inputPath}
	switch strategy {
	case "keyframe":
		args = append(args, "-vf", "select='eq(pict_type\\,I)'", "-vsync", "vfr")
	case "interval":
		args = append(args, "-vf", fmt.Sprintf("fps=1/%s", strconv.FormatFloat(intervalSec, 'f', -1, 64)))
	default:
		return fmt.Errorf("unsupported strategy %q", strategy)
	}
	if maxFrames > 0 {
		args = append(args, "-frames:v", strconv.Itoa(maxFrames))
	}
	args = append(args, outputPattern)
	_, stderr, err := a.runCommand(ctx, a.ffmpegPath, args...)
	if err != nil {
		return fmt.Errorf("ffmpeg extract failed: %s", firstNonEmpty(strings.TrimSpace(stderr), err.Error()))
	}
	return nil
}

func collectArtifacts(root, outputDir, strategy string, intervalSec float64, maxFrames int, keyframeTimestamps []float64) ([]Frame, string, error) {
	files, err := filepath.Glob(filepath.Join(outputDir, "frame_*.jpg"))
	if err != nil {
		return nil, "", fmt.Errorf("list extracted frames: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, "", fmt.Errorf("no frames extracted")
	}
	if maxFrames > 0 && len(files) > maxFrames {
		files = files[:maxFrames]
	}
	out := make([]Frame, 0, len(files))
	warning := ""
	rootDir := strings.TrimSpace(root)
	if resolvedRoot, err := resolveRootDir(root); err == nil {
		rootDir = resolvedRoot
	}
	for idx, file := range files {
		info, err := os.Lstat(file)
		if err != nil {
			return nil, "", fmt.Errorf("inspect extracted frame: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, "", fmt.Errorf("extracted frame is not a regular file: %s", filepath.Base(file))
		}
		if err := os.Chmod(file, 0o600); err != nil {
			return nil, "", fmt.Errorf("protect extracted frame: %w", err)
		}
		display := filepath.ToSlash(file)
		if rootDir != "" {
			if rel, err := filepath.Rel(rootDir, file); err == nil {
				display = filepath.ToSlash(rel)
			}
		}
		item := Frame{FramePath: display, Index: idx + 1}
		switch strategy {
		case "interval":
			item.TimestampSec = float64(idx) * intervalSec
		case "keyframe":
			if idx < len(keyframeTimestamps) {
				item.TimestampSec = keyframeTimestamps[idx]
			} else if warning == "" {
				warning = "some keyframe timestamps were unavailable from ffprobe; extracted frames still returned in order"
			}
		}
		bytes := fileBytes(file)
		descriptor := agentxmedia.Descriptor{Source: "video", Kind: "frame", Path: display, MIMEType: mimeTypeFromPath(file), Format: formatFromPath(file), Bytes: bytes, Index: idx + 1, TimestampMs: int64(item.TimestampSec * 1000)}
		item.Media = &descriptor
		out = append(out, item)
	}
	return out, warning, nil
}

func (a *LocalAdapter) runCommand(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	captured, err := runBoundedCommand(cmd, commandLimits{stdoutBytes: stdoutLimitBytes, stderrBytes: stderrLimitBytes})
	if err != nil {
		return "", captured.summary(), err
	}
	return string(captured.stdout), string(captured.stderr), nil
}

func outputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"tool": stringSchema("Tool name that produced this response."), "action": stringSchema("Video action performed by the tool."), "status": stringSchema("Execution status, usually success."),
		"source_video": stringSchema("Workspace display path for the source video."), "strategy": stringEnumSchema("Frame extraction strategy used by the tool.", "interval", "keyframe"),
		"interval_sec": numberSchema("Interval in seconds used for interval extraction."), "output_dir": stringSchema("Workspace artifact directory used for extracted frames."),
		"frame_count": intSchema("Number of frame artifacts returned.", 0), "files_touched": stringArraySchema("Workspace frame artifacts actually written by video_frames."),
		"probe":   map[string]any{"type": "object", "description": "FFprobe metadata for the source video.", "additionalProperties": false, "properties": map[string]any{"duration_sec": numberSchema("Video duration in seconds reported by ffprobe."), "width": intSchema("Video width in pixels reported by ffprobe.", 0), "height": intSchema("Video height in pixels reported by ffprobe.", 0), "fps": numberSchema("Video frame rate reported by ffprobe.")}},
		"frames":  objectArraySchema("Extracted frame artifacts.", map[string]any{"frame_path": stringSchema("Workspace display path for the frame artifact."), "index": intSchema("One-based frame index in the returned artifact list.", 1), "timestamp_sec": numberSchema("Frame timestamp in seconds."), "media": looseObjectSchema("Media descriptor for downstream multimodal routing.")}),
		"warning": stringSchema("Non-fatal extraction warning, such as truncation."),
	}, []string{"tool", "action", "status"})
}

func rejectHostOwnedArguments(params map[string]any, fields ...string) error {
	present := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := params[field]; ok {
			present = append(present, field)
		}
	}
	if len(present) == 0 {
		return nil
	}
	return agentxtoolerrors.NewInvalidToolArgumentError("video_frames", present, "video_frames: workspace, memory, and artifact roots are host-owned and cannot be overridden by tool arguments")
}

func normalizeStrategy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "interval":
		return "interval"
	case "keyframe", "keyframes":
		return "keyframe"
	default:
		return ""
	}
}

func artifactStem(path, strategy string, intervalSec float64) string {
	stem := sanitizeArtifactName(strings.TrimSuffix(filepath.Base(strings.TrimSpace(path)), filepath.Ext(strings.TrimSpace(path))))
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(path))))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strings.TrimSpace(strategy)))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strconv.FormatFloat(intervalSec, 'f', -1, 64)))
	return fmt.Sprintf("%s-%x", stem, h.Sum64())
}

func parseRate(raw string) float64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	if strings.Contains(trimmed, "/") {
		parts := strings.SplitN(trimmed, "/", 2)
		num, den := parseFloat(parts[0]), parseFloat(parts[1])
		if den <= 0 {
			return 0
		}
		return num / den
	}
	return parseFloat(trimmed)
}

func parseFloat(raw string) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return value
}

func readFloatDefault(params map[string]any, fallback float64, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := readFloat(params, key); ok {
			return value
		}
	}
	return fallback
}

func readFloat(params map[string]any, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	raw, ok := params[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, false
		}
		return value, true
	case float32:
		number := float64(value)
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, false
		}
		return number, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	default:
		return 0, false
	}
}

func clampLimit(value, defaultValue, hardMax int) int {
	if value <= 0 {
		value = defaultValue
	}
	if hardMax > 0 && value > hardMax {
		value = hardMax
	}
	return value
}

func firstNonEmpty(parts ...string) string {
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			return part
		}
	}
	return ""
}

func formatFromPath(path string) string {
	switch strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(path))), ".") {
	case "jpeg":
		return "jpg"
	case "jpg", "png", "gif", "webp", "bmp", "tiff", "mp4", "mov", "m4v", "webm":
		return strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(path))), ".")
	default:
		return ""
	}
}

func mimeTypeFromPath(path string) string {
	format := formatFromPath(path)
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

func fileBytes(path string) int64 {
	info, err := os.Stat(strings.TrimSpace(path))
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}

func sanitizeArtifactName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "pdf"
	}
	var out strings.Builder
	lastDash := false
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				out.WriteByte('-')
				lastDash = true
			}
		}
	}
	cleaned := strings.Trim(strings.ToLower(out.String()), "-")
	if cleaned == "" {
		return "pdf"
	}
	return cleaned
}
