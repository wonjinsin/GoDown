package main

import (
	"cheetah/config"
	"cheetah/pkg/logger"
	"cheetah/template"
	"runtime"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger with configuration
	loggerConfig := &logger.Config{
		Level:      logger.LogLevel(cfg.Logger.Level),
		TimeFormat: "2006-01-02T15:04:05Z07:00",
		Pretty:     cfg.Logger.Pretty,
	}
	logger.Setup(loggerConfig)

	log.Info().
		Str("app_name", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("environment", cfg.App.Environment).
		Int("cpu_count", runtime.NumCPU()).
		Int("batch_size", cfg.Download.BatchSize).
		Int("max_errors", cfg.Download.MaxErrors).
		Msg("Starting GoDown application")

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Start GUI with configuration
	template.ShowMain(cfg)

	log.Info().Msg("Application shutdown")
}
