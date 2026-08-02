// Package memory provides portable memory search/get tool coordination.
package memory

import (
	"context"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

const (
	// SearchName is the catalog name of unified memory search.
	SearchName = "memory_search"
	// GetName is the catalog name of memory snippet reads.
	GetName = "memory_get"
)

const (
	defaultMaxSearchResults = 8
	defaultMaxReadLines     = 40
)

// Options configures model-facing memory tool coordination.
type Options struct {
	Backend             Backend
	MaxSearchResults    int
	MaxReadLines        int
	StructuredByDefault bool
}

// Backend is the narrow Host-owned memory data plane. Implementations own
// concrete stores, ranking models, visibility rules and source availability.
type Backend interface {
	Search(context.Context, SearchRequest) (string, error)
	Get(context.Context, GetRequest) (string, error)
}

// BackendFuncs adapts private Host functions to Backend.
type BackendFuncs struct {
	SearchFunc func(context.Context, SearchRequest) (string, error)
	GetFunc    func(context.Context, GetRequest) (string, error)
}

// Search delegates to SearchFunc and fails closed when unavailable.
func (b BackendFuncs) Search(ctx context.Context, request SearchRequest) (string, error) {
	if b.SearchFunc == nil {
		return "", fmt.Errorf("%s: search backend is unavailable", SearchName)
	}
	return b.SearchFunc(ctx, request)
}

// Get delegates to GetFunc and fails closed when unavailable.
func (b BackendFuncs) Get(ctx context.Context, request GetRequest) (string, error) {
	if b.GetFunc == nil {
		return "", fmt.Errorf("%s: get backend is unavailable", GetName)
	}
	return b.GetFunc(ctx, request)
}

// SearchRequest is the normalized unified recall request passed to a Host.
type SearchRequest struct {
	Query             string
	Limit             int
	Sources           []string
	IncludeMemory     bool
	IncludeStructured bool
	IncludeSessions   bool
	Session           SessionSearch
}

// SessionSearch contains optional persisted-session recall filters. The Host
// remains responsible for visibility, lineage, hydration and ranking policy.
type SessionSearch struct {
	SessionID             string
	Status                string
	Statuses              []string
	Model                 string
	Models                []string
	Tag                   string
	Tags                  []string
	Mode                  string
	CandidateLimit        int
	RerankLimit           int
	IncludeClusters       bool
	ClusterLimit          int
	HistoryLines          int
	ExcludeCurrentLineage *bool
}

// GetRequest is the normalized bounded snippet request.
type GetRequest struct {
	Path  string
	From  int
	Lines int
}

// Register adds memory_search and memory_get when a Backend is available.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.Backend == nil {
		return
	}
	reg.Register(SearchDefinition(), NewSearchHandler(opts))
	reg.Register(GetDefinition(), NewGetHandler(opts))
}

// NewSearchHandler constructs the unified memory search implementation.
func NewSearchHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		if _, exists := params["root"]; exists {
			return "", agentxtoolerrors.NewInvalidToolArgumentError(SearchName, []string{"root"}, SearchName+": workspace, memory, and artifact roots are host-owned and cannot be overridden by tool arguments")
		}
		query := readString(params, "query")
		if query == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(SearchName, []string{"query"}, SearchName+": query is required")
		}
		maxSearch := opts.MaxSearchResults
		if maxSearch <= 0 {
			maxSearch = defaultMaxSearchResults
		}
		limit := readInt(params, "limit")
		if limit <= 0 {
			limit = maxSearch
		}
		if limit > maxSearch {
			limit = maxSearch
		}
		includeMemory, includeStructured, includeSessions, sources := resolveSources(params, opts.StructuredByDefault)
		if !includeMemory && !includeStructured && !includeSessions {
			return "", agentxtoolerrors.NewInvalidToolArgumentError(
				SearchName,
				[]string{"sources", "include_memory", "include_structured", "include_sessions"},
				SearchName+": at least one source must be selected",
			)
		}
		request := SearchRequest{
			Query: query, Limit: limit, Sources: sources,
			IncludeMemory: includeMemory, IncludeStructured: includeStructured, IncludeSessions: includeSessions,
			Session: SessionSearch{
				SessionID: readString(params, "session_id"), Status: readString(params, "status"),
				Statuses: readStringSlice(params, "statuses"), Model: readString(params, "model"),
				Models: readStringSlice(params, "models"), Tag: readString(params, "tag"),
				Tags: readStringSlice(params, "tags"), Mode: readString(params, "mode"),
				CandidateLimit: readInt(params, "candidate_limit"), RerankLimit: readInt(params, "rerank_limit"),
				IncludeClusters: readBool(params, "include_clusters"), ClusterLimit: readInt(params, "cluster_limit"),
				HistoryLines: readInt(params, "history_lines"), ExcludeCurrentLineage: readOptionalBool(params, "exclude_current_lineage"),
			},
		}
		return opts.Backend.Search(ctx, request)
	}
}

// NewGetHandler constructs the bounded memory snippet implementation.
func NewGetHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		if _, exists := params["root"]; exists {
			return "", agentxtoolerrors.NewInvalidToolArgumentError(GetName, []string{"root"}, GetName+": workspace, memory, and artifact roots are host-owned and cannot be overridden by tool arguments")
		}
		path := readString(params, "path")
		if path == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(GetName, []string{"path"}, GetName+": path is required")
		}
		maxLines := opts.MaxReadLines
		if maxLines <= 0 {
			maxLines = defaultMaxReadLines
		}
		lines := readInt(params, "lines")
		if lines <= 0 {
			lines = 20
		}
		if lines > maxLines {
			lines = maxLines
		}
		return opts.Backend.Get(ctx, GetRequest{Path: path, From: readInt(params, "from"), Lines: lines})
	}
}

func resolveSources(params map[string]any, structuredDefault bool) (bool, bool, bool, []string) {
	includeMemory := true
	includeStructured := structuredDefault
	includeSessions := readBool(params, "include_sessions")
	raw, exists := params["sources"]
	if exists && raw != nil {
		includeMemory, includeStructured, includeSessions = false, false, false
		for _, source := range readSourceValues(raw) {
			switch normalizeSource(source) {
			case "memory":
				includeMemory = true
			case "structured":
				includeStructured = true
			case "sessions":
				includeSessions = true
			}
		}
		if readBool(params, "include_sessions") {
			includeSessions = true
		}
	}
	sources := make([]string, 0, 3)
	if includeMemory {
		sources = append(sources, "memory")
	}
	if includeStructured {
		sources = append(sources, "structured")
	}
	if includeSessions {
		sources = append(sources, "sessions")
	}
	return includeMemory, includeStructured, includeSessions, sources
}

func readSourceValues(raw any) []string {
	switch typed := raw.(type) {
	case string:
		return []string{typed}
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeSource(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	switch normalized {
	case "memory", "files", "durable", "memory.md":
		return "memory"
	case "structured", "memory_records", "records", "canonical":
		return "structured"
	case "sessions", "session", "recall":
		return "sessions"
	}
	if strings.HasPrefix(normalized, "memory/") {
		return "memory"
	}
	return normalized
}
