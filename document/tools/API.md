# Document Tools API（Experimental）

`github.com/wsnacj/agentx-go/document/tools` 提供 AgentX 推荐的文档解析与 PDF
工具入口。该包拥有 tool schema、参数归一化、结果压缩、PDF focus/evidence、structured
projection、缓存键和 unified route coordination；它不拥有文件授权、网络、credential、
provider 配置、OCR 配置发现、native PDF 请求、页面渲染或 artifact 写入策略。

当前状态为 **Experimental / Developer Preview candidate**，不是 Public/Beta/Stable 兼容承诺。

## 入口

```go
func RegisterDocumentParseTools(tool.Registrar, DocumentParseToolOptions) error
func RegisterPDFTools(tool.Registrar, PDFToolOptions) error

func DocumentParseDefinition() llm.Tool
func DocumentSpecRecommendDefinition() llm.Tool
func PDFDefinition(name string, unifiedAvailable bool) (llm.Tool, bool)
```

文档工具名：

- `document_parse`：使用显式 `DocumentParser` 运行 canonical document pipeline；
- `document_spec_recommend`：对 Host 提供的页面文本和候选 spec 做 deterministic 排序；
- `pdf`：统一 PDF 问答、总结和多文档比较；
- `pdf_extract`、`pdf_read_pages`、`pdf_outline`：精确 specialist 路径；
- `pdf_analyze`、`pdf_extract_structured`：focus、evidence、visual/structured 路径。

`EnabledTools` 为空时延续兼容行为：允许注册该组全部可用工具。`document_parse` 与
`document_spec_recommend` 在 HS 兼容入口中仍由调用方显式启用。PDF unified `pdf` 只有在
调用方提供至少一个 `Models` candidate 时才可注册；具体 provider 不由本包发现。

## Document Host

```go
type DocumentHost struct {
    Runtime   DocumentParser
    Paths     PathResolver
    Text      DocumentTextLoader
    Artifacts ArtifactLister
    Errors    ErrorProjector
}
```

- `Runtime` 与 `Paths` 是必需项；
- `Text` 只在 `document_spec_recommend` 执行时需要；
- `Artifacts` 未提供时不返回 `files_touched`，不会自行扫描未授权目录；
- `Errors` 负责把 Host/provider 错误投影为稳定分类和 display-safe 文本；
- `*pipeline.Runtime` 直接满足 `DocumentParser`，兼容 Host 可用 `DocumentParserFunc`。

## PDF Host

```go
type PDFHost struct {
    Inputs          PDFInputResolver
    LayoutName      string
    Layout          func(context.Context, string, []int) (PDFTextResult, error)
    Chat            func(context.Context, llm.ChatInput) (*llm.ChatResponse, error)
    Vision          func(context.Context, llm.VisionInput) (*llm.VisualResponse, error)
    Native          func(context.Context, PDFNativeRequest) (string, error)
    OCR             func(context.Context, PDFOCRRequest) ([]PDFPageText, error)
    Render          func(context.Context, string, []int, int) ([]PDFRenderedPage, func() error, error)
    PublishRendered func(context.Context, string, string, []PDFRenderedPage) ([]mediaartifact.Descriptor, error)
}
```

`Inputs` 是必需项，统一处理 workspace path、`file://` 和远程 PDF 的授权与物化。`Layout`
用于 Host 显式提供 layout-preserving 文本提取；其余能力
按实际路由可选；缺少能力时保持 fail closed 或走既有 deterministic fallback，不会读取环境
变量、发起网络请求或启动进程。

`PDFToolOptions.Backend` 必须由 Host 显式提供并实现 `PDFBackend`。可选
`PDFLayoutBackend` 用于 layout-preserving evidence。`Models` 是 Host 已解析的模型候选快照，
可包含 `NativePDF`、`SupportsVision` 和 opaque `ConfigKey`；本包不会从配置文件或凭据环境
重新发现模型。

`Render` 返回的 `PDFRenderedPage` 必须携带页面图像 `Data`，可选 `MIMEType`；`Path` 只用于
Host 自己的 artifact 发布。canonical 不读取 Host 文件系统。

## 并发、取消和清理

- 注册后的 handler 将调用方 `context.Context` 传递给 input/backend/model/OCR/render ports；
- cancellation/deadline 原样返回，Host 不应在 adapter 内替换为 `context.Background()`；
- `ResolvedPDFInput.Cleanup` 在每次调用结束时执行；多输入中途失败也会清理已物化输入；
- cache 为进程内 bounded cache，key 保留原 PDF identity、page selection、query 和显式 OCR
  profile；它不持有 credential；
- 并发安全取决于注入的 Host ports 和 backend。canonical cache 本身支持并发访问。

## 明确 non-goal

- 默认 provider、model route、credential/env/file discovery；
- 私网/代理/CIDR/端口授权；
- 真实 HTTP fetch、native provider 请求、Poppler/Python 启动；
- 客户 spec、财务模板、review/evaluator、Scene 或 Runner policy；
- Public/Beta/Stable 或正式发行承诺。
