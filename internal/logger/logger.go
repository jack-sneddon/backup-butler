// Package logger provides a logging interface using Uber's zap logger
//
// Log Levels (lowest to highest):
// - DEBUG: Verbose information for debugging issues
// - INFO:  General operational events
// - WARN:  Potentially harmful situations
// - ERROR: Error events that might still allow the application to continue
// - FATAL: Very severe error events that will lead to application termination
//
// Command Line Usage:
//
//	# Default level (error)
//	backup-butler sync -c config.yaml
//
//	# Debug level for verbose output
//	backup-butler sync -c config.yaml --log-level debug
//
//	# Info level for operational events
//	backup-butler sync -c config.yaml --log-level info
//
// Code Examples:
//
//	log := logger.Get()
//
//	logger.Debug("Starting function", "param1", value1)
//	logger.Info("Operation complete", "files", fileCount)
//	logger.Warn("Retrying operation", "attempt", retryCount)
//	logger.Error("Operation failed", "error", err)
//	log.Fatalw("Unrecoverable error", "error", err)
//
// Structured Logging:
//
//	logger.Info("message",
//	  "key1", value1,
//	  "key2", value2)
//
// internal/logger/logger.go
// internal/logger/logger.go
package logger

import (
	"log/slog"
	"os"
	"sync"
)

// Log levels
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

var (
	defaultLogger *slog.Logger
	logLevel      = new(slog.LevelVar)
	once          sync.Once
)

// Init initializes the default logger with custom options
func Init() error {
	once.Do(func() {
		opts := &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Remove time from the output for cleaner logs
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}

		handler := slog.NewTextHandler(os.Stderr, opts)
		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger)
	})
	return nil
}

// SetLevel sets the logging level
func SetLevel(level string) error {
	switch level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelError) // Default to error level
	}
	return nil
}

// Get returns the default logger instance
func Get() *slog.Logger {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger
}

// Helper functions for structured logging

// Debug logs at debug level with structured key-value pairs
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs at info level with structured key-value pairs
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs at warn level with structured key-value pairs
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs at error level with structured key-value pairs
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// WithGroup creates a new logger with the specified group
func WithGroup(name string) *slog.Logger {
	return Get().WithGroup(name)
}

// With creates a new logger with the specified attributes
func With(attrs ...any) *slog.Logger {
	return Get().With(attrs...)
}
