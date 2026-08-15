package arch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackagingManifestsExcludeTheDevMCP: the development MCP command
// must never ship (add-graph-workflow-harness, task 7.1) — every
// packaging manifest and release workflow is scanned for its name.
func TestPackagingManifestsExcludeTheDevMCP(t *testing.T) {
	patterns := []string{
		filepath.Join("..", "..", "..", "scripts", "*"),
		filepath.Join("..", "..", "..", ".github", "workflows", "*"),
	}
	scanned := 0
	for _, pattern := range patterns {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			scanned++
			if strings.Contains(string(raw), "dev-mcp") {
				t.Errorf("packaging manifest %s mentions dev-mcp", path)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no packaging manifest scanned; the gate proves nothing")
	}
}
