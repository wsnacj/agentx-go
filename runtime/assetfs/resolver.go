package assetfs

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
)

const uriScheme = "assetfs://"

// ErrResolverSealed reports an attempt to add an asset provider after startup
// wiring has been frozen.
var ErrResolverSealed = errors.New("assetfs resolver is sealed")

// Resolver maps assetfs:// URIs to immutable providers. Providers may be added
// only during startup wiring; Seal makes the resolver safe to share across
// concurrent runs and isolated child runtimes.
type Resolver struct {
	mu        sync.RWMutex
	providers map[string]Provider
	sealed    bool
}

// NewResolver creates an empty provider resolver.
func NewResolver() *Resolver {
	return &Resolver{providers: map[string]Provider{}}
}

// Add registers an immutable provider by its stable ID. Re-registering the
// same ID and fingerprint is idempotent; conflicting content fails closed.
func (r *Resolver) Add(provider Provider) error {
	return r.AddAll(provider)
}

// AddAll atomically registers immutable providers. No provider is added when
// validation, identity collision, or sealed-state checks fail.
func (r *Resolver) AddAll(providers ...Provider) error {
	if r == nil {
		return fmt.Errorf("assetfs resolver is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		return ErrResolverSealed
	}
	pending := make(map[string]Provider, len(providers))
	for _, provider := range providers {
		if provider.IsZero() {
			return fmt.Errorf("assetfs resolver requires a non-zero provider")
		}
		if existing, ok := r.providers[provider.ID()]; ok {
			if existing.Fingerprint() != provider.Fingerprint() {
				return fmt.Errorf("assetfs provider %q is already registered with different content", provider.ID())
			}
			continue
		}
		if existing, ok := pending[provider.ID()]; ok {
			if existing.Fingerprint() != provider.Fingerprint() {
				return fmt.Errorf("assetfs provider %q is repeated with different content", provider.ID())
			}
			continue
		}
		pending[provider.ID()] = provider
	}
	for id, provider := range pending {
		r.providers[id] = provider
	}
	return nil
}

// Seal freezes provider registration. It is idempotent.
func (r *Resolver) Seal() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.sealed = true
	r.mu.Unlock()
}

// IsSealed reports whether provider registration has been frozen.
func (r *Resolver) IsSealed() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sealed
}

// Open resolves and opens an immutable asset URI. The URI form is
// assetfs://<provider-id>/<provider-relative-path>.
func (r *Resolver) Open(uri string) (fs.File, error) {
	provider, name, err := r.resolve(uri)
	if err != nil {
		return nil, err
	}
	file, err := provider.Open(name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// CanOpen reports whether uri resolves to a regular file without returning an
// open handle.
func (r *Resolver) CanOpen(uri string) bool {
	file, err := r.Open(uri)
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	return err == nil && info.Mode().IsRegular()
}

func (r *Resolver) resolve(uri string) (Provider, string, error) {
	if r == nil {
		return Provider{}, "", &fs.PathError{Op: "resolve", Path: uri, Err: fs.ErrNotExist}
	}
	raw := strings.TrimSpace(uri)
	if len(raw) < len(uriScheme) || !strings.EqualFold(raw[:len(uriScheme)], uriScheme) {
		return Provider{}, "", &fs.PathError{Op: "resolve", Path: uri, Err: fs.ErrInvalid}
	}
	remainder := raw[len(uriScheme):]
	if remainder == "" || strings.ContainsAny(remainder, "?#\\") {
		return Provider{}, "", &fs.PathError{Op: "resolve", Path: uri, Err: fs.ErrInvalid}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var (
		matched Provider
		matchID string
	)
	for id, provider := range r.providers {
		if remainder != id && !strings.HasPrefix(remainder, id+"/") {
			continue
		}
		if len(id) > len(matchID) {
			matchID = id
			matched = provider
		}
	}
	if matched.IsZero() {
		return Provider{}, "", &fs.PathError{Op: "resolve", Path: uri, Err: fs.ErrNotExist}
	}
	rawName := strings.TrimPrefix(remainder, matchID)
	rawName = strings.TrimPrefix(rawName, "/")
	name := path.Clean(rawName)
	if rawName != name || name == "." || !fs.ValidPath(name) {
		return Provider{}, "", &fs.PathError{Op: "resolve", Path: uri, Err: fs.ErrInvalid}
	}
	return matched, name, nil
}
