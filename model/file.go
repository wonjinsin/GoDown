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
	Repo      string
	URL       string
	Separator string
	Folder    string
}

// SetFileSavePath ...
func (f *File) SetFileSavePath() string {
	return fmt.Sprintf("%s/%s", f.Repo, f.Folder)
}

// GetReplacedURL ...
func (f File) GetReplacedURL(num uint64) (url string) {

	var path, param string
	arr := strings.Split(f.URL, "?")

	path = arr[0]
	if len(arr) > 1 {
		param = arr[1]
	}

	url = f.getReplacedPath(path, num)
	if param != "" {
		url += fmt.Sprintf("?%s", param)
	}
	return
}

func (f File) getReplacedPath(path string, num uint64) string {
	r := regexp.MustCompile("\\w\\/([a-zA-Z0-9-_]+).([a-z0-9]+)")
	arr := r.FindStringSubmatch(path)

	r.FindStringSubmatch("https://www.test.kr/test-001.ts?test")
	arr := strings.Split(path, "/")
	fileName := arr[len(arr)-1]
	replacedFileName := f.getReplacedFileName(fileName, num)
	return strings.Join(arr[:len(arr)-1], "/") + replacedFileName
}

func (f File) getReplacedFileName(fileName string, num uint64) string {
	regex := regexp.MustCompile(`[0-9]+`)
	minIndexLen := len(regex.FindString(f.Separator))
	return regex.ReplaceAllString(fileName, fmt.Sprintf("%0"+strconv.Itoa(minIndexLen)+"d", num))
}

// MakeDirectory ...
func (f *File) MakeDirectory() (err error) {
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
	fullPath := fmt.Sprintf("%s/%s/%s", f.Repo, f.Folder, filename)
	file, err = os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// StartCmd ...
func (f *File) StartCmd() (err error) {
	if _, err := exec.Command("/bin/sh", "ffmpeg.sh", fmt.Sprintf("%s/%s", f.Repo, f.Folder), f.Folder).Output(); err != nil {
		return err
	}
	return nil
}
