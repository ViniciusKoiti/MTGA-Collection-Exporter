// Package normalize turns raw source observations into snapshot entries
// (OpenSpec introduce-agentic-go-companion, task 3.2): catalog provenance
// is attached to resolved cards, unresolved source records are preserved
// verbatim, and skipped rows leave a diagnostic — never silence.
package normalize

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Result is the outcome of normalizing one observation.
type Result struct {
	Entries     []collection.Entry
	Diagnostics []collection.Diagnostic
	Resolved    int
	Unresolved  int
}

// Observation resolves every observed quantity through the catalog. The
// same function serves the collection-sync graph and any future caller —
// there is exactly one normalization in the product.
func Observation(
	ctx context.Context,
	catalog ports.Catalog,
	obs collection.Observation,
) (Result, error) {
	result := Result{Diagnostics: append([]collection.Diagnostic(nil), obs.Diagnostics...)}
	for _, raw := range obs.Quantities {
		if raw.Quantity == 0 {
			result.Diagnostics = append(result.Diagnostics, collection.Diagnostic{
				Code: "quantidade_zero", Detail: "line skipped",
				Severity: collection.SeverityInfo})
			continue
		}
		identity, found, err := catalog.ResolvePorArena(ctx, raw.Arena)
		if err != nil {
			return Result{}, err
		}
		if raw.Arena != 0 && found {
			result.Entries = append(result.Entries, collection.Entry{
				Identity: identity, Quantity: raw.Quantity})
			result.Resolved++
			continue
		}
		result.Entries = append(result.Entries, collection.Entry{
			Unresolved: true, Raw: raw.RawIdentity, Quantity: raw.Quantity})
		result.Unresolved++
	}
	return result, nil
}
