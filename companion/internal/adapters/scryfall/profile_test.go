package scryfall

import (
	"slices"
	"testing"
	"time"
)

// TestProfilesComeFromTheBulkMetadata: a card with catalog metadata
// yields the full substitution profile; one without (paper-only entry
// carries none in the fixture) degrades to ok=false.
func TestProfilesComeFromTheBulkMetadata(t *testing.T) {
	ts := server(t, nil)
	now := time.Unix(1_700_000_000, 0)
	catalog, err := Open(t.Context(), client(ts.URL, t.TempDir(), 1),
		now, time.Hour)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	profile, ok := catalog.Profile("LIGHTNING strike")
	if !ok || profile.ManaValue != 2 ||
		!slices.Equal(profile.Colors, []string{"R"}) ||
		!slices.Equal(profile.Types, []string{"Instant"}) {
		t.Fatalf("the profile must map colors, mana and types: %+v %v",
			profile, ok)
	}
	if !slices.Contains(profile.Formats, "standard") ||
		slices.Contains(profile.Formats, "pauper") {
		t.Fatalf("only legal formats may tag the profile: %v", profile.Formats)
	}
	if _, ok := catalog.Profile("Paper Only Card"); ok {
		t.Fatal("cards without metadata must degrade to no profile")
	}
	if _, ok := catalog.Profile("Unknown Card"); ok {
		t.Fatal("unknown cards have no profile")
	}
}
