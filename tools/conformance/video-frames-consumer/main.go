package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/videoframes"
)

type result struct {
	FrameCount   int      `json:"frame_count"`
	FilesTouched []string `json:"files_touched"`
	MediaSource  string   `json:"media_source"`
	Verified     bool     `json:"verified"`
}

func run(ctx context.Context) (result, error) {
	if runtime.GOOS == "windows" {
		return result{}, fmt.Errorf("video frames conformance shell stubs are unavailable on windows")
	}
	root, err := os.MkdirTemp("", "agentx-video-frames-consumer-")
	if err != nil {
		return result{}, err
	}
	defer os.RemoveAll(root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		return result{}, err
	}
	ffprobe := filepath.Join(binDir, "ffprobe")
	ffmpeg := filepath.Join(binDir, "ffmpeg")
	if err := os.WriteFile(ffprobe, []byte(`#!/bin/sh
set -eu
case "$*" in
  *"frame=best_effort_timestamp_time"*) printf '0.500000\n3.250000\n' ;;
  *) printf '%s\n' '{"streams":[{"width":1280,"height":720,"r_frame_rate":"30/1","duration":"10.0"}],"format":{"duration":"10.0"}}' ;;
esac
`), 0o700); err != nil {
		return result{}, err
	}
	if err := os.WriteFile(ffmpeg, []byte(`#!/bin/sh
set -eu
last=""
for arg in "$@"; do last="$arg"; done
out_dir=$(dirname "$last")
mkdir -p "$out_dir"
printf 'frame-a' > "$out_dir/frame_00001.jpg"
printf 'frame-b' > "$out_dir/frame_00002.jpg"
`), 0o700); err != nil {
		return result{}, err
	}
	if err := os.WriteFile(filepath.Join(root, "clip.mp4"), []byte("stub-video"), 0o600); err != nil {
		return result{}, err
	}

	registry := tools.NewRegistry()
	adapter := videoframes.NewLocalAdapter(videoframes.LocalOptions{Root: root, FFprobePath: ffprobe, FFmpegPath: ffmpeg})
	videoframes.Register(registry, adapter)
	raw, err := registry.Execute(ctx, toolcontract.Call{Name: "video_frames", Arguments: `{"path":"clip.mp4","interval_sec":5}`})
	if err != nil {
		return result{}, err
	}
	var payload videoframes.Result
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return result{}, err
	}
	value := result{FrameCount: payload.FrameCount, FilesTouched: payload.FilesTouched}
	if len(payload.Frames) > 0 && payload.Frames[0].Media != nil {
		value.MediaSource = payload.Frames[0].Media.Source
	}
	value.Verified = value.FrameCount == 2 && len(value.FilesTouched) == 2 && value.MediaSource == "video" && strings.HasPrefix(value.FilesTouched[0], ".agentx/artifacts/video_frames/")
	return value, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !value.Verified {
		fmt.Fprintln(os.Stderr, "video frames conformance failed")
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
