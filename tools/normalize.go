package tools

import (
	"sort"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

var toolNameAliases = map[string]string{
	"bash":        "exec",
	"apply-patch": "apply_patch",
}

// NormalizeToolName returns the stable lower-case catalog name and applies the
// two legacy aliases already accepted by AgentX hosts.
func NormalizeToolName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return ""
	}
	if alias, ok := toolNameAliases[normalized]; ok {
		return alias
	}
	return normalized
}

// SortByName sorts definitions by their normalized function name.
func SortByName(definitions []toolcontract.Definition) {
	sort.SliceStable(definitions, func(i, j int) bool {
		return NormalizeToolName(definitions[i].Function.Name) < NormalizeToolName(definitions[j].Function.Name)
	})
}
