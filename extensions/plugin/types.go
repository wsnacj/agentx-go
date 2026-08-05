package plugin

// SchemaVersionV1 is the first portable plugin manifest schema.
const SchemaVersionV1 = "v1"

// TrustBoundary describes provenance claimed by a bundle. It is descriptive
// metadata only and never grants execution authorization.
type TrustBoundary string

const (
	TrustBoundaryWorkspace TrustBoundary = "workspace"
	TrustBoundaryReviewed  TrustBoundary = "reviewed"
	TrustBoundaryTrusted   TrustBoundary = "trusted"
)

// Entrypoints contains bundle-relative asset directories. Paths use forward
// slashes and must remain within one of Manifest.Roots.
type Entrypoints struct {
	Skills   string `json:"skills,omitempty"`
	Tools    string `json:"tools,omitempty"`
	Hooks    string `json:"hooks,omitempty"`
	Commands string `json:"commands,omitempty"`
}

// Dependency is a requested plugin or connector dependency. Version is an
// opaque host-interpreted constraint; this contract does not resolve it.
type Dependency struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version string `json:"version,omitempty"`
}

// PermissionRequest describes a capability requested by a bundle. It is not
// a permission grant and must be evaluated by the Host before activation.
type PermissionRequest struct {
	Capability string `json:"capability"`
	Reason     string `json:"reason,omitempty"`
}

// Manifest is the portable, credential-free description of one filesystem
// capability bundle. Commands, hooks, tool handlers, credentials and policy
// remain owned by the Host and their respective runtime packages.
type Manifest struct {
	Name                 string              `json:"name"`
	SchemaVersion        string              `json:"schema_version"`
	Version              string              `json:"version,omitempty"`
	Description          string              `json:"description,omitempty"`
	TrustBoundary        TrustBoundary       `json:"trust_boundary"`
	Roots                []string            `json:"roots,omitempty"`
	Entrypoints          Entrypoints         `json:"entrypoints,omitempty"`
	Dependencies         []Dependency        `json:"dependencies,omitempty"`
	RequestedPermissions []PermissionRequest `json:"requested_permissions,omitempty"`
}
