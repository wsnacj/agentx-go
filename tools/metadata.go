package tools

// Tool source values describe where a catalog entry is owned. They are
// descriptive metadata only; they do not grant authorization or select a
// backend.
const (
	ToolSourceUnknown   = "unknown"
	ToolSourceBuiltin   = "builtin"
	ToolSourceExtension = "extension"
	ToolSourceProject   = "project"
	ToolSourceCustom    = "custom"
)

// ToolMetadata describes a tool catalog entry without owning authorization,
// approval, sandbox, credential, or backend policy.
//
// Pointer booleans preserve the distinction between an explicit false value
// and an unspecified hint.
type ToolMetadata struct {
	Plugin          string
	Groups          []string
	Type            string
	Source          string
	Capabilities    []string
	AuditTags       []string
	RiskProfile     string
	ReadOnly        *bool
	ConcurrencySafe *bool
	Destructive     *bool
}
