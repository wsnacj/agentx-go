package browserruntime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserDoctorSearchRootsUsesExplicitRootThenCurrentDirectory(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	current := t.TempDir()
	if err := os.Chdir(current); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	current, err = os.Getwd()
	if err != nil {
		t.Fatalf("resolve changed working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	explicit := filepath.Join(t.TempDir(), "operator-root")
	got := browserDoctorSearchRoots(explicit)
	if len(got) != 2 || got[0] != filepath.Clean(explicit) || got[1] != filepath.Clean(current) {
		t.Fatalf("unexpected browser doctor roots: %#v", got)
	}
	if deduped := browserDoctorSearchRoots(current); len(deduped) != 1 || deduped[0] != filepath.Clean(current) {
		t.Fatalf("expected current directory to be deduplicated, got %#v", deduped)
	}
}
