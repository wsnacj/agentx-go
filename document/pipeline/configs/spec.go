package configs

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DocSpec 驱动整个解析流程的规格
type DocSpec struct {
	Meta        MetaSpec         `yaml:"meta"`
	Chapters    []ChapterSpec    `yaml:"chapters"`
	Validations []ValidationSpec `yaml:"validations,omitempty"`

	// ConfigDir 配置文件夹路径（运行时设置，不在YAML中配置）
	ConfigDir string `yaml:"-"`
}

type MetaSpec struct {
	DocType     string `yaml:"doc_type"`
	Version     string `yaml:"version"`
	HeaderLines int    `yaml:"header_lines"`
	FooterLines int    `yaml:"footer_lines"`

	// DetectPriorityGroups 章节归类优先级分组，groups[0] 为最高优先级；同优先级并发归类
	DetectPriorityGroups [][]string `yaml:"detect_priority_groups,omitempty"`
	// DetectBatchPages 每个并发批次包含的页面数（如 100）
	DetectBatchPages int `yaml:"detect_batch_pages,omitempty"`
	// DetectMaxConcurrent 最大并发批次数；<=0 表示按 pages/DetectBatchPages 动态计算
	DetectMaxConcurrent int `yaml:"detect_max_concurrent,omitempty"`

	// DetectBodyPatterns 可选：用于从页内提取关键信息（小标题/段首句等）的正则列表
	DetectBodyPatterns []string `yaml:"detect_body_patterns,omitempty"`
	// DetectBodyMaxMatches 单页最多提取的"页中匹配"条目数（总数），<=0 使用默认 2
	DetectBodyMaxMatches int `yaml:"detect_body_max_matches,omitempty"`
	// DetectBodyClip 单条"页中匹配"最大字符数，<=0 使用与 LineClip 相同
	DetectBodyClip int `yaml:"detect_body_clip,omitempty"`

	// AttemptTimeout 单次LLM请求超时时间（秒）
	AttemptTimeout int `yaml:"attempt_timeout,omitempty"`
	// TotalTimeout LLM总超时时间（秒）
	TotalTimeout int `yaml:"total_timeout,omitempty"`
	// MaxChunkChars 每个章节文本最大字符数
	MaxChunkChars int `yaml:"max_chunk_chars,omitempty"`
	// PDFParseMode PDF 文本提取模式：simple | normal | ocr
	PDFParseMode string `yaml:"pdf_parse_mode,omitempty"`
	// HeaderFooterCleanup 页眉页脚清理模式：none | programmatic | llm | auto
	HeaderFooterCleanup string `yaml:"header_footer_cleanup,omitempty"`
}

type ChapterSpec struct {
	Key           string      `yaml:"key"`
	TitleKeywords []string    `yaml:"title_keywords"`
	Fields        []FieldSpec `yaml:"fields"`
	LLMPrompt     string      `yaml:"llm_prompt,omitempty"`
	// Priority 可选，章节归类优先级（正整数，数值越小优先级越高）。
	// 未配置或 <=0 的章节会在使用优先级模式时自动归入“最低优先级”的一组。
	Priority int `yaml:"priority,omitempty"`
}

type FieldSpec struct {
	Key               string          `yaml:"key"`
	Type              string          `yaml:"type"` // string|number|date|enum|array
	Required          bool            `yaml:"required"`
	Unit              string          `yaml:"unit,omitempty"`
	Normalize         string          `yaml:"normalize,omitempty"` // e.g. number
	Extractors        []ExtractorSpec `yaml:"extractors"`
	DerivedFormula    string          `yaml:"derived,omitempty"` // 简易表达式 a/b 等
	Description       string          `yaml:"description,omitempty"`
	Aliases           []string        `yaml:"aliases,omitempty"`
	PreferredChapters []string        `yaml:"preferred_chapters,omitempty"`
	PeriodPolicy      string          `yaml:"period_policy,omitempty"` // any|current
	UnitPolicy        string          `yaml:"unit_policy,omitempty"`   // any|prefer|required
	DisallowPatterns  []string        `yaml:"disallow_patterns,omitempty"`
	PreferPatterns    []string        `yaml:"prefer_patterns,omitempty"`
}

// ValidationSpec 定义跨字段/跨章节的校验规则
// expr 支持比较符号 > >= < <= == != 以及 approx(a,b,tol), between(x,a,b)，以及 && || 组合
type ValidationSpec struct {
	Name     string `yaml:"name"`
	Expr     string `yaml:"expr"`
	Severity string `yaml:"severity"` // warn|error
	Message  string `yaml:"message,omitempty"`
}

type ExtractorSpec struct {
	Type string `yaml:"type"` // regex|table|llm|script
	// regex
	Scope   string `yaml:"scope,omitempty"`   // header|full|footer
	Pattern string `yaml:"pattern,omitempty"` // 正则
	// table
	RowLabels     []string `yaml:"row_labels,omitempty"`     // row labels or aliases to match
	ColumnLabels  []string `yaml:"column_labels,omitempty"`  // preferred column labels, e.g. current period
	ValueColumn   int      `yaml:"value_column,omitempty"`   // 1-based value column fallback
	MaxCandidates int      `yaml:"max_candidates,omitempty"` // per extractor candidate cap
	// script
	Script string `yaml:"script,omitempty"` // normalize_number 等
	// llm
	Prompt string `yaml:"prompt,omitempty"` // 可选，优先使用 chapter 的 llm_prompt
}

func LoadSpec(configPath string) (*DocSpec, error) {
	var mainYamlPath string

	// 判断输入是文件还是文件夹
	info, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// 如果是文件夹，自动拼接main.yaml
		mainYamlPath = filepath.Join(configPath, "main.yaml")
	} else {
		// 如果是文件，直接使用
		mainYamlPath = configPath
	}

	b, err := os.ReadFile(mainYamlPath)
	if err != nil {
		return nil, err
	}
	var s DocSpec
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	// 如果输入是文件夹，设置配置文件夹路径用于后续文件发现
	if info.IsDir() {
		s.ConfigDir = configPath
	} else {
		// 如果输入是文件，使用文件所在目录
		s.ConfigDir = filepath.Dir(configPath)
	}

	// defaults
	if s.Meta.HeaderLines <= 0 {
		s.Meta.HeaderLines = 6
	}
	if s.Meta.FooterLines <= 0 {
		s.Meta.FooterLines = 4
	}
	if s.Meta.AttemptTimeout <= 0 {
		s.Meta.AttemptTimeout = 90 // 默认90秒
	}
	if s.Meta.TotalTimeout <= 0 {
		s.Meta.TotalTimeout = 240 // 默认4分钟
	}
	if s.Meta.MaxChunkChars <= 0 {
		s.Meta.MaxChunkChars = 12000 // 默认12000字符
	}
	if s.Meta.PDFParseMode == "" {
		s.Meta.PDFParseMode = "simple" // 默认使用简单模式
	}
	if s.Meta.HeaderFooterCleanup == "" {
		s.Meta.HeaderFooterCleanup = "none" // 默认不清理
	}

	return &s, nil
}
