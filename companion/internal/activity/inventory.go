package activity

import wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"

// Inventario monta o inventário canônico de atividades do companion
// (tarefa 1.3; espelho executável de docs/activity-inventory.md). A
// composição do processo chama esta função e depois Validate contra o
// registry de grafos — atividade sem grafo registrado reprova o boot.
func Inventario() (*Registry, error) {
	inventario := New()
	grafos := map[string]wf.Identity{
		"sync-collection":  {Kind: "collection-sync", Version: 1},
		"recommend-decks":  {Kind: "meta-deck-recommendation", Version: 1},
		"export-approved":  {Kind: "approved-export", Version: 1},
		"flush-telemetry":  {Kind: "telemetry-flush", Version: 1},
		"run-dev-scenario": {Kind: "development-scenario", Version: 1},
	}
	for nome, graph := range grafos {
		if err := inventario.RegisterGraph(nome, graph); err != nil {
			return nil, err
		}
	}
	consultas := map[string]string{
		"query-collection": "leitura direta do snapshot exportado, sem efeito",
		"query-stats":      "agregação determinística em memória, sem efeito",
		"search-cards":     "filtro puro sobre o snapshot, sem efeito",
		"check-deck":       "aritmética de posse sem efeito nem estado resumível",
	}
	for nome, justificativa := range consultas {
		if err := inventario.RegisterPureQuery(nome, justificativa); err != nil {
			return nil, err
		}
	}
	return inventario, nil
}
