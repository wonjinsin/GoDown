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

	// Behavior configuration
	processFilesError     error
	extractMetadataResult map[string]interface{}
	extractMetadataError  error
	getDurationResult     float64
	getDurationError      error
	convertFormatError    error
	compressFileError     error
}

// Call structs for tracking
type ProcessMediaFilesCall struct {
	Ctx            context.Context
	InputDir       string
	OutputFilename string
	Extension      string
}

type ExtractMetadataCall struct {
	Ctx      context.Context
	FilePath string
}

// NewMediaProcessorMock creates a new media processor mock
func NewMediaProcessorMock() *MediaProcessorMock {
	return &MediaProcessorMock{}
}

// ProcessFiles mocks processing media files
func (m *MediaProcessorMock) ProcessFiles(ctx context.Context, inputDir, outputFilename, extension string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processMediaFilesCalls = append(m.processMediaFilesCalls, ProcessMediaFilesCall{
		Ctx:            ctx,
		InputDir:       inputDir,
		OutputFilename: outputFilename,
		Extension:      extension,
	})

	return m.processFilesError
}

// ValidateMediaFile mocks validating a media file
func (m *MediaProcessorMock) ValidateMediaFile(ctx context.Context, filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return nil
}

// ExtractMetadata mocks extracting metadata from a file
func (m *MediaProcessorMock) ExtractMetadata(ctx context.Context, filePath string) (map[string]interface{}, error) {
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
	return map[string]interface{}{
		"duration":   "00:10:00",
		"resolution": "1920x1080",
		"bitrate":    "5000kbps",
		"format":     "mp4",
		"size":       1024 * 1024, // 1MB
	}, nil
}

// GetDuration mocks getting file duration
func (m *MediaProcessorMock) GetDuration(ctx context.Context, filePath string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.getDurationError != nil {
		return 0, m.getDurationError
	}

	if m.getDurationResult != 0 {
		return m.getDurationResult, nil
	}

	return 600.0, nil // 10 minutes
}

// ConvertFormat mocks converting file format
func (m *MediaProcessorMock) ConvertFormat(ctx context.Context, inputPath, outputPath, format string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.convertFormatError
}

// CompressFile mocks compressing a file
func (m *MediaProcessorMock) CompressFile(ctx context.Context, inputPath, outputPath string, quality int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.compressFileError
}

// Mock behavior configuration methods
func (m *MediaProcessorMock) SetProcessFilesError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processFilesError = err
}

func (m *MediaProcessorMock) SetExtractMetadataResult(metadata map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extractMetadataResult = metadata
}

func (m *MediaProcessorMock) SetExtractMetadataError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extractMetadataError = err
}

func (m *MediaProcessorMock) SetGetDurationResult(duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getDurationResult = duration
}

func (m *MediaProcessorMock) SetGetDurationError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getDurationError = err
}

func (m *MediaProcessorMock) SetConvertFormatError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.convertFormatError = err
}

func (m *MediaProcessorMock) SetCompressFileError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compressFileError = err
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

// Reset clears all call history and mock data
func (m *MediaProcessorMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processMediaFilesCalls = nil
	m.extractMetadataCalls = nil

	m.processFilesError = nil
	m.extractMetadataResult = nil
	m.extractMetadataError = nil
	m.getDurationResult = 0
	m.getDurationError = nil
	m.convertFormatError = nil
	m.compressFileError = nil
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

func (m *MediaProcessorMock) ProcessFilesCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processMediaFilesCalls)
}

func (m *MediaProcessorMock) ExtractMetadataCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.extractMetadataCalls)
}

// Verify interface compliance
var _ repository.MediaProcessor = (*MediaProcessorMock)(nil)
