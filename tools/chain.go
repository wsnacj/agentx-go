package tools

import (
	"context"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

// ChainExecutor dispatches calls across multiple executors.
// It falls back only when the current executor does not know the tool.
type ChainExecutor struct {
	Executors []toolcontract.Executor
}

func (c ChainExecutor) Execute(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
	name := strings.TrimSpace(call.Name)
	if name == "" {
		return "", fmt.Errorf("llmx: tool name is required")
	}

	executors := normalizeExecutors(c.Executors)
	if len(executors) == 0 {
		return "", fmt.Errorf("llmx: tool %s not registered", name)
	}

	for _, exec := range executors {
		knownTool := false
		if defs, ok := executorDefinitions(exec); ok {
			if len(defs) == 0 || !definitionsContainTool(defs, name) {
				continue
			}
			knownTool = true
		}
		out, err := exec.Execute(ctx, call)
		if err == nil {
			return out, nil
		}
		if knownTool {
			return "", err
		}
		if isToolNotRegistered(err, name) {
			continue
		}
		return "", err
	}
	return "", fmt.Errorf("llmx: tool %s not registered", name)
}

func (c ChainExecutor) Definitions() []toolcontract.Definition {
	executors := normalizeExecutors(c.Executors)
	if len(executors) == 0 {
		return nil
	}
	defs := make([]toolcontract.Definition, 0)
	seen := map[string]bool{}
	for _, exec := range executors {
		defProvider, ok := exec.(interface {
			Definitions() []toolcontract.Definition
		})
		if !ok {
			continue
		}
		for _, def := range defProvider.Definitions() {
			name := strings.ToLower(strings.TrimSpace(def.Function.Name))
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			defs = append(defs, def)
		}
	}
	SortByName(defs)
	return defs
}

func normalizeExecutors(execs []toolcontract.Executor) []toolcontract.Executor {
	if len(execs) == 0 {
		return nil
	}
	out := make([]toolcontract.Executor, 0, len(execs))
	for _, exec := range execs {
		if exec == nil {
			continue
		}
		out = append(out, exec)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func executorDefinitions(exec toolcontract.Executor) ([]toolcontract.Definition, bool) {
	provider, ok := exec.(interface {
		Definitions() []toolcontract.Definition
	})
	if !ok || provider == nil {
		return nil, false
	}
	return provider.Definitions(), true
}

func definitionsContainTool(defs []toolcontract.Definition, name string) bool {
	normalized := NormalizeToolName(name)
	for _, def := range defs {
		defName := strings.TrimSpace(def.Function.Name)
		if defName == "" {
			continue
		}
		if strings.EqualFold(defName, name) || NormalizeToolName(defName) == normalized {
			return true
		}
	}
	return false
}

func isToolNotRegistered(err error, name string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if msg == "" {
		return false
	}
	if strings.Contains(msg, "not registered") {
		return true
	}
	return strings.Contains(msg, strings.ToLower(strings.TrimSpace(name))) && strings.Contains(msg, "not found")
}
