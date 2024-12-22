package filesystem

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FilesystemLogger struct {
	mu       sync.Mutex
	file     *os.File
	logger   *log.Logger
	level    LogLevel
	basePath string
}

type LogLevel int

const (
	ErrorLevel LogLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

func NewFilesystemLogger(basePath string) (*FilesystemLogger, error) {
	// Create logs directory
	logDir := filepath.Join(basePath, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("backup_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	return &FilesystemLogger{
		file:     file,
		logger:   log.New(file, "", log.LstdFlags),
		level:    InfoLevel,
		basePath: basePath,
	}, nil
}

func (l *FilesystemLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *FilesystemLogger) Debug(format string, v ...interface{}) {
	if l.level >= DebugLevel {
		l.log("DEBUG", format, v...)
	}
}

func (l *FilesystemLogger) Info(format string, v ...interface{}) {
	if l.level >= InfoLevel {
		l.log("INFO", format, v...)
	}
}

func (l *FilesystemLogger) Warn(format string, v ...interface{}) {
	if l.level >= WarnLevel {
		l.log("WARN", format, v...)
	}
}

func (l *FilesystemLogger) Error(format string, v ...interface{}) {
	if l.level >= ErrorLevel {
		l.log("ERROR", format, v...)
	}
}

func (l *FilesystemLogger) log(level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s", level, msg)
}

func (l *FilesystemLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
