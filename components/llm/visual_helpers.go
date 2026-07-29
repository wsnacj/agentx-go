package llm

// 常用 detail 等级，可根据供应商文档扩展。
const (
	DetailAuto = "auto"
	DetailHigh = "high"
)

type VisualOption func(*VisualContent)

func WithDetail(detail string) VisualOption {
	return func(v *VisualContent) {
		if detail != "" {
			v.Detail = detail
		}
	}
}

func WithLabels(labels ...string) VisualOption {
	return func(v *VisualContent) {
		if len(labels) == 0 {
			return
		}
		v.Labels = append(v.Labels, labels...)
	}
}

func WithFPS(fps float32) VisualOption {
	return func(v *VisualContent) {
		if fps > 0 {
			v.FPS = &fps
		}
	}
}

func NewImageURL(url string, opts ...VisualOption) VisualContent {
	vc := VisualContent{
		Type:     "image_url",
		ImageURL: url,
		Detail:   DetailAuto,
	}
	for _, opt := range opts {
		opt(&vc)
	}
	return vc
}

func NewLocalImage(path string, opts ...VisualOption) VisualContent {
	vc := VisualContent{
		Type:     "image_url",
		ImageURL: path,
		Detail:   DetailAuto,
	}
	for _, opt := range opts {
		opt(&vc)
	}
	return vc
}

func NewVideoURL(url string, opts ...VisualOption) VisualContent {
	vc := VisualContent{
		Type:     "video_url",
		VideoURL: url,
	}
	for _, opt := range opts {
		opt(&vc)
	}
	return vc
}

func NewTextBlock(text string, opts ...VisualOption) VisualContent {
	vc := VisualContent{
		Type: "text",
		Text: text,
	}
	for _, opt := range opts {
		opt(&vc)
	}
	return vc
}

// NewImageList 将多个图片路径转换为 VisualContent 列表。
func NewImageList(paths []string, opts ...VisualOption) []VisualContent {
	visuals := make([]VisualContent, 0, len(paths))
	for _, p := range paths {
		visuals = append(visuals, NewImageURL(p, opts...))
	}
	return visuals
}
