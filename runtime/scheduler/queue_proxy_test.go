package scheduler

import (
	"context"
	"testing"
	"time"
)

type queueProxyTargetStub struct {
	acked          []Result
	dequeueJobs    []Job
	failed         []Result
	heartbeatJobs  []Job
	results        map[string]Result
	pending        map[string]bool
	heartbeatEvery time.Duration
	runtimeVisible bool
}

func (s *queueProxyTargetStub) Enqueue(_ context.Context, job Job) error {
	if s.pending == nil {
		s.pending = map[string]bool{}
	}
	s.pending[job.ID] = true
	return nil
}

func (s *queueProxyTargetStub) Dequeue(_ context.Context, lane Lane) (Job, error) {
	if len(s.dequeueJobs) > 0 {
		job := s.dequeueJobs[0]
		s.dequeueJobs = s.dequeueJobs[1:]
		if job.Lane == "" {
			job.Lane = lane
		}
		return job, nil
	}
	for jobID, pending := range s.pending {
		if pending {
			s.pending[jobID] = false
			return Job{ID: jobID, Lane: lane}, nil
		}
	}
	return Job{}, ErrQueueEmpty
}

func (s *queueProxyTargetStub) Ack(_ context.Context, result Result) error {
	s.acked = append(s.acked, result)
	if s.results == nil {
		s.results = map[string]Result{}
	}
	s.results[result.JobID] = Result{
		JobID:     result.JobID,
		Lane:      result.Lane,
		Status:    "completed",
		Succeeded: true,
		Attempt:   result.Attempt,
	}
	if s.pending == nil {
		s.pending = map[string]bool{}
	}
	s.pending[result.JobID] = false
	return nil
}

func (s *queueProxyTargetStub) Fail(_ context.Context, result Result) error {
	s.failed = append(s.failed, result)
	if s.results == nil {
		s.results = map[string]Result{}
	}
	s.results[result.JobID] = Result{
		JobID:     result.JobID,
		Lane:      result.Lane,
		Status:    "failed",
		Succeeded: false,
		Attempt:   result.Attempt,
		Error:     result.Error,
	}
	if s.pending == nil {
		s.pending = map[string]bool{}
	}
	s.pending[result.JobID] = false
	return nil
}

func (s *queueProxyTargetStub) Result(_ context.Context, jobID string) (Result, bool, error) {
	if s.results == nil {
		return Result{}, false, nil
	}
	result, ok := s.results[jobID]
	return result, ok, nil
}

func (s *queueProxyTargetStub) Pending(_ context.Context, jobID string) (bool, error) {
	if s.pending == nil {
		return false, nil
	}
	return s.pending[jobID], nil
}

func (s *queueProxyTargetStub) Heartbeat(_ context.Context, job Job) error {
	s.heartbeatJobs = append(s.heartbeatJobs, job)
	return nil
}

func (s *queueProxyTargetStub) HeartbeatInterval() time.Duration {
	return s.heartbeatEvery
}

func (s *queueProxyTargetStub) HasRuntimeVisibility() bool {
	return s.runtimeVisible
}

func TestQueueProxyUnavailable(t *testing.T) {
	proxy := NewQueueProxy(nil)
	ctx := context.Background()
	if err := proxy.Enqueue(ctx, Job{ID: "job-1", Lane: LaneSubtask}); err != ErrQueueUnavailable {
		t.Fatalf("expected queue unavailable, got %v", err)
	}
	if _, err := proxy.Dequeue(ctx, LaneSubtask); err != ErrQueueUnavailable {
		t.Fatalf("expected queue unavailable, got %v", err)
	}
}

func TestQueueProxySwapTarget(t *testing.T) {
	proxy := NewQueueProxy(nil)
	if proxy.Available() {
		t.Fatalf("expected unavailable proxy")
	}
	target := NewMemoryQueue(QueueConfig{})
	proxy.SetTarget(target)
	if !proxy.Available() {
		t.Fatalf("expected available proxy")
	}

	ctx := context.Background()
	if err := proxy.Enqueue(ctx, Job{
		ID:        "job-1",
		Lane:      LaneSubtask,
		SessionID: "session-1",
		Payload:   "{}",
	}); err != nil {
		t.Fatalf("enqueue through proxy: %v", err)
	}
	pending, err := proxy.Pending(ctx, "job-1")
	if err != nil {
		t.Fatalf("pending through proxy: %v", err)
	}
	if !pending {
		t.Fatalf("expected pending job through proxy")
	}
	job, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue through proxy: %v", err)
	}
	if job.ID != "job-1" {
		t.Fatalf("unexpected job through proxy: %#v", job)
	}
}

func TestQueueProxyPassesThroughResultAndHeartbeat(t *testing.T) {
	target := &queueProxyTargetStub{
		results: map[string]Result{
			"job-1": {
				JobID:     "job-1",
				Lane:      LaneSubtask,
				Status:    "completed",
				Succeeded: true,
			},
		},
		pending:        map[string]bool{"job-1": true},
		heartbeatEvery: 3 * time.Second,
		runtimeVisible: true,
	}
	proxy := NewQueueProxy(target)
	ctx := context.Background()

	pending, err := proxy.Pending(ctx, "job-1")
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if !pending {
		t.Fatalf("expected pending job")
	}
	result, ok, err := proxy.Result(ctx, "job-1")
	if err != nil {
		t.Fatalf("result: %v", err)
	}
	if !ok || !result.Succeeded || result.Status != "completed" {
		t.Fatalf("unexpected result passthrough: %#v ok=%t", result, ok)
	}
	if err := proxy.Ack(ctx, Result{JobID: "job-1", Lane: LaneSubtask, Status: "completed", Succeeded: true}); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if err := proxy.Fail(ctx, Result{JobID: "job-2", Lane: LaneSubtask, Status: "failed", Error: "boom"}); err != nil {
		t.Fatalf("fail: %v", err)
	}
	if err := proxy.Heartbeat(ctx, Job{ID: "job-1", Lane: LaneSubtask}); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if len(target.acked) != 1 || target.acked[0].JobID != "job-1" {
		t.Fatalf("expected ack to pass through, got %#v", target.acked)
	}
	if len(target.failed) != 1 || target.failed[0].JobID != "job-2" {
		t.Fatalf("expected fail to pass through, got %#v", target.failed)
	}
	if len(target.heartbeatJobs) != 1 || target.heartbeatJobs[0].ID != "job-1" {
		t.Fatalf("expected heartbeat to pass through, got %#v", target.heartbeatJobs)
	}
	if got := proxy.HeartbeatInterval(); got != 3*time.Second {
		t.Fatalf("expected heartbeat interval passthrough, got %s", got)
	}
	if !proxy.HasRuntimeVisibility() {
		t.Fatalf("expected runtime visibility passthrough")
	}
}

func TestQueueProxySwapTargetPreservesClaimedRoutingWithoutAttempt(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-old", Lane: LaneSubtask}},
		heartbeatEvery: 2 * time.Second,
	}
	newTarget := &queueProxyTargetStub{
		heartbeatEvery: 5 * time.Second,
	}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	job, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue old target: %v", err)
	}
	proxy.SetTarget(newTarget)

	if err := proxy.Heartbeat(ctx, job); err != nil {
		t.Fatalf("heartbeat old claimed job: %v", err)
	}
	if err := proxy.Ack(ctx, Result{JobID: job.ID, Lane: job.Lane}); err != nil {
		t.Fatalf("ack old claimed job: %v", err)
	}
	if len(oldTarget.heartbeatJobs) != 1 || oldTarget.heartbeatJobs[0].ID != "job-old" {
		t.Fatalf("expected heartbeat to stay on old target, got %#v", oldTarget.heartbeatJobs)
	}
	if len(newTarget.heartbeatJobs) != 0 {
		t.Fatalf("expected no heartbeat on new target, got %#v", newTarget.heartbeatJobs)
	}
	if len(oldTarget.acked) != 1 || oldTarget.acked[0].JobID != "job-old" {
		t.Fatalf("expected ack to stay on old target, got %#v", oldTarget.acked)
	}
	if len(newTarget.acked) != 0 {
		t.Fatalf("expected no ack on new target, got %#v", newTarget.acked)
	}
	result, ok, err := proxy.Result(ctx, "job-old")
	if err != nil {
		t.Fatalf("result old claimed job: %v", err)
	}
	if !ok || result.Status != "completed" || !result.Succeeded {
		t.Fatalf("expected terminal result to stay on old target, got %#v ok=%t", result, ok)
	}
}

func TestQueueProxySwapTargetPreservesClaimedRoutingPerAttempt(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-attempt", Lane: LaneSubtask, Attempt: 2}},
		heartbeatEvery: 2 * time.Second,
	}
	newTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-new", Lane: LaneSubtask, Attempt: 1}},
		heartbeatEvery: 5 * time.Second,
	}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	claimedJob, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue claimed attempt: %v", err)
	}
	proxy.SetTarget(newTarget)

	if err := proxy.Heartbeat(ctx, claimedJob); err != nil {
		t.Fatalf("heartbeat claimed attempt: %v", err)
	}
	if err := proxy.Fail(ctx, Result{JobID: claimedJob.ID, Lane: claimedJob.Lane, Attempt: claimedJob.Attempt, Error: "boom"}); err != nil {
		t.Fatalf("fail claimed attempt: %v", err)
	}

	newJob, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue new target after swap: %v", err)
	}
	if err := proxy.Ack(ctx, Result{JobID: newJob.ID, Lane: newJob.Lane, Attempt: newJob.Attempt}); err != nil {
		t.Fatalf("ack new target job: %v", err)
	}

	if len(oldTarget.heartbeatJobs) != 1 || oldTarget.heartbeatJobs[0].Attempt != 2 {
		t.Fatalf("expected attempt heartbeat to stay on old target, got %#v", oldTarget.heartbeatJobs)
	}
	if len(newTarget.heartbeatJobs) != 0 {
		t.Fatalf("expected no attempt heartbeat on new target, got %#v", newTarget.heartbeatJobs)
	}
	if len(oldTarget.failed) != 1 || oldTarget.failed[0].Attempt != 2 || oldTarget.failed[0].JobID != "job-attempt" {
		t.Fatalf("expected fail to stay on old target, got %#v", oldTarget.failed)
	}
	if len(newTarget.failed) != 0 {
		t.Fatalf("expected no fail on new target, got %#v", newTarget.failed)
	}
	if len(newTarget.acked) != 1 || newTarget.acked[0].JobID != "job-new" || newTarget.acked[0].Attempt != 1 {
		t.Fatalf("expected new target ack after swap, got %#v", newTarget.acked)
	}
}

func TestQueueProxySwapTargetPreservesClaimedRoutingPerAttemptWhenResultOmitsAttempt(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-attempt", Lane: LaneSubtask, Attempt: 2}},
		heartbeatEvery: 2 * time.Second,
	}
	newTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-new", Lane: LaneSubtask, Attempt: 1}},
		heartbeatEvery: 5 * time.Second,
	}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	claimedJob, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue claimed attempt: %v", err)
	}
	proxy.SetTarget(newTarget)

	if err := proxy.Fail(ctx, Result{JobID: claimedJob.ID, Lane: claimedJob.Lane, Error: "boom"}); err != nil {
		t.Fatalf("fail claimed attempt without result attempt: %v", err)
	}

	newJob, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue new target after swap: %v", err)
	}
	if err := proxy.Ack(ctx, Result{JobID: newJob.ID, Lane: newJob.Lane}); err != nil {
		t.Fatalf("ack new target job without result attempt: %v", err)
	}

	if len(oldTarget.failed) != 1 || oldTarget.failed[0].JobID != "job-attempt" || oldTarget.failed[0].Attempt != 0 {
		t.Fatalf("expected omitted-attempt fail to stay on old target, got %#v", oldTarget.failed)
	}
	if len(newTarget.failed) != 0 {
		t.Fatalf("expected no fail on new target, got %#v", newTarget.failed)
	}
	if len(newTarget.acked) != 1 || newTarget.acked[0].JobID != "job-new" || newTarget.acked[0].Attempt != 0 {
		t.Fatalf("expected omitted-attempt ack to stay on new target, got %#v", newTarget.acked)
	}
	if len(oldTarget.acked) != 0 {
		t.Fatalf("expected no ack on old target, got %#v", oldTarget.acked)
	}
}

func TestQueueProxyNestedProxyPreservesClaimedRoutingAcrossInnerTargetSwap(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-stacked", Lane: LaneSubtask, Attempt: 2}},
		heartbeatEvery: 2 * time.Second,
		runtimeVisible: true,
	}
	newTarget := &queueProxyTargetStub{
		dequeueJobs:    []Job{{ID: "job-fresh", Lane: LaneSubtask, Attempt: 1}},
		heartbeatEvery: 5 * time.Second,
		runtimeVisible: true,
	}
	inner := NewQueueProxy(oldTarget)
	outer := NewQueueProxy(inner)
	ctx := context.Background()

	claimedJob, err := outer.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue stacked claimed attempt: %v", err)
	}
	inner.SetTarget(newTarget)

	if !outer.HasRuntimeVisibility() {
		t.Fatalf("expected outer proxy to preserve nested runtime visibility")
	}
	if got := outer.HeartbeatInterval(); got != 5*time.Second {
		t.Fatalf("expected outer proxy heartbeat interval to reflect inner current target, got=%s", got)
	}
	if err := outer.Heartbeat(ctx, claimedJob); err != nil {
		t.Fatalf("heartbeat stacked claimed attempt: %v", err)
	}
	if err := outer.Fail(ctx, Result{JobID: claimedJob.ID, Lane: claimedJob.Lane, Error: "boom"}); err != nil {
		t.Fatalf("fail stacked claimed attempt without result attempt: %v", err)
	}

	freshJob, err := outer.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue stacked fresh attempt: %v", err)
	}
	if err := outer.Ack(ctx, Result{JobID: freshJob.ID, Lane: freshJob.Lane}); err != nil {
		t.Fatalf("ack stacked fresh attempt without result attempt: %v", err)
	}

	if len(oldTarget.heartbeatJobs) != 1 || oldTarget.heartbeatJobs[0].ID != "job-stacked" || oldTarget.heartbeatJobs[0].Attempt != 2 {
		t.Fatalf("expected stacked heartbeat to stay on old target, got %#v", oldTarget.heartbeatJobs)
	}
	if len(newTarget.heartbeatJobs) != 0 {
		t.Fatalf("expected no stacked heartbeat on new target, got %#v", newTarget.heartbeatJobs)
	}
	if len(oldTarget.failed) != 1 || oldTarget.failed[0].JobID != "job-stacked" || oldTarget.failed[0].Attempt != 0 {
		t.Fatalf("expected stacked fail to stay on old target, got %#v", oldTarget.failed)
	}
	if len(newTarget.failed) != 0 {
		t.Fatalf("expected no stacked fail on new target, got %#v", newTarget.failed)
	}
	if len(newTarget.acked) != 1 || newTarget.acked[0].JobID != "job-fresh" || newTarget.acked[0].Attempt != 0 {
		t.Fatalf("expected stacked ack to stay on new target, got %#v", newTarget.acked)
	}
	if len(oldTarget.acked) != 0 {
		t.Fatalf("expected no stacked ack on old target, got %#v", oldTarget.acked)
	}
}

func TestQueueProxyNestedProxyReflectsInnerTargetVisibilityAndHeartbeatInterval(t *testing.T) {
	firstTarget := &queueProxyTargetStub{
		heartbeatEvery: 3 * time.Second,
		runtimeVisible: true,
	}
	secondTarget := &queueProxyTargetStub{
		heartbeatEvery: 0,
		runtimeVisible: false,
	}
	inner := NewQueueProxy(firstTarget)
	outer := NewQueueProxy(inner)

	if !outer.HasRuntimeVisibility() {
		t.Fatalf("expected outer proxy to expose initial nested runtime visibility")
	}
	if got := outer.HeartbeatInterval(); got != 3*time.Second {
		t.Fatalf("expected outer proxy to expose initial nested heartbeat interval, got=%s", got)
	}

	inner.SetTarget(secondTarget)

	if outer.HasRuntimeVisibility() {
		t.Fatalf("expected outer proxy visibility to follow inner target swap")
	}
	if got := outer.HeartbeatInterval(); got != 0 {
		t.Fatalf("expected outer proxy heartbeat interval to follow inner target swap, got=%s", got)
	}
}

func TestQueueProxySwapTargetPreservesQueuedPendingVisibility(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		pending: map[string]bool{"job-queued": true},
	}
	newTarget := &queueProxyTargetStub{}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	if err := proxy.Enqueue(ctx, Job{ID: "job-queued", Lane: LaneSubtask}); err != nil {
		t.Fatalf("enqueue queued job: %v", err)
	}
	proxy.SetTarget(newTarget)

	pending, err := proxy.Pending(ctx, "job-queued")
	if err != nil {
		t.Fatalf("pending queued job after swap: %v", err)
	}
	if !pending {
		t.Fatalf("expected queued job to remain pending through old route")
	}
}

func TestQueueProxySwapTargetFallsBackToCurrentTargetWhenRoutedQueueHasNoVisibleState(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs: []Job{{ID: "job-fallback", Lane: LaneSubtask, Attempt: 1}},
	}
	newTarget := &queueProxyTargetStub{
		results: map[string]Result{
			"job-fallback": {
				JobID:     "job-fallback",
				Lane:      LaneSubtask,
				Status:    "completed",
				Succeeded: true,
				Attempt:   1,
			},
		},
		pending: map[string]bool{"job-fallback": true},
	}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	if _, err := proxy.Dequeue(ctx, LaneSubtask); err != nil {
		t.Fatalf("dequeue fallback job: %v", err)
	}
	proxy.SetTarget(newTarget)

	pending, err := proxy.Pending(ctx, "job-fallback")
	if err != nil {
		t.Fatalf("pending fallback job: %v", err)
	}
	if !pending {
		t.Fatalf("expected pending to fall back to current target")
	}
	result, ok, err := proxy.Result(ctx, "job-fallback")
	if err != nil {
		t.Fatalf("result fallback job: %v", err)
	}
	if !ok || !result.Succeeded || result.Status != "completed" {
		t.Fatalf("expected result to fall back to current target, got %#v ok=%t", result, ok)
	}
}

func TestQueueProxyEnqueueClearsStaleRoutedTerminalStateAfterSwap(t *testing.T) {
	oldTarget := &queueProxyTargetStub{
		dequeueJobs: []Job{{ID: "job-reuse", Lane: LaneSubtask}},
	}
	newTarget := &queueProxyTargetStub{}
	proxy := NewQueueProxy(oldTarget)
	ctx := context.Background()

	job, err := proxy.Dequeue(ctx, LaneSubtask)
	if err != nil {
		t.Fatalf("dequeue reusable job: %v", err)
	}
	proxy.SetTarget(newTarget)
	if err := proxy.Ack(ctx, Result{JobID: job.ID, Lane: job.Lane}); err != nil {
		t.Fatalf("ack reusable job on old target: %v", err)
	}

	if err := proxy.Enqueue(ctx, Job{ID: "job-reuse", Lane: LaneSubtask}); err != nil {
		t.Fatalf("reenqueue reusable job on new target: %v", err)
	}
	pending, err := proxy.Pending(ctx, "job-reuse")
	if err != nil {
		t.Fatalf("pending reusable job after reenqueue: %v", err)
	}
	if !pending {
		t.Fatalf("expected reenqueue to clear old terminal route and expose new pending state")
	}
	if result, ok, err := proxy.Result(ctx, "job-reuse"); err != nil {
		t.Fatalf("result reusable job after reenqueue: %v", err)
	} else if ok {
		t.Fatalf("expected no terminal result after fresh reenqueue, got %#v", result)
	}
}
