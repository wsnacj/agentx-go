package expert

import capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"

const SchemaVersionV1 = "v1"

// Requirement declares one capability an Expert expects the Host to make
// available. It is not an authorization grant or execution route.
type Requirement struct {
	Kind     capabilitycatalog.Kind `json:"kind"`
	ID       string                 `json:"id"`
	Optional bool                   `json:"optional,omitempty"`
}

// Spec is one portable role asset. Instructions are untrusted content and
// must be reviewed and assembled by the Host before use.
type Spec struct {
	ID            string        `json:"id"`
	SchemaVersion string        `json:"schema_version"`
	Name          string        `json:"name"`
	Version       string        `json:"version,omitempty"`
	Description   string        `json:"description,omitempty"`
	Instructions  string        `json:"instructions"`
	Requirements  []Requirement `json:"requirements,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
}
