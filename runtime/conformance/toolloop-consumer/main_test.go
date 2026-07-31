package main

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

func TestFixedVersionConsumerRunsPortableLoop(t *testing.T) {
	result, err := run(context.Background(), 3)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if result.Kind != toolloop.OutcomeCompleted || result.Round != 3 {
		t.Fatalf("result = %#v", result)
	}
}

func TestFixedVersionConsumerUsesDetectorAndFuse(t *testing.T) {
	detector := toolloop.NewLoopDetector(toolloop.LoopDetectorConfig{
		Enabled:             true,
		RepeatThreshold:     2,
		PingPongThreshold:   4,
		NoProgressThreshold: 2,
	})
	calls := []toolloop.Call{{Name: "lookup", Arguments: `{"id":"same"}`}}
	runs := []toolloop.RunObservation{{Name: "lookup", Output: "same"}}
	_, _ = detector.Observe(1, calls, runs)
	signal, ok := detector.Observe(2, calls, runs)
	if !ok || signal.Kind != toolloop.LoopKindNoProgress {
		t.Fatalf("signal = %#v ok=%t", signal, ok)
	}

	fuse := toolloop.NewFailureFuse(toolloop.FailureFuseConfig{
		Enabled:   true,
		Threshold: 2,
	})
	observations := []toolloop.FailureObservation{{
		Tool:       "lookup",
		Failed:     true,
		ErrorClass: "timeout",
	}}
	_, _ = fuse.Observe(1, observations)
	failure, ok := fuse.Observe(2, observations)
	if !ok || failure.Tool != "lookup" || failure.Count != 2 {
		t.Fatalf("failure = %#v ok=%t", failure, ok)
	}
}
