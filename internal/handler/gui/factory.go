package gui

import (
	"cheetah/config"
	"cheetah/service"
)

// Factory provides methods to create GUI handler instances
type Factory struct {
	config *config.Config
}

// NewFactory creates a new GUI handler factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateFyneHandler creates a new Fyne GUI handler
func (f *Factory) CreateFyneHandler(orchestrationService service.OrchestrationService) GUIHandler {
	return NewFyneHandler(orchestrationService, f.config)
}

// CreateDefaultHandler creates the default GUI handler (Fyne)
func (f *Factory) CreateDefaultHandler(orchestrationService service.OrchestrationService) GUIHandler {
	return f.CreateFyneHandler(orchestrationService)
}
