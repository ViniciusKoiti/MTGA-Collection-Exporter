// Command companion é o ponto de entrada do desktop Go (Wails chega na
// tarefa 4.1 do OpenSpec introduce-agentic-go-companion). O comando nunca
// importa o runtime de grafos diretamente: toda atividade externa passa
// pelo inventário de internal/activity (teste de arquitetura garante).
package main

import (
	"log/slog"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
)

func main() {
	classificador := policy.NewClassifier(
		[]string{"query-collection", "query-stats", "search-cards", "check-deck"},
		[]string{"export-approved", "sync-collection"},
	)
	slog.Info("companion: esqueleto inicializado",
		"politica_default", string(classificador.Classify("ferramenta-desconhecida")))
}
