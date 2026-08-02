package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerOnce sync.Once
	requests     *prometheus.CounterVec
	retries      *prometheus.CounterVec
	durations    *prometheus.HistogramVec
)

func ensureInit() {
	registerOnce.Do(func() {
		requests = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "ocrx",
			Name:      "provider_requests_total",
			Help:      "Total provider calls",
		}, []string{"kind", "cached", "success", "error"})
		retries = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "ocrx",
			Name:      "provider_retries_total",
			Help:      "Total retry attempts performed",
		}, []string{"kind"})
		durations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "ocrx",
			Name:      "provider_duration_seconds",
			Help:      "Duration of provider calls",
			Buckets:   prometheus.DefBuckets,
		}, []string{"kind", "cached", "success", "error"})

		prometheus.MustRegister(requests, retries, durations)
	})
}

// ObserveRequest 记录一次 provider 调用的统计信息。
func ObserveRequest(kind string, cached bool, success bool, retryCount int, duration time.Duration, errCategory string) {
	ensureInit()
	cachedLabel := fmt.Sprintf("%t", cached)
	successLabel := fmt.Sprintf("%t", success)
	requests.WithLabelValues(kind, cachedLabel, successLabel, errCategory).Inc()
	durations.WithLabelValues(kind, cachedLabel, successLabel, errCategory).Observe(duration.Seconds())
	if retryCount > 0 {
		retries.WithLabelValues(kind).Add(float64(retryCount))
	}
}
