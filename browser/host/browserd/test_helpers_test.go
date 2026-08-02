package browserd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func configureIsolatedBrowserdUserCache(t *testing.T) string {
	t.Helper()
	homeRoot := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", homeRoot)
	if os.PathSeparator == '\\' {
		t.Setenv("LOCALAPPDATA", filepath.Join(homeRoot, "AppData", "Local"))
	} else {
		t.Setenv("XDG_CACHE_HOME", filepath.Join(homeRoot, ".cache"))
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("resolve isolated user cache dir: %v", err)
	}
	return cacheRoot
}

func writeFakeBundledPlaywrightChromiumExecutable(t *testing.T, cacheRoot string, revision string) string {
	t.Helper()
	candidates := bundledPlaywrightChromiumExecutableCandidates(cacheRoot, "chromium-"+strings.TrimSpace(revision))
	if len(candidates) == 0 {
		t.Fatalf("expected playwright chromium executable candidate for revision %q", revision)
	}
	executablePath := candidates[0]
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatalf("mkdir fake playwright executable dir: %v", err)
	}
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake playwright executable: %v", err)
	}
	return executablePath
}
