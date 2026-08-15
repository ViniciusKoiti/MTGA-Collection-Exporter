package publication

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
)

func budgets() config.Budgets {
	return config.Budgets{
		FetchWorkers: 4, NormalizeWorkers: 4,
		ProviderBuffer: 2, ObservationBuffer: 4, NormalizedBuffer: 4,
		BatchSize: 10,
	}
}

func jobs(n int) []catalog.ProviderJob {
	js := make([]catalog.ProviderJob, n)
	for i := range js {
		js[i] = catalog.ProviderJob{Provider: fmt.Sprintf("prov-%d", i), Kind: "cards"}
	}
	return js
}

func TestRunPublicaOrdenadoEmLotes(t *testing.T) {
	fake := newFakeDeps(25, nil)
	man, err := Run(t.Context(), budgets(), "v1", jobs(8), fake.deps())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cards := fake.cartasEscritas()
	if len(cards) != 8*25 {
		t.Fatalf("cartas persistidas: %d", len(cards))
	}
	for i := 1; i < len(cards); i++ {
		if cards[i-1].GrpID > cards[i].GrpID {
			t.Fatalf("redução não determinística na posição %d", i)
		}
	}
	for i, lote := range fake.lotes {
		if len(lote) > budgets().BatchSize {
			t.Fatalf("lote %d excede o orçamento: %d", i, len(lote))
		}
	}
	if fake.ativacoes != 1 || man.Object.SHA256 == "" {
		t.Fatalf("publicação deveria ativar exatamente um manifesto: %+v", man)
	}
}

func TestRunFalhaDeAtivacaoNaoViraCorrente(t *testing.T) {
	falha := errors.New("transação de ativação falhou")
	fake := newFakeDeps(5, falha)
	if _, err := Run(t.Context(), budgets(), "v1", jobs(2), fake.deps()); !errors.Is(err, falha) {
		t.Fatalf("esperava erro de ativação, veio: %v", err)
	}
	if fake.ativacoes != 0 {
		t.Fatal("manifesto não pode virar corrente quando a ativação falha")
	}
}

func TestRunCancelamentoParaTudo(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	fake := newFakeDeps(5, nil)
	if _, err := Run(ctx, budgets(), "v1", jobs(4), fake.deps()); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, veio: %v", err)
	}
	if fake.ativacoes != 0 {
		t.Fatal("execução cancelada não pode ativar manifesto")
	}
}
