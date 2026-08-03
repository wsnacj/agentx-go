//go:build !windows

package videoframes_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/videoframes"
)

func TestRegisterVideoFramesToolsRejectsFIFOWithoutOpeningIt(t *testing.T) {
	restoreVideoFramesBins(t)
	binDir := t.TempDir()
	writeVideoFramesStubFFProbe(t, binDir)
	writeVideoFramesStubFFmpeg(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "stream.mp4"), 0o600); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}
	reg := tools.NewRegistry()
	registerVideoFramesTools(reg, videoframes.LocalOptions{Root: root})
	_, err := reg.Execute(context.Background(), toolcontract.Call{Name: "video_frames", Arguments: `{"path":"stream.mp4"}`})
	if !errors.Is(err, videoframes.ErrUnsafeFile) {
		t.Fatalf("FIFO input error=%v want ErrUnsafeFile", err)
	}
}
