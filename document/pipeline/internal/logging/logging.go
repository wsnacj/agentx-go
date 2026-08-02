// Package logging contains intentionally silent logging hooks for deterministic
// document helpers. Runtime-level observability is exposed by pipeline.Observer;
// low-level helpers must not discover or configure a host logger.
package logging

func Debug(string, ...any)               {}
func Info(string, ...any)                {}
func Warn(string, ...any)                {}
func LogChunks(string, []string)         {}
func LogChunksDetailed(string, []string) {}
