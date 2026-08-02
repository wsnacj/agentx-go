package pipeline

import (
	"fmt"
	"os"
	"strings"
)

func readPromptsFromTxt(filePath string) (map[string]string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	prompts := make(map[string]string)
	for _, section := range strings.Split(string(contentBytes), "### ") {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		lines := strings.SplitN(section, "\n", 2)
		if len(lines) > 1 {
			prompts[strings.TrimSpace(lines[0])] = strings.TrimSpace(lines[1])
			continue
		}
		prompts[fmt.Sprintf("PROMPT_%d", len(prompts)+1)] = ""
	}
	return prompts, nil
}
