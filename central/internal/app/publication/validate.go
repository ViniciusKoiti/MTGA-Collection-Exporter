package publication

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// SoundCard is the default publication validator: a publishable card
// must carry a positive GrpID, a name and a set.
type SoundCard struct{}

var _ ports.CardValidator = SoundCard{}

// Check refuses structurally unsound cards.
func (SoundCard) Check(_ context.Context, card catalog.Card) error {
	if card.GrpID <= 0 || card.Name == "" || card.Set == "" {
		return fmt.Errorf("publication: unsound card %+v", card)
	}
	return nil
}
