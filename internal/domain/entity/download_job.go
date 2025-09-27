package entity

import (
	"cheetah/util"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DownloadJob represents a complete download job with all its parameters
type DownloadJob struct {
	ID          string
	URL         string
	Folder      string
	Host        *string
	Origin      *string
	Separator   *string
	RepoDir     string
	Status      DownloadStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       error
}

// DownloadStatus represents the status of a download job
type DownloadStatus int

const (
	StatusPending DownloadStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s DownloadStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// NewDownloadJob creates a new download job
func NewDownloadJob(url, folder, repoDir string) *DownloadJob {
	return &DownloadJob{
		ID:        generateJobID(),
		URL:       url,
		Folder:    folder,
		RepoDir:   repoDir,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

// SetHost sets the host header value
func (dj *DownloadJob) SetHost(host string) {
	if host != "" {
		dj.Host = &host
	}
}

// SetOrigin sets the origin header value
func (dj *DownloadJob) SetOrigin(origin string) {
	if origin != "" {
		dj.Origin = &origin
	}
}

// SetSeparator sets the file name separator
func (dj *DownloadJob) SetSeparator(separator string) {
	if separator != "" {
		dj.Separator = &separator
	}
}

// Start marks the job as started
func (dj *DownloadJob) Start() {
	dj.Status = StatusRunning
	now := time.Now()
	dj.StartedAt = &now
}

// Complete marks the job as completed
func (dj *DownloadJob) Complete() {
	dj.Status = StatusCompleted
	now := time.Now()
	dj.CompletedAt = &now
}

// Fail marks the job as failed with an error
func (dj *DownloadJob) Fail(err error) {
	dj.Status = StatusFailed
	dj.Error = err
	now := time.Now()
	dj.CompletedAt = &now
}

// Cancel marks the job as cancelled
func (dj *DownloadJob) Cancel() {
	dj.Status = StatusCancelled
	now := time.Now()
	dj.CompletedAt = &now
}

// GetFileSequence creates a FileSequence from this download job
func (dj *DownloadJob) GetFileSequence() (*FileSequence, error) {
	extension, err := extractExtensionFromURL(dj.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", util.ErrInvalidURL, err)
	}

	return NewFileSequence(dj.URL, extension, dj.Separator), nil
}

// GetSavePath returns the full save path for downloaded files
func (dj *DownloadJob) GetSavePath() string {
	return fmt.Sprintf("%s/%s", dj.RepoDir, dj.Folder)
}

// Validate validates the download job parameters
func (dj *DownloadJob) Validate() error {
	if dj.URL == "" {
		return util.NewValidationError("url", dj.URL, "URL is required")
	}

	if dj.Folder == "" {
		return util.NewValidationError("folder", dj.Folder, "Folder is required")
	}

	if dj.RepoDir == "" {
		return util.NewValidationError("repo_dir", dj.RepoDir, "Repository directory is required")
	}

	// Validate URL format
	if !isValidURL(dj.URL) {
		return util.NewValidationError("url", dj.URL, "Invalid URL format")
	}

	return nil
}

// Helper functions
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func extractExtensionFromURL(url string) (string, error) {
	r := regexp.MustCompile(`\.(\w+)$|\.(\w+)\?`)
	matches := r.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("cannot extract file extension from URL: %s", url)
	}

	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i], nil
		}
	}

	return "", fmt.Errorf("cannot extract file extension from URL: %s", url)
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
