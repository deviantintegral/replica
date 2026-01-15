package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deviantintegral/replica/internal/config"
	"github.com/rs/zerolog"
)

func TestNew_JSONFormat(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	// Capture output
	var buf bytes.Buffer
	logger := zerolog.New(&buf).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Log a test message
	logger.Info().Msg("test message")

	// Verify JSON format
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "info" {
		t.Errorf("level = %v, want %v", logEntry["level"], "info")
	}
	if logEntry["message"] != "test message" {
		t.Errorf("message = %v, want %v", logEntry["message"], "test message")
	}
	if _, ok := logEntry["time"]; !ok {
		t.Error("timestamp not present in log output")
	}
}

func TestNew_TextFormat(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
	}

	// Capture output using ConsoleWriter
	var buf bytes.Buffer
	output := zerolog.ConsoleWriter{
		Out:        &buf,
		TimeFormat: "2006-01-02T15:04:05Z07:00", // RFC3339
		NoColor:    true,
	}

	logger := zerolog.New(output).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Log a test message
	logger.Info().Msg("test message")

	// Verify text format contains expected elements
	logOutput := buf.String()
	if !strings.Contains(logOutput, "INF") {
		t.Errorf("text output should contain 'INF', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test message") {
		t.Errorf("text output should contain 'test message', got: %s", logOutput)
	}
}

func TestNew_DebugLevel(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "debug",
		Format: "json",
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Debug message should be logged
	logger.Debug().Msg("debug message")

	if buf.Len() == 0 {
		t.Error("debug message should be logged at debug level")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "debug" {
		t.Errorf("level = %v, want %v", logEntry["level"], "debug")
	}
}

func TestNew_InfoLevel(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Debug message should NOT be logged at info level
	logger.Debug().Msg("debug message")

	if buf.Len() != 0 {
		t.Error("debug message should not be logged at info level")
	}

	// Info message should be logged
	logger.Info().Msg("info message")

	if buf.Len() == 0 {
		t.Error("info message should be logged at info level")
	}
}

func TestNew_WarnLevel(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "warn",
		Format: "json",
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Info message should NOT be logged at warn level
	logger.Info().Msg("info message")

	if buf.Len() != 0 {
		t.Error("info message should not be logged at warn level")
	}

	// Warn message should be logged
	logger.Warn().Msg("warn message")

	if buf.Len() == 0 {
		t.Error("warn message should be logged at warn level")
	}
}

func TestNew_ErrorLevel(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "error",
		Format: "json",
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).
		Level(parseLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	// Warn message should NOT be logged at error level
	logger.Warn().Msg("warn message")

	if buf.Len() != 0 {
		t.Error("warn message should not be logged at error level")
	}

	// Error message should be logged
	logger.Error().Msg("error message")

	if buf.Len() == 0 {
		t.Error("error message should be logged at error level")
	}
}

func TestParseLevel_ValidLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"Debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"WARN", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"ERROR", zerolog.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLevel_UnknownDefaultsToInfo(t *testing.T) {
	tests := []string{
		"unknown",
		"verbose",
		"trace",
		"",
		"invalid",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result := parseLevel(input)
			if result != zerolog.InfoLevel {
				t.Errorf("parseLevel(%q) = %v, want %v (default to info)", input, result, zerolog.InfoLevel)
			}
		})
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := zerolog.New(&buf).With().Timestamp().Logger()

	// Create component logger
	componentLogger := WithComponent(baseLogger, "database")

	// Log a message
	componentLogger.Info().Msg("component test")

	// Verify component field is present
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["component"] != "database" {
		t.Errorf("component = %v, want %v", logEntry["component"], "database")
	}
}

func TestWithComponent_MultipleComponents(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := zerolog.New(&buf).With().Timestamp().Logger()

	// Create different component loggers
	dbLogger := WithComponent(baseLogger, "database")
	httpLogger := WithComponent(baseLogger, "http")

	// Log with database logger
	dbLogger.Info().Msg("database message")
	dbOutput := buf.String()
	buf.Reset()

	// Log with http logger
	httpLogger.Info().Msg("http message")
	httpOutput := buf.String()

	// Verify each has correct component
	var dbEntry map[string]interface{}
	if err := json.Unmarshal([]byte(dbOutput), &dbEntry); err != nil {
		t.Errorf("Failed to parse database log output: %v", err)
	}
	if dbEntry["component"] != "database" {
		t.Errorf("database logger component = %v, want %v", dbEntry["component"], "database")
	}

	var httpEntry map[string]interface{}
	if err := json.Unmarshal([]byte(httpOutput), &httpEntry); err != nil {
		t.Errorf("Failed to parse http log output: %v", err)
	}
	if httpEntry["component"] != "http" {
		t.Errorf("http logger component = %v, want %v", httpEntry["component"], "http")
	}
}

func TestNew_Integration(t *testing.T) {
	// Test the actual New function creates a working logger
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	logger := New(cfg)

	// The logger should be usable (this test just verifies no panic)
	logger.Info().Str("key", "value").Msg("integration test")
}

func TestNew_CaseInsensitiveFormat(t *testing.T) {
	formats := []string{"TEXT", "Text", "text", "JSON", "Json", "json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cfg := config.LoggingConfig{
				Level:  "info",
				Format: format,
			}

			// Should not panic
			logger := New(cfg)
			logger.Info().Msg("test")
		})
	}
}
