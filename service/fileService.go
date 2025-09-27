package service

import (
	"cheetah/config"
	"cheetah/internal/adapter"
	"cheetah/model"
)

// NewFileService creates a new file service using the modern use case architecture
func NewFileService(input *model.Input, cfg *config.Config) FileUsecase {
	// Use the new service adapter for better architecture
	return adapter.NewLegacyServiceAdapter(input, cfg)
}
