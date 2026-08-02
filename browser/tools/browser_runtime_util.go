package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func runBrowserCommand(ctx context.Context, runner BrowserCommandRunner, name string, args []string) ([]byte, error) {
	if runner == nil {
		return nil, fmt.Errorf("browser command runner is required")
	}
	return runner(ctx, name, append([]string(nil), args...))
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func firstBool(params map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := params[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			switch strings.ToLower(strings.TrimSpace(typed)) {
			case "true", "1", "yes", "on":
				return true
			case "false", "0", "no", "off":
				return false
			}
		}
	}
	return false
}
