package util

import (
	"errors"
	"fmt"
)

// Error constants for different failure scenarios
const (
	ErrMakeFile           = "file creation failed"
	ErrInvalidURL         = "invalid URL format"
	ErrInvalidPath        = "invalid file path"
	ErrInvalidFileName    = "invalid file name pattern"
	ErrHTTPRequest        = "HTTP request failed"
	ErrFileDownload       = "file download failed"
	ErrDirectoryCreate    = "directory creation failed"
	ErrFFmpegExecution    = "FFmpeg execution failed"
	ErrEmptyFile          = "downloaded file is empty"
	ErrMaxRetriesExceeded = "maximum retries exceeded"
)

// Custom error types
var (
	ErrNetworkTimeout    = errors.New("network timeout")
	ErrInsufficientSpace = errors.New("insufficient disk space")
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrCancelled         = errors.New("operation cancelled")
)

// DownloadError represents a download-specific error
type DownloadError struct {
	URL      string
	Filename string
	Err      error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("download failed for %s (file: %s): %v", e.URL, e.Filename, e.Err)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

// NewDownloadError creates a new download error
func NewDownloadError(url, filename string, err error) *DownloadError {
	return &DownloadError{
		URL:      url,
		Filename: filename,
		Err:      err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}
