package model

import (
	"cheetah/config"
	"cheetah/pkg/logger"
	"cheetah/util"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// File ...
type File struct {
	Repo      string
	URL       string
	Separator *string
	Extension string
	Folder    string
	logger    zerolog.Logger
}

// MakeFile ...
func MakeFile(input *Input, cfg *config.Config) *File {
	f := &File{
		Repo:   cfg.App.RepoDir,
		URL:    input.URL,
		Folder: input.Folder,
		logger: logger.GetLogger("file_model"),
	}
	f.SetSeparator(input.Separator)
	f.SetExtension()
	return f
}

// SetFileSavePath ...
func (f *File) SetFileSavePath() string {
	return fmt.Sprintf("%s/%s", f.Repo, f.Folder)
}

// GetReplacedURL ...
func (f File) GetReplacedURL(num uint64) (url string, err error) {
	var path, param string
	arr := strings.Split(f.URL, "?")

	path = arr[0]
	if len(arr) > 1 {
		param = arr[1]
	}

	url, err = f.getReplacedPath(path, num)
	if err != nil {
		return "", err
	}
	if param != "" {
		url += fmt.Sprintf("?%s", param)
	}
	return
}

func (f File) getReplacedPath(path string, num uint64) (replaced string, err error) {
	r := regexp.MustCompile("\\w\\/([a-zA-Z0-9-_]+)\\.[a-z0-9]+")
	arr := r.FindStringSubmatch(path)
	if len(arr) < 2 {
		f.logger.Error().
			Str("path", path).
			Uint64("file_number", num).
			Msg("Invalid path format - cannot extract filename")
		return "", fmt.Errorf("%s: %s", util.ErrInvalidPath, path)
	}
	fileName := arr[1]
	replacedFileName := f.getReplacedFileName(fileName, num)
	if replacedFileName == "" {
		f.logger.Error().
			Str("filename", fileName).
			Uint64("file_number", num).
			Msg("Failed to generate replacement filename")
		return "", fmt.Errorf("%s: %s", util.ErrInvalidFileName, fileName)
	}
	return strings.Replace(path, fmt.Sprintf("/%s.", fileName), fmt.Sprintf("/%s.", replacedFileName), 1), nil
}

func (f File) getReplacedFileName(fileName string, num uint64) string {
	r := regexp.MustCompile(`^([0-9]+)$|(-[0-9]+)|(_[0-9]+)|([a-zA-Z]+[0-9]+)`)
	arr := r.FindStringSubmatch(fileName)
	var separator string
	var separatorLen int

	if f.Separator != nil {
		separator = *f.Separator
		separatorLen = len(separator)
	}

	if separator == "" {
		for i, v := range arr {
			if i == 0 || v == "" {
				continue
			}
			separator = v
			separatorLen = len(v)
			if i > 1 {
				separatorLen--
			}
			break
		}
	}

	replaced := regexp.MustCompile(`[0-9]+`).ReplaceAllString(separator, fmt.Sprintf("%0"+strconv.Itoa(separatorLen)+"d", num))
	// replaced := regexp.MustCompile(`[0-9]+`).ReplaceAllString(separator, fmt.Sprintf("%d", num))
	return strings.Replace(fileName, separator, replaced, 1)
}

// MakeDirectory ...
func (f File) MakeDirectory() (err error) {
	dir := fmt.Sprintf("%s/%s", f.Repo, f.Folder)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeFile ...
func (f File) MakeFile(filename string, body io.ReadCloser) (err error) {
	file, err := f.makeEmptyFile(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", util.ErrMakeFile, err)
	}
	defer file.Close()

	written, err := io.Copy(file, body)
	if err != nil {
		f.logger.Error().
			Err(err).
			Str("filename", filename).
			Msg("Failed to write file content")
		return fmt.Errorf("%s: %w", util.ErrMakeFile, err)
	}
	if written == 0 {
		f.logger.Warn().
			Str("filename", filename).
			Msg("Downloaded file is empty")
		return fmt.Errorf("%s", util.ErrEmptyFile)
	}

	f.logger.Debug().
		Str("filename", filename).
		Int64("bytes_written", written).
		Msg("File saved successfully")
	return nil
}

// makeEmptyFile ...
func (f File) makeEmptyFile(filename string) (file *os.File, err error) {
	fullPath := fmt.Sprintf("%s/%s/%s", f.Repo, f.Folder, filename)
	file, err = os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// SetSeparator ...
func (f *File) SetSeparator(s *string) {
	if s == nil {
		return
	}
	f.Separator = s
}

// SetExtension ...
func (f *File) SetExtension() string {
	r := regexp.MustCompile("\\.(\\w+)$|\\.(\\w+)\\?")
	arr := r.FindStringSubmatch(f.URL)
	if len(arr) < 2 {
		return ""
	}
	for i, v := range arr {
		if i == 0 || v == "" {
			continue
		}
		f.Extension = v
		break
	}
	return f.Extension
}

// GetExtension ...
func (f File) GetExtension() string {
	return f.Extension
}

// StartCmd ...
func (f File) StartCmd() (err error) {
	scriptPath := "ffmpeg.sh"
	folderPath := fmt.Sprintf("%s/%s", f.Repo, f.Folder)

	f.logger.Info().
		Str("script", scriptPath).
		Str("folder_path", folderPath).
		Str("folder_name", f.Folder).
		Str("extension", f.Extension).
		Msg("Starting FFmpeg processing")

	output, err := exec.Command("/bin/sh", scriptPath, folderPath, f.Folder, f.Extension).Output()
	if err != nil {
		f.logger.Error().
			Err(err).
			Str("script", scriptPath).
			Str("folder_path", folderPath).
			Str("output", string(output)).
			Msg("FFmpeg command failed")
		return fmt.Errorf("%s: %w", util.ErrFFmpegExecution, err)
	}

	f.logger.Info().
		Str("output", string(output)).
		Msg("FFmpeg processing completed successfully")
	return nil
}
