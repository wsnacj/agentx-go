// Package usage defines provider-neutral usage collection seams.
package usage

import (
	"context"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

// Collector receives usage records after a host-selected provider call.
type Collector interface {
	Record(context.Context, llm.UsageRecord) error
}

// NoopCollector discards usage records.
type NoopCollector struct{}

// Record implements Collector.
func (NoopCollector) Record(context.Context, llm.UsageRecord) error { return nil }
