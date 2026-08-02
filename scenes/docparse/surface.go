package docparse

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
)

// PackDefinition returns a caller-owned copy of the Docparse Pack definition.
func PackDefinition() agentxpack.Definition { return Definition() }

// ToolNames returns the semantic tool identities exposed by this kit.
func ToolNames() []string { return hostkit.ToolNames() }

// SkillNames returns the embedded skill identities exposed by this kit.
func SkillNames() []string { return []string{"document-operations"} }

// RegisterPacksIntoRegistry registers the Docparse Pack into a canonical registry.
func RegisterPacksIntoRegistry(reg agentxpack.Registry) error { return RegisterInto(reg) }
