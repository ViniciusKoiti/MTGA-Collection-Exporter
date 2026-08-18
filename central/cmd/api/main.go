// Command api é o processo HTTP stateless do serviço central.
// Nesta fase apenas valida a configuração estrita e emite diagnósticos
// redigidos; o servidor net/http com limites chega na tarefa 3.1 do
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
		slog.Error("api: falha na inicialização", "err", err)
		os.Exit(1)
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"api: configuração validada", cfg.Diagnostics()...)
}
