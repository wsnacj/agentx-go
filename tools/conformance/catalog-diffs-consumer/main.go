package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/diffs"
)

type result struct {
	Registered []string `json:"registered"`
	Repaired   bool     `json:"repaired"`
	Additions  int      `json:"additions"`
	Deletions  int      `json:"deletions"`
	Path       string   `json:"path"`
}

func run(ctx context.Context) (result, error) {
	registry := tools.NewRegistry()
	diffs.Register(registry)
	definitions := registry.Definitions()
	names := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		names = append(names, definition.Function.Name)
	}
	resolution := tools.RepairToolName("Diffs-Tool", names)
	output, err := registry.Execute(ctx, toolcontract.Call{
		Name: "Diffs-Tool", Arguments: `{"before":"alpha\nbeta\n","after":"alpha\ngamma\n","path":"sample.txt"}`,
	})
	if err != nil {
		return result{}, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return result{}, err
	}
	return result{
		Registered: names, Repaired: resolution.Repaired && resolution.Name == "diffs",
		Additions: int(payload["additions"].(float64)), Deletions: int(payload["deletions"].(float64)),
		Path: payload["path"].(string),
	}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
