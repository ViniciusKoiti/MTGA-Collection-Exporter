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
	unsoundGrp   int // this GrpID normalizes into an unsound card
	quarantined  []catalog.Card
}

func newFakeDeps(porProvedor int, erroAtivacao error) *fakeDeps {
	// GrpIDs start at 1: zero is unsound by the publication validator.
	return &fakeDeps{porProvedor: porProvedor, erroAtivacao: erroAtivacao,
		proximoGrp: 1}
}

func (f *fakeDeps) deps() Deps {
	return Deps{Fetcher: f, Normalize: f, Validator: SoundCard{},
		Quarantine: f, Writer: f, Objects: f, Signer: f, Activator: f}
}

// Quarantine records refused cards so tests can assert they left the
// pipeline without failing the run.
func (f *fakeDeps) Quarantine(_ context.Context, card catalog.Card,
	_ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.quarantined = append(f.quarantined, card)
	return nil
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
	if f.unsoundGrp != 0 && o.GrpID == f.unsoundGrp {
		return catalog.Card{GrpID: o.GrpID}, nil // no name/set: unsound
	}
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
