package arch

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompanionImportsNoCentralPackage proves the desktop companion
// never links against the central platform (OpenSpec
// add-central-go-platform, task 1.4): the two deploy independently
// and may only talk over the versioned HTTP contract.
func TestCompanionImportsNoCentralPackage(t *testing.T) {
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil,
			parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			if strings.Contains(imported.Path.Value,
				"MTGA-Collection-Exporter/central") {
				t.Errorf("%s imports the central platform: %s",
					path, imported.Path.Value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
