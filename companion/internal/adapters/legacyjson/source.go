// Package legacyjson reads the legacy Python export
// (mtga_collection.json) as a collection source (OpenSpec
// introduce-agentic-go-companion, task 3.1). The accepted shape is
// frozen by golden contract fixtures: unknown fields are rejected so the
// Go adapter and the Python exporter cannot drift silently.
package legacyjson

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// legacyCard is exactly one entry of the Python export (enrich_collection
// in mtga/collection.py); every field is part of the frozen contract.
type legacyCard struct {
	Count           int      `json:"count"`
	Name            string   `json:"name"`
	Set             string   `json:"set"`
	CollectorNumber string   `json:"collector_number"`
	Rarity          string   `json:"rarity"`
	Colors          []string `json:"colors"`
	TypeLine        string   `json:"type_line"`
	ManaCost        string   `json:"mana_cost"`
	CMC             *float64 `json:"cmc"`
	Image           string   `json:"image"`
	GrpID           int      `json:"grp_id"`
}

// Source observes the collection from a legacy export file.
type Source struct {
	Path  string
	Clock ports.Clock
}

var _ ports.CollectionSource = (*Source)(nil)

// Kind identifies this source as the explicit JSON import.
func (s *Source) Kind() collection.SourceKind { return collection.SourceJSONImport }

// Observe parses the export strictly and maps each entry to a raw
// observed quantity. Cards without an arena id stay observable as raw
// records with a diagnostic — never silently dropped.
func (s *Source) Observe(_ context.Context) (collection.Observation, error) {
	payload, err := os.ReadFile(s.Path)
	if err != nil {
		return collection.Observation{}, fmt.Errorf("legacyjson: read export: %w", err)
	}
	cards, err := decode(payload)
	if err != nil {
		return collection.Observation{}, err
	}
	var quantities []collection.ObservedQuantity
	var diagnostics []collection.Diagnostic
	for _, card := range cards {
		if card.Count < 0 {
			return collection.Observation{}, fmt.Errorf(
				"legacyjson: negative count for %q", card.Name)
		}
		if card.GrpID == 0 {
			diagnostics = append(diagnostics, collection.Diagnostic{
				Code: "unknown_arena_id", Detail: card.Name,
				Severity: collection.SeverityWarning})
		}
		quantities = append(quantities, collection.ObservedQuantity{
			RawIdentity: fmt.Sprintf("%s (%s)", card.Name, card.Set),
			Arena:       collection.ArenaID(card.GrpID),
			Quantity:    card.Count,
		})
	}
	return collection.NewObservation(collection.SourceJSONImport, s.Path,
		s.Clock.Now(), quantities, diagnostics)
}

// decode enforces the frozen contract: a single JSON array of known-shape
// objects, nothing more.
func decode(payload []byte) ([]legacyCard, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var cards []legacyCard
	if err := decoder.Decode(&cards); err != nil {
		return nil, fmt.Errorf("legacyjson: export outside the frozen contract: %w", err)
	}
	if decoder.More() {
		return nil, fmt.Errorf("legacyjson: trailing content after the export array")
	}
	return cards, nil
}
