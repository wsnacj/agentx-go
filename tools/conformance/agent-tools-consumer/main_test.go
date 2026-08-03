package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	if err := run(context.Background()); err != nil {
		t.Fatal(err)
	}
}
