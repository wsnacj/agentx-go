package tools

import (
	"context"
	"fmt"
	"strings"
)

func browserValidateFinalURL(ctx context.Context, policy outboundNetworkPolicy, finalURL string) (string, error) {
	raw := strings.TrimSpace(finalURL)
	if raw == "" {
		return "", nil
	}
	parsed, err := policy.validateURL(ctx, raw)
	if err != nil {
		return "", fmt.Errorf("final_url %q: %w", raw, err)
	}
	return parsed.String(), nil
}
