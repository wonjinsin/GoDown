package main

import (
	"cheetah/model"
	"cheetah/service"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("goDown")
	w := a.NewWindow("goDown")
	w.Resize(fyne.NewSize(600, 400))

	url, folder, separator, origin := widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "URL", Widget: url},
			{Text: "Folder", Widget: folder},
			{Text: "Separator(optional)", Widget: separator},
			{Text: "Origin(optional)", Widget: origin},
		},
		OnSubmit: func() { // optional, handle form submission
			input := &model.Input{
				URL:    url.Text,
				Folder: folder.Text,
			}
			if separator.Text != "" {
				input.Separator = &separator.Text
			}
			if origin.Text != "" {
				input.Origin = &origin.Text
			}
			log.Printf("input is: %+v\n", input)
			svc := service.NewFileService(input)
			svc.Do()
		},
	}

	w.SetContent(form)
	w.ShowAndRun()

	// input := &model.Input{}
	// fmt.Print("Write URL: ")
	// fmt.Scan(&input.URL)

	// fmt.Print("Write Folder: ")
	// fmt.Scanln(&input.Folder)

	// svc := service.NewFileService(input)
	// svc.Do()
}
