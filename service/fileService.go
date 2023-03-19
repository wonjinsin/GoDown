package service

import (
	"cheetah/model"
	"fmt"
	"net/http"
	"sync"
)

// FileService ...
type FileService struct {
	File   *model.File
	Client *model.Client
}

// NewFileService ...
func NewFileService(file *model.File) FileUsecase {
	fs := &FileService{
		File: file,
	}
	return fs
}

// Do ...
func (t *FileService) Do() {
	err := t.File.MakeDirectory()
	if err != nil {
		fmt.Printf("Error occurred: %s", err.Error())
	}

	startNum := 1
	batchCount := 20
	errCount := 0

	for {
		if errCount > 100 {
			fmt.Println("File download done")
			break
		}

		var wg sync.WaitGroup
		wg.Add(batchCount)

		for j := 0; j < batchCount; j++ {
			go func(num uint64, wg *sync.WaitGroup) {
				url, err := t.File.GetReplacedURL(uint64(num))
				if err != nil {
					fmt.Printf("Error occured: %s", err.Error())
					errCount++
				}
				if err := t.DownloadFile(url, fmt.Sprintf("%d.%s", num, t.File.GetExtension())); err != nil {
					fmt.Printf("Error occured: %s", err.Error())
					errCount++
				}
				wg.Done()
			}(uint64(startNum+j), &wg)
		}

		startNum += batchCount
		wg.Wait()
	}

	if err := t.File.StartCmd(); err != nil {
		fmt.Println(err.Error())
	}
}

// DownloadFile ...
func (t *FileService) DownloadFile(url string, filename string) error {
	fmt.Println(fmt.Sprintf("start: %s, url: %s", filename, url))
	client := &model.Client{Client: &http.Client{}}
	client.SetRequest(url)

	resp, err := client.Do()
	if err != nil {
		fmt.Println(fmt.Printf("Error occurred: %s", err.Error()))
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Println("Request failed: Status Code is not valid", "resp", resp)
		return fmt.Errorf("Request failed: Status Code is not valid")
	}

	if resp.ContentLength <= 0 {
		fmt.Println("Request failed: Content length is not valid", "resp", resp)
		return fmt.Errorf("Request failed: Content length is not valid")
	}

	if err := t.File.MakeFile(filename, resp.Body); err != nil {
		fmt.Printf("Error occurred: %s", err.Error())
		return err
	}

	fmt.Println(fmt.Sprintf("Downloaded a file, filename: %s", filename))
	return nil
}
