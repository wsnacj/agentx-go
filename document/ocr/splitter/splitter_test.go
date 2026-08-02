package splitter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

func TestPopplerSplitterUsesBatchPagesAndMaxParallel(t *testing.T) {
	t.Parallel()

	pdfPath := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	var mu sync.Mutex
	var ranges []string
	currentRuns := 0
	maxConcurrentRuns := 0

	split := popplerSplitter{
		cfg: config.SplitterConfig{
			DPI:         200,
			BatchPages:  2,
			MaxParallel: 2,
		},
		execCommand: func(ctx context.Context, workDir, name string, args []string) error {
			start := argInt(args, "-f")
			end := argInt(args, "-l")
			prefix := args[len(args)-1]

			mu.Lock()
			currentRuns++
			if currentRuns > maxConcurrentRuns {
				maxConcurrentRuns = currentRuns
			}
			ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			mu.Unlock()

			time.Sleep(20 * time.Millisecond)
			for page := start; page <= end; page++ {
				if err := os.WriteFile(fmt.Sprintf("%s-%d.png", prefix, page), []byte("png"), 0o644); err != nil {
					return err
				}
			}

			mu.Lock()
			currentRuns--
			mu.Unlock()
			return nil
		},
		detectPageCount: func(ctx context.Context, inputPath string) (int, error) {
			return 5, nil
		},
	}

	result, err := split.Split(context.Background(), Request{Path: pdfPath})
	if err != nil {
		t.Fatalf("split pdf: %v", err)
	}
	defer func() {
		if result.Cleanup != nil {
			_ = result.Cleanup()
		}
	}()

	var pages []int
	for _, image := range result.Images {
		page, ok := popplerPageNumber(image)
		if !ok {
			t.Fatalf("expected page number in %q", image)
		}
		pages = append(pages, page)
	}
	if want := []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(pages, want) {
		t.Fatalf("unexpected page order: got %v want %v", pages, want)
	}

	sort.Strings(ranges)
	if want := []string{"1-2", "3-4", "5-5"}; !reflect.DeepEqual(ranges, want) {
		t.Fatalf("unexpected ranges: got %v want %v", ranges, want)
	}
	if maxConcurrentRuns != 2 {
		t.Fatalf("expected max 2 concurrent runs, got %d", maxConcurrentRuns)
	}
	if got := result.Stats["batching_mode"]; got != "batched" {
		t.Fatalf("unexpected batching mode: %v", got)
	}
	if got := result.Stats["batches"]; got != 3 {
		t.Fatalf("unexpected batch count: %v", got)
	}
	if got := result.Stats["batch_pages"]; got != 2 {
		t.Fatalf("unexpected batch pages: %v", got)
	}
	if got := result.Stats["max_parallel"]; got != 2 {
		t.Fatalf("unexpected max_parallel: %v", got)
	}
}

func TestPopplerSplitterFallsBackToSingleRunWithoutPageCount(t *testing.T) {
	t.Parallel()

	pdfPath := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	execCalls := 0
	split := popplerSplitter{
		cfg: config.SplitterConfig{
			BatchPages:  2,
			MaxParallel: 4,
		},
		execCommand: func(ctx context.Context, workDir, name string, args []string) error {
			execCalls++
			if hasArg(args, "-l") {
				t.Fatalf("did not expect -l when page count lookup fails: %v", args)
			}
			prefix := args[len(args)-1]
			for page := 1; page <= 3; page++ {
				if err := os.WriteFile(fmt.Sprintf("%s-%d.png", prefix, page), []byte("png"), 0o644); err != nil {
					return err
				}
			}
			return nil
		},
		detectPageCount: func(ctx context.Context, inputPath string) (int, error) {
			return 0, errors.New("pdfinfo unavailable")
		},
	}

	result, err := split.Split(context.Background(), Request{Path: pdfPath})
	if err != nil {
		t.Fatalf("split pdf: %v", err)
	}
	defer func() {
		if result.Cleanup != nil {
			_ = result.Cleanup()
		}
	}()

	if execCalls != 1 {
		t.Fatalf("expected single pdftocairo run, got %d", execCalls)
	}
	if got := result.Stats["batching_mode"]; got != "single" {
		t.Fatalf("unexpected batching mode: %v", got)
	}
	if got := result.Stats["batching_fallback"]; got != "page_count_unavailable" {
		t.Fatalf("unexpected fallback reason: %v", got)
	}
}

func TestRemoteSplitterStreamsMultipartRequest(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(filePath, []byte("stream-me"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	var gotContentLength int64
	var gotMaxPages string
	var gotFile string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentLength = r.ContentLength

		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for {
			part, err := reader.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			data, err := io.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch part.FormName() {
			case "file":
				gotFile = string(data)
			case "max_pages":
				gotMaxPages = string(data)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"images":["/tmp/page-2.png","/tmp/page-1.png"]}`)
	}))
	defer server.Close()

	split := remoteSplitter{
		cfg:     config.SplitterConfig{},
		baseURL: server.URL,
		client: &http.Client{
			Timeout:   1500 * time.Millisecond,
			Transport: server.Client().Transport,
		},
	}

	result, err := split.Split(context.Background(), Request{Path: filePath, MaxPages: 2})
	if err != nil {
		t.Fatalf("remote split: %v", err)
	}

	if gotContentLength != -1 {
		t.Fatalf("expected streaming request with unknown content length, got %d", gotContentLength)
	}
	if gotFile != "stream-me" {
		t.Fatalf("unexpected uploaded file content: %q", gotFile)
	}
	if gotMaxPages != "2" {
		t.Fatalf("unexpected max_pages value: %q", gotMaxPages)
	}
	if want := []string{"/tmp/page-1.png", "/tmp/page-2.png"}; !reflect.DeepEqual(result.Images, want) {
		t.Fatalf("unexpected images: got %v want %v", result.Images, want)
	}
	if got := result.Stats["timeout"]; got != "1.5s" {
		t.Fatalf("unexpected timeout stat: %v", got)
	}
}

func TestNewRemoteSplitterUsesConfiguredTimeout(t *testing.T) {
	t.Parallel()

	created, err := newRemoteSplitter(config.SplitterConfig{
		Options: map[string]any{
			"base_url": "http://example.com",
			"timeout":  "2500ms",
		},
	})
	if err != nil {
		t.Fatalf("new remote splitter: %v", err)
	}

	remote, ok := created.(*remoteSplitter)
	if !ok {
		t.Fatalf("unexpected splitter type %T", created)
	}
	if remote.client == nil {
		t.Fatalf("expected client to be initialized")
	}
	if remote.client.Timeout != 2500*time.Millisecond {
		t.Fatalf("unexpected timeout: %v", remote.client.Timeout)
	}
}

func argInt(args []string, flag string) int {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != flag {
			continue
		}
		value, err := strconv.Atoi(args[i+1])
		if err != nil {
			return 0
		}
		return value
	}
	return 0
}

func hasArg(args []string, flag string) bool {
	for _, arg := range args {
		if strings.TrimSpace(arg) == flag {
			return true
		}
	}
	return false
}
