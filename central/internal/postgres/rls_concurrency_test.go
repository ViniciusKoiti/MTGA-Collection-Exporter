package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
)

// TestConcurrentDuplicateBatchHasExactlyOneWinner: idempotency under
// concurrency — the unique (installation, sequence) constraint lets
// exactly one of two simultaneous duplicates through (task 2.6).
func TestConcurrentDuplicateBatchHasExactlyOneWinner(t *testing.T) {
	dsn := containerDSN(t)
	ctx := context.Background()
	if err := RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(ctx, `INSERT INTO installations
		(id, token_hash, deletion_hash) VALUES ('inst-1','t1','d1')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var wg sync.WaitGroup
	winners := make([]error, 2)
	for i := range winners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, winners[i] = db.ExecContext(ctx, `INSERT INTO telemetry_batches
				(id, installation_id, sequence) VALUES ($1,'inst-1',9)`,
				fmt.Sprintf("b-9-%d", i))
		}()
	}
	wg.Wait()
	if (winners[0] == nil) == (winners[1] == nil) {
		t.Fatalf("exactly one concurrent duplicate must win: %v / %v",
			winners[0], winners[1])
	}
}
