package controller

import (
	"cheetah/model"
	"cheetah/service"
)

// DoFileDownload ...
func DoFileDownload(input *model.Input) {
	svc := service.NewFileService(input)
	svc.Do()
}
