package activity

import "testing"

// TestInventarioCompletoDoCompanion congela o inventário da tarefa 1.3:
// toda atividade externa mapeada, consultas puras justificadas.
func TestInventarioCompletoDoCompanion(t *testing.T) {
	inventario, err := Inventario()
	if err != nil {
		t.Fatalf("inventário: %v", err)
	}
	esperadas := map[string]Kind{
		"sync-collection":  KindGraph,
		"recommend-decks":  KindGraph,
		"export-approved":  KindGraph,
		"flush-telemetry":  KindGraph,
		"run-dev-scenario": KindGraph,
		"query-collection": KindPureQuery,
		"query-stats":      KindPureQuery,
		"search-cards":     KindPureQuery,
		"check-deck":       KindPureQuery,
	}
	for nome, tipo := range esperadas {
		entrada, err := inventario.Resolve(nome)
		if err != nil || entrada.Kind != tipo {
			t.Errorf("%s: esperava %s, veio %+v (%v)", nome, tipo, entrada, err)
		}
		if tipo == KindPureQuery && entrada.Justification == "" {
			t.Errorf("%s: consulta pura sem justificativa", nome)
		}
		if tipo == KindGraph && (entrada.Graph.Kind == "" || entrada.Graph.Version < 1) {
			t.Errorf("%s: atividade de grafo sem identidade: %+v", nome, entrada)
		}
	}
}
