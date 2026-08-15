package compatibility

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

func TestExporterMatchesLegacyProjectionShapes(t *testing.T) {
	exporter := Exporter{}
	jsonPayload, err := exporter.Export(t.Context(), snapshotFixture(t), ports.ExportJSON)
	if err != nil {
		t.Fatalf("JSON export: %v", err)
	}
	var cards []card
	if err := json.Unmarshal(jsonPayload, &cards); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if len(cards) != 2 || cards[0].Name != "Alpha Card" || cards[0].Count != 4 ||
		cards[0].GrpID != 10 || cards[1].Name != "Zeta Card" {
		t.Fatalf("legacy projection is not sorted and aggregated: %+v", cards)
	}
	if strings.Contains(string(jsonPayload), "private unknown") {
		t.Fatal("unresolved source data must not enter compatibility exports")
	}

	csvPayload, err := exporter.Export(t.Context(), snapshotFixture(t), ports.ExportCSV)
	if err != nil || !strings.Contains(string(csvPayload),
		"4,Alpha Card,ONE,Near Mint,English,,") {
		t.Fatalf("CSV export mismatch: %v\n%s", err, csvPayload)
	}
	textPayload, err := exporter.Export(t.Context(), snapshotFixture(t), ports.ExportText)
	if err != nil || string(textPayload) != "4 Alpha Card (ONE)\n1 Zeta Card (TST)\n" {
		t.Fatalf("text export mismatch: %v\n%s", err, textPayload)
	}
}

func TestExporterRejectsUnsupportedFormat(t *testing.T) {
	if _, err := (Exporter{}).Export(t.Context(), snapshotFixture(t), "xml"); err == nil {
		t.Fatal("unsupported formats must fail")
	}
}
