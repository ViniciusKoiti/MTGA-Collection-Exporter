package workflow

import "context"

// EffectExecutor aplica o efeito real (escrita de arquivo, chamada de API).
// A entrega do outbox é at-least-once: o executor DEVE ser idempotente pela
// chave EffectRecord.ID — reexecução após crash não pode duplicar o efeito.
type EffectExecutor interface {
	Execute(ctx context.Context, rec EffectRecord) error
}

// Dispatch drena o outbox em ordem: executa cada efeito pendente e só
// então confirma (tarefa 3.4). Os dois pontos de crash são cobertos:
//
//   - morte antes de Execute: o registro segue pendente e é reentregue;
//   - morte entre Execute e Ack: a reentrega chama Execute de novo e a
//     idempotência por ID garante efeito único.
//
// Qualquer erro interrompe o despacho; a próxima chamada continua de onde
// parou, pois só o Ack remove a pendência.
func Dispatch(ctx context.Context, outbox Outbox, exec EffectExecutor) (int, error) {
	pendentes, err := outbox.Pending(ctx)
	if err != nil {
		return 0, err
	}
	confirmados := 0
	for _, rec := range pendentes {
		if err := ctx.Err(); err != nil {
			return confirmados, err
		}
		if err := exec.Execute(ctx, rec); err != nil {
			return confirmados, err
		}
		if err := outbox.Ack(ctx, rec.ID); err != nil {
			return confirmados, err
		}
		confirmados++
	}
	return confirmados, nil
}
