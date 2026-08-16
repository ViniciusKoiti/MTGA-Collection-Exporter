//go:build windows

// Command desktop is the Wails shell of the MTGA companion (OpenSpec
// introduce-agentic-go-companion, task 4.1): a TypeScript frontend
// with Home, Collection, Decks, Assistant and Settings navigation
// over this thin Go host. The shell holds no domain logic — data
// arrives through the application services in later tasks, and graph
// activities may only ever run via internal/activity.Launcher.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "MTGA Companion",
		Width:            1200,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 24, G: 26, B: 32, A: 1},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
	})
	if err != nil {
		println("desktop:", err.Error())
	}
}
