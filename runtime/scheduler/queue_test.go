package scheduler

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMemoryQueueEnqueueDequeueAckFail(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	job := Job{
		ID:        "job-1",
		Lane:      LaneMain,
		SessionID: "session-1",
		Payload:   "{}",
	}
	if err := queue.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	dequeued, err := queue.Dequeue(ctx, LaneMain)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if dequeued.ID != "job-1" {
		t.Fatalf("unexpected dequeued job: %#v", dequeued)
	}
	if err := queue.Ack(ctx, Result{JobID: dequeued.ID, Lane: LaneMain}); err != nil {
		t.Fatalf("ack: %v", err)
	}
	ackResult, ok, err := queue.Result(ctx, dequeued.ID)
	if err != nil {
		t.Fatalf("result ack: %v", err)
	}
	if !ok || !ackResult.Succeeded {
		t.Fatalf("expected ack result, got %#v", ackResult)
	}
	if err := queue.Fail(ctx, Result{JobID: dequeued.ID, Lane: LaneMain, Error: "forced"}); err != nil {
		t.Fatalf("fail: %v", err)
	}
	failResult, ok, err := queue.Result(ctx, dequeued.ID)
	if err != nil {
		t.Fatalf("result fail: %v", err)
	}
	if !ok || failResult.Succeeded || failResult.Error != "forced" {
		t.Fatalf("expected failed result override, got %#v", failResult)
	}
}

func TestMemoryQueueDequeueByKindKeepsOtherJobsPending(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Enqueue(ctx, Job{ID: "job-other", Lane: LaneBackground, JobKind: "other_kind"}); err != nil {
		t.Fatalf("enqueue other: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-target", Lane: LaneBackground, JobKind: "target_kind"}); err != nil {
		t.Fatalf("enqueue target: %v", err)
	}
	target, err := queue.DequeueByKind(ctx, LaneBackground, "target_kind")
	if err != nil {
		t.Fatalf("dequeue by kind: %v", err)
	}
	if target.ID != "job-target" {
		t.Fatalf("unexpected target job: %#v", target)
	}
	otherPending, err := queue.Pending(ctx, "job-other")
	if err != nil {
		t.Fatalf("pending other: %v", err)
	}
	if !otherPending {
		t.Fatalf("expected other job to remain pending")
	}
	other, err := queue.Dequeue(ctx, LaneBackground)
	if err != nil {
		t.Fatalf("dequeue other: %v", err)
	}
	if other.ID != "job-other" {
		t.Fatalf("unexpected other job: %#v", other)
	}
}

func TestMemoryQueueRespectsLaneLimit(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneMain: 1,
		},
	})
	if err := queue.Enqueue(ctx, Job{ID: "job-1", Lane: LaneMain}); err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-2", Lane: LaneMain}); err == nil {
		t.Fatalf("expected queue limit error")
	}
}

func TestMemoryQueuePending(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Enqueue(ctx, Job{ID: "job-pending", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	pending, err := queue.Pending(ctx, "job-pending")
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if !pending {
		t.Fatalf("expected pending job")
	}
	if _, err := queue.Dequeue(ctx, LaneSubtask); err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	pending, err = queue.Pending(ctx, "job-pending")
	if err != nil {
		t.Fatalf("pending after dequeue: %v", err)
	}
	if pending {
		t.Fatalf("expected no pending job after dequeue")
	}
}

func TestMemoryQueuePendingIgnoresTerminalGhostResult(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Enqueue(ctx, Job{ID: "job-ghost", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := queue.Ack(ctx, Result{JobID: "job-ghost", Lane: LaneSubtask}); err != nil {
		t.Fatalf("ack terminal ghost: %v", err)
	}
	pending, err := queue.Pending(ctx, "job-ghost")
	if err != nil {
		t.Fatalf("pending after terminal ghost: %v", err)
	}
	if pending {
		t.Fatalf("expected pending=false when terminal result exists for queued ghost")
	}
}

func TestMemoryQueueReenqueueClearsTerminalGhostPending(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Enqueue(ctx, Job{ID: "job-ghost", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue initial ghost: %v", err)
	}
	if err := queue.Ack(ctx, Result{JobID: "job-ghost", Lane: LaneSubtask}); err != nil {
		t.Fatalf("ack terminal ghost: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-ghost", Lane: LaneSubtask}); err != nil {
		t.Fatalf("reenqueue after terminal ghost: %v", err)
	}
	job, err := queue.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue reenqueued job: %v", err)
	}
	if job.ID != "job-ghost" {
		t.Fatalf("unexpected dequeued job after reenqueuing ghost: %#v", job)
	}
	if _, err := queue.Dequeue(ctx, LaneSubtask); !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("expected queue empty after single fresh enqueue, got %v", err)
	}
}

func TestMemoryQueueDequeueSkipsTerminalJobs(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Enqueue(ctx, Job{ID: "job-canceled", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue canceled candidate: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-next", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue next job: %v", err)
	}
	if err := queue.Fail(ctx, Result{JobID: "job-canceled", Lane: LaneSubtask, Outcome: ResultOutcomeCanceled, Error: "manual cancel"}); err != nil {
		t.Fatalf("mark canceled candidate: %v", err)
	}
	job, err := queue.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if job.ID != "job-next" {
		t.Fatalf("expected next runnable job, got %#v", job)
	}
}

func TestMemoryQueueAckDoesNotOverrideFailure(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Fail(ctx, Result{JobID: "job-1", Lane: LaneMain, Outcome: ResultOutcomeCanceled, Error: "manual cancel"}); err != nil {
		t.Fatalf("fail: %v", err)
	}
	if err := queue.Ack(ctx, Result{JobID: "job-1", Lane: LaneMain}); err != nil {
		t.Fatalf("ack: %v", err)
	}
	result, ok, err := queue.Result(ctx, "job-1")
	if err != nil {
		t.Fatalf("result: %v", err)
	}
	if !ok {
		t.Fatalf("expected result after fail+ack")
	}
	if result.Succeeded {
		t.Fatalf("expected failed result to remain terminal, got %#v", result)
	}
	if !strings.Contains(result.Error, "cancel") {
		t.Fatalf("expected preserved cancel reason, got %#v", result)
	}
}

func TestMemoryQueueErrorTextCannotSelectCanceledOutcome(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{})
	if err := queue.Fail(ctx, Result{
		JobID: "job-cancel-word",
		Lane:  LaneMain,
		Error: "cancel order failed upstream",
	}); err != nil {
		t.Fatalf("fail: %v", err)
	}
	result, ok, err := queue.Result(ctx, "job-cancel-word")
	if err != nil {
		t.Fatalf("result: %v", err)
	}
	if !ok {
		t.Fatal("expected terminal result")
	}
	if result.Outcome != ResultOutcomeFailed || result.Status != string(ResultOutcomeFailed) {
		t.Fatalf("expected typed failed outcome independent of error text, got %#v", result)
	}
}

func TestMemoryQueueEnqueueIsIdempotentForPendingJob(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneSubtask: 1,
		},
	})
	if err := queue.Enqueue(ctx, Job{ID: "job-repeat", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-repeat", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue duplicate should be idempotent, got: %v", err)
	}
	dequeued, err := queue.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue first: %v", err)
	}
	if dequeued.ID != "job-repeat" {
		t.Fatalf("unexpected dequeued job: %#v", dequeued)
	}
	if _, err := queue.Dequeue(ctx, LaneSubtask); err != ErrQueueEmpty {
		t.Fatalf("expected queue empty after single logical enqueue, got %v", err)
	}
}

func TestMemoryQueueLimitCountsOnlyPendingJobs(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		LaneQueueLimit: map[Lane]int{
			LaneSubtask: 1,
		},
	})
	if err := queue.Enqueue(ctx, Job{ID: "job-1", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if _, err := queue.Dequeue(ctx, LaneSubtask); err != nil {
		t.Fatalf("dequeue first: %v", err)
	}
	if err := queue.Enqueue(ctx, Job{ID: "job-2", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue after dequeue should pass queue limit check, got %v", err)
	}
	job, err := queue.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue second: %v", err)
	}
	if job.ID != "job-2" {
		t.Fatalf("expected second job to be dequeued, got %#v", job)
	}
}

func TestMemoryQueueResultLimitEvictsOldest(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		ResultLimit: 2,
		ResultTTL:   time.Hour,
	})
	if err := queue.Fail(ctx, Result{JobID: "job-1", Lane: LaneMain, Error: "e1"}); err != nil {
		t.Fatalf("fail job-1: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := queue.Fail(ctx, Result{JobID: "job-2", Lane: LaneMain, Error: "e2"}); err != nil {
		t.Fatalf("fail job-2: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := queue.Fail(ctx, Result{JobID: "job-3", Lane: LaneMain, Error: "e3"}); err != nil {
		t.Fatalf("fail job-3: %v", err)
	}
	if _, ok, err := queue.Result(ctx, "job-1"); err != nil {
		t.Fatalf("result job-1: %v", err)
	} else if ok {
		t.Fatalf("expected oldest result to be evicted")
	}
	if _, ok, err := queue.Result(ctx, "job-2"); err != nil {
		t.Fatalf("result job-2: %v", err)
	} else if !ok {
		t.Fatalf("expected recent result to remain")
	}
	if _, ok, err := queue.Result(ctx, "job-3"); err != nil {
		t.Fatalf("result job-3: %v", err)
	} else if !ok {
		t.Fatalf("expected recent result to remain")
	}
	size, evicted := queue.ResultStats()
	if size != 2 || evicted == 0 {
		t.Fatalf("unexpected result stats: size=%d evicted=%d", size, evicted)
	}
}

func TestMemoryQueueResultTTLEvictsExpiredEntries(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue(QueueConfig{
		ResultLimit: 16,
		ResultTTL:   15 * time.Millisecond,
	})
	if err := queue.Ack(ctx, Result{JobID: "job-ttl", Lane: LaneMain}); err != nil {
		t.Fatalf("ack: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok, err := queue.Result(ctx, "job-ttl"); err != nil {
		t.Fatalf("result: %v", err)
	} else if ok {
		t.Fatalf("expected ttl-evicted result")
	}
	size, evicted := queue.ResultStats()
	if size != 0 || evicted == 0 {
		t.Fatalf("unexpected ttl stats: size=%d evicted=%d", size, evicted)
	}
}
