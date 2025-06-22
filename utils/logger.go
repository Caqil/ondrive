package utils

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger interface defines logging methods
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
	WithContext(ctx map[string]interface{}) Logger
}

// AppLogger implements Logger interface
type AppLogger struct {
	logger zerolog.Logger
}

// LogConfig holds configuration for logger
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // "json" or "console"
	Output     string `json:"output"` // "stdout", "stderr", or file path
	TimeFormat string `json:"time_format"`
}

// NewLogger creates a new logger instance
func NewLogger(config LogConfig) Logger {
	// Set global error stack marshaler
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Set time format
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set output
	var output io.Writer = os.Stdout
	if config.Output != "" {
		switch config.Output {
		case "stdout":
			output = os.Stdout
		case "stderr":
			output = os.Stderr
		default:
			// File output
			if file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				output = file
			}
		}
	}

	// Create logger
	var logger zerolog.Logger
	if config.Format == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	}

	return &AppLogger{logger: logger}
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() Logger {
	config := LogConfig{
		Level:      "info",
		Format:     "console",
		Output:     "stdout",
		TimeFormat: time.RFC3339,
	}
	return NewLogger(config)
}

// Debug creates a debug level log event
func (l *AppLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info creates an info level log event
func (l *AppLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn creates a warn level log event
func (l *AppLogger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error creates an error level log event
func (l *AppLogger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal creates a fatal level log event
func (l *AppLogger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic creates a panic level log event
func (l *AppLogger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// WithContext adds context fields to logger
func (l *AppLogger) WithContext(ctx map[string]interface{}) Logger {
	context := l.logger.With()
	for key, value := range ctx {
		context = context.Interface(key, value)
	}
	return &AppLogger{logger: context.Logger()}
}

// RequestLogger creates a logger for HTTP requests
func RequestLogger() *AppLogger {
	logger := log.With().
		Str("component", "http").
		Logger()
	return &AppLogger{logger: logger}
}

// DatabaseLogger creates a logger for database operations
func DatabaseLogger() *AppLogger {
	logger := log.With().
		Str("component", "database").
		Logger()
	return &AppLogger{logger: logger}
}

// WebSocketLogger creates a logger for WebSocket operations
func WebSocketLogger() *AppLogger {
	logger := log.With().
		Str("component", "websocket").
		Logger()
	return &AppLogger{logger: logger}
}

// ServiceLogger creates a logger for service operations
func ServiceLogger(serviceName string) *AppLogger {
	logger := log.With().
		Str("component", "service").
		Str("service", serviceName).
		Logger()
	return &AppLogger{logger: logger}
}

// ControllerLogger creates a logger for controller operations
func ControllerLogger(controllerName string) *AppLogger {
	logger := log.With().
		Str("component", "controller").
		Str("controller", controllerName).
		Logger()
	return &AppLogger{logger: logger}
}

// LogRequestStart logs the start of an HTTP request
func LogRequestStart(logger Logger, method, path, userID string) {
	logger.Info().
		Str("method", method).
		Str("path", path).
		Str("user_id", userID).
		Msg("Request started")
}

// LogRequestEnd logs the end of an HTTP request
func LogRequestEnd(logger Logger, method, path string, statusCode int, duration time.Duration) {
	event := logger.Info()
	if statusCode >= 400 {
		event = logger.Error()
	}

	event.
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Msg("Request completed")
}

// LogError logs an error with context
func LogError(logger Logger, err error, context map[string]interface{}) {
	event := logger.Error().Err(err)

	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg("Error occurred")
}

// LogWebSocketEvent logs WebSocket events
func LogWebSocketEvent(logger Logger, event, userID, rideID string) {
	logger.Info().
		Str("event", event).
		Str("user_id", userID).
		Str("ride_id", rideID).
		Msg("WebSocket event")
}
