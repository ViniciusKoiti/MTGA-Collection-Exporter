package publication

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// fakeDeps implementa todas as ports em memória, com mutex porque os pools
// chamam Fetch e Normalize concorrentemente.
type fakeDeps struct {
	mu           sync.Mutex
	porProvedor  int
	erroAtivacao error
	lotes        [][]catalog.Card
	ativacoes    int
	proximoGrp   int
}

func newFakeDeps(porProvedor int, erroAtivacao error) *fakeDeps {
	return &fakeDeps{porProvedor: porProvedor, erroAtivacao: erroAtivacao}
}

func (f *fakeDeps) deps() Deps {
	return Deps{Fetcher: f, Normalize: f, Writer: f, Objects: f, Signer: f, Activator: f}
}

func (f *fakeDeps) Fetch(_ context.Context, job catalog.ProviderJob) ([]catalog.Observation, error) {
	f.mu.Lock()
	base := f.proximoGrp
	f.proximoGrp += f.porProvedor
	f.mu.Unlock()
	obs := make([]catalog.Observation, f.porProvedor)
	for i := range obs {
		obs[i] = catalog.Observation{Provider: job.Provider, GrpID: base + i}
	}
	return obs, nil
}

func (f *fakeDeps) Normalize(_ context.Context, o catalog.Observation) (catalog.Card, error) {
	return catalog.Card{GrpID: o.GrpID, Name: fmt.Sprintf("card-%d", o.GrpID), Set: "TST"}, nil
}

func (f *fakeDeps) WriteBatch(_ context.Context, cards []catalog.Card) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lotes = append(f.lotes, cards)
	return nil
}

func (f *fakeDeps) WriteSnapshot(_ context.Context, s catalog.Snapshot) (catalog.ObjectRef, error) {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", s.SchemaVersion, len(s.Cards))))
	key := fmt.Sprintf("catalog/%x.json.gz", sum[:8])
	return catalog.ObjectRef{Key: key, SHA256: fmt.Sprintf("%x", sum), Size: int64(len(s.Cards))}, nil
}

func (f *fakeDeps) Sign(_ context.Context, ref catalog.ObjectRef) (catalog.Manifest, error) {
	return catalog.Manifest{KeyID: "k1", Signature: "assinado:" + ref.SHA256, Object: ref}, nil
}

func (f *fakeDeps) Activate(_ context.Context, _ catalog.Manifest) error {
	if f.erroAtivacao != nil {
		return f.erroAtivacao
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ativacoes++
	return nil
}

func (f *fakeDeps) cartasEscritas() []catalog.Card {
	f.mu.Lock()
	defer f.mu.Unlock()
	var todas []catalog.Card
	for _, lote := range f.lotes {
		todas = append(todas, lote...)
	}
	return todas
}
