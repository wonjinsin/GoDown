package controller

import (
	"cheetah/model"
	"cheetah/service"
)

// DoFileDownload ...
func DoFileDownload(input *model.Input, c chan int) (err error) {
	svc := service.NewFileService(input)
	return svc.Do(c)
}
