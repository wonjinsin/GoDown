package filesequence

import (
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	"cheetah/internal/domain/service"
	"cheetah/internal/domain/value"
	"cheetah/pkg/logger"
	"cheetah/util"
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// fileSequenceUseCase implements the FileSequenceService interface
type fileSequenceUseCase struct {
	httpRepo repository.HTTPRepository
	logger   zerolog.Logger
}

// NewFileSequenceUseCase creates a new file sequence use case
func NewFileSequenceUseCase(httpRepo repository.HTTPRepository) service.FileSequenceService {
	return &fileSequenceUseCase{
		httpRepo: httpRepo,
		logger:   logger.GetLogger("file_sequence_usecase"),
	}
}

// AnalyzeSequence analyzes a URL to create a file sequence
func (fsu *fileSequenceUseCase) AnalyzeSequence(ctx context.Context, baseURL string) (*entity.FileSequence, error) {
	fsu.logger.Info().
		Str("base_url", baseURL).
		Msg("Analyzing file sequence")

	// Validate URL
	if err := fsu.httpRepo.ValidateURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create URL value object
	urlValue, err := value.NewURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract file extension
	extension, err := urlValue.GetFileExtension()
	if err != nil {
		return nil, fmt.Errorf("failed to extract file extension: %w", err)
	}

	// Create file sequence
	sequence := entity.NewFileSequence(baseURL, extension, nil)

	fsu.logger.Info().
		Str("extension", extension).
		Interface("pattern", sequence.GetPattern()).
		Msg("File sequence analyzed successfully")

	return sequence, nil
}

// GenerateFileList generates a list of media files from a sequence
func (fsu *fileSequenceUseCase) GenerateFileList(ctx context.Context, sequence *entity.FileSequence, count int) ([]*entity.MediaFile, error) {
	fsu.logger.Debug().
		Int("count", count).
		Msg("Generating file list from sequence")

	if count <= 0 {
		return nil, util.NewValidationError("count", fmt.Sprintf("%d", count), "count must be positive")
	}

	var files []*entity.MediaFile
	for i := 0; i < count; i++ {
		fileNum := uint64(i)

		// Generate URL and filename
		url, err := sequence.GenerateURL(fileNum)
		if err != nil {
			fsu.logger.Warn().
				Err(err).
				Uint64("file_number", fileNum).
				Msg("Failed to generate URL for file number")
			continue
		}

		filename := sequence.GenerateFilename(fileNum)

		// Create media file entity
		mediaFile := entity.NewMediaFile(filename, url, "")
		files = append(files, mediaFile)
	}

	fsu.logger.Info().
		Int("generated_files", len(files)).
		Int("requested_count", count).
		Msg("File list generated successfully")

	return files, nil
}

// ValidateSequence validates a file sequence
func (fsu *fileSequenceUseCase) ValidateSequence(ctx context.Context, sequence *entity.FileSequence) error {
	fsu.logger.Debug().
		Str("base_url", sequence.BaseURL).
		Str("extension", sequence.Extension).
		Msg("Validating file sequence")

	// Validate base URL
	if err := fsu.httpRepo.ValidateURL(sequence.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Validate extension
	if sequence.Extension == "" {
		return util.NewValidationError("extension", sequence.Extension, "file extension is required")
	}

	// Test URL generation for first few files
	testCount := 3
	for i := 0; i < testCount; i++ {
		url, err := sequence.GenerateURL(uint64(i))
		if err != nil {
			return fmt.Errorf("failed to generate URL for file %d: %w", i, err)
		}

		// Validate generated URL
		if err := fsu.httpRepo.ValidateURL(url); err != nil {
			return fmt.Errorf("generated invalid URL for file %d: %w", i, err)
		}
	}

	fsu.logger.Debug().Msg("File sequence validation successful")
	return nil
}

// DetectPattern detects the file naming pattern from a URL
func (fsu *fileSequenceUseCase) DetectPattern(ctx context.Context, url string) (*entity.FilePattern, error) {
	fsu.logger.Debug().
		Str("url", url).
		Msg("Detecting file pattern")

	// Create file sequence to analyze pattern
	sequence := entity.NewFileSequence(url, "", nil)
	pattern := sequence.GetPattern()

	if pattern == nil {
		return nil, fmt.Errorf("no pattern detected in URL: %s", url)
	}

	fsu.logger.Info().
		Str("prefix", pattern.Prefix).
		Str("separator", pattern.Separator).
		Int("number_length", pattern.NumberLength).
		Uint64("start_number", pattern.StartNumber).
		Msg("Pattern detected successfully")

	return pattern, nil
}

// EstimateFileCount estimates the number of files in a sequence
func (fsu *fileSequenceUseCase) EstimateFileCount(ctx context.Context, sequence *entity.FileSequence) (int, error) {
	fsu.logger.Debug().
		Str("base_url", sequence.BaseURL).
		Msg("Estimating file count")

	count := sequence.EstimateFileCount()

	fsu.logger.Info().
		Int("estimated_count", count).
		Msg("File count estimated")

	return count, nil
}
