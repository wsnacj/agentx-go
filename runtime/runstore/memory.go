package runstore

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu        sync.RWMutex
	runs      map[string]Run
	events    map[string][]Event
	nodeExecs map[string]map[string]NodeExecution
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		runs:      map[string]Run{},
		events:    map[string][]Event{},
		nodeExecs: map[string]map[string]NodeExecution{},
	}
}

func (s *MemoryStore) CreateRun(_ context.Context, run Run) error {
	if s == nil {
		return nil
	}
	normalized, err := normalizeRun(run)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runs[normalized.RunID]; exists {
		return fmt.Errorf("%w: run_id=%s", ErrAlreadyExists, normalized.RunID)
	}
	s.runs[normalized.RunID] = normalized
	return nil
}

func (s *MemoryStore) UpdateRun(_ context.Context, run Run) error {
	if s == nil {
		return nil
	}
	normalized, err := normalizeRun(run)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runs[normalized.RunID]; !exists {
		return fmt.Errorf("%w: run_id=%s", ErrNotFound, normalized.RunID)
	}
	s.runs[normalized.RunID] = normalized
	return nil
}

func (s *MemoryStore) GetRun(_ context.Context, runID string) (Run, error) {
	if s == nil {
		return Run{}, ErrNotFound
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Run{}, fmt.Errorf("agentx/runstore: run id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[runID]
	if !ok {
		return Run{}, fmt.Errorf("%w: run_id=%s", ErrNotFound, runID)
	}
	return run, nil
}

func (s *MemoryStore) AppendEvent(_ context.Context, event Event) error {
	if s == nil {
		return nil
	}
	normalized, err := normalizeEvent(event)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runs[normalized.RunID]; !exists {
		return fmt.Errorf("%w: run_id=%s", ErrNotFound, normalized.RunID)
	}
	for _, events := range s.events {
		for _, existing := range events {
			if existing.EventID == normalized.EventID {
				return fmt.Errorf("%w: event_id=%s", ErrAlreadyExists, normalized.EventID)
			}
		}
	}
	s.events[normalized.RunID] = append(s.events[normalized.RunID], normalized)
	sort.SliceStable(s.events[normalized.RunID], func(i, j int) bool {
		left := s.events[normalized.RunID][i]
		right := s.events[normalized.RunID][j]
		if left.CreatedAt != right.CreatedAt {
			return left.CreatedAt < right.CreatedAt
		}
		return left.EventID < right.EventID
	})
	return nil
}

func (s *MemoryStore) ListEvents(_ context.Context, runID string, limit int) ([]Event, error) {
	if s == nil {
		return nil, nil
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("agentx/runstore: run id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]Event(nil), s.events[runID]...)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *MemoryStore) UpsertNodeExecution(_ context.Context, node NodeExecution) error {
	if s == nil {
		return nil
	}
	normalized, err := normalizeNodeExecution(node)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runs[normalized.RunID]; !exists {
		return fmt.Errorf("%w: run_id=%s", ErrNotFound, normalized.RunID)
	}
	if s.nodeExecs[normalized.RunID] == nil {
		s.nodeExecs[normalized.RunID] = map[string]NodeExecution{}
	}
	s.nodeExecs[normalized.RunID][normalized.NodeExecID] = normalized
	return nil
}

func (s *MemoryStore) ListNodeExecutions(_ context.Context, runID string) ([]NodeExecution, error) {
	if s == nil {
		return nil, nil
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("agentx/runstore: run id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]NodeExecution, 0, len(s.nodeExecs[runID]))
	for _, item := range s.nodeExecs[runID] {
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartedAt != items[j].StartedAt {
			return items[i].StartedAt < items[j].StartedAt
		}
		return items[i].NodeExecID < items[j].NodeExecID
	})
	return items, nil
}

func normalizeRun(run Run) (Run, error) {
	run.RunID = strings.TrimSpace(run.RunID)
	if run.RunID == "" {
		return Run{}, fmt.Errorf("agentx/runstore: run id is required")
	}
	if run.Attempt < 0 {
		run.Attempt = 0
	}
	return run, nil
}

func normalizeEvent(event Event) (Event, error) {
	event.EventID = strings.TrimSpace(event.EventID)
	event.RunID = strings.TrimSpace(event.RunID)
	if event.EventID == "" || event.RunID == "" {
		return Event{}, fmt.Errorf("agentx/runstore: event id and run id are required")
	}
	return event, nil
}

func normalizeNodeExecution(node NodeExecution) (NodeExecution, error) {
	node.NodeExecID = strings.TrimSpace(node.NodeExecID)
	node.RunID = strings.TrimSpace(node.RunID)
	if node.NodeExecID == "" || node.RunID == "" {
		return NodeExecution{}, fmt.Errorf("agentx/runstore: node exec id and run id are required")
	}
	if node.Attempt < 0 {
		node.Attempt = 0
	}
	return node, nil
}
