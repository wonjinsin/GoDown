package mocks

import (
	"cheetah/internal/domain/repository"
	"context"
	"io"
	"strings"
	"sync"
	"time"
)

// HTTPClientMock is a mock implementation of HTTPClient
type HTTPClientMock struct {
	mu sync.RWMutex

	// Call tracking
	getCalls      []GetCall
	headCalls     []HeadCall
	downloadCalls []DownloadCall

	// Behavior configuration
	getResponse      *repository.HTTPResponse
	getError         error
	headResponse     *repository.HTTPResponse
	headError        error
	downloadError    error
	downloadProgress []ProgressUpdate
	timeout          time.Duration
}

// Call structs for tracking
type GetCall struct {
	Ctx     context.Context
	URL     string
	Headers map[string]string
}

type HeadCall struct {
	Ctx     context.Context
	URL     string
	Headers map[string]string
}

type DownloadCall struct {
	Ctx              context.Context
	URL              string
	Headers          map[string]string
	ProgressCallback repository.ProgressCallback
}

type ProgressUpdate struct {
	Downloaded int64
	Total      int64
}

// NewHTTPClientMock creates a new HTTP client mock
func NewHTTPClientMock() *HTTPClientMock {
	return &HTTPClientMock{
		timeout: 30 * time.Second,
	}
}

// Get mocks HTTP GET request
func (m *HTTPClientMock) Get(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getCalls = append(m.getCalls, GetCall{
		Ctx:     ctx,
		URL:     url,
		Headers: headers,
	})

	if m.getError != nil {
		return nil, m.getError
	}

	if m.getResponse != nil {
		return m.getResponse, nil
	}

	// Default response
	return &repository.HTTPResponse{
		StatusCode:    200,
		Headers:       map[string]string{"Content-Type": "text/html"},
		Body:          io.NopCloser(strings.NewReader("mock response body")),
		ContentLength: 18,
		URL:           url,
	}, nil
}

// Head mocks HTTP HEAD request
func (m *HTTPClientMock) Head(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.headCalls = append(m.headCalls, HeadCall{
		Ctx:     ctx,
		URL:     url,
		Headers: headers,
	})

	if m.headError != nil {
		return nil, m.headError
	}

	if m.headResponse != nil {
		return m.headResponse, nil
	}

	// Default response
	return &repository.HTTPResponse{
		StatusCode:    200,
		Headers:       map[string]string{"Content-Length": "1024"},
		Body:          nil,
		ContentLength: 1024,
		URL:           url,
	}, nil
}

// Download mocks file download
func (m *HTTPClientMock) Download(ctx context.Context, url string, headers map[string]string) (*repository.HTTPResponse, error) {
	return m.DownloadWithProgress(ctx, url, headers, nil)
}

// DownloadWithProgress mocks file download with progress tracking
func (m *HTTPClientMock) DownloadWithProgress(ctx context.Context, url string, headers map[string]string, progressCallback repository.ProgressCallback) (*repository.HTTPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.downloadCalls = append(m.downloadCalls, DownloadCall{
		Ctx:              ctx,
		URL:              url,
		Headers:          headers,
		ProgressCallback: progressCallback,
	})

	if m.downloadError != nil {
		return nil, m.downloadError
	}

	// Simulate progress updates if callback is provided
	if progressCallback != nil && len(m.downloadProgress) > 0 {
		go func() {
			for _, update := range m.downloadProgress {
				select {
				case <-ctx.Done():
					return
				default:
					progressCallback(update.Downloaded, update.Total)
					time.Sleep(10 * time.Millisecond) // Simulate download time
				}
			}
		}()
	}

	// Default response
	return &repository.HTTPResponse{
		StatusCode:    200,
		Headers:       map[string]string{"Content-Type": "application/octet-stream"},
		Body:          io.NopCloser(strings.NewReader("mock file content")),
		ContentLength: 17,
		URL:           url,
	}, nil
}

// SetTimeout mocks setting timeout
func (m *HTTPClientMock) SetTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeout = timeout
}

// SetUserAgent sets the user agent string
func (m *HTTPClientMock) SetUserAgent(userAgent string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Mock implementation - just store it
}

// SetRetryPolicy sets the retry policy
func (m *HTTPClientMock) SetRetryPolicy(maxRetries int, retryDelay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Mock implementation - just store it
}

// Mock behavior configuration methods
func (m *HTTPClientMock) SetGetResponse(response *repository.HTTPResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getResponse = response
}

func (m *HTTPClientMock) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

func (m *HTTPClientMock) SetHeadResponse(response *repository.HTTPResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.headResponse = response
}

func (m *HTTPClientMock) SetHeadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.headError = err
}

func (m *HTTPClientMock) SetDownloadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.downloadError = err
}

func (m *HTTPClientMock) SetDownloadProgress(progress []ProgressUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.downloadProgress = progress
}

// Call verification methods
func (m *HTTPClientMock) GetGetCalls() []GetCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]GetCall{}, m.getCalls...)
}

func (m *HTTPClientMock) GetHeadCalls() []HeadCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]HeadCall{}, m.headCalls...)
}

func (m *HTTPClientMock) GetDownloadCalls() []DownloadCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]DownloadCall{}, m.downloadCalls...)
}

func (m *HTTPClientMock) GetTimeout() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timeout
}

// Reset clears all call history and mock data
func (m *HTTPClientMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getCalls = nil
	m.headCalls = nil
	m.downloadCalls = nil

	m.getResponse = nil
	m.getError = nil
	m.headResponse = nil
	m.headError = nil
	m.downloadError = nil
	m.downloadProgress = nil
	m.timeout = 30 * time.Second
}

// Helper methods for testing
func (m *HTTPClientMock) WasGetCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.getCalls) > 0
}

func (m *HTTPClientMock) WasHeadCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.headCalls) > 0
}

func (m *HTTPClientMock) WasDownloadCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.downloadCalls) > 0
}

func (m *HTTPClientMock) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.getCalls)
}

func (m *HTTPClientMock) HeadCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.headCalls)
}

func (m *HTTPClientMock) DownloadCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.downloadCalls)
}

// Verify interface compliance
var _ repository.HTTPClient = (*HTTPClientMock)(nil)
