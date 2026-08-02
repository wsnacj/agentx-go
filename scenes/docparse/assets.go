package docparse

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// AssetRoots lists filesystem locations for reusable docparse module assets.
type AssetRoots struct {
	DomainRoot string
	SkillsRoot string
	ToolsRoot  string
	PacksRoot  string
	DocsRoot   string
}

//go:embed skills tools/*.tool.json
var embeddedExtensionFS embed.FS

// ExtensionFS returns the immutable skills and declarative tool manifests
// compiled into the docparse package.
func ExtensionFS() fs.FS {
	return embeddedExtensionFS
}

// LocateAssets is retained for source compatibility only.
//
// Deprecated: implicit source-checkout discovery is not distribution safe.
// Use ExtensionFS or LocateAssetsAt.
func LocateAssets() (AssetRoots, error) {
	return AssetRoots{}, fmt.Errorf("locate agentx_docparse assets: implicit source-checkout discovery is disabled; use ExtensionFS or LocateAssetsAt")
}

// LocateAssetsAt returns filesystem roots below an explicit host-owned root.
func LocateAssetsAt(root string) (AssetRoots, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return AssetRoots{}, fmt.Errorf("locate agentx_docparse assets: explicit root is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return AssetRoots{}, fmt.Errorf("locate agentx_docparse assets: resolve explicit root: %w", err)
	}
	root = filepath.Clean(absolute)
	return AssetRoots{
		DomainRoot: root,
		SkillsRoot: filepath.Join(root, "skills"),
		ToolsRoot:  filepath.Join(root, "tools"),
		PacksRoot:  filepath.Join(root, "packs"),
		DocsRoot:   filepath.Join(root, "docs"),
	}, nil
}

// DomainRoot no longer derives a package source directory.
//
// Deprecated: use ExtensionFS or LocateAssetsAt.
func DomainRoot() (string, error) {
	return "", fmt.Errorf("locate agentx_docparse domain root: implicit source-checkout discovery is disabled; use ExtensionFS or LocateAssetsAt")
}

// ExtensionRoot returns the root that exposes reusable docparse skills and tools.
func (r AssetRoots) ExtensionRoot() string {
	return r.DomainRoot
}

// SkillPath returns the SKILL.md path for a docparse skill name.
func (r AssetRoots) SkillPath(name string) string {
	return filepath.Join(r.SkillsRoot, name, "SKILL.md")
}

// ToolManifestPath returns the .tool.json path for a docparse tool name.
func (r AssetRoots) ToolManifestPath(name string) string {
	return filepath.Join(r.ToolsRoot, name+".tool.json")
}
