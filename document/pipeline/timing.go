package pipeline

import "time"

// The legacy pipeline emitted wall-clock log messages through a package global.
// Stage diagnostics retain the observable durations; these helpers intentionally
// avoid introducing a global logger into the canonical implementation.
func startTiming(string) time.Time { return time.Now() }
func endTiming(time.Time, string)  {}
