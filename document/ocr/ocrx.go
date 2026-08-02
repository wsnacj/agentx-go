package ocrx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ocrxconfig "github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
)

// Mode 是快捷接口使用的识别类型。
type Mode = OperationKind

const (
	ModeOCR   Mode = OperationKindOCR
	ModeTable Mode = OperationKindTable
	ModeStamp Mode = OperationKindStamp
)

// Client 封装 Service，提供按需的便捷接口。
type Client struct {
	service     *Service
	defaultOpts RequestOptions
}

// ClientOption 配置 Client 的默认行为。
type ClientOption func(*Client)

// RequestOptions 控制单次请求。
type RequestOptions struct {
	NeedCharacter *bool
	DiffEnabled   *bool
	DiffBaseline  *string
	DiffPreview   *int
	MaxPages      *int
	Context       context.Context
}

// Option 覆盖单次请求的设置。
type Option func(*RequestOptions)

// BaseResult 为快捷结果的通用部分。
type BaseResult struct {
	Source string
	Mode   Mode
	Meta   Meta
	Diff   *DiffSummary
}

// TextResult 返回拼接后的纯文本。
type TextResult struct {
	BaseResult
	Text           string
	NormalizedText string
}

// HTMLResult 返回整体 HTML 及按页 HTML。
type HTMLResult struct {
	BaseResult
	HTML  string
	Pages []string
}

// PageTextResult 返回按页文本。
type PageTextResult struct {
	BaseResult
	Pages []PageText
}

type PageText struct {
	Index          int
	Text           string
	NormalizedText string
}

// TableResult 返回表格识别结果。
type TableResult struct {
	BaseResult
	Pages                  []TablePageResult
	CombinedText           string
	NormalizedCombinedText string
}

type TablePageResult struct {
	Index          int
	Text           string
	NormalizedText string
	HTML           string
	Coordinates    []Coordinate
}

// StampResult 返回印章识别结果。
type StampResult struct {
	BaseResult
	Pages []StampPageResult
}

type StampPageResult struct {
	Index  int
	Stamps []StampDetail
}

// FileResult 用于批量场景。
type FileResult struct {
	BaseResult
	Output interface{}
	Err    error
}

// NewClientFromConfig 从 YAML 配置构建 Client。
func NewClientFromConfig(path string, opts ...ClientOption) (*Client, error) {
	cfg, err := ocrxconfig.Load(path)
	if err != nil {
		return nil, err
	}
	if err := applyProviderConfigDefaults(&cfg); err != nil {
		return nil, err
	}
	svc, err := NewService(cfg, Dependencies{})
	if err != nil {
		return nil, err
	}
	return NewClientFromService(svc, opts...), nil
}

// NewClientFromService 从现有 Service 构建 Client。
func NewClientFromService(svc *Service, opts ...ClientOption) *Client {
	c := &Client{
		service: svc,
		defaultOpts: RequestOptions{
			NeedCharacter: boolPtr(true),
			Context:       context.Background(),
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithDefaultNeedCharacter 设置默认 NeedCharacter。
func WithDefaultNeedCharacter(on bool) ClientOption {
	return func(c *Client) {
		c.defaultOpts.NeedCharacter = boolPtr(on)
	}
}

// WithDefaultDiff 配置默认 diff 行为。
func WithDefaultDiff(enabled bool, baseline string) ClientOption {
	return func(c *Client) {
		c.defaultOpts.DiffEnabled = boolPtr(enabled)
		c.defaultOpts.DiffBaseline = stringPtr(baseline)
	}
}

// WithDefaultContext 配置默认 Context。
func WithDefaultContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.defaultOpts.Context = ctx
	}
}

// WithDefaultDiffPreview 配置默认 diff 预览数量。
func WithDefaultDiffPreview(preview int) ClientOption {
	return func(c *Client) {
		c.defaultOpts.DiffPreview = intPtr(preview)
	}
}

// WithDefaultMaxPages sets a default per-request page cap for PDF inputs.
func WithDefaultMaxPages(maxPages int) ClientOption {
	return func(c *Client) {
		c.defaultOpts.MaxPages = positiveIntPtr(maxPages)
	}
}

// Option helpers.
func WithNeedCharacter(on bool) Option {
	return func(ro *RequestOptions) { ro.NeedCharacter = boolPtr(on) }
}

func WithDiffEnabled(on bool, baseline string) Option {
	return func(ro *RequestOptions) {
		ro.DiffEnabled = boolPtr(on)
		ro.DiffBaseline = stringPtr(baseline)
	}
}

func WithContext(ctx context.Context) Option {
	return func(ro *RequestOptions) { ro.Context = ctx }
}

// WithDiffPreview 覆盖本次请求的 diff 预览限制。
func WithDiffPreview(preview int) Option {
	return func(ro *RequestOptions) { ro.DiffPreview = intPtr(preview) }
}

// WithMaxPages limits how many pages a recognition request may process.
// Values <= 0 leave the service/pipeline default unchanged.
func WithMaxPages(maxPages int) Option {
	return func(ro *RequestOptions) { ro.MaxPages = positiveIntPtr(maxPages) }
}

// RecognizeText 返回拼接文本。
func (c *Client) RecognizeText(path string, opts ...Option) (TextResult, error) {
	base, payload, err := c.recognize(path, ModeOCR, opts...)
	if err != nil {
		return TextResult{BaseResult: base}, err
	}
	ocrPayload, ok := payload.(OCRPayload)
	if !ok {
		return TextResult{BaseResult: base}, fmt.Errorf("unexpected payload type %T", payload)
	}
	return TextResult{
		BaseResult:     base,
		Text:           ocrPayload.RecognizedText,
		NormalizedText: ocrPayload.NormalizedText,
	}, nil
}

// RecognizeHTML 返回整体 HTML 与按页 HTML。
func (c *Client) RecognizeHTML(path string, opts ...Option) (HTMLResult, error) {
	base, payload, err := c.recognize(path, ModeOCR, opts...)
	if err != nil {
		return HTMLResult{BaseResult: base}, err
	}
	ocrPayload, ok := payload.(OCRPayload)
	if !ok {
		return HTMLResult{BaseResult: base}, fmt.Errorf("unexpected payload type %T", payload)
	}
	return HTMLResult{BaseResult: base, HTML: ocrPayload.CombinedHTML, Pages: append([]string(nil), ocrPayload.HTMLPages...)}, nil
}

// RecognizeTextPages 返回按页文本。
func (c *Client) RecognizeTextPages(path string, opts ...Option) (PageTextResult, error) {
	base, payload, err := c.recognize(path, ModeOCR, opts...)
	if err != nil {
		return PageTextResult{BaseResult: base}, err
	}
	ocrPayload, ok := payload.(OCRPayload)
	if !ok {
		return PageTextResult{BaseResult: base}, fmt.Errorf("unexpected payload type %T", payload)
	}
	pages := make([]PageText, 0, len(ocrPayload.PageTexts))
	for i, text := range ocrPayload.PageTexts {
		page := PageText{Index: i + 1, Text: text}
		if i < len(ocrPayload.NormalizedPageTexts) {
			page.NormalizedText = ocrPayload.NormalizedPageTexts[i]
		}
		pages = append(pages, page)
	}
	return PageTextResult{BaseResult: base, Pages: pages}, nil
}

// RecognizeTable 返回表格识别结果。
func (c *Client) RecognizeTable(path string, opts ...Option) (TableResult, error) {
	base, payload, err := c.recognize(path, ModeTable, opts...)
	if err != nil {
		return TableResult{BaseResult: base}, err
	}
	tablePayload, ok := payload.(TablePayload)
	if !ok {
		return TableResult{BaseResult: base}, fmt.Errorf("unexpected payload type %T", payload)
	}
	pages := make([]TablePageResult, 0, len(tablePayload.Pages))
	for i, page := range tablePayload.Pages {
		html := ""
		if i < len(tablePayload.HTMLPages) {
			html = tablePayload.HTMLPages[i]
		}
		normalizedText := ""
		if i < len(tablePayload.NormalizedTexts) {
			normalizedText = tablePayload.NormalizedTexts[i]
		}
		pages = append(pages, TablePageResult{
			Index:          i + 1,
			Text:           page.Recognized,
			NormalizedText: normalizedText,
			HTML:           html,
			Coordinates:    append([]Coordinate(nil), page.Coordinates...),
		})
	}
	combined := strings.Join(tablePayload.Text, "\n")
	normalizedCombined := tablePayload.NormalizedCombinedText
	if normalizedCombined == "" && len(tablePayload.NormalizedTexts) > 0 {
		normalizedCombined = strings.Join(tablePayload.NormalizedTexts, "\n\n")
	}
	return TableResult{
		BaseResult:             base,
		Pages:                  pages,
		CombinedText:           combined,
		NormalizedCombinedText: normalizedCombined,
	}, nil
}

// RecognizeStamp 返回印章识别结果。
func (c *Client) RecognizeStamp(path string, opts ...Option) (StampResult, error) {
	base, payload, err := c.recognize(path, ModeStamp, opts...)
	if err != nil {
		return StampResult{BaseResult: base}, err
	}
	stampPayload, ok := payload.(StampPayload)
	if !ok {
		return StampResult{BaseResult: base}, fmt.Errorf("unexpected payload type %T", payload)
	}
	pages := make([]StampPageResult, 0, len(stampPayload.Pages))
	for _, page := range stampPayload.Pages {
		pages = append(pages, StampPageResult{Index: page.Index, Stamps: append([]StampDetail(nil), page.Stamp...)})
	}
	return StampResult{BaseResult: base, Pages: pages}, nil
}

// RecognizeFiles 批量处理文件。
func (c *Client) RecognizeFiles(paths []string, mode Mode, opts ...Option) []FileResult {
	results := make([]FileResult, 0, len(paths))
	for _, path := range paths {
		switch mode {
		case ModeOCR:
			res, err := c.RecognizeText(path, opts...)
			results = append(results, FileResult{BaseResult: res.BaseResult, Output: res, Err: err})
		case ModeTable:
			res, err := c.RecognizeTable(path, opts...)
			results = append(results, FileResult{BaseResult: res.BaseResult, Output: res, Err: err})
		case ModeStamp:
			res, err := c.RecognizeStamp(path, opts...)
			results = append(results, FileResult{BaseResult: res.BaseResult, Output: res, Err: err})
		default:
			results = append(results, FileResult{BaseResult: BaseResult{Source: path, Mode: mode}, Err: fmt.Errorf("unsupported mode %s", mode)})
		}
	}
	return results
}

// RecognizeDirectory 会递归识别目录。
func (c *Client) RecognizeDirectory(root string, mode Mode, opts ...Option) ([]FileResult, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if supportedExt(strings.ToLower(filepath.Ext(d.Name()))) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return c.RecognizeFiles(paths, mode, opts...), nil
}

// RecognizeDetailed 返回 BaseResult 和原始 Payload。
func (c *Client) RecognizeDetailed(path string, mode Mode, opts ...Option) (BaseResult, interface{}, error) {
	return c.recognize(path, mode, opts...)
}

// WriteJSONResult 将结果以 JSON 形式写入文件。
func WriteJSONResult(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteText 将文本写入文件。
func WriteText(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// PreviewText 返回文本预览。
func PreviewText(text string, limit int) string {
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) <= limit {
		return trimmed
	}
	return string(runes[:limit]) + "..."
}

// 内部调用入口。
func (c *Client) recognize(path string, mode Mode, opts ...Option) (BaseResult, interface{}, error) {
	options := c.mergeOptions(opts...)
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	req := Request{
		Paths:     []string{path},
		Options:   make(map[string]any),
		CreatedAt: time.Now(),
	}
	if val := options.MaxPages; val != nil && *val > 0 {
		req.MaxPages = *val
	} else if c.defaultOpts.MaxPages != nil && *c.defaultOpts.MaxPages > 0 {
		req.MaxPages = *c.defaultOpts.MaxPages
	}
	if val := options.NeedCharacter; val != nil {
		req.Options["need_character"] = *val
	} else if c.defaultOpts.NeedCharacter != nil {
		req.Options["need_character"] = *c.defaultOpts.NeedCharacter
	}
	if val := options.DiffEnabled; val != nil {
		req.Options["diff_enabled"] = *val
	} else if c.defaultOpts.DiffEnabled != nil {
		req.Options["diff_enabled"] = *c.defaultOpts.DiffEnabled
	}
	if baseline := options.DiffBaseline; baseline != nil {
		req.Options["diff_baseline"] = *baseline
	} else if baseline := c.defaultOpts.DiffBaseline; baseline != nil {
		req.Options["diff_baseline"] = *baseline
	}
	if preview := options.DiffPreview; preview != nil {
		req.Options["diff_preview"] = *preview
	} else if preview := c.defaultOpts.DiffPreview; preview != nil {
		req.Options["diff_preview"] = *preview
	}

	switch mode {
	case ModeOCR:
		resp, err := c.service.RecognizeOCR(ctx, req)
		base := BaseResult{Source: path, Mode: ModeOCR, Meta: resp.Meta, Diff: resp.Diff}
		return base, resp.Payload, err
	case ModeTable:
		resp, err := c.service.RecognizeTable(ctx, req)
		base := BaseResult{Source: path, Mode: ModeTable, Meta: resp.Meta, Diff: resp.Diff}
		return base, resp.Payload, err
	case ModeStamp:
		resp, err := c.service.RecognizeStamp(ctx, req)
		base := BaseResult{Source: path, Mode: ModeStamp, Meta: resp.Meta, Diff: nil}
		return base, resp.Payload, err
	default:
		return BaseResult{Source: path, Mode: mode}, nil, fmt.Errorf("unsupported mode %s", mode)
	}
}

func (c *Client) mergeOptions(opts ...Option) RequestOptions {
	merged := c.defaultOpts
	for _, opt := range opts {
		opt(&merged)
	}
	return merged
}

func applyProviderConfigDefaults(cfg *ocrxconfig.ServiceConfig) error {
	validators := provider.DefaultConfigValidators()
	for name, pipe := range cfg.Pipelines {
		if validator, ok := validators[pipe.Provider.Kind]; ok {
			p := pipe
			if err := validator(&p.Provider); err != nil {
				return fmt.Errorf("配置 %s 无效: %w", name, err)
			}
			cfg.Pipelines[name] = p
		}
	}
	return nil
}

func boolPtr(v bool) *bool       { return &v }
func stringPtr(v string) *string { return &v }
func intPtr(v int) *int          { return &v }

func positiveIntPtr(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

var supportedExtensions = map[string]struct{}{
	".pdf":  {},
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".tif":  {},
	".tiff": {},
	".gif":  {},
	".webp": {},
	".bmp":  {},
	".ico":  {},
	".icns": {},
	".sgi":  {},
	".jp2":  {},
}

func supportedExt(ext string) bool {
	_, ok := supportedExtensions[ext]
	return ok
}
