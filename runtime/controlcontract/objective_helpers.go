package controlcontract

func normalizeOneDisplaySafeRef(value DisplaySafeRef) DisplaySafeRef {
	ref, _ := NormalizeDisplaySafeRef(string(value))
	return ref
}

func firstDisplaySafeRef(values ...DisplaySafeRef) DisplaySafeRef {
	for _, value := range values {
		if ref := normalizeOneDisplaySafeRef(value); ref != "" {
			return ref
		}
	}
	return ""
}

func maxNonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeControlTokenList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		token := normalizeControlToken(value)
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func appendUniqueControlToken(in []string, value string) []string {
	return normalizeControlTokenList(append(cloneStringSlice(in), value))
}
