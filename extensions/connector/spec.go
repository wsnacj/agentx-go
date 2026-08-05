package connector

import (
	"fmt"
	"strings"

	capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"
)

// Normalize validates and canonicalizes a connector declaration without
// mutating caller-owned state.
func Normalize(raw Spec) (Spec, error) {
	raw.ID = normalizeID(raw.ID)
	raw.Name = strings.TrimSpace(raw.Name)
	raw.Description = strings.TrimSpace(raw.Description)
	raw.Version = strings.TrimSpace(raw.Version)
	raw.Protocol = Protocol(strings.ToLower(strings.TrimSpace(string(raw.Protocol))))
	raw.Transport = Transport(strings.ToLower(strings.TrimSpace(string(raw.Transport))))
	if raw.ID == "" || raw.Protocol != ProtocolMCP || !validTransport(raw.Transport) {
		return Spec{}, &Error{Code: ErrorCodeInvalidSpec, Cause: fmt.Errorf("unsupported identity, protocol, or transport")}
	}
	if raw.Name == "" {
		raw.Name = raw.ID
	}
	return raw, nil
}

// Project creates one discovery-only catalog asset. It does not connect,
// authorize, or activate the connector.
func Project(sourceRef string, spec Spec) (capabilitycatalog.Asset, error) {
	normalized, err := Normalize(spec)
	if err != nil {
		return capabilitycatalog.Asset{}, err
	}
	return capabilitycatalog.Asset{
		Identity:    capabilitycatalog.Identity{Kind: capabilitycatalog.KindConnector, ID: normalized.ID},
		Name:        normalized.Name,
		Description: normalized.Description,
		Version:     normalized.Version,
		SourceRef:   strings.TrimSpace(sourceRef),
		Tags:        []string{string(normalized.Protocol), string(normalized.Transport)},
	}, nil
}

func validTransport(value Transport) bool {
	return value == TransportStdio || value == TransportStreamableHTTP
}

func normalizeID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || len(value) > 128 {
		return ""
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return ""
		}
	}
	return value
}
