package config

import (
	"fmt"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/model"
)

const (
	textInOCREndpoint   = "https://api.textin.com/ai/service/v2/recognize/multipage"
	textInTableEndpoint = "https://api.textin.com/ai/service/v2/recognize/table/multipage"
	textInStampEndpoint = "https://api.textin.com/ai/service/v1/recognize_stamp"
)

// DefaultTextInConfig 根据提供的鉴权信息构造三条 TextIn 管线的默认配置。
func DefaultTextInConfig(appID, secret string) (ServiceConfig, error) {
	if appID == "" || secret == "" {
		return ServiceConfig{}, fmt.Errorf("textin default config requires appID and secret")
	}

	providerAuth := map[string]string{
		"app_id":      appID,
		"secret_code": secret,
	}

	commonCache := CacheConfig{
		Enabled: false,
	}

	commonSplitter := SplitterConfig{
		Kind:        "poppler",
		DPI:         300,
		BatchPages:  0,
		MaxParallel: 4,
	}

	commonWorker := WorkerConfig{MaxConcurrent: 5, QueueSize: 0}
	commonLimits := LimitConfig{MaxPages: 100, MaxFiles: 50}
	commonRetry := RetryConfig{MaxAttempts: 3, Backoff: time.Second, MaxBackoff: 5 * time.Second}

	return ServiceConfig{Pipelines: map[string]PipelineConfig{
		string(model.OperationKindOCR): {
			Provider: ProviderConfig{
				Kind:    "textin",
				BaseURL: textInOCREndpoint,
				Auth:    cloneStringMap(providerAuth),
				Timeout: 30 * time.Second,
			},
			Splitter: commonSplitter,
			Cache:    commonCache,
			Worker:   commonWorker,
			Limits:   commonLimits,
			Retry:    commonRetry,
			Diff:     DiffConfig{Enabled: false, Preview: 3},
		},
		string(model.OperationKindTable): {
			Provider: ProviderConfig{
				Kind:    "textin",
				BaseURL: textInTableEndpoint,
				Auth:    cloneStringMap(providerAuth),
				Timeout: 30 * time.Second,
			},
			Splitter: commonSplitter,
			Cache:    commonCache,
			Worker:   commonWorker,
			Limits:   commonLimits,
			Retry:    commonRetry,
			Diff:     DiffConfig{Enabled: false, Preview: 3},
		},
		string(model.OperationKindStamp): {
			Provider: ProviderConfig{
				Kind:    "textin",
				BaseURL: textInStampEndpoint,
				Auth:    cloneStringMap(providerAuth),
				Timeout: 30 * time.Second,
			},
			Splitter: commonSplitter,
			Cache:    commonCache,
			Worker:   commonWorker,
			Limits:   commonLimits,
			Retry:    commonRetry,
			Diff:     DiffConfig{Enabled: false, Preview: 3},
		},
	}}, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	dup := make(map[string]string, len(in))
	for k, v := range in {
		dup[k] = v
	}
	return dup
}
