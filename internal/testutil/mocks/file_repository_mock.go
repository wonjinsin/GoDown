package mocks

import (
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	"context"
	"io"
	"sync"
)

// FileRepositoryMock is a mock implementation of FileRepository
type FileRepositoryMock struct {
	mu sync.RWMutex

	// Mock data
	directories map[string]bool
	files       map[string]*entity.MediaFile

	// Call tracking
	createDirectoryCalls []CreateDirectoryCall
	saveFileCalls        []SaveFileCall
	fileExistsCalls      []FileExistsCall
	deleteFileCalls      []DeleteFileCall
	listFilesCalls       []ListFilesCall
	validateFileCalls    []ValidateFileCall

	// Behavior configuration
	createDirectoryError error
	saveFileResult       *entity.MediaFile
	saveFileError        error
	fileExistsResult     bool
	deleteFileError      error
	listFilesResult      []*entity.MediaFile
	listFilesError       error
	validateFileError    error
	checksumResult       string
	checksumError        error
}

// Call structs for tracking
type CreateDirectoryCall struct {
	Ctx  context.Context
	Path string
}

type SaveFileCall struct {
	Ctx      context.Context
	Filename string
	Content  io.Reader
	Path     string
}

type FileExistsCall struct {
	Ctx      context.Context
	FilePath string
}

type DeleteFileCall struct {
	Ctx      context.Context
	FilePath string
}

type ListFilesCall struct {
	Ctx       context.Context
	Directory string
}

type ValidateFileCall struct {
	Ctx  context.Context
	File *entity.MediaFile
}

// NewFileRepositoryMock creates a new file repository mock
func NewFileRepositoryMock() *FileRepositoryMock {
	return &FileRepositoryMock{
		directories: make(map[string]bool),
		files:       make(map[string]*entity.MediaFile),
	}
}

// CreateDirectory mocks creating a directory
func (m *FileRepositoryMock) CreateDirectory(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createDirectoryCalls = append(m.createDirectoryCalls, CreateDirectoryCall{
		Ctx:  ctx,
		Path: path,
	})

	if m.createDirectoryError != nil {
		return m.createDirectoryError
	}

	m.directories[path] = true
	return nil
}

// DirectoryExists mocks checking if directory exists
func (m *FileRepositoryMock) DirectoryExists(ctx context.Context, path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exists, found := m.directories[path]
	return found && exists, nil
}

// SaveFile mocks saving a file
func (m *FileRepositoryMock) SaveFile(ctx context.Context, filename string, content io.Reader, path string) (*entity.MediaFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveFileCalls = append(m.saveFileCalls, SaveFileCall{
		Ctx:      ctx,
		Filename: filename,
		Content:  content,
		Path:     path,
	})

	if m.saveFileError != nil {
		return nil, m.saveFileError
	}

	if m.saveFileResult != nil {
		return m.saveFileResult, nil
	}

	// Default result
	file := &entity.MediaFile{
		Filename: filename,
		Size:     1024,
		Status:   entity.FileStatusCompleted,
	}
	m.files[filename] = file
	return file, nil
}

// FileExists mocks checking if a file exists
func (m *FileRepositoryMock) FileExists(ctx context.Context, filePath string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fileExistsCalls = append(m.fileExistsCalls, FileExistsCall{
		Ctx:      ctx,
		FilePath: filePath,
	})

	if m.fileExistsResult {
		return m.fileExistsResult, nil
	}

	_, exists := m.files[filePath]
	return exists, nil
}

// GetFileInfo mocks getting file info
func (m *FileRepositoryMock) GetFileInfo(ctx context.Context, filePath string) (*entity.MediaFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if file, exists := m.files[filePath]; exists {
		return file, nil
	}

	return nil, nil
}

// DeleteFile mocks deleting a file
func (m *FileRepositoryMock) DeleteFile(ctx context.Context, filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deleteFileCalls = append(m.deleteFileCalls, DeleteFileCall{
		Ctx:      ctx,
		FilePath: filePath,
	})

	if m.deleteFileError != nil {
		return m.deleteFileError
	}

	delete(m.files, filePath)
	return nil
}

// ListFiles mocks listing files in a directory
func (m *FileRepositoryMock) ListFiles(ctx context.Context, directory string) ([]*entity.MediaFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listFilesCalls = append(m.listFilesCalls, ListFilesCall{
		Ctx:       ctx,
		Directory: directory,
	})

	if m.listFilesError != nil {
		return nil, m.listFilesError
	}

	if m.listFilesResult != nil {
		return m.listFilesResult, nil
	}

	var files []*entity.MediaFile
	for _, file := range m.files {
		files = append(files, file)
	}

	return files, nil
}

// DeleteDirectory mocks deleting a directory
func (m *FileRepositoryMock) DeleteDirectory(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.directories, path)
	return nil
}

// ValidateFile mocks validating a file
func (m *FileRepositoryMock) ValidateFile(ctx context.Context, file *entity.MediaFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validateFileCalls = append(m.validateFileCalls, ValidateFileCall{
		Ctx:  ctx,
		File: file,
	})

	return m.validateFileError
}

// CalculateChecksum mocks calculating file checksum
func (m *FileRepositoryMock) CalculateChecksum(ctx context.Context, filePath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.checksumError != nil {
		return "", m.checksumError
	}

	if m.checksumResult != "" {
		return m.checksumResult, nil
	}

	return "mock-checksum", nil
}

// Mock behavior configuration methods
func (m *FileRepositoryMock) SetCreateDirectoryError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createDirectoryError = err
}

func (m *FileRepositoryMock) SetSaveFileResult(file *entity.MediaFile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveFileResult = file
}

func (m *FileRepositoryMock) SetSaveFileError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveFileError = err
}

func (m *FileRepositoryMock) SetFileExistsResult(exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fileExistsResult = exists
}

func (m *FileRepositoryMock) SetDeleteFileError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteFileError = err
}

func (m *FileRepositoryMock) SetListFilesResult(files []*entity.MediaFile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listFilesResult = files
}

func (m *FileRepositoryMock) SetListFilesError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listFilesError = err
}

func (m *FileRepositoryMock) SetValidateFileError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validateFileError = err
}

func (m *FileRepositoryMock) SetChecksumResult(checksum string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checksumResult = checksum
}

func (m *FileRepositoryMock) SetChecksumError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checksumError = err
}

// Call verification methods
func (m *FileRepositoryMock) GetCreateDirectoryCalls() []CreateDirectoryCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]CreateDirectoryCall{}, m.createDirectoryCalls...)
}

func (m *FileRepositoryMock) GetSaveFileCalls() []SaveFileCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]SaveFileCall{}, m.saveFileCalls...)
}

func (m *FileRepositoryMock) GetFileExistsCalls() []FileExistsCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]FileExistsCall{}, m.fileExistsCalls...)
}

func (m *FileRepositoryMock) GetDeleteFileCalls() []DeleteFileCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]DeleteFileCall{}, m.deleteFileCalls...)
}

func (m *FileRepositoryMock) GetListFilesCalls() []ListFilesCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ListFilesCall{}, m.listFilesCalls...)
}

func (m *FileRepositoryMock) GetValidateFileCalls() []ValidateFileCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ValidateFileCall{}, m.validateFileCalls...)
}

// Reset clears all call history and mock data
func (m *FileRepositoryMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.directories = make(map[string]bool)
	m.files = make(map[string]*entity.MediaFile)

	m.createDirectoryCalls = nil
	m.saveFileCalls = nil
	m.fileExistsCalls = nil
	m.deleteFileCalls = nil
	m.listFilesCalls = nil
	m.validateFileCalls = nil

	m.createDirectoryError = nil
	m.saveFileResult = nil
	m.saveFileError = nil
	m.fileExistsResult = false
	m.deleteFileError = nil
	m.listFilesResult = nil
	m.listFilesError = nil
	m.validateFileError = nil
	m.checksumResult = ""
	m.checksumError = nil
}

// Verify interface compliance
var _ repository.FileRepository = (*FileRepositoryMock)(nil)
