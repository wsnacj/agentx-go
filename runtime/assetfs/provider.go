// Package assetfs exposes immutable, read-only asset providers for resources
// that are compiled into an AgentX binary.
package assetfs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

const fingerprintVersion = "agentx-assetfs-v1"

// Provider identifies an immutable, read-only filesystem and its content.
//
// New snapshots the supplied filesystem and computes its content fingerprint.
// Later changes to the input filesystem cannot affect the provider.
type Provider struct {
	id          string
	fingerprint string
	source      fs.FS
}

// New constructs a provider and computes a deterministic content fingerprint.
func New(id string, source fs.FS) (Provider, error) {
	id = strings.TrimSpace(id)
	if err := validateID(id); err != nil {
		return Provider{}, err
	}
	if source == nil {
		return Provider{}, fmt.Errorf("assetfs provider %q requires a filesystem", id)
	}
	snapshot, err := captureSnapshot(source)
	if err != nil {
		return Provider{}, fmt.Errorf("snapshot assetfs provider %q: %w", id, err)
	}
	fingerprint, err := fingerprintFS(snapshot)
	if err != nil {
		return Provider{}, fmt.Errorf("fingerprint assetfs provider %q: %w", id, err)
	}
	return Provider{id: id, fingerprint: fingerprint, source: snapshot}, nil
}

// MustNew is New with panic-on-error semantics for package-level embedded
// providers whose filesystem is fixed at build time.
func MustNew(id string, source fs.FS) Provider {
	provider, err := New(id, source)
	if err != nil {
		panic(err)
	}
	return provider
}

// IsZero reports whether the provider has not been initialized.
func (p Provider) IsZero() bool {
	return p.source == nil || strings.TrimSpace(p.id) == "" || strings.TrimSpace(p.fingerprint) == ""
}

// ID returns the stable logical identity of the provider.
func (p Provider) ID() string {
	return p.id
}

// Fingerprint returns a deterministic SHA-256 digest of paths and file content.
func (p Provider) Fingerprint() string {
	return p.fingerprint
}

// FS returns a read-only filesystem whose root is ".".
func (p Provider) FS() fs.FS {
	if p.IsZero() {
		return nil
	}
	return p
}

// Open implements fs.FS without exposing the provider's backing snapshot.
func (p Provider) Open(name string) (fs.File, error) {
	if p.IsZero() {
		return nil, invalidProviderPathError("open", name)
	}
	return p.source.Open(name)
}

// ReadFile reads a provider-relative file and returns a detached byte slice.
func (p Provider) ReadFile(name string) ([]byte, error) {
	if p.IsZero() {
		return nil, invalidProviderPathError("read", name)
	}
	content, err := fs.ReadFile(p.source, name)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), content...), nil
}

// ReadDir reads a provider-relative directory.
func (p Provider) ReadDir(name string) ([]fs.DirEntry, error) {
	if p.IsZero() {
		return nil, invalidProviderPathError("readdir", name)
	}
	entries, err := fs.ReadDir(p.source, name)
	if err != nil {
		return nil, err
	}
	return append([]fs.DirEntry(nil), entries...), nil
}

// Sub returns an independently identified provider rooted at dir.
func (p Provider) Sub(dir string) (Provider, error) {
	if p.IsZero() {
		return Provider{}, invalidProviderPathError("sub", dir)
	}
	dir = path.Clean(strings.TrimSpace(dir))
	if !fs.ValidPath(dir) {
		return Provider{}, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid}
	}
	if dir == "." {
		return p, nil
	}
	childID := p.id + "/" + dir
	if err := validateID(childID); err != nil {
		return Provider{}, err
	}
	source, err := fs.Sub(p.source, dir)
	if err != nil {
		return Provider{}, err
	}
	fingerprint, err := fingerprintFS(source)
	if err != nil {
		return Provider{}, fmt.Errorf("fingerprint assetfs provider %q subdirectory %q: %w", p.id, dir, err)
	}
	return Provider{
		id:          childID,
		fingerprint: fingerprint,
		source:      source,
	}, nil
}

// IsProviderFS verifies that source is the read-only filesystem view of the
// provider identity supplied alongside it.
func IsProviderFS(source fs.FS, id string, fingerprint string) bool {
	var provider Provider
	switch value := source.(type) {
	case Provider:
		provider = value
	case *Provider:
		if value == nil {
			return false
		}
		provider = *value
	default:
		return false
	}
	return !provider.IsZero() &&
		provider.ID() == strings.TrimSpace(id) &&
		provider.Fingerprint() == strings.TrimSpace(fingerprint)
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("assetfs provider id is required")
	}
	if strings.ContainsAny(id, "?#\\") || strings.ContainsFunc(id, func(r rune) bool {
		return r <= ' ' || r == '\u007f'
	}) {
		return fmt.Errorf("assetfs provider id %q contains unsupported characters", id)
	}
	cleaned := path.Clean(id)
	if cleaned != id || id == "." || strings.HasPrefix(id, "/") || strings.HasPrefix(id, "../") {
		return fmt.Errorf("assetfs provider id %q is not canonical", id)
	}
	return nil
}

func fingerprintFS(source fs.FS) (string, error) {
	hasher := sha256.New()
	writeFingerprintField(hasher, fingerprintVersion)
	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry == nil {
			return &fs.PathError{Op: "fingerprint", Path: name, Err: fs.ErrInvalid}
		}
		writeFingerprintField(hasher, name)
		if entry.IsDir() {
			writeFingerprintField(hasher, "dir")
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return &fs.PathError{Op: "fingerprint", Path: name, Err: fs.ErrInvalid}
		}
		content, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		writeFingerprintField(hasher, "file")
		writeFingerprintField(hasher, string(content))
		return nil
	})
	if err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func writeFingerprintField(hasher interface{ Write([]byte) (int, error) }, value string) {
	_, _ = hasher.Write([]byte(strconv.Itoa(len(value))))
	_, _ = hasher.Write([]byte{':'})
	_, _ = hasher.Write([]byte(value))
}

func invalidProviderPathError(op string, name string) error {
	return &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
}
