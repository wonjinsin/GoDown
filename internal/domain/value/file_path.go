package value

import (
	"cheetah/util"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// FilePath represents a validated file path value object
type FilePath struct {
	raw       string
	directory string
	filename  string
	extension string
}

// NewFilePath creates a new FilePath value object
func NewFilePath(path string) (*FilePath, error) {
	if path == "" {
		return nil, util.NewValidationError("path", path, "File path cannot be empty")
	}

	// Validate path format
	if err := validatePath(path); err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")

	return &FilePath{
		raw:       path,
		directory: dir,
		filename:  filename,
		extension: ext,
	}, nil
}

// String returns the raw path
func (fp *FilePath) String() string {
	return fp.raw
}

// Directory returns the directory part
func (fp *FilePath) Directory() string {
	return fp.directory
}

// Filename returns the filename with extension
func (fp *FilePath) Filename() string {
	return fp.filename
}

// Extension returns the file extension without the dot
func (fp *FilePath) Extension() string {
	return fp.extension
}

// BaseName returns the filename without extension
func (fp *FilePath) BaseName() string {
	return strings.TrimSuffix(fp.filename, "."+fp.extension)
}

// Join creates a new FilePath by joining with another path component
func (fp *FilePath) Join(component string) *FilePath {
	newPath := filepath.Join(fp.raw, component)
	// We assume this is valid since we're joining with an existing valid path
	newFp, _ := NewFilePath(newPath)
	return newFp
}

// WithExtension returns a new FilePath with a different extension
func (fp *FilePath) WithExtension(ext string) *FilePath {
	baseName := fp.BaseName()
	newFilename := baseName + "." + ext
	newPath := filepath.Join(fp.directory, newFilename)

	newFp, _ := NewFilePath(newPath)
	return newFp
}

// WithFilename returns a new FilePath with a different filename
func (fp *FilePath) WithFilename(filename string) *FilePath {
	newPath := filepath.Join(fp.directory, filename)
	newFp, _ := NewFilePath(newPath)
	return newFp
}

// IsAbsolute returns true if the path is absolute
func (fp *FilePath) IsAbsolute() bool {
	return filepath.IsAbs(fp.raw)
}

// Clean returns a cleaned version of the path
func (fp *FilePath) Clean() *FilePath {
	cleaned := filepath.Clean(fp.raw)
	newFp, _ := NewFilePath(cleaned)
	return newFp
}

// Validate checks if the file path is valid
func (fp *FilePath) Validate() error {
	return validatePath(fp.raw)
}

// Helper functions
func validatePath(path string) error {
	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return util.NewValidationError("path", path, fmt.Sprintf("Path contains invalid character: %s", char))
		}
	}

	// Check for reserved names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	filename := filepath.Base(path)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	for _, reserved := range reservedNames {
		if strings.EqualFold(baseName, reserved) {
			return util.NewValidationError("path", path, fmt.Sprintf("Path contains reserved name: %s", reserved))
		}
	}

	// Check path length (reasonable limit)
	if len(path) > 260 {
		return util.NewValidationError("path", path, "Path is too long (max 260 characters)")
	}

	// Check for valid filename pattern
	if filename != "." && filename != ".." {
		validPattern := regexp.MustCompile(`^[^<>:"/\\|?*\x00-\x1f]+$`)
		if !validPattern.MatchString(filename) {
			return util.NewValidationError("path", path, "Invalid filename format")
		}
	}

	return nil
}
