package http

import (
	"cheetah/internal/domain/repository"
	"cheetah/pkg/logger"
	"cheetah/util"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// httpClient implements the HTTPClient interface
type httpClient struct {
	client     *http.Client
	userAgent  string
	maxRetries int
	retryDelay time.Duration
	logger     zerolog.Logger
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config repository.HTTPClientConfig) repository.HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: config.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= config.MaxRedirects {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		userAgent:  config.UserAgent,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
		logger:     logger.GetLogger("http_client"),
	}
}

// Get performs a GET request
func (hc *httpClient) Get(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	return hc.doRequest(ctx, "GET", url, headers, nil)
}

// Head performs a HEAD request
func (hc *httpClient) Head(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	return hc.doRequest(ctx, "HEAD", url, headers, nil)
}

// Download performs a GET request for downloading files
func (hc *httpClient) Download(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	return hc.Get(ctx, url, headers)
}

// DownloadWithProgress performs a GET request with progress callback
func (hc *httpClient) DownloadWithProgress(ctx context.Context, url string, headers map[string]string, progressCallback repository.ProgressCallback) (*repository.HTTPResponse, error) {
	resp, err := hc.Get(ctx, url, headers)
	if err != nil {
		return nil, err
	}

	if progressCallback != nil && resp.ContentLength > 0 {
		// Wrap the response body with progress tracking
		resp.Body = &progressReader{
			Reader:   resp.Body,
			total:    resp.ContentLength,
			callback: progressCallback,
		}
	}

	return resp, nil
}

// SetTimeout sets the HTTP client timeout
func (hc *httpClient) SetTimeout(timeout time.Duration) {
	hc.client.Timeout = timeout
}

// SetUserAgent sets the user agent string
func (hc *httpClient) SetUserAgent(userAgent string) {
	hc.userAgent = userAgent
}

// SetRetryPolicy sets the retry policy
func (hc *httpClient) SetRetryPolicy(maxRetries int, retryDelay time.Duration) {
	hc.maxRetries = maxRetries
	hc.retryDelay = retryDelay
}

// doRequest performs the actual HTTP request with retry logic
func (hc *httpClient) doRequest(ctx context.Context, method, url string, headers map[string]string, body io.Reader) (*repository.HTTPResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= hc.maxRetries; attempt++ {
		if attempt > 0 {
			hc.logger.Warn().
				Int("attempt", attempt).
				Str("url", url).
				Dur("delay", hc.retryDelay).
				Msg("Retrying HTTP request")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(hc.retryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", util.ErrHTTPRequest, err)
		}

		// Set default headers
		req.Header.Set("User-Agent", hc.userAgent)

		// Set custom headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		hc.logger.Debug().
			Str("method", method).
			Str("url", url).
			Int("attempt", attempt+1).
			Msg("Sending HTTP request")

		resp, err := hc.client.Do(req)
		if err != nil {
			lastErr = err
			hc.logger.Warn().
				Err(err).
				Str("url", url).
				Int("attempt", attempt+1).
				Msg("HTTP request failed")
			continue
		}

		// Convert to our response format
		httpResp := &repository.HTTPResponse{
			StatusCode:    resp.StatusCode,
			Headers:       make(map[string]string),
			Body:          resp.Body,
			ContentLength: resp.ContentLength,
			URL:           url,
		}

		// Copy headers
		for key, values := range resp.Header {
			if len(values) > 0 {
				httpResp.Headers[key] = values[0]
			}
		}

		hc.logger.Debug().
			Str("url", url).
			Int("status_code", resp.StatusCode).
			Int64("content_length", resp.ContentLength).
			Msg("HTTP request completed")

		return httpResp, nil
	}

	return nil, fmt.Errorf("%s: max retries exceeded: %w", util.ErrMaxRetriesExceeded, lastErr)
}

// progressReader wraps an io.Reader to track download progress
type progressReader struct {
	io.Reader
	total      int64
	downloaded int64
	callback   repository.ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.downloaded += int64(n)

	if pr.callback != nil {
		pr.callback(pr.downloaded, pr.total)
	}

	return n, err
}

func (pr *progressReader) Close() error {
	if closer, ok := pr.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
