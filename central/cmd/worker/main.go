// Command worker executa jobs duráveis de publicação de catálogo e agregação.
// Nesta fase apenas valida a configuração estrita e emite diagnósticos
// redigidos; o poller com lease em PostgreSQL chega na tarefa 6.1/6.2 do
// OpenSpec add-central-go-platform.
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
		slog.Error("worker: falha na inicialização", "err", err)
		os.Exit(1)
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"worker: configuração validada", cfg.Diagnostics()...)
}
