package service

import (
	"cheetah/internal/testutil"
	"context"
	"testing"
	"time"
)

func TestOrchestrationService_StartDownload(t *testing.T) {
	t.Run("Successful download", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
			Options: DownloadOptions{
				MaxConcurrency: 5,
				BatchSize:      10,
				Timeout:        30 * time.Second,
				RetryCount:     3,
			},
		}

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session, err := service.StartDownload(ctx, request)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, "", session.ID)
		testutil.AssertEqual(t, StatusPending, session.Status)
		testutil.AssertEqual(t, request.URL, session.Request.URL)
		testutil.AssertEqual(t, request.Folder, session.Request.Folder)
	})

	t.Run("Download with validation error", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "", // Invalid URL
			Folder: "test-folder",
		}

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session, err := service.StartDownload(ctx, request)

		// Verify
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, (*DownloadSession)(nil), session)
		testutil.AssertContains(t, err.Error(), "validation failed")
	})
}

func TestOrchestrationService_GetDownloadStatus(t *testing.T) {
	t.Run("Get status for existing session", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Start a download
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session, err := service.StartDownload(ctx, request)
		testutil.AssertNoError(t, err)

		// Wait a bit for the download to start
		time.Sleep(50 * time.Millisecond)

		// Get status
		status, err := service.GetDownloadStatus(ctx, session.ID)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, session.ID, status.SessionID)
		testutil.AssertTrue(t, status.Status == StatusPending || status.Status == StatusRunning)
	})

	t.Run("Get status for non-existent session", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		status, err := service.GetDownloadStatus(ctx, "non-existent-session")

		// Verify
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, (*DownloadStatus)(nil), status)
		testutil.AssertContains(t, err.Error(), "session not found")
	})
}

func TestOrchestrationService_CancelDownload(t *testing.T) {
	t.Run("Cancel existing download", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Start a download
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session, err := service.StartDownload(ctx, request)
		testutil.AssertNoError(t, err)

		// Cancel the download
		err = service.CancelDownload(ctx, session.ID)

		// Verify
		testutil.AssertNoError(t, err)

		// Check that status was updated
		testutil.WaitForCondition(t, func() bool {
			status, err := service.GetDownloadStatus(ctx, session.ID)
			return err == nil && status.Status == StatusCancelled
		}, 1*time.Second, "download should be cancelled")
	})
}

func TestOrchestrationService_ListActiveSessions(t *testing.T) {
	t.Run("List sessions with no active sessions", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		sessions, err := service.ListActiveSessions(ctx)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, len(sessions))
	})

	t.Run("List sessions with active sessions", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Start downloads
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session1, err := service.StartDownload(ctx, request)
		testutil.AssertNoError(t, err)

		session2, err := service.StartDownload(ctx, request)
		testutil.AssertNoError(t, err)

		// List active sessions
		sessions, err := service.ListActiveSessions(ctx)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertTrue(t, len(sessions) >= 2)

		// Check that our sessions are in the list
		found1, found2 := false, false
		for _, session := range sessions {
			if session.ID == session1.ID {
				found1 = true
			}
			if session.ID == session2.ID {
				found2 = true
			}
		}
		testutil.AssertTrue(t, found1)
		testutil.AssertTrue(t, found2)
	})
}

func TestOrchestrationService_ProgressTracking(t *testing.T) {
	t.Run("Subscribe to progress updates", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Start a download
		ctx, cancel := testutil.TestContext()
		defer cancel()

		session, err := service.StartDownload(ctx, request)
		testutil.AssertNoError(t, err)

		// Subscribe to progress
		progressChan, err := service.Subscribe(ctx, session.ID)
		testutil.AssertNoError(t, err)

		// Wait for at least one progress update
		select {
		case progress := <-progressChan:
			testutil.AssertEqual(t, session.ID, progress.SessionID)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected to receive progress update within 2 seconds")
		}

		// Unsubscribe
		err = service.Unsubscribe(session.ID)
		testutil.AssertNoError(t, err)
	})
}

func TestOrchestrationService_ValidationService(t *testing.T) {
	t.Run("Validate valid request", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := service.ValidateDownloadRequest(ctx, request)

		// Verify
		testutil.AssertNoError(t, err)
	})

	t.Run("Validate invalid request", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "", // Invalid URL
			Folder: "test-folder",
		}

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := service.ValidateDownloadRequest(ctx, request)

		// Verify
		testutil.AssertError(t, err)
	})

	t.Run("Validate URL", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		err := service.ValidateURL(ctx, "https://example.com/file{}.jpg")

		// Verify
		testutil.AssertNoError(t, err)
	})

	t.Run("Estimate download size", func(t *testing.T) {
		// Setup
		cfg := testutil.TestConfig()
		service := NewOrchestrationService(cfg)

		request := DownloadRequest{
			URL:    "https://example.com/file{}.jpg",
			Folder: "test-folder",
		}

		// Execute
		ctx, cancel := testutil.TestContext()
		defer cancel()

		size, err := service.EstimateDownloadSize(ctx, request)

		// Verify
		testutil.AssertNoError(t, err)
		testutil.AssertTrue(t, size > 0)
	})
}

// Integration test with mock dependencies
func TestOrchestrationService_WithMocks(t *testing.T) {
	t.Run("Integration test with mocked repositories", func(t *testing.T) {
		// This test would require refactoring OrchestrationService to accept
		// injected dependencies, which is a good practice for testing.
		// For now, we'll skip this as it would require significant changes
		// to the production code.
		t.Skip("Requires dependency injection refactoring")
	})
}

// Benchmark tests
func BenchmarkOrchestrationService_StartDownload(b *testing.B) {
	cfg := testutil.TestConfig()
	service := NewOrchestrationService(cfg)

	request := DownloadRequest{
		URL:    "https://example.com/file{}.jpg",
		Folder: "test-folder",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session, err := service.StartDownload(ctx, request)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}

		// Cancel immediately to avoid resource buildup
		service.CancelDownload(ctx, session.ID)
	}
}

func BenchmarkOrchestrationService_GetDownloadStatus(b *testing.B) {
	cfg := testutil.TestConfig()
	service := NewOrchestrationService(cfg)

	request := DownloadRequest{
		URL:    "https://example.com/file{}.jpg",
		Folder: "test-folder",
	}

	ctx := context.Background()

	// Create a session for testing
	session, err := service.StartDownload(ctx, request)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetDownloadStatus(ctx, session.ID)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
