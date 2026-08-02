package astock

import (
	"embed"
	"io/fs"

	"github.com/wsnacj/agentx-go/runtime/assetfs"
)

//go:embed skills tools/*.tool.json
var embeddedAssets embed.FS

var assets = assetfs.MustNew("agentx/astock", embeddedAssets)

// Assets returns the immutable A-share skill/tool asset provider.
func Assets() assetfs.Provider { return assets }

// ExtensionFS returns the immutable filesystem view used by extension hosts.
func ExtensionFS() fs.FS { return assets.FS() }
