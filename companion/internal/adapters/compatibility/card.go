package compatibility

import (
	"sort"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// card mirrors the legacy Python collection export contract.
type card struct {
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

func projectCards(snapshot collection.Snapshot) []card {
	byPrinting := make(map[string]card)
	for _, entry := range snapshot.Entries {
		if entry.Unresolved {
			continue
		}
		key := entry.Identity.Name + "\x00" + entry.Identity.Set
		current := byPrinting[key]
		if current.Name == "" {
			current = card{Name: entry.Identity.Name, Set: entry.Identity.Set,
				Colors: []string{}, GrpID: int(entry.Identity.Arena)}
		}
		current.Count += entry.Quantity
		byPrinting[key] = current
	}
	cards := make([]card, 0, len(byPrinting))
	for _, item := range byPrinting {
		cards = append(cards, item)
	}
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Name == cards[j].Name {
			return cards[i].Set < cards[j].Set
		}
		return cards[i].Name < cards[j].Name
	})
	return cards
}
