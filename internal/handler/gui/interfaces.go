package gui

import (
	"context"
	"time"
)

// GUIHandler defines the interface for GUI operations
type GUIHandler interface {
	// ShowMainWindow displays the main application window
	ShowMainWindow(ctx context.Context) error

	// HandleDownloadRequest handles a download request from the GUI
	HandleDownloadRequest(ctx context.Context, request DownloadRequest) error

	// ShowProgress displays download progress
	ShowProgress(ctx context.Context, progress ProgressInfo) error

	// ShowResult displays the final result
	ShowResult(ctx context.Context, result ResultInfo) error

	// Close closes the GUI handler
	Close() error
}

// EventHandler defines the interface for handling GUI events
type EventHandler interface {
	// OnDownloadStart is called when a download starts
	OnDownloadStart(ctx context.Context, sessionID string)

	// OnDownloadProgress is called during download progress
	OnDownloadProgress(ctx context.Context, sessionID string, progress ProgressInfo)

	// OnDownloadComplete is called when a download completes
	OnDownloadComplete(ctx context.Context, sessionID string, result ResultInfo)

	// OnDownloadError is called when a download fails
	OnDownloadError(ctx context.Context, sessionID string, err error)
}

// ProgressReporter defines the interface for progress reporting
type ProgressReporter interface {
	// ReportProgress reports download progress
	ReportProgress(ctx context.Context, progress ProgressInfo) error

	// Subscribe subscribes to progress updates
	Subscribe(ctx context.Context, sessionID string) (<-chan ProgressInfo, error)

	// Unsubscribe unsubscribes from progress updates
	Unsubscribe(sessionID string) error
}

// WindowManager defines the interface for window management
type WindowManager interface {
	// CreateWindow creates a new window
	CreateWindow(title string, size WindowSize) Window

	// ShowWindow shows a window
	ShowWindow(window Window) error

	// HideWindow hides a window
	HideWindow(window Window) error

	// CloseWindow closes a window
	CloseWindow(window Window) error
}

// Window represents a GUI window
type Window interface {
	// SetContent sets the window content
	SetContent(content interface{})

	// Show shows the window
	Show()

	// Hide hides the window
	Hide()

	// Close closes the window
	Close()

	// SetTitle sets the window title
	SetTitle(title string)

	// Resize resizes the window
	Resize(size WindowSize)
}

// Data structures
type DownloadRequest struct {
	URL       string  `json:"url"`
	Folder    string  `json:"folder"`
	Host      *string `json:"host,omitempty"`
	Origin    *string `json:"origin,omitempty"`
	Separator *string `json:"separator,omitempty"`
}

type ProgressInfo struct {
	SessionID      string        `json:"session_id"`
	Progress       float64       `json:"progress"`
	TotalFiles     int           `json:"total_files"`
	CompletedFiles int           `json:"completed_files"`
	FailedFiles    int           `json:"failed_files"`
	CurrentFile    string        `json:"current_file"`
	Speed          int64         `json:"speed_bps"`
	ETA            time.Duration `json:"eta"`
	Message        string        `json:"message,omitempty"`
}

type ResultInfo struct {
	SessionID string        `json:"session_id"`
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
