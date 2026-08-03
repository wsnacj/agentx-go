package main

import (
	"fmt"

	mediaartifact "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

func run() (string, error) {
	refs, err := mediaartifact.RefsFromValue(map[string]any{
		"tool":  "pdf_analyze",
		"title": "Portable report",
		"rendered_pages": []any{
			map[string]any{
				"path": "page-1.png",
				"kind": "rendered_page",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(refs) != 1 {
		return "", fmt.Errorf("artifact ref count = %d, want 1", len(refs))
	}
	ref := refs[0]
	if ref.Raw != "page-1.png" || ref.ArtifactSource != "pdf" || ref.ArtifactKind != "rendered_page" || ref.ModeHint != "document" {
		return "", fmt.Errorf("unexpected artifact ref: %#v", ref)
	}
	return fmt.Sprintf(
		"agentx-media-artifact-ok:%s:%s:%s",
		ref.ArtifactSource,
		ref.ArtifactKind,
		ref.Raw,
	), nil
}

func main() {
	output, err := run()
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
