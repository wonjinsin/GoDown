package gui

import (
	"cheetah/config"
	"cheetah/pkg/logger"
	"cheetah/service"
	"context"
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/rs/zerolog"
)

// fyneHandler implements the GUIHandler interface using Fyne
type fyneHandler struct {
	app                  fyne.App
	mainWindow           fyne.Window
	orchestrationService service.OrchestrationService
	config               *config.Config
	logger               zerolog.Logger

	// Progress tracking
	progressMu    sync.RWMutex
	progressChans map[string]chan ProgressInfo

	// Window management
	windowMu sync.RWMutex
	windows  map[string]fyne.Window
}

// NewFyneHandler creates a new Fyne GUI handler
func NewFyneHandler(orchestrationService service.OrchestrationService, cfg *config.Config) GUIHandler {
	return &fyneHandler{
		app:                  app.NewWithID("goDown"),
		orchestrationService: orchestrationService,
		config:               cfg,
		logger:               logger.GetLogger("fyne_handler"),
		progressChans:        make(map[string]chan ProgressInfo),
		windows:              make(map[string]fyne.Window),
	}
}

// ShowMainWindow displays the main application window
func (fh *fyneHandler) ShowMainWindow(ctx context.Context) error {
	fh.logger.Info().Msg("Showing main window")

	fh.mainWindow = fh.app.NewWindow("GoDown")
	fh.mainWindow.Resize(fyne.NewSize(600, 400))

	content := fh.createMainContent(ctx)
	fh.mainWindow.SetContent(content)

	fh.mainWindow.ShowAndRun()
	return nil
}

// HandleDownloadRequest handles a download request from the GUI
func (fh *fyneHandler) HandleDownloadRequest(ctx context.Context, request DownloadRequest) error {
	fh.logger.Info().
		Str("url", request.URL).
		Str("folder", request.Folder).
		Msg("Handling download request")

	// Convert to service request
	serviceRequest := service.DownloadRequest{
		URL:       request.URL,
		Folder:    request.Folder,
		Host:      request.Host,
		Origin:    request.Origin,
		Separator: request.Separator,
		Options: service.DownloadOptions{
			MaxConcurrency: fh.config.Download.MaxConcurrency,
			BatchSize:      fh.config.Download.BatchSize,
			Timeout:        fh.config.Download.HTTPTimeout,
			RetryCount:     fh.config.Download.MaxRetries,
		},
	}

	// Start download
	session, err := fh.orchestrationService.StartDownload(ctx, serviceRequest)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Show progress window
	go fh.showProgressWindow(ctx, session.ID)

	// Monitor progress
	go fh.monitorProgress(ctx, session.ID)

	return nil
}

// ShowProgress displays download progress
func (fh *fyneHandler) ShowProgress(ctx context.Context, progress ProgressInfo) error {
	fh.progressMu.RLock()
	progressChan, exists := fh.progressChans[progress.SessionID]
	fh.progressMu.RUnlock()

	if exists {
		select {
		case progressChan <- progress:
		default:
			// Channel is full, skip this update
		}
	}

	return nil
}

// ShowResult displays the final result
func (fh *fyneHandler) ShowResult(ctx context.Context, result ResultInfo) error {
	fh.logger.Info().
		Str("session_id", result.SessionID).
		Bool("success", result.Success).
		Msg("Showing result")

	resultWindow := fh.app.NewWindow("Result")
	resultWindow.Resize(fyne.NewSize(300, 100))

	var message string
	var textColor color.Color

	if result.Success {
		message = fmt.Sprintf("Success: %s", result.Message)
		textColor = color.RGBA{0, 128, 0, 255} // Green
	} else {
		message = fmt.Sprintf("Failed: %s", result.Error)
		textColor = color.RGBA{255, 0, 0, 255} // Red
	}

	text := canvas.NewText(message, textColor)
	text.Alignment = fyne.TextAlignCenter
	text.TextStyle = fyne.TextStyle{Bold: true}

	content := container.New(layout.NewCenterLayout(), text)
	resultWindow.SetContent(content)
	resultWindow.Show()

	// Auto-close after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		resultWindow.Close()
	}()

	return nil
}

// Close closes the GUI handler gracefully
func (fh *fyneHandler) Close() error {
	fh.logger.Info().Msg("Starting graceful shutdown of GUI handler")

	// Cancel all active download sessions
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	activeSessions, err := fh.orchestrationService.ListActiveSessions(ctx)
	if err != nil {
		fh.logger.Error().Err(err).Msg("Failed to get active sessions during shutdown")
	} else {
		for _, session := range activeSessions {
			if err := fh.orchestrationService.CancelDownload(ctx, session.ID); err != nil {
				fh.logger.Error().
					Err(err).
					Str("session_id", session.ID).
					Msg("Failed to cancel session during shutdown")
			} else {
				fh.logger.Info().
					Str("session_id", session.ID).
					Msg("Session cancelled during shutdown")
			}
		}
	}

	// Close all progress channels
	fh.progressMu.Lock()
	for sessionID, ch := range fh.progressChans {
		close(ch)
		fh.logger.Debug().
			Str("session_id", sessionID).
			Msg("Progress channel closed")
	}
	fh.progressChans = make(map[string]chan ProgressInfo)
	fh.progressMu.Unlock()

	// Close all windows
	fh.windowMu.Lock()
	for sessionID, window := range fh.windows {
		window.Close()
		fh.logger.Debug().
			Str("session_id", sessionID).
			Msg("Window closed")
	}
	fh.windows = make(map[string]fyne.Window)
	fh.windowMu.Unlock()

	if fh.mainWindow != nil {
		fh.mainWindow.Close()
		fh.logger.Debug().Msg("Main window closed")
	}

	fh.logger.Info().Msg("GUI handler shutdown completed")
	return nil
}

// createMainContent creates the main window content
func (fh *fyneHandler) createMainContent(ctx context.Context) fyne.CanvasObject {
	// Create form fields
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL...")

	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Enter folder name...")

	separatorEntry := widget.NewEntry()
	separatorEntry.SetPlaceHolder("Optional separator...")

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("Optional host...")

	originEntry := widget.NewEntry()
	originEntry.SetPlaceHolder("Optional origin...")

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "URL", Widget: urlEntry},
			{Text: "Folder", Widget: folderEntry},
			{Text: "Separator (optional)", Widget: separatorEntry},
			{Text: "Host (optional)", Widget: hostEntry},
			{Text: "Origin (optional)", Widget: originEntry},
		},
		OnSubmit: func() {
			request := DownloadRequest{
				URL:    strings.TrimSpace(urlEntry.Text),
				Folder: strings.TrimSpace(folderEntry.Text),
			}

			if sep := strings.TrimSpace(separatorEntry.Text); sep != "" {
				request.Separator = &sep
			}
			if host := strings.TrimSpace(hostEntry.Text); host != "" {
				request.Host = &host
			}
			if origin := strings.TrimSpace(originEntry.Text); origin != "" {
				request.Origin = &origin
			}

			if err := fh.HandleDownloadRequest(ctx, request); err != nil {
				fh.showError("Download Error", err.Error())
			}
		},
	}

	return form
}

// showProgressWindow shows a progress window for a download session
func (fh *fyneHandler) showProgressWindow(ctx context.Context, sessionID string) {
	progressWindow := fh.app.NewWindow("Downloading")
	progressWindow.Resize(fyne.NewSize(400, 150))

	// Create progress widgets
	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("Preparing download...")
	fileLabel := widget.NewLabel("")
	speedLabel := widget.NewLabel("")

	content := container.NewVBox(
		statusLabel,
		progressBar,
		fileLabel,
		speedLabel,
	)

	progressWindow.SetContent(content)
	progressWindow.Show()

	// Store window reference
	fh.windowMu.Lock()
	fh.windows[sessionID] = progressWindow
	fh.windowMu.Unlock()

	// Create progress channel
	progressChan := make(chan ProgressInfo, 10)
	fh.progressMu.Lock()
	fh.progressChans[sessionID] = progressChan
	fh.progressMu.Unlock()

	// Update progress UI
	go func() {
		defer progressWindow.Close()

		for progress := range progressChan {
			progressBar.SetValue(progress.Progress / 100.0)

			statusLabel.SetText(fmt.Sprintf("Progress: %.1f%% (%d/%d files)",
				progress.Progress, progress.CompletedFiles, progress.TotalFiles))

			if progress.CurrentFile != "" {
				fileLabel.SetText(fmt.Sprintf("Current: %s", progress.CurrentFile))
			}

			if progress.Speed > 0 {
				speedLabel.SetText(fmt.Sprintf("Speed: %s/s, ETA: %s",
					formatBytes(progress.Speed), progress.ETA.String()))
			}

			// Check if completed
			if progress.Progress >= 100 {
				time.Sleep(1 * time.Second) // Show completion briefly
				return
			}
		}
	}()
}

// monitorProgress monitors download progress and updates the UI
func (fh *fyneHandler) monitorProgress(ctx context.Context, sessionID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := fh.orchestrationService.GetDownloadStatus(ctx, sessionID)
			if err != nil {
				fh.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get download status")
				continue
			}

			progress := ProgressInfo{
				SessionID:      sessionID,
				Progress:       status.Progress,
				TotalFiles:     status.TotalFiles,
				CompletedFiles: status.CompletedFiles,
				FailedFiles:    status.FailedFiles,
				CurrentFile:    status.CurrentFile,
				Speed:          status.Speed,
				ETA:            status.ETA,
			}

			if err := fh.ShowProgress(ctx, progress); err != nil {
				fh.logger.Error().Err(err).Msg("Failed to show progress")
			}

			// Check if completed
			if status.Status == service.StatusCompleted {
				result := ResultInfo{
					SessionID: sessionID,
					Success:   true,
					Message:   "Download completed successfully",
					EndTime:   time.Now(),
				}
				fh.ShowResult(ctx, result)
				return
			} else if status.Status == service.StatusFailed {
				result := ResultInfo{
					SessionID: sessionID,
					Success:   false,
					Error:     status.Error,
					EndTime:   time.Now(),
				}
				fh.ShowResult(ctx, result)
				return
			}
		}
	}
}

// showError shows an error dialog
func (fh *fyneHandler) showError(title, message string) {
	errorWindow := fh.app.NewWindow(title)
	errorWindow.Resize(fyne.NewSize(400, 150))

	text := widget.NewLabel(message)
	text.Wrapping = fyne.TextWrapWord

	okButton := widget.NewButton("OK", func() {
		errorWindow.Close()
	})

	content := container.NewVBox(
		text,
		container.New(layout.NewCenterLayout(), okButton),
	)

	errorWindow.SetContent(content)
	errorWindow.Show()
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
