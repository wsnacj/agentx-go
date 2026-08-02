package publicnews_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/runtime/workflow"
	publicnews "github.com/wsnacj/agentx-go/scenes/publicnews"
	"github.com/wsnacj/agentx-go/scenes/publicnews/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error { return nil }

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func TestPortablePublicNewsContract(t *testing.T) {
	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if err := publicnews.RegisterInto(registry); err != nil {
		t.Fatalf("RegisterInto(): %v", err)
	}
	spec, err := publicnews.MaterializedDefaultWorkflow(coordinator)
	if err != nil || spec.ID != publicnews.DefaultWorkflow {
		t.Fatalf("MaterializedDefaultWorkflow() = %#v, %v", spec, err)
	}

	payload, err := hostkit.BuildLatestNewsLookupPayload(context.Background(), hostkit.LatestNewsLookupConfig{
		Handlers: hostkit.LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent publicnews.LatestNewsLookupIntent) (publicnews.LatestNewsSourcesPayload, error) {
				return publicnews.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					Intent:        intent,
					PrimarySource: publicnews.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-08-02T09:00:00Z",
						KeyUpdate:   "示例事件出现可核验的新进展。",
						Text:        "示例事件出现可核验的新进展。",
					},
					SupportingSources: []publicnews.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-08-02T09:05:00Z",
						KeyUpdate:   "第二来源独立确认该进展。",
						Text:        "第二来源独立确认该进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "示例事件最新新闻",
		"topic":        "示例事件",
	})
	if err != nil || payload.Tool != publicnews.ToolLatestNewsLookup {
		t.Fatalf("BuildLatestNewsLookupPayload() = %#v, %v", payload, err)
	}
}
