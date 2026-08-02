// pdf_parser.go - 修复后的版本
package pdfparser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// TableResponse 表示完整的 PDF解析 API 响应
type TableResponse struct {
	Msg      string  `json:"message"` // 注意：Python返回的是"message"，不是"msg"
	Code     int     `json:"code"`
	Version  string  `json:"version"`
	Duration int     `json:"duration"`
	Result   TResult `json:"result"`
}

// TResult 包含 Pages
type TResult struct {
	Pages []TPage `json:"pages"`
}

// TPage 表示每个页面的解析结果
type TPage struct {
	Angle  int     `json:"angle"`
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Tables []Table `json:"tables"`
	Images []Image `json:"images,omitempty"` // 图片信息（可选）
}

// Image 表示提取的图片
type Image struct {
	Index            int    `json:"index"`
	Position         []int  `json:"position"` // [x0, y0, x1, y1]
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	Data             string `json:"data"`                        // base64编码的图片数据
	Format           string `json:"format"`                      // 图片格式
	SizeBytes        int    `json:"size_bytes,omitempty"`        // 图片大小（字节）
	ExtractionMethod string `json:"extraction_method,omitempty"` // 提取方法：pdfimages, pymupdf, page_render
}

// Table 表示表格结构
type Table struct {
	HeightOfRows []int       `json:"height_of_rows"`
	Type         string      `json:"type"`
	TableCells   []TableCell `json:"table_cells"`
	TableRows    int         `json:"table_rows"`
	WidthOfCols  []int       `json:"width_of_cols"`
	Position     []int       `json:"position"`
	Lines        []TLine     `json:"lines"`
	TableCols    int         `json:"table_cols"`
}

// TableCell 表示表格单元格
type TableCell struct {
	StartRow int    `json:"start_row"`
	StartCol int    `json:"start_col"`
	EndRow   int    `json:"end_row"`
	EndCol   int    `json:"end_col"`
	Text     string `json:"text"`
	Borders  struct {
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
		Left   int `json:"left"`
		Top    int `json:"top"`
	} `json:"borders"`
	Position []int   `json:"position"`
	Lines    []TLine `json:"lines"`
}

// TLine 表示 OCR 识别的文本行，包含字符级信息
type TLine struct {
	Angle       int     `json:"angle"`
	Text        string  `json:"text"`
	Direction   int     `json:"direction"`
	Handwritten int     `json:"handwritten"`
	Position    []int   `json:"position"` // [x0, y0, x1, y1]
	Score       float64 `json:"score"`
	Type        string  `json:"type"`
	// 以下字段在 needCharacter 为 true 时返回
	CharAttributes      []string    `json:"char_attributes,omitempty"`
	CharCandidates      [][]string  `json:"char_candidates,omitempty"`
	CharCandidatesScore [][]float64 `json:"char_candidates_score,omitempty"`
	CharCenters         [][]int     `json:"char_centers,omitempty"`
	CharPositions       [][]int     `json:"char_positions,omitempty"`
	CharScores          []float64   `json:"char_scores,omitempty"`
}

// Runner 是 PDF parser 唯一允许的外部执行端口。实现可以调用 Python、远程服务或
// deterministic fake；canonical parser 不发现解释器、不读取环境变量，也不自行启动进程。
type Runner interface {
	Run(context.Context, RunRequest) (RunResult, error)
}

// RunRequest 是发送给显式 Runner 的解析请求。
type RunRequest struct {
	PDFPath string
	Options PDFParserOptions
}

// RunResult 保留 adapter 的 stdout/stderr，供兼容错误投影与诊断使用。
type RunResult struct {
	Stdout []byte
	Stderr []byte
}

// ParserOption 配置 parser 本身而非具体 backend。
type ParserOption func(*PDFParser)

// WithTimeout 设置单次调用的默认上限。零值表示使用五分钟兼容默认值。
func WithTimeout(timeout time.Duration) ParserOption {
	return func(parser *PDFParser) {
		if timeout > 0 {
			parser.timeout = timeout
		}
	}
}

// PDFParser PDF解析器。构造时必须显式注入 Runner。
type PDFParser struct {
	runner  Runner
	timeout time.Duration
}

// NewParser 使用显式 Runner 构造 parser。
func NewParser(runner Runner, opts ...ParserOption) (*PDFParser, error) {
	if runner == nil {
		return nil, fmt.Errorf("pdfparser runner is required")
	}
	parser := &PDFParser{runner: runner, timeout: 5 * time.Minute}
	for _, opt := range opts {
		if opt != nil {
			opt(parser)
		}
	}
	return parser, nil
}

// ParsePDF 解析PDF文件
func (p *PDFParser) ParsePDF(pdfPath string, needCharacter bool) (*TableResponse, error) {
	return p.ParsePDFContext(context.Background(), pdfPath, needCharacter)
}

// ParsePDFContext 解析 PDF 并传播调用方 cancellation/deadline。
func (p *PDFParser) ParsePDFContext(ctx context.Context, pdfPath string, needCharacter bool) (*TableResponse, error) {
	// 首先检查PDF文件是否存在
	if err := IsValidPDFPath(pdfPath); err != nil {
		return nil, fmt.Errorf("PDF文件路径错误: %v", err)
	}
	if p == nil || p.runner == nil {
		return nil, fmt.Errorf("PDF解析失败: pdfparser runner is required")
	}
	ctx, cancel := p.withTimeout(ctx)
	defer cancel()
	result, err := p.runner.Run(ctx, RunRequest{
		PDFPath: pdfPath,
		Options: PDFParserOptions{OutputFormat: "json", PageRange: "all", TableEngine: "hybrid", NeedCharacter: needCharacter},
	})
	if err != nil {
		// 处理超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("PDF parsing timeout after 5 minutes")
		}

		// 尝试解析stderr中的错误JSON
		if len(result.Stderr) > 0 {
			var errorResp TableResponse
			if jsonErr := json.Unmarshal(result.Stderr, &errorResp); jsonErr == nil {
				return &errorResp, nil // 返回错误响应而不是错误
			}
		}

		// 如果stderr中有内容，分析错误类型并提供更清晰的提示
		if len(result.Stderr) > 0 {
			stderrStr := string(result.Stderr)
			log.Printf("Python script stderr: %s", stderrStr)

			// 检查是否是Python脚本文件不存在的错误
			if strings.Contains(stderrStr, "can't open file") && strings.Contains(stderrStr, "pdfparser.py") {
				return nil, fmt.Errorf("PDF解析脚本不存在，请检查pdfparser.py文件路径: %v", err)
			}

			// 检查是否是PDF文件相关的错误
			if strings.Contains(stderrStr, "No such file") || strings.Contains(stderrStr, "FileNotFoundError") {
				return nil, fmt.Errorf("PDF文件不存在或无法访问，请检查文件路径: %s", pdfPath)
			}
		}

		return nil, fmt.Errorf("PDF解析失败: %w", err)
	}

	output := result.Stdout

	// 清理可能的MuPDF错误信息污染
	cleanOutput := cleanMuPDFOutput(output)

	// 解析JSON输出
	var response TableResponse
	if err := json.Unmarshal(cleanOutput, &response); err != nil {
		// 记录原始输出以便调试
		log.Printf("Failed to parse JSON response. Raw output: %s", string(output))
		log.Printf("Cleaned output: %s", string(cleanOutput))
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 检查响应状态
	if response.Code != 0 {
		return &response, fmt.Errorf("parsing failed: %s", response.Msg)
	}

	return &response, nil
}

func (p *PDFParser) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := p.timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return context.WithTimeout(ctx, timeout)
}

// IsValidPDFPath 检查PDF文件是否存在
func IsValidPDFPath(pdfPath string) error {
	if pdfPath == "" {
		return fmt.Errorf("PDF文件路径为空")
	}

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF文件不存在: %s", pdfPath)
	}

	return nil
}

// ParsePDFSafe 安全的PDF解析，带预检查
func (p *PDFParser) ParsePDFSafe(pdfPath string, needCharacter bool) (*TableResponse, error) {
	// 预检查
	if err := IsValidPDFPath(pdfPath); err != nil {
		return nil, err
	}

	return p.ParsePDF(pdfPath, needCharacter)
}

// ParsePDFToText 解析PDF并返回按行规整的文本
func (p *PDFParser) ParsePDFToText(pdfPath string, needCharacter bool) (string, error) {
	// 先解析PDF
	response, err := p.ParsePDF(pdfPath, needCharacter)
	if err != nil {
		return "", err
	}

	if response.Code != 0 {
		return "", fmt.Errorf("parsing failed: %s", response.Msg)
	}

	// 格式化为文本
	formatter := NewTextFormatter(response)
	return formatter.FormatToText(), nil
}

// ParsePDFToLines 解析PDF并返回行数组
func (p *PDFParser) ParsePDFToLines(pdfPath string, needCharacter bool) ([]string, error) {
	// 先解析PDF
	response, err := p.ParsePDF(pdfPath, needCharacter)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("parsing failed: %s", response.Msg)
	}

	// 格式化为行数组
	formatter := NewTextFormatter(response)
	return formatter.FormatToLines(), nil
}

// ParsePDFToPages 解析PDF并返回按页面分组的文本数组
func (p *PDFParser) ParsePDFToPages(pdfPath string, needCharacter bool) ([]string, error) {
	// 先解析PDF
	response, err := p.ParsePDF(pdfPath, needCharacter)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("parsing failed: %s", response.Msg)
	}

	// 格式化为页面数组
	formatter := NewTextFormatter(response)
	return formatter.FormatToPages(), nil
}

// ParsePDFToTextOnly 解析PDF并返回纯文本（不包含表格格式）
func (p *PDFParser) ParsePDFToTextOnly(pdfPath string, needCharacter bool) (string, error) {
	// 先解析PDF
	response, err := p.ParsePDF(pdfPath, needCharacter)
	if err != nil {
		return "", err
	}

	if response.Code != 0 {
		return "", fmt.Errorf("parsing failed: %s", response.Msg)
	}

	// 提取纯文本
	formatter := NewTextFormatter(response)
	return formatter.ExtractTextOnly(), nil
}

// ParsePDFWithOptionsToText 使用自定义选项解析PDF并返回文本
func (p *PDFParser) ParsePDFWithOptionsToText(pdfPath string, opts *PDFParserOptions) (string, error) {
	// 先解析PDF
	response, err := p.ParsePDFWithOptions(pdfPath, opts)
	if err != nil {
		return "", err
	}

	if response.Code != 0 {
		return "", fmt.Errorf("parsing failed: %s", response.Msg)
	}

	// 格式化为文本
	formatter := NewTextFormatter(response)
	return formatter.FormatToText(), nil
}

// cleanMuPDFOutput 清理MuPDF错误信息，确保只返回纯净的JSON
func cleanMuPDFOutput(output []byte) []byte {
	outputStr := string(output)

	// 查找JSON的开始位置 - 寻找第一个 '{'
	jsonStart := strings.Index(outputStr, "{")
	if jsonStart == -1 {
		// 如果没有找到JSON开始，返回原始输出
		return output
	}

	// 如果JSON不是从头开始，说明前面有错误信息
	if jsonStart > 0 {
		// 记录被清理的错误信息
		errorPart := strings.TrimSpace(outputStr[:jsonStart])
		if errorPart != "" {
			log.Printf("Cleaned MuPDF error from output: %s", errorPart)
		}

		// 返回从JSON开始的部分
		return []byte(outputStr[jsonStart:])
	}

	// JSON从头开始，直接返回原始输出
	return output
}
