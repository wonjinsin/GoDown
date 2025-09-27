package service

import (
	"cheetah/config"
	"cheetah/internal/adapter"
	"cheetah/model"
	"cheetah/pkg/logger"

	"github.com/rs/zerolog"
)

// NewFileService creates a new file service using the modern orchestration architecture
func NewFileService(input *model.Input, cfg *config.Config) FileUsecase {
	return NewModernFileService(input, cfg)
}

// ModernFileService implements the legacy FileUsecase interface using the new orchestration service
type ModernFileService struct {
	adapter *adapter.ModernServiceAdapter
	input   *model.Input
	logger  zerolog.Logger
}

// NewModernFileService creates a new modern file service
func NewModernFileService(input *model.Input, cfg *config.Config) FileUsecase {
	return &ModernFileService{
		adapter: adapter.NewModernServiceAdapter(cfg),
		input:   input,
		logger:  logger.GetLogger("modern_file_service"),
	}
}

// Do implements the legacy FileUsecase interface
func (mfs *ModernFileService) Do(progressChan chan int) error {
	mfs.logger.Info().
		Str("url", mfs.input.URL).
		Str("folder", mfs.input.Folder).
		Msg("Executing download via modern file service")

	return mfs.adapter.Do(mfs.input, progressChan)
}
