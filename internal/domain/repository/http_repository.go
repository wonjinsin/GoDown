package repository

import (
	"context"
	"io"
	"time"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	// Basic HTTP operations
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Head(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)

	// Download operations
	Download(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	DownloadWithProgress(ctx context.Context, url string, headers map[string]string, progressCallback ProgressCallback) (*HTTPResponse, error)

	// Configuration
	SetTimeout(timeout time.Duration)
	SetUserAgent(userAgent string)
	SetRetryPolicy(maxRetries int, retryDelay time.Duration)
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode    int
	Headers       map[string]string
	Body          io.ReadCloser
	ContentLength int64
	URL           string
}

// ProgressCallback is called during download to report progress
type ProgressCallback func(downloaded, total int64)

// HTTPRepository defines repository-level HTTP operations
type HTTPRepository interface {
	// Client management
	CreateClient(config HTTPClientConfig) HTTPClient

	// Utility operations
	ValidateURL(url string) error
	ExtractDomain(url string) string
	BuildHeaders(host, origin, userAgent string) map[string]string
}

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	UserAgent     string
	MaxRedirects  int
	EnableCookies bool
}
