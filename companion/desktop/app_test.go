//go:build windows

package main

import "testing"

// TestShellNavigatesExactlyTheFiveViews pins the navigation contract
// of task 4.1: Home, Collection, Decks, Assistant and Settings, in
// this order, with stable IDs the frontend routes by.
func TestShellNavigatesExactlyTheFiveViews(t *testing.T) {
	views := NewApp().Views()
	want := []View{
		{ID: "home", Label: "Home"},
		{ID: "collection", Label: "Collection"},
		{ID: "decks", Label: "Decks"},
		{ID: "assistant", Label: "Assistant"},
		{ID: "settings", Label: "Settings"},
	}
	if len(views) != len(want) {
		t.Fatalf("the shell must expose exactly five views: %+v", views)
	}
	for i, view := range views {
		if view != want[i] {
			t.Fatalf("view %d must be %+v, got %+v", i, want[i], view)
		}
	}
}
