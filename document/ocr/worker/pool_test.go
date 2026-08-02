package worker

import (
	"context"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

func TestNewPoolAppliesDefaults(t *testing.T) {
	pool := NewPool(config.WorkerConfig{MaxConcurrent: 0, QueueSize: -1})
	if cap(pool.sem) != 1 {
		t.Fatalf("expected default concurrency 1, got %d", cap(pool.sem))
	}
	if pool.queue != nil {
		t.Fatalf("expected negative queue size to disable queue")
	}
}

func TestAcquireBlocksWithoutQueueAndRespectsContext(t *testing.T) {
	pool := NewPool(config.WorkerConfig{MaxConcurrent: 1})
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := pool.Acquire(ctx)
	if err == nil {
		t.Fatal("expected acquire to fail when context times out")
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatalf("expected acquire to block before timing out, elapsed=%s", time.Since(start))
	}
}

func TestAcquireUsesQueueBackpressure(t *testing.T) {
	pool := NewPool(config.WorkerConfig{MaxConcurrent: 1, QueueSize: 1})
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- pool.Acquire(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := pool.Acquire(ctx); err == nil {
		t.Fatal("expected queued acquire to fail when queue is full")
	}

	pool.Release()
	if err := <-waitDone; err != nil {
		t.Fatalf("expected queued acquire to succeed after release: %v", err)
	}
	pool.Release()
}

func TestReleaseIsSafeWithoutAcquire(t *testing.T) {
	pool := NewPool(config.WorkerConfig{MaxConcurrent: 1})
	pool.Release()
}
