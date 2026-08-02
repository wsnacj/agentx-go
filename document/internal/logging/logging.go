// Package logger provides a module-local logging bridge without importing a
// product host. Applications may configure zap's global logger explicitly.
package logger

import "go.uber.org/zap"

func DebugWithFields(message string, fields ...zap.Field) { zap.L().Debug(message, fields...) }
func InfoWithFields(message string, fields ...zap.Field)  { zap.L().Info(message, fields...) }
func WarnWithFields(message string, fields ...zap.Field)  { zap.L().Warn(message, fields...) }
func ErrorWithFields(message string, fields ...zap.Field) { zap.L().Error(message, fields...) }
