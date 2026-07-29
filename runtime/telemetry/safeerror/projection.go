// Package safeerror projects errors into observation-safe classifications.
package safeerror

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const maxIdentifierChars = 64

type Projection struct {
	Class    string `json:"class,omitempty"`
	Code     string `json:"code,omitempty"`
	Identity string `json:"identity,omitempty"`
}

type wrappedError struct {
	message  string
	cause    error
	identity string
}

func (e *wrappedError) Error() string {
	if e == nil {
		return "safe error"
	}
	if message := strings.TrimSpace(e.message); message != "" {
		return message
	}
	return "safe error"
}

func (e *wrappedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func Wrap(cause error, message string) error {
	return WrapWithIdentity(cause, message, "")
}

func WrapWithIdentity(cause error, message string, identity string) error {
	if cause == nil {
		return nil
	}
	if identity = strings.TrimSpace(identity); identity == "" {
		identity = Identity(errorIdentityMaterial(cause))
	}
	return &wrappedError{
		message:  strings.TrimSpace(message),
		cause:    cause,
		identity: identity,
	}
}

func Project(err error, class string, code string) Projection {
	return ProjectWithIdentity(err, class, code, "")
}

func ProjectWithIdentity(err error, class string, code string, identity string) Projection {
	projection := Projection{
		Class:    normalizeIdentifier(class, "error"),
		Code:     normalizeIdentifier(code, "unknown"),
		Identity: strings.TrimSpace(identity),
	}
	if err == nil {
		return projection
	}
	var wrapped *wrappedError
	if projection.Identity == "" && errors.As(err, &wrapped) && wrapped != nil {
		projection.Identity = wrapped.identity
	}
	if projection.Identity == "" {
		projection.Identity = Identity(errorIdentityMaterial(err))
	}
	return projection
}

func ProjectText(value string, class string, code string) Projection {
	projection := Projection{
		Class: normalizeIdentifier(class, "error"),
		Code:  normalizeIdentifier(code, "unknown"),
	}
	if value = strings.TrimSpace(value); value != "" {
		projection.Identity = Identity(value)
	}
	return projection
}

func Identity(material string) string {
	material = strings.TrimSpace(material)
	if material == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])
}

func Summary(projection Projection) string {
	parts := make([]string, 0, 3)
	if projection.Class != "" {
		parts = append(parts, "class="+projection.Class)
	}
	if projection.Code != "" {
		parts = append(parts, "code="+projection.Code)
	}
	if projection.Identity != "" {
		parts = append(parts, "identity="+projection.Identity)
	}
	if len(parts) == 0 {
		return "class=error code=unknown"
	}
	return strings.Join(parts, " ")
}

func AppendAttrs(attrs map[string]any, prefix string, projection Projection) map[string]any {
	if attrs == nil {
		attrs = map[string]any{}
	}
	prefix = strings.TrimSpace(prefix)
	if projection.Class != "" {
		attrs[prefix+"error_class"] = projection.Class
	}
	if projection.Code != "" {
		attrs[prefix+"error_code"] = projection.Code
	}
	if projection.Identity != "" {
		attrs[prefix+"error_identity"] = projection.Identity
	}
	return attrs
}

func AppendDetails(details map[string]string, prefix string, projection Projection) map[string]string {
	if details == nil {
		details = map[string]string{}
	}
	prefix = strings.TrimSpace(prefix)
	if projection.Class != "" {
		details[prefix+"error_class"] = projection.Class
	}
	if projection.Code != "" {
		details[prefix+"error_code"] = projection.Code
	}
	if projection.Identity != "" {
		details[prefix+"error_identity"] = projection.Identity
	}
	return details
}

func errorIdentityMaterial(err error) string {
	parts := make([]string, 0, 4)
	for current := err; current != nil; current = errors.Unwrap(current) {
		parts = append(parts, fmt.Sprintf("%T:%s", current, current.Error()))
	}
	return strings.Join(parts, "\n")
}

func normalizeIdentifier(value string, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastUnderscore := false
	for i := 0; i < len(value) && builder.Len() < maxIdentifierChars; i++ {
		char := value[i]
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '.' {
			builder.WriteByte(char)
			lastUnderscore = false
		} else if builder.Len() > 0 && !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return fallback
	}
	return result
}
