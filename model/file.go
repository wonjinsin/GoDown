package model

import (
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
	Repo  string
	Input *Input
}

// SetFileSavePath ...
func (f *File) SetFileSavePath() string {
	return fmt.Sprintf("%s/%s", f.Repo, f.Input.Folder)
}

// GetReplacedURL ...
func (f File) GetReplacedURL(num uint64) string {
	arr := strings.Split(f.Input.URL, "/")
	fileName := arr[len(arr)-1]
	replacedFileName := f.getReplacedFileName(fileName, num)
	return strings.Join(arr[:len(arr)-1], "/") + replacedFileName
}

func (f File) getReplacedFileName(fileName string, num uint64) string {
	regex := regexp.MustCompile(`[0-9]+`)
	minIndexLen := len(regex.FindString(f.Input.Separator))
	return regex.ReplaceAllString(fileName, fmt.Sprintf("%0"+strconv.Itoa(minIndexLen)+"d", num))
}

// MakeDirectory ...
func (f *File) MakeDirectory() (err error) {
	dir := fmt.Sprintf("%s/%s", f.Repo, f.Input.Folder)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeFile ...
func (f *File) MakeFile(filename string, body io.ReadCloser) (err error) {
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
func (f *File) makeEmptyFile(filename string) (file *os.File, err error) {
	fullPath := fmt.Sprintf("%s/%s/%s", f.Repo, f.Input.Folder, filename)
	file, err = os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// StartCmd ...
func (f *File) StartCmd() (err error) {
	if _, err := exec.Command("/bin/sh", "ffmpeg.sh", fmt.Sprintf("%s/%s", f.Repo, f.Input.Folder), f.Input.Folder).Output(); err != nil {
		return err
	}
	return nil
}
