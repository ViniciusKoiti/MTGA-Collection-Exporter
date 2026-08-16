package scryfall

import (
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// profile maps the bulk metadata into the substitution profile the
// decks domain consumes; a card cached before profiles existed has an
// empty type line and yields no profile.
func (c bulkCard) profile() (decks.CardProfile, bool) {
	if c.TypeLine == "" {
		return decks.CardProfile{}, false
	}
	types := strings.Fields(strings.Split(c.TypeLine, "—")[0])
	var formats []string
	for format, status := range c.Legalities {
		if status == "legal" {
			formats = append(formats, format)
		}
	}
	return decks.CardProfile{
		Name:      c.Name,
		Colors:    c.Colors,
		ManaValue: int(c.Cmc),
		Types:     types,
		Formats:   formats,
	}, true
}

// Profile answers the substitution profile of a card by name. Caches
// written before profiles carry no metadata and answer ok=false, so
// substitutions degrade honestly until the next refresh.
func (c *Catalog) Profile(name string) (decks.CardProfile, bool) {
	profile, ok := c.byProfile[strings.ToLower(name)]
	return profile, ok
}
