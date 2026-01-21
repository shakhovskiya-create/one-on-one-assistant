package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Caller    string                 `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	mu     sync.Mutex
	output *os.File
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	once.Do(func() {
		level := INFO
		if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
			switch strings.ToUpper(lvl) {
			case "DEBUG":
				level = DEBUG
			case "INFO":
				level = INFO
			case "WARN", "WARNING":
				level = WARN
			case "ERROR":
				level = ERROR
			}
		}
		defaultLogger = &Logger{
			level:  level,
			output: os.Stdout,
		}
	})
	return defaultLogger
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		// Get just the file name, not full path
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Caller:    caller,
		Fields:    fields,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, `{"timestamp":"%s","level":"ERROR","message":"failed to marshal log entry: %v"}`+"\n",
			time.Now().UTC().Format(time.RFC3339), err)
		return
	}

	fmt.Fprintln(l.output, string(jsonBytes))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f)
}

// WithFields creates a new log entry with fields
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerWithFields {
	return &LoggerWithFields{
		logger: l,
		fields: fields,
	}
}

// LoggerWithFields is a logger with pre-set fields
type LoggerWithFields struct {
	logger *Logger
	fields map[string]interface{}
}

func (lf *LoggerWithFields) Debug(message string) {
	lf.logger.log(DEBUG, message, lf.fields)
}

func (lf *LoggerWithFields) Info(message string) {
	lf.logger.log(INFO, message, lf.fields)
}

func (lf *LoggerWithFields) Warn(message string) {
	lf.logger.log(WARN, message, lf.fields)
}

func (lf *LoggerWithFields) Error(message string) {
	lf.logger.log(ERROR, message, lf.fields)
}

// HTTPRequestLog logs HTTP request details
func (l *Logger) HTTPRequestLog(method, path string, status int, duration time.Duration, clientIP string) {
	l.Info("HTTP Request", map[string]interface{}{
		"method":      method,
		"path":        path,
		"status":      status,
		"duration_ms": duration.Milliseconds(),
		"client_ip":   clientIP,
	})
}
