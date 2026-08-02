package browserd

import "strings"

const (
	// NodeBackendSourceExternalProxy identifies a caller-owned remote browser endpoint.
	NodeBackendSourceExternalProxy = "external_proxy"
	// NodeBackendSourceManagedBrowser identifies a browserd process managed by this host.
	NodeBackendSourceManagedBrowser = "managed_browserd"
)

// Plan is the explicit process and storage contract for a managed browserd.
// It contains no credential discovery, product defaults, or implicit workspace policy.
type Plan struct {
	Enabled           bool
	Command           string
	Args              []string
	Host              string
	Port              int
	Endpoint          string
	Token             string
	StateRoot         string
	ProfilesRoot      string
	ArtifactsRoot     string
	LogsRoot          string
	AttachCDPEndpoint string
}

// NodeBackendPlan contains the minimum route facts needed to derive browserd
// capabilities without depending on an application configuration package.
type NodeBackendPlan struct {
	Source  string
	Managed Plan
}

// UsesManagedBrowserd reports whether the route selects the managed browserd host.
func (p NodeBackendPlan) UsesManagedBrowserd() bool {
	return strings.TrimSpace(p.Source) == NodeBackendSourceManagedBrowser
}
