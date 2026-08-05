package connector

// Protocol identifies the external capability protocol.
type Protocol string

const (
	ProtocolMCP Protocol = "mcp"
)

// Transport identifies the connection mechanism without carrying endpoint,
// process, credential, or tenant state.
type Transport string

const (
	TransportStdio          Transport = "stdio"
	TransportStreamableHTTP Transport = "streamable_http"
)

// Spec is a portable, display-safe connector declaration. Concrete command,
// endpoint, credential, retry, proxy and lifecycle configuration belong to
// the Host.
type Spec struct {
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Version     string    `json:"version,omitempty"`
	Protocol    Protocol  `json:"protocol"`
	Transport   Transport `json:"transport"`
}
