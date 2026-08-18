package metarecommend

import (
	"context"
	"sort"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// no adapta uma função tipada sobre Estado ao contrato wf.Node.
type no struct {
	id wf.NodeID
	fn func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error)
}

func (n no) ID() wf.NodeID { return n.id }
func (n no) Execute(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	estado, outcome, err := n.fn(ctx, estadoDe(st))
	return estado, outcome, err
}

// recomenda carrega snapshot + catálogo, compara e ranqueia em um único
// passo determinístico (a comparação em si é função pura de domínio).
func recomenda(deps Deps) wf.Node {
	return no{"recomenda", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		snap, existe, err := deps.Snapshots.Latest(ctx)
		if err != nil {
			return estado, "", err
		}
		if !existe {
			return estado, "sem_snapshot", nil
		}
		estado.Snapshot = snap.ID
		metas, err := deps.Meta.Decks(ctx, estado.Formato)
		if err != nil {
			return estado, "", err
		}
		if len(metas) == 0 {
			return estado, "sem_catalogo", nil
		}
		for _, meta := range metas {
			estado.Recomendacoes = append(estado.Recomendacoes, avalia(meta, snap, deps.MaxLacunas))
			estado.Avaliados++
		}
		sort.SliceStable(estado.Recomendacoes, func(i, j int) bool {
			a, b := estado.Recomendacoes[i], estado.Recomendacoes[j]
			if a.Montabilidade != b.Montabilidade {
				return a.Montabilidade > b.Montabilidade
			}
			return a.Nome < b.Nome
		})
		return estado, "ok", nil
	}}
}

// avalia compara um meta deck com o snapshot e produz a linha do ranking.
func avalia(meta ports.MetaDeck, snap collection.Snapshot, maxLacunas int) Recomendacao {
	posse := decks.CompareOwnership(meta.Deck, snap)
	exigido := 0
	for _, linha := range posse.Lines {
		exigido += linha.Required
	}
	rec := Recomendacao{Nome: meta.Nome, Fonte: meta.Fonte, Faltantes: posse.TotalMissing}
	if exigido > 0 {
		rec.Montabilidade = (exigido - posse.TotalMissing) * 100 / exigido
	}
	for _, linha := range posse.Lines { // Lines já vem ordenado por nome
		if linha.Missing > 0 && len(rec.Lacunas) < maxLacunas {
			rec.Lacunas = append(rec.Lacunas, linha.Name)
		}
	}
	return rec
}
