package collection

import (
	"fmt"
	"time"
)

// SnapshotID identifica um snapshot imutável de coleção.
type SnapshotID string

// SnapshotSchema é a versão corrente do modelo de snapshot.
const SnapshotSchema = "collection-snapshot/v1"

// Entry é uma linha do snapshot: identidade (resolvida ou não), quantidade
// e confiança. Cartas desconhecidas permanecem visíveis como não
// resolvidas em vez de descartadas silenciosamente (decisão 3 do design).
type Entry struct {
	Identity   CardIdentity
	Quantity   int
	Unresolved bool
	Raw        string // identidade crua preservada quando não resolvida
}

// Snapshot é o resultado imutável e autoritativo de uma sincronização.
// Depois de construído por NewSnapshot ele nunca muda; qualquer correção
// gera um snapshot novo.
type Snapshot struct {
	ID             SnapshotID
	Schema         string
	Source         SourceKind
	SourceInstance string
	ObservedAt     time.Time
	ImportedAt     time.Time
	Entries        []Entry
	Diagnostics    []Diagnostic
}

// NewSnapshot valida os invariantes e devolve o snapshot com cópias
// defensivas: entradas resolvidas exigem identidade completa e quantidade
// positiva; não resolvidas exigem a identidade crua preservada.
func NewSnapshot(
	id SnapshotID,
	obs Observation,
	importedAt time.Time,
	entries []Entry,
	diagnostics []Diagnostic,
) (Snapshot, error) {
	if id == "" {
		return Snapshot{}, fmt.Errorf("collection: snapshot sem ID")
	}
	if importedAt.Before(obs.ObservedAt) {
		return Snapshot{}, fmt.Errorf("collection: importação anterior à observação")
	}
	for i, e := range entries {
		if e.Quantity < 1 {
			return Snapshot{}, fmt.Errorf("collection: entrada %d com quantidade inválida %d", i, e.Quantity)
		}
		if e.Unresolved && e.Raw == "" {
			return Snapshot{}, fmt.Errorf("collection: entrada %d não resolvida sem identidade crua", i)
		}
		if !e.Unresolved && !e.Identity.Resolvida() {
			return Snapshot{}, fmt.Errorf("collection: entrada %d resolvida sem identidade completa", i)
		}
	}
	return Snapshot{
		ID:             id,
		Schema:         SnapshotSchema,
		Source:         obs.Source,
		SourceInstance: obs.SourceInstance,
		ObservedAt:     obs.ObservedAt,
		ImportedAt:     importedAt,
		Entries:        append([]Entry(nil), entries...),
		Diagnostics:    append([]Diagnostic(nil), diagnostics...),
	}, nil
}

// TotalCartas soma as quantidades, incluindo entradas não resolvidas.
func (s Snapshot) TotalCartas() int {
	total := 0
	for _, e := range s.Entries {
		total += e.Quantity
	}
	return total
}
