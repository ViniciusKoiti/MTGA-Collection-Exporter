//go:build windows

package main

import "context"

// App is the Wails-bound host of the desktop shell.
type App struct {
	ctx context.Context
}

// NewApp builds the shell host.
func NewApp() *App { return &App{} }

// startup keeps the runtime context for later runtime calls.
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// View is one top-level navigation destination of the shell.
type View struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Views lists the five navigation destinations in their fixed order.
// The frontend renders exactly this list — the navigation contract
// lives in ONE place and is unit-tested here.
func (a *App) Views() []View {
	return []View{
		{ID: "home", Label: "Home"},
		{ID: "collection", Label: "Collection"},
		{ID: "decks", Label: "Decks"},
		{ID: "assistant", Label: "Assistant"},
		{ID: "settings", Label: "Settings"},
	}
}
