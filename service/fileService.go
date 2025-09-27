package service

import (
	"cheetah/model"
	"cheetah/util"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// FileService ...
type FileService struct {
	Input *model.Input
	File  *model.File
}

// NewFileService ...
func NewFileService(input *model.Input) FileUsecase {
	fs := &FileService{
		Input: input,
		File:  model.MakeFile(input),
	}
	return fs
}

// Do ...
func (t *FileService) Do(c chan int) (err error) {
	if err = t.File.MakeDirectory(); err != nil {
		fmt.Printf("Error occurred: %s", err.Error())
		return err
	}

	startNum := 0
	batchCount := 50
	errCount := 0
	errMax := 100
	routineErrMax := 10

	for {
		if errCount > errMax {
			fmt.Println("File download done")
			break
		}

		var wg sync.WaitGroup
		wg.Add(batchCount)

		for j := 0; j < batchCount; j++ {
			go func(num uint64, wg *sync.WaitGroup) {
				defer wg.Done()
				url, err := t.File.GetReplacedURL(uint64(num))
				if err != nil {
					fmt.Printf("Error occured: %s", err.Error())
					errCount++
					return
				}
				for {
					errCnt := 0
					if err := t.DownloadFile(url, fmt.Sprintf("%d.%s", num, t.File.GetExtension())); err != nil {
						if err.Error() == util.ErrorIsMakeFile && errCnt < routineErrMax {
							fmt.Printf("error occured, try again %s", url)
							continue
						}
						fmt.Printf("Error occured: %s", err.Error())
						errCount++
					}
					return
				}
			}(uint64(startNum+j), &wg)
		}

		startNum += batchCount
		c <- int(startNum)
		wg.Wait()
	}

	if err = t.File.StartCmd(); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// DownloadFile ...
func (t *FileService) DownloadFile(url string, filename string) error {
	fmt.Println(fmt.Sprintf("filename: %s, url: %s", filename, url))
	client, err := model.MakeClient(url, t.Input.Host, t.Input.Origin)
	if err != nil {
		return err
	}

	resp, err := client.Do()
	defer func(resp *http.Response) {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}(resp)

	if err != nil {
		fmt.Println(fmt.Printf("Error occurred: %s", err.Error()))
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Println("Request failed: Status Code is not valid", "url", url, "resp", resp)
		return fmt.Errorf("Request failed: Status Code is not valid")
	}

	// if resp.ContentLength <= 0 {
	// 	if err := t.File.MakeFile(filename, resp.Body); err != nil {
	// 		fmt.Println("Request failed: Content length is not valid", "url", url, "length", resp.ContentLength)
	// 		return fmt.Errorf("Request failed: Content length is not valid")
	// 	}
	// } else {
	// 	if err := t.File.MakeFile(filename, resp.Body); err != nil {
	// 		fmt.Printf("Error occurred: %s", err.Error())
	// 		return errors.New("MakeFile failed")
	// 	}
	// }

	if err := t.File.MakeFile(filename, resp.Body); err != nil {
		fmt.Printf("Error occurred: %s", err.Error())
		return errors.New("MakeFile failed")
	}

	fmt.Println(fmt.Sprintf("Downloaded a file, filename: %s", filename))
	return nil
}
