package compatibility

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

func TestWriterAtomicallyReplacesCompatibilityFiles(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{JSONName, CSVName, TextName} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("old"), 0o600); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	paths, err := (Writer{Exporter: Exporter{}}).WriteAll(
		t.Context(), directory, snapshotFixture(t))
	if err != nil {
		t.Fatalf("write all: %v", err)
	}
	for format, name := range exportNames {
		payload, err := os.ReadFile(paths[format])
		if err != nil || string(payload) == "old" || filepath.Base(paths[format]) != name {
			t.Fatalf("%s was not replaced correctly: %q (%v)", name, payload, err)
		}
	}
	temporary, err := filepath.Glob(filepath.Join(directory, ".mtga-export-*"))
	if err != nil || len(temporary) != 0 {
		t.Fatalf("temporary files leaked: %v (%v)", temporary, err)
	}
	if paths[ports.ExportJSON] != filepath.Join(directory, JSONName) {
		t.Fatalf("unexpected JSON path: %s", paths[ports.ExportJSON])
	}
}
