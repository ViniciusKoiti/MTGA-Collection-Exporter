package loadtest

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/httpapi"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// pgTelemetry adapts the real ingestion to the handler port.
type pgTelemetry struct{ pool *pgxpool.Pool }

func (s pgTelemetry) IngestBatch(ctx context.Context, installationID,
	batchID string, sequence int64,
	events []httpapi.TelemetryEventInput) error {
	batch := postgres.TelemetryBatch{ID: batchID,
		InstallationID: installationID, Sequence: sequence}
	for _, event := range events {
		batch.Events = append(batch.Events, postgres.TelemetryEvent{
			Name: event.Name, Attrs: event.Attrs})
	}
	err := postgres.IngestBatch(ctx, s.pool, batch, time.Hour)
	if errors.Is(err, postgres.ErrDuplicate) {
		return httpapi.ErrDuplicateBatch
	}
	return err
}
