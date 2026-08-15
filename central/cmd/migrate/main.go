// Command migrate aplica migrações SQL forward com papel dedicado de
// migrador. Nesta fase apenas valida a configuração estrita; lock,
// preflight e verificação chegam na tarefa 2.5 do OpenSpec
// add-central-go-platform.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
)

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		slog.Error("migrate: falha na inicialização", "err", err)
		os.Exit(1)
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"migrate: configuração validada", cfg.Diagnostics()...)
}
