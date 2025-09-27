package mocks

import (
	"cheetah/internal/domain/repository"
	"net/url"
	"sync"
)

// HTTPRepositoryMock is a mock implementation of HTTPRepository
type HTTPRepositoryMock struct {
	mu sync.RWMutex

	// Call tracking
	buildHeadersCalls  []BuildHeadersCall
	extractDomainCalls []ExtractDomainCall
	createClientCalls  []CreateClientCall
	validateURLCalls   []ValidateURLCall

	// Behavior configuration
	buildHeadersResult  map[string]string
	extractDomainResult string
	createClientResult  repository.HTTPClient
	createClientError   error
	validateURLError    error
}

// Call structs for tracking
type BuildHeadersCall struct {
	UserAgent string
	Host      *string
	Origin    *string
}

type ExtractDomainCall struct {
	URL string
}

type CreateClientCall struct {
	Config interface{}
}

type ValidateURLCall struct {
	URL string
}

// NewHTTPRepositoryMock creates a new HTTP repository mock
func NewHTTPRepositoryMock() *HTTPRepositoryMock {
	return &HTTPRepositoryMock{}
}

// BuildHeaders mocks building HTTP headers
func (m *HTTPRepositoryMock) BuildHeaders(userAgent string, host, origin *string) map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buildHeadersCalls = append(m.buildHeadersCalls, BuildHeadersCall{
		UserAgent: userAgent,
		Host:      host,
		Origin:    origin,
	})

	if m.buildHeadersResult != nil {
		return m.buildHeadersResult
	}

	// Default headers
	headers := map[string]string{
		"User-Agent": userAgent,
	}

	if host != nil {
		headers["Host"] = *host
	}

	if origin != nil {
		headers["Origin"] = *origin
	}

	return headers
}

// ExtractDomain mocks extracting domain from URL
func (m *HTTPRepositoryMock) ExtractDomain(rawURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.extractDomainCalls = append(m.extractDomainCalls, ExtractDomainCall{
		URL: rawURL,
	})

	if m.extractDomainResult != "" {
		return m.extractDomainResult, nil
	}

	// Default implementation
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return parsedURL.Host, nil
}

// CreateClient mocks creating an HTTP client
func (m *HTTPRepositoryMock) CreateClient(config interface{}) (repository.HTTPClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createClientCalls = append(m.createClientCalls, CreateClientCall{
		Config: config,
	})

	if m.createClientError != nil {
		return nil, m.createClientError
	}

	if m.createClientResult != nil {
		return m.createClientResult, nil
	}

	// Return a new mock client
	return NewHTTPClientMock(), nil
}

// ValidateURL mocks URL validation
func (m *HTTPRepositoryMock) ValidateURL(rawURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validateURLCalls = append(m.validateURLCalls, ValidateURLCall{
		URL: rawURL,
	})

	if m.validateURLError != nil {
		return m.validateURLError
	}

	// Default validation
	_, err := url.Parse(rawURL)
	return err
}

// Mock behavior configuration methods
func (m *HTTPRepositoryMock) SetBuildHeadersResult(headers map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buildHeadersResult = headers
}

func (m *HTTPRepositoryMock) SetExtractDomainResult(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extractDomainResult = domain
}

func (m *HTTPRepositoryMock) SetCreateClientResult(client repository.HTTPClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createClientResult = client
}

func (m *HTTPRepositoryMock) SetCreateClientError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createClientError = err
}

func (m *HTTPRepositoryMock) SetValidateURLError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validateURLError = err
}

// Call verification methods
func (m *HTTPRepositoryMock) GetBuildHeadersCalls() []BuildHeadersCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]BuildHeadersCall{}, m.buildHeadersCalls...)
}

func (m *HTTPRepositoryMock) GetExtractDomainCalls() []ExtractDomainCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ExtractDomainCall{}, m.extractDomainCalls...)
}

func (m *HTTPRepositoryMock) GetCreateClientCalls() []CreateClientCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]CreateClientCall{}, m.createClientCalls...)
}

func (m *HTTPRepositoryMock) GetValidateURLCalls() []ValidateURLCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ValidateURLCall{}, m.validateURLCalls...)
}

// Reset clears all call history and mock data
func (m *HTTPRepositoryMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buildHeadersCalls = nil
	m.extractDomainCalls = nil
	m.createClientCalls = nil
	m.validateURLCalls = nil

	m.buildHeadersResult = nil
	m.extractDomainResult = ""
	m.createClientResult = nil
	m.createClientError = nil
	m.validateURLError = nil
}

// Helper methods for testing
func (m *HTTPRepositoryMock) WasBuildHeadersCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.buildHeadersCalls) > 0
}

func (m *HTTPRepositoryMock) WasExtractDomainCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.extractDomainCalls) > 0
}

func (m *HTTPRepositoryMock) WasCreateClientCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.createClientCalls) > 0
}

func (m *HTTPRepositoryMock) WasValidateURLCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.validateURLCalls) > 0
}

// Verify interface compliance
var _ repository.HTTPRepository = (*HTTPRepositoryMock)(nil)
