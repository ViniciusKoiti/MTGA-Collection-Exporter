//go:build windows

package main

import (
	"context"
	"errors"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/detailedlogs"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/setup"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// clientProbe probes the installed client's Player.log.
type clientProbe struct{}

func (clientProbe) Decide(context.Context) (collection.SourceKind, error) {
	path, err := detailedlogs.DefaultLogPath()
	if err != nil {
		return collection.SourceJSONImport, nil
	}
	caps, err := detailedlogs.ProbeFile(path)
	if err != nil {
		return collection.SourceJSONImport, nil
	}
	return detailedlogs.DecideSource(caps,
		detailedlogs.RecognizedPayloadVersions()), nil
}

// sourceVerifier: Detailed Logs verify only when the payload is
// really there; the explicit import is operable by definition — the
// user picks the file.
type sourceVerifier struct{}

func (sourceVerifier) Verify(_ context.Context,
	source collection.SourceKind) error {
	if source != collection.SourceDetailedLogs {
		return nil
	}
	path, err := detailedlogs.DefaultLogPath()
	if err != nil {
		return err
	}
	caps, err := detailedlogs.ProbeFile(path)
	if err != nil {
		return err
	}
	if !caps.CollectionPayload {
		return errors.New("detailed logs carry no collection payload")
	}
	return nil
}

// FirstRun drives the first-run setup against the installed client.
func (a *App) FirstRun(telemetryOptIn, assistantEnabled bool) setup.Result {
	flow := setup.Flow{Probe: clientProbe{}, Verify: sourceVerifier{}}
	return flow.Run(context.Background(), setup.PrivacyChoices{
		TelemetryOptIn:   telemetryOptIn,
		AssistantEnabled: assistantEnabled,
	})
}
