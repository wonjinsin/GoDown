package testutil

import (
	"cheetah/config"
	"cheetah/internal/domain/entity"
	"cheetah/internal/domain/service"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestConfig creates a test configuration
func TestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			RepoDir: "test_repo",
		},
		Download: config.DownloadConfig{
			BatchSize:      5,
			MaxErrors:      10,
			RoutineErrMax:  3,
			MaxRetries:     3,
			HTTPTimeout:    30 * time.Second,
			MaxConcurrency: 10,
			UserAgent:      "GoDown-Test/1.0",
		},
		Logger: config.LoggerConfig{
			Level:  "debug",
			Pretty: true,
		},
	}
}

// TestLogger creates a test logger that discards output
func TestLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).Level(zerolog.Disabled)
}

// CreateTestDownloadJob creates a test download job
func CreateTestDownloadJob() *entity.DownloadJob {
	job := entity.NewDownloadJob(
		"https://example.com/file{}.jpg",
		"test-folder",
		"test-repo",
	)
	return job
}

// CreateTestDownloadJobRequest creates a test download job request
func CreateTestDownloadJobRequest() service.CreateDownloadJobRequest {
	return service.CreateDownloadJobRequest{
		URL:       "https://example.com/file{}.jpg",
		Folder:    "test-folder",
		Host:      nil,
		Origin:    nil,
		Separator: nil,
	}
}

// CreateTestFileSequence creates a test file sequence
func CreateTestFileSequence() *entity.FileSequence {
	separator := ""
	sequence := entity.NewFileSequence(
		"https://example.com/file{}.jpg",
		"test-folder",
		&separator,
	)
	return sequence
}

// CreateTestMediaFile creates a test media file
func CreateTestMediaFile() *entity.MediaFile {
	return &entity.MediaFile{
		Filename: "test-file.jpg",
		Size:     1024,
		Status:   entity.FileStatusPending,
	}
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "godown-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// CreateTestFile creates a test file with content
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	filepath := filepath.Join(dir, filename)
	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return filepath
}

// AssertNoError asserts that there is no error
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError asserts that there is an error
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected == actual {
		t.Fatalf("Expected values to be different, but both were %v", expected)
	}
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Fatal("Expected condition to be true")
	}
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Fatal("Expected condition to be false")
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr string) {
	t.Helper()
	if !contains(str, substr) {
		t.Fatalf("Expected %q to contain %q", str, substr)
	}
}

// AssertNotContains asserts that a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string) {
	t.Helper()
	if contains(str, substr) {
		t.Fatalf("Expected %q to not contain %q", str, substr)
	}
}

// contains is a helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(substr) <= len(str) && findSubstring(str, substr)))
}

// findSubstring finds a substring in a string
func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ContextWithTimeout creates a context with timeout for testing
func ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// TestContext creates a test context with reasonable timeout
func TestContext() (context.Context, context.CancelFunc) {
	return ContextWithTimeout(10 * time.Second)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// ExpectPanic expects a function to panic
func ExpectPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected function to panic, but it didn't")
		}
	}()

	fn()
}

// TableTest represents a table-driven test case
type TableTest[T any] struct {
	Name     string
	Input    T
	Expected interface{}
	Error    bool
}

// RunTableTests runs table-driven tests
func RunTableTests[T any](t *testing.T, tests []TableTest[T], testFunc func(t *testing.T, input T) (interface{}, error)) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := testFunc(t, tt.Input)

			if tt.Error {
				AssertError(t, err)
			} else {
				AssertNoError(t, err)
				AssertEqual(t, tt.Expected, result)
			}
		})
	}
}

// BenchmarkHelper provides utilities for benchmarking
type BenchmarkHelper struct {
	b *testing.B
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{b: b}
}

// TimeOperation times an operation
func (bh *BenchmarkHelper) TimeOperation(name string, operation func()) {
	bh.b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			operation()
		}
	})
}

// MeasureAllocs measures allocations for an operation
func (bh *BenchmarkHelper) MeasureAllocs(name string, operation func()) {
	bh.b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			operation()
		}
	})
}

// TestProgress represents test progress information
type TestProgress struct {
	Total     int
	Completed int
	Failed    int
	Current   string
}

// String returns string representation of test progress
func (tp TestProgress) String() string {
	return fmt.Sprintf("Progress: %d/%d (failed: %d) - %s",
		tp.Completed, tp.Total, tp.Failed, tp.Current)
}
