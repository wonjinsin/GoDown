package controller

import (
	"cheetah/config"
	"cheetah/model"
	"cheetah/service"
)

// DoFileDownload ...
func DoFileDownload(input *model.Input, c chan int, cfg *config.Config) (err error) {
	svc := service.NewFileService(input, cfg)
	return svc.Do(c)
}
