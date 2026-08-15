package compatibility

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPythonMCPCollectionLoaderParsesGeneratedJSON(t *testing.T) {
	directory := t.TempDir()
	if _, err := (Writer{Exporter: Exporter{}}).WriteAll(
		t.Context(), directory, snapshotFixture(t)); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	python, err := pythonCommand()
	if err != nil {
		t.Fatal(err)
	}
	root := repositoryRoot(t)
	code := "from pathlib import Path\n" +
		"from mtga.collection import load_exported_collection\n" +
		"cards = load_exported_collection(Path(__import__('sys').argv[1]))\n" +
		"assert len(cards) == 2 and cards[0]['name'] == 'Alpha Card'\n"
	command := exec.Command(python, "-c", code, directory)
	command.Env = append(os.Environ(), "PYTHONPATH="+root)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("legacy MCP loader rejected generated JSON: %v\n%s", err, output)
	}
}

func pythonCommand() (string, error) {
	for _, candidate := range []string{"python3", "python"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("python is required for the legacy MCP contract test")
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}
