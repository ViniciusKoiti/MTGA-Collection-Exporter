// Package windowsdetect locates MTGA logs and observes process presence
// without opening or reading the game process.
package windowsdetect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const processName = "MTGA.exe"

// ErrUnsupported reports that process discovery is unavailable on this OS.
var ErrUnsupported = errors.New("windowsdetect: process discovery requires Windows")

// ProcessSource lists executable names from the operating-system process table.
type ProcessSource interface {
	Names(ctx context.Context) ([]string, error)
}

// Status is the read-only environment state needed by setup and sync flows.
type Status struct {
	PlayerLog    string
	LogAvailable bool
	MTGARunning  bool
}

// Detector combines an explicit user home with a process-table source.
type Detector struct {
	Home      string
	Processes ProcessSource
}

// New builds the production Windows detector for the current user.
func New() (Detector, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Detector{}, fmt.Errorf("windowsdetect: resolve user home: %w", err)
	}
	return Detector{Home: home, Processes: systemProcesses{}}, nil
}

// PlayerLogPath returns the documented current-session MTGA log path.
func (d Detector) PlayerLogPath() string {
	return filepath.Join(d.Home, "AppData", "LocalLow", "Wizards Of The Coast",
		"MTGA", "player.log")
}

// Inspect reports path availability and process presence without reading either.
func (d Detector) Inspect(ctx context.Context) (Status, error) {
	status := Status{PlayerLog: d.PlayerLogPath()}
	if _, err := os.Stat(status.PlayerLog); err == nil {
		status.LogAvailable = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return Status{}, fmt.Errorf("windowsdetect: inspect player log: %w", err)
	}
	names, err := d.Processes.Names(ctx)
	if err != nil {
		return Status{}, err
	}
	status.MTGARunning = containsProcess(names, processName)
	return status, nil
}
