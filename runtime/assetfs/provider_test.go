package assetfs

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestProviderIdentityReadsAndSub(t *testing.T) {
	source := fstest.MapFS{
		"skills/weather/SKILL.md": &fstest.MapFile{Data: []byte("weather")},
		"skills/weather/assets/a": &fstest.MapFile{Data: []byte("asset")},
	}
	provider, err := New("agentx.test-assets", source)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	if provider.IsZero() || provider.ID() != "agentx.test-assets" {
		t.Fatalf("unexpected provider identity: id=%q zero=%v", provider.ID(), provider.IsZero())
	}
	if provider.Fingerprint() == "" {
		t.Fatal("expected content fingerprint")
	}
	if content, err := provider.ReadFile("skills/weather/SKILL.md"); err != nil || string(content) != "weather" {
		t.Fatalf("read provider file: content=%q err=%v", content, err)
	}
	entries, err := provider.ReadDir("skills")
	if err != nil || len(entries) != 1 || entries[0].Name() != "weather" {
		t.Fatalf("read provider dir: entries=%#v err=%v", entries, err)
	}

	skills, err := provider.Sub("skills")
	if err != nil {
		t.Fatalf("sub provider: %v", err)
	}
	if skills.ID() != "agentx.test-assets/skills" {
		t.Fatalf("unexpected sub provider id %q", skills.ID())
	}
	if skills.Fingerprint() == "" || skills.Fingerprint() == provider.Fingerprint() {
		t.Fatalf("sub provider should have its own content fingerprint: parent=%q sub=%q", provider.Fingerprint(), skills.Fingerprint())
	}
	if content, err := fs.ReadFile(skills.FS(), "weather/SKILL.md"); err != nil || string(content) != "weather" {
		t.Fatalf("read sub filesystem: content=%q err=%v", content, err)
	}
}

func TestProviderFingerprintIsContentAddressed(t *testing.T) {
	first, err := New("agentx.first", fstest.MapFS{
		"a.txt": &fstest.MapFile{Data: []byte("same")},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := New("agentx.second", fstest.MapFS{
		"a.txt": &fstest.MapFile{Data: []byte("same")},
	})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := New("agentx.changed", fstest.MapFS{
		"a.txt": &fstest.MapFile{Data: []byte("different")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Fingerprint() != second.Fingerprint() {
		t.Fatalf("provider id must not affect content fingerprint: %q != %q", first.Fingerprint(), second.Fingerprint())
	}
	if first.Fingerprint() == changed.Fingerprint() {
		t.Fatalf("content change must affect fingerprint: %q", first.Fingerprint())
	}
}

func TestProviderSnapshotsMutableInput(t *testing.T) {
	source := fstest.MapFS{
		"skill/SKILL.md": &fstest.MapFile{Data: []byte("original")},
	}
	provider, err := New("agentx.snapshot", source)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := provider.Fingerprint()

	source["skill/SKILL.md"] = &fstest.MapFile{Data: []byte("mutated")}
	source["skill/new.txt"] = &fstest.MapFile{Data: []byte("new")}

	content, err := provider.ReadFile("skill/SKILL.md")
	if err != nil {
		t.Fatalf("read snapshotted file: %v", err)
	}
	if string(content) != "original" {
		t.Fatalf("provider changed with mutable input: %q", content)
	}
	if provider.Fingerprint() != fingerprint {
		t.Fatalf("provider fingerprint changed after input mutation: %q != %q", provider.Fingerprint(), fingerprint)
	}
	if _, err := provider.ReadFile("skill/new.txt"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("new input file must not appear in provider snapshot, got %v", err)
	}
	if !IsProviderFS(provider.FS(), provider.ID(), provider.Fingerprint()) {
		t.Fatal("provider filesystem should attest its exact identity")
	}
	if IsProviderFS(provider.FS(), provider.ID(), "sha256:wrong") {
		t.Fatal("provider filesystem must reject mismatched fingerprint")
	}
}

func TestProviderRejectsInvalidInputsAndZeroReads(t *testing.T) {
	for _, id := range []string{"", "../escape", "contains space", "bad?id"} {
		if _, err := New(id, fstest.MapFS{}); err == nil {
			t.Fatalf("expected invalid id %q to fail", id)
		}
	}
	if _, err := New("agentx.nil", nil); err == nil {
		t.Fatal("expected nil filesystem to fail")
	}

	var zero Provider
	if !zero.IsZero() {
		t.Fatal("zero provider should report zero")
	}
	if _, err := zero.ReadFile("x"); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("zero provider read should fail with fs.ErrInvalid, got %v", err)
	}
	if _, err := zero.ReadDir("."); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("zero provider readdir should fail with fs.ErrInvalid, got %v", err)
	}
	if _, err := zero.Sub("."); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("zero provider sub should fail with fs.ErrInvalid, got %v", err)
	}
}
