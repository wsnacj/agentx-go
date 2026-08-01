package resourcepolicy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathPolicyAllowsDefaultExactAndConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	defaultPath := filepath.Join(root, "default.sqlite")
	resolved, err := (PathPolicy{}).Resolve(defaultPath, defaultPath)
	if err != nil || resolved == "" {
		t.Fatalf("resolve default: path=%q err=%v", resolved, err)
	}
	allowed := filepath.Join(root, "artifacts", "output.json")
	resolved, err = (PathPolicy{AllowedRoots: []string{root}}).Resolve("", allowed)
	if err != nil || resolved == "" {
		t.Fatalf("resolve allowed root: path=%q err=%v", resolved, err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if _, err := (PathPolicy{AllowedRoots: []string{root}}).Resolve("", outside); !errors.Is(err, ErrPathNotAllowed) {
		t.Fatalf("outside path error = %v, want not allowed", err)
	}
}

func TestHostBudgetsOnlyNarrow(t *testing.T) {
	if got, err := NarrowPositiveInt(50, 10); err != nil || got != 10 {
		t.Fatalf("narrow integer budget: got=%d err=%v", got, err)
	}
	if got, err := NarrowPositiveInt(50, 0); err != nil || got != 50 {
		t.Fatalf("inherit integer budget: got=%d err=%v", got, err)
	}
	for _, requested := range []int{-1, 51} {
		if _, err := NarrowPositiveInt(50, requested); !errors.Is(err, ErrBudgetNotAllowed) {
			t.Fatalf("integer request %d should be rejected, got %v", requested, err)
		}
	}
	if got, err := NarrowDurationMilliseconds(30*time.Second, 500); err != nil || got != 500*time.Millisecond {
		t.Fatalf("narrow duration budget: got=%s err=%v", got, err)
	}
	if _, err := NarrowDurationMilliseconds(30*time.Second, 30_001); !errors.Is(err, ErrBudgetNotAllowed) {
		t.Fatalf("expanded duration should be rejected, got %v", err)
	}
}

func TestHostBooleanPoliciesOnlyNarrow(t *testing.T) {
	trueValue := true
	falseValue := false
	if _, err := NarrowPermission(false, &trueValue); !errors.Is(err, ErrBudgetNotAllowed) {
		t.Fatalf("disabled host permission must not be enabled, got %v", err)
	}
	if got, err := NarrowPermission(true, &falseValue); err != nil || got {
		t.Fatalf("host permission should allow request narrowing: got=%t err=%v", got, err)
	}
	if _, err := NarrowRequirement(true, &falseValue); !errors.Is(err, ErrBudgetNotAllowed) {
		t.Fatalf("host requirement must not be disabled, got %v", err)
	}
	if got, err := NarrowRequirement(false, &trueValue); err != nil || !got {
		t.Fatalf("request should be allowed to add a requirement: got=%t err=%v", got, err)
	}
}

func TestPathPolicyRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := (PathPolicy{AllowedRoots: []string{root}}).Resolve("", filepath.Join(link, "secret")); !errors.Is(err, ErrPathNotAllowed) {
		t.Fatalf("symlink escape error = %v, want not allowed", err)
	}
}

func TestValuePolicyAllowsOnlyHostValues(t *testing.T) {
	policy := ValuePolicy{AllowedValues: []string{"postgres://allowed-two"}}
	if got, err := policy.Resolve("postgres://allowed-one", "postgres://allowed-two"); err != nil || got != "postgres://allowed-two" {
		t.Fatalf("allowed value = %q err=%v", got, err)
	}
	if _, err := policy.Resolve("postgres://allowed-one", "postgres://attacker"); !errors.Is(err, ErrValueNotAllowed) {
		t.Fatalf("unexpected value error = %v, want not allowed", err)
	}
}
