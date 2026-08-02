package tools

import (
	"net"
	"net/url"
	"strings"
)

const (
	browserRouteFallbackRemoteLocalURLReason = "remote_browser_local_url_fallback"
)

type browserRemoteTargetURLGuardProvider interface {
	BrowserRemoteTargetURLGuardEnabled() bool
}

func browserBackendRemoteTargetURLGuardEnabled(backend BrowserBackend) bool {
	provider, ok := backend.(browserRemoteTargetURLGuardProvider)
	return ok && provider.BrowserRemoteTargetURLGuardEnabled()
}

func browserURLLooksPrivateOrLocal(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return false
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && isPrivateOrLocalIP(ip)
}
