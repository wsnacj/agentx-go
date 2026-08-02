package adapters

import "strings"

func uniqueAppend(existing []string, values ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(existing)+len(values))
	for _, value := range append(existing, values...) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
