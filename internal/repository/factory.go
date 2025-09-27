package repository

import (
	"cheetah/config"
	"cheetah/internal/domain/repository"
	"cheetah/internal/repository/filesystem"
	"cheetah/internal/repository/http"
	"cheetah/internal/repository/media"
)

// Factory provides methods to create repository instances
type Factory struct {
	config *config.Config
}

// NewFactory creates a new repository factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateFileRepository creates a new file repository
func (f *Factory) CreateFileRepository() repository.FileRepository {
	return filesystem.NewFileRepository()
}

// CreateHTTPRepository creates a new HTTP repository
func (f *Factory) CreateHTTPRepository() repository.HTTPRepository {
	return http.NewHTTPRepository()
}

// CreateHTTPClient creates a new HTTP client with configuration
func (f *Factory) CreateHTTPClient() repository.HTTPClient {
	httpRepo := f.CreateHTTPRepository()

	config := repository.HTTPClientConfig{
		Timeout:       f.config.Download.HTTPTimeout,
		MaxRetries:    f.config.Download.MaxRetries,
		RetryDelay:    f.config.Download.RetryDelay,
		UserAgent:     f.config.Download.UserAgent,
		MaxRedirects:  10, // reasonable default
		EnableCookies: false,
	}

	return httpRepo.CreateClient(config)
}

// CreateMediaProcessor creates a new media processor
func (f *Factory) CreateMediaProcessor() repository.MediaProcessor {
	return media.NewMediaProcessor()
}

// Repositories holds all repository instances
type Repositories struct {
	FileRepo       repository.FileRepository
	HTTPRepo       repository.HTTPRepository
	HTTPClient     repository.HTTPClient
	MediaProcessor repository.MediaProcessor
}

// CreateAll creates all repository instances
func (f *Factory) CreateAll() *Repositories {
	return &Repositories{
		FileRepo:       f.CreateFileRepository(),
		HTTPRepo:       f.CreateHTTPRepository(),
		HTTPClient:     f.CreateHTTPClient(),
		MediaProcessor: f.CreateMediaProcessor(),
	}
}
