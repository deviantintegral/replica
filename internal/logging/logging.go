// Package logging provides structured logging utilities for the replica application.
package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/deviantintegral/replica/internal/config"
	"github.com/rs/zerolog"
)

// New creates a new zerolog.Logger based on the provided logging configuration.
// The logger outputs JSON format by default, or human-readable text format when
// configured for development. All log entries include timestamps.
func New(cfg config.LoggingConfig) zerolog.Logger {
	var output io.Writer = os.Stdout

	// Configure output format
	if strings.ToLower(cfg.Format) == "text" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger with configured level and timestamp
	level := parseLevel(cfg.Level)
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger
}

// parseLevel converts a string log level to a zerolog.Level.
// Unknown levels default to Info level.
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithComponent creates a new logger with a component field added.
// This is useful for creating component-specific loggers that identify
// which part of the application generated the log entry.
func WithComponent(logger zerolog.Logger, component string) zerolog.Logger {
	return logger.With().Str("component", component).Logger()
}
