package homesvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

type fakeReader struct {
	latest   collection.Snapshot
	previous collection.Snapshot
	hasBoth  bool
	err      error
}

func (f fakeReader) Latest(context.Context) (collection.Snapshot, bool, error) {
	return f.latest, f.err == nil && len(f.latest.Entries) > 0, f.err
}

func (f fakeReader) Previous(context.Context) (collection.Snapshot, bool, error) {
	return f.previous, f.hasBoth, nil
}

func snapshotWith(total int, at time.Time) collection.Snapshot {
	return collection.Snapshot{ImportedAt: at,
		Entries: []collection.Entry{{Quantity: total}}}
}

func depsWith(reader SnapshotReader, now time.Time) Deps {
	return Deps{
		Presence: func(context.Context) bool { return true },
		Source: func(context.Context) (collection.SourceKind, bool) {
			return collection.SourceJSONImport, true
		},
		Reader: reader,
		Now:    func() time.Time { return now },
	}
}

func TestFreshSnapshotRendersTotalsDeltaAndSuccess(t *testing.T) {
	now := time.Unix(1_700_100_000, 0)
	reader := fakeReader{latest: snapshotWith(1200, now.Add(-time.Hour)),
		previous: snapshotWith(1150, now.Add(-30*time.Hour)), hasBoth: true}
	model := Compose(context.Background(), depsWith(reader, now))
	if !model.MTGARunning || !model.SourceHealthy || model.Totals != 1200 {
		t.Fatalf("presence, health and totals must render: %+v", model)
	}
	if model.Delta == nil || *model.Delta != 50 {
		t.Fatalf("the delta against the previous snapshot must be 50: %+v",
			model.Delta)
	}
	if model.State.Status != viewstate.StatusSuccess || len(model.Actions) != 0 {
		t.Fatalf("a fresh snapshot needs no recovery: %+v", model)
	}
}

func TestStaleSnapshotOffersResync(t *testing.T) {
	now := time.Unix(1_700_100_000, 0)
	reader := fakeReader{latest: snapshotWith(900, now.Add(-48*time.Hour))}
	model := Compose(context.Background(), depsWith(reader, now))
	if model.State.Status != viewstate.StatusStale ||
		len(model.Actions) != 1 || model.Actions[0].ID != "resync" {
		t.Fatalf("stale must offer resync in context: %+v", model)
	}
	if model.Delta != nil {
		t.Fatalf("without history the delta must be unknown: %+v", model.Delta)
	}
}

func TestEmptyAndErrorPathsStillRenderWithRecovery(t *testing.T) {
	now := time.Unix(1_700_100_000, 0)
	empty := Compose(context.Background(), depsWith(fakeReader{}, now))
	if empty.State.Status != viewstate.StatusEmpty ||
		empty.Actions[0].ID != "import_collection" {
		t.Fatalf("no snapshot must offer the import action: %+v", empty)
	}
	broken := Compose(context.Background(),
		depsWith(fakeReader{err: errors.New("db locked")}, now))
	if broken.State.Status != viewstate.StatusError ||
		broken.Actions[0].ID != "retry_sync" {
		t.Fatalf("a read failure must offer retry: %+v", broken)
	}
}
