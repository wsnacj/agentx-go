package artifact

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryRegistry struct {
	mu        sync.RWMutex
	byID      map[string]Record
	byRun     map[string][]string
	bySession map[string][]string
	links     []Link
}

func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{
		byID:      map[string]Record{},
		byRun:     map[string][]string{},
		bySession: map[string][]string{},
	}
}

func (r *MemoryRegistry) Register(_ context.Context, record Record) error {
	if r == nil {
		return nil
	}
	normalized := normalizeRecord(record)
	r.mu.Lock()
	defer r.mu.Unlock()
	_, existed := r.byID[normalized.ArtifactID]
	if existing, ok := r.byID[normalized.ArtifactID]; ok {
		normalized = mergeRecord(existing, normalized)
	}
	r.byID[normalized.ArtifactID] = normalized
	if !existed && normalized.RunID != "" {
		r.byRun[normalized.RunID] = append(r.byRun[normalized.RunID], normalized.ArtifactID)
	}
	if !existed && normalized.SessionID != "" {
		r.bySession[normalized.SessionID] = append(r.bySession[normalized.SessionID], normalized.ArtifactID)
	}
	return nil
}

func (r *MemoryRegistry) Link(_ context.Context, link Link) error {
	if r == nil {
		return nil
	}
	normalized := normalizeLink(link)
	if normalized.SourceArtifactID == "" || normalized.TargetArtifactID == "" || normalized.Relation == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for idx, existing := range r.links {
		if existing.SourceArtifactID == normalized.SourceArtifactID &&
			existing.TargetArtifactID == normalized.TargetArtifactID &&
			existing.Relation == normalized.Relation {
			r.links[idx] = mergeLink(existing, normalized)
			return nil
		}
	}
	r.links = append(r.links, normalized)
	return nil
}

func (r *MemoryRegistry) ListByRun(_ context.Context, runID string) ([]Record, error) {
	if r == nil {
		return nil, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recordsForIDs(r.byRun[strings.TrimSpace(runID)]), nil
}

func (r *MemoryRegistry) ListBySession(_ context.Context, sessionID string) ([]Record, error) {
	if r == nil {
		return nil, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recordsForIDs(r.bySession[strings.TrimSpace(sessionID)]), nil
}

func (r *MemoryRegistry) ListLinks(_ context.Context, filter LinkFilter) ([]Link, error) {
	if r == nil {
		return nil, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	artifactID := strings.TrimSpace(filter.ArtifactID)
	relation := strings.TrimSpace(filter.Relation)
	direction := strings.ToLower(strings.TrimSpace(filter.Direction))
	out := make([]Link, 0, len(r.links))
	for _, link := range r.links {
		if relation != "" && link.Relation != relation {
			continue
		}
		if artifactID != "" {
			switch direction {
			case "inbound":
				if link.TargetArtifactID != artifactID {
					continue
				}
			case "outbound":
				if link.SourceArtifactID != artifactID {
					continue
				}
			default:
				if link.SourceArtifactID != artifactID && link.TargetArtifactID != artifactID {
					continue
				}
			}
		}
		out = append(out, link)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		if out[i].SourceArtifactID != out[j].SourceArtifactID {
			return out[i].SourceArtifactID < out[j].SourceArtifactID
		}
		if out[i].TargetArtifactID != out[j].TargetArtifactID {
			return out[i].TargetArtifactID < out[j].TargetArtifactID
		}
		return out[i].Relation < out[j].Relation
	})
	return out, nil
}

func (r *MemoryRegistry) recordsForIDs(ids []string) []Record {
	if len(ids) == 0 {
		return nil
	}
	out := make([]Record, 0, len(ids))
	for _, id := range ids {
		record, ok := r.byID[id]
		if !ok {
			continue
		}
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		return out[i].ArtifactID < out[j].ArtifactID
	})
	return out
}

func normalizeRecord(record Record) Record {
	record.ArtifactID = strings.TrimSpace(record.ArtifactID)
	record.RunID = strings.TrimSpace(record.RunID)
	record.NodeExecID = strings.TrimSpace(record.NodeExecID)
	record.SessionID = strings.TrimSpace(record.SessionID)
	record.ToolName = strings.TrimSpace(record.ToolName)
	record.Producer = strings.TrimSpace(record.Producer)
	record.Source = strings.TrimSpace(record.Source)
	record.Kind = strings.TrimSpace(record.Kind)
	record.Role = strings.TrimSpace(record.Role)
	record.Path = strings.TrimSpace(record.Path)
	record.StorageRef = strings.TrimSpace(record.StorageRef)
	record.URL = strings.TrimSpace(record.URL)
	record.Digest = strings.TrimSpace(record.Digest)
	record.MIMEType = strings.TrimSpace(record.MIMEType)
	record.Format = strings.TrimSpace(record.Format)
	record.Summary = strings.TrimSpace(record.Summary)
	record.Labels = normalizeLabels(record.Labels)
	record.MetadataJSON = strings.TrimSpace(record.MetadataJSON)
	return record
}

func normalizeLink(link Link) Link {
	link.SourceArtifactID = strings.TrimSpace(link.SourceArtifactID)
	link.TargetArtifactID = strings.TrimSpace(link.TargetArtifactID)
	link.Relation = strings.TrimSpace(link.Relation)
	link.MetadataJSON = strings.TrimSpace(link.MetadataJSON)
	return link
}

func mergeLink(base Link, incoming Link) Link {
	if base.MetadataJSON == "" {
		base.MetadataJSON = incoming.MetadataJSON
	}
	switch {
	case base.CreatedAt <= 0:
		base.CreatedAt = incoming.CreatedAt
	case incoming.CreatedAt > 0 && incoming.CreatedAt < base.CreatedAt:
		base.CreatedAt = incoming.CreatedAt
	}
	return base
}

func mergeRecord(base Record, incoming Record) Record {
	base.RunID = firstArtifactString(base.RunID, incoming.RunID)
	base.NodeExecID = firstArtifactString(base.NodeExecID, incoming.NodeExecID)
	base.SessionID = firstArtifactString(base.SessionID, incoming.SessionID)
	base.ToolName = firstArtifactString(base.ToolName, incoming.ToolName)
	base.Producer = firstArtifactString(base.Producer, incoming.Producer)
	base.Source = firstArtifactString(base.Source, incoming.Source)
	base.Kind = firstArtifactString(base.Kind, incoming.Kind)
	base.Role = firstArtifactString(base.Role, incoming.Role)
	base.Path = firstArtifactString(base.Path, incoming.Path)
	base.StorageRef = firstArtifactString(base.StorageRef, incoming.StorageRef)
	base.URL = firstArtifactString(base.URL, incoming.URL)
	base.Digest = firstArtifactString(base.Digest, incoming.Digest)
	base.MIMEType = firstArtifactString(base.MIMEType, incoming.MIMEType)
	base.Format = firstArtifactString(base.Format, incoming.Format)
	if incoming.Bytes > base.Bytes {
		base.Bytes = incoming.Bytes
	}
	base.Summary = firstArtifactString(base.Summary, incoming.Summary)
	base.Labels = mergeArtifactLabels(base.Labels, incoming.Labels)
	base.MetadataJSON = firstArtifactString(base.MetadataJSON, incoming.MetadataJSON)
	switch {
	case base.CreatedAt <= 0:
		base.CreatedAt = incoming.CreatedAt
	case incoming.CreatedAt > 0 && incoming.CreatedAt < base.CreatedAt:
		base.CreatedAt = incoming.CreatedAt
	}
	return base
}

func firstArtifactString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func mergeArtifactLabels(base []string, incoming []string) []string {
	if len(base) == 0 && len(incoming) == 0 {
		return nil
	}
	return normalizeLabels(append(append([]string(nil), base...), incoming...))
}

func normalizeLabels(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		value := strings.TrimSpace(raw)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
