package mocks

import (
	"cheetah/internal/domain/repository"
	"context"
	"sync"
)

// MediaProcessorMock is a mock implementation of MediaProcessor
type MediaProcessorMock struct {
	mu sync.RWMutex

	// Call tracking
	processMediaFilesCalls []ProcessMediaFilesCall
	extractMetadataCalls   []ExtractMetadataCall
	generateChecksumCalls  []GenerateChecksumCall
	cleanupTempFilesCalls  []CleanupTempFilesCall

	// Behavior configuration
	processMediaFilesError error
	extractMetadataResult  *repository.MediaMetadata
	extractMetadataError   error
	generateChecksumResult string
	generateChecksumError  error
	cleanupTempFilesError  error
}

// Call structs for tracking
type ProcessMediaFilesCall struct {
	Ctx         context.Context
	InputFiles  []string
	OutputFile  string
	ProcessType string
}

type ExtractMetadataCall struct {
	Ctx      context.Context
	FilePath string
}

type GenerateChecksumCall struct {
	Ctx      context.Context
	FilePath string
}

type CleanupTempFilesCall struct {
	Ctx       context.Context
	FilePaths []string
}

// NewMediaProcessorMock creates a new media processor mock
func NewMediaProcessorMock() *MediaProcessorMock {
	return &MediaProcessorMock{}
}

// ProcessMediaFiles mocks processing media files
func (m *MediaProcessorMock) ProcessMediaFiles(ctx context.Context, inputFiles []string, outputFile, processType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processMediaFilesCalls = append(m.processMediaFilesCalls, ProcessMediaFilesCall{
		Ctx:         ctx,
		InputFiles:  inputFiles,
		OutputFile:  outputFile,
		ProcessType: processType,
	})

	return m.processMediaFilesError
}

// ExtractMetadata mocks extracting metadata from a file
func (m *MediaProcessorMock) ExtractMetadata(ctx context.Context, filePath string) (*repository.MediaMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.extractMetadataCalls = append(m.extractMetadataCalls, ExtractMetadataCall{
		Ctx:      ctx,
		FilePath: filePath,
	})

	if m.extractMetadataError != nil {
		return nil, m.extractMetadataError
	}

	if m.extractMetadataResult != nil {
		return m.extractMetadataResult, nil
	}

	// Default metadata
	return &repository.MediaMetadata{
		Duration:   "00:10:00",
		Resolution: "1920x1080",
		Bitrate:    "5000kbps",
		Format:     "mp4",
		Size:       1024 * 1024, // 1MB
	}, nil
}

// GenerateChecksum mocks generating a checksum for a file
func (m *MediaProcessorMock) GenerateChecksum(ctx context.Context, filePath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.generateChecksumCalls = append(m.generateChecksumCalls, GenerateChecksumCall{
		Ctx:      ctx,
		FilePath: filePath,
	})

	if m.generateChecksumError != nil {
		return "", m.generateChecksumError
	}

	if m.generateChecksumResult != "" {
		return m.generateChecksumResult, nil
	}

	// Default checksum
	return "abcdef1234567890", nil
}

// CleanupTempFiles mocks cleaning up temporary files
func (m *MediaProcessorMock) CleanupTempFiles(ctx context.Context, filePaths []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupTempFilesCalls = append(m.cleanupTempFilesCalls, CleanupTempFilesCall{
		Ctx:       ctx,
		FilePaths: filePaths,
	})

	return m.cleanupTempFilesError
}

// Mock behavior configuration methods
func (m *MediaProcessorMock) SetProcessMediaFilesError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processMediaFilesError = err
}

func (m *MediaProcessorMock) SetExtractMetadataResult(metadata *repository.MediaMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extractMetadataResult = metadata
}

func (m *MediaProcessorMock) SetExtractMetadataError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extractMetadataError = err
}

func (m *MediaProcessorMock) SetGenerateChecksumResult(checksum string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.generateChecksumResult = checksum
}

func (m *MediaProcessorMock) SetGenerateChecksumError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.generateChecksumError = err
}

func (m *MediaProcessorMock) SetCleanupTempFilesError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupTempFilesError = err
}

// Call verification methods
func (m *MediaProcessorMock) GetProcessMediaFilesCalls() []ProcessMediaFilesCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ProcessMediaFilesCall{}, m.processMediaFilesCalls...)
}

func (m *MediaProcessorMock) GetExtractMetadataCalls() []ExtractMetadataCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ExtractMetadataCall{}, m.extractMetadataCalls...)
}

func (m *MediaProcessorMock) GetGenerateChecksumCalls() []GenerateChecksumCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]GenerateChecksumCall{}, m.generateChecksumCalls...)
}

func (m *MediaProcessorMock) GetCleanupTempFilesCalls() []CleanupTempFilesCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]CleanupTempFilesCall{}, m.cleanupTempFilesCalls...)
}

// Reset clears all call history and mock data
func (m *MediaProcessorMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processMediaFilesCalls = nil
	m.extractMetadataCalls = nil
	m.generateChecksumCalls = nil
	m.cleanupTempFilesCalls = nil

	m.processMediaFilesError = nil
	m.extractMetadataResult = nil
	m.extractMetadataError = nil
	m.generateChecksumResult = ""
	m.generateChecksumError = nil
	m.cleanupTempFilesError = nil
}

// Helper methods for testing
func (m *MediaProcessorMock) WasProcessMediaFilesCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processMediaFilesCalls) > 0
}

func (m *MediaProcessorMock) WasExtractMetadataCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.extractMetadataCalls) > 0
}

func (m *MediaProcessorMock) WasGenerateChecksumCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.generateChecksumCalls) > 0
}

func (m *MediaProcessorMock) WasCleanupTempFilesCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cleanupTempFilesCalls) > 0
}

func (m *MediaProcessorMock) ProcessMediaFilesCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processMediaFilesCalls)
}

func (m *MediaProcessorMock) ExtractMetadataCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.extractMetadataCalls)
}

func (m *MediaProcessorMock) GenerateChecksumCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.generateChecksumCalls)
}

func (m *MediaProcessorMock) CleanupTempFilesCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cleanupTempFilesCalls)
}

// Verify interface compliance
var _ repository.MediaProcessor = (*MediaProcessorMock)(nil)
