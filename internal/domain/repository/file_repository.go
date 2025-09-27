package repository

import (
	"cheetah/internal/domain/entity"
	"context"
	"io"
)

// FileRepository defines the interface for file storage operations
type FileRepository interface {
	// Directory operations
	CreateDirectory(ctx context.Context, path string) error
	DirectoryExists(ctx context.Context, path string) (bool, error)

	// File operations
	SaveFile(ctx context.Context, filename string, content io.Reader, path string) (*entity.MediaFile, error)
	FileExists(ctx context.Context, filepath string) (bool, error)
	GetFileInfo(ctx context.Context, filepath string) (*entity.MediaFile, error)
	DeleteFile(ctx context.Context, filepath string) error

	// Batch operations
	ListFiles(ctx context.Context, directory string) ([]*entity.MediaFile, error)
	DeleteDirectory(ctx context.Context, path string) error

	// File validation
	ValidateFile(ctx context.Context, file *entity.MediaFile) error
	CalculateChecksum(ctx context.Context, filepath string) (string, error)
}

// MediaProcessor defines the interface for media processing operations
type MediaProcessor interface {
	// Processing operations
	ProcessFiles(ctx context.Context, inputDir, outputFilename, extension string) error
	ValidateMediaFile(ctx context.Context, filepath string) error

	// Metadata operations
	ExtractMetadata(ctx context.Context, filepath string) (map[string]interface{}, error)
	GetDuration(ctx context.Context, filepath string) (float64, error)

	// Conversion operations
	ConvertFormat(ctx context.Context, inputPath, outputPath, format string) error
	CompressFile(ctx context.Context, inputPath, outputPath string, quality int) error
}
