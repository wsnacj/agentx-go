package main

import "testing"

func TestFixedVersionSkillsConsumer(t *testing.T) {
	result, err := run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result != "agentx-skills-ok:portable-review:bundled:true:true:fork:2" {
		t.Fatalf("unexpected result %q", result)
	}
}
