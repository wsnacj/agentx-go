package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	got, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := "agentx-sourceacquisition-ok:public-source-readonly-pack:wechat-article-readonly-pack:1:1"
	if got != want {
		t.Fatalf("output=%q want=%q", got, want)
	}
}
