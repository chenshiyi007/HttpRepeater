package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"httpRepeater/internal/repeater"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	repeaterService := repeater.NewService()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "HttpRepeater",
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 16, G: 16, B: 20, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			repeaterService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
