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
//	log.Debugw("Starting function", "param1", value1)
//	log.Infow("Operation complete", "files", fileCount)
//	log.Warnw("Retrying operation", "attempt", retryCount)
//	log.Errorw("Operation failed", "error", err)
//	log.Fatalw("Unrecoverable error", "error", err)
//
// Structured Logging:
//
//	log.Infow("message",
//	  "key1", value1,
//	  "key2", value2)
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger
var LogLevel = zap.NewAtomicLevel()

func Init() error {
	config := zap.NewDevelopmentConfig()
	config.Level = LogLevel

	// Customize output format
	config.DisableStacktrace = true
	config.DisableCaller = true // Remove file/line references
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "" // Remove timestamp
	config.EncoderConfig.NameKey = "" // Remove logger name
	config.EncoderConfig.CallerKey = ""
	config.EncoderConfig.FunctionKey = ""

	// Only show configuration block in debug mode
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.ConsoleSeparator = " " // Clean up JSON formatting

	logger, err := config.Build()
	if err != nil {
		return err
	}
	log = logger.Sugar()
	return nil
}

func SetLevel(level string) error {
	return LogLevel.UnmarshalText([]byte(level))
}

func Get() *zap.SugaredLogger {
	return log
}
