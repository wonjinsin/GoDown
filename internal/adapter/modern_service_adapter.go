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
	"time"

	"github.com/rs/zerolog"
)

// ModernServiceAdapter adapts the new use case layer to work with legacy code
type ModernServiceAdapter struct {
	downloadService        service.DownloadService
	fileSequenceService    service.FileSequenceService
	mediaProcessingService service.MediaProcessingService
	config                 *config.Config
	logger                 zerolog.Logger
}

// NewModernServiceAdapter creates a new modern service adapter
func NewModernServiceAdapter(cfg *config.Config) *ModernServiceAdapter {
	// Initialize repositories
	repoFactory := repoImpl.NewFactory(cfg)
	repos := repoFactory.CreateAll()

	// Initialize use cases
	usecaseFactory := usecase.NewFactory(
		cfg,
		repos.FileRepo,
		repos.HTTPClient,
		repos.HTTPRepo,
		repos.MediaProcessor,
	)
	usecases := usecaseFactory.CreateAll()

	return &ModernServiceAdapter{
		downloadService:        usecases.DownloadService,
		fileSequenceService:    usecases.FileSequenceService,
		mediaProcessingService: usecases.MediaProcessingService,
		config:                 cfg,
		logger:                 logger.GetLogger("modern_service_adapter"),
	}
}

// Do executes the download process using the new use case layer
func (msa *ModernServiceAdapter) Do(input *model.Input, progressChan chan int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	msa.logger.Info().
		Str("url", input.URL).
		Str("folder", input.Folder).
		Msg("Starting download via modern service adapter")

	// Convert legacy input to use case request
	request := service.CreateDownloadJobRequest{
		URL:       input.URL,
		Folder:    input.Folder,
		Host:      input.Host,
		Origin:    input.Origin,
		Separator: input.Separator,
	}

	// Create download job
	job, err := msa.downloadService.CreateDownloadJob(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create download job: %w", err)
	}

	msa.logger.Info().
		Str("job_id", job.ID).
		Msg("Download job created")

	// Start progress monitoring if channel is provided
	if progressChan != nil {
		go msa.monitorProgressLegacy(ctx, job.ID, progressChan)
	}

	// Execute download
	if err := msa.downloadService.StartDownload(ctx, job); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	msa.logger.Info().
		Str("job_id", job.ID).
		Msg("Download completed successfully")

	return nil
}

// monitorProgressLegacy monitors progress and sends updates to the legacy channel
func (msa *ModernServiceAdapter) monitorProgressLegacy(ctx context.Context, jobID string, progressChan chan int) {
	defer close(progressChan)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress, err := msa.downloadService.GetDownloadProgress(ctx, jobID)
			if err != nil {
				continue
			}

			// Convert progress percentage to legacy format (completed files count)
			legacyProgress := int(progress.ProgressPercent * float64(progress.TotalFiles) / 100)
			if legacyProgress > 0 {
				select {
				case progressChan <- legacyProgress:
				case <-ctx.Done():
					return
				}
			}

			// Stop if completed
			if progress.Status == "completed" || progress.Status == "failed" || progress.Status == "cancelled" {
				return
			}
		}
	}
}