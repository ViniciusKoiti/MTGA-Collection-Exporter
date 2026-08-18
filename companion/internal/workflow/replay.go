package workflow

import (
	"context"
	"fmt"
	"sync"
)

// ReplayResult é o resultado de um replay dry-run: desfecho, journal do
// store descartável e os efeitos que SERIAM executados — nenhum deles é
// despachado (tarefa 6.5; decisão 6 do design).
type ReplayResult struct {
	Run     Run
	Steps   []Step
	Efeitos []EffectPreview
}

// replayPolicy libera todo efeito registrando o preview exato; o replay
// nunca chega ao outbox nem ao executor, então nada é repetido de fato.
type replayPolicy struct {
	mu      sync.Mutex
	efeitos []EffectPreview
}

func (p *replayPolicy) Decide(_ context.Context, _ Identity, preview EffectPreview) (PolicyDecision, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.efeitos = append(p.efeitos, preview)
	return PolicyAllow, nil
}

// ReplayDryRun reexecuta o grafo do run `id` na MESMA versão pinada, sobre
// a fixture sanitizada dada e um store descartável fornecido pelo chamador.
// O store e o journal de produção não são tocados; eventos não são
// emitidos; efeitos são apenas colecionados no resultado.
func (e *Engine) ReplayDryRun(
	ctx context.Context,
	id RunID,
	fixture State,
	scratch RunStore,
) (ReplayResult, error) {
	original, err := e.store.Get(ctx, id)
	if err != nil {
		return ReplayResult{}, err
	}
	if !original.Status.Terminal() {
		return ReplayResult{}, fmt.Errorf("workflow: replay exige run terminal, %s está %s",
			id, original.Status)
	}
	if _, err := e.registry.Get(original.Graph); err != nil {
		return ReplayResult{}, fmt.Errorf("replay com versão pinada indisponível: %w", err)
	}
	registro := &replayPolicy{}
	sombra := &Engine{
		registry: e.registry,
		clock:    e.clock,
		ids:      replayIDs{base: id},
		store:    scratch,
		policy:   registro,
	}
	run, err := sombra.Start(ctx, original.Graph, fixture)
	if err != nil {
		return ReplayResult{}, err
	}
	steps, err := scratch.Steps(ctx, run.ID)
	if err != nil {
		return ReplayResult{}, err
	}
	return ReplayResult{Run: run, Steps: steps, Efeitos: registro.efeitos}, nil
}

// replayIDs deriva o ID do replay do run original, mantendo o dry-run
// determinístico e rastreável.
type replayIDs struct {
	base RunID
}

func (r replayIDs) NewRunID() RunID {
	return RunID("replay-" + string(r.base))
}
