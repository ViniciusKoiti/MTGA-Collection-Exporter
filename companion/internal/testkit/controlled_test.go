package testkit

import (
	"context"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

func TestControlledIDsAndModelAreDeterministic(t *testing.T) {
	ids := &ControlledIDs{Prefix: "run"}
	if ids.New() != "run-0001" || ids.New() != "run-0002" {
		t.Fatal("ids must be sequential, never random")
	}
	model := ControlledModel{Script: map[string]string{
		"mono red options": "build burn"}}
	reply, err := model.Respond(context.Background(), "mono red options")
	if err != nil || reply != "build burn" {
		t.Fatalf("scripted context must answer: %q %v", reply, err)
	}
	if _, err := model.Respond(context.Background(), "improvise"); err == nil {
		t.Fatal("unscripted contexts must be refused")
	}
}

func TestControlledCatalogResolvesFixturesDeterministically(t *testing.T) {
	catalog := ControlledCatalog{DB: FixtureCardDB()}
	ctx := context.Background()
	identity, ok, err := catalog.ResolvePorArena(ctx, 1001)
	if err != nil || !ok || identity.Name != "Synthetic Bolt" {
		t.Fatalf("known arena id must resolve: %+v %v", identity, err)
	}
	if _, ok, err := catalog.ResolvePorArena(ctx, 999999); err != nil || ok {
		t.Fatalf("absence is not an error: ok=%v err=%v", ok, err)
	}
	byName, ok, err := catalog.ResolvePorNome(ctx, "Synthetic Bolt")
	if err != nil || !ok || byName.Arena != 1001 {
		t.Fatalf("name resolution must pick the lowest printing: %+v", byName)
	}
}

func TestControlledCentralFailsItsScriptedBudgetThenRecords(t *testing.T) {
	central := &ControlledCentral{FailFirst: 2}
	lote := []ports.TelemetryEvent{{Seq: 1, Name: "scan_completed"}}
	ctx := context.Background()
	for i := range 2 {
		if err := central.SendBatch(ctx, lote); err == nil {
			t.Fatalf("send %d must fail during the scripted outage", i)
		}
	}
	if err := central.SendBatch(ctx, lote); err != nil {
		t.Fatalf("the outage must end on schedule: %v", err)
	}
	if len(central.Eventos) != 1 || central.Eventos[0][0].Seq != 1 {
		t.Fatalf("recovered sends must be recorded: %+v", central.Eventos)
	}
	clipboard := &ControlledClipboard{}
	if err := clipboard.Write(ctx, "1 Synthetic Bolt"); err != nil ||
		len(clipboard.Writes) != 1 {
		t.Fatalf("effects must be recorded, never touch the OS: %v", err)
	}
}
