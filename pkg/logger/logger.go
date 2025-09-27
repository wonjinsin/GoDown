package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel `json:"level"`
	TimeFormat string   `json:"time_format"`
	Pretty     bool     `json:"pretty"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      LevelInfo,
		TimeFormat: time.RFC3339,
		Pretty:     false,
	}
}

// Setup initializes the global logger with the given configuration
func Setup(cfg *Config) {
	// Set log level
	switch strings.ToLower(string(cfg.Level)) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Set time format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Configure output format
	if cfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Add caller information in development
	if cfg.Level == LevelDebug {
		log.Logger = log.With().Caller().Logger()
	}
}

// GetLogger returns a logger with component context
func GetLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// WithRequestID adds request ID to logger context
func WithRequestID(logger zerolog.Logger, requestID string) zerolog.Logger {
	return logger.With().Str("request_id", requestID).Logger()
}

// WithError adds error to logger context
func WithError(logger zerolog.Logger, err error) zerolog.Logger {
	return logger.With().Err(err).Logger()
}
