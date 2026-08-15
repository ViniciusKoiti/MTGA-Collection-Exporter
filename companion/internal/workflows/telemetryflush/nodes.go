package telemetryflush

import (
	"context"

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

// consente verifica o opt-in explícito; sem ele, nada sai do dispositivo.
func consente(deps Deps) wf.Node {
	return no{"consente", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		consentimento, err := deps.Consent.Current(ctx)
		if err != nil {
			return estado, "", err
		}
		if !consentimento.OptIn {
			return estado, "sem_optin", nil
		}
		return estado, "ok", nil
	}}
}

// minimiza monta o lote limitado aplicando a allowlist de atributos:
// chave fora da lista ou valor acima do teto é descartado do evento.
func minimiza(deps Deps) wf.Node {
	return no{"minimiza", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		pendentes, err := deps.Queue.Pending(ctx, deps.TamanhoLote)
		if err != nil {
			return estado, "", err
		}
		if len(pendentes) == 0 {
			return estado, "fila_vazia", nil
		}
		for _, evento := range pendentes {
			limpo := ports.TelemetryEvent{Seq: evento.Seq, Name: evento.Name, At: evento.At,
				Attrs: make(map[string]string, len(evento.Attrs))}
			for chave, valor := range evento.Attrs {
				if !deps.AtributosPermitidos[chave] || len(valor) > deps.MaxValorBytes {
					estado.Minimizados++
					continue
				}
				limpo.Attrs[chave] = valor
			}
			estado.Lote = append(estado.Lote, limpo)
			estado.UltimoSeq = evento.Seq
		}
		return estado, "ok", nil
	}}
}

// envia tenta o lote com retry limitado; esgotado o orçamento, os eventos
// permanecem locais (fallback local-only decidido pela transição).
func envia(deps Deps) wf.Node {
	return no{"envia", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		for estado.Tentativas < deps.MaxTentativas {
			if err := ctx.Err(); err != nil {
				return estado, "", err
			}
			estado.Tentativas++
			if err := deps.Sink.SendBatch(ctx, estado.Lote); err == nil {
				estado.Enviados = len(estado.Lote)
				return estado, "ok", nil
			}
		}
		return estado, "envio_esgotado", nil
	}}
}

// confirma faz o acknowledgement até o último seq enviado.
func confirma(deps Deps) wf.Node {
	return no{"confirma", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		if err := deps.Queue.Ack(ctx, estado.UltimoSeq); err != nil {
			return estado, "", err
		}
		return estado, "ok", nil
	}}
}
