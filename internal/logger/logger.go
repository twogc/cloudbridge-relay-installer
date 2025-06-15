package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents the severity of the log message
type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// Logger represents a structured logger
type Logger struct {
	output io.Writer
	level  LogLevel
}

// NewLogger creates a new logger
func NewLogger(output io.Writer, level LogLevel) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		output: output,
		level:  level,
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields interface{}, err error) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields interface{}) {
	if l.level == DEBUG {
		l.log(DEBUG, message, fields, nil)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, fields interface{}) {
	if l.level == DEBUG || l.level == INFO {
		l.log(INFO, message, fields, nil)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(message string, fields interface{}) {
	if l.level != ERROR {
		l.log(WARNING, message, fields, nil)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields interface{}) {
	l.log(ERROR, message, fields, err)
} 