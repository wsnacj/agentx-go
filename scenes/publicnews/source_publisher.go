package publicnews

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

var knownPublisherDomainFamilies = map[string]string{
	"sina.cn":     "sina",
	"sina.com":    "sina",
	"sina.com.cn": "sina",
}

// SourcePublisherFamily returns a conservative publisher identity for a URL.
// Registrable domains are distinct by default; known alternate domains owned
// by one publisher share an explicit scene-level family.
func SourcePublisherFamily(sourceURL string) string {
	return PublisherHostFamily(SourceSite(sourceURL))
}

// PublisherHostFamily is the host-oriented form used when an adapter already
// extracted a source site.
func PublisherHostFamily(host string) string {
	host = NormalizeSourceHost(host)
	if host == "" || host == "unknown" {
		return "unknown"
	}
	registrable, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil || strings.TrimSpace(registrable) == "" {
		registrable = host
	}
	registrable = NormalizeSourceHost(registrable)
	if family := knownPublisherDomainFamilies[registrable]; family != "" {
		return family
	}
	return registrable
}
