package service

import (
	"cheetah/config"
	"sync"
)

// Factory provides methods to create service instances
type Factory struct {
	config *config.Config
	
	// Singleton instances
	orchestrationService OrchestrationService
	once                sync.Once
}

// NewFactory creates a new service factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateOrchestrationService creates or returns the singleton orchestration service
func (f *Factory) CreateOrchestrationService() OrchestrationService {
	f.once.Do(func() {
		f.orchestrationService = NewOrchestrationService(f.config)
	})
	return f.orchestrationService
}

// CreateDownloadService creates a download service
func (f *Factory) CreateDownloadService() DownloadService {
	return f.CreateOrchestrationService()
}

// CreateProgressService creates a progress service
func (f *Factory) CreateProgressService() ProgressService {
	return f.CreateOrchestrationService()
}

// CreateValidationService creates a validation service
func (f *Factory) CreateValidationService() ValidationService {
	return f.CreateOrchestrationService()
}

// Services holds all service instances
type Services struct {
	OrchestrationService OrchestrationService
	DownloadService      DownloadService
	ProgressService      ProgressService
	ValidationService    ValidationService
}

// CreateAll creates all service instances
func (f *Factory) CreateAll() *Services {
	orchestrationService := f.CreateOrchestrationService()
	
	return &Services{
		OrchestrationService: orchestrationService,
		DownloadService:      orchestrationService,
		ProgressService:      orchestrationService,
		ValidationService:    orchestrationService,
	}
}
