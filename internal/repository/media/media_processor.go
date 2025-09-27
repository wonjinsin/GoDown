package media

import (
	"cheetah/internal/domain/repository"
	"cheetah/pkg/logger"
	"cheetah/util"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// mediaProcessor implements the MediaProcessor interface
type mediaProcessor struct {
	logger zerolog.Logger
}

// NewMediaProcessor creates a new media processor
func NewMediaProcessor() repository.MediaProcessor {
	return &mediaProcessor{
		logger: logger.GetLogger("media_processor"),
	}
}

// ProcessFiles processes downloaded files using FFmpeg
func (mp *mediaProcessor) ProcessFiles(ctx context.Context, inputDir, outputFilename, extension string) error {
	mp.logger.Info().
		Str("input_dir", inputDir).
		Str("output_filename", outputFilename).
		Str("extension", extension).
		Msg("Starting media processing")

	// Create file list for FFmpeg concat
	fileListPath := filepath.Join(inputDir, "fl.txt")
	if err := mp.createFileList(inputDir, extension, fileListPath); err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer os.Remove(fileListPath) // Clean up file list

	// Run FFmpeg command
	outputPath := filepath.Join(inputDir, outputFilename+".mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-f", "concat",
		"-i", fileListPath,
		"-c", "copy",
		"-y", outputPath,
	)

	mp.logger.Debug().
		Str("command", cmd.String()).
		Msg("Executing FFmpeg command")

	output, err := cmd.CombinedOutput()
	if err != nil {
		mp.logger.Error().
			Err(err).
			Str("output", string(output)).
			Str("command", cmd.String()).
			Msg("FFmpeg command failed")
		return fmt.Errorf("%s: %w", util.ErrFFmpegExecution, err)
	}

	// Check if output file was created successfully
	if _, err := os.Stat(outputPath); err != nil {
		mp.logger.Error().
			Err(err).
			Str("output_path", outputPath).
			Msg("Output file was not created")
		return fmt.Errorf("output file was not created: %w", err)
	}

	// Clean up individual files after successful processing
	if err := mp.cleanupInputFiles(inputDir, extension); err != nil {
		mp.logger.Warn().
			Err(err).
			Msg("Failed to clean up input files")
		// Don't return error as the main processing succeeded
	}

	mp.logger.Info().
		Str("output_path", outputPath).
		Msg("Media processing completed successfully")

	return nil
}

// ValidateMediaFile validates a media file
func (mp *mediaProcessor) ValidateMediaFile(ctx context.Context, filepath string) error {
	mp.logger.Debug().
		Str("filepath", filepath).
		Msg("Validating media file")

	// Check if file exists
	info, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	// Check if file is not empty
	if info.Size() == 0 {
		return fmt.Errorf("%s: file is empty", util.ErrEmptyFile)
	}

	// Use FFprobe to validate media file
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filepath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		mp.logger.Warn().
			Err(err).
			Str("filepath", filepath).
			Str("output", string(output)).
			Msg("FFprobe validation failed")
		return fmt.Errorf("invalid media file: %w", err)
	}

	mp.logger.Debug().
		Str("filepath", filepath).
		Msg("Media file validation successful")

	return nil
}

// ExtractMetadata extracts metadata from a media file
func (mp *mediaProcessor) ExtractMetadata(ctx context.Context, filepath string) (map[string]interface{}, error) {
	mp.logger.Debug().
		Str("filepath", filepath).
		Msg("Extracting media metadata")

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filepath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// For now, return basic metadata
	metadata := make(map[string]interface{})
	metadata["raw_output"] = string(output)

	// Extract duration if possible
	if duration, err := mp.GetDuration(ctx, filepath); err == nil {
		metadata["duration"] = duration
	}

	return metadata, nil
}

// GetDuration gets the duration of a media file
func (mp *mediaProcessor) GetDuration(ctx context.Context, filepath string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filepath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// ConvertFormat converts a media file to a different format
func (mp *mediaProcessor) ConvertFormat(ctx context.Context, inputPath, outputPath, format string) error {
	mp.logger.Info().
		Str("input_path", inputPath).
		Str("output_path", outputPath).
		Str("format", format).
		Msg("Converting media format")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-f", format,
		"-y", outputPath,
	)

	if err := cmd.Run(); err != nil {
		mp.logger.Error().
			Err(err).
			Str("input_path", inputPath).
			Str("output_path", outputPath).
			Msg("Format conversion failed")
		return fmt.Errorf("format conversion failed: %w", err)
	}

	mp.logger.Info().
		Str("output_path", outputPath).
		Msg("Format conversion completed")

	return nil
}

// CompressFile compresses a media file
func (mp *mediaProcessor) CompressFile(ctx context.Context, inputPath, outputPath string, quality int) error {
	mp.logger.Info().
		Str("input_path", inputPath).
		Str("output_path", outputPath).
		Int("quality", quality).
		Msg("Compressing media file")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-crf", strconv.Itoa(quality),
		"-y", outputPath,
	)

	if err := cmd.Run(); err != nil {
		mp.logger.Error().
			Err(err).
			Str("input_path", inputPath).
			Str("output_path", outputPath).
			Msg("File compression failed")
		return fmt.Errorf("file compression failed: %w", err)
	}

	mp.logger.Info().
		Str("output_path", outputPath).
		Msg("File compression completed")

	return nil
}

// createFileList creates a file list for FFmpeg concat
func (mp *mediaProcessor) createFileList(inputDir, extension, fileListPath string) error {
	// Find all files with the given extension
	pattern := filepath.Join(inputDir, "*."+extension)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no files found with extension: %s", extension)
	}

	// Sort files numerically
	mp.sortFilesByNumber(matches)

	// Create file list
	file, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer file.Close()

	for _, match := range matches {
		filename := filepath.Base(match)
		// Escape spaces in filenames
		escapedFilename := strings.ReplaceAll(filename, " ", "\\ ")
		if _, err := fmt.Fprintf(file, "file '%s'\n", escapedFilename); err != nil {
			return fmt.Errorf("failed to write to file list: %w", err)
		}
	}

	mp.logger.Debug().
		Str("file_list_path", fileListPath).
		Int("file_count", len(matches)).
		Msg("File list created")

	return nil
}

// sortFilesByNumber sorts files by their numeric prefix
func (mp *mediaProcessor) sortFilesByNumber(files []string) {
	// Simple numeric sort based on filename
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			num1 := mp.extractNumber(filepath.Base(files[i]))
			num2 := mp.extractNumber(filepath.Base(files[j]))
			if num1 > num2 {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

// extractNumber extracts the numeric part from a filename
func (mp *mediaProcessor) extractNumber(filename string) int {
	r := regexp.MustCompile(`^(\d+)`)
	matches := r.FindStringSubmatch(filename)
	if len(matches) >= 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return 0
}

// cleanupInputFiles removes individual files after successful processing
func (mp *mediaProcessor) cleanupInputFiles(inputDir, extension string) error {
	pattern := filepath.Join(inputDir, "*."+extension)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find files for cleanup: %w", err)
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			mp.logger.Warn().
				Err(err).
				Str("file", match).
				Msg("Failed to remove file during cleanup")
		}
	}

	mp.logger.Info().
		Int("files_removed", len(matches)).
		Msg("Input files cleaned up")

	return nil
}
