package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/retry"
)

func TestDoRetriesBoundedRetryableFailures(t *testing.T) {
	attempts := 0
	err := retry.Do(context.Background(), &retry.Options{
		MaxRetries: 2, BackoffBase: time.Millisecond, Classify: retry.DefaultClassifier,
	}, func(context.Context) error {
		attempts++
		if attempts < 3 {
			return &providers.APIError{StatusCode: 503, Body: "unavailable"}
		}
		return nil
	})
	if err != nil || attempts != 3 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestDoStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := retry.Do(ctx, &retry.Options{MaxRetries: 2, BackoffBase: time.Second, Classify: func(error) bool { return true }}, func(context.Context) error { return errors.New("retry") })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}
