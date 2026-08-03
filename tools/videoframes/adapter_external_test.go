package videoframes_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/videoframes"
)

func TestVideoFramesOutputSchemaDeclaresFilesTouched(t *testing.T) {
	definition := videoframes.Definition()
	properties, ok := definition.Function.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected video_frames input properties, got %#v", definition.Function.Parameters)
	}
	for _, forbidden := range []string{"root", "workspace_root", "artifact_root", "out_dir", "output_dir"} {
		if _, exists := properties[forbidden]; exists {
			t.Fatalf("video_frames input schema exposes host-owned field %q", forbidden)
		}
	}
	schema := definition.Function.OutputSchema
	if len(schema) == 0 {
		t.Fatal("expected video_frames output schema")
	}
	if schema["additionalProperties"] != false {
		t.Fatalf("expected video_frames output schema to be closed, got %#v", schema["additionalProperties"])
	}
	assertRequiredFields(t, schema, []string{"tool", "action", "status"})
	assertSchemaProperties(t, schema, []string{
		"tool",
		"action",
		"status",
		"source_video",
		"strategy",
		"interval_sec",
		"output_dir",
		"frame_count",
		"files_touched",
		"probe",
		"frames",
		"warning",
	})
}

func TestRegisterVideoFramesTools_IntervalExtractsFramesManifest(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	videoPath := filepath.Join(root, "clip.mp4")
	if err := os.WriteFile(videoPath, []byte("stub-video"), 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}

	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	if len(reg.Definitions()) != 1 {
		t.Fatalf("expected one video_frames tool definition, got %#v", reg.Definitions())
	}

	out, err := reg.Execute(context.Background(), toolcontract.Call{
		Name: "video_frames",
		Arguments: mustJSON(t, map[string]any{
			"path":         "clip.mp4",
			"strategy":     "interval",
			"interval_sec": 5,
		}),
	})
	if err != nil {
		t.Fatalf("video_frames interval: %v", err)
	}
	var payload videoframes.Result
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Status != "success" || payload.Tool != "video_frames" || payload.Action != "extract" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Strategy != "interval" || payload.IntervalSec != 5 {
		t.Fatalf("unexpected strategy payload: %#v", payload)
	}
	if payload.Probe == nil || payload.Probe.Width != 1280 || payload.Probe.Height != 720 || payload.Probe.FPS != 30 || payload.Probe.DurationSec != 10 {
		t.Fatalf("unexpected probe payload: %#v", payload.Probe)
	}
	if payload.FrameCount != 2 || len(payload.Frames) != 2 {
		t.Fatalf("expected two frames, got %#v", payload.Frames)
	}
	if !reflect.DeepEqual(payload.FilesTouched, []string{
		payload.Frames[0].FramePath,
		payload.Frames[1].FramePath,
	}) {
		t.Fatalf("expected files_touched to mirror extracted frames, got %#v", payload.FilesTouched)
	}
	if payload.Frames[0].TimestampSec != 0 || payload.Frames[1].TimestampSec != 5 {
		t.Fatalf("unexpected frame timestamps: %#v", payload.Frames)
	}
	for idx, frame := range payload.Frames {
		if !strings.HasPrefix(frame.FramePath, ".agentx/artifacts/video_frames/") {
			t.Fatalf("expected frame path under .agentx/artifacts/video_frames, got %#v", frame.FramePath)
		}
		if frame.Media == nil || frame.Media.Source != "video" || frame.Media.Kind != "frame" || frame.Media.Index != idx+1 {
			t.Fatalf("unexpected frame media descriptor: %#v", frame.Media)
		}
		if frame.Media.TimestampMs != int64(frame.TimestampSec*1000) {
			t.Fatalf("expected timestamp_ms to match timestamp_sec, got %#v", frame.Media)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(frame.FramePath))); err != nil {
			t.Fatalf("expected extracted frame to exist: %v", err)
		}
		if runtime.GOOS != "windows" {
			info, err := os.Stat(filepath.Join(root, filepath.FromSlash(frame.FramePath)))
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("frame mode=%#o want 0600", info.Mode().Perm())
			}
		}
	}
	if snapshots, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(payload.OutputDir), "source*")); err != nil || len(snapshots) != 0 {
		t.Fatalf("input snapshot must not remain in artifact output: paths=%v err=%v", snapshots, err)
	}
}

func TestRegisterVideoFramesTools_KeyframeUsesFFProbeTimestamps(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	videoPath := filepath.Join(root, "clip.mp4")
	if err := os.WriteFile(videoPath, []byte("stub-video"), 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}

	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})

	out, err := reg.Execute(context.Background(), toolcontract.Call{
		Name: "video_frames",
		Arguments: mustJSON(t, map[string]any{
			"path":     "clip.mp4",
			"strategy": "keyframe",
		}),
	})
	if err != nil {
		t.Fatalf("video_frames keyframe: %v", err)
	}
	var payload videoframes.Result
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Strategy != "keyframe" || payload.IntervalSec != 0 {
		t.Fatalf("unexpected keyframe payload: %#v", payload)
	}
	if !reflect.DeepEqual(payload.FilesTouched, []string{
		payload.Frames[0].FramePath,
		payload.Frames[1].FramePath,
	}) {
		t.Fatalf("expected keyframe files_touched to mirror extracted frames, got %#v", payload.FilesTouched)
	}
	if len(payload.Frames) != 2 || payload.Frames[0].TimestampSec != 0.5 || payload.Frames[1].TimestampSec != 3.25 {
		t.Fatalf("expected keyframe timestamps from ffprobe, got %#v", payload.Frames)
	}
}

func TestRegisterVideoFramesTools_HidesWhenBinariesUnavailable(t *testing.T) {
	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{
		Root:        t.TempDir(),
		FFmpegPath:  filepath.Join(t.TempDir(), "missing-ffmpeg"),
		FFprobePath: filepath.Join(t.TempDir(), "missing-ffprobe"),
	})
	if len(reg.Definitions()) != 0 {
		t.Fatalf("expected video_frames hidden when ffmpeg/ffprobe unavailable, got %#v", reg.Definitions())
	}
}

func TestRegisterVideoFramesTools_MissingPathReturnsStructuredArgumentError(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: t.TempDir()})
	_, err := reg.Execute(context.Background(), toolcontract.Call{
		Name:      "video_frames",
		Arguments: `{}`,
	})
	if err == nil {
		t.Fatalf("expected missing path error")
	}
	argErr, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("unexpected argument error code: %#v", argErr)
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"path"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
}

func TestRegisterVideoFramesTools_RejectsCallerControlledOutputDirWithoutDeletingIt(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	videoPath := filepath.Join(root, "clip.mp4")
	if err := os.WriteFile(videoPath, []byte("stub-video"), 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}
	protectedDir := filepath.Join(root, "protected")
	if err := os.MkdirAll(protectedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(protectedDir, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	_, err := reg.Execute(context.Background(), toolcontract.Call{
		Name: "video_frames",
		Arguments: mustJSON(t, map[string]any{
			"path":    "clip.mp4",
			"out_dir": "protected",
		}),
	})
	argErr, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok || argErr.Code != agentxtoolerrors.ToolArgumentErrorCodeInvalidArgument || !reflect.DeepEqual(argErr.InvalidFields, []string{"out_dir"}) {
		t.Fatalf("expected host-owned out_dir error, got %T %#v", err, argErr)
	}
	if raw, readErr := os.ReadFile(sentinel); readErr != nil || string(raw) != "keep" {
		t.Fatalf("caller directory was modified: content=%q err=%v", raw, readErr)
	}
}

func TestRegisterVideoFramesTools_RejectsModelRootOverride(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: t.TempDir()})
	_, err := reg.Execute(context.Background(), toolcontract.Call{
		Name:      "video_frames",
		Arguments: mustJSON(t, map[string]any{"path": "missing.mp4", "root": t.TempDir()}),
	})
	argErr, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok || argErr.Code != agentxtoolerrors.ToolArgumentErrorCodeInvalidArgument || !reflect.DeepEqual(argErr.InvalidFields, []string{"root"}) {
		t.Fatalf("expected host-owned root error, got %T %#v", err, argErr)
	}
}

func TestRegisterVideoFramesTools_RepeatedRunsPreservePriorArtifacts(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "clip.mp4"), []byte("stub-video"), 0o600); err != nil {
		t.Fatal(err)
	}
	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	run := func() videoframes.Result {
		t.Helper()
		out, err := reg.Execute(context.Background(), toolcontract.Call{Name: "video_frames", Arguments: `{"path":"clip.mp4"}`})
		if err != nil {
			t.Fatal(err)
		}
		var payload videoframes.Result
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}
	first := run()
	second := run()
	if first.OutputDir == second.OutputDir {
		t.Fatalf("runtime-owned output directories must be unique: %q", first.OutputDir)
	}
	if len(first.Frames) == 0 {
		t.Fatal("first run returned no frames")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(first.Frames[0].FramePath))); err != nil {
		t.Fatalf("second run removed first artifacts: %v", err)
	}
}

func TestRegisterVideoFramesTools_EnforcesInputByteBudget(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	content := []byte("stub-video")
	if err := os.WriteFile(filepath.Join(root, "clip.mp4"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Run("exact limit", func(t *testing.T) {
		reg := tools.NewRegistry()
		registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root, MaxInputBytes: int64(len(content))})
		if _, err := reg.Execute(context.Background(), toolcontract.Call{Name: "video_frames", Arguments: `{"path":"clip.mp4"}`}); err != nil {
			t.Fatalf("exact-limit input: %v", err)
		}
	})

	t.Run("overflow", func(t *testing.T) {
		reg := tools.NewRegistry()
		registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root, MaxInputBytes: int64(len(content) - 1)})
		_, err := reg.Execute(context.Background(), toolcontract.Call{Name: "video_frames", Arguments: `{"path":"clip.mp4"}`})
		if !errors.Is(err, videoframes.ErrFileTooLarge) {
			t.Fatalf("input budget error=%v want ErrFileTooLarge", err)
		}
	})
}

func TestRegisterVideoFramesTools_RejectsSymlinkInput(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	target := filepath.Join(root, "target.mp4")
	if err := os.WriteFile(target, []byte("stub-video"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "link.mp4")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	_, err := reg.Execute(context.Background(), toolcontract.Call{Name: "video_frames", Arguments: `{"path":"link.mp4"}`})
	if !errors.Is(err, videoframes.ErrUnsafeFile) {
		t.Fatalf("symlink input error=%v want ErrUnsafeFile", err)
	}
}

func TestRegisterVideoFramesTools_HonorsCanceledContextBeforeArtifacts(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	_, err := reg.Execute(ctx, toolcontract.Call{Name: "video_frames", Arguments: `{"path":"clip.mp4"}`})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled video_frames error=%v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".agentx")); !os.IsNotExist(statErr) {
		t.Fatalf("canceled call created artifacts: %v", statErr)
	}
}

func restoreVideoFramesBins(t *testing.T) { t.Helper() }

func registerVideoFramesTools(reg *tools.Registry, opts videoframes.LocalOptions) {
	videoframes.Register(reg, videoframes.NewLocalAdapter(opts))
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	blob, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(blob)
}

func assertRequiredFields(t *testing.T, schema map[string]any, fields []string) {
	t.Helper()
	raw, ok := schema["required"].([]string)
	if !ok {
		t.Fatalf("required=%#v", schema["required"])
	}
	for _, field := range fields {
		found := false
		for _, value := range raw {
			if value == field {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("required missing %q: %#v", field, raw)
		}
	}
}

func assertSchemaProperties(t *testing.T, schema map[string]any, fields []string) {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties=%#v", schema["properties"])
	}
	for _, field := range fields {
		if _, ok := properties[field]; !ok {
			t.Fatalf("properties missing %q", field)
		}
	}
}

func writeVideoFramesStubFFProbe(t *testing.T, dir string) {
	t.Helper()
	content := `#!/bin/sh
set -eu
case "$*" in
  *"frame=best_effort_timestamp_time"*)
    printf '0.500000\n3.250000\n'
    ;;
  *)
    cat <<'EOF'
{"streams":[{"width":1280,"height":720,"r_frame_rate":"30/1","duration":"10.0"}],"format":{"duration":"10.0"}}
EOF
    ;;
esac
`
	writeExecutableScript(t, filepath.Join(dir, "ffprobe"), content)
}

func writeVideoFramesStubFFmpeg(t *testing.T, dir string) {
	t.Helper()
	content := `#!/bin/sh
set -eu
last=""
for arg in "$@"; do
  last="$arg"
done
out_dir=$(dirname "$last")
mkdir -p "$out_dir"
printf 'frame-a' > "$out_dir/frame_00001.jpg"
printf 'frame-b' > "$out_dir/frame_00002.jpg"
`
	writeExecutableScript(t, filepath.Join(dir, "ffmpeg"), content)
}

func writeExecutableScript(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write script %s: %v", path, err)
	}
}
