package assetfs

import (
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestResolverOpensLongestMatchingProviderAndSeals(t *testing.T) {
	parent := MustNew("scene.demo", fstest.MapFS{
		"skills/parent.txt": {Data: []byte("parent")},
	})
	child := MustNew("scene.demo/skills", fstest.MapFS{
		"child.txt": {Data: []byte("child")},
	})
	resolver := NewResolver()
	if err := resolver.Add(parent); err != nil {
		t.Fatalf("add parent: %v", err)
	}
	if err := resolver.Add(child); err != nil {
		t.Fatalf("add child: %v", err)
	}
	if err := resolver.Add(child); err != nil {
		t.Fatalf("idempotent add: %v", err)
	}

	file, err := resolver.Open("assetfs://scene.demo/skills/child.txt")
	if err != nil {
		t.Fatalf("open child: %v", err)
	}
	content, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil || string(content) != "child" {
		t.Fatalf("read child: content=%q err=%v", content, err)
	}
	if !resolver.CanOpen("assetfs://scene.demo/skills/child.txt") {
		t.Fatal("expected regular child asset to be readable")
	}
	if resolver.CanOpen("assetfs://scene.demo/skills") {
		t.Fatal("provider root is not a readable regular file")
	}

	resolver.Seal()
	if !resolver.IsSealed() {
		t.Fatal("expected resolver to be sealed")
	}
	if err := resolver.Add(MustNew("other", fstest.MapFS{"file": {Data: []byte("x")}})); !errors.Is(err, ErrResolverSealed) {
		t.Fatalf("sealed add error=%v, want ErrResolverSealed", err)
	}
}

func TestResolverRejectsUnknownAndInvalidURIs(t *testing.T) {
	resolver := NewResolver()
	if err := resolver.Add(MustNew("known", fstest.MapFS{"file.txt": {Data: []byte("ok")}})); err != nil {
		t.Fatalf("add provider: %v", err)
	}
	for _, uri := range []string{
		"",
		"/tmp/file.txt",
		"assetfs://",
		"assetfs://unknown/file.txt",
		"assetfs://known",
		"assetfs://known/../file.txt",
		"assetfs://known/file.txt?query=1",
	} {
		if _, err := resolver.Open(uri); err == nil {
			t.Fatalf("expected %q to fail", uri)
		} else {
			var pathErr *fs.PathError
			if !errors.As(err, &pathErr) {
				t.Fatalf("expected PathError for %q, got %T: %v", uri, err, err)
			}
		}
	}
}

func TestResolverRejectsProviderIdentityCollision(t *testing.T) {
	resolver := NewResolver()
	if err := resolver.Add(MustNew("same", fstest.MapFS{"file": {Data: []byte("one")}})); err != nil {
		t.Fatalf("add first provider: %v", err)
	}
	if err := resolver.Add(MustNew("same", fstest.MapFS{"file": {Data: []byte("two")}})); err == nil {
		t.Fatal("expected provider identity collision to fail")
	}
}

func TestResolverAddAllIsAtomicOnIdentityCollision(t *testing.T) {
	resolver := NewResolver()
	if err := resolver.Add(MustNew("existing", fstest.MapFS{"file": {Data: []byte("one")}})); err != nil {
		t.Fatal(err)
	}
	err := resolver.AddAll(
		MustNew("pending", fstest.MapFS{"file": {Data: []byte("pending")}}),
		MustNew("existing", fstest.MapFS{"file": {Data: []byte("different")}}),
	)
	if err == nil {
		t.Fatal("expected batch collision")
	}
	if resolver.CanOpen("assetfs://pending/file") {
		t.Fatal("failed batch must not partially register providers")
	}
	if !resolver.CanOpen("assetfs://existing/file") {
		t.Fatal("existing provider must remain available")
	}
}
