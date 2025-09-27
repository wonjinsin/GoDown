package adapter

import (
	"cheetah/config"
	"cheetah/internal/domain/service"
	repoImpl "cheetah/internal/repository"
	"cheetah/internal/usecase"
	"cheetah/model"
	"cheetah/pkg/logger"
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// ServiceAdapter adapts the new use case layer to work with legacy service interface
type ServiceAdapter struct {
	downloadService service.DownloadService
	config          *config.Config
	logger          zerolog.Logger
}

// NewServiceAdapter creates a new service adapter
func NewServiceAdapter(cfg *config.Config) *ServiceAdapter {
	// Create repositories
	repoFactory := repoImpl.NewFactory(cfg)
	repos := repoFactory.CreateAll()

	// Create use cases
	usecaseFactory := usecase.NewFactory(
		cfg,
		repos.FileRepo,
		repos.HTTPClient,
		repos.HTTPRepo,
		repos.MediaProcessor,
	)
	usecases := usecaseFactory.CreateAll()

	return &ServiceAdapter{
		downloadService: usecases.DownloadService,
		config:          cfg,
		logger:          logger.GetLogger("service_adapter"),
	}
}

// Do executes the download process using the new use case layer
func (sa *ServiceAdapter) Do(input *model.Input, progressChan chan int) error {
	ctx := context.Background()

	sa.logger.Info().
		Str("url", input.URL).
		Str("folder", input.Folder).
		Msg("Starting download via service adapter")

	// Convert legacy input to use case request
	request := service.CreateDownloadJobRequest{
		URL:       input.URL,
		Folder:    input.Folder,
		Host:      input.Host,
		Origin:    input.Origin,
		Separator: input.Separator,
	}

	// Create download job
	job, err := sa.downloadService.CreateDownloadJob(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create download job: %w", err)
	}

	sa.logger.Info().
		Str("job_id", job.ID).
		Msg("Download job created")

	// Start download with progress monitoring
	if progressChan != nil {
		go sa.monitorProgress(ctx, job.ID, progressChan)
	}

	// Execute download
	if err := sa.downloadService.StartDownload(ctx, job); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	sa.logger.Info().
		Str("job_id", job.ID).
		Msg("Download completed successfully via service adapter")

	return nil
}

// monitorProgress monitors download progress and sends updates to the channel
func (sa *ServiceAdapter) monitorProgress(ctx context.Context, jobID string, progressChan chan int) {
	defer close(progressChan)

	ticker := make(chan struct{})
	go func() {
		// Simple progress simulation - in a real implementation,
		// this would read actual progress from the use case
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				progressChan <- i * 10 // Send progress updates
				if i >= 99 {
					return
				}
			}
		}
	}()

	<-ticker
}

// LegacyServiceAdapter provides backward compatibility with the existing FileService interface
type LegacyServiceAdapter struct {
	serviceAdapter *ServiceAdapter
	input          *model.Input
	file           *model.File
}

// NewLegacyServiceAdapter creates a legacy service adapter
func NewLegacyServiceAdapter(input *model.Input, cfg *config.Config) *LegacyServiceAdapter {
	return &LegacyServiceAdapter{
		serviceAdapter: NewServiceAdapter(cfg),
		input:          input,
		file:           model.MakeFile(input, cfg),
	}
}

// Do implements the legacy FileUsecase interface
func (lsa *LegacyServiceAdapter) Do(progressChan chan int) error {
	return lsa.serviceAdapter.Do(lsa.input, progressChan)
}

// GetInput returns the input for backward compatibility
func (lsa *LegacyServiceAdapter) GetInput() *model.Input {
	return lsa.input
}

// GetFile returns the file for backward compatibility
func (lsa *LegacyServiceAdapter) GetFile() *model.File {
	return lsa.file
}
