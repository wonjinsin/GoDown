package service

import (
	"cheetah/config"
	"cheetah/model"
	"cheetah/pkg/logger"
	"cheetah/util"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
)

// FileService ...
type FileService struct {
	Input  *model.Input
	File   *model.File
	Config *config.Config
	logger zerolog.Logger
}

// NewFileService ...
func NewFileService(input *model.Input, cfg *config.Config) FileUsecase {
	fs := &FileService{
		Input:  input,
		File:   model.MakeFile(input, cfg),
		Config: cfg,
		logger: logger.GetLogger("file_service"),
	}
	return fs
}

// Do ...
func (t *FileService) Do(c chan int) (err error) {
	t.logger.Info().
		Str("folder", t.File.Folder).
		Str("repo_dir", t.File.Repo).
		Msg("Starting file download process")

	if err = t.File.MakeDirectory(); err != nil {
		t.logger.Error().
			Err(err).
			Str("directory", fmt.Sprintf("%s/%s", t.File.Repo, t.File.Folder)).
			Msg("Failed to create directory")
		return fmt.Errorf("%s: %w", util.ErrDirectoryCreate, err)
	}

	startNum := 0
	batchCount := t.Config.Download.BatchSize
	errCount := 0
	errMax := t.Config.Download.MaxErrors
	routineErrMax := t.Config.Download.RoutineErrMax

	t.logger.Info().
		Int("batch_size", batchCount).
		Int("max_errors", errMax).
		Int("routine_err_max", routineErrMax).
		Msg("Starting batch download process")

	for {
		if errCount > errMax {
			t.logger.Warn().
				Int("error_count", errCount).
				Int("max_errors", errMax).
				Msg("Maximum errors reached, stopping download")
			break
		}

		var wg sync.WaitGroup
		wg.Add(batchCount)

		for j := 0; j < batchCount; j++ {
			go func(num uint64, wg *sync.WaitGroup) {
				defer wg.Done()
				url, err := t.File.GetReplacedURL(uint64(num))
				if err != nil {
					t.logger.Error().
						Err(err).
						Uint64("file_number", num).
						Msg("Failed to generate URL for file")
					errCount++
					return
				}

				filename := fmt.Sprintf("%d.%s", num, t.File.GetExtension())
				for {
					errCnt := 0
					if err := t.DownloadFile(url, filename); err != nil {
						var downloadErr *util.DownloadError
						if errors.As(err, &downloadErr) && errCnt < routineErrMax {
							t.logger.Warn().
								Err(err).
								Str("url", url).
								Str("filename", filename).
								Int("retry_count", errCnt).
								Msg("Download failed, retrying")
							errCnt++
							continue
						}
						t.logger.Error().
							Err(err).
							Str("url", url).
							Str("filename", filename).
							Msg("Download failed after retries")
						errCount++
					}
					return
				}
			}(uint64(startNum+j), &wg)
		}

		startNum += batchCount
		c <- int(startNum)
		wg.Wait()
	}

	t.logger.Info().Msg("Starting FFmpeg processing")
	if err = t.File.StartCmd(); err != nil {
		t.logger.Error().
			Err(err).
			Str("folder", t.File.Folder).
			Str("extension", t.File.Extension).
			Msg("FFmpeg processing failed")
		return fmt.Errorf("%s: %w", util.ErrFFmpegExecution, err)
	}

	t.logger.Info().Msg("Download and processing completed successfully")
	return nil
}

// DownloadFile ...
func (t *FileService) DownloadFile(url string, filename string) error {
	t.logger.Debug().
		Str("filename", filename).
		Str("url", url).
		Msg("Starting file download")

	client, err := model.MakeClient(url, t.Input.Host, t.Input.Origin, t.Config)
	if err != nil {
		return util.NewDownloadError(url, filename, fmt.Errorf("%s: %w", util.ErrHTTPRequest, err))
	}

	resp, err := client.Do()
	defer func(resp *http.Response) {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}(resp)

	if err != nil {
		return util.NewDownloadError(url, filename, fmt.Errorf("%s: %w", util.ErrHTTPRequest, err))
	}
	if resp.StatusCode != 200 {
		t.logger.Warn().
			Str("url", url).
			Int("status_code", resp.StatusCode).
			Msg("HTTP request returned non-200 status code")
		return util.NewDownloadError(url, filename, fmt.Errorf("%s: status code %d", util.ErrHTTPRequest, resp.StatusCode))
	}

	// if resp.ContentLength <= 0 {
	// 	if err := t.File.MakeFile(filename, resp.Body); err != nil {
	// 		fmt.Println("Request failed: Content length is not valid", "url", url, "length", resp.ContentLength)
	// 		return fmt.Errorf("Request failed: Content length is not valid")
	// 	}
	// } else {
	// 	if err := t.File.MakeFile(filename, resp.Body); err != nil {
	// 		fmt.Printf("Error occurred: %s", err.Error())
	// 		return errors.New("MakeFile failed")
	// 	}
	// }

	if err := t.File.MakeFile(filename, resp.Body); err != nil {
		t.logger.Error().
			Err(err).
			Str("filename", filename).
			Msg("Failed to save downloaded file")
		return util.NewDownloadError(url, filename, fmt.Errorf("%s: %w", util.ErrMakeFile, err))
	}

	t.logger.Debug().
		Str("filename", filename).
		Int64("content_length", resp.ContentLength).
		Msg("File downloaded successfully")
	return nil
}
