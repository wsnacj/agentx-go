package artifact

import "context"

type Record struct {
	ArtifactID   string   `json:"artifact_id"`
	RunID        string   `json:"run_id,omitempty"`
	NodeExecID   string   `json:"node_exec_id,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	ToolName     string   `json:"tool_name,omitempty"`
	Producer     string   `json:"producer,omitempty"`
	Source       string   `json:"source,omitempty"`
	Kind         string   `json:"kind,omitempty"`
	Role         string   `json:"role,omitempty"`
	Path         string   `json:"path,omitempty"`
	StorageRef   string   `json:"storage_ref,omitempty"`
	URL          string   `json:"url,omitempty"`
	Digest       string   `json:"digest,omitempty"`
	MIMEType     string   `json:"mime_type,omitempty"`
	Format       string   `json:"format,omitempty"`
	Bytes        int64    `json:"bytes,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	MetadataJSON string   `json:"metadata_json,omitempty"`
	CreatedAt    int64    `json:"created_at"`
}

type Link struct {
	SourceArtifactID string `json:"source_artifact_id"`
	TargetArtifactID string `json:"target_artifact_id"`
	Relation         string `json:"relation"`
	MetadataJSON     string `json:"metadata_json,omitempty"`
	CreatedAt        int64  `json:"created_at"`
}

type LinkFilter struct {
	ArtifactID string `json:"artifact_id,omitempty"`
	Relation   string `json:"relation,omitempty"`
	Direction  string `json:"direction,omitempty"`
}

type BlobPutInput struct {
	Namespace string `json:"namespace,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Extension string `json:"extension,omitempty"`
	Data      []byte `json:"-"`
}

type BlobRef struct {
	StorageRef string `json:"storage_ref,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Bytes      int64  `json:"bytes,omitempty"`
}

type BlobStore interface {
	Put(ctx context.Context, input BlobPutInput) (BlobRef, error)
}

type AuthoringRegistry interface {
	Register(ctx context.Context, record Record) error
	Link(ctx context.Context, link Link) error
}

type QueryRegistry interface {
	ListByRun(ctx context.Context, runID string) ([]Record, error)
	ListBySession(ctx context.Context, sessionID string) ([]Record, error)
}

type Registry interface {
	AuthoringRegistry
	QueryRegistry
	ListLinks(ctx context.Context, filter LinkFilter) ([]Link, error)
}
