package browserd

import (
	"context"
	"sync"
	"time"
)

type ManagedStarterOptions struct {
	WorkspaceRoot    string
	Plan             Plan
	Probe            StatusProbe
	TransportTimeout int
	HealthTimeout    time.Duration
	ProbeInterval    time.Duration
}

type ManagedStarter struct {
	opts ManagedStarterOptions

	mu      sync.Mutex
	current *Manager
}

func NewManagedStarter(opts ManagedStarterOptions) *ManagedStarter {
	return &ManagedStarter{opts: opts}
}

func (s *ManagedStarter) EnsureStarted(ctx context.Context) error {
	if s == nil {
		return nil
	}
	manager, err := s.ensureManager()
	if err != nil {
		return err
	}
	if err := manager.EnsureStarted(ctx); err != nil {
		s.discardManager(manager)
		return err
	}
	return nil
}

func (s *ManagedStarter) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	manager := s.current
	s.current = nil
	s.mu.Unlock()
	if manager == nil {
		return nil
	}
	return manager.Close()
}

func (s *ManagedStarter) ensureManager() (*Manager, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		return s.current, nil
	}
	manager, err := NewManager(ManagerOptions{
		WorkspaceRoot:    s.opts.WorkspaceRoot,
		Plan:             s.opts.Plan,
		Probe:            s.opts.Probe,
		TransportTimeout: s.opts.TransportTimeout,
		HealthTimeout:    s.opts.HealthTimeout,
		ProbeInterval:    s.opts.ProbeInterval,
	})
	if err != nil {
		return nil, err
	}
	s.current = manager
	return manager, nil
}

func (s *ManagedStarter) discardManager(manager *Manager) {
	s.mu.Lock()
	if s.current == manager {
		s.current = nil
	}
	s.mu.Unlock()
}
