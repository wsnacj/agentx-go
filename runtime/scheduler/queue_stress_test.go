package scheduler

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryQueueConcurrentLifecycle(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneMain:       512,
			LaneSubtask:    512,
			LaneBackground: 512,
		},
	})

	const totalJobs = 240
	var processed atomic.Int64
	var producersDone atomic.Bool

	jobIDs := make([]string, 0, totalJobs)
	for i := 0; i < totalJobs; i++ {
		jobIDs = append(jobIDs, fmt.Sprintf("concurrent-job-%03d", i))
	}

	var seen sync.Map
	var producerWG sync.WaitGroup
	for producer := 0; producer < 4; producer++ {
		producerWG.Add(1)
		go func(worker int) {
			defer producerWG.Done()
			for i := worker; i < totalJobs; i += 4 {
				lane := LaneMain
				switch i % 3 {
				case 1:
					lane = LaneSubtask
				case 2:
					lane = LaneBackground
				}
				if err := queue.Enqueue(ctx, Job{
					ID:        jobIDs[i],
					Lane:      lane,
					SessionID: fmt.Sprintf("session-%03d", i),
					Payload:   fmt.Sprintf(`{"index":%d}`, i),
				}); err != nil {
					t.Errorf("enqueue %s: %v", jobIDs[i], err)
					return
				}
			}
		}(producer)
	}

	var consumerWG sync.WaitGroup
	lanes := []Lane{LaneMain, LaneSubtask, LaneBackground}
	for _, lane := range lanes {
		for worker := 0; worker < 2; worker++ {
			consumerWG.Add(1)
			go func(lane Lane) {
				defer consumerWG.Done()
				for {
					job, err := queue.Dequeue(ctx, lane)
					if err == nil {
						if _, loaded := seen.LoadOrStore(job.ID, lane); loaded {
							t.Errorf("duplicate dequeue for job %s", job.ID)
							return
						}
						if err := queue.Ack(ctx, Result{JobID: job.ID, Lane: lane}); err != nil {
							t.Errorf("ack %s: %v", job.ID, err)
							return
						}
						if processed.Add(1) == totalJobs {
							return
						}
						continue
					}
					if !errors.Is(err, ErrQueueEmpty) {
						t.Errorf("dequeue %s: %v", lane, err)
						return
					}
					if producersDone.Load() && processed.Load() >= totalJobs {
						return
					}
					runtime.Gosched()
					time.Sleep(1 * time.Millisecond)
				}
			}(lane)
		}
	}

	producerWG.Wait()
	producersDone.Store(true)
	consumerWG.Wait()

	if got := processed.Load(); got != totalJobs {
		t.Fatalf("expected %d processed jobs, got %d", totalJobs, got)
	}
	for _, jobID := range jobIDs {
		pending, err := queue.Pending(ctx, jobID)
		if err != nil {
			t.Fatalf("pending %s: %v", jobID, err)
		}
		if pending {
			t.Fatalf("expected job %s not pending after lifecycle", jobID)
		}
		result, ok, err := queue.Result(ctx, jobID)
		if err != nil {
			t.Fatalf("result %s: %v", jobID, err)
		}
		if !ok || !result.Succeeded || result.Status != "completed" {
			t.Fatalf("expected completed result for %s, got %#v ok=%t", jobID, result, ok)
		}
	}
}

func BenchmarkMemoryQueueLifecycle(b *testing.B) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneMain: 2 * b.N,
		},
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobID := fmt.Sprintf("bench-job-%d", i)
		if err := queue.Enqueue(ctx, Job{
			ID:        jobID,
			Lane:      LaneMain,
			SessionID: "bench-session",
			Payload:   "{}",
		}); err != nil {
			b.Fatalf("enqueue: %v", err)
		}
		job, err := queue.Dequeue(ctx, LaneMain)
		if err != nil {
			b.Fatalf("dequeue: %v", err)
		}
		if job.ID != jobID {
			b.Fatalf("unexpected dequeued job: got=%s want=%s", job.ID, jobID)
		}
		if err := queue.Ack(ctx, Result{JobID: jobID, Lane: LaneMain}); err != nil {
			b.Fatalf("ack: %v", err)
		}
	}
}

func BenchmarkMemoryQueueMixedLanesBatchLifecycle(b *testing.B) {
	ctx := context.Background()
	const batchSize = 12
	lanes := []Lane{LaneMain, LaneSubtask, LaneBackground}
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneMain:       batchSize * b.N,
			LaneSubtask:    batchSize * b.N,
			LaneBackground: batchSize * b.N,
		},
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			lane := lanes[j%len(lanes)]
			jobID := fmt.Sprintf("bench-mixed-job-%d-%02d", i, j)
			if err := queue.Enqueue(ctx, Job{
				ID:        jobID,
				Lane:      lane,
				SessionID: "bench-mixed-session",
				Payload:   "{}",
			}); err != nil {
				b.Fatalf("enqueue %s: %v", jobID, err)
			}
		}
		for j := 0; j < batchSize; j++ {
			lane := lanes[j%len(lanes)]
			jobID := fmt.Sprintf("bench-mixed-job-%d-%02d", i, j)
			job, err := queue.Dequeue(ctx, lane)
			if err != nil {
				b.Fatalf("dequeue %s: %v", lane, err)
			}
			if job.ID != jobID {
				b.Fatalf("unexpected dequeued job: got=%s want=%s", job.ID, jobID)
			}
			if err := queue.Ack(ctx, Result{JobID: jobID, Lane: lane}); err != nil {
				b.Fatalf("ack %s: %v", jobID, err)
			}
		}
	}
	b.StopTimer()

	for _, lane := range lanes {
		if _, err := queue.Dequeue(ctx, lane); !errors.Is(err, ErrQueueEmpty) {
			b.Fatalf("expected %s empty after mixed benchmark, got %v", lane, err)
		}
	}
}
