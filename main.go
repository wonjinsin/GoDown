package main

import (
	"cheetah/model"
	"cheetah/service"
	"fmt"
)

func main() {
	input := &model.Input{}
	fmt.Print("Write URL: ")
	fmt.Scan(&input.URL)

	fmt.Print("Write Folder: ")
	fmt.Scanln(&input.Folder)

	file := model.MakeFile(input)
	svc := service.NewFileService(file)
	svc.Do()
}
