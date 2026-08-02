package scheduler

import (
	"sync"
	"sync/atomic"
)

type LaneMetrics struct {
	Enqueued                  uint64
	Dequeued                  uint64
	Acked                     uint64
	Failed                    uint64
	Requeued                  uint64
	DeadLetter                uint64
	HeartbeatFailed           uint64
	LeaseCancellationObserved uint64
	StaleHandlerCompleted     uint64
	AckFailed                 uint64
	FailFailed                uint64
}

type Metrics struct {
	main       laneCounter
	subtask    laneCounter
	background laneCounter
}

type laneCounter struct {
	enqueued                  uint64
	dequeued                  uint64
	acked                     uint64
	failed                    uint64
	requeued                  uint64
	deadLetter                uint64
	heartbeatFailed           uint64
	leaseCancellationObserved uint64
	staleHandlerCompleted     uint64
	ackFailed                 uint64
	failFailed                uint64
}

func (m *Metrics) RecordEnqueue(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.enqueued, 1)
}

func (m *Metrics) RecordDequeue(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.dequeued, 1)
}

func (m *Metrics) RecordAck(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.acked, 1)
}

func (m *Metrics) RecordFail(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.failed, 1)
}

func (m *Metrics) RecordRequeue(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.requeued, 1)
}

func (m *Metrics) RecordDeadLetter(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.deadLetter, 1)
}

func (m *Metrics) RecordHeartbeatFailed(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.heartbeatFailed, 1)
}

func (m *Metrics) RecordLeaseCancellationObserved(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.leaseCancellationObserved, 1)
}

func (m *Metrics) RecordStaleHandlerCompletion(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.staleHandlerCompleted, 1)
}

func (m *Metrics) RecordAckFailed(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.ackFailed, 1)
}

func (m *Metrics) RecordFailFailed(lane Lane) {
	counter := m.counter(lane)
	atomic.AddUint64(&counter.failFailed, 1)
}

func (m *Metrics) Snapshot() map[Lane]LaneMetrics {
	out := map[Lane]LaneMetrics{
		LaneMain:       m.snapshotLane(LaneMain),
		LaneSubtask:    m.snapshotLane(LaneSubtask),
		LaneBackground: m.snapshotLane(LaneBackground),
	}
	return out
}

func (m *Metrics) snapshotLane(lane Lane) LaneMetrics {
	counter := m.counter(lane)
	return LaneMetrics{
		Enqueued:                  atomic.LoadUint64(&counter.enqueued),
		Dequeued:                  atomic.LoadUint64(&counter.dequeued),
		Acked:                     atomic.LoadUint64(&counter.acked),
		Failed:                    atomic.LoadUint64(&counter.failed),
		Requeued:                  atomic.LoadUint64(&counter.requeued),
		DeadLetter:                atomic.LoadUint64(&counter.deadLetter),
		HeartbeatFailed:           atomic.LoadUint64(&counter.heartbeatFailed),
		LeaseCancellationObserved: atomic.LoadUint64(&counter.leaseCancellationObserved),
		StaleHandlerCompleted:     atomic.LoadUint64(&counter.staleHandlerCompleted),
		AckFailed:                 atomic.LoadUint64(&counter.ackFailed),
		FailFailed:                atomic.LoadUint64(&counter.failFailed),
	}
}

func (m *Metrics) counter(lane Lane) *laneCounter {
	switch lane {
	case LaneSubtask:
		return &m.subtask
	case LaneBackground:
		return &m.background
	default:
		return &m.main
	}
}

type handlerRegistry struct {
	mu       sync.RWMutex
	handlers map[Lane]Handler
}

func newHandlerRegistry() *handlerRegistry {
	return &handlerRegistry{
		handlers: map[Lane]Handler{},
	}
}

func (r *handlerRegistry) Set(lane Lane, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[lane] = handler
}

func (r *handlerRegistry) Get(lane Lane) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[lane]
	return handler, ok
}
