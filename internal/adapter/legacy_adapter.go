package adapter

import (
	"cheetah/config"
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/repository"
	repoImpl "cheetah/internal/repository"
	"cheetah/model"
	"cheetah/pkg/logger"
	"context"
	"io"

	"github.com/rs/zerolog"
)

// LegacyFileAdapter adapts the new repository pattern to work with legacy code
type LegacyFileAdapter struct {
	fileRepo       repository.FileRepository
	mediaProcessor repository.MediaProcessor
	config         *config.Config
	logger         zerolog.Logger
}

// NewLegacyFileAdapter creates a new legacy file adapter
func NewLegacyFileAdapter(cfg *config.Config) *LegacyFileAdapter {
	factory := repoImpl.NewFactory(cfg)

	return &LegacyFileAdapter{
		fileRepo:       factory.CreateFileRepository(),
		mediaProcessor: factory.CreateMediaProcessor(),
		config:         cfg,
		logger:         logger.GetLogger("legacy_file_adapter"),
	}
}

// MakeDirectory creates a directory using the new repository
func (lfa *LegacyFileAdapter) MakeDirectory(path string) error {
	ctx := context.Background()
	return lfa.fileRepo.CreateDirectory(ctx, path)
}

// MakeFile saves a file using the new repository
func (lfa *LegacyFileAdapter) MakeFile(filename string, body io.ReadCloser, directory string) error {
	ctx := context.Background()
	defer body.Close()

	mediaFile, err := lfa.fileRepo.SaveFile(ctx, filename, body, directory)
	if err != nil {
		return err
	}

	lfa.logger.Debug().
		Str("filename", mediaFile.Filename).
		Int64("size", mediaFile.Size).
		Msg("File saved via legacy adapter")

	return nil
}

// StartCmd processes files using the new media processor
func (lfa *LegacyFileAdapter) StartCmd(folderPath, folderName, extension string) error {
	ctx := context.Background()
	return lfa.mediaProcessor.ProcessFiles(ctx, folderPath, folderName, extension)
}

// LegacyClientAdapter adapts the new HTTP client to work with legacy code
type LegacyClientAdapter struct {
	httpClient repository.HTTPClient
	httpRepo   repository.HTTPRepository
	logger     zerolog.Logger
}

// NewLegacyClientAdapter creates a new legacy client adapter
func NewLegacyClientAdapter(cfg *config.Config) *LegacyClientAdapter {
	factory := repoImpl.NewFactory(cfg)

	return &LegacyClientAdapter{
		httpClient: factory.CreateHTTPClient(),
		httpRepo:   factory.CreateHTTPRepository(),
		logger:     logger.GetLogger("legacy_client_adapter"),
	}
}

// Do performs an HTTP request using the new HTTP client
func (lca *LegacyClientAdapter) Do(url string, host, origin *string) (*LegacyHTTPResponse, error) {
	ctx := context.Background()

	// Build headers
	var hostStr, originStr string
	if host != nil {
		hostStr = *host
	} else {
		hostStr = lca.httpRepo.ExtractDomain(url)
	}
	if origin != nil {
		originStr = *origin
	} else {
		originStr = lca.httpRepo.ExtractDomain(url)
	}

	headers := lca.httpRepo.BuildHeaders(hostStr, originStr, "")

	// Perform request
	resp, err := lca.httpClient.Download(ctx, url, headers)
	if err != nil {
		return nil, err
	}

	// Convert to legacy response format
	return &LegacyHTTPResponse{
		StatusCode:    resp.StatusCode,
		Body:          resp.Body,
		ContentLength: resp.ContentLength,
	}, nil
}

// LegacyHTTPResponse represents the legacy HTTP response format
type LegacyHTTPResponse struct {
	StatusCode    int
	Body          io.ReadCloser
	ContentLength int64
}

// ConvertLegacyInputToDownloadJob converts legacy input to new domain entity
func ConvertLegacyInputToDownloadJob(input *model.Input, cfg *config.Config) (*entity.DownloadJob, error) {
	legacyInput := &entity.LegacyInput{
		URL:       input.URL,
		Host:      input.Host,
		Origin:    input.Origin,
		Folder:    input.Folder,
		Separator: input.Separator,
	}

	return legacyInput.ToDownloadJob(cfg)
}

// ConvertLegacyFileToEntities converts legacy file to new domain entities
func ConvertLegacyFileToEntities(file *model.File) (*entity.DownloadJob, *entity.FileSequence, error) {
	legacyFile := &entity.LegacyFile{
		Repo:      file.Repo,
		URL:       file.URL,
		Separator: file.Separator,
		Extension: file.Extension,
		Folder:    file.Folder,
	}

	job, err := legacyFile.ToDownloadJob()
	if err != nil {
		return nil, nil, err
	}

	sequence, err := legacyFile.ToFileSequence()
	if err != nil {
		return nil, nil, err
	}

	return job, sequence, nil
}
