package download

import (
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/service"
	"cheetah/internal/testutil"
	"cheetah/internal/testutil/mocks"
	"context"
	"errors"
	"testing"
	"time"
)

func TestDownloadUseCase_CreateDownloadJob(t *testing.T) {
	tests := []testutil.TableTest[service.CreateDownloadJobRequest]{
		{
			Name: "Valid request",
			Input: service.CreateDownloadJobRequest{
				URL:    "https://example.com/file{}.jpg",
				Folder: "test-folder",
			},
			Expected: "test-folder",
			Error:    false,
		},
		{
			Name: "Empty URL",
			Input: service.CreateDownloadJobRequest{
				URL:    "",
				Folder: "test-folder",
			},
			Expected: nil,
			Error:    true,
		},
		{
			Name: "Empty folder",
			Input: service.CreateDownloadJobRequest{
				URL:    "https://example.com/file{}.jpg",
				Folder: "",
			},
			Expected: nil,
			Error:    true,
		},
	}

	testutil.RunTableTests(t, tests, func(t *testing.T, input service.CreateDownloadJobRequest) (interface{}, error) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		job, err := useCase.CreateDownloadJob(ctx, input)
		if err != nil {
			return nil, err
		}

		return job.Folder, nil
	})
}

func TestDownloadUseCase_StartDownload(t *testing.T) {
	t.Run("Successful download", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create test job
		job := testutil.CreateTestDownloadJob()

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := useCase.StartDownload(ctx, job)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, entity.StatusCompleted, job.Status)

		// Verify repository calls
		createDirCalls := fileRepo.GetCreateDirectoryCalls()
		testutil.AssertTrue(t, len(createDirCalls) > 0)
	})

	t.Run("Download with context cancellation", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		// Configure slow download to test cancellation
		httpClient.SetDownloadProgress([]mocks.ProgressUpdate{
			{Downloaded: 0, Total: 1000},
			{Downloaded: 100, Total: 1000},
		})

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create test job
		job := testutil.CreateTestDownloadJob()

		// Execute with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := useCase.StartDownload(ctx, job)

		// Verify cancellation
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "cancelled")
	})

	t.Run("Download with file repository error", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		// Configure repository to fail
		fileRepo.SetCreateDirectoryError(errors.New("directory creation failed"))

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create test job
		job := testutil.CreateTestDownloadJob()

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := useCase.StartDownload(ctx, job)

		// Verify error handling
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, entity.StatusFailed, job.Status)
		testutil.AssertContains(t, err.Error(), "failed to create directory")
	})
}

func TestDownloadUseCase_CancelDownload(t *testing.T) {
	t.Run("Cancel existing download", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create and start a job
		job := testutil.CreateTestDownloadJob()
		jobID := job.ID

		// Start download in background
		ctx, cancel := testutil.TestContext()
		defer cancel()

		go func() {
			useCase.StartDownload(ctx, job)
		}()

		// Wait a bit to ensure download has started
		time.Sleep(10 * time.Millisecond)

		// Cancel the download
		err := useCase.CancelDownload(ctx, jobID)

		// Verify
		testutil.AssertNoError(t, err)

		// Check progress was updated
		progress, err := useCase.GetDownloadProgress(ctx, jobID)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "cancelled", progress.Status)
	})
}

func TestDownloadUseCase_GetDownloadProgress(t *testing.T) {
	t.Run("Get progress for existing job", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create test job
		job := testutil.CreateTestDownloadJob()
		jobID := job.ID

		// Start download to initialize progress
		ctx, cancel := testutil.TestContext()
		defer cancel()

		go func() {
			useCase.StartDownload(ctx, job)
		}()

		// Wait for progress to be initialized
		time.Sleep(10 * time.Millisecond)

		// Get progress
		progress, err := useCase.GetDownloadProgress(ctx, jobID)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, "", progress.Status)
	})

	t.Run("Get progress for non-existent job", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		progress, err := useCase.GetDownloadProgress(ctx, "non-existent-job")

		// Verify
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, (*service.DownloadProgress)(nil), progress)
	})
}

func TestDownloadUseCase_ValidateDownloadRequest(t *testing.T) {
	tests := []testutil.TableTest[service.CreateDownloadJobRequest]{
		{
			Name: "Valid request",
			Input: service.CreateDownloadJobRequest{
				URL:    "https://example.com/file{}.jpg",
				Folder: "test-folder",
			},
			Expected: true,
			Error:    false,
		},
		{
			Name: "Invalid URL",
			Input: service.CreateDownloadJobRequest{
				URL:    "not-a-url",
				Folder: "test-folder",
			},
			Expected: false,
			Error:    true,
		},
		{
			Name: "Empty folder",
			Input: service.CreateDownloadJobRequest{
				URL:    "https://example.com/file{}.jpg",
				Folder: "",
			},
			Expected: false,
			Error:    true,
		},
	}

	testutil.RunTableTests(t, tests, func(t *testing.T, input service.CreateDownloadJobRequest) (interface{}, error) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := useCase.ValidateDownloadRequest(ctx, input)
		if err != nil {
			return false, err
		}

		return true, nil
	})
}

func TestDownloadUseCase_EstimateDownloadSize(t *testing.T) {
	t.Run("Estimate size for valid job", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		fileRepo := mocks.NewFileRepositoryMock()
		httpClient := mocks.NewHTTPClientMock()
		httpRepo := mocks.NewHTTPRepositoryMock()
		mediaProcessor := mocks.NewMediaProcessorMock()

		useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

		// Create test job
		job := testutil.CreateTestDownloadJob()

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		size, err := useCase.EstimateDownloadSize(ctx, job)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertTrue(t, size > 0)
	})
}

// Benchmark tests
func BenchmarkDownloadUseCase_CreateDownloadJob(b *testing.B) {
	cfg := testutil.TestConfig()
	fileRepo := mocks.NewFileRepositoryMock()
	httpClient := mocks.NewHTTPClientMock()
	httpRepo := mocks.NewHTTPRepositoryMock()
	mediaProcessor := mocks.NewMediaProcessorMock()

	useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

	request := testutil.CreateTestDownloadJobRequest()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := useCase.CreateDownloadJob(ctx, request)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkDownloadUseCase_GetDownloadProgress(b *testing.B) {
	cfg := testutil.TestConfig()
	fileRepo := mocks.NewFileRepositoryMock()
	httpClient := mocks.NewHTTPClientMock()
	httpRepo := mocks.NewHTTPRepositoryMock()
	mediaProcessor := mocks.NewMediaProcessorMock()

	useCase := NewDownloadUseCase(fileRepo, httpClient, httpRepo, mediaProcessor, cfg)

	// Create and initialize a job
	job := testutil.CreateTestDownloadJob()
	ctx := context.Background()

	// Initialize progress
	go useCase.StartDownload(ctx, job)
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := useCase.GetDownloadProgress(ctx, job.ID)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
