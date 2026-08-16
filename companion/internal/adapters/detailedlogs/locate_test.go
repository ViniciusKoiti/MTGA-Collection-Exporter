package detailedlogs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProbeFileHandlesAbsenceRealLogsAndFutureLogs(t *testing.T) {
	missing, err := ProbeFile(filepath.Join(t.TempDir(), "Player.log"))
	if err != nil || missing.CollectionPayload {
		t.Fatalf("absence must probe empty without error: %+v %v", missing, err)
	}
	current, err := ProbeFile("testdata/current_client_boot.log")
	if err != nil || current.CollectionPayload {
		t.Fatalf("the real capture must probe empty: %+v %v", current, err)
	}
	future, err := ProbeFile("testdata/hypothetical_v3_collection.log")
	if err != nil || !future.CollectionPayload || future.PayloadVersion != "v3" {
		t.Fatalf("the future capture must probe full: %+v %v", future, err)
	}
}

// TestProbeFileReadsOnlyTheTailOfHugeLogs: a payload near the end of
// an oversized log is still seen; the read stays bounded.
func TestProbeFileReadsOnlyTheTailOfHugeLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Player.log")
	payload, err := os.ReadFile("testdata/hypothetical_v3_collection.log")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	padding := strings.Repeat("[noise] filler line\n", 500_000)
	if err := os.WriteFile(path,
		append([]byte(padding), payload...), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	caps, err := ProbeFile(path)
	if err != nil || !caps.CollectionPayload {
		t.Fatalf("a tail payload must be seen in a huge log: %+v %v",
			caps, err)
	}
}

func TestDefaultLogPathPointsAtTheUnityLocalLow(t *testing.T) {
	path, err := DefaultLogPath()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if !strings.Contains(path, "LocalLow") ||
		!strings.HasSuffix(path, "Player.log") {
		t.Fatalf("unexpected default path: %s", path)
	}
}
