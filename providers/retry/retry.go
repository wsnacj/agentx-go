// Package retry provides bounded, context-aware retry for provider calls.
package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/wsnacj/agentx-go/providers/fault"
)

type Options struct {
	MaxRetries        int
	AttemptTimeout    time.Duration
	TotalTimeout      time.Duration
	BackoffBase       time.Duration
	BackoffMultiplier float64
	BackoffJitter     float64
	MaxBackoff        time.Duration
	Classify          func(error) bool
	OnAttempt         func(AttemptInfo)
}
type AttemptInfo struct {
	Attempt        int
	Err            error
	Delay          time.Duration
	Classification fault.Classification
}
type AttemptSummary struct {
	AttemptCount        int
	DelayedAttemptCount int
	Last                fault.Classification
	Faults              fault.Summary
}

func Default() *Options {
	return &Options{MaxRetries: 3, AttemptTimeout: 90 * time.Second, BackoffBase: 300 * time.Millisecond, BackoffMultiplier: 2, BackoffJitter: .2, MaxBackoff: 30 * time.Second, Classify: DefaultClassifier}
}
func DefaultClassifier(err error) bool { return fault.IsRetryable(err) }
func SummarizeAttempts(attempts []AttemptInfo) AttemptSummary {
	if len(attempts) == 0 {
		return AttemptSummary{}
	}
	items := make([]fault.Classification, 0, len(attempts))
	summary := AttemptSummary{AttemptCount: len(attempts)}
	for _, attempt := range attempts {
		classification := attempt.Classification
		if classification == (fault.Classification{}) && attempt.Err != nil {
			classification = fault.Classify(attempt.Err)
		}
		if classification != (fault.Classification{}) {
			items = append(items, classification)
			summary.Last = classification
		}
		if attempt.Delay > 0 {
			summary.DelayedAttemptCount++
		}
	}
	summary.Faults = fault.Summarize(items)
	return summary
}

func backoff(base time.Duration, multiplier, jitter float64, attempt int, max time.Duration) time.Duration {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if multiplier < 1 {
		multiplier = 1
	}
	wait := time.Duration(float64(base) * math.Pow(multiplier, float64(attempt)))
	if jitter > 0 {
		factor := 1 + (rand.Float64()*2-1)*jitter
		if factor < 0 {
			factor = 0
		}
		wait = time.Duration(float64(wait) * factor)
	}
	if max > 0 && wait > max {
		wait = max
	}
	return wait
}

// Do invokes fn until it succeeds, becomes non-retryable, or exhausts its bounds.
func Do(ctx context.Context, opts *Options, fn func(context.Context) error) error {
	if opts == nil {
		opts = Default()
	}
	if opts.TotalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.TotalTimeout)
		defer cancel()
	}
	var lastErr error
	for i := 0; i < opts.MaxRetries+1; i++ {
		attemptCtx := ctx
		var cancel context.CancelFunc
		if opts.AttemptTimeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, opts.AttemptTimeout)
		}
		err := fn(attemptCtx)
		if cancel != nil {
			cancel()
		}
		if err == nil {
			return nil
		}
		lastErr = err
		classification := fault.Classify(err)
		retryable := true
		if opts.Classify != nil {
			retryable = opts.Classify(err)
		}
		if !retryable || i == opts.MaxRetries {
			if opts.OnAttempt != nil {
				opts.OnAttempt(AttemptInfo{Attempt: i + 1, Err: err, Classification: classification})
			}
			break
		}
		delay := backoff(opts.BackoffBase, opts.BackoffMultiplier, opts.BackoffJitter, i, opts.MaxBackoff)
		if opts.OnAttempt != nil {
			opts.OnAttempt(AttemptInfo{Attempt: i + 1, Err: err, Delay: delay, Classification: classification})
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if lastErr == nil {
		return fmt.Errorf("unknown error")
	}
	return lastErr
}
