// Package inmem fornece adaptadores em memória de todos os ports do
// companion, com asserções de implementação em tempo de compilação
// (tarefa 2.3). São os adaptadores do perfil determinístico do harness.
package inmem

import (
	"context"
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Source devolve uma observação fixa (fixture) como fonte de coleção.
type Source struct {
	Tipo collection.SourceKind
	Obs  collection.Observation
	Err  error
}

var _ ports.CollectionSource = (*Source)(nil)

func (s *Source) Kind() collection.SourceKind { return s.Tipo }
func (s *Source) Observe(context.Context) (collection.Observation, error) {
	return s.Obs, s.Err
}

// Catalog resolve identidades a partir de mapas em memória.
type Catalog struct {
	PorArena map[collection.ArenaID]collection.CardIdentity
	PorNome  map[string]collection.CardIdentity
}

var _ ports.Catalog = (*Catalog)(nil)

func (c *Catalog) ResolvePorArena(_ context.Context, id collection.ArenaID) (collection.CardIdentity, bool, error) {
	identidade, ok := c.PorArena[id]
	return identidade, ok, nil
}

func (c *Catalog) ResolvePorNome(_ context.Context, nome string) (collection.CardIdentity, bool, error) {
	identidade, ok := c.PorNome[nome]
	return identidade, ok, nil
}

// SnapshotStore guarda snapshots imutáveis em memória.
type SnapshotStore struct {
	mu     sync.Mutex
	itens  map[collection.SnapshotID]collection.Snapshot
	ultimo collection.SnapshotID
}

var _ ports.SnapshotStore = (*SnapshotStore)(nil)

// NewSnapshotStore cria o store vazio.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{itens: make(map[collection.SnapshotID]collection.Snapshot)}
}

func (s *SnapshotStore) Save(_ context.Context, snap collection.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, existe := s.itens[snap.ID]; existe {
		return fmt.Errorf("inmem: snapshot %s já existe (snapshots são imutáveis)", snap.ID)
	}
	s.itens[snap.ID] = snap
	s.ultimo = snap.ID
	return nil
}

func (s *SnapshotStore) Get(_ context.Context, id collection.SnapshotID) (collection.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.itens[id]
	if !ok {
		return collection.Snapshot{}, fmt.Errorf("inmem: snapshot %s não existe", id)
	}
	return snap, nil
}

func (s *SnapshotStore) Latest(_ context.Context) (collection.Snapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ultimo == "" {
		return collection.Snapshot{}, false, nil
	}
	return s.itens[s.ultimo], true, nil
}
