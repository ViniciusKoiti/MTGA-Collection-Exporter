package worker

import (
	"errors"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// isNotFound distinguishes "queue drained" from real claim failures.
func isNotFound(err error) bool {
	return errors.Is(err, postgres.ErrNotFound)
}
