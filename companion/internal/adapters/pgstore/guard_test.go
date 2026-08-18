package pgstore

import (
	"strings"
	"testing"
)

// TestProductionIdentitiesAreRejectedBeforeDialing: any DSN carrying
// a production marker is refused by Open without a single dial —
// there is no network in this test at all.
func TestProductionIdentitiesAreRejectedBeforeDialing(t *testing.T) {
	refused := []string{
		"postgres://central@db-prod.internal:5432/central",
		"postgres://user:pw@host/central_production",
		"postgres://user@release-db.example/wf",
		"postgres://user@db.example/live",
	}
	for _, dsn := range refused {
		if _, err := Open(dsn); err == nil ||
			!strings.Contains(err.Error(), "production marker") {
			t.Fatalf("production dsn must be refused before dialing: %s (%v)",
				dsn, err)
		}
	}
	isolated := []string{
		"postgres://postgres:postgres@127.0.0.1:5432/wf_test",
		"postgres://ci@localhost/companion_ci",
	}
	for _, dsn := range isolated {
		if err := GuardDSN(dsn); err != nil {
			t.Fatalf("isolated dsn must pass the guard: %s (%v)", dsn, err)
		}
	}
}
