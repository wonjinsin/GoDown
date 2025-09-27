package entity

import (
	"testing"
)

func TestNewDownloadJob(t *testing.T) {
	job := NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")

	if job.URL != "https://example.com/file{}.jpg" {
		t.Errorf("expected URL to be set")
	}
	if job.Folder != "test-folder" {
		t.Errorf("expected Folder to be set")
	}
	if job.Status != StatusPending {
		t.Errorf("expected StatusPending, got %v", job.Status)
	}
}

func TestDownloadJob_Start(t *testing.T) {
	job := NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")

	if job.Status != StatusPending {
		t.Errorf("expected StatusPending, got %v", job.Status)
	}

	job.Start()

	if job.Status != StatusRunning {
		t.Errorf("expected StatusRunning, got %v", job.Status)
	}
	if job.StartedAt == nil {
		t.Error("expected StartedAt to be set")
	}
}

func TestDownloadJob_Complete(t *testing.T) {
	job := NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")
	job.Start()

	job.Complete()

	if job.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", job.Status)
	}
	if job.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestDownloadJob_GetSavePath(t *testing.T) {
	job := NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")
	path := job.GetSavePath()

	expected := "test-repo/test-folder"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestDownloadStatus_String(t *testing.T) {
	tests := []struct {
		status   DownloadStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status.String())
		}
	}
}

// Benchmark tests
func BenchmarkNewDownloadJob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")
	}
}

func BenchmarkDownloadJob_Start(b *testing.B) {
	for i := 0; i < b.N; i++ {
		job := NewDownloadJob("https://example.com/file{}.jpg", "test-folder", "test-repo")
		job.Start()
	}
}
