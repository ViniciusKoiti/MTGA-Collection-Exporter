package approvedexport

import (
	"context"
	"crypto/sha256"
	"fmt"

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

// carrega fixa o preview imutável: snapshot corrente, payload projetado e
// seu hash exato. A aprovação que vier depois cobre este hash e nada mais.
func carrega(deps Deps) wf.Node {
	return no{"carrega", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		snap, existe, err := deps.Snapshots.Latest(ctx)
		if err != nil {
			return estado, "", err
		}
		if !existe {
			return estado, "sem_snapshot", nil
		}
		saida, err := deps.Exporter.Export(ctx, snap, estado.Formato)
		if err != nil {
			return estado, "", err
		}
		soma := sha256.Sum256(saida)
		estado.Snapshot = snap.ID
		estado.Hash = fmt.Sprintf("%x", soma)
		estado.Bytes = len(saida)
		return estado, "ok", nil
	}}
}

// aprova apresenta o preview exato ao effect gate: a política libera,
// nega, pausa aguardando aprovação ou expira — sempre com outcome estável.
func aprova() wf.Node {
	return no{"aprova", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		preview := wf.EffectPreview{
			Effect:      "export-write",
			Target:      estado.Destino,
			PayloadHash: estado.Hash,
		}
		if err := wf.RequestEffect(ctx, preview); err != nil {
			return estado, "", err
		}
		return estado, "ok", nil
	}}
}

// escreve enfileira o efeito aprovado no outbox com chave de idempotência
// derivada do hash: reexecução ou run repetido do mesmo payload não
// duplica a escrita (entrega at-least-once deduplicada).
func escreve(deps Deps) wf.Node {
	return no{"escreve", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		estado.EfeitoID = "export-" + estado.Hash
		efeito := wf.EffectRecord{
			ID: estado.EfeitoID,
			Preview: wf.EffectPreview{
				Effect:      "export-write",
				Target:      estado.Destino,
				PayloadHash: estado.Hash,
			},
			EnqueuedAt: deps.Clock.Now(),
		}
		if err := deps.Outbox.Enqueue(ctx, efeito); err != nil {
			return estado, "", err
		}
		return estado, "ok", nil
	}}
}
