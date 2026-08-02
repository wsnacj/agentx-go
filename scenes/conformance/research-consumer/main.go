package main

import (
	"context"
	"fmt"
	"os"

	companyresearch "github.com/wsnacj/agentx-go/scenes/companyresearch"
	companyhostkit "github.com/wsnacj/agentx-go/scenes/companyresearch/hostkit"
	publicnews "github.com/wsnacj/agentx-go/scenes/publicnews"
	newshostkit "github.com/wsnacj/agentx-go/scenes/publicnews/hostkit"
)

func run() (string, error) {
	ctx := context.Background()
	news, err := newshostkit.BuildLatestNewsLookupPayload(ctx, newshostkit.LatestNewsLookupConfig{
		Handlers: newshostkit.LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent publicnews.LatestNewsLookupIntent) (publicnews.LatestNewsSourcesPayload, error) {
				return publicnews.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					Intent:        intent,
					PrimarySource: publicnews.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-08-02T09:00:00Z",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "发布时间: 2026-08-02 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []publicnews.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-08-02T09:05:00Z",
						KeyUpdate:   "第二来源确认示例事件出现最新进展。",
						Text:        "发布时间: 2026-08-02 09:05 UTC\n第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
		"topic":        "示例事件",
	})
	if err != nil {
		return "", err
	}

	company, err := companyhostkit.BuildCompanyResearchLookupPayload(ctx, companyhostkit.CompanyResearchConfig{
		Handlers: companyhostkit.CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"adapter_status": "ok",
					"source_url":     "https://example.com/filing",
				}, nil
			},
		},
	}, map[string]any{
		"user_message":         "research Example Corp financials",
		"entity_name":          "Example Corp",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		return "", err
	}

	if news.Tool != publicnews.ToolLatestNewsLookup || company.Tool != companyresearch.ToolCompanyResearchLookup {
		return "", fmt.Errorf("canonical research identity mismatch")
	}
	return fmt.Sprintf("agentx-research-ok:%s:%s:%s:%s:%t:%t",
		publicnews.PackID,
		publicnews.DefaultWorkflow,
		companyresearch.PackID,
		companyresearch.DefaultWorkflow,
		news.Passed,
		company.AnswerReadiness.AnswerReady,
	), nil
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
