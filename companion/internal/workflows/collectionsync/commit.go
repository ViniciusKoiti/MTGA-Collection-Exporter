package collectionsync

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// persiste constrói o snapshot imutável e o grava no store autoritativo.
// O ID é determinístico sob o relógio controlado do harness.
func persiste(deps Deps) wf.Node {
	return no{"persiste", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		agora := deps.Clock.Now()
		id := collection.SnapshotID(fmt.Sprintf("snap-%d-%s",
			agora.Unix(), estado.Observacao.SourceInstance))
		if _, err := deps.Snapshots.Get(ctx, id); err == nil {
			// Idempotent orchestration (task 3.3): the same observation at
			// the same instant is already committed — reuse it.
			estado.Snapshot = id
			return estado, "ok", nil
		}
		snap, err := collection.NewSnapshot(id, estado.Observacao, agora,
			estado.Entradas, estado.Diagnosticos)
		if err != nil {
			return estado, "", err
		}
		if err := deps.Snapshots.Save(ctx, snap); err != nil {
			return estado, "", err
		}
		estado.Snapshot = id
		return estado, "ok", nil
	}}
}

// projeta gera as projeções de compatibilidade a partir do snapshot
// recém-commitado; a escrita em disco é um efeito aprovado à parte
// (grafo approved-export).
func projeta(deps Deps) wf.Node {
	return no{"projeta", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		snap, err := deps.Snapshots.Get(ctx, estado.Snapshot)
		if err != nil {
			return estado, "", err
		}
		estado.Projecoes = make(map[string]int, 3)
		for _, formato := range []ports.ExportFormat{ports.ExportJSON, ports.ExportCSV, ports.ExportText} {
			saida, err := deps.Exporter.Export(ctx, snap, formato)
			if err != nil {
				return estado, "", err
			}
			estado.Projecoes[string(formato)] = len(saida)
		}
		return estado, "ok", nil
	}}
}
