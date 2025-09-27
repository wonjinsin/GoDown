package media

import (
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	"cheetah/internal/domain/service"
	"cheetah/pkg/logger"
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
)

// mediaProcessingUseCase implements the MediaProcessingService interface
type mediaProcessingUseCase struct {
	fileRepo       repository.FileRepository
	mediaProcessor repository.MediaProcessor
	logger         zerolog.Logger
}

// NewMediaProcessingUseCase creates a new media processing use case
func NewMediaProcessingUseCase(
	fileRepo repository.FileRepository,
	mediaProcessor repository.MediaProcessor,
) service.MediaProcessingService {
	return &mediaProcessingUseCase{
		fileRepo:       fileRepo,
		mediaProcessor: mediaProcessor,
		logger:         logger.GetLogger("media_processing_usecase"),
	}
}

// ProcessDownloadedFiles processes all downloaded files for a job
func (mpu *mediaProcessingUseCase) ProcessDownloadedFiles(ctx context.Context, job *entity.DownloadJob) error {
	mpu.logger.Info().
		Str("job_id", job.ID).
		Str("folder", job.Folder).
		Msg("Processing downloaded files")

	// Get file sequence to determine extension
	sequence, err := job.GetFileSequence()
	if err != nil {
		return fmt.Errorf("failed to get file sequence: %w", err)
	}

	// Process files using media processor
	downloadPath := job.GetSavePath()
	if err := mpu.mediaProcessor.ProcessFiles(ctx, downloadPath, job.Folder, sequence.Extension); err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	mpu.logger.Info().
		Str("job_id", job.ID).
		Msg("File processing completed successfully")

	return nil
}

// ValidateMediaFiles validates a list of media files
func (mpu *mediaProcessingUseCase) ValidateMediaFiles(ctx context.Context, files []*entity.MediaFile) error {
	mpu.logger.Info().
		Int("file_count", len(files)).
		Msg("Validating media files")

	var wg sync.WaitGroup
	errors := make(chan error, len(files))
	semaphore := make(chan struct{}, 10) // Limit concurrent validations

	for _, file := range files {
		wg.Add(1)
		go func(f *entity.MediaFile) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := mpu.validateSingleFile(ctx, f); err != nil {
				errors <- fmt.Errorf("validation failed for %s: %w", f.Filename, err)
			}
		}(file)
	}

	wg.Wait()
	close(errors)

	// Collect all errors
	var validationErrors []error
	for err := range errors {
		validationErrors = append(validationErrors, err)
		mpu.logger.Warn().Err(err).Msg("File validation error")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed for %d files", len(validationErrors))
	}

	mpu.logger.Info().
		Int("validated_files", len(files)).
		Msg("All files validated successfully")

	return nil
}

// ExtractFileMetadata extracts metadata for a media file
func (mpu *mediaProcessingUseCase) ExtractFileMetadata(ctx context.Context, file *entity.MediaFile) error {
	mpu.logger.Debug().
		Str("filename", file.Filename).
		Msg("Extracting file metadata")

	fullPath := file.GetFullPath()

	// Extract metadata using media processor
	metadata, err := mpu.mediaProcessor.ExtractMetadata(ctx, fullPath)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Add metadata to file entity
	for key, value := range metadata {
		file.AddMetadata(key, value)
	}

	mpu.logger.Debug().
		Str("filename", file.Filename).
		Int("metadata_count", len(metadata)).
		Msg("Metadata extracted successfully")

	return nil
}

// GenerateChecksums generates checksums for a list of files
func (mpu *mediaProcessingUseCase) GenerateChecksums(ctx context.Context, files []*entity.MediaFile) error {
	mpu.logger.Info().
		Int("file_count", len(files)).
		Msg("Generating file checksums")

	var wg sync.WaitGroup
	errors := make(chan error, len(files))
	semaphore := make(chan struct{}, 5) // Limit concurrent checksum calculations

	for _, file := range files {
		wg.Add(1)
		go func(f *entity.MediaFile) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := mpu.generateSingleChecksum(ctx, f); err != nil {
				errors <- fmt.Errorf("checksum generation failed for %s: %w", f.Filename, err)
			}
		}(file)
	}

	wg.Wait()
	close(errors)

	// Collect all errors
	var checksumErrors []error
	for err := range errors {
		checksumErrors = append(checksumErrors, err)
		mpu.logger.Warn().Err(err).Msg("Checksum generation error")
	}

	if len(checksumErrors) > 0 {
		return fmt.Errorf("checksum generation failed for %d files", len(checksumErrors))
	}

	mpu.logger.Info().
		Int("processed_files", len(files)).
		Msg("All checksums generated successfully")

	return nil
}

// CleanupFailedDownloads removes failed or incomplete downloads
func (mpu *mediaProcessingUseCase) CleanupFailedDownloads(ctx context.Context, job *entity.DownloadJob) error {
	mpu.logger.Info().
		Str("job_id", job.ID).
		Msg("Cleaning up failed downloads")

	downloadPath := job.GetSavePath()

	// List all files in the download directory
	files, err := mpu.fileRepo.ListFiles(ctx, downloadPath)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	cleanedCount := 0
	for _, file := range files {
		// Check if file is valid
		if err := mpu.fileRepo.ValidateFile(ctx, file); err != nil {
			mpu.logger.Warn().
				Str("filename", file.Filename).
				Err(err).
				Msg("Removing invalid file")

			if err := mpu.fileRepo.DeleteFile(ctx, file.GetFullPath()); err != nil {
				mpu.logger.Error().
					Str("filename", file.Filename).
					Err(err).
					Msg("Failed to delete invalid file")
				continue
			}
			cleanedCount++
		}
	}

	mpu.logger.Info().
		Str("job_id", job.ID).
		Int("cleaned_files", cleanedCount).
		Msg("Failed download cleanup completed")

	return nil
}

// CleanupTemporaryFiles removes temporary files from a directory
func (mpu *mediaProcessingUseCase) CleanupTemporaryFiles(ctx context.Context, directory string) error {
	mpu.logger.Info().
		Str("directory", directory).
		Msg("Cleaning up temporary files")

	// List all files in the directory
	files, err := mpu.fileRepo.ListFiles(ctx, directory)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	cleanedCount := 0
	for _, file := range files {
		// Check if it's a temporary file (e.g., .tmp, .part, etc.)
		if mpu.isTemporaryFile(file.Filename) {
			mpu.logger.Debug().
				Str("filename", file.Filename).
				Msg("Removing temporary file")

			if err := mpu.fileRepo.DeleteFile(ctx, file.GetFullPath()); err != nil {
				mpu.logger.Error().
					Str("filename", file.Filename).
					Err(err).
					Msg("Failed to delete temporary file")
				continue
			}
			cleanedCount++
		}
	}

	mpu.logger.Info().
		Str("directory", directory).
		Int("cleaned_files", cleanedCount).
		Msg("Temporary file cleanup completed")

	return nil
}

// validateSingleFile validates a single media file
func (mpu *mediaProcessingUseCase) validateSingleFile(ctx context.Context, file *entity.MediaFile) error {
	// Validate file exists and is not empty
	if err := mpu.fileRepo.ValidateFile(ctx, file); err != nil {
		return err
	}

	// Validate media file format if it's a media file
	if file.IsVideo() || file.IsAudio() {
		fullPath := file.GetFullPath()
		if err := mpu.mediaProcessor.ValidateMediaFile(ctx, fullPath); err != nil {
			return fmt.Errorf("media validation failed: %w", err)
		}
	}

	return nil
}

// generateSingleChecksum generates a checksum for a single file
func (mpu *mediaProcessingUseCase) generateSingleChecksum(ctx context.Context, file *entity.MediaFile) error {
	fullPath := file.GetFullPath()

	checksum, err := mpu.fileRepo.CalculateChecksum(ctx, fullPath)
	if err != nil {
		return err
	}

	file.SetChecksum(checksum)

	mpu.logger.Debug().
		Str("filename", file.Filename).
		Str("checksum", checksum).
		Msg("Checksum generated")

	return nil
}

// isTemporaryFile checks if a filename indicates a temporary file
func (mpu *mediaProcessingUseCase) isTemporaryFile(filename string) bool {
	ext := filepath.Ext(filename)
	tempExtensions := []string{".tmp", ".temp", ".part", ".download", ".crdownload"}

	for _, tempExt := range tempExtensions {
		if ext == tempExt {
			return true
		}
	}

	return false
}
