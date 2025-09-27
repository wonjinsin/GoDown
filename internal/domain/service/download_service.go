package service

import (
	"cheetah/internal/domain/entity"
	"context"
)

// DownloadService defines the core business logic for downloads
type DownloadService interface {
	// Job management
	CreateDownloadJob(ctx context.Context, request CreateDownloadJobRequest) (*entity.DownloadJob, error)
	StartDownload(ctx context.Context, job *entity.DownloadJob) error
	CancelDownload(ctx context.Context, jobID string) error

	// Progress tracking
	GetDownloadProgress(ctx context.Context, jobID string) (*DownloadProgress, error)
	GetDownloadStatus(ctx context.Context, jobID string) (*entity.DownloadJob, error)

	// File operations
	ValidateDownloadRequest(ctx context.Context, request CreateDownloadJobRequest) error
	EstimateDownloadSize(ctx context.Context, job *entity.DownloadJob) (int64, error)
}

// FileSequenceService handles file sequence operations
type FileSequenceService interface {
	// Sequence operations
	AnalyzeSequence(ctx context.Context, baseURL string) (*entity.FileSequence, error)
	GenerateFileList(ctx context.Context, sequence *entity.FileSequence, count int) ([]*entity.MediaFile, error)
	ValidateSequence(ctx context.Context, sequence *entity.FileSequence) error

	// Pattern detection
	DetectPattern(ctx context.Context, url string) (*entity.FilePattern, error)
	EstimateFileCount(ctx context.Context, sequence *entity.FileSequence) (int, error)
}

// MediaProcessingService handles media file processing
type MediaProcessingService interface {
	// Processing operations
	ProcessDownloadedFiles(ctx context.Context, job *entity.DownloadJob) error
	ValidateMediaFiles(ctx context.Context, files []*entity.MediaFile) error

	// Metadata operations
	ExtractFileMetadata(ctx context.Context, file *entity.MediaFile) error
	GenerateChecksums(ctx context.Context, files []*entity.MediaFile) error

	// Cleanup operations
	CleanupFailedDownloads(ctx context.Context, job *entity.DownloadJob) error
	CleanupTemporaryFiles(ctx context.Context, directory string) error
}

// Request/Response types
type CreateDownloadJobRequest struct {
	URL       string  `json:"url" validate:"required,url"`
	Folder    string  `json:"folder" validate:"required"`
	Host      *string `json:"host,omitempty"`
	Origin    *string `json:"origin,omitempty"`
	Separator *string `json:"separator,omitempty"`
}

type DownloadProgress struct {
	JobID           string  `json:"job_id"`
	TotalFiles      int     `json:"total_files"`
	CompletedFiles  int     `json:"completed_files"`
	FailedFiles     int     `json:"failed_files"`
	TotalBytes      int64   `json:"total_bytes"`
	DownloadedBytes int64   `json:"downloaded_bytes"`
	ProgressPercent float64 `json:"progress_percent"`
	Speed           int64   `json:"speed_bps"` // bytes per second
	ETA             int64   `json:"eta_seconds"`
	Status          string  `json:"status"`
	CurrentFile     string  `json:"current_file"`
	Error           string  `json:"error,omitempty"`
}
