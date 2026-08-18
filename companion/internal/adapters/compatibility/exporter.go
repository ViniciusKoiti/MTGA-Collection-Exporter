// Package compatibility projects Go snapshots into the legacy Python
// JSON, CSV, and text collection formats used during migration.
package compatibility

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Exporter renders compatibility projections without performing I/O.
type Exporter struct{}

var _ ports.Exporter = Exporter{}

// Export returns a deterministic projection of resolved snapshot entries.
func (Exporter) Export(ctx context.Context, snapshot collection.Snapshot,
	format ports.ExportFormat) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cards := projectCards(snapshot)
	switch format {
	case ports.ExportJSON:
		return json.MarshalIndent(cards, "", "  ")
	case ports.ExportCSV:
		return renderCSV(cards)
	case ports.ExportText:
		return renderText(cards), nil
	default:
		return nil, fmt.Errorf("compatibility: unsupported export format %q", format)
	}
}

func renderCSV(cards []card) ([]byte, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)
	writer.UseCRLF = true
	_ = writer.Write([]string{"Count", "Name", "Edition", "Condition",
		"Language", "Foil", "Tag"})
	for _, item := range cards {
		_ = writer.Write([]string{strconv.Itoa(item.Count), item.Name, item.Set,
			"Near Mint", "English", "", ""})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return []byte(output.String()), nil
}

func renderText(cards []card) []byte {
	var output strings.Builder
	for _, item := range cards {
		fmt.Fprintf(&output, "%d %s", item.Count, item.Name)
		if item.Set != "" {
			fmt.Fprintf(&output, " (%s)", item.Set)
		}
		output.WriteByte('\n')
	}
	return []byte(output.String())
}
