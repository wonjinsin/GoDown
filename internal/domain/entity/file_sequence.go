package entity

import (
	"cheetah/util"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FileSequence handles sequential file URL generation and naming
type FileSequence struct {
	BaseURL   string
	Extension string
	Separator *string
	Pattern   *FilePattern
}

// FilePattern represents the detected pattern in filename
type FilePattern struct {
	Prefix       string
	Separator    string
	NumberLength int
	StartNumber  uint64
}

// NewFileSequence creates a new file sequence
func NewFileSequence(baseURL, extension string, separator *string) *FileSequence {
	fs := &FileSequence{
		BaseURL:   baseURL,
		Extension: extension,
		Separator: separator,
	}

	// Analyze the URL pattern
	fs.analyzePattern()

	return fs
}

// GenerateURL generates a URL for the given file number
func (fs *FileSequence) GenerateURL(fileNumber uint64) (string, error) {
	var path, params string
	parts := strings.Split(fs.BaseURL, "?")

	path = parts[0]
	if len(parts) > 1 {
		params = parts[1]
	}

	newPath, err := fs.replacePathNumber(path, fileNumber)
	if err != nil {
		return "", err
	}

	if params != "" {
		return fmt.Sprintf("%s?%s", newPath, params), nil
	}

	return newPath, nil
}

// GenerateFilename generates a filename for the given file number
func (fs *FileSequence) GenerateFilename(fileNumber uint64) string {
	return fmt.Sprintf("%d.%s", fileNumber, fs.Extension)
}

// replacePathNumber replaces the number in the file path
func (fs *FileSequence) replacePathNumber(path string, fileNumber uint64) (string, error) {
	// Extract filename from path
	r := regexp.MustCompile(`\w/([a-zA-Z0-9-_]+)\.[a-z0-9]+`)
	matches := r.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("%s: cannot extract filename from path: %s", util.ErrInvalidPath, path)
	}

	originalFilename := matches[1]
	newFilename := fs.replaceFilenameNumber(originalFilename, fileNumber)
	if newFilename == "" {
		return "", fmt.Errorf("%s: failed to generate new filename for: %s", util.ErrInvalidFileName, originalFilename)
	}

	return strings.Replace(path, fmt.Sprintf("/%s.", originalFilename), fmt.Sprintf("/%s.", newFilename), 1), nil
}

// replaceFilenameNumber replaces the number in filename
func (fs *FileSequence) replaceFilenameNumber(filename string, fileNumber uint64) string {
	r := regexp.MustCompile(`^([0-9]+)$|(-[0-9]+)|(_[0-9]+)|([a-zA-Z]+[0-9]+)`)
	matches := r.FindStringSubmatch(filename)

	var separator string
	var separatorLen int

	// Use provided separator or detect from filename
	if fs.Separator != nil {
		separator = *fs.Separator
		separatorLen = len(separator)
	} else {
		// Auto-detect separator from matches
		for i, match := range matches {
			if i == 0 || match == "" {
				continue
			}
			separator = match
			separatorLen = len(match)
			if i > 1 {
				separatorLen-- // Account for separator character
			}
			break
		}
	}

	if separator == "" {
		return ""
	}

	// Generate new number with proper padding
	newNumber := fmt.Sprintf("%0"+strconv.Itoa(separatorLen)+"d", fileNumber)
	numberRegex := regexp.MustCompile(`[0-9]+`)
	newSeparator := numberRegex.ReplaceAllString(separator, newNumber)

	return strings.Replace(filename, separator, newSeparator, 1)
}

// analyzePattern analyzes the URL pattern for better understanding
func (fs *FileSequence) analyzePattern() {
	// Extract filename pattern from URL
	r := regexp.MustCompile(`\w/([a-zA-Z0-9-_]+)\.[a-z0-9]+`)
	matches := r.FindStringSubmatch(fs.BaseURL)
	if len(matches) < 2 {
		return
	}

	filename := matches[1]

	// Analyze number pattern
	numberRegex := regexp.MustCompile(`([a-zA-Z]*)([_-]?)([0-9]+)`)
	numberMatches := numberRegex.FindStringSubmatch(filename)
	if len(numberMatches) >= 4 {
		prefix := numberMatches[1]
		separator := numberMatches[2]
		numberStr := numberMatches[3]

		if startNum, err := strconv.ParseUint(numberStr, 10, 64); err == nil {
			fs.Pattern = &FilePattern{
				Prefix:       prefix,
				Separator:    separator,
				NumberLength: len(numberStr),
				StartNumber:  startNum,
			}
		}
	}
}

// GetPattern returns the detected file pattern
func (fs *FileSequence) GetPattern() *FilePattern {
	return fs.Pattern
}

// EstimateFileCount estimates the number of files based on common patterns
func (fs *FileSequence) EstimateFileCount() int {
	if fs.Pattern != nil && fs.Pattern.NumberLength > 0 {
		// Estimate based on number length (e.g., 001 suggests up to 999 files)
		maxNumber := 1
		for i := 0; i < fs.Pattern.NumberLength; i++ {
			maxNumber *= 10
		}
		return maxNumber - 1
	}

	// Default conservative estimate
	return 100
}
