package collection

import (
	"testing"
	"time"
)

func observacaoValida(t *testing.T) Observation {
	t.Helper()
	obs, err := NewObservation(SourceJSONImport, "export-1",
		time.Unix(1_700_000_000, 0), []ObservedQuantity{
			{RawIdentity: "12345", Arena: 12345, Quantity: 4},
			{RawIdentity: "??? carta estranha", Quantity: 1},
		}, nil)
	if err != nil {
		t.Fatalf("observação: %v", err)
	}
	return obs
}

func TestObservacaoRejeitaInvariantesQuebrados(t *testing.T) {
	agora := time.Unix(1_700_000_000, 0)
	if _, err := NewObservation("planilha", "x", agora, nil, nil); err == nil {
		t.Fatal("fonte desconhecida deveria falhar")
	}
	if _, err := NewObservation(SourceJSONImport, "x", time.Time{}, nil, nil); err == nil {
		t.Fatal("observação sem instante deveria falhar")
	}
	negativa := []ObservedQuantity{{RawIdentity: "1", Quantity: -1}}
	if _, err := NewObservation(SourceJSONImport, "x", agora, negativa, nil); err == nil {
		t.Fatal("quantidade negativa deveria falhar")
	}
}

func TestSnapshotImutavelPreservaNaoResolvidas(t *testing.T) {
	obs := observacaoValida(t)
	entradas := []Entry{
		{Identity: CardIdentity{Printing: "p1", Arena: 12345, Name: "Carta A", Set: "TST"}, Quantity: 4},
		{Unresolved: true, Raw: "??? carta estranha", Quantity: 1},
	}
	snap, err := NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute), entradas, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.Schema != SnapshotSchema || snap.TotalCartas() != 5 {
		t.Fatalf("snapshot inesperado: %+v", snap)
	}
	entradas[0].Quantity = 99 // mutação externa não pode vazar
	if snap.Entries[0].Quantity != 4 {
		t.Fatal("snapshot deveria copiar defensivamente as entradas")
	}
}

func TestSnapshotRejeitaInvariantesQuebrados(t *testing.T) {
	obs := observacaoValida(t)
	depois := obs.ObservedAt.Add(time.Minute)
	casos := map[string]func() (Snapshot, error){
		"sem id": func() (Snapshot, error) {
			return NewSnapshot("", obs, depois, nil, nil)
		},
		"importação antes da observação": func() (Snapshot, error) {
			return NewSnapshot("s", obs, obs.ObservedAt.Add(-time.Minute), nil, nil)
		},
		"quantidade zero": func() (Snapshot, error) {
			return NewSnapshot("s", obs, depois, []Entry{{Unresolved: true, Raw: "x"}}, nil)
		},
		"não resolvida sem identidade crua": func() (Snapshot, error) {
			return NewSnapshot("s", obs, depois, []Entry{{Unresolved: true, Quantity: 1}}, nil)
		},
		"resolvida sem identidade": func() (Snapshot, error) {
			return NewSnapshot("s", obs, depois, []Entry{{Quantity: 1}}, nil)
		},
	}
	for nome, constroi := range casos {
		if _, err := constroi(); err == nil {
			t.Errorf("%s: deveria falhar", nome)
		}
	}
}
