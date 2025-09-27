package middleware

import (
	"cheetah/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// LoggingMiddleware provides logging functionality for operations
type LoggingMiddleware struct {
	logger zerolog.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.GetLogger("middleware"),
	}
}

// LogOperation logs the start and completion of an operation
func (lm *LoggingMiddleware) LogOperation(ctx context.Context, operationName string, fn func() error) error {
	correlationID := GetCorrelationID(ctx)
	sessionID := GetSessionID(ctx)

	start := time.Now()

	lm.logger.Info().
		Str("operation", operationName).
		Str("correlation_id", correlationID).
		Str("session_id", sessionID).
		Msg("Operation started")

	err := fn()

	duration := time.Since(start)

	if err != nil {
		lm.logger.Error().
			Err(err).
			Str("operation", operationName).
			Str("correlation_id", correlationID).
			Str("session_id", sessionID).
			Dur("duration", duration).
			Msg("Operation failed")
		return err
	}

	lm.logger.Info().
		Str("operation", operationName).
		Str("correlation_id", correlationID).
		Str("session_id", sessionID).
		Dur("duration", duration).
		Msg("Operation completed")

	return nil
}

// LogHTTPRequest logs HTTP request details
func (lm *LoggingMiddleware) LogHTTPRequest(ctx context.Context, method, url string, statusCode int, duration time.Duration, size int64) {
	correlationID := GetCorrelationID(ctx)
	sessionID := GetSessionID(ctx)

	event := lm.logger.Info()

	if statusCode >= 400 {
		event = lm.logger.Warn()
	}
	if statusCode >= 500 {
		event = lm.logger.Error()
	}

	event.
		Str("type", "http_request").
		Str("method", method).
		Str("url", url).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Int64("response_size", size).
		Str("correlation_id", correlationID).
		Str("session_id", sessionID).
		Msg("HTTP request completed")
}

// LogDownloadProgress logs download progress with structured fields
func (lm *LoggingMiddleware) LogDownloadProgress(ctx context.Context, progress DownloadProgress) {
	correlationID := GetCorrelationID(ctx)
	sessionID := GetSessionID(ctx)

	lm.logger.Info().
		Str("type", "download_progress").
		Str("session_id", sessionID).
		Str("correlation_id", correlationID).
		Float64("progress_percent", progress.ProgressPercent).
		Int("total_files", progress.TotalFiles).
		Int("completed_files", progress.CompletedFiles).
		Int("failed_files", progress.FailedFiles).
		Str("current_file", progress.CurrentFile).
		Int64("download_speed", progress.Speed).
		Dur("eta", time.Duration(progress.ETA)*time.Second).
		Msg("Download progress update")
}

// LogFileOperation logs file operation with details
func (lm *LoggingMiddleware) LogFileOperation(ctx context.Context, operation, filepath string, size int64, err error) {
	correlationID := GetCorrelationID(ctx)
	sessionID := GetSessionID(ctx)

	event := lm.logger.Info()
	if err != nil {
		event = lm.logger.Error().Err(err)
	}

	event.
		Str("type", "file_operation").
		Str("operation", operation).
		Str("filepath", filepath).
		Int64("size", size).
		Str("correlation_id", correlationID).
		Str("session_id", sessionID).
		Msg("File operation")
}

// LogError logs structured error information
func (lm *LoggingMiddleware) LogError(ctx context.Context, err error, component, operation string, fields map[string]interface{}) {
	correlationID := GetCorrelationID(ctx)
	sessionID := GetSessionID(ctx)

	event := lm.logger.Error().
		Err(err).
		Str("component", component).
		Str("operation", operation).
		Str("correlation_id", correlationID).
		Str("session_id", sessionID)

	// Add custom fields
	for key, value := range fields {
		switch v := value.(type) {
		case string:
			event.Str(key, v)
		case int:
			event.Int(key, v)
		case int64:
			event.Int64(key, v)
		case float64:
			event.Float64(key, v)
		case bool:
			event.Bool(key, v)
		case time.Duration:
			event.Dur(key, v)
		default:
			event.Interface(key, v)
		}
	}

	event.Msg("Error occurred")
}

// DownloadProgress represents download progress information
type DownloadProgress struct {
	ProgressPercent float64
	TotalFiles      int
	CompletedFiles  int
	FailedFiles     int
	CurrentFile     string
	Speed           int64
	ETA             int64
}

// Context key types
type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	SessionIDKey     contextKey = "session_id"
	UserIDKey        contextKey = "user_id"
	OperationKey     contextKey = "operation"
)

// Context helper functions
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return generateCorrelationID()
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(SessionIDKey).(string); ok {
		return id
	}
	return ""
}

func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, OperationKey, operation)
}

func GetOperation(ctx context.Context) string {
	if op, ok := ctx.Value(OperationKey).(string); ok {
		return op
	}
	return ""
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	return fmt.Sprintf("corr_%d", time.Now().UnixNano())
}
