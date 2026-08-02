package worker

import (
	"context"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

// Pool is a lightweight concurrency limiter around a semaphore.
type Pool struct {
	sem   chan struct{}
	queue chan struct{}
}

// NewPool creates a new pool from configuration.
func NewPool(cfg config.WorkerConfig) *Pool {
	concurrency := cfg.MaxConcurrent
	if concurrency <= 0 {
		concurrency = 1
	}
	queueSize := cfg.QueueSize
	if queueSize < 0 {
		queueSize = 0
	}
	var queue chan struct{}
	if queueSize > 0 {
		queue = make(chan struct{}, queueSize)
	}
	return &Pool{
		sem:   make(chan struct{}, concurrency),
		queue: queue,
	}
}

// Acquire obtains capacity failing fast when context is done.
func (p *Pool) Acquire(ctx context.Context) error {
	select {
	case p.sem <- struct{}{}:
		return nil
	default:
	}

	if p.queue == nil {
		select {
		case p.sem <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case p.queue <- struct{}{}:
		defer func() { <-p.queue }()
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case p.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot.
func (p *Pool) Release() {
	select {
	case <-p.sem:
	default:
	}
}
