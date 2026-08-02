package provider

import "github.com/wsnacj/agentx-go/document/ocr/config"

// DefaultFactories returns built-in provider factories keyed by ProviderConfig.Kind.
func DefaultFactories() map[string]Factory {
	return map[string]Factory{
		"textin":     NewTextInProvider,
		"volcengine": NewVolcEngineProvider,
		"baidu":      NewBaiduProvider,
	}
}

// NewTextInProvider constructs a TextIn OCR provider.
func NewTextInProvider(cfg config.ProviderConfig) (Provider, error) {
	return newTextIn(cfg)
}
