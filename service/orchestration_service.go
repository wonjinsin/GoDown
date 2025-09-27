package service

import (
	"cheetah/config"
	"cheetah/internal/domain/service"
	repoImpl "cheetah/internal/repository"
	"cheetah/internal/usecase"
	"cheetah/pkg/logger"
	"cheetah/util"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// orchestrationService implements the OrchestrationService interface
type orchestrationService struct {
	downloadUseCase        service.DownloadService
	fileSequenceUseCase    service.FileSequenceService
	mediaProcessingUseCase service.MediaProcessingService
	config                 *config.Config
	logger                 zerolog.Logger
	
	// Session management
	sessionMu sync.RWMutex
	sessions  map[string]*DownloadSession
	
	// Progress tracking
	progressMu    sync.RWMutex
	progressChans map[string][]chan ProgressUpdate
}

// NewOrchestrationService creates a new orchestration service
func NewOrchestrationService(cfg *config.Config) OrchestrationService {
	// Initialize repositories
	repoFactory := repoImpl.NewFactory(cfg)
	repos := repoFactory.CreateAll()
	
	// Initialize use cases
	usecaseFactory := usecase.NewFactory(
		cfg,
		repos.FileRepo,
		repos.HTTPClient,
		repos.HTTPRepo,
		repos.MediaProcessor,
	)
	usecases := usecaseFactory.CreateAll()
	
	return &orchestrationService{
		downloadUseCase:        usecases.DownloadService,
		fileSequenceUseCase:    usecases.FileSequenceService,
		mediaProcessingUseCase: usecases.MediaProcessingService,
		config:                 cfg,
		logger:                 logger.GetLogger("orchestration_service"),
		sessions:               make(map[string]*DownloadSession),
		progressChans:          make(map[string][]chan ProgressUpdate),
	}
}

// StartDownload initiates a download process
func (os *orchestrationService) StartDownload(ctx context.Context, request DownloadRequest) (*DownloadSession, error) {
	os.logger.Info().
		Str("url", request.URL).
		Str("folder", request.Folder).
		Msg("Starting download orchestration")

	// Validate request
	if err := os.ValidateDownloadRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create session
	session := &DownloadSession{
		ID:        generateSessionID(),
		Request:   request,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Store session
	os.sessionMu.Lock()
	os.sessions[session.ID] = session
	os.sessionMu.Unlock()

	// Start download asynchronously
	go func() {
		if err := os.executeDownload(ctx, session); err != nil {
			os.updateSessionError(session.ID, err)
			os.logger.Error().
				Err(err).
				Str("session_id", session.ID).
				Msg("Download execution failed")
		}
	}()

	os.logger.Info().
		Str("session_id", session.ID).
		Msg("Download session created")

	return session, nil
}

// GetDownloadStatus returns the current status of a download
func (os *orchestrationService) GetDownloadStatus(ctx context.Context, sessionID string) (*DownloadStatus, error) {
	os.sessionMu.RLock()
	session, exists := os.sessions[sessionID]
	os.sessionMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Get progress from use case
	progress, err := os.downloadUseCase.GetDownloadProgress(ctx, sessionID)
	if err != nil {
		// Return basic status if progress is not available
		return &DownloadStatus{
			SessionID: sessionID,
			Status:    session.Status,
			Error:     session.Error,
		}, nil
	}

	status := &DownloadStatus{
		SessionID:      sessionID,
		Status:         session.Status,
		Progress:       progress.ProgressPercent,
		TotalFiles:     progress.TotalFiles,
		CompletedFiles: progress.CompletedFiles,
		FailedFiles:    progress.FailedFiles,
		CurrentFile:    progress.CurrentFile,
		Speed:          progress.Speed,
		ETA:            time.Duration(progress.ETA) * time.Second,
		Error:          progress.Error,
	}

	return status, nil
}

// CancelDownload cancels an ongoing download
func (os *orchestrationService) CancelDownload(ctx context.Context, sessionID string) error {
	os.logger.Info().
		Str("session_id", sessionID).
		Msg("Cancelling download")

	// Update session status
	os.sessionMu.Lock()
	if session, exists := os.sessions[sessionID]; exists {
		session.Status = StatusCancelled
		now := time.Now()
		session.CompletedAt = &now
	}
	os.sessionMu.Unlock()

	// Cancel in use case
	if err := os.downloadUseCase.CancelDownload(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to cancel download: %w", err)
	}

	// Notify progress subscribers
	os.notifyProgress(sessionID, ProgressUpdate{
		SessionID: sessionID,
		Timestamp: time.Now(),
		Message:   "Download cancelled",
	})

	return nil
}

// ListActiveSessions returns all active download sessions
func (os *orchestrationService) ListActiveSessions(ctx context.Context) ([]*DownloadSession, error) {
	os.sessionMu.RLock()
	defer os.sessionMu.RUnlock()

	var activeSessions []*DownloadSession
	for _, session := range os.sessions {
		if session.Status == StatusRunning || session.Status == StatusPending {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// Subscribe subscribes to progress updates for a download session
func (os *orchestrationService) Subscribe(ctx context.Context, sessionID string) (<-chan ProgressUpdate, error) {
	os.progressMu.Lock()
	defer os.progressMu.Unlock()

	progressChan := make(chan ProgressUpdate, 100)
	
	if os.progressChans[sessionID] == nil {
		os.progressChans[sessionID] = make([]chan ProgressUpdate, 0)
	}
	os.progressChans[sessionID] = append(os.progressChans[sessionID], progressChan)

	os.logger.Debug().
		Str("session_id", sessionID).
		Msg("Progress subscription created")

	return progressChan, nil
}

// GetProgress returns the current progress of a download
func (os *orchestrationService) GetProgress(ctx context.Context, sessionID string) (*ProgressUpdate, error) {
	status, err := os.GetDownloadStatus(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return &ProgressUpdate{
		SessionID:      status.SessionID,
		Timestamp:      time.Now(),
		Progress:       status.Progress,
		TotalFiles:     status.TotalFiles,
		CompletedFiles: status.CompletedFiles,
		FailedFiles:    status.FailedFiles,
		CurrentFile:    status.CurrentFile,
		Speed:          status.Speed,
		ETA:            status.ETA,
	}, nil
}

// Unsubscribe unsubscribes from progress updates
func (os *orchestrationService) Unsubscribe(sessionID string) error {
	os.progressMu.Lock()
	defer os.progressMu.Unlock()

	if channels, exists := os.progressChans[sessionID]; exists {
		for _, ch := range channels {
			close(ch)
		}
		delete(os.progressChans, sessionID)
	}

	os.logger.Debug().
		Str("session_id", sessionID).
		Msg("Progress subscription removed")

	return nil
}

// ValidateDownloadRequest validates a download request
func (os *orchestrationService) ValidateDownloadRequest(ctx context.Context, request DownloadRequest) error {
	if request.URL == "" {
		return util.NewValidationError("url", request.URL, "URL is required")
	}

	if request.Folder == "" {
		return util.NewValidationError("folder", request.Folder, "Folder is required")
	}

	// Validate URL using use case
	usecaseRequest := service.CreateDownloadJobRequest{
		URL:       request.URL,
		Folder:    request.Folder,
		Host:      request.Host,
		Origin:    request.Origin,
		Separator: request.Separator,
	}

	return os.downloadUseCase.ValidateDownloadRequest(ctx, usecaseRequest)
}

// ValidateURL checks if a URL is valid and accessible
func (os *orchestrationService) ValidateURL(ctx context.Context, url string) error {
	// Use file sequence use case to validate URL
	_, err := os.fileSequenceUseCase.AnalyzeSequence(ctx, url)
	return err
}

// EstimateDownloadSize estimates the total download size
func (os *orchestrationService) EstimateDownloadSize(ctx context.Context, request DownloadRequest) (int64, error) {
	// Create a temporary download job for estimation
	usecaseRequest := service.CreateDownloadJobRequest{
		URL:       request.URL,
		Folder:    request.Folder,
		Host:      request.Host,
		Origin:    request.Origin,
		Separator: request.Separator,
	}

	job, err := os.downloadUseCase.CreateDownloadJob(ctx, usecaseRequest)
	if err != nil {
		return 0, err
	}

	return os.downloadUseCase.EstimateDownloadSize(ctx, job)
}

// executeDownload executes the actual download process
func (os *orchestrationService) executeDownload(ctx context.Context, session *DownloadSession) error {
	// Update session status
	os.updateSessionStatus(session.ID, StatusRunning)

	// Create download job
	usecaseRequest := service.CreateDownloadJobRequest{
		URL:       session.Request.URL,
		Folder:    session.Request.Folder,
		Host:      session.Request.Host,
		Origin:    session.Request.Origin,
		Separator: session.Request.Separator,
	}

	job, err := os.downloadUseCase.CreateDownloadJob(ctx, usecaseRequest)
	if err != nil {
		return fmt.Errorf("failed to create download job: %w", err)
	}

	// Start progress monitoring
	go os.monitorProgress(ctx, session.ID, job.ID)

	// Execute download
	if err := os.downloadUseCase.StartDownload(ctx, job); err != nil {
		os.updateSessionStatus(session.ID, StatusFailed)
		return fmt.Errorf("download failed: %w", err)
	}

	// Update session status
	os.updateSessionStatus(session.ID, StatusCompleted)

	os.logger.Info().
		Str("session_id", session.ID).
		Msg("Download completed successfully")

	return nil
}

// Helper methods
func (os *orchestrationService) updateSessionStatus(sessionID string, status SessionStatus) {
	os.sessionMu.Lock()
	defer os.sessionMu.Unlock()

	if session, exists := os.sessions[sessionID]; exists {
		session.Status = status
		if status == StatusRunning && session.StartedAt == nil {
			now := time.Now()
			session.StartedAt = &now
		}
		if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
			now := time.Now()
			session.CompletedAt = &now
		}
	}
}

func (os *orchestrationService) updateSessionError(sessionID string, err error) {
	os.sessionMu.Lock()
	defer os.sessionMu.Unlock()

	if session, exists := os.sessions[sessionID]; exists {
		session.Status = StatusFailed
		session.Error = err.Error()
		now := time.Now()
		session.CompletedAt = &now
	}
}

func (os *orchestrationService) monitorProgress(ctx context.Context, sessionID, jobID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress, err := os.downloadUseCase.GetDownloadProgress(ctx, jobID)
			if err != nil {
				continue
			}

			update := ProgressUpdate{
				SessionID:      sessionID,
				Timestamp:      time.Now(),
				Progress:       progress.ProgressPercent,
				TotalFiles:     progress.TotalFiles,
				CompletedFiles: progress.CompletedFiles,
				FailedFiles:    progress.FailedFiles,
				CurrentFile:    progress.CurrentFile,
				Speed:          progress.Speed,
				ETA:            time.Duration(progress.ETA) * time.Second,
			}

			os.notifyProgress(sessionID, update)

			// Stop monitoring if completed
			if progress.Status == "completed" || progress.Status == "failed" || progress.Status == "cancelled" {
				return
			}
		}
	}
}

func (os *orchestrationService) notifyProgress(sessionID string, update ProgressUpdate) {
	os.progressMu.RLock()
	channels, exists := os.progressChans[sessionID]
	os.progressMu.RUnlock()

	if !exists {
		return
	}

	for _, ch := range channels {
		select {
		case ch <- update:
		default:
			// Channel is full, skip this update
		}
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
