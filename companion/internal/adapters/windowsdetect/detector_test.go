package windowsdetect

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeProcesses struct {
	names []string
	err   error
}

func (f fakeProcesses) Names(context.Context) ([]string, error) {
	return f.names, f.err
}

func TestInspectFindsDocumentedLogAndMTGAProcess(t *testing.T) {
	home := t.TempDir()
	detector := Detector{Home: home, Processes: fakeProcesses{
		names: []string{"explorer.exe", `C:\\Games\\mtga.EXE`},
	}}
	logPath := detector.PlayerLogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log directory: %v", err)
	}
	if err := os.WriteFile(logPath, nil, 0o600); err != nil {
		t.Fatalf("create player log: %v", err)
	}
	status, err := detector.Inspect(t.Context())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !status.LogAvailable || !status.MTGARunning || status.PlayerLog != logPath {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestInspectDoesNotClaimMissingResources(t *testing.T) {
	detector := Detector{Home: t.TempDir(), Processes: fakeProcesses{
		names: []string{"MTGALauncher.exe", "not-mtga.exe"},
	}}
	status, err := detector.Inspect(t.Context())
	if err != nil || status.LogAvailable || status.MTGARunning {
		t.Fatalf("missing resources were misclassified: %+v (%v)", status, err)
	}
}
