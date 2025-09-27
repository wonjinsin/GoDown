package usecase

import (
	"cheetah/config"
	"cheetah/internal/domain/repository"
	"cheetah/internal/domain/service"
	"cheetah/internal/usecase/download"
	"cheetah/internal/usecase/filesequence"
	"cheetah/internal/usecase/media"
)

// Factory provides methods to create use case instances
type Factory struct {
	config         *config.Config
	fileRepo       repository.FileRepository
	httpClient     repository.HTTPClient
	httpRepo       repository.HTTPRepository
	mediaProcessor repository.MediaProcessor
}

// NewFactory creates a new use case factory
func NewFactory(
	cfg *config.Config,
	fileRepo repository.FileRepository,
	httpClient repository.HTTPClient,
	httpRepo repository.HTTPRepository,
	mediaProcessor repository.MediaProcessor,
) *Factory {
	return &Factory{
		config:         cfg,
		fileRepo:       fileRepo,
		httpClient:     httpClient,
		httpRepo:       httpRepo,
		mediaProcessor: mediaProcessor,
	}
}

// CreateDownloadUseCase creates a new download use case
func (f *Factory) CreateDownloadUseCase() service.DownloadService {
	return download.NewDownloadUseCase(
		f.fileRepo,
		f.httpClient,
		f.httpRepo,
		f.mediaProcessor,
		f.config,
	)
}

// CreateFileSequenceUseCase creates a new file sequence use case
func (f *Factory) CreateFileSequenceUseCase() service.FileSequenceService {
	return filesequence.NewFileSequenceUseCase(f.httpRepo)
}

// CreateMediaProcessingUseCase creates a new media processing use case
func (f *Factory) CreateMediaProcessingUseCase() service.MediaProcessingService {
	return media.NewMediaProcessingUseCase(f.fileRepo, f.mediaProcessor)
}

// UseCases holds all use case instances
type UseCases struct {
	DownloadService        service.DownloadService
	FileSequenceService    service.FileSequenceService
	MediaProcessingService service.MediaProcessingService
}

// CreateAll creates all use case instances
func (f *Factory) CreateAll() *UseCases {
	return &UseCases{
		DownloadService:        f.CreateDownloadUseCase(),
		FileSequenceService:    f.CreateFileSequenceUseCase(),
		MediaProcessingService: f.CreateMediaProcessingUseCase(),
	}
}
