// Package requestjson provides the shared JSON request boundary for AgentX
// host-deployed HTTP adapters.
package requestjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrBodyTooLarge = errors.New("request JSON body exceeds limit")
	ErrInvalidLimit = errors.New("request JSON body limit must be positive")
	ErrTrailingData = errors.New("request JSON body contains trailing data")
)

// Decode reads at most maxBytes of one JSON value, rejects unknown fields and
// trailing data, and treats an empty body as the zero value of target.
func Decode(body io.Reader, maxBytes int64, target any) error {
	if maxBytes <= 0 {
		return ErrInvalidLimit
	}
	if body == nil {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(body, maxBytes+1))
	if err != nil {
		return fmt.Errorf("read request JSON body: %w", err)
	}
	if int64(len(raw)) > maxBytes {
		return ErrBodyTooLarge
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrTrailingData
		}
		return fmt.Errorf("%w: %v", ErrTrailingData, err)
	}
	return nil
}
