package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// QueueProxy allows swapping the concrete queue implementation at runtime.
// This keeps runner/tools wiring stable while enabling backend replacement.
type QueueProxy struct {
	mu     sync.RWMutex
	target Queue
	claims map[string]Queue
	routes map[string]Queue
}

func NewQueueProxy(target Queue) *QueueProxy {
	return &QueueProxy{target: target}
}

func (p *QueueProxy) SetTarget(target Queue) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.target = target
	p.mu.Unlock()
}

func (p *QueueProxy) Available() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.target != nil
}

func (p *QueueProxy) Enqueue(ctx context.Context, job Job) error {
	target, err := p.targetQueue()
	if err != nil {
		return err
	}
	if err := target.Enqueue(ctx, job); err != nil {
		return err
	}
	p.resetJobRouting(job.ID)
	p.recordJobRoute(job.ID, target)
	return nil
}

func (p *QueueProxy) Dequeue(ctx context.Context, lane Lane) (Job, error) {
	target, err := p.targetQueue()
	if err != nil {
		return Job{}, err
	}
	job, err := target.Dequeue(ctx, lane)
	if err != nil {
		return Job{}, err
	}
	p.recordClaim(job, target)
	p.recordJobRoute(job.ID, target)
	return job, nil
}

func (p *QueueProxy) DequeueByKind(ctx context.Context, lane Lane, jobKind string) (Job, error) {
	target, err := p.targetQueue()
	if err != nil {
		return Job{}, err
	}
	kindAware, ok := target.(KindAwareQueue)
	if !ok || kindAware == nil {
		return Job{}, ErrQueueUnavailable
	}
	job, err := kindAware.DequeueByKind(ctx, lane, jobKind)
	if err != nil {
		return Job{}, err
	}
	p.recordClaim(job, target)
	p.recordJobRoute(job.ID, target)
	return job, nil
}

func (p *QueueProxy) Ack(ctx context.Context, result Result) error {
	target, err := p.targetQueueForResult(result)
	if err != nil {
		return err
	}
	if err := target.Ack(ctx, result); err != nil {
		return err
	}
	p.recordJobRoute(result.JobID, target)
	p.releaseClaim(result.JobID, result.Attempt, target)
	return nil
}

func (p *QueueProxy) Fail(ctx context.Context, result Result) error {
	target, err := p.targetQueueForResult(result)
	if err != nil {
		return err
	}
	if err := target.Fail(ctx, result); err != nil {
		return err
	}
	p.recordJobRoute(result.JobID, target)
	p.releaseClaim(result.JobID, result.Attempt, target)
	return nil
}

func (p *QueueProxy) Cancel(ctx context.Context, result Result) error {
	target, err := p.targetQueueForCancellation(result)
	if err != nil {
		return err
	}
	cancelable, ok := target.(CancelableQueue)
	if !ok || cancelable == nil {
		return ErrCancellationUnsupported
	}
	if err := cancelable.Cancel(ctx, result); err != nil {
		return err
	}
	p.recordJobRoute(result.JobID, target)
	p.releaseClaim(result.JobID, result.Attempt, target)
	return nil
}

func (p *QueueProxy) Result(ctx context.Context, jobID string) (Result, bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return Result{}, false, ErrUnknownJob
	}
	if target := p.jobRoute(jobID); target != nil {
		result, ok, err := target.Result(ctx, jobID)
		if err != nil {
			return Result{}, false, err
		}
		if ok {
			return result, true, nil
		}
		if current := p.currentTarget(); current == nil || current == target {
			return Result{}, false, nil
		}
	}
	target, err := p.targetQueue()
	if err != nil {
		return Result{}, false, err
	}
	return target.Result(ctx, jobID)
}

func (p *QueueProxy) Pending(ctx context.Context, jobID string) (bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, ErrUnknownJob
	}
	if target := p.jobRoute(jobID); target != nil {
		pending, err := target.Pending(ctx, jobID)
		if err != nil {
			return false, err
		}
		if pending {
			return true, nil
		}
		if current := p.currentTarget(); current == nil || current == target {
			return false, nil
		}
	}
	target, err := p.targetQueue()
	if err != nil {
		return false, err
	}
	return target.Pending(ctx, jobID)
}

func (p *QueueProxy) HasRuntimeVisibility() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	target := p.target
	p.mu.RUnlock()
	if target == nil {
		return false
	}
	runtimeVisible, ok := target.(RuntimeVisibleQueue)
	if !ok {
		return false
	}
	return runtimeVisible.HasRuntimeVisibility()
}

func (p *QueueProxy) Heartbeat(ctx context.Context, job Job) error {
	target, err := p.targetQueueForJob(job)
	if err != nil {
		return err
	}
	heartbeatQueue, ok := target.(HeartbeatCapableQueue)
	if !ok {
		return nil
	}
	return heartbeatQueue.Heartbeat(ctx, job)
}

func (p *QueueProxy) HeartbeatInterval() time.Duration {
	if p == nil {
		return 0
	}
	p.mu.RLock()
	target := p.target
	p.mu.RUnlock()
	if target == nil {
		return 0
	}
	heartbeatQueue, ok := target.(HeartbeatCapableQueue)
	if !ok {
		return 0
	}
	return heartbeatQueue.HeartbeatInterval()
}

func (p *QueueProxy) LeaseIdentity(job Job) string {
	if p == nil {
		return ""
	}
	target, err := p.targetQueueForJob(job)
	if err != nil {
		return ""
	}
	provider, ok := target.(LeaseIdentityProvider)
	if !ok || provider == nil {
		return ""
	}
	return provider.LeaseIdentity(job)
}

func (p *QueueProxy) targetQueue() (Queue, error) {
	if p == nil {
		return nil, ErrQueueUnavailable
	}
	target := p.currentTarget()
	if target == nil {
		return nil, ErrQueueUnavailable
	}
	return target, nil
}

func (p *QueueProxy) currentTarget() Queue {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	target := p.target
	p.mu.RUnlock()
	return target
}

func (p *QueueProxy) targetQueueForResult(result Result) (Queue, error) {
	if target := p.claimTarget(result.JobID, result.Attempt); target != nil {
		return target, nil
	}
	if result.Attempt <= 0 {
		if _, target, ok := p.singleClaimTarget(result.JobID); ok {
			return target, nil
		}
	}
	return p.targetQueue()
}

func (p *QueueProxy) targetQueueForCancellation(result Result) (Queue, error) {
	if target := p.claimTarget(result.JobID, result.Attempt); target != nil {
		return target, nil
	}
	if result.Attempt <= 0 {
		if _, target, ok := p.singleClaimTarget(result.JobID); ok {
			return target, nil
		}
	}
	if target := p.jobRoute(result.JobID); target != nil {
		return target, nil
	}
	return p.targetQueue()
}

func (p *QueueProxy) targetQueueForJob(job Job) (Queue, error) {
	if target := p.claimTarget(job.ID, job.Attempt); target != nil {
		return target, nil
	}
	if job.Attempt <= 0 {
		if _, target, ok := p.singleClaimTarget(job.ID); ok {
			return target, nil
		}
	}
	return p.targetQueue()
}

func (p *QueueProxy) recordClaim(job Job, target Queue) {
	if p == nil || target == nil {
		return
	}
	claimKey := queueProxyClaimKey(job.ID, job.Attempt)
	if claimKey == "" {
		return
	}
	p.mu.Lock()
	if p.claims == nil {
		p.claims = map[string]Queue{}
	}
	p.claims[claimKey] = target
	p.mu.Unlock()
}

func (p *QueueProxy) recordJobRoute(jobID string, target Queue) {
	jobID = strings.TrimSpace(jobID)
	if p == nil || target == nil || jobID == "" {
		return
	}
	p.mu.Lock()
	if p.routes == nil {
		p.routes = map[string]Queue{}
	}
	p.routes[jobID] = target
	p.mu.Unlock()
}

func (p *QueueProxy) claimTarget(jobID string, attempt int) Queue {
	if p == nil {
		return nil
	}
	claimKey := queueProxyClaimKey(jobID, attempt)
	if claimKey == "" {
		return nil
	}
	p.mu.RLock()
	target := p.claims[claimKey]
	p.mu.RUnlock()
	return target
}

func (p *QueueProxy) jobRoute(jobID string) Queue {
	jobID = strings.TrimSpace(jobID)
	if p == nil || jobID == "" {
		return nil
	}
	p.mu.RLock()
	target := p.routes[jobID]
	p.mu.RUnlock()
	return target
}

func (p *QueueProxy) releaseClaim(jobID string, attempt int, target Queue) {
	if p == nil || target == nil {
		return
	}
	if attempt <= 0 {
		claimKey, claimTarget, ok := p.singleClaimTarget(jobID)
		if !ok || claimTarget != target {
			return
		}
		p.mu.Lock()
		if p.claims != nil && p.claims[claimKey] == target {
			delete(p.claims, claimKey)
		}
		p.mu.Unlock()
		return
	}
	claimKey := queueProxyClaimKey(jobID, attempt)
	if claimKey == "" {
		return
	}
	p.mu.Lock()
	if p.claims != nil && p.claims[claimKey] == target {
		delete(p.claims, claimKey)
	}
	p.mu.Unlock()
}

func (p *QueueProxy) singleClaimTarget(jobID string) (string, Queue, bool) {
	jobID = strings.TrimSpace(jobID)
	if p == nil || jobID == "" {
		return "", nil, false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.claims) == 0 {
		return "", nil, false
	}
	var (
		foundKey    string
		foundTarget Queue
	)
	for claimKey, target := range p.claims {
		if !queueProxyClaimMatchesJob(claimKey, jobID) {
			continue
		}
		// Omitted-attempt routing is only safe when exactly one active claim is
		// present for the job. If multiple attempts are still in-flight, callers
		// must provide Result.Attempt / Job.Attempt to preserve per-attempt routing.
		if foundKey != "" {
			return "", nil, false
		}
		foundKey = claimKey
		foundTarget = target
	}
	if foundKey == "" || foundTarget == nil {
		return "", nil, false
	}
	return foundKey, foundTarget, true
}

func (p *QueueProxy) resetJobRouting(jobID string) {
	jobID = strings.TrimSpace(jobID)
	if p == nil || jobID == "" {
		return
	}
	p.mu.Lock()
	if p.routes != nil {
		delete(p.routes, jobID)
	}
	if p.claims != nil {
		for claimKey := range p.claims {
			if queueProxyClaimMatchesJob(claimKey, jobID) {
				delete(p.claims, claimKey)
			}
		}
	}
	p.mu.Unlock()
}

func queueProxyClaimKey(jobID string, attempt int) string {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ""
	}
	if attempt > 0 {
		return fmt.Sprintf("%s#%d", jobID, attempt)
	}
	return jobID
}

func queueProxyClaimMatchesJob(claimKey string, jobID string) bool {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false
	}
	return claimKey == jobID || strings.HasPrefix(claimKey, jobID+"#")
}
