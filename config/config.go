package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Download DownloadConfig `json:"download"`
	Logger   LoggerConfig   `json:"logger"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	RepoDir     string `json:"repo_dir"`
}

// DownloadConfig holds download-specific configuration
type DownloadConfig struct {
	BatchSize      int           `json:"batch_size"`
	MaxErrors      int           `json:"max_errors"`
	RoutineErrMax  int           `json:"routine_err_max"`
	HTTPTimeout    time.Duration `json:"http_timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	MaxConcurrency int           `json:"max_concurrency"`
	UserAgent      string        `json:"user_agent"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `json:"level"`
	Pretty bool   `json:"pretty"`
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnvOrDefault("APP_NAME", "cheetah"),
			Version:     getEnvOrDefault("APP_VERSION", "1.0.0"),
			Environment: getEnvOrDefault("APP_ENV", "development"),
			RepoDir:     getEnvOrDefault("REPO_DIR", "repo"),
		},
		Download: DownloadConfig{
			BatchSize:      parseIntOrDefault("DOWNLOAD_BATCH_SIZE", 50),
			MaxErrors:      parseIntOrDefault("DOWNLOAD_MAX_ERRORS", 100),
			RoutineErrMax:  parseIntOrDefault("DOWNLOAD_ROUTINE_ERR_MAX", 10),
			HTTPTimeout:    parseDurationOrDefault("HTTP_TIMEOUT", 30*time.Second),
			MaxRetries:     parseIntOrDefault("HTTP_MAX_RETRIES", 3),
			RetryDelay:     parseDurationOrDefault("HTTP_RETRY_DELAY", 1*time.Second),
			MaxConcurrency: parseIntOrDefault("HTTP_MAX_CONCURRENCY", 50),
			UserAgent:      getEnvOrDefault("HTTP_USER_AGENT", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"),
		},
		Logger: LoggerConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Pretty: parseBoolOrDefault("LOG_PRETTY", true),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if c.Download.BatchSize <= 0 {
		return fmt.Errorf("download batch size must be positive")
	}

	if c.Download.MaxErrors < 0 {
		return fmt.Errorf("download max errors cannot be negative")
	}

	if c.Download.RoutineErrMax < 0 {
		return fmt.Errorf("download routine error max cannot be negative")
	}

	if c.Download.HTTPTimeout <= 0 {
		return fmt.Errorf("HTTP timeout must be positive")
	}

	if c.Download.MaxRetries < 0 {
		return fmt.Errorf("HTTP max retries cannot be negative")
	}

	if c.Download.MaxConcurrency <= 0 {
		return fmt.Errorf("HTTP max concurrency must be positive")
	}

	return nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
