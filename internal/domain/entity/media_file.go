package entity

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// MediaFile represents a downloaded media file
type MediaFile struct {
	ID           string
	Filename     string
	OriginalURL  string
	LocalPath    string
	Size         int64
	Extension    string
	MimeType     string
	Status       FileStatus
	DownloadedAt time.Time
	Checksum     string
	Metadata     map[string]interface{}
}

// FileStatus represents the status of a file
type FileStatus int

const (
	FileStatusPending FileStatus = iota
	FileStatusDownloading
	FileStatusCompleted
	FileStatusFailed
	FileStatusCorrupted
)

func (s FileStatus) String() string {
	switch s {
	case FileStatusPending:
		return "pending"
	case FileStatusDownloading:
		return "downloading"
	case FileStatusCompleted:
		return "completed"
	case FileStatusFailed:
		return "failed"
	case FileStatusCorrupted:
		return "corrupted"
	default:
		return "unknown"
	}
}

// NewMediaFile creates a new media file entity
func NewMediaFile(filename, originalURL, localPath string) *MediaFile {
	extension := strings.TrimPrefix(filepath.Ext(filename), ".")

	return &MediaFile{
		ID:          generateFileID(filename),
		Filename:    filename,
		OriginalURL: originalURL,
		LocalPath:   localPath,
		Extension:   extension,
		Status:      FileStatusPending,
		Metadata:    make(map[string]interface{}),
	}
}

// SetSize sets the file size
func (mf *MediaFile) SetSize(size int64) {
	mf.Size = size
}

// SetMimeType sets the MIME type
func (mf *MediaFile) SetMimeType(mimeType string) {
	mf.MimeType = mimeType
}

// SetChecksum sets the file checksum
func (mf *MediaFile) SetChecksum(checksum string) {
	mf.Checksum = checksum
}

// AddMetadata adds metadata to the file
func (mf *MediaFile) AddMetadata(key string, value interface{}) {
	mf.Metadata[key] = value
}

// GetMetadata retrieves metadata by key
func (mf *MediaFile) GetMetadata(key string) (interface{}, bool) {
	value, exists := mf.Metadata[key]
	return value, exists
}

// StartDownload marks the file as being downloaded
func (mf *MediaFile) StartDownload() {
	mf.Status = FileStatusDownloading
}

// CompleteDownload marks the file as successfully downloaded
func (mf *MediaFile) CompleteDownload(size int64) {
	mf.Status = FileStatusCompleted
	mf.Size = size
	mf.DownloadedAt = time.Now()
}

// FailDownload marks the file download as failed
func (mf *MediaFile) FailDownload() {
	mf.Status = FileStatusFailed
}

// MarkCorrupted marks the file as corrupted
func (mf *MediaFile) MarkCorrupted() {
	mf.Status = FileStatusCorrupted
}

// IsDownloaded returns true if the file has been successfully downloaded
func (mf *MediaFile) IsDownloaded() bool {
	return mf.Status == FileStatusCompleted
}

// IsVideo returns true if the file is a video file
func (mf *MediaFile) IsVideo() bool {
	videoExtensions := []string{"mp4", "avi", "mkv", "mov", "wmv", "flv", "webm", "m4v"}
	return mf.isExtensionIn(videoExtensions)
}

// IsAudio returns true if the file is an audio file
func (mf *MediaFile) IsAudio() bool {
	audioExtensions := []string{"mp3", "wav", "flac", "aac", "ogg", "wma", "m4a"}
	return mf.isExtensionIn(audioExtensions)
}

// IsImage returns true if the file is an image file
func (mf *MediaFile) IsImage() bool {
	imageExtensions := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "tiff"}
	return mf.isExtensionIn(imageExtensions)
}

// GetFullPath returns the complete file path
func (mf *MediaFile) GetFullPath() string {
	return filepath.Join(mf.LocalPath, mf.Filename)
}

// GetSizeFormatted returns human-readable file size
func (mf *MediaFile) GetSizeFormatted() string {
	return formatFileSize(mf.Size)
}

// Validate validates the media file
func (mf *MediaFile) Validate() error {
	if mf.Filename == "" {
		return fmt.Errorf("filename is required")
	}

	if mf.OriginalURL == "" {
		return fmt.Errorf("original URL is required")
	}

	if mf.LocalPath == "" {
		return fmt.Errorf("local path is required")
	}

	return nil
}

// Helper functions
func (mf *MediaFile) isExtensionIn(extensions []string) bool {
	ext := strings.ToLower(mf.Extension)
	for _, validExt := range extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

func generateFileID(filename string) string {
	return fmt.Sprintf("file_%s_%d", strings.ReplaceAll(filename, ".", "_"), time.Now().UnixNano())
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
