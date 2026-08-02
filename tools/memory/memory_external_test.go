package memory_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxtools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/memory"
)

func TestExternalConsumerRunsSearchAndGet(t *testing.T) {
	var searchRequest memory.SearchRequest
	var getRequest memory.GetRequest
	backend := memory.BackendFuncs{
		SearchFunc: func(ctx context.Context, request memory.SearchRequest) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			searchRequest = request
			return `{"query":"deploy","sources_requested":["memory","sessions"],"hits":[]}`, nil
		},
		GetFunc: func(ctx context.Context, request memory.GetRequest) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			getRequest = request
			return `{"path":"MEMORY.md","line_start":1,"line_end":1,"text":"ready"}`, nil
		},
	}
	registry := agentxtools.NewRegistry()
	memory.Register(registry, memory.Options{Backend: backend, MaxSearchResults: 3, MaxReadLines: 5})

	searchResult, err := registry.Execute(context.Background(), toolcontract.Call{
		Name:      memory.SearchName,
		Arguments: `{"query":"deploy","limit":9,"sources":["memory.md","session"],"session_id":"s1","statuses":["done"],"candidate_limit":10}`,
	})
	if err != nil || searchResult == "" {
		t.Fatalf("search result=%q err=%v", searchResult, err)
	}
	if searchRequest.Limit != 3 || !searchRequest.IncludeMemory || !searchRequest.IncludeSessions || searchRequest.IncludeStructured {
		t.Fatalf("search request=%#v", searchRequest)
	}
	if !reflect.DeepEqual(searchRequest.Sources, []string{"memory", "sessions"}) || searchRequest.Session.SessionID != "s1" || !reflect.DeepEqual(searchRequest.Session.Statuses, []string{"done"}) {
		t.Fatalf("normalized search request=%#v", searchRequest)
	}

	getResult, err := registry.Execute(context.Background(), toolcontract.Call{
		Name: memory.GetName, Arguments: `{"path":"MEMORY.md","from":1,"lines":99}`,
	})
	if err != nil || getResult == "" || getRequest.Lines != 5 {
		t.Fatalf("get result=%q request=%#v err=%v", getResult, getRequest, err)
	}
}

func TestErrorsAndCancellationRemainTyped(t *testing.T) {
	backend := memory.BackendFuncs{
		SearchFunc: func(ctx context.Context, _ memory.SearchRequest) (string, error) { return "", ctx.Err() },
		GetFunc:    func(ctx context.Context, _ memory.GetRequest) (string, error) { return "", ctx.Err() },
	}
	registry := agentxtools.NewRegistry()
	memory.Register(registry, memory.Options{Backend: backend})
	_, err := registry.Execute(context.Background(), toolcontract.Call{Name: memory.SearchName, Arguments: `{}`})
	argumentError, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok || argumentError.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("argument error=%T %v", err, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = registry.Execute(ctx, toolcontract.Call{Name: memory.SearchName, Arguments: `{"query":"x"}`})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestDefinitionsRemainStable(t *testing.T) {
	registry := agentxtools.NewRegistry()
	memory.Register(registry, memory.Options{Backend: memory.BackendFuncs{
		SearchFunc: func(context.Context, memory.SearchRequest) (string, error) { return `{}`, nil },
		GetFunc:    func(context.Context, memory.GetRequest) (string, error) { return `{}`, nil },
	}})
	definitions := registry.Definitions()
	if len(definitions) != 2 || definitions[0].Function.Name != memory.GetName || definitions[1].Function.Name != memory.SearchName {
		t.Fatalf("definitions=%#v", definitions)
	}
	if definitions[1].Function.Parameters["additionalProperties"] != false || len(definitions[1].Function.OutputSchema) == 0 {
		t.Fatalf("search schema=%#v", definitions[1].Function)
	}
}
