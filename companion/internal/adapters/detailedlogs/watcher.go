package detailedlogs

import (
	"errors"
	"io"
	"os"
)

// FeatureFlagEnv gates the watcher; the log source ships DISABLED and
// only the exact value "true" turns it on (companion task 3.7).
const FeatureFlagEnv = "COMPANION_FLAG_DETAILED_LOGS"

// ErrFeatureDisabled is the stable refusal of a dark feature.
var ErrFeatureDisabled = errors.New(
	"detailedlogs: the log source is disabled by default; set " +
		FeatureFlagEnv + "=true to enable it")

// Watcher tails one log file incrementally: every Poll reads only the
// bytes appended since the previous one, and a truncated or rotated
// file restarts from zero instead of missing data.
type Watcher struct {
	path   string
	offset int64
}

// NewWatcher builds the watcher only when the feature flag is on.
func NewWatcher(path string, getenv func(string) string) (*Watcher, error) {
	if getenv(FeatureFlagEnv) != "true" {
		return nil, ErrFeatureDisabled
	}
	return &Watcher{path: path}, nil
}

// Poll returns the records appended since the last poll. A missing
// file yields no records — the game may simply not be running.
func (w *Watcher) Poll() ([]Record, error) {
	file, err := os.Open(w.path)
	if errors.Is(err, os.ErrNotExist) {
		w.offset = 0
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < w.offset {
		w.offset = 0 // truncation or rotation: start over
	}
	if info.Size() == w.offset {
		return nil, nil
	}
	if _, err := file.Seek(w.offset, io.SeekStart); err != nil {
		return nil, err
	}
	appended, err := io.ReadAll(io.LimitReader(file, maxProbeBytes))
	if err != nil {
		return nil, err
	}
	w.offset += int64(len(appended))
	return ParseRecords(appended), nil
}
