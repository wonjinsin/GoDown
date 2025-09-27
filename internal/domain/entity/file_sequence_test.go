package entity

import (
	"testing"
)

func TestNewFileSequence(t *testing.T) {
	sequence := NewFileSequence("https://example.com/file{}.jpg", "jpg", nil)

	if sequence.BaseURL != "https://example.com/file{}.jpg" {
		t.Errorf("expected BaseURL to be set")
	}
	if sequence.Extension != "jpg" {
		t.Errorf("expected Extension to be jpg")
	}
}

func TestFileSequence_GenerateURL(t *testing.T) {
	sequence := NewFileSequence("https://example.com/file{}.jpg", "jpg", nil)

	url, err := sequence.GenerateURL(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "https://example.com/file1.jpg"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestFileSequence_GenerateURL_WithSeparator(t *testing.T) {
	separator := "-"
	sequence := NewFileSequence("https://example.com/file{}.jpg", "jpg", &separator)

	url, err := sequence.GenerateURL(10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "https://example.com/file-10.jpg"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestFileSequence_EstimateFileCount(t *testing.T) {
	sequence := NewFileSequence("https://example.com/file{}.jpg", "jpg", nil)
	count := sequence.EstimateFileCount()

	if count <= 0 {
		t.Error("expected positive file count")
	}
}

// Benchmark tests
func BenchmarkFileSequence_GenerateURL(b *testing.B) {
	sequence := NewFileSequence("https://example.com/file{}.jpg", "jpg", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sequence.GenerateURL(uint64(i))
	}
}
