package tools

import (
	"strings"
	"sync"
)

type browserRuntimeRouteAssessmentCache struct {
	mu          sync.Mutex
	assessments map[string]browserConcreteRouteAssessment
}

func newBrowserRuntimeRouteAssessmentCache() *browserRuntimeRouteAssessmentCache {
	return &browserRuntimeRouteAssessmentCache{
		assessments: map[string]browserConcreteRouteAssessment{},
	}
}

func (c *browserRuntimeRouteAssessmentCache) Load(profile string, target string) (browserConcreteRouteAssessment, bool) {
	if c == nil {
		return browserConcreteRouteAssessment{}, false
	}
	key := browserRuntimeRouteAssessmentCacheKey(profile, target)
	if key == "" {
		return browserConcreteRouteAssessment{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	assessment, ok := c.assessments[key]
	return assessment, ok
}

func (c *browserRuntimeRouteAssessmentCache) Store(profile string, target string, assessment browserConcreteRouteAssessment) {
	if c == nil {
		return
	}
	key := browserRuntimeRouteAssessmentCacheKey(profile, target)
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.assessments[key] = assessment
}

func browserRuntimeRouteAssessmentCacheKey(profile string, target string) string {
	profile = strings.ToLower(strings.TrimSpace(profile))
	target = strings.ToLower(strings.TrimSpace(target))
	if profile == "" || target == "" {
		return ""
	}
	return target + "\x00" + profile
}
