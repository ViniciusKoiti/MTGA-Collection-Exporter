// Command migrate applies the central forward migrations through the
// dedicated migrator path (OpenSpec add-central-go-platform, task 2.5):
// cluster-wide advisory lock, ahead-of-binary preflight, apply, verify.
// API replicas never migrate — they gate readiness on SchemaReady.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		slog.Error("migrate: invalid configuration", "err", err)
		os.Exit(1)
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"migrate: configuration validated", cfg.Diagnostics()...)

	// The DSN is resolved by the deployment from the secret reference in
	// the configuration and injected as CENTRAL_DB_DSN.
	dsn := cfg.DBDsn
	if dsn == "" {
		slog.Info("migrate: CENTRAL_DB_DSN not set; nothing to do")
		return
	}
	if err := postgres.RunMigrator(context.Background(), dsn); err != nil {
		slog.Error("migrate: failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migrate: schema is current and verified")
}
