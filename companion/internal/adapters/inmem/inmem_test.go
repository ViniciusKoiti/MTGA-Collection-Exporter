package inmem

import (
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

func snapshotExemplo(t *testing.T) collection.Snapshot {
	t.Helper()
	obs, err := collection.NewObservation(collection.SourceJSONImport, "fixture",
		time.Unix(1_700_000_000, 0), nil, nil)
	if err != nil {
		t.Fatalf("observação: %v", err)
	}
	snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{
			{Identity: collection.CardIdentity{Printing: "p1", Arena: 101, Name: "Carta A", Set: "TST"}, Quantity: 4},
			{Unresolved: true, Raw: "misteriosa", Quantity: 1},
		}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func TestSnapshotStoreImutavelComLatest(t *testing.T) {
	store := NewSnapshotStore()
	snap := snapshotExemplo(t)
	if err := store.Save(t.Context(), snap); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.Save(t.Context(), snap); err == nil {
		t.Fatal("snapshot é imutável: segundo save do mesmo ID deveria falhar")
	}
	ultimo, existe, err := store.Latest(t.Context())
	if err != nil || !existe || ultimo.ID != snap.ID {
		t.Fatalf("latest inesperado: %+v (%v)", ultimo, err)
	}
}

func TestAprovacaoDeUsoUnicoViaPorts(t *testing.T) {
	clock := NewClock(time.Unix(1_700_000_000, 0))
	servico := NewApprovalService(clock)
	token, err := servico.Request(t.Context(), "export-approved", "hash-1")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if err := servico.Redeem(t.Context(), token.ID, "export-approved", "hash-2"); err == nil {
		t.Fatal("argumentos divergentes não resgatam o token")
	}
	if err := servico.Redeem(t.Context(), token.ID, "export-approved", "hash-1"); err != nil {
		t.Fatalf("resgate exato deveria passar: %v", err)
	}
	if err := servico.Redeem(t.Context(), token.ID, "export-approved", "hash-1"); err == nil {
		t.Fatal("token é de uso único")
	}
}

func TestExporterProjetaSomenteResolvidas(t *testing.T) {
	saida, err := Exporter{}.Export(t.Context(), snapshotExemplo(t), ports.ExportCSV)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	csv := string(saida)
	if !strings.Contains(csv, "101,Carta A,TST,4") || strings.Contains(csv, "misteriosa") {
		t.Fatalf("projeção CSV inesperada:\n%s", csv)
	}
	if _, err := (Exporter{}).Export(t.Context(), snapshotExemplo(t), "xml"); err == nil {
		t.Fatal("formato desconhecido deveria falhar")
	}
}
