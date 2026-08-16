package detailedlogs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func enabled(key string) string {
	if key == FeatureFlagEnv {
		return "true"
	}
	return ""
}

func TestWatcherShipsDisabledByDefault(t *testing.T) {
	if _, err := NewWatcher("x.log", func(string) string { return "" }); !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("the dark feature must refuse: %v", err)
	}
	if _, err := NewWatcher("x.log", func(string) string { return "TRUE" }); !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("only the exact value true may enable: %v", err)
	}
}

// TestWatcherPollsIncrementallyAndSurvivesRotation: each poll sees
// only the appended records, and truncation restarts from zero.
func TestWatcherPollsIncrementallyAndSurvivesRotation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Player.log")
	watcher, err := NewWatcher(path, enabled)
	if err != nil {
		t.Fatalf("watcher: %v", err)
	}
	if records, err := watcher.Poll(); err != nil || records != nil {
		t.Fatalf("a missing file polls empty: %v %v", records, err)
	}
	if err := os.WriteFile(path, fixture(t, "current_client_boot.log"),
		0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	first, err := watcher.Poll()
	if err != nil || len(first) != 12 {
		t.Fatalf("first poll must parse the fixture (12 records): %d %v",
			len(first), err)
	}
	appended := "[999] [UnityCrossThreadLogger]==> RankGetSeasonAndRankDetails " +
		`{"id":"00000000-0000-4000-8000-000000000099"}` + "\n"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open append: %v", err)
	}
	if _, err := file.WriteString(appended); err != nil {
		t.Fatalf("append: %v", err)
	}
	_ = file.Close()
	second, err := watcher.Poll()
	if err != nil || len(second) != 1 ||
		second[0].Method != "RankGetSeasonAndRankDetails" {
		t.Fatalf("polls must be incremental: %+v %v", second, err)
	}
	if err := os.WriteFile(path, []byte(appended), 0o644); err != nil {
		t.Fatalf("rotate: %v", err)
	}
	rotated, err := watcher.Poll()
	if err != nil || len(rotated) != 1 {
		t.Fatalf("rotation must restart from zero, not miss data: %+v %v",
			rotated, err)
	}
}
