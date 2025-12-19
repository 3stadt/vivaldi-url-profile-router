package gui

import (
	"maps"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ChooseProfile(defaultProfile *string, profiles map[string]string) (choosenProfile string, cancel bool) {

	cancel = false
	myApp := app.New()
	myWindow := myApp.NewWindow("Form Layout")
	myWindow.Resize(fyne.NewSize(800, 400))

	label := widget.NewLabel("Profile to use:")
	combo := widget.NewSelect(slices.Collect(maps.Keys(profiles)), func(value string) {
		choosenProfile = profiles[value]
	})
	cancelButton := widget.NewButton("Cancel", func() {
		cancel = true
		myApp.Quit()
	})
	okButton := widget.NewButton("Open", func() {
		myApp.Quit()
	})
	grid := container.New(layout.NewFormLayout(), label, combo, cancelButton, okButton)

	myWindow.SetContent(grid)
	myWindow.ShowAndRun()

	return choosenProfile, cancel
}
