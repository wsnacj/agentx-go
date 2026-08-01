// Package resourcepolicy provides host-owned narrowing policies for HTTP
// request resources and budgets.
package resourcepolicy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrPathNotAllowed   = errors.New("host resource path is not allowed")
	ErrValueNotAllowed  = errors.New("host resource value is not allowed")
	ErrBudgetNotAllowed = errors.New("host resource budget is not allowed")
)

// PathPolicy allows exact paths and descendants of explicitly allowed roots.
type PathPolicy struct {
	AllowedPaths []string
	AllowedRoots []string
}

// NarrowPositiveInt applies an optional positive request limit without allowing
// it to exceed the host-owned limit. Zero means inherit the host limit.
func NarrowPositiveInt(hostLimit int, requested int) (int, error) {
	if hostLimit <= 0 || requested < 0 || requested > hostLimit {
		return 0, ErrBudgetNotAllowed
	}
	if requested == 0 {
		return hostLimit, nil
	}
	return requested, nil
}

// NarrowDurationMilliseconds applies an optional millisecond request timeout
// without converting an overflowing or host-expanding value to time.Duration.
func NarrowDurationMilliseconds(hostLimit time.Duration, requested int) (time.Duration, error) {
	if hostLimit <= 0 || requested < 0 {
		return 0, ErrBudgetNotAllowed
	}
	if requested == 0 {
		return hostLimit, nil
	}
	maxMilliseconds := hostLimit / time.Millisecond
	if maxMilliseconds <= 0 || int64(requested) > int64(maxMilliseconds) {
		return 0, ErrBudgetNotAllowed
	}
	return time.Duration(requested) * time.Millisecond, nil
}

// NarrowPermission lets a request disable a host permission but never enable a
// permission the host disabled. Nil means inherit the host setting.
func NarrowPermission(hostAllows bool, requested *bool) (bool, error) {
	if requested == nil {
		return hostAllows, nil
	}
	if *requested && !hostAllows {
		return false, ErrBudgetNotAllowed
	}
	return *requested, nil
}

// NarrowRequirement lets a request enable a host requirement but never disable
// a requirement the host enabled. Nil means inherit the host setting.
func NarrowRequirement(hostRequires bool, requested *bool) (bool, error) {
	if requested == nil {
		return hostRequires, nil
	}
	if !*requested && hostRequires {
		return false, ErrBudgetNotAllowed
	}
	return *requested, nil
}

// Resolve returns the canonical allowed request path or the host default.
func (p PathPolicy) Resolve(defaultPath string, requestedPath string) (string, error) {
	requestedPath = strings.TrimSpace(requestedPath)
	if requestedPath == "" {
		return strings.TrimSpace(defaultPath), nil
	}
	requested, err := canonicalPath(requestedPath)
	if err != nil {
		return "", ErrPathNotAllowed
	}
	allowedPaths := append([]string{strings.TrimSpace(defaultPath)}, p.AllowedPaths...)
	for _, allowedPath := range allowedPaths {
		if strings.TrimSpace(allowedPath) == "" {
			continue
		}
		allowed, err := canonicalPath(allowedPath)
		if err == nil && requested == allowed {
			return requested, nil
		}
	}
	for _, allowedRoot := range p.AllowedRoots {
		if strings.TrimSpace(allowedRoot) == "" {
			continue
		}
		root, err := canonicalPath(allowedRoot)
		if err != nil {
			continue
		}
		relative, err := filepath.Rel(root, requested)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return requested, nil
		}
	}
	return "", ErrPathNotAllowed
}

// ValuePolicy allows exact opaque values such as host-owned database DSNs.
type ValuePolicy struct {
	AllowedValues []string
}

// Resolve returns an allowed request value or the host default.
func (p ValuePolicy) Resolve(defaultValue string, requestedValue string) (string, error) {
	requestedValue = strings.TrimSpace(requestedValue)
	if requestedValue == "" {
		return strings.TrimSpace(defaultValue), nil
	}
	allowedValues := append([]string{strings.TrimSpace(defaultValue)}, p.AllowedValues...)
	for _, allowed := range allowedValues {
		if requestedValue == strings.TrimSpace(allowed) && requestedValue != "" {
			return requestedValue, nil
		}
	}
	return "", ErrValueNotAllowed
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	absolute = filepath.Clean(absolute)
	cursor := absolute
	missing := make([]string, 0, 2)
	for {
		resolved, err := filepath.EvalSymlinks(cursor)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			return "", err
		}
		missing = append(missing, filepath.Base(cursor))
		cursor = parent
	}
}
