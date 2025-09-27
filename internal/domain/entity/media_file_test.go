package entity

import (
	"testing"
)

func TestNewMediaFile(t *testing.T) {
	mediaFile := NewMediaFile("test.jpg", "https://example.com/test.jpg", "/path/to/test.jpg")

	if mediaFile.Filename != "test.jpg" {
		t.Errorf("expected filename to be set")
	}
	if mediaFile.OriginalURL != "https://example.com/test.jpg" {
		t.Errorf("expected OriginalURL to be set")
	}
	if mediaFile.LocalPath != "/path/to/test.jpg" {
		t.Errorf("expected LocalPath to be set")
	}
	if mediaFile.Status != FileStatusPending {
		t.Errorf("expected status to be pending")
	}
}

func TestFileStatus_String(t *testing.T) {
	tests := []struct {
		status   FileStatus
		expected string
	}{
		{FileStatusPending, "pending"},
		{FileStatusDownloading, "downloading"},
		{FileStatusCompleted, "completed"},
		{FileStatusFailed, "failed"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status.String())
		}
	}
}

// Benchmark tests
func BenchmarkNewMediaFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewMediaFile("test.jpg", "https://example.com/test.jpg", "/path/to/test.jpg")
	}
}
