package scheduler

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type QueueConfig struct {
	LaneQueueLimit map[Lane]int
	ResultLimit    int
	ResultTTL      time.Duration
}

type MemoryQueue struct {
	mu             sync.Mutex
	lanes          map[Lane][]Job
	heads          map[Lane]int
	limits         map[Lane]int
	results        map[string]Result
	active         map[string]Job
	resultLimit    int
	resultTTL      time.Duration
	evictedResults uint64
}

func NewMemoryQueue(cfg QueueConfig) *MemoryQueue {
	limits := map[Lane]int{
		LaneMain:       128,
		LaneSubtask:    256,
		LaneBackground: 256,
	}
	for lane, value := range cfg.LaneQueueLimit {
		if value > 0 {
			limits[lane] = value
		}
	}
	resultLimit := cfg.ResultLimit
	if resultLimit <= 0 {
		resultLimit = 2048
	}
	resultTTL := cfg.ResultTTL
	if resultTTL <= 0 {
		resultTTL = 30 * time.Minute
	}
	return &MemoryQueue{
		lanes: map[Lane][]Job{
			LaneMain:       nil,
			LaneSubtask:    nil,
			LaneBackground: nil,
		},
		heads: map[Lane]int{
			LaneMain:       0,
			LaneSubtask:    0,
			LaneBackground: 0,
		},
		limits:      limits,
		results:     map[string]Result{},
		active:      map[string]Job{},
		resultLimit: resultLimit,
		resultTTL:   resultTTL,
	}
}

func (q *MemoryQueue) HasRuntimeVisibility() bool {
	return false
}

func (q *MemoryQueue) Enqueue(_ context.Context, job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	if !isKnownLane(job.Lane) {
		return ErrInvalidLane
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.ID = strings.TrimSpace(job.ID)
	if job.ID == "" {
		return ErrUnknownJob
	}
	if q.terminalResultLocked(job.ID) {
		q.dropPendingJobLocked(job.ID)
		delete(q.results, job.ID)
		delete(q.active, job.ID)
	}
	if q.pendingLocked(job.ID) {
		return nil
	}
	// Re-enqueue with the same job ID should represent a fresh attempt.
	delete(q.results, job.ID)
	limit := q.limits[job.Lane]
	if limit > 0 && q.lanePendingLenLocked(job.Lane) >= limit {
		return ErrQueueLimit
	}
	q.lanes[job.Lane] = append(q.lanes[job.Lane], job)
	return nil
}

func (q *MemoryQueue) Dequeue(_ context.Context, lane Lane) (Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	if !isKnownLane(lane) {
		return Job{}, ErrInvalidLane
	}
	for {
		job, ok := q.popLaneJobLocked(lane)
		if !ok {
			return Job{}, ErrQueueEmpty
		}
		if _, terminal := q.results[strings.TrimSpace(job.ID)]; terminal {
			continue
		}
		q.active[strings.TrimSpace(job.ID)] = job
		return job, nil
	}
}

func (q *MemoryQueue) DequeueByKind(_ context.Context, lane Lane, jobKind string) (Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	if !isKnownLane(lane) {
		return Job{}, ErrInvalidLane
	}
	jobKind = strings.TrimSpace(jobKind)
	if jobKind == "" {
		return Job{}, ErrUnknownJob
	}
	start := q.heads[lane]
	if start < 0 {
		start = 0
	}
	queue := q.lanes[lane]
	if start >= len(queue) {
		q.lanes[lane] = nil
		q.heads[lane] = 0
		return Job{}, ErrQueueEmpty
	}
	for index := start; index < len(queue); {
		job := queue[index]
		if q.terminalResultLocked(strings.TrimSpace(job.ID)) {
			if index == start {
				queue[index] = Job{}
				q.heads[lane] = start + 1
				start++
				index++
				continue
			}
			copy(queue[index:], queue[index+1:])
			queue[len(queue)-1] = Job{}
			queue = queue[:len(queue)-1]
			q.lanes[lane] = queue
			continue
		}
		if strings.TrimSpace(job.JobKind) != jobKind {
			index++
			continue
		}
		queue[index] = Job{}
		if index == start {
			q.heads[lane] = start + 1
			q.maybeCompactLaneLocked(lane)
			q.active[strings.TrimSpace(job.ID)] = job
			return job, nil
		}
		copy(queue[index:], queue[index+1:])
		queue[len(queue)-1] = Job{}
		q.lanes[lane] = queue[:len(queue)-1]
		q.maybeCompactLaneLocked(lane)
		q.active[strings.TrimSpace(job.ID)] = job
		return job, nil
	}
	q.maybeCompactLaneLocked(lane)
	return Job{}, ErrQueueEmpty
}

func (q *MemoryQueue) Ack(_ context.Context, result Result) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	q.pruneResultCacheLocked(now)
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return ErrUnknownJob
	}
	delete(q.active, jobID)
	if existing, exists := q.results[jobID]; exists && !existing.Succeeded {
		return nil
	}
	result.JobID = jobID
	result.Outcome = ResultOutcomeCompleted
	result.Status = string(result.Outcome)
	result.Succeeded = true
	if result.FinishedAt.IsZero() {
		result.FinishedAt = now
	}
	q.results[jobID] = result
	q.pruneResultCacheLocked(now)
	return nil
}

func (q *MemoryQueue) Fail(_ context.Context, result Result) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	q.pruneResultCacheLocked(now)
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return ErrUnknownJob
	}
	delete(q.active, jobID)
	if existing, ok := q.results[jobID]; ok && (normalizeResultOutcome(existing.Outcome) == ResultOutcomeCanceled || strings.EqualFold(strings.TrimSpace(existing.Status), string(ResultOutcomeCanceled))) {
		return nil
	}
	result.JobID = jobID
	result.Outcome = resultOutcome(result, ResultOutcomeFailed)
	result.Status = string(result.Outcome)
	result.Succeeded = false
	if result.FinishedAt.IsZero() {
		result.FinishedAt = now
	}
	q.results[jobID] = result
	q.pruneResultCacheLocked(now)
	return nil
}

func (q *MemoryQueue) Cancel(_ context.Context, result Result) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	q.pruneResultCacheLocked(now)
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return ErrUnknownJob
	}
	if existing, ok := q.results[jobID]; ok {
		if normalizeResultOutcome(existing.Outcome) == ResultOutcomeCanceled || strings.EqualFold(strings.TrimSpace(existing.Status), string(ResultOutcomeCanceled)) {
			return nil
		}
		return ErrUnknownJob
	}
	pending := q.pendingLocked(jobID)
	active, running := q.active[jobID]
	if !pending && !running {
		return ErrUnknownJob
	}
	if running && result.Attempt > 0 && active.Attempt > 0 && result.Attempt != active.Attempt {
		return ErrLeaseLost
	}
	if result.Attempt <= 0 && running {
		result.Attempt = active.Attempt
	}
	result.JobID = jobID
	result.Outcome = ResultOutcomeCanceled
	result.Status = string(ResultOutcomeCanceled)
	result.Succeeded = false
	if result.FinishedAt.IsZero() {
		result.FinishedAt = now
	}
	q.dropPendingJobLocked(jobID)
	delete(q.active, jobID)
	q.results[jobID] = result
	q.pruneResultCacheLocked(now)
	return nil
}

func (q *MemoryQueue) Result(_ context.Context, jobID string) (Result, bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	result, ok := q.results[strings.TrimSpace(jobID)]
	return result, ok, nil
}

func (q *MemoryQueue) Pending(_ context.Context, jobID string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	target := strings.TrimSpace(jobID)
	if target == "" {
		return false, ErrUnknownJob
	}
	if q.terminalResultLocked(target) {
		return false, nil
	}
	return q.pendingLocked(target), nil
}

func (q *MemoryQueue) ResultStats() (size int, evicted uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneResultCacheLocked(time.Now())
	return len(q.results), atomic.LoadUint64(&q.evictedResults)
}

func (q *MemoryQueue) pendingLocked(jobID string) bool {
	target := strings.TrimSpace(jobID)
	if target == "" {
		return false
	}
	for _, lane := range []Lane{LaneMain, LaneSubtask, LaneBackground} {
		queue := q.lanes[lane]
		start := q.heads[lane]
		if start < 0 {
			start = 0
		}
		if start >= len(queue) {
			continue
		}
		for _, job := range queue[start:] {
			if strings.EqualFold(strings.TrimSpace(job.ID), target) {
				return true
			}
		}
	}
	return false
}

func (q *MemoryQueue) terminalResultLocked(jobID string) bool {
	target := strings.TrimSpace(jobID)
	if target == "" {
		return false
	}
	result, ok := q.results[target]
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(result.Status)) {
	case "completed", "failed", "canceled", "dead_letter":
		return true
	default:
		return false
	}
}

func (q *MemoryQueue) dropPendingJobLocked(jobID string) {
	target := strings.TrimSpace(jobID)
	if target == "" {
		return
	}
	for _, lane := range []Lane{LaneMain, LaneSubtask, LaneBackground} {
		queue := q.lanes[lane]
		start := q.heads[lane]
		if start < 0 {
			start = 0
		}
		if start >= len(queue) {
			q.lanes[lane] = nil
			q.heads[lane] = 0
			continue
		}
		filtered := queue[:start]
		for _, job := range queue[start:] {
			if strings.EqualFold(strings.TrimSpace(job.ID), target) {
				continue
			}
			filtered = append(filtered, job)
		}
		q.lanes[lane] = filtered
		q.maybeCompactLaneLocked(lane)
	}
}

func (q *MemoryQueue) lanePendingLenLocked(lane Lane) int {
	queue := q.lanes[lane]
	start := q.heads[lane]
	if start <= 0 {
		return len(queue)
	}
	if start >= len(queue) {
		return 0
	}
	return len(queue) - start
}

func (q *MemoryQueue) popLaneJobLocked(lane Lane) (Job, bool) {
	queue := q.lanes[lane]
	start := q.heads[lane]
	if start < 0 {
		start = 0
	}
	if start >= len(queue) {
		q.lanes[lane] = nil
		q.heads[lane] = 0
		return Job{}, false
	}
	job := queue[start]
	// Break references from consumed slot for better GC behavior.
	queue[start] = Job{}
	start++
	q.heads[lane] = start
	q.maybeCompactLaneLocked(lane)
	return job, true
}

func (q *MemoryQueue) maybeCompactLaneLocked(lane Lane) {
	queue := q.lanes[lane]
	start := q.heads[lane]
	if start <= 0 {
		return
	}
	if start >= len(queue) {
		q.lanes[lane] = nil
		q.heads[lane] = 0
		return
	}
	// Compact only when enough head slack accumulates to avoid frequent copies.
	if start >= 64 && start*2 >= len(queue) {
		compacted := append([]Job(nil), queue[start:]...)
		q.lanes[lane] = compacted
		q.heads[lane] = 0
	}
}

func (q *MemoryQueue) pruneResultCacheLocked(now time.Time) {
	if len(q.results) == 0 {
		return
	}
	if q.resultTTL > 0 {
		for jobID, result := range q.results {
			finished := result.FinishedAt
			if finished.IsZero() || now.Sub(finished) < q.resultTTL {
				continue
			}
			delete(q.results, jobID)
			atomic.AddUint64(&q.evictedResults, 1)
		}
	}
	if q.resultLimit <= 0 {
		return
	}
	for len(q.results) > q.resultLimit {
		oldestID := ""
		var oldestAt time.Time
		for jobID, result := range q.results {
			finished := result.FinishedAt
			if oldestID == "" || finished.Before(oldestAt) || (finished.Equal(oldestAt) && jobID < oldestID) {
				oldestID = jobID
				oldestAt = finished
			}
		}
		if oldestID == "" {
			return
		}
		delete(q.results, oldestID)
		atomic.AddUint64(&q.evictedResults, 1)
	}
}

func isKnownLane(lane Lane) bool {
	switch lane {
	case LaneMain, LaneSubtask, LaneBackground:
		return true
	default:
		return false
	}
}
