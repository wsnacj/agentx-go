package tools

import "fmt"

func browserCompatRegisteredOrEnabledToolForActKind(ctx browserRegistrationContext, kind string) string {
	name := browserCompatToolForManagedOptInActKind(kind)
	if name == "" || !browserRuntimeRegisteredOrEnabledTool(ctx, name) {
		return ""
	}
	return name
}

func browserCompatToolErrorf(kind string, format string, args ...any) error {
	prefix := browserCompatToolForManagedOptInActKind(kind)
	if prefix == "" {
		prefix = "browser"
	}
	all := make([]any, 0, len(args)+1)
	all = append(all, prefix)
	all = append(all, args...)
	return fmt.Errorf("%s: "+format, all...)
}
