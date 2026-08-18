package detailedlogs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// maxProbeBytes bounds how much of a log file the probe reads: the
// interesting payloads sit near the tail.
const maxProbeBytes = 8 << 20

// DefaultLogPath is where the current Windows client writes its
// Player.log (Unity LocalLow directory).
func DefaultLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "AppData", "LocalLow",
		"Wizards Of The Coast", "MTGA", "Player.log"), nil
}

// ProbeFile probes one log file, reading at most its last
// maxProbeBytes. A missing file probes empty — absence is a normal
// first-run condition, not an error.
func ProbeFile(path string) (Capabilities, error) {
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Capabilities{}, nil
	}
	if err != nil {
		return Capabilities{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Capabilities{}, err
	}
	if info.Size() > maxProbeBytes {
		if _, err := file.Seek(-maxProbeBytes, io.SeekEnd); err != nil {
			return Capabilities{}, err
		}
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxProbeBytes))
	if err != nil {
		return Capabilities{}, err
	}
	return Probe(raw), nil
}
