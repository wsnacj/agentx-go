package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/section"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---- 落盘 ----

func saveArtifacts(docPath, specPath string, spec *configs.DocSpec, pages []string, nodes []*section.Node, res *types.DocumentResult, outDir string, policy ArtifactPolicy) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	artifacts := []string{}
	addArtifact := func(name string) {
		artifacts = append(artifacts, name)
	}

	if err := writeJSON(filepath.Join(outDir, "result.json"), res); err != nil {
		return err
	}
	addArtifact("result.json")
	if res.Diagnostics != nil {
		if err := writeJSON(filepath.Join(outDir, "diagnostics.json"), res.Diagnostics); err != nil {
			return err
		}
		addArtifact("diagnostics.json")
	}

	if policy == ArtifactPolicyFull {
		if err := saveFullArtifacts(docPath, pages, nodes, res, outDir, addArtifact); err != nil {
			return err
		}
	}

	manifest := buildArtifactManifest(docPath, specPath, spec, pages, res, outDir, policy, artifacts)
	if err := writeJSON(filepath.Join(outDir, "manifest.json"), manifest); err != nil {
		return err
	}
	return nil
}

func saveFullArtifacts(docPath string, pages []string, nodes []*section.Node, res *types.DocumentResult, outDir string, addArtifact func(string)) error {
	// 拷贝PDF
	if err := copyFile(docPath, filepath.Join(outDir, "input.pdf")); err != nil {
		// The input copy is optional. Remaining artifacts are still useful when
		// the host cannot reopen the source document.
	} else {
		addArtifact("input.pdf")
	}
	// pages
	if err := writeJSON(filepath.Join(outDir, "pages.json"), pages); err != nil {
		return err
	}
	addArtifact("pages.json")
	if err := writeText(filepath.Join(outDir, "pages.txt"), strings.Join(pages, "\n\n---PAGE---\n\n")); err != nil {
		return err
	}
	addArtifact("pages.txt")
	// nodes 摘要
	type NodeView struct {
		Name      string   `json:"name"`
		Matched   []string `json:"matched,omitempty"`
		PageCount int      `json:"page_count"`
	}
	var views []NodeView
	for _, n := range nodes {
		views = append(views, NodeView{Name: n.Name, Matched: n.Matched, PageCount: len(n.Pages)})
	}
	if err := writeJSON(filepath.Join(outDir, "sections_nodes.json"), views); err != nil {
		return err
	}
	addArtifact("sections_nodes.json")
	// 每章
	for idx, key := range res.ChapterOrder {
		r := res.Chapters[key]
		if r == nil {
			continue
		}
		stem := fmt.Sprintf("%02d_%s", idx+1, safeName(key))
		parsedName := stem + "_parsed.json"
		if err := writeJSON(filepath.Join(outDir, parsedName), r); err != nil {
			return err
		}
		addArtifact(parsedName)
		if r.RawLLM != "" {
			rawName := stem + "_raw.json"
			if err := writeText(filepath.Join(outDir, rawName), r.RawLLM); err != nil {
				return err
			}
			addArtifact(rawName)
		}
		if r.Prompt != "" {
			promptName := stem + "_prompt.txt"
			if err := writeText(filepath.Join(outDir, promptName), r.Prompt); err != nil {
				return err
			}
			addArtifact(promptName)
		}
	}
	// 保存校验
	if len(res.Validations) > 0 {
		if err := writeJSON(filepath.Join(outDir, "validations.json"), res.Validations); err != nil {
			return err
		}
		addArtifact("validations.json")
	}
	return nil
}

func buildArtifactManifest(docPath, specPath string, spec *configs.DocSpec, pages []string, res *types.DocumentResult, outDir string, policy ArtifactPolicy, artifacts []string) map[string]any {
	return map[string]any{
		"pdf_path":        docPath,
		"spec_path":       specPath,
		"segmentation":    filepath.Join(spec.ConfigDir, "sections.yaml"),
		"pages":           len(pages),
		"chapters":        res.ChapterOrder,
		"fingerprint":     res.Fingerprint,
		"diagnostics":     res.Diagnostics,
		"artifact_policy": string(policy),
		"artifacts":       artifacts,
		"saved_at":        time.Now().Format(time.RFC3339),
		"output_dir":      outDir,
	}
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func writeText(path, s string) error { return os.WriteFile(path, []byte(s), 0644) }

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "untitled"
	}
	repl := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			fallthrough
		case r >= 'A' && r <= 'Z':
			fallthrough
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}
	return strings.Map(repl, s)
}
