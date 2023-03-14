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

	fmt.Print("Write Seperator: ")
	fmt.Scan(&input.Separator)

	fmt.Print("Write Folder: ")
	fmt.Scanln(&input.Folder)

	svc := service.NewFileService(input)
	svc.Do()
}
