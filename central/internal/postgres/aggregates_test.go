package postgres

import (
	"context"
	"testing"
	"time"
)

// TestAggregationIsAtomicAndSuppressesDuplicates: events roll up into
// hourly buckets exactly once, replays add nothing, and expired
// events are purged without ever being mined.
func TestAggregationIsAtomicAndSuppressesDuplicates(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	repo := InstallationsRepo{Q: pool}
	if err := repo.Enroll(ctx, Installation{ID: "agg-1", TokenHash: "h",
		DeletionHash: "d"}); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	for batch, count := range map[string]int{"ab-1": 2, "ab-2": 1} {
		events := make([]TelemetryEvent, count)
		for i := range events {
			events[i] = TelemetryEvent{Name: "scan_completed",
				Attrs: map[string]string{"result": "ok"}}
		}
		if err := IngestBatch(ctx, pool, TelemetryBatch{ID: batch,
			InstallationID: "agg-1", Sequence: int64(len(batch) + count),
			Events: events}, time.Hour); err != nil {
			t.Fatalf("ingest %s: %v", batch, err)
		}
	}
	if _, err := AggregateEvents(ctx, pool); err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	totals, err := Totals(ctx, pool, "scan_completed")
	if err != nil || len(totals) != 1 || totals[0].Value != 3 {
		t.Fatalf("3 events must land in one hourly bucket: %+v %v", totals, err)
	}
	var remaining int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM accepted_events`).Scan(&remaining); err != nil ||
		remaining != 0 {
		t.Fatalf("consumed events must be gone: %d %v", remaining, err)
	}
	if _, err := AggregateEvents(ctx, pool); err != nil {
		t.Fatalf("replay: %v", err)
	}
	totals, _ = Totals(ctx, pool, "scan_completed")
	if totals[0].Value != 3 {
		t.Fatalf("replayed aggregation must add nothing: %+v", totals)
	}
	// An expired event is purged, never mined.
	if err := IngestBatch(ctx, pool, TelemetryBatch{ID: "ab-3",
		InstallationID: "agg-1", Sequence: 99, Events: []TelemetryEvent{
			{Name: "scan_completed"}}}, -time.Hour); err != nil {
		t.Fatalf("ingest expired: %v", err)
	}
	purged, err := PurgeExpiredEvents(ctx, pool)
	if err != nil || purged != 1 {
		t.Fatalf("expired event must purge: %d %v", purged, err)
	}
	totals, _ = Totals(ctx, pool, "scan_completed")
	if totals[0].Value != 3 {
		t.Fatalf("purged data must never be mined: %+v", totals)
	}
}
