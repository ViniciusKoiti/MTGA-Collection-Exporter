package arch

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkflowFoundationPackagesExist pins the layering precondition
// of the graph-workflow change (add-graph-workflow-harness, task 1.2):
// the domain, application, policy and storage-port packages the
// workflow adapters build on must exist and parse as valid Go — a
// rename or removal fails here before any adapter breaks.
func TestWorkflowFoundationPackagesExist(t *testing.T) {
	foundations := []string{
		filepath.Join("..", "domain"),
		filepath.Join("..", "application"),
		filepath.Join("..", "policy"),
		filepath.Join("..", "ports"),
		filepath.Join("..", "adapters", "sqlitestore"),
		filepath.Join("..", "workflow"),
	}
	for _, dir := range foundations {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("foundation package missing: %s (%v)", dir, err)
		}
		parsed := false
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			if _, err := parser.ParseFile(token.NewFileSet(), path, nil,
				parser.PackageClauseOnly); err != nil {
				t.Fatalf("foundation file does not parse: %s (%v)", path, err)
			}
			parsed = true
		}
		if !parsed && !hasGoSubpackages(t, dir) {
			t.Fatalf("foundation package %s holds no Go source", dir)
		}
	}
}

func hasGoSubpackages(t *testing.T, dir string) bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sub, err := filepath.Glob(filepath.Join(dir, entry.Name(), "*.go"))
		if err == nil && len(sub) > 0 {
			return true
		}
	}
	return false
}
