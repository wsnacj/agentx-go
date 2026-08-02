package publicnews

import (
	"strings"
	"time"
)

const defaultRecentNewsWindowDays = 30

// LatestNewsPublishedAfter returns the explicit or resolved lower publication
// bound from the structured intent.
func LatestNewsPublishedAfter(intent LatestNewsLookupIntent) string {
	return LatestNewsPublishedAfterAt(intent, time.Now())
}

// LatestNewsPublishedAfterAt resolves common relative windows against a
// host-controlled clock when no concrete cutoff was supplied.
func LatestNewsPublishedAfterAt(intent LatestNewsLookupIntent, now time.Time) string {
	for _, key := range []string{"published_after", "since", "start_time"} {
		if value := strings.TrimSpace(StringArg(intent.Freshness[key])); value != "" {
			return value
		}
	}
	days := latestNewsRelativeWindowDays(intent)
	if days == 0 || now.IsZero() {
		return ""
	}
	location := now.Location()
	cutoffDate := now.In(location).AddDate(0, 0, -days)
	cutoff := time.Date(cutoffDate.Year(), cutoffDate.Month(), cutoffDate.Day(), 0, 0, 0, 0, location)
	return cutoff.Format(time.RFC3339)
}

func latestNewsRelativeWindowDays(intent LatestNewsLookupIntent) int {
	values := []string{}
	for _, key := range []string{"relative_date_hint", "time_range", "recency"} {
		values = append(values, StringArg(intent.Freshness[key]))
	}
	values = append(values, intent.UserMessage, intent.OriginalIntent)
	for _, raw := range values {
		value := strings.ToLower(strings.TrimSpace(raw))
		switch {
		case containsAny(value, "最近一周", "过去一周", "近一周", "一星期", "7天", "7 天", "last week", "past week", "last 7", "past 7"):
			return 7
		case containsAny(value, "最近一天", "过去一天", "近一天", "24小时", "24 小时", "今天", "今日", "last day", "past day", "last 24", "past 24"):
			return 1
		case containsAny(value, "最近一个月", "过去一个月", "近一个月", "近一月", "30天", "30 天", "last month", "past month", "last 30", "past 30"):
			return 30
		}
	}
	for _, key := range []string{"relative_date_hint", "time_range", "recency"} {
		switch strings.ToLower(strings.TrimSpace(StringArg(intent.Freshness[key]))) {
		case "最近", "近期", "recent", "recently":
			return defaultRecentNewsWindowDays
		}
	}
	return 0
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

// LatestNewsSourceWithinFreshnessWindow checks an explicit or recognized
// relative publication cutoff. Without either, existing ranking remains.
func LatestNewsSourceWithinFreshnessWindow(source LatestNewsLookupSource, intent LatestNewsLookupIntent) bool {
	return LatestNewsSourceWithinFreshnessWindowAt(source, intent, time.Now())
}

// LatestNewsSourceWithinFreshnessWindowAt is the deterministic-clock variant
// used by live adapters and tests.
func LatestNewsSourceWithinFreshnessWindowAt(source LatestNewsLookupSource, intent LatestNewsLookupIntent, now time.Time) bool {
	cutoff := LatestNewsPublishedAfterAt(intent, now)
	return cutoff == "" || LatestNewsFreshnessConfirmed(source.PublishedAt, cutoff)
}
