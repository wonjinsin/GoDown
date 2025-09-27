package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Collector collects and manages application metrics
type Collector struct {
	mu     sync.RWMutex
	logger zerolog.Logger

	// Download metrics
	downloadMetrics     map[string]*DownloadMetrics
	globalDownloadStats *GlobalDownloadStats

	// System metrics
	systemMetrics *SystemMetrics

	// Error metrics
	errorMetrics *ErrorMetrics

	// HTTP metrics
	httpMetrics *HTTPMetrics
}

// DownloadMetrics represents metrics for a specific download session
type DownloadMetrics struct {
	SessionID       string
	StartTime       time.Time
	EndTime         *time.Time
	TotalFiles      int
	CompletedFiles  int
	FailedFiles     int
	TotalBytes      int64
	DownloadedBytes int64
	AverageSpeed    float64 // bytes per second
	CurrentSpeed    float64 // bytes per second
	LastUpdateTime  time.Time
	Status          string
}

// GlobalDownloadStats represents aggregate download statistics
type GlobalDownloadStats struct {
	TotalSessions        int64
	CompletedSessions    int64
	FailedSessions       int64
	CancelledSessions    int64
	TotalFilesDownloaded int64
	TotalBytesDownloaded int64
	AverageSessionTime   time.Duration
}

// SystemMetrics represents system resource metrics
type SystemMetrics struct {
	MemoryUsage    int64 // bytes
	GoroutineCount int
	CPUUsage       float64 // percentage
	LastUpdated    time.Time
}

// ErrorMetrics represents error tracking metrics
type ErrorMetrics struct {
	TotalErrors      int64
	HTTPErrors       int64
	FileSystemErrors int64
	NetworkErrors    int64
	ValidationErrors int64
	ErrorRate        float64 // errors per minute
	LastErrorTime    time.Time
	ErrorsByType     map[string]int64
}

// HTTPMetrics represents HTTP request metrics
type HTTPMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	RequestsByStatus   map[int]int64
	LastRequestTime    time.Time
}

// NewCollector creates a new metrics collector
func NewCollector(logger zerolog.Logger) *Collector {
	return &Collector{
		logger:              logger,
		downloadMetrics:     make(map[string]*DownloadMetrics),
		globalDownloadStats: &GlobalDownloadStats{},
		systemMetrics:       &SystemMetrics{},
		errorMetrics: &ErrorMetrics{
			ErrorsByType: make(map[string]int64),
		},
		httpMetrics: &HTTPMetrics{
			RequestsByStatus: make(map[int]int64),
		},
	}
}

// StartDownloadSession starts tracking metrics for a download session
func (c *Collector) StartDownloadSession(sessionID string, totalFiles int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.downloadMetrics[sessionID] = &DownloadMetrics{
		SessionID:      sessionID,
		StartTime:      time.Now(),
		TotalFiles:     totalFiles,
		Status:         "running",
		LastUpdateTime: time.Now(),
	}

	c.globalDownloadStats.TotalSessions++

	c.logger.Debug().
		Str("session_id", sessionID).
		Int("total_files", totalFiles).
		Msg("Started tracking download session metrics")
}

// UpdateDownloadProgress updates download progress metrics
func (c *Collector) UpdateDownloadProgress(sessionID string, completedFiles, failedFiles int, downloadedBytes int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, exists := c.downloadMetrics[sessionID]
	if !exists {
		c.logger.Warn().
			Str("session_id", sessionID).
			Msg("Attempted to update metrics for unknown session")
		return
	}

	now := time.Now()
	timeDiff := now.Sub(metrics.LastUpdateTime).Seconds()

	if timeDiff > 0 {
		bytesDiff := downloadedBytes - metrics.DownloadedBytes
		metrics.CurrentSpeed = float64(bytesDiff) / timeDiff

		// Update average speed
		totalTime := now.Sub(metrics.StartTime).Seconds()
		if totalTime > 0 {
			metrics.AverageSpeed = float64(downloadedBytes) / totalTime
		}
	}

	metrics.CompletedFiles = completedFiles
	metrics.FailedFiles = failedFiles
	metrics.DownloadedBytes = downloadedBytes
	metrics.LastUpdateTime = now

	c.logger.Debug().
		Str("session_id", sessionID).
		Int("completed_files", completedFiles).
		Int("failed_files", failedFiles).
		Int64("downloaded_bytes", downloadedBytes).
		Float64("current_speed", metrics.CurrentSpeed).
		Msg("Updated download progress metrics")
}

// CompleteDownloadSession marks a download session as completed
func (c *Collector) CompleteDownloadSession(sessionID string, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, exists := c.downloadMetrics[sessionID]
	if !exists {
		return
	}

	now := time.Now()
	metrics.EndTime = &now

	if success {
		metrics.Status = "completed"
		c.globalDownloadStats.CompletedSessions++
		c.globalDownloadStats.TotalFilesDownloaded += int64(metrics.CompletedFiles)
		c.globalDownloadStats.TotalBytesDownloaded += metrics.DownloadedBytes
	} else {
		metrics.Status = "failed"
		c.globalDownloadStats.FailedSessions++
	}

	// Update average session time
	sessionDuration := now.Sub(metrics.StartTime)
	totalSessions := c.globalDownloadStats.CompletedSessions + c.globalDownloadStats.FailedSessions
	if totalSessions > 0 {
		c.globalDownloadStats.AverageSessionTime = time.Duration(
			(int64(c.globalDownloadStats.AverageSessionTime)*totalSessions + int64(sessionDuration)) / totalSessions,
		)
	}

	c.logger.Info().
		Str("session_id", sessionID).
		Bool("success", success).
		Dur("duration", sessionDuration).
		Msg("Completed download session metrics")
}

// CancelDownloadSession marks a download session as cancelled
func (c *Collector) CancelDownloadSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, exists := c.downloadMetrics[sessionID]
	if !exists {
		return
	}

	now := time.Now()
	metrics.EndTime = &now
	metrics.Status = "cancelled"
	c.globalDownloadStats.CancelledSessions++

	c.logger.Info().
		Str("session_id", sessionID).
		Msg("Cancelled download session")
}

// RecordError records an error occurrence
func (c *Collector) RecordError(errorType string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errorMetrics.TotalErrors++
	c.errorMetrics.ErrorsByType[errorType]++
	c.errorMetrics.LastErrorTime = time.Now()

	// Calculate error rate (errors per minute)
	// This is a simple implementation - in production you'd want a sliding window
	if c.errorMetrics.TotalErrors > 1 {
		duration := time.Since(c.errorMetrics.LastErrorTime).Minutes()
		if duration > 0 {
			c.errorMetrics.ErrorRate = float64(c.errorMetrics.TotalErrors) / duration
		}
	}

	switch errorType {
	case "http":
		c.errorMetrics.HTTPErrors++
	case "filesystem":
		c.errorMetrics.FileSystemErrors++
	case "network":
		c.errorMetrics.NetworkErrors++
	case "validation":
		c.errorMetrics.ValidationErrors++
	}

	c.logger.Warn().
		Err(err).
		Str("error_type", errorType).
		Int64("total_errors", c.errorMetrics.TotalErrors).
		Float64("error_rate", c.errorMetrics.ErrorRate).
		Msg("Recorded error metric")
}

// RecordHTTPRequest records HTTP request metrics
func (c *Collector) RecordHTTPRequest(statusCode int, latency time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.httpMetrics.TotalRequests++
	c.httpMetrics.RequestsByStatus[statusCode]++
	c.httpMetrics.LastRequestTime = time.Now()

	if success {
		c.httpMetrics.SuccessfulRequests++
	} else {
		c.httpMetrics.FailedRequests++
	}

	// Update average latency
	if c.httpMetrics.TotalRequests > 0 {
		c.httpMetrics.AverageLatency = time.Duration(
			(int64(c.httpMetrics.AverageLatency)*c.httpMetrics.TotalRequests + int64(latency)) / c.httpMetrics.TotalRequests,
		)
	}

	c.logger.Debug().
		Int("status_code", statusCode).
		Dur("latency", latency).
		Bool("success", success).
		Msg("Recorded HTTP request metric")
}

// UpdateSystemMetrics updates system resource metrics
func (c *Collector) UpdateSystemMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.systemMetrics.MemoryUsage = int64(m.Alloc)
	c.systemMetrics.GoroutineCount = runtime.NumGoroutine()
	c.systemMetrics.LastUpdated = time.Now()

	c.logger.Debug().
		Int64("memory_usage", c.systemMetrics.MemoryUsage).
		Int("goroutine_count", c.systemMetrics.GoroutineCount).
		Msg("Updated system metrics")
}

// GetDownloadMetrics returns metrics for a specific download session
func (c *Collector) GetDownloadMetrics(sessionID string) *DownloadMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if metrics, exists := c.downloadMetrics[sessionID]; exists {
		// Return a copy to avoid data races
		metricsCopy := *metrics
		return &metricsCopy
	}

	return nil
}

// GetGlobalStats returns global download statistics
func (c *Collector) GetGlobalStats() *GlobalDownloadStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid data races
	statsCopy := *c.globalDownloadStats
	return &statsCopy
}

// GetSystemMetrics returns current system metrics
func (c *Collector) GetSystemMetrics() *SystemMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update metrics before returning
	c.mu.RUnlock()
	c.UpdateSystemMetrics()
	c.mu.RLock()

	// Return a copy to avoid data races
	metricsCopy := *c.systemMetrics
	return &metricsCopy
}

// GetErrorMetrics returns error metrics
func (c *Collector) GetErrorMetrics() *ErrorMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid data races
	errorsCopy := *c.errorMetrics
	errorsCopy.ErrorsByType = make(map[string]int64)
	for k, v := range c.errorMetrics.ErrorsByType {
		errorsCopy.ErrorsByType[k] = v
	}

	return &errorsCopy
}

// GetHTTPMetrics returns HTTP metrics
func (c *Collector) GetHTTPMetrics() *HTTPMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid data races
	httpCopy := *c.httpMetrics
	httpCopy.RequestsByStatus = make(map[int]int64)
	for k, v := range c.httpMetrics.RequestsByStatus {
		httpCopy.RequestsByStatus[k] = v
	}

	return &httpCopy
}

// StartPeriodicSystemMetricsCollection starts collecting system metrics periodically
func (c *Collector) StartPeriodicSystemMetricsCollection(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.UpdateSystemMetrics()
			}
		}
	}()

	c.logger.Info().
		Dur("interval", interval).
		Msg("Started periodic system metrics collection")
}

// CleanupOldSessions removes metrics for old completed sessions
func (c *Collector) CleanupOldSessions(maxAge time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for sessionID, metrics := range c.downloadMetrics {
		if metrics.EndTime != nil && now.Sub(*metrics.EndTime) > maxAge {
			delete(c.downloadMetrics, sessionID)
			removed++
		}
	}

	if removed > 0 {
		c.logger.Info().
			Int("removed_sessions", removed).
			Dur("max_age", maxAge).
			Msg("Cleaned up old session metrics")
	}
}
