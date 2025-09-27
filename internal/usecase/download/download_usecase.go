package download

import (
	"cheetah/config"
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	"cheetah/internal/domain/service"
	contextPkg "cheetah/pkg/context"
	"cheetah/pkg/logger"
	"cheetah/pkg/metrics"
	"cheetah/util"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// downloadUseCase implements the DownloadService interface
type downloadUseCase struct {
	fileRepo       repository.FileRepository
	httpClient     repository.HTTPClient
	httpRepo       repository.HTTPRepository
	mediaProcessor repository.MediaProcessor
	config         *config.Config
	logger         zerolog.Logger

	// Progress tracking
	progressMu sync.RWMutex
	progress   map[string]*service.DownloadProgress

	// Context management
	contextManager *contextPkg.Manager

	// Metrics collection
	metricsCollector *metrics.Collector
}

// NewDownloadUseCase creates a new download use case
func NewDownloadUseCase(
	fileRepo repository.FileRepository,
	httpClient repository.HTTPClient,
	httpRepo repository.HTTPRepository,
	mediaProcessor repository.MediaProcessor,
	cfg *config.Config,
) service.DownloadService {
	logger := logger.GetLogger("download_usecase")
	return &downloadUseCase{
		fileRepo:         fileRepo,
		httpClient:       httpClient,
		httpRepo:         httpRepo,
		mediaProcessor:   mediaProcessor,
		config:           cfg,
		logger:           logger,
		progress:         make(map[string]*service.DownloadProgress),
		contextManager:   contextPkg.NewManager(logger),
		metricsCollector: metrics.NewCollector(logger),
	}
}

// CreateDownloadJob creates a new download job
func (du *downloadUseCase) CreateDownloadJob(ctx context.Context, request service.CreateDownloadJobRequest) (*entity.DownloadJob, error) {
	du.logger.Info().
		Str("url", request.URL).
		Str("folder", request.Folder).
		Msg("Creating download job")

	// Validate request
	if err := du.ValidateDownloadRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create download job entity
	job := entity.NewDownloadJob(request.URL, request.Folder, du.config.App.RepoDir)

	if request.Host != nil {
		job.SetHost(*request.Host)
	}
	if request.Origin != nil {
		job.SetOrigin(*request.Origin)
	}
	if request.Separator != nil {
		job.SetSeparator(*request.Separator)
	}

	// Initialize progress tracking
	du.progressMu.Lock()
	du.progress[job.ID] = &service.DownloadProgress{
		JobID:           job.ID,
		TotalFiles:      0,
		CompletedFiles:  0,
		FailedFiles:     0,
		TotalBytes:      0,
		DownloadedBytes: 0,
		ProgressPercent: 0,
		Speed:           0,
		ETA:             0,
		Status:          job.Status.String(),
		CurrentFile:     "",
	}
	du.progressMu.Unlock()

	du.logger.Info().
		Str("job_id", job.ID).
		Msg("Download job created successfully")

	return job, nil
}

// StartDownload starts the download process
func (du *downloadUseCase) StartDownload(ctx context.Context, job *entity.DownloadJob) error {
	// Create a dedicated context for this download with timeout
	downloadTimeout := 30 * time.Minute // Default timeout
	downloadCtx, cancel := du.contextManager.CreateContext(ctx, job.ID, downloadTimeout)
	defer cancel()

	// Add context values
	downloadCtx = contextPkg.WithSessionID(downloadCtx, job.ID)
	downloadCtx = contextPkg.WithOperation(downloadCtx, "download")

	du.logger.Info().
		Str("job_id", job.ID).
		Str("url", job.URL).
		Str("folder", job.Folder).
		Dur("timeout", downloadTimeout).
		Msg("Starting download with context")

	// Mark job as started
	job.Start()
	du.updateProgressStatus(job.ID, job.Status.String())

	// Start metrics collection for this session
	sequence, err := job.GetFileSequence()
	if err == nil {
		estimatedFiles := sequence.EstimateFileCount()
		du.metricsCollector.StartDownloadSession(job.ID, estimatedFiles)
	}

	// Create download directory
	downloadPath := job.GetSavePath()
	if err := du.fileRepo.CreateDirectory(downloadCtx, downloadPath); err != nil {
		job.Fail(fmt.Errorf("failed to create directory: %w", err))
		du.updateProgressError(job.ID, err.Error())
		return err
	}

	// Get file sequence (already retrieved above for metrics)
	if sequence == nil {
		sequence, err = job.GetFileSequence()
		if err != nil {
			job.Fail(fmt.Errorf("failed to analyze file sequence: %w", err))
			du.updateProgressError(job.ID, err.Error())
			du.metricsCollector.CompleteDownloadSession(job.ID, false)
			return err
		}
	}

	// Start concurrent download with cancellation support
	if err := du.downloadFiles(downloadCtx, job, sequence); err != nil {
		if downloadCtx.Err() != nil {
			job.Cancel()
			du.updateProgressStatus(job.ID, job.Status.String())
			du.metricsCollector.CancelDownloadSession(job.ID)
			return fmt.Errorf("download cancelled: %w", downloadCtx.Err())
		}
		job.Fail(err)
		du.updateProgressError(job.ID, err.Error())
		du.metricsCollector.CompleteDownloadSession(job.ID, false)
		return err
	}

	// Process downloaded files
	if err := du.processFiles(downloadCtx, job, sequence); err != nil {
		if downloadCtx.Err() != nil {
			job.Cancel()
			du.updateProgressStatus(job.ID, job.Status.String())
			return fmt.Errorf("processing cancelled: %w", downloadCtx.Err())
		}
		job.Fail(fmt.Errorf("failed to process files: %w", err))
		du.updateProgressError(job.ID, err.Error())
		return err
	}

	// Mark job as completed
	job.Complete()
	du.updateProgressStatus(job.ID, job.Status.String())
	du.metricsCollector.CompleteDownloadSession(job.ID, true)

	du.logger.Info().
		Str("job_id", job.ID).
		Msg("Download completed successfully")

	return nil
}

// CancelDownload cancels a download job
func (du *downloadUseCase) CancelDownload(ctx context.Context, jobID string) error {
	du.logger.Info().
		Str("job_id", jobID).
		Msg("Cancelling download")

	// Cancel the context for this job
	if du.contextManager.CancelContext(jobID) {
		du.logger.Info().
			Str("job_id", jobID).
			Msg("Download context cancelled successfully")
	}

	du.progressMu.Lock()
	if progress, exists := du.progress[jobID]; exists {
		progress.Status = entity.StatusCancelled.String()
		progress.Error = "Download cancelled by user"
	}
	du.progressMu.Unlock()

	return nil
}

// GetDownloadProgress returns the current progress of a download
func (du *downloadUseCase) GetDownloadProgress(ctx context.Context, jobID string) (*service.DownloadProgress, error) {
	du.progressMu.RLock()
	defer du.progressMu.RUnlock()

	progress, exists := du.progress[jobID]
	if !exists {
		return nil, fmt.Errorf("download job not found: %s", jobID)
	}

	// Create a copy to avoid race conditions
	progressCopy := *progress
	return &progressCopy, nil
}

// GetDownloadStatus returns the status of a download job
func (du *downloadUseCase) GetDownloadStatus(ctx context.Context, jobID string) (*entity.DownloadJob, error) {
	// Note: In a real implementation, we would store job state in a repository
	// For now, we'll return a basic status based on progress

	progress, err := du.GetDownloadProgress(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Create a basic job status (this would normally come from storage)
	job := &entity.DownloadJob{
		ID:        jobID,
		Status:    parseStatus(progress.Status),
		CreatedAt: time.Now(), // This would be the actual creation time
	}

	if progress.Error != "" {
		job.Error = fmt.Errorf("%s", progress.Error)
	}

	return job, nil
}

// ValidateDownloadRequest validates a download request
func (du *downloadUseCase) ValidateDownloadRequest(ctx context.Context, request service.CreateDownloadJobRequest) error {
	if request.URL == "" {
		return util.NewValidationError("url", request.URL, "URL is required")
	}

	if request.Folder == "" {
		return util.NewValidationError("folder", request.Folder, "Folder is required")
	}

	// Validate URL format
	if err := du.httpRepo.ValidateURL(request.URL); err != nil {
		return err
	}

	return nil
}

// EstimateDownloadSize estimates the total download size
func (du *downloadUseCase) EstimateDownloadSize(ctx context.Context, job *entity.DownloadJob) (int64, error) {
	// This is a basic implementation - in practice, you might want to
	// make HEAD requests to estimate file sizes
	sequence, err := job.GetFileSequence()
	if err != nil {
		return 0, err
	}

	estimatedCount := sequence.EstimateFileCount()
	estimatedSizePerFile := int64(1024 * 1024) // 1MB per file as rough estimate

	return int64(estimatedCount) * estimatedSizePerFile, nil
}

// downloadFiles handles concurrent file downloading
func (du *downloadUseCase) downloadFiles(ctx context.Context, job *entity.DownloadJob, sequence *entity.FileSequence) error {
	batchSize := du.config.Download.BatchSize
	maxErrors := du.config.Download.MaxErrors
	routineErrMax := du.config.Download.RoutineErrMax

	startNum := 0
	errCount := 0
	totalDownloaded := 0

	du.logger.Info().
		Int("batch_size", batchSize).
		Int("max_errors", maxErrors).
		Msg("Starting concurrent file download")

	for {
		// Check for cancellation first
		select {
		case <-ctx.Done():
			du.logger.Info().Msg("Download cancelled by context")
			return ctx.Err()
		default:
		}

		if errCount > maxErrors {
			du.logger.Warn().
				Int("error_count", errCount).
				Int("max_errors", maxErrors).
				Msg("Maximum errors reached, stopping download")
			break
		}

		var wg sync.WaitGroup
		wg.Add(batchSize)
		batchErrors := make(chan error, batchSize)

		for j := 0; j < batchSize; j++ {
			go func(fileNum int) {
				defer wg.Done()

				// Check for cancellation before starting file download
				select {
				case <-ctx.Done():
					batchErrors <- ctx.Err()
					return
				default:
				}

				if err := du.downloadSingleFile(ctx, job, sequence, uint64(fileNum), routineErrMax); err != nil {
					batchErrors <- err
					return
				}

				totalDownloaded++
				du.updateProgress(job.ID, totalDownloaded, 0, 0)
			}(startNum + j)
		}

		wg.Wait()
		close(batchErrors)

		// Count errors from this batch
		for err := range batchErrors {
			du.logger.Error().Err(err).Msg("File download error")
			errCount++
		}

		startNum += batchSize
	}

	du.logger.Info().
		Int("total_downloaded", totalDownloaded).
		Int("total_errors", errCount).
		Msg("File download phase completed")

	return nil
}

// downloadSingleFile downloads a single file with retry logic
func (du *downloadUseCase) downloadSingleFile(ctx context.Context, job *entity.DownloadJob, sequence *entity.FileSequence, fileNum uint64, maxRetries int) error {
	url, err := sequence.GenerateURL(fileNum)
	if err != nil {
		return fmt.Errorf("failed to generate URL for file %d: %w", fileNum, err)
	}

	filename := sequence.GenerateFilename(fileNum)

	du.updateCurrentFile(job.ID, filename)

	// Build headers
	var hostStr, originStr string
	if job.Host != nil {
		hostStr = *job.Host
	} else {
		hostStr = du.httpRepo.ExtractDomain(url)
	}
	if job.Origin != nil {
		originStr = *job.Origin
	} else {
		originStr = du.httpRepo.ExtractDomain(url)
	}

	headers := du.httpRepo.BuildHeaders(hostStr, originStr, du.config.Download.UserAgent)

	// Retry loop
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			du.logger.Warn().
				Str("url", url).
				Str("filename", filename).
				Int("attempt", attempt+1).
				Msg("Retrying file download")
		}

		// Download file
		resp, err := du.httpClient.Download(ctx, url, headers)
		if err != nil {
			if attempt == maxRetries-1 {
				return util.NewDownloadError(url, filename, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			if attempt == maxRetries-1 {
				return util.NewDownloadError(url, filename, fmt.Errorf("HTTP %d", resp.StatusCode))
			}
			continue
		}

		// Save file
		_, err = du.fileRepo.SaveFile(ctx, filename, resp.Body, job.GetSavePath())
		if err != nil {
			if attempt == maxRetries-1 {
				return util.NewDownloadError(url, filename, err)
			}
			continue
		}

		du.logger.Debug().
			Str("filename", filename).
			Int64("size", resp.ContentLength).
			Msg("File downloaded successfully")

		return nil
	}

	return util.NewDownloadError(url, filename, fmt.Errorf("max retries exceeded"))
}

// processFiles processes downloaded files using media processor
func (du *downloadUseCase) processFiles(ctx context.Context, job *entity.DownloadJob, sequence *entity.FileSequence) error {
	du.logger.Info().
		Str("job_id", job.ID).
		Msg("Starting file processing")

	return du.mediaProcessor.ProcessFiles(ctx, job.GetSavePath(), job.Folder, sequence.Extension)
}

// Helper methods for progress tracking
func (du *downloadUseCase) updateProgress(jobID string, completed, failed int, downloadedBytes int64) {
	du.progressMu.Lock()
	defer du.progressMu.Unlock()

	if progress, exists := du.progress[jobID]; exists {
		progress.CompletedFiles = completed
		progress.FailedFiles = failed
		progress.DownloadedBytes = downloadedBytes

		if progress.TotalFiles > 0 {
			progress.ProgressPercent = float64(completed) / float64(progress.TotalFiles) * 100
		}
	}
}

func (du *downloadUseCase) updateProgressStatus(jobID, status string) {
	du.progressMu.Lock()
	defer du.progressMu.Unlock()

	if progress, exists := du.progress[jobID]; exists {
		progress.Status = status
	}
}

func (du *downloadUseCase) updateProgressError(jobID, errorMsg string) {
	du.progressMu.Lock()
	defer du.progressMu.Unlock()

	if progress, exists := du.progress[jobID]; exists {
		progress.Error = errorMsg
	}
}

func (du *downloadUseCase) updateCurrentFile(jobID, filename string) {
	du.progressMu.Lock()
	defer du.progressMu.Unlock()

	if progress, exists := du.progress[jobID]; exists {
		progress.CurrentFile = filename
	}
}

// parseStatus converts string status to DownloadStatus
func parseStatus(status string) entity.DownloadStatus {
	switch status {
	case "pending":
		return entity.StatusPending
	case "running":
		return entity.StatusRunning
	case "completed":
		return entity.StatusCompleted
	case "failed":
		return entity.StatusFailed
	case "cancelled":
		return entity.StatusCancelled
	default:
		return entity.StatusPending
	}
}
