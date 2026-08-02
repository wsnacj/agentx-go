package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatcherProcessesJobsAcrossLanes(t *testing.T) {
	queue := NewMemoryQueue(QueueConfig{})
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{
			LaneMain:       1,
			LaneSubtask:    2,
			LaneBackground: 1,
		},
		PollInterval: 5 * time.Millisecond,
	})
	var mainCount int64
	var subtaskCount int64
	var backgroundCount int64
	if err := dispatcher.RegisterHandler(LaneMain, func(_ context.Context, _ Job) error {
		atomic.AddInt64(&mainCount, 1)
		return nil
	}); err != nil {
		t.Fatalf("register main: %v", err)
	}
	if err := dispatcher.RegisterHandler(LaneSubtask, func(_ context.Context, _ Job) error {
		atomic.AddInt64(&subtaskCount, 1)
		return nil
	}); err != nil {
		t.Fatalf("register subtask: %v", err)
	}
	if err := dispatcher.RegisterHandler(LaneBackground, func(_ context.Context, _ Job) error {
		atomic.AddInt64(&backgroundCount, 1)
		return nil
	}); err != nil {
		t.Fatalf("register background: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	for i := 0; i < 3; i++ {
		if err := dispatcher.Enqueue(ctx, Job{ID: fmt.Sprintf("m-%d", i), Lane: LaneMain}); err != nil {
			t.Fatalf("enqueue main %d: %v", i, err)
		}
	}
	for i := 0; i < 4; i++ {
		if err := dispatcher.Enqueue(ctx, Job{ID: fmt.Sprintf("s-%d", i), Lane: LaneSubtask}); err != nil {
			t.Fatalf("enqueue subtask %d: %v", i, err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := dispatcher.Enqueue(ctx, Job{ID: fmt.Sprintf("b-%d", i), Lane: LaneBackground}); err != nil {
			t.Fatalf("enqueue background %d: %v", i, err)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&mainCount) == 3 &&
			atomic.LoadInt64(&subtaskCount) == 4 &&
			atomic.LoadInt64(&backgroundCount) == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if atomic.LoadInt64(&mainCount) != 3 ||
		atomic.LoadInt64(&subtaskCount) != 4 ||
		atomic.LoadInt64(&backgroundCount) != 2 {
		t.Fatalf(
			"unexpected handler counts: main=%d subtask=%d background=%d",
			mainCount,
			subtaskCount,
			backgroundCount,
		)
	}
	cancel()
	dispatcher.Wait()
	metrics := dispatcher.Metrics()
	if metrics[LaneMain].Acked != 3 || metrics[LaneSubtask].Acked != 4 || metrics[LaneBackground].Acked != 2 {
		t.Fatalf("unexpected ack metrics: %#v", metrics)
	}
}

func TestDispatcherProjectsHandlerErrorAndPanicInQueueResult(t *testing.T) {
	tests := map[string]struct {
		handler Handler
		outcome ResultOutcome
	}{
		"error": {
			handler: func(context.Context, Job) error { return errors.New("scheduler-error-secret-sentinel") },
			outcome: ResultOutcomeFailed,
		},
		"panic": {
			handler: func(context.Context, Job) error { panic("scheduler-panic-secret-sentinel") },
			outcome: ResultOutcomeFailed,
		},
		"context_canceled": {
			handler: func(context.Context, Job) error { return context.Canceled },
			outcome: ResultOutcomeCanceled,
		},
	}
	for name, tt := range tests {
		handler := tt.handler
		expectedOutcome := tt.outcome
		t.Run(name, func(t *testing.T) {
			queue := NewMemoryQueue(QueueConfig{})
			dispatcher := NewDispatcher(queue, DispatcherConfig{LaneConcurrency: map[Lane]int{LaneMain: 1}, PollInterval: time.Millisecond})
			if err := dispatcher.RegisterHandler(LaneMain, handler); err != nil {
				t.Fatalf("register: %v", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			dispatcher.Run(ctx)
			jobID := "safe-" + name
			if err := dispatcher.Enqueue(ctx, Job{ID: jobID, Lane: LaneMain}); err != nil {
				t.Fatalf("enqueue: %v", err)
			}
			var result Result
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				if current, ok, err := queue.Result(ctx, jobID); err == nil && ok {
					result = current
					break
				}
				time.Sleep(time.Millisecond)
			}
			cancel()
			dispatcher.Wait()
			if result.JobID == "" {
				t.Fatal("missing queue result")
			}
			if strings.Contains(result.Error, "secret-sentinel") || !strings.Contains(result.Error, "identity=") {
				t.Fatalf("unsafe queue result: %#v", result)
			}
			if result.Outcome != expectedOutcome {
				t.Fatalf("expected typed %s outcome, got %#v", expectedOutcome, result)
			}
		})
	}
}

func TestDispatcherWaitContextReportsNonCooperativeHandler(t *testing.T) {
	queue := NewMemoryQueue(QueueConfig{})
	dispatcher := NewDispatcher(queue, DispatcherConfig{LaneConcurrency: map[Lane]int{
		LaneMain: 1,
	}})
	started := make(chan struct{})
	release := make(chan struct{})
	if err := dispatcher.RegisterHandler(LaneMain, func(context.Context, Job) error {
		close(started)
		<-release
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	if err := dispatcher.Enqueue(context.Background(), Job{ID: "non-cooperative", Lane: LaneMain}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	runCtx, cancelRun := context.WithCancel(context.Background())
	dispatcher.Run(runCtx)
	<-started
	cancelRun()
	waitCtx, cancelWait := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelWait()
	if err := dispatcher.WaitContext(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait error=%v, want deadline", err)
	}
	close(release)
	waitCtx, cancelWait = context.WithTimeout(context.Background(), time.Second)
	defer cancelWait()
	if err := dispatcher.WaitContext(waitCtx); err != nil {
		t.Fatalf("wait after release: %v", err)
	}
}

type heartbeatQueueForTest struct {
	*MemoryQueue
	beats    atomic.Int64
	interval time.Duration
}

func (q *heartbeatQueueForTest) Heartbeat(_ context.Context, _ Job) error {
	q.beats.Add(1)
	return nil
}

func (q *heartbeatQueueForTest) HeartbeatInterval() time.Duration {
	return q.interval
}

func TestDispatcherLeaseHeartbeat(t *testing.T) {
	queue := &heartbeatQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		interval:    5 * time.Millisecond,
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneSubtask: 1},
		PollInterval:    2 * time.Millisecond,
	})
	if err := dispatcher.RegisterHandler(LaneSubtask, func(_ context.Context, _ Job) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "hb-1", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		metrics := dispatcher.Metrics()
		if metrics[LaneSubtask].Acked >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	dispatcher.Wait()
	if queue.beats.Load() == 0 {
		t.Fatalf("expected dispatcher to emit heartbeats while handling running job")
	}
}

type ackFailQueueForTest struct {
	*MemoryQueue
	ackErr  error
	failErr error
	failed  atomic.Int64
}

func (q *ackFailQueueForTest) Ack(_ context.Context, _ Result) error {
	if q.ackErr == nil {
		return errors.New("ack failed")
	}
	return q.ackErr
}

func (q *ackFailQueueForTest) Fail(ctx context.Context, result Result) error {
	q.failed.Add(1)
	if q.failErr != nil {
		return q.failErr
	}
	return q.MemoryQueue.Fail(ctx, result)
}

func TestDispatcherAckFailureCompensatesAndRecordsMetrics(t *testing.T) {
	queue := &ackFailQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		ackErr:      errors.New("ack failed for test"),
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneMain: 1},
		PollInterval:    2 * time.Millisecond,
	})
	if err := dispatcher.RegisterHandler(LaneMain, func(_ context.Context, _ Job) error { return nil }); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "ack-fail-1", Lane: LaneMain}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		metrics := dispatcher.Metrics()[LaneMain]
		if metrics.AckFailed >= 1 && metrics.Failed >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	dispatcher.Wait()
	metrics := dispatcher.Metrics()[LaneMain]
	if metrics.AckFailed == 0 || metrics.Failed == 0 {
		t.Fatalf("expected ack failure metrics, got: %#v", metrics)
	}
	if metrics.Acked != 0 {
		t.Fatalf("expected acked=0 when ack always fails, got: %#v", metrics)
	}
	if queue.failed.Load() == 0 {
		t.Fatalf("expected fail compensation to be called at least once")
	}
}

func TestDispatcherFailPathFailureIsRecorded(t *testing.T) {
	queue := &ackFailQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		ackErr:      errors.New("ack failed for test"),
		failErr:     errors.New("fail failed for test"),
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneMain: 1},
		PollInterval:    2 * time.Millisecond,
	})
	if err := dispatcher.RegisterHandler(LaneMain, func(_ context.Context, _ Job) error { return nil }); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "fail-fail-1", Lane: LaneMain}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		metrics := dispatcher.Metrics()[LaneMain]
		if metrics.FailFailed >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	dispatcher.Wait()
	metrics := dispatcher.Metrics()[LaneMain]
	if metrics.FailFailed == 0 {
		t.Fatalf("expected fail_failed metric to increase, got: %#v", metrics)
	}
}

type heartbeatFailQueueForTest struct {
	*MemoryQueue
	interval time.Duration
	beats    atomic.Int64
}

func (q *heartbeatFailQueueForTest) Heartbeat(_ context.Context, _ Job) error {
	q.beats.Add(1)
	return errors.New("heartbeat failed for test")
}

func (q *heartbeatFailQueueForTest) HeartbeatInterval() time.Duration {
	return q.interval
}

type heartbeatCancelSensitiveQueueForTest struct {
	*MemoryQueue
	interval      time.Duration
	heartbeatErr  error
	failDelay     time.Duration
	beats         atomic.Int64
	failCalls     atomic.Int64
	failCanceled  atomic.Int64
	failCompleted atomic.Int64
}

type heartbeatOrderingQueueForTest struct {
	*MemoryQueue
	interval       time.Duration
	beats          atomic.Int64
	failCalls      atomic.Int64
	handlerDone    atomic.Bool
	failBeforeDone atomic.Bool
}

func (q *heartbeatOrderingQueueForTest) Heartbeat(context.Context, Job) error {
	q.beats.Add(1)
	return errors.New("heartbeat lease lost for ordering test")
}

func (q *heartbeatOrderingQueueForTest) HeartbeatInterval() time.Duration {
	return q.interval
}

func (q *heartbeatOrderingQueueForTest) Fail(ctx context.Context, result Result) error {
	q.failCalls.Add(1)
	if !q.handlerDone.Load() {
		q.failBeforeDone.Store(true)
	}
	return q.MemoryQueue.Fail(ctx, result)
}

func (q *heartbeatCancelSensitiveQueueForTest) Heartbeat(_ context.Context, _ Job) error {
	q.beats.Add(1)
	if q.heartbeatErr != nil {
		return q.heartbeatErr
	}
	return errors.New("heartbeat failed for test")
}

func (q *heartbeatCancelSensitiveQueueForTest) HeartbeatInterval() time.Duration {
	return q.interval
}

func (q *heartbeatCancelSensitiveQueueForTest) Fail(ctx context.Context, result Result) error {
	q.failCalls.Add(1)
	if q.failDelay > 0 {
		time.Sleep(q.failDelay)
	}
	select {
	case <-ctx.Done():
		q.failCanceled.Add(1)
		return ctx.Err()
	default:
	}
	q.failCompleted.Add(1)
	return q.MemoryQueue.Fail(ctx, result)
}

func TestDispatcherHeartbeatFailureRecordsMetrics(t *testing.T) {
	queue := &heartbeatFailQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		interval:    5 * time.Millisecond,
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneSubtask: 1},
		PollInterval:    2 * time.Millisecond,
	})
	if err := dispatcher.RegisterHandler(LaneSubtask, func(_ context.Context, _ Job) error {
		time.Sleep(25 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "heartbeat-fail-1", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		metrics := dispatcher.Metrics()[LaneSubtask]
		if metrics.HeartbeatFailed >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	dispatcher.Wait()
	metrics := dispatcher.Metrics()[LaneSubtask]
	if metrics.HeartbeatFailed == 0 || metrics.Failed == 0 {
		t.Fatalf("expected heartbeat failure metrics, got: %#v", metrics)
	}
	if metrics.Acked != 0 {
		t.Fatalf("expected no ack after heartbeat failure, got: %#v", metrics)
	}
}

func TestDispatcherHeartbeatFailureUsesDetachedFailContext(t *testing.T) {
	queue := &heartbeatCancelSensitiveQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		interval:    2 * time.Millisecond,
		failDelay:   20 * time.Millisecond,
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneSubtask: 1},
		PollInterval:    2 * time.Millisecond,
	})
	if err := dispatcher.RegisterHandler(LaneSubtask, func(_ context.Context, _ Job) error {
		deadline := time.Now().Add(100 * time.Millisecond)
		for queue.beats.Load() == 0 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "heartbeat-detached-fail-1", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		result, ok, err := queue.Result(context.Background(), "heartbeat-detached-fail-1")
		if err != nil {
			t.Fatalf("result: %v", err)
		}
		if ok {
			if result.Status != "failed" {
				t.Fatalf("expected failed result, got %#v", result)
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	dispatcher.Wait()
	if queue.failCalls.Load() == 0 {
		t.Fatalf("expected heartbeat fail path to call queue.Fail")
	}
	if queue.failCanceled.Load() != 0 {
		t.Fatalf("expected detached fail context to avoid cancellation, got failCanceled=%d", queue.failCanceled.Load())
	}
	if queue.failCompleted.Load() == 0 {
		t.Fatalf("expected detached fail context to persist terminal failure")
	}
}

func TestDispatcherHeartbeatFailureCancelsHandlerBeforeFail(t *testing.T) {
	queue := &heartbeatOrderingQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		interval:    2 * time.Millisecond,
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneSubtask: 1},
		PollInterval:    2 * time.Millisecond,
	})
	attempts := make(chan ExecutionAttempt, 1)
	causes := make(chan error, 1)
	if err := dispatcher.RegisterHandler(LaneSubtask, func(ctx context.Context, _ Job) error {
		attempt, ok := ExecutionAttemptFromContext(ctx)
		if !ok {
			return errors.New("scheduler attempt missing from handler context")
		}
		attempts <- attempt
		<-ctx.Done()
		causes <- context.Cause(ctx)
		queue.handlerDone.Store(true)
		return ctx.Err()
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	job := Job{
		ID:             "heartbeat-cancel-handler-1",
		Lane:           LaneSubtask,
		Attempt:        3,
		IdempotencyKey: "idem-heartbeat-cancel-handler-1",
	}
	if err := dispatcher.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	select {
	case attempt := <-attempts:
		if attempt.JobID != job.ID || attempt.Lane != job.Lane || attempt.Attempt != job.Attempt || attempt.IdempotencyKey != job.IdempotencyKey {
			t.Fatalf("unexpected execution attempt: %#v", attempt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for handler attempt context")
	}
	select {
	case cause := <-causes:
		if !errors.Is(cause, ErrLeaseLost) {
			t.Fatalf("expected lease-lost cancellation cause, got %v", cause)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for handler cancellation")
	}

	deadline := time.Now().Add(2 * time.Second)
	for queue.failCalls.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if queue.failCalls.Load() == 0 {
		t.Fatal("expected queue failure after handler cancellation")
	}
	if queue.failBeforeDone.Load() {
		t.Fatal("queue failure ran before cooperative handler exited")
	}
	metrics := dispatcher.Metrics()[LaneSubtask]
	if metrics.LeaseCancellationObserved != 1 || metrics.StaleHandlerCompleted != 0 {
		t.Fatalf("unexpected lease cancellation metrics: %#v", metrics)
	}
	cancel()
	dispatcher.Wait()
}

func TestDispatcherHeartbeatFailureDoesNotFailWhileNonCooperativeHandlerRuns(t *testing.T) {
	queue := &heartbeatOrderingQueueForTest{
		MemoryQueue: NewMemoryQueue(QueueConfig{}),
		interval:    2 * time.Millisecond,
	}
	dispatcher := NewDispatcher(queue, DispatcherConfig{
		LaneConcurrency: map[Lane]int{LaneSubtask: 1},
		PollInterval:    2 * time.Millisecond,
	})
	started := make(chan struct{})
	release := make(chan struct{})
	if err := dispatcher.RegisterHandler(LaneSubtask, func(context.Context, Job) error {
		close(started)
		<-release
		queue.handlerDone.Store(true)
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.Run(ctx)
	if err := dispatcher.Enqueue(ctx, Job{ID: "heartbeat-noncooperative-1", Lane: LaneSubtask, Attempt: 1}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for non-cooperative handler")
	}
	deadline := time.Now().Add(2 * time.Second)
	for dispatcher.Metrics()[LaneSubtask].HeartbeatFailed == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if dispatcher.Metrics()[LaneSubtask].HeartbeatFailed == 0 {
		t.Fatal("timed out waiting for heartbeat failure")
	}
	if queue.failCalls.Load() != 0 {
		t.Fatal("dispatcher must not fail/requeue while stale handler is still running")
	}
	close(release)
	deadline = time.Now().Add(2 * time.Second)
	for queue.failCalls.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if queue.failCalls.Load() == 0 || queue.failBeforeDone.Load() {
		t.Fatalf("unexpected fail ordering calls=%d fail_before_done=%t", queue.failCalls.Load(), queue.failBeforeDone.Load())
	}
	metrics := dispatcher.Metrics()[LaneSubtask]
	if metrics.StaleHandlerCompleted != 1 || metrics.LeaseCancellationObserved != 0 {
		t.Fatalf("expected stale completion to be visible, got %#v", metrics)
	}
	cancel()
	dispatcher.Wait()
}
