package template

import (
	"cheetah/controller"
	"cheetah/model"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// ShowMain ...
func ShowMain() {
	a := app.NewWithID("goDown")
	w := a.NewWindow("goDown")
	w.Resize(fyne.NewSize(600, 400))

	url, folder, separator, host, origin := widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "URL", Widget: url},
			{Text: "Folder", Widget: folder},
			{Text: "Separator(optional)", Widget: separator},
			{Text: "Host(optional)", Widget: host},
			{Text: "Origin(optional)", Widget: origin},
		},
		OnSubmit: func() {
			input := &model.Input{
				URL:    url.Text,
				Folder: folder.Text,
			}
			if separator.Text != "" {
				input.Separator = &separator.Text
			}
			if host.Text != "" {
				input.Host = &host.Text
			}
			if origin.Text != "" {
				input.Origin = &origin.Text
			}
			log.Printf("input is: %+v\n", input)
			controller.DoFileDownload(input)
		},
	}
	w.SetContent(form)
	w.ShowAndRun()
}
