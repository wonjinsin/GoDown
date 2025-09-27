package filesystem

import (
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	"cheetah/pkg/logger"
	"cheetah/util"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

// fileRepository implements the FileRepository interface
type fileRepository struct {
	logger zerolog.Logger
}

// NewFileRepository creates a new file repository
func NewFileRepository() repository.FileRepository {
	return &fileRepository{
		logger: logger.GetLogger("file_repository"),
	}
}

// CreateDirectory creates a directory if it doesn't exist
func (fr *fileRepository) CreateDirectory(ctx context.Context, path string) error {
	fr.logger.Debug().
		Str("path", path).
		Msg("Creating directory")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			fr.logger.Error().
				Err(err).
				Str("path", path).
				Msg("Failed to create directory")
			return fmt.Errorf("%s: %w", util.ErrDirectoryCreate, err)
		}

		fr.logger.Info().
			Str("path", path).
			Msg("Directory created successfully")
	} else {
		fr.logger.Debug().
			Str("path", path).
			Msg("Directory already exists")
	}

	return nil
}

// DirectoryExists checks if a directory exists
func (fr *fileRepository) DirectoryExists(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check directory: %w", err)
	}
	return info.IsDir(), nil
}

// SaveFile saves content to a file and returns a MediaFile entity
func (fr *fileRepository) SaveFile(ctx context.Context, filename string, content io.Reader, path string) (*entity.MediaFile, error) {
	fullPath := filepath.Join(path, filename)

	fr.logger.Debug().
		Str("filename", filename).
		Str("path", path).
		Str("full_path", fullPath).
		Msg("Saving file")

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		fr.logger.Error().
			Err(err).
			Str("full_path", fullPath).
			Msg("Failed to create file")
		return nil, fmt.Errorf("%s: %w", util.ErrMakeFile, err)
	}
	defer file.Close()

	// Copy content to file
	written, err := io.Copy(file, content)
	if err != nil {
		fr.logger.Error().
			Err(err).
			Str("full_path", fullPath).
			Msg("Failed to write file content")
		return nil, fmt.Errorf("%s: %w", util.ErrMakeFile, err)
	}

	if written == 0 {
		fr.logger.Warn().
			Str("full_path", fullPath).
			Msg("File is empty")
		return nil, fmt.Errorf("%s", util.ErrEmptyFile)
	}

	// Create MediaFile entity
	mediaFile := entity.NewMediaFile(filename, "", path)
	mediaFile.CompleteDownload(written)

	fr.logger.Info().
		Str("filename", filename).
		Int64("size", written).
		Msg("File saved successfully")

	return mediaFile, nil
}

// FileExists checks if a file exists
func (fr *fileRepository) FileExists(ctx context.Context, filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// GetFileInfo returns information about a file
func (fr *fileRepository) GetFileInfo(ctx context.Context, filepath string) (*entity.MediaFile, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	filename := info.Name()
	directory := filepath[:len(filepath)-len(filename)-1]

	mediaFile := entity.NewMediaFile(filename, "", directory)
	mediaFile.SetSize(info.Size())
	mediaFile.CompleteDownload(info.Size())

	// Add file system metadata
	mediaFile.AddMetadata("modified_time", info.ModTime())
	mediaFile.AddMetadata("mode", info.Mode().String())

	return mediaFile, nil
}

// DeleteFile deletes a file
func (fr *fileRepository) DeleteFile(ctx context.Context, filepath string) error {
	fr.logger.Debug().
		Str("filepath", filepath).
		Msg("Deleting file")

	if err := os.Remove(filepath); err != nil {
		fr.logger.Error().
			Err(err).
			Str("filepath", filepath).
			Msg("Failed to delete file")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fr.logger.Info().
		Str("filepath", filepath).
		Msg("File deleted successfully")

	return nil
}

// ListFiles lists all files in a directory
func (fr *fileRepository) ListFiles(ctx context.Context, directory string) ([]*entity.MediaFile, error) {
	fr.logger.Debug().
		Str("directory", directory).
		Msg("Listing files in directory")

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []*entity.MediaFile
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				fr.logger.Warn().
					Err(err).
					Str("filename", entry.Name()).
					Msg("Failed to get file info, skipping")
				continue
			}

			mediaFile := entity.NewMediaFile(entry.Name(), "", directory)
			mediaFile.SetSize(info.Size())
			mediaFile.CompleteDownload(info.Size())
			mediaFile.AddMetadata("modified_time", info.ModTime())

			files = append(files, mediaFile)
		}
	}

	fr.logger.Info().
		Str("directory", directory).
		Int("file_count", len(files)).
		Msg("Files listed successfully")

	return files, nil
}

// DeleteDirectory deletes a directory and all its contents
func (fr *fileRepository) DeleteDirectory(ctx context.Context, path string) error {
	fr.logger.Debug().
		Str("path", path).
		Msg("Deleting directory")

	if err := os.RemoveAll(path); err != nil {
		fr.logger.Error().
			Err(err).
			Str("path", path).
			Msg("Failed to delete directory")
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	fr.logger.Info().
		Str("path", path).
		Msg("Directory deleted successfully")

	return nil
}

// ValidateFile validates a media file
func (fr *fileRepository) ValidateFile(ctx context.Context, file *entity.MediaFile) error {
	fullPath := file.GetFullPath()

	// Check if file exists
	exists, err := fr.FileExists(ctx, fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Check file size
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("%s: file is empty", util.ErrEmptyFile)
	}

	// Update file entity with actual size
	if file.Size != info.Size() {
		fr.logger.Warn().
			Str("filename", file.Filename).
			Int64("expected_size", file.Size).
			Int64("actual_size", info.Size()).
			Msg("File size mismatch")
		file.SetSize(info.Size())
	}

	return nil
}

// CalculateChecksum calculates MD5 checksum of a file
func (fr *fileRepository) CalculateChecksum(ctx context.Context, filepath string) (string, error) {
	fr.logger.Debug().
		Str("filepath", filepath).
		Msg("Calculating file checksum")

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	fr.logger.Debug().
		Str("filepath", filepath).
		Str("checksum", checksum).
		Msg("Checksum calculated successfully")

	return checksum, nil
}
