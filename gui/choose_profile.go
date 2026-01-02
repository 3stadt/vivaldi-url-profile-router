package gui

import (
	"maps"
	"os"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

func ChooseProfile(launchURL string, defaultProfile *string, maxButtons int, profiles map[string]string) (choosenProfile string) {

	profileChooser := app.New()
	choiceWindow := profileChooser.NewWindow("Profile Selection")
	choiceWindow.Resize(fyne.NewSize(800, 400))

	// If window is closed, do not open the browser
	choiceWindow.SetCloseIntercept(func() {
		choiceWindow.Close()
		profileChooser.Quit()
		os.Exit(0)
	})

	titleLabel := widget.NewLabel("Please select a profile to launch.")
	urlLabel := ttwidget.NewLabelWithStyle(truncateText(launchURL, 80), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	if len(launchURL) > 80 {
		urlLabel.SetToolTip(launchURL)
	}
	header := container.NewVBox(titleLabel, urlLabel)

	keys := slices.Collect(maps.Keys(profiles))
	slices.Sort(keys)
	defaultKey := ""
	if defaultProfile != nil && *defaultProfile != "" {
		for _, name := range keys {
			if profiles[name] == *defaultProfile {
				defaultKey = name
				break
			}
		}
	}
	if defaultKey == "" {
		if _, ok := profiles["Default"]; ok {
			defaultKey = "Default"
		}
	}

	buttons := make([]fyne.CanvasObject, 0, len(keys)+2)
	buttons = append(buttons, header)

	if defaultKey != "" {
		profileFolder := profiles[defaultKey]
		profileName := defaultKey
		defaultButton := widget.NewButton(profileName, func() {
			choosenProfile = profileFolder
			profileChooser.Quit()
		})
		buttons = append(buttons, defaultButton)
	}

	for _, name := range keys {
		if name == defaultKey {
			continue
		}
		profileFolder := profiles[name]
		profileName := name
		button := widget.NewButton(profileName, func() {
			choosenProfile = profileFolder
			profileChooser.Quit()
		})
		buttons = append(buttons, button)
	}

	cancelButton := widget.NewButton("Cancel", func() {
		choiceWindow.Close()
		profileChooser.Quit()
		os.Exit(0)
	})
	buttons = append(buttons, cancelButton)

	var content fyne.CanvasObject
	buttonCount := len(keys) + 1
	if maxButtons > 0 && buttonCount > maxButtons {
		content = container.NewVScroll(container.NewVBox(buttons...))
	} else {
		content = container.NewVBox(buttons...)
	}
	choiceWindow.SetContent(fynetooltip.AddWindowToolTipLayer(content, choiceWindow.Canvas()))
	choiceWindow.ShowAndRun()

	return choosenProfile
}

func truncateText(s string, max int) string {
	if len(s) > max {
		r := 0
		for i := range s {
			r++
			if r > max {
				return s[:i] + "..."
			}
		}
	}
	return s
}
