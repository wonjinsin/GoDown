package template

import (
	"cheetah/controller"
	"cheetah/model"
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
			downloading := ShowDownloading(a)
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
			c := make(chan int)
			go SetDownloadingContents(downloading, c)
			log.Printf("test")
			if err := controller.DoFileDownload(input, c); err != nil {
				downloading.Close()
				ShowSuccess(a)
				return
			}
			fmt.Println("come to here")
		},
	}
	w.SetContent(form)
	w.ShowAndRun()
}

// ShowDownloading ...
func ShowDownloading(a fyne.App) (w fyne.Window) {
	w = a.NewWindow("Downloading")
	w.Resize(fyne.NewSize(300, 100))
	text := canvas.NewText("Downloading now...", color.Black)
	text.Alignment = fyne.TextAlignTrailing
	text.TextStyle = fyne.TextStyle{Monospace: true}
	content := container.New(layout.NewCenterLayout(), text)
	w.SetContent(content)
	w.Show()

	return w
}

// SetDownloadingContents ...
func SetDownloadingContents(w fyne.Window, c chan int) {
	for i := range c {
		text := canvas.NewText(fmt.Sprintf("%d", i), color.Black)
		text.Alignment = fyne.TextAlignTrailing
		text.TextStyle = fyne.TextStyle{Monospace: true}
		content := container.New(layout.NewCenterLayout(), text)
		w.SetContent(content)
	}
}

// ShowSuccess ...
func ShowSuccess(a fyne.App) (w fyne.Window) {
	w = a.NewWindow("Success")
	text := canvas.NewText("Success", color.Black)
	text.Alignment = fyne.TextAlignTrailing
	text.TextStyle = fyne.TextStyle{Italic: true}
	w.SetContent(text)
	w.Show()
	return w
}
