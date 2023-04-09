package template

import (
	"cheetah/controller"
	"cheetah/model"
	"cheetah/util"
	"fmt"
	"image/color"
	"log"
	"strings"

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
			downloading := showDownloading(a)
			input := &model.Input{
				URL:    strings.TrimSpace(url.Text),
				Folder: strings.TrimSpace(folder.Text),
			}
			if separator.Text != "" {
				input.Separator = util.ToPointer(strings.TrimSpace(separator.Text))
			}
			if host.Text != "" {
				input.Host = util.ToPointer(strings.TrimSpace(host.Text))
			}
			if origin.Text != "" {
				input.Origin = util.ToPointer(strings.TrimSpace(origin.Text))
			}
			log.Printf("input is: %+v\n", input)
			c := make(chan int)
			go setDownloadingContents(downloading, c)
			if err := controller.DoFileDownload(input, c); err != nil {
				showResult(a, fmt.Sprintf("Faild: %s", err.Error()))
			} else {
				showResult(a, "Success")
			}
			downloading.Close()
		},
	}
	w.SetContent(form)
	w.ShowAndRun()
}

func showDownloading(a fyne.App) (w fyne.Window) {
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

func setDownloadingContents(w fyne.Window, c chan int) {
	for i := range c {
		text := canvas.NewText(fmt.Sprintf("Downloading now ... %d", i), color.Black)
		text.Alignment = fyne.TextAlignTrailing
		text.TextStyle = fyne.TextStyle{Monospace: true}
		content := container.New(layout.NewCenterLayout(), text)
		w.SetContent(content)
	}
}

func showResult(a fyne.App, t string) (w fyne.Window) {
	w = a.NewWindow("Result")
	w.Resize(fyne.NewSize(300, 100))
	text := canvas.NewText(t, color.Black)
	text.Alignment = fyne.TextAlignTrailing
	text.TextStyle = fyne.TextStyle{Italic: true}
	content := container.New(layout.NewCenterLayout(), text)
	w.SetContent(content)
	w.Show()
	return w
}
