package setup

import (
	"context"
	"errors"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

type fakeProbe struct {
	source collection.SourceKind
	err    error
}

func (f fakeProbe) Decide(context.Context) (collection.SourceKind, error) {
	return f.source, f.err
}

type fakeVerify struct {
	failing map[collection.SourceKind]bool
}

func (f fakeVerify) Verify(_ context.Context,
	source collection.SourceKind) error {
	if f.failing[source] {
		return errors.New("not operable")
	}
	return nil
}

func TestCurrentClientLandsOnVerifiedExplicitImportWithGuidance(t *testing.T) {
	flow := Flow{Probe: fakeProbe{source: collection.SourceJSONImport},
		Verify: fakeVerify{}}
	result := flow.Run(context.Background(), PrivacyChoices{})
	if result.Source != collection.SourceJSONImport || !result.Verified ||
		!result.NeedsGuidance || result.FellBack {
		t.Fatalf("import must verify with guidance shown: %+v", result)
	}
	if result.Privacy.TelemetryOptIn || result.Privacy.AssistantEnabled {
		t.Fatalf("nothing may default to sharing: %+v", result.Privacy)
	}
}

func TestWorkingDetailedLogsNeedNoGuidance(t *testing.T) {
	flow := Flow{Probe: fakeProbe{source: collection.SourceDetailedLogs},
		Verify: fakeVerify{}}
	result := flow.Run(context.Background(),
		PrivacyChoices{TelemetryOptIn: true})
	if result.Source != collection.SourceDetailedLogs || !result.Verified ||
		result.NeedsGuidance || result.FellBack {
		t.Fatalf("verified logs must serve without guidance: %+v", result)
	}
	if !result.Privacy.TelemetryOptIn {
		t.Fatal("explicit choices must be carried through")
	}
}

// TestUnverifiableLogsFallBackExplicitly: the fallback is visible
// (FellBack), guidance turns on, and the import still verifies.
func TestUnverifiableLogsFallBackExplicitly(t *testing.T) {
	flow := Flow{Probe: fakeProbe{source: collection.SourceDetailedLogs},
		Verify: fakeVerify{failing: map[collection.SourceKind]bool{
			collection.SourceDetailedLogs: true}}}
	result := flow.Run(context.Background(), PrivacyChoices{})
	if result.Source != collection.SourceJSONImport || !result.FellBack ||
		!result.NeedsGuidance || !result.Verified {
		t.Fatalf("the fallback must be explicit and verified: %+v", result)
	}
}

func TestProbeFailureStillYieldsAWorkingSetup(t *testing.T) {
	flow := Flow{Probe: fakeProbe{err: errors.New("log unreadable")},
		Verify: fakeVerify{}}
	result := flow.Run(context.Background(), PrivacyChoices{})
	if result.Source != collection.SourceJSONImport || !result.Verified {
		t.Fatalf("a broken probe must not break first-run: %+v", result)
	}
}
