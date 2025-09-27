package service

import (
	"context"
	"time"
)

// DownloadService defines the interface for download operations
type DownloadService interface {
	// StartDownload initiates a download process
	StartDownload(ctx context.Context, request DownloadRequest) (*DownloadSession, error)
	
	// GetDownloadStatus returns the current status of a download
	GetDownloadStatus(ctx context.Context, sessionID string) (*DownloadStatus, error)
	
	// CancelDownload cancels an ongoing download
	CancelDownload(ctx context.Context, sessionID string) error
	
	// ListActiveSessions returns all active download sessions
	ListActiveSessions(ctx context.Context) ([]*DownloadSession, error)
}

// ProgressService defines the interface for progress tracking
type ProgressService interface {
	// Subscribe subscribes to progress updates for a download session
	Subscribe(ctx context.Context, sessionID string) (<-chan ProgressUpdate, error)
	
	// GetProgress returns the current progress of a download
	GetProgress(ctx context.Context, sessionID string) (*ProgressUpdate, error)
	
	// Unsubscribe unsubscribes from progress updates
	Unsubscribe(sessionID string) error
}

// ValidationService defines the interface for validation operations
type ValidationService interface {
	// ValidateDownloadRequest validates a download request
	ValidateDownloadRequest(ctx context.Context, request DownloadRequest) error
	
	// ValidateURL checks if a URL is valid and accessible
	ValidateURL(ctx context.Context, url string) error
	
	// EstimateDownloadSize estimates the total download size
	EstimateDownloadSize(ctx context.Context, request DownloadRequest) (int64, error)
}

// OrchestrationService defines the main orchestration interface
type OrchestrationService interface {
	DownloadService
	ProgressService
	ValidationService
}

// Request/Response types
type DownloadRequest struct {
	URL       string  `json:"url" validate:"required,url"`
	Folder    string  `json:"folder" validate:"required"`
	Host      *string `json:"host,omitempty"`
	Origin    *string `json:"origin,omitempty"`
	Separator *string `json:"separator,omitempty"`
	Options   DownloadOptions `json:"options,omitempty"`
}

type DownloadOptions struct {
	MaxConcurrency int           `json:"max_concurrency,omitempty"`
	BatchSize      int           `json:"batch_size,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	RetryCount     int           `json:"retry_count,omitempty"`
	SkipExisting   bool          `json:"skip_existing,omitempty"`
}

type DownloadSession struct {
	ID          string                 `json:"id"`
	Request     DownloadRequest        `json:"request"`
	Status      SessionStatus          `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	StartedAt   *time.Time            `json:"started_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Error       string                `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type DownloadStatus struct {
	SessionID       string        `json:"session_id"`
	Status          SessionStatus `json:"status"`
	Progress        float64       `json:"progress"`
	TotalFiles      int           `json:"total_files"`
	CompletedFiles  int           `json:"completed_files"`
	FailedFiles     int           `json:"failed_files"`
	CurrentFile     string        `json:"current_file"`
	Speed           int64         `json:"speed_bps"`
	ETA             time.Duration `json:"eta"`
	Error           string        `json:"error,omitempty"`
}

type ProgressUpdate struct {
	SessionID       string        `json:"session_id"`
	Timestamp       time.Time     `json:"timestamp"`
	Progress        float64       `json:"progress"`
	TotalFiles      int           `json:"total_files"`
	CompletedFiles  int           `json:"completed_files"`
	FailedFiles     int           `json:"failed_files"`
	CurrentFile     string        `json:"current_file"`
	Speed           int64         `json:"speed_bps"`
	ETA             time.Duration `json:"eta"`
	Message         string        `json:"message,omitempty"`
}

type SessionStatus string

const (
	StatusPending    SessionStatus = "pending"
	StatusRunning    SessionStatus = "running"
	StatusCompleted  SessionStatus = "completed"
	StatusFailed     SessionStatus = "failed"
	StatusCancelled  SessionStatus = "cancelled"
	StatusPaused     SessionStatus = "paused"
)

func (s SessionStatus) String() string {
	return string(s)
}

// Legacy interface for backward compatibility
type FileUsecase interface {
	Do(c chan int) (err error)
}
