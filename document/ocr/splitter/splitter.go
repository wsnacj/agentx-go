package splitter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

// Request encapsulates a single file that may need splitting.
type Request struct {
	Path     string
	MaxPages int
	Options  map[string]any
}

// Result represents the outcome of a splitter operation.
type Result struct {
	Images  []string
	Stats   map[string]any
	Cleanup func() error
}

// Splitter is responsible for producing per-page images from PDFs or passing through images.
type Splitter interface {
	Split(ctx context.Context, req Request) (Result, error)
}

// Factory constructs a splitter instance from configuration.
type Factory func(config.SplitterConfig) (Splitter, error)

// DefaultFactories provides the built-in splitter registry.
func DefaultFactories() map[string]Factory {
	return map[string]Factory{
		"poppler": newPopplerSplitter,
		"remote":  newRemoteSplitter,
	}
}

func newPopplerSplitter(cfg config.SplitterConfig) (Splitter, error) {
	return &popplerSplitter{cfg: cfg}, nil
}

func newRemoteSplitter(cfg config.SplitterConfig) (Splitter, error) {
	if cfg.Options == nil {
		return nil, fmt.Errorf("remote splitter requires options.base_url")
	}
	baseURL, _ := cfg.Options["base_url"].(string)
	if baseURL == "" {
		return nil, fmt.Errorf("remote splitter requires options.base_url")
	}
	return &remoteSplitter{
		cfg:     cfg,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: resolveRemoteTimeout(cfg.Options)},
	}, nil
}

type popplerSplitter struct {
	cfg             config.SplitterConfig
	execCommand     func(context.Context, string, string, []string) error
	detectPageCount func(context.Context, string) (int, error)
}

func (p *popplerSplitter) Split(ctx context.Context, req Request) (Result, error) {
	inputPath, err := filepath.Abs(req.Path)
	if err != nil {
		return Result{}, fmt.Errorf("poppler splitter: resolve path: %w", err)
	}

	if _, err := os.Stat(inputPath); err != nil {
		return Result{}, fmt.Errorf("poppler splitter: %w", err)
	}

	workDir, err := os.MkdirTemp("", "ocrx-poppler-*")
	if err != nil {
		return Result{}, fmt.Errorf("poppler splitter: create temp dir: %w", err)
	}
	cleanup := func() error {
		return os.RemoveAll(workDir)
	}

	dpi := p.cfg.DPI
	if dpi <= 0 {
		dpi = 300
	}

	limit := resolveMaxPages(req.MaxPages, p.cfg.Options)
	requestedBatchPages := p.cfg.BatchPages
	effectiveBatchPages := 0
	maxParallel := 1
	batchCount := 1
	batchingMode := "single"
	fallbackReason := ""
	totalPages := limit

	if requestedBatchPages > 0 {
		if totalPages <= 0 {
			detectedPages, detectErr := p.lookupPageCount(ctx, inputPath)
			if detectErr != nil || detectedPages <= 0 {
				fallbackReason = "page_count_unavailable"
			} else {
				totalPages = detectedPages
			}
		}
		if totalPages > 0 {
			effectiveBatchPages = requestedBatchPages
			if effectiveBatchPages > totalPages {
				effectiveBatchPages = totalPages
			}
		}
	}

	var files []string
	if totalPages > 0 && effectiveBatchPages > 0 && totalPages > effectiveBatchPages {
		maxParallel = p.cfg.MaxParallel
		if maxParallel <= 0 {
			maxParallel = 1
		}
		files, err = p.runBatches(ctx, inputPath, workDir, dpi, totalPages, effectiveBatchPages, maxParallel)
		if err != nil {
			cleanup()
			return Result{}, err
		}
		batchCount = len(buildPageRanges(totalPages, effectiveBatchPages))
		batchingMode = "batched"
	} else {
		endPage := totalPages
		files, err = p.runRange(ctx, inputPath, workDir, dpi, 1, endPage, 0)
		if err != nil {
			cleanup()
			return Result{}, err
		}
		if effectiveBatchPages > 0 {
			maxParallel = p.cfg.MaxParallel
			if maxParallel <= 0 {
				maxParallel = 1
			}
		}
	}

	stats := map[string]any{
		"dpi":             dpi,
		"work_dir":        workDir,
		"pages":           len(files),
		"batch_pages":     effectiveBatchPages,
		"max_parallel":    maxParallel,
		"batches":         batchCount,
		"batching_mode":   batchingMode,
		"requested_limit": limit,
	}
	if fallbackReason != "" {
		stats["batching_fallback"] = fallbackReason
	}

	return Result{
		Images:  files,
		Stats:   stats,
		Cleanup: cleanup,
	}, nil
}

type remoteSplitter struct {
	cfg     config.SplitterConfig
	baseURL string
	client  *http.Client
}

func (r *remoteSplitter) Split(ctx context.Context, req Request) (Result, error) {
	if r.baseURL == "" {
		return Result{}, errors.New("remote splitter: base_url missing")
	}
	file, err := os.Open(req.Path)
	if err != nil {
		return Result{}, fmt.Errorf("remote splitter: open file: %w", err)
	}

	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/split", pipeReader)
	if err != nil {
		_ = pipeReader.Close()
		_ = pipeWriter.Close()
		_ = file.Close()
		return Result{}, fmt.Errorf("remote splitter: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	streamErrCh := make(chan error, 1)
	go streamRemoteMultipart(file, filepath.Base(req.Path), req.MaxPages, writer, pipeWriter, streamErrCh)

	resp, err := r.httpClient().Do(httpReq)
	if err != nil {
		_ = pipeReader.Close()
		streamErr := <-streamErrCh
		if streamErr != nil && !errors.Is(streamErr, io.ErrClosedPipe) && !errors.Is(streamErr, context.Canceled) {
			return Result{}, fmt.Errorf("remote splitter: stream: %w", streamErr)
		}
		return Result{}, fmt.Errorf("remote splitter: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		_ = drainStreamError(streamErrCh)
		return Result{}, fmt.Errorf("remote splitter: status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var parsed struct {
		Images []string `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		_ = drainStreamError(streamErrCh)
		return Result{}, fmt.Errorf("remote splitter: decode: %w", err)
	}
	if err := drainStreamError(streamErrCh); err != nil {
		return Result{}, fmt.Errorf("remote splitter: stream: %w", err)
	}
	if len(parsed.Images) == 0 {
		return Result{}, errors.New("remote splitter: no images returned")
	}
	sort.Strings(parsed.Images)
	if req.MaxPages > 0 && len(parsed.Images) > req.MaxPages {
		parsed.Images = parsed.Images[:req.MaxPages]
	}
	return Result{
		Images: parsed.Images,
		Stats: map[string]any{
			"pages":   len(parsed.Images),
			"kind":    "remote",
			"timeout": r.httpClient().Timeout.String(),
		},
	}, nil
}

func (p *popplerSplitter) runBatches(ctx context.Context, inputPath, workDir string, dpi, totalPages, batchPages, maxParallel int) ([]string, error) {
	ranges := buildPageRanges(totalPages, batchPages)
	results := make([][]string, len(ranges))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for idx, currentRange := range ranges {
		wg.Add(1)
		go func(i int, r pageRange) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			files, err := p.runRange(ctx, inputPath, workDir, dpi, r.start, r.end, i+1)
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
				return
			}
			results[i] = files
		}(idx, currentRange)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	combined := make([]string, 0, totalPages)
	for _, files := range results {
		combined = append(combined, files...)
	}
	return combined, nil
}

func (p *popplerSplitter) runRange(ctx context.Context, inputPath, workDir string, dpi, startPage, endPage, batchIndex int) ([]string, error) {
	args := []string{"-png", "-r", strconv.Itoa(dpi)}
	if startPage > 0 {
		args = append(args, "-f", strconv.Itoa(startPage))
	}
	if endPage > 0 {
		args = append(args, "-l", strconv.Itoa(endPage))
	}

	prefix := "page"
	if batchIndex > 0 {
		prefix = fmt.Sprintf("page-b%04d", batchIndex)
	}
	outPrefix := filepath.Join(workDir, prefix)
	args = append(args, inputPath, outPrefix)

	if err := p.runCommand(ctx, workDir, "pdftocairo", args); err != nil {
		return nil, fmt.Errorf("poppler splitter: pdftocairo: %w", err)
	}

	files, err := filepath.Glob(outPrefix + "-*.png")
	if err != nil {
		return nil, fmt.Errorf("poppler splitter: glob: %w", err)
	}
	sortPopplerImages(files)
	if len(files) == 0 {
		return nil, errors.New("poppler splitter: no images produced")
	}
	return files, nil
}

func (p *popplerSplitter) runCommand(ctx context.Context, workDir, name string, args []string) error {
	if p.execCommand != nil {
		return p.execCommand(ctx, workDir, name, args)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	return cmd.Run()
}

func (p *popplerSplitter) lookupPageCount(ctx context.Context, inputPath string) (int, error) {
	if p.detectPageCount != nil {
		return p.detectPageCount(ctx, inputPath)
	}

	cmd := exec.CommandContext(ctx, "pdfinfo", inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Pages:") {
			continue
		}
		pages, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
		if err != nil {
			return 0, fmt.Errorf("pdfinfo: parse pages: %w", err)
		}
		if pages <= 0 {
			return 0, errors.New("pdfinfo: pages must be positive")
		}
		return pages, nil
	}
	return 0, errors.New("pdfinfo: pages not found")
}

func buildPageRanges(totalPages, batchPages int) []pageRange {
	if totalPages <= 0 || batchPages <= 0 {
		return nil
	}
	ranges := make([]pageRange, 0, (totalPages+batchPages-1)/batchPages)
	for start := 1; start <= totalPages; start += batchPages {
		end := start + batchPages - 1
		if end > totalPages {
			end = totalPages
		}
		ranges = append(ranges, pageRange{start: start, end: end})
	}
	return ranges
}

type pageRange struct {
	start int
	end   int
}

func sortPopplerImages(files []string) {
	sort.Slice(files, func(i, j int) bool {
		left, okLeft := popplerPageNumber(files[i])
		right, okRight := popplerPageNumber(files[j])
		switch {
		case okLeft && okRight && left != right:
			return left < right
		case okLeft != okRight:
			return okLeft
		default:
			return files[i] < files[j]
		}
	})
}

func popplerPageNumber(path string) (int, bool) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	idx := strings.LastIndex(name, "-")
	if idx < 0 || idx+1 >= len(name) {
		return 0, false
	}
	value, err := strconv.Atoi(name[idx+1:])
	if err != nil {
		return 0, false
	}
	return value, true
}

func resolveMaxPages(reqMaxPages int, options map[string]any) int {
	if reqMaxPages > 0 {
		return reqMaxPages
	}
	if options == nil {
		return 0
	}
	value, ok := intFromAny(options["max_pages"])
	if !ok || value <= 0 {
		return 0
	}
	return value
}

func intFromAny(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func resolveRemoteTimeout(options map[string]any) time.Duration {
	if options != nil {
		if timeout, ok := durationFromAny(options["timeout"]); ok && timeout > 0 {
			return timeout
		}
	}
	return 60 * time.Second
}

func durationFromAny(value any) (time.Duration, bool) {
	switch v := value.(type) {
	case time.Duration:
		return v, true
	case string:
		parsed, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return parsed, true
	case int:
		return time.Duration(v) * time.Second, true
	case int8:
		return time.Duration(v) * time.Second, true
	case int16:
		return time.Duration(v) * time.Second, true
	case int32:
		return time.Duration(v) * time.Second, true
	case int64:
		return time.Duration(v) * time.Second, true
	case float32:
		return time.Duration(v * float32(time.Second)), true
	case float64:
		return time.Duration(v * float64(time.Second)), true
	default:
		return 0, false
	}
}

func (r *remoteSplitter) httpClient() *http.Client {
	if r.client != nil {
		return r.client
	}
	r.client = &http.Client{Timeout: resolveRemoteTimeout(r.cfg.Options)}
	return r.client
}

func streamRemoteMultipart(file *os.File, filename string, maxPages int, writer *multipart.Writer, pipeWriter *io.PipeWriter, errCh chan<- error) {
	defer close(errCh)
	defer file.Close()

	var streamErr error
	defer func() {
		if closeErr := writer.Close(); streamErr == nil && closeErr != nil {
			streamErr = closeErr
		}
		if streamErr != nil {
			_ = pipeWriter.CloseWithError(streamErr)
		} else {
			_ = pipeWriter.Close()
		}
		errCh <- streamErr
	}()

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		streamErr = err
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		streamErr = err
		return
	}
	if maxPages > 0 {
		if err := writer.WriteField("max_pages", strconv.Itoa(maxPages)); err != nil {
			streamErr = err
			return
		}
	}
}

func drainStreamError(errCh <-chan error) error {
	streamErr := <-errCh
	if streamErr == nil || errors.Is(streamErr, io.ErrClosedPipe) || errors.Is(streamErr, context.Canceled) {
		return nil
	}
	return streamErr
}
