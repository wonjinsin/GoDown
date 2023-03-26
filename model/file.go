package model

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// File ...
type File struct {
	Repo      string
	URL       string
	Separator *string
	Extension string
	Folder    string
}

// MakeFile ...
func MakeFile(input *Input) *File {
	f := &File{
		Repo:   "repo",
		URL:    input.URL,
		Folder: input.Folder,
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
		return "", errors.New("Invalid path")
	}
	fileName := arr[1]
	replacedFileName := f.getReplacedFileName(fileName, num)
	if replacedFileName == "" {
		return "", errors.New("Invalid replacedFileName")
	}
	return strings.Replace(path, fmt.Sprintf("/%s.", fileName), fmt.Sprintf("/%s.", replacedFileName), 1), nil
}

func (f File) getReplacedFileName(fileName string, num uint64) string {
	r := regexp.MustCompile(`^([0-9]+)$|(-[0-9]+)|(_[0-9]+)`)
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
	return r.ReplaceAllString(fileName, replaced)
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
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	defer body.Close()
	if err != nil {
		return err
	}
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
func (f *File) SetExtension() {
	r := regexp.MustCompile("\\.(\\w+)$|\\.(\\w+)\\?")
	arr := r.FindStringSubmatch(f.URL)
	if len(arr) < 2 {
		return
	}
	f.Extension = arr[1]
}

// GetExtension ...
func (f File) GetExtension() string {
	return f.Extension
}

// StartCmd ...
func (f *File) StartCmd() (err error) {
	if _, err := exec.Command("/bin/sh", "ffmpeg.sh", fmt.Sprintf("%s/%s", f.Repo, f.Folder), f.Folder, f.Extension).Output(); err != nil {
		return err
	}
	return nil
}
